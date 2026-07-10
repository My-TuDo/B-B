package comment

import (
	"context"
	"fmt"

	commentmodel "github.com/My-TuDo/B-B/backend/internal/model/comment"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	commentrepo "github.com/My-TuDo/B-B/backend/internal/repository/comment"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/redis/go-redis/v9"
)

type Notifier interface {
	SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error
}

type Service struct {
	repo     *commentrepo.Repository
	rdb      *redis.Client
	notifier Notifier
}

func NewService(repo *commentrepo.Repository, rdb *redis.Client, notifier Notifier) *Service {
	return &Service{repo: repo, rdb: rdb, notifier: notifier}
}

func (s *Service) GetComments(ctx context.Context, videoID uint, page, pageSize int, sort string) (*commentmodel.CommentListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	list, total, err := s.repo.FindRootComments(ctx, videoID, sort, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("comment.service.GetComments: %w", err)
	}

	// Collect root comment IDs for batch reply fetch
	rootIDs := make([]uint, 0, len(list))
	for _, c := range list {
		rootIDs = append(rootIDs, c.ID)
	}

	// Fetch all replies in one query
	var replyMap map[uint][]commentmodel.CommentResp
	if len(rootIDs) > 0 {
		replies, err := s.repo.FindRepliesByRootIDs(ctx, rootIDs)
		if err != nil {
			return nil, fmt.Errorf("comment.service.GetComments: %w", err)
		}
		if len(replies) > 0 {
			replyMap = make(map[uint][]commentmodel.CommentResp, len(rootIDs))
			for i := range replies {
				r := toCommentResp(ctx, &replies[i])
				replyMap[r.RootID] = append(replyMap[r.RootID], *r)
			}
		}
	}

	items := make([]commentmodel.CommentResp, 0, len(list))
	for _, c := range list {
		resp := toCommentResp(ctx, &c)
		if replyMap != nil {
			resp.Replies = replyMap[c.ID]
		}
		items = append(items, *resp)
	}

	return &commentmodel.CommentListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) CreateComment(ctx context.Context, videoID uint, userID uint, req *commentmodel.CommentReq) (*commentmodel.CommentResp, error) {
	parentID := uint(0)
	rootID := uint(0)
	if req.ParentID != nil {
		parentID = *req.ParentID
	}
	if req.RootID != nil {
		rootID = *req.RootID
	}

	// Bug 2 fix: validate video exists before creating comment
	exists, err := s.repo.ExistsVideo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("comment.service.CreateComment: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("comment.service.CreateComment: %w", newError(errcode.VideoNotFound))
	}

	// Bug 1 fix: auto-derive root_id from parent comment
	if parentID > 0 && rootID == 0 {
		parent, err := s.repo.FindByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("comment.service.CreateComment: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("comment.service.CreateComment: %w", newError(errcode.CommentNotFound))
		}
		if parent.ParentID == 0 {
			rootID = parentID
		} else {
			rootID = parent.RootID
		}
	}

	c := &commentmodel.Comment{
		VideoID:  videoID,
		UserID:   userID,
		ParentID: parentID,
		RootID:   rootID,
		Content:  req.Content,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("comment.service.CreateComment: %w", err)
	}

	// Re-fetch with User
	created, _ := s.repo.FindByID(ctx, c.ID)
	if created == nil {
		created = c
	}

	// Send notification for reply (not self-reply)
	if parentID > 0 && s.notifier != nil {
		parentComment, err := s.repo.FindByID(ctx, parentID)
		if err == nil && parentComment != nil && parentComment.UserID != userID {
			replyUsername := "有人"
			if created.User.ID != 0 {
				replyUsername = created.User.Nickname
				if replyUsername == "" {
					replyUsername = created.User.Username
				}
			}
			content := fmt.Sprintf("%s 回复了你的评论: %s", replyUsername, truncateContent(req.Content, 50))
			_ = s.notifier.SendNotification(ctx, parentComment.UserID, userID, 1, videoID, content)
		}
	}

	return toCommentResp(ctx, created), nil
}

func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}

func (s *Service) LikeComment(ctx context.Context, commentID uint, userID uint) (*commentmodel.CommentLikeResp, error) {
	key := fmt.Sprintf("comment:like:%d", commentID)

	liked := false
	if s.rdb != nil {
		added, err := s.rdb.SAdd(ctx, key, userID).Result()
		if err != nil {
			return nil, fmt.Errorf("comment.service.LikeComment: %w", err)
		}
		if added == 0 {
			// Already liked → unlike
			s.rdb.SRem(ctx, key, userID)
		} else {
			liked = true
		}
	}

	// Count likes from Redis
	likeCount := int64(0)
	if s.rdb != nil {
		likeCount, _ = s.rdb.SCard(ctx, key).Result()
	}

	return &commentmodel.CommentLikeResp{
		Liked: liked,
		Likes: uint(likeCount),
	}, nil
}

func (s *Service) DeleteComment(ctx context.Context, commentID uint, userID uint, videoID uint) error {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("comment.service.DeleteComment: %w", err)
	}
	if comment == nil {
		return fmt.Errorf("comment.service.DeleteComment: %w", newError(errcode.CommentNotFound))
	}

	// Check if user is comment author OR video author
	if comment.UserID == userID {
		// Comment author - can delete
	} else if videoID > 0 {
		// Check if user is the video author
		// We use a raw query since comment repo doesn't have video lookup
		var videoAuthorID uint
		if err := s.repo.FindVideoAuthor(ctx, videoID, &videoAuthorID); err != nil || videoAuthorID != userID {
			return fmt.Errorf("comment.service.DeleteComment: %w", newError(errcode.Forbidden))
		}
	} else {
		return fmt.Errorf("comment.service.DeleteComment: %w", newError(errcode.Forbidden))
	}

	if err := s.repo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("comment.service.DeleteComment: %w", err)
	}
	return nil
}

func (s *Service) SyncLikes(ctx context.Context) {
	// Called periodically to sync Redis likes to MySQL
}

func toCommentResp(ctx context.Context, c *commentmodel.Comment) *commentmodel.CommentResp {
	resp := &commentmodel.CommentResp{
		ID:        c.ID,
		VideoID:   c.VideoID,
		UserID:    c.UserID,
		ParentID:  c.ParentID,
		RootID:    c.RootID,
		Content:   c.Content,
		Likes:     c.Likes,
		CreatedAt: c.CreatedAt,
	}
	if c.User.ID != 0 {
		avatar := ""
		if c.User.Avatar != "" {
			avatar = storage.GetObjectURL(c.User.Avatar)
		}
		resp.User = &usermodel.UserBrief{
			ID:       c.User.ID,
			Username: c.User.Username,
			Nickname: c.User.Nickname,
			Avatar:   avatar,
		}
	}
	return resp
}

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

func newError(code int) *Error {
	return &Error{Code: code, Msg: errcode.Message(code)}
}

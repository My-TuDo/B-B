// Package comment 提供评论相关的业务逻辑服务，
// 包括评论列表查询、创建评论、评论点赞/取消点赞、删除评论等功能，
// 支持根评论与子回复的层级结构。
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

// Notifier 通知发送接口，用于在评论回复时向被回复用户发送通知。
type Notifier interface {
	SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error
}

// Service 评论服务，封装评论相关的业务逻辑。
type Service struct {
	repo     *commentrepo.Repository
	rdb      *redis.Client
	notifier Notifier
}

// NewService 创建评论服务实例。
func NewService(repo *commentrepo.Repository, rdb *redis.Client, notifier Notifier) *Service {
	return &Service{repo: repo, rdb: rdb, notifier: notifier}
}

// GetComments 分页获取指定视频的根评论列表，同时批量拉取每个根评论的子回复。
// sort 控制排序方式，如 "hot" 按热度、"new" 按时间。
func (s *Service) GetComments(ctx context.Context, videoID uint, page, pageSize int, sort string) (*commentmodel.CommentListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	// 查询根评论
	offset := (page - 1) * pageSize
	list, total, err := s.repo.FindRootComments(ctx, videoID, sort, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("comment.service.GetComments: %w", err)
	}

	// Collect root comment IDs for batch reply fetch
	// 收集所有根评论ID，用于批量查询子回复
	rootIDs := make([]uint, 0, len(list))
	for _, c := range list {
		rootIDs = append(rootIDs, c.ID)
	}

	// Fetch all replies in one query — 一次性批量查询所有子回复
	var replyMap map[uint][]commentmodel.CommentResp
	if len(rootIDs) > 0 {
		replies, err := s.repo.FindRepliesByRootIDs(ctx, rootIDs)
		if err != nil {
			return nil, fmt.Errorf("comment.service.GetComments: %w", err)
		}
		if len(replies) > 0 {
			// 按 rootID 分组
			replyMap = make(map[uint][]commentmodel.CommentResp, len(rootIDs))
			for i := range replies {
				r := toCommentResp(ctx, &replies[i])
				replyMap[r.RootID] = append(replyMap[r.RootID], *r)
			}
		}
	}

	// 组装响应，将子回复挂到对应的根评论下
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

// CreateComment 创建评论或回复。如果是回复，会自动推导 rootID；
// 同时验证目标视频存在性。创建成功后向被回复用户推送通知。
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
	// 验证视频是否存在
	exists, err := s.repo.ExistsVideo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("comment.service.CreateComment: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("comment.service.CreateComment: %w", newError(errcode.VideoNotFound))
	}

	// Bug 1 fix: auto-derive root_id from parent comment
	// 自动推导 rootID：若为子回复且未显式指定rootID，则从父评论推导
	if parentID > 0 && rootID == 0 {
		parent, err := s.repo.FindByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("comment.service.CreateComment: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("comment.service.CreateComment: %w", newError(errcode.CommentNotFound))
		}
		if parent.ParentID == 0 {
			// 父评论即为根评论
			rootID = parentID
		} else {
			// 父评论本身也是回复，继承其 rootID
			rootID = parent.RootID
		}
	}

	// 构建并创建评论记录
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

	// Re-fetch with User — 重新查询以获取关联的用户信息
	created, _ := s.repo.FindByID(ctx, c.ID)
	if created == nil {
		created = c
	}

	// Send notification for reply (not self-reply)
	// 若为回复评论且不是自回复，则向被回复用户发送通知
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

// truncateContent 截断内容至指定最大字符数（按rune计数），超出部分用 "..." 替代。
func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}

// LikeComment 点赞/取消点赞评论。使用Redis Set存储点赞用户ID，
// 已点赞则移除（取消），未点赞则添加（点赞）。
// 返回点赞状态和当前点赞总数。
func (s *Service) LikeComment(ctx context.Context, commentID uint, userID uint) (*commentmodel.CommentLikeResp, error) {
	key := fmt.Sprintf("comment:like:%d", commentID)

	liked := false
	if s.rdb != nil {
		// SAdd 返回1表示成功添加（点赞），0表示已存在
		added, err := s.rdb.SAdd(ctx, key, userID).Result()
		if err != nil {
			return nil, fmt.Errorf("comment.service.LikeComment: %w", err)
		}
		if added == 0 {
			// Already liked → unlike — 已点赞则取消
			s.rdb.SRem(ctx, key, userID)
		} else {
			liked = true
		}
	}

	// Count likes from Redis — 从Redis获取最新点赞数
	likeCount := int64(0)
	if s.rdb != nil {
		likeCount, _ = s.rdb.SCard(ctx, key).Result()
	}

	return &commentmodel.CommentLikeResp{
		Liked: liked,
		Likes: uint(likeCount),
	}, nil
}

// DeleteComment 删除评论。仅评论作者或视频作者可以删除。
func (s *Service) DeleteComment(ctx context.Context, commentID uint, userID uint, videoID uint) error {
	// 查询评论是否存在
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("comment.service.DeleteComment: %w", err)
	}
	if comment == nil {
		return fmt.Errorf("comment.service.DeleteComment: %w", newError(errcode.CommentNotFound))
	}

	// Check if user is comment author OR video author
	// 权限校验：评论作者或视频作者可以删除
	if comment.UserID == userID {
		// Comment author - can delete — 评论作者直接删除
	} else if videoID > 0 {
		// Check if user is the video author — 检查是否为视频作者
		var videoAuthorID uint
		if err := s.repo.FindVideoAuthor(ctx, videoID, &videoAuthorID); err != nil || videoAuthorID != userID {
			return fmt.Errorf("comment.service.DeleteComment: %w", newError(errcode.Forbidden))
		}
	} else {
		return fmt.Errorf("comment.service.DeleteComment: %w", newError(errcode.Forbidden))
	}

	// 执行删除
	if err := s.repo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("comment.service.DeleteComment: %w", err)
	}
	return nil
}

// SyncLikes 定时将Redis中的点赞数据同步回MySQL。
// Called periodically to sync Redis likes to MySQL
func (s *Service) SyncLikes(ctx context.Context) {
	// TODO: 实现Redis点赞数据到MySQL的同步逻辑
}

// toCommentResp 将评论模型转换为对外响应结构，包括用户头像预签名。
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
	// 填充评论作者简要信息
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

// Error 服务层错误类型，携带错误码以支持在HTTP层映射为合适的响应。
type Error struct {
	Code int
	Msg  string
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	return e.Msg
}

// newError 根据错误码创建带本地化消息的服务错误。
func newError(code int) *Error {
	return &Error{Code: code, Msg: errcode.Message(code)}
}

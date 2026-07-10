package follow

import (
	"context"
	"fmt"

	followmodel "github.com/My-TuDo/B-B/backend/internal/model/follow"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	followrepo "github.com/My-TuDo/B-B/backend/internal/repository/follow"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"gorm.io/gorm"
)

type Notifier interface {
	SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error
}

type Service struct {
	repo     *followrepo.Repository
	db       *gorm.DB
	notifier Notifier
}

func NewService(repo *followrepo.Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

func NewServiceWithNotifier(repo *followrepo.Repository, db *gorm.DB, notifier Notifier) *Service {
	return &Service{repo: repo, db: db, notifier: notifier}
}

func (s *Service) ToggleFollow(ctx context.Context, followerID, followingID uint) (*followmodel.FollowResp, error) {
	if followerID == followingID {
		return nil, fmt.Errorf("follow.service.ToggleFollow: %w", newError(errcode.CannotFollowSelf))
	}

	exists, err := s.repo.Exists(ctx, followerID, followingID)
	if err != nil {
		return nil, fmt.Errorf("follow.service.ToggleFollow: %w", err)
	}

	if exists {
		if err := s.repo.Delete(ctx, followerID, followingID); err != nil {
			return nil, fmt.Errorf("follow.service.ToggleFollow: %w", err)
		}
		return &followmodel.FollowResp{Following: false}, nil
	}

	f := &followmodel.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("follow.service.ToggleFollow: %w", err)
	}

	// Send notification to the followed user
	if s.notifier != nil {
		// Look up follower's name for the notification content
		var followerUser usermodel.User
		followerName := "有人"
		if err := s.db.WithContext(ctx).Select("nickname, username").First(&followerUser, followerID).Error; err == nil {
			followerName = followerUser.Nickname
			if followerName == "" {
				followerName = followerUser.Username
			}
		}
		content := fmt.Sprintf("%s 关注了你", followerName)
		_ = s.notifier.SendNotification(ctx, followingID, followerID, 3, 0, content)
	}

	return &followmodel.FollowResp{Following: true}, nil
}

func (s *Service) GetFollowers(ctx context.Context, userID uint, page, pageSize int) (*followmodel.UserListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	users, total, err := s.repo.FindFollowers(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("follow.service.GetFollowers: %w", err)
	}

	items := make([]usermodel.UserBrief, len(users))
	for i, u := range users {
		avatar := ""
		if u.Avatar != "" {
			avatar = storage.GetObjectURL(u.Avatar)
		}
		items[i] = usermodel.UserBrief{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   avatar,
		}
	}

	return &followmodel.UserListResp{Items: items, Total: total}, nil
}

func (s *Service) GetFollowing(ctx context.Context, userID uint, page, pageSize int) (*followmodel.UserListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	users, total, err := s.repo.FindFollowing(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("follow.service.GetFollowing: %w", err)
	}

	items := make([]usermodel.UserBrief, len(users))
	for i, u := range users {
		avatar := ""
		if u.Avatar != "" {
			avatar = storage.GetObjectURL(u.Avatar)
		}
		items[i] = usermodel.UserBrief{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   avatar,
		}
	}

	return &followmodel.UserListResp{Items: items, Total: total}, nil
}

func (s *Service) GetProfile(ctx context.Context, targetUserID uint, viewerID uint) (*followmodel.ProfileResp, error) {
	var u usermodel.User
	if err := s.db.WithContext(ctx).First(&u, targetUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("follow.service.GetProfile: %w", newError(errcode.UserNotFound))
		}
		return nil, fmt.Errorf("follow.service.GetProfile: %w", err)
	}

	videos, _ := s.repo.CountVideosByUser(ctx, targetUserID)
	followers, _ := s.repo.CountFollowers(ctx, targetUserID)
	following, _ := s.repo.CountFollowing(ctx, targetUserID)

	avatar := ""
	if u.Avatar != "" {
		avatar = storage.GetObjectURL(u.Avatar)
	}

	return &followmodel.ProfileResp{
		User: &usermodel.UserResp{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Avatar:    avatar,
			Bio:       u.Bio,
			CreatedAt: u.CreatedAt,
		},
		Stats: followmodel.ProfileStats{
			Videos:    videos,
			Followers: followers,
			Following: following,
		},
	}, nil
}

func (s *Service) GetFeed(ctx context.Context, userID uint, page, pageSize int) (*videomodel.VideoListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	followingIDs, err := s.repo.FindFollowingIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("follow.service.GetFeed: %w", err)
	}
	if len(followingIDs) == 0 {
		return &videomodel.VideoListResp{
			Items:    []videomodel.VideoResp{},
			Total:    0,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	var total int64
	s.db.WithContext(ctx).Model(&videomodel.Video{}).Where("user_id IN ? AND status = 1", followingIDs).Count(&total)

	offset := (page - 1) * pageSize
	var videos []videomodel.Video
	s.db.WithContext(ctx).Preload("User").Where("user_id IN ? AND status = 1", followingIDs).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos)

	items := make([]videomodel.VideoResp, len(videos))
	for i, v := range videos {
		items[i] = videomodel.VideoResp{
			ID:          v.ID,
			Title:       v.Title,
			Description: v.Description,
			CoverURL:    v.CoverURL,
			VideoURL:    v.VideoURL,
			Duration:    v.Duration,
			FileSize:    v.FileSize,
			CategoryID:  v.CategoryID,
			Status:      v.Status,
			Views:       v.Views,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
		if v.User.ID != 0 {
			avatar := ""
			if v.User.Avatar != "" {
				avatar = storage.GetObjectURL(v.User.Avatar)
			}
			items[i].User = &usermodel.UserBrief{
				ID:       v.User.ID,
				Username: v.User.Username,
				Nickname: v.User.Nickname,
				Avatar:   avatar,
			}
		}
		// Presign cover URL — use direct public URL since bucket is public
		if items[i].CoverURL != "" {
			items[i].CoverURL = storage.GetObjectURL(items[i].CoverURL)
		}
	}

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
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

// Package follow 提供关注/粉丝相关的业务逻辑服务，
// 包括关注/取关、粉丝列表、关注列表、用户主页和关注动态流等功能。
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

// Notifier 通知发送接口，用于在关注时向被关注用户发送通知。
type Notifier interface {
	SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error
}

// Service 关注服务，封装关注/粉丝相关的业务逻辑。
type Service struct {
	repo     *followrepo.Repository
	db       *gorm.DB
	notifier Notifier
}

// NewService 创建关注服务实例（不含通知功能）。
func NewService(repo *followrepo.Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// NewServiceWithNotifier 创建关注服务实例（含通知功能）。
func NewServiceWithNotifier(repo *followrepo.Repository, db *gorm.DB, notifier Notifier) *Service {
	return &Service{repo: repo, db: db, notifier: notifier}
}

// ToggleFollow 切换关注状态：已关注则取消，未关注则关注。
// 不允许自己关注自己。关注成功后向被关注用户发送通知。
func (s *Service) ToggleFollow(ctx context.Context, followerID, followingID uint) (*followmodel.FollowResp, error) {
	// 不允许自关注
	if followerID == followingID {
		return nil, fmt.Errorf("follow.service.ToggleFollow: %w", newError(errcode.CannotFollowSelf))
	}

	// 检查当前关注状态
	exists, err := s.repo.Exists(ctx, followerID, followingID)
	if err != nil {
		return nil, fmt.Errorf("follow.service.ToggleFollow: %w", err)
	}

	if exists {
		// 已关注 → 取消关注
		if err := s.repo.Delete(ctx, followerID, followingID); err != nil {
			return nil, fmt.Errorf("follow.service.ToggleFollow: %w", err)
		}
		return &followmodel.FollowResp{Following: false}, nil
	}

	// 未关注 → 添加关注
	f := &followmodel.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("follow.service.ToggleFollow: %w", err)
	}

	// Send notification to the followed user — 向被关注用户发送通知
	if s.notifier != nil {
		// Look up follower's name for the notification content — 查询关注者的昵称
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

// GetFollowers 分页获取指定用户的粉丝列表。
func (s *Service) GetFollowers(ctx context.Context, userID uint, page, pageSize int) (*followmodel.UserListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	// 查询粉丝
	offset := (page - 1) * pageSize
	users, total, err := s.repo.FindFollowers(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("follow.service.GetFollowers: %w", err)
	}

	// 组装响应，生成头像预签名URL
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

// GetFollowing 分页获取指定用户正在关注的人的列表。
func (s *Service) GetFollowing(ctx context.Context, userID uint, page, pageSize int) (*followmodel.UserListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	// 查询关注列表
	offset := (page - 1) * pageSize
	users, total, err := s.repo.FindFollowing(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("follow.service.GetFollowing: %w", err)
	}

	// 组装响应，生成头像预签名URL
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

// GetProfile 获取目标用户的个人主页信息，包含用户资料和统计数据
// （视频数、粉丝数、关注数）。viewerID 为当前查看者（0 表示未登录）。
func (s *Service) GetProfile(ctx context.Context, targetUserID uint, viewerID uint) (*followmodel.ProfileResp, error) {
	// 查询目标用户
	var u usermodel.User
	if err := s.db.WithContext(ctx).First(&u, targetUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("follow.service.GetProfile: %w", newError(errcode.UserNotFound))
		}
		return nil, fmt.Errorf("follow.service.GetProfile: %w", err)
	}

	// 统计数据
	videos, _ := s.repo.CountVideosByUser(ctx, targetUserID)
	followers, _ := s.repo.CountFollowers(ctx, targetUserID)
	following, _ := s.repo.CountFollowing(ctx, targetUserID)

	// 生成头像预签名URL
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

// GetFeed 获取当前用户的关注动态流：展示已关注用户发布的视频，
// 按发布时间倒序排列并分页。若用户未关注任何人则返回空列表。
func (s *Service) GetFeed(ctx context.Context, userID uint, page, pageSize int) (*videomodel.VideoListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 获取当前用户关注的所有用户ID
	followingIDs, err := s.repo.FindFollowingIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("follow.service.GetFeed: %w", err)
	}
	// 未关注任何人则返回空列表
	if len(followingIDs) == 0 {
		return &videomodel.VideoListResp{
			Items:    []videomodel.VideoResp{},
			Total:    0,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	// 统计总数
	var total int64
	s.db.WithContext(ctx).Model(&videomodel.Video{}).Where("user_id IN ? AND status = 1", followingIDs).Count(&total)

	// 分页查询视频
	offset := (page - 1) * pageSize
	var videos []videomodel.Video
	s.db.WithContext(ctx).Preload("User").Where("user_id IN ? AND status = 1", followingIDs).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos)

	// 组装响应
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
		// Presign cover URL — 为封面生成公网访问URL
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

// Package history 提供观看历史相关的业务逻辑服务，
// 包括观看记录的创建/更新和历史记录的分页查询。
package history

import (
	"context"
	"fmt"
	"time"

	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	historyrepo "github.com/My-TuDo/B-B/backend/internal/repository/history"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

// Service 观看历史服务，封装观看记录相关的业务逻辑。
type Service struct {
	repo      *historyrepo.Repository
	videoRepo *videorepo.Repository
}

// NewService 创建观看历史服务实例。
func NewService(repo *historyrepo.Repository, videoRepo *videorepo.Repository) *Service {
	return &Service{repo: repo, videoRepo: videoRepo}
}

// CreateOrUpdate 创建或更新观看历史记录（按用户+视频去重，upsert模式）。
// 同时尝试更新视频时长信息。
func (s *Service) CreateOrUpdate(ctx context.Context, userID uint, req *historymodel.CreateHistoryReq) error {
	// 构建观看记录实体
	h := &historymodel.VideoHistory{
		UserID:    userID,
		VideoID:   req.VideoID,
		Progress:  req.Progress,
		WatchedAt: time.Now(),
	}

	// 写入或更新记录
	if err := s.repo.CreateOrUpdate(ctx, h); err != nil {
		return fmt.Errorf("history.service.CreateOrUpdate: %w", err)
	}

	// 如果传入了视频时长，同步更新视频记录的时长
	if req.Duration > 0 {
		_ = s.videoRepo.UpdateDuration(ctx, req.VideoID, uint(req.Duration))
	}

	return nil
}

// List 分页获取当前用户的观看历史，按观看时间倒序排列。
func (s *Service) List(ctx context.Context, userID uint, page, pageSize int) (*historymodel.HistoryListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 查询历史记录（含关联视频和用户信息）
	histories, total, err := s.repo.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("history.service.List: %w", err)
	}

	// 组装响应列表
	items := make([]historymodel.HistoryItemResp, len(histories))
	for i, h := range histories {
		v := h.Video
		vr := videomodel.VideoResp{
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
		// 填充视频作者信息
		if v.User.ID != 0 {
			vr.User = &usermodel.UserBrief{
				ID:       v.User.ID,
				Username: v.User.Username,
				Nickname: v.User.Nickname,
				Avatar:   v.User.Avatar,
			}
		}
		items[i] = historymodel.HistoryItemResp{
			Video:     vr,
			Progress:  h.Progress,
			WatchedAt: h.WatchedAt,
		}
	}

	// Presign cover URLs and user avatars — 为封面和头像生成公网URL
	for i := range items {
		if items[i].Video.CoverURL != "" {
			items[i].Video.CoverURL = storage.GetObjectURL(items[i].Video.CoverURL)
		}
		if items[i].Video.User != nil && items[i].Video.User.Avatar != "" {
			items[i].Video.User.Avatar = storage.GetObjectURL(items[i].Video.User.Avatar)
		}
	}

	return &historymodel.HistoryListResp{
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

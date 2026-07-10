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

type Service struct {
	repo      *historyrepo.Repository
	videoRepo *videorepo.Repository
}

func NewService(repo *historyrepo.Repository, videoRepo *videorepo.Repository) *Service {
	return &Service{repo: repo, videoRepo: videoRepo}
}

func (s *Service) CreateOrUpdate(ctx context.Context, userID uint, req *historymodel.CreateHistoryReq) error {
	h := &historymodel.VideoHistory{
		UserID:    userID,
		VideoID:   req.VideoID,
		Progress:  req.Progress,
		WatchedAt: time.Now(),
	}

	if err := s.repo.CreateOrUpdate(ctx, h); err != nil {
		return fmt.Errorf("history.service.CreateOrUpdate: %w", err)
	}

	if req.Duration > 0 {
		_ = s.videoRepo.UpdateDuration(ctx, req.VideoID, uint(req.Duration))
	}

	return nil
}

func (s *Service) List(ctx context.Context, userID uint, page, pageSize int) (*historymodel.HistoryListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	histories, total, err := s.repo.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("history.service.List: %w", err)
	}

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

	// Presign cover URLs and user avatars
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

// Package video 提供视频数据访问层，封装视频相关的数据库 CRUD 操作。
// 支持视频的创建、查询（按 ID、用户、状态、分类）、更新、以及播放量自增和时长更新。
// 是视频业务逻辑与数据库之间的核心桥梁。
package video

import (
	"context"
	"fmt"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

// Repository 视频数据仓库，封装视频相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建视频数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一条视频记录。
// 视频元信息（标题、描述、封面、时长、文件大小等）通过指针传入并写入数据库。
func (r *Repository) Create(ctx context.Context, video *videomodel.Video) error {
	result := r.db.WithContext(ctx).Create(video)
	if result.Error != nil {
		return fmt.Errorf("video.repository.Create: %w", result.Error)
	}
	return nil
}

// FindByID 根据视频 ID 查找视频，同时预加载作者（User）信息。
// 若未找到则返回 (nil, nil)。
func (r *Repository) FindByID(ctx context.Context, id uint) (*videomodel.Video, error) {
	var video videomodel.Video
	result := r.db.WithContext(ctx).Preload("User").First(&video, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("video.repository.FindByID: %w", result.Error)
	}
	return &video, nil
}

// Update 全量更新视频记录（使用 Save 方法）。
// 注意：Save 会更新所有字段，包括零值字段。
func (r *Repository) Update(ctx context.Context, video *videomodel.Video) error {
	result := r.db.WithContext(ctx).Save(video)
	if result.Error != nil {
		return fmt.Errorf("video.repository.Update: %w", result.Error)
	}
	return nil
}

// List 分页查询已审核通过的公开视频列表（status = 1）。
// 支持按分类 ID 筛选，categoryID 为 0 时查询所有分类。
// 按创建时间倒序排列，预加载作者信息。
func (r *Repository) List(ctx context.Context, page, pageSize int, categoryID uint) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	// 基础查询：仅查询已审核通过的视频
	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("status = ?", 1)
	// 若指定分类 ID，则添加分类筛选条件
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	// 统计符合条件的视频总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.List.Count: %w", err)
	}

	// 分页查询，预加载作者信息
	offset := (page - 1) * pageSize
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.List.Find: %w", err)
	}

	return videos, total, nil
}

// ListByUser 查询指定用户的视频列表，支持按审核状态筛选。
// status 为 nil 时排除已删除视频（status != 3）；非 nil 时精确匹配指定状态。
// 按创建时间倒序排列，预加载作者信息。
func (r *Repository) ListByUser(ctx context.Context, userID uint, status *int8, page, pageSize int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	// 筛选属于该用户的视频
	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("user_id = ?", userID)
	// 根据 status 参数决定筛选逻辑
	if status != nil {
		query = query.Where("status = ?", *status)
	} else {
		query = query.Where("status != ?", 3) // 未指定时排除已删除
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.ListByUser.Count: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.ListByUser.Find: %w", err)
	}

	return videos, total, nil
}

// ListAllPublic 查询所有已审核通过的公开视频（不分页）。
// 通常用于站点地图或 RSS 订阅等需要全量数据的场景。
func (r *Repository) ListAllPublic(ctx context.Context) ([]videomodel.Video, error) {
	var videos []videomodel.Video
	result := r.db.WithContext(ctx).Preload("User").Where("status = ?", 1).Find(&videos)
	if result.Error != nil {
		return nil, fmt.Errorf("video.repository.ListAllPublic: %w", result.Error)
	}
	return videos, nil
}

// ListByStatus 根据审核状态分页查询视频列表。
// 用于管理后台按状态筛选视频（待审/通过/拒绝/删除）。
func (r *Repository) ListByStatus(ctx context.Context, status int8, page, pageSize int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.ListByStatus.Count: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.ListByStatus.Find: %w", err)
	}

	return videos, total, nil
}

// UpdateDuration 更新视频的时长字段。
// 仅当 duration 当前值为 0 时才更新（防止覆盖已解析的时长），
// 适用于异步解析视频元信息后回填时长的场景。
func (r *Repository) UpdateDuration(ctx context.Context, videoID uint, duration uint) error {
	result := r.db.WithContext(ctx).Model(&videomodel.Video{}).
		Where("id = ? AND duration = 0", videoID). // 仅在时长为 0 时更新
		Update("duration", duration)
	if result.Error != nil {
		return fmt.Errorf("video.repository.UpdateDuration: %w", result.Error)
	}
	return nil
}

// IncrementViews 原子性地将视频播放量加 1。
// 使用 gorm.Expr("views + 1") 保证并发安全，避免读-改-写竞态。
func (r *Repository) IncrementViews(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1")) // 原子自增，避免竞态
	if result.Error != nil {
		return fmt.Errorf("video.repository.IncrementViews: %w", result.Error)
	}
	return nil
}

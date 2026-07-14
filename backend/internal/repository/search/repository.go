// Package search 提供搜索数据访问层，封装视频全文搜索相关的数据库操作。
// 支持 MySQL FULLTEXT 全文搜索和 LIKE 模糊搜索的自动降级策略，
// 同时提供搜索建议（自动补全）功能，结合视频标题和标签名称。
package search

import (
	"context"
	"fmt"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

// Repository 搜索数据仓库，封装搜索相关的数据库查询操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建搜索数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Search 根据关键词搜索视频，支持分页。
// 优先尝试 MySQL FULLTEXT 全文索引搜索（MATCH ... AGAINST），
// 若全文搜索失败或无结果则自动降级为 LIKE 模糊搜索。
// 搜索范围包括视频标题、描述以及关联的标签名称。
func (r *Repository) Search(ctx context.Context, q string, page, pageSize int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	offset := (page - 1) * pageSize
	likeQ := "%" + q + "%"

	// 第一步：尝试 FULLTEXT 全文搜索统计总数
	countSQL := "SELECT COUNT(DISTINCT v.id) FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND MATCH(v.title, v.description) AGAINST(? IN BOOLEAN MODE)"
	fulltextErr := r.db.WithContext(ctx).Raw(countSQL, 1, q).Scan(&total).Error

	// 第二步：若全文搜索失败或结果为空，降级为 LIKE 模糊搜索
	if fulltextErr != nil || total == 0 {
		// 使用 LIKE 进行模糊匹配（包括标签名称匹配）
		if err := r.db.WithContext(ctx).Raw(
			"SELECT COUNT(DISTINCT v.id) FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND (v.title LIKE ? OR v.description LIKE ? OR t.name LIKE ?)",
			1, likeQ, likeQ, likeQ,
		).Scan(&total).Error; err != nil {
			return nil, 0, fmt.Errorf("search.repository.Search.Count: %w", err)
		}

		// 无结果时提前返回
		if total == 0 {
			return videos, 0, nil
		}

		// LIKE 模糊搜索分页查询
		querySQL := "SELECT DISTINCT v.* FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND (v.title LIKE ? OR v.description LIKE ? OR t.name LIKE ?) ORDER BY v.created_at DESC LIMIT ? OFFSET ?"
		if err := r.db.WithContext(ctx).Raw(querySQL, 1, likeQ, likeQ, likeQ, pageSize, offset).Scan(&videos).Error; err != nil {
			return nil, 0, fmt.Errorf("search.repository.Search.Find: %w", err)
		}
	} else {
		// 全文搜索分页查询
		querySQL := "SELECT DISTINCT v.* FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND MATCH(v.title, v.description) AGAINST(? IN BOOLEAN MODE) ORDER BY v.created_at DESC LIMIT ? OFFSET ?"
		if err := r.db.WithContext(ctx).Raw(querySQL, 1, q, pageSize, offset).Scan(&videos).Error; err != nil {
			return nil, 0, fmt.Errorf("search.repository.Search.Find: %w", err)
		}
	}

	// 第三步：为每个视频加载关联的用户信息
	for i := range videos {
		var user struct {
			ID       uint
			Username string
			Nickname string
			Avatar   string
		}
		r.db.WithContext(ctx).Table("users").Where("id = ?", videos[i].UserID).
			Select("id, username, nickname, avatar").Scan(&user)
		videos[i].User.ID = user.ID
		videos[i].User.Username = user.Username
		videos[i].User.Nickname = user.Nickname
		videos[i].User.Avatar = user.Avatar
	}

	return videos, total, nil
}

// SearchSuggestions 获取搜索建议（自动补全），结合视频标题和标签名称。
// 从公开视频标题和标签名称中匹配关键词，按视频播放量降序排列。
// limit 控制返回的建议条数。
func (r *Repository) SearchSuggestions(ctx context.Context, q string, limit int) ([]struct {
	Keyword string
	Count   int64
}, error) {
	type row struct {
		Keyword string
		Count   int64
	}
	var rows []row

	likeQ := "%" + q + "%"
	// UNION 查询：合并视频标题匹配和标签名称匹配的结果
	querySQL := `
		(SELECT title AS keyword, views AS count FROM videos WHERE status = ? AND title LIKE ?)
		UNION
		(SELECT DISTINCT t.name COLLATE utf8mb4_unicode_ci AS keyword, 0 AS count FROM tags t WHERE t.name LIKE ?)
		ORDER BY count DESC LIMIT ?
	`
	if err := r.db.WithContext(ctx).Raw(querySQL, 1, likeQ, likeQ, limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("search.repository.SearchSuggestions: %w", err)
	}

	// 转换匿名结构体切片为返回类型
	result := make([]struct {
		Keyword string
		Count   int64
	}, len(rows))
	for i, r := range rows {
		result[i].Keyword = r.Keyword
		result[i].Count = r.Count
	}

	return result, nil
}

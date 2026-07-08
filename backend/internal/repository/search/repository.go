package search

import (
	"context"
	"fmt"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Search(ctx context.Context, q string, page, pageSize int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	offset := (page - 1) * pageSize
	likeQ := "%" + q + "%"

	// Try FULLTEXT search first
	countSQL := "SELECT COUNT(DISTINCT v.id) FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND MATCH(v.title, v.description) AGAINST(? IN BOOLEAN MODE)"
	fulltextErr := r.db.WithContext(ctx).Raw(countSQL, 1, q).Scan(&total).Error

	// If FULLTEXT fails or returns 0, fallback to LIKE
	if fulltextErr != nil || total == 0 {
		// Use LIKE for count (including tag name match)
		if err := r.db.WithContext(ctx).Raw(
			"SELECT COUNT(DISTINCT v.id) FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND (v.title LIKE ? OR v.description LIKE ? OR t.name LIKE ?)",
			1, likeQ, likeQ, likeQ,
		).Scan(&total).Error; err != nil {
			return nil, 0, fmt.Errorf("search.repository.Search.Count: %w", err)
		}

		if total == 0 {
			return videos, 0, nil
		}

		querySQL := "SELECT DISTINCT v.* FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND (v.title LIKE ? OR v.description LIKE ? OR t.name LIKE ?) ORDER BY v.created_at DESC LIMIT ? OFFSET ?"
		if err := r.db.WithContext(ctx).Raw(querySQL, 1, likeQ, likeQ, likeQ, pageSize, offset).Scan(&videos).Error; err != nil {
			return nil, 0, fmt.Errorf("search.repository.Search.Find: %w", err)
		}
	} else {
		querySQL := "SELECT DISTINCT v.* FROM videos v LEFT JOIN video_tags vt ON v.id = vt.video_id LEFT JOIN tags t ON vt.tag_id = t.id WHERE v.status = ? AND MATCH(v.title, v.description) AGAINST(? IN BOOLEAN MODE) ORDER BY v.created_at DESC LIMIT ? OFFSET ?"
		if err := r.db.WithContext(ctx).Raw(querySQL, 1, q, pageSize, offset).Scan(&videos).Error; err != nil {
			return nil, 0, fmt.Errorf("search.repository.Search.Find: %w", err)
		}
	}

	// Load user for each video
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

func (r *Repository) SearchSuggestions(ctx context.Context, q string, limit int) ([]struct {
	Keyword string
	Count   int64
}, error) {
	type row struct {
		Keyword string
		Count   int64
	}
	var rows []row

	// Combine video title suggestions and tag name suggestions
	likeQ := "%" + q + "%"
	querySQL := `
		(SELECT title AS keyword, views AS count FROM videos WHERE status = ? AND title LIKE ?)
		UNION
		(SELECT DISTINCT t.name COLLATE utf8mb4_unicode_ci AS keyword, 0 AS count FROM tags t WHERE t.name LIKE ?)
		ORDER BY count DESC LIMIT ?
	`
	if err := r.db.WithContext(ctx).Raw(querySQL, 1, likeQ, likeQ, limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("search.repository.SearchSuggestions: %w", err)
	}

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

package admin

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Stats represents dashboard summary statistics.
type Stats struct {
	TotalUsers    int64 `json:"total_users"`
	TotalVideos   int64 `json:"total_videos"`
	TotalViews    int64 `json:"total_views"`
	TotalComments int64 `json:"total_comments"`
	TotalDanmaku  int64 `json:"total_danmaku"`
	TodayNewUsers int64 `json:"today_new_users"`
	TodayNewVideos int64 `json:"today_new_videos"`
}

// UserListItem represents a user row in the admin users table.
type UserListItem struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Role      uint8     `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UsersListResp is the paginated response for admin users list.
type UsersListResp struct {
	Items    []UserListItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

func countTable(db *gorm.DB, ctx context.Context, tableName string) (int64, error) {
	var count int64
	if err := db.WithContext(ctx).Table(tableName).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count %s: %w", tableName, err)
	}
	return count, nil
}

func sumColumn(db *gorm.DB, ctx context.Context, tableName, colName string) (int64, error) {
	var sum int64
	if err := db.WithContext(ctx).Table(tableName).Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", colName)).Scan(&sum).Error; err != nil {
		return 0, fmt.Errorf("sum %s.%s: %w", tableName, colName, err)
	}
	return sum, nil
}

func countToday(db *gorm.DB, ctx context.Context, tableName string) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	if err := db.WithContext(ctx).Table(tableName).
		Where("DATE(created_at) = ?", today).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count today %s: %w", tableName, err)
	}
	return count, nil
}

func getAdminStats(db *gorm.DB, ctx context.Context) (*Stats, error) {
	s := &Stats{}
	var err error

	s.TotalUsers, err = countTable(db, ctx, "users")
	if err != nil {
		return nil, err
	}
	s.TotalVideos, err = countTable(db, ctx, "videos")
	if err != nil {
		return nil, err
	}
	s.TotalViews, err = sumColumn(db, ctx, "videos", "views")
	if err != nil {
		return nil, err
	}
	s.TotalComments, err = countTable(db, ctx, "comments")
	if err != nil {
		return nil, err
	}
	s.TotalDanmaku, err = countTable(db, ctx, "danmaku")
	if err != nil {
		return nil, err
	}
	s.TodayNewUsers, err = countToday(db, ctx, "users")
	if err != nil {
		return nil, err
	}
	s.TodayNewVideos, err = countToday(db, ctx, "videos")
	if err != nil {
		return nil, err
	}

	return s, nil
}

func searchUsers(db *gorm.DB, ctx context.Context, q string, page, pageSize int) ([]UserListItem, int64, error) {
	var items []UserListItem
	var total int64

	query := db.WithContext(ctx).Table("users")

	if q != "" {
		like := "%" + q + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("searchUsers count: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.
		Select("id, username, nickname, avatar, role, created_at").
		Order("id ASC").
		Offset(offset).
		Limit(pageSize).
		Scan(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("searchUsers find: %w", err)
	}

	return items, total, nil
}

func updateUserRole(db *gorm.DB, ctx context.Context, userID uint, role uint8) error {
	result := db.WithContext(ctx).Table("users").
		Where("id = ?", userID).
		Update("role", role)
	if result.Error != nil {
		return fmt.Errorf("updateUserRole: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("updateUserRole: user not found")
	}
	return nil
}

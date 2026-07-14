// Package admin 提供管理后台统计、用户搜索和系统信息查询的辅助函数与数据类型。
package admin

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// ServerStartTime 记录服务器启动时间（在 main.go 中初始化）。
var ServerStartTime = time.Now()

// Stats 代表管理后台仪表盘的聚合统计数据。
type Stats struct {
	TotalUsers    int64 `json:"total_users"`    // 用户总数
	TotalVideos   int64 `json:"total_videos"`   // 视频总数
	TotalViews    int64 `json:"total_views"`    // 总播放量
	TotalComments int64 `json:"total_comments"` // 评论总数
	TotalDanmaku  int64 `json:"total_danmaku"`  // 弹幕总数
	TodayNewUsers int64 `json:"today_new_users"` // 今日新增用户
	TodayNewVideos int64 `json:"today_new_videos"` // 今日新增视频
}

// UserListItem 代表管理后台用户列表中的一行用户数据。
type UserListItem struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Role      uint8     `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UsersListResp 是管理后台用户列表的分页响应结构。
type UsersListResp struct {
	Items    []UserListItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// allowedSumColumns 允许进行 SUM 聚合的列名白名单，防止 SQL 注入。
var allowedSumColumns = map[string]bool{"views": true}

// countTable 统计指定表的记录总数。
func countTable(db *gorm.DB, ctx context.Context, tableName string) (int64, error) {
	var count int64
	if err := db.WithContext(ctx).Table(tableName).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count %s: %w", tableName, err)
	}
	return count, nil
}

// sumColumn 对指定表的指定列进行 SUM 聚合，仅允许白名单中的列名。
func sumColumn(db *gorm.DB, ctx context.Context, tableName, colName string) (int64, error) {
	// 列名白名单校验，防止 SQL 注入
	if !allowedSumColumns[colName] {
		return 0, fmt.Errorf("sumColumn: column %q is not whitelisted", colName)
	}
	var sum int64
	if err := db.WithContext(ctx).Table(tableName).Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", colName)).Scan(&sum).Error; err != nil {
		return 0, fmt.Errorf("sum %s.%s: %w", tableName, colName, err)
	}
	return sum, nil
}

// countToday 统计指定表中当天创建的记录数。
func countToday(db *gorm.DB, ctx context.Context, tableName string) (int64, error) {
	var count int64
	now := time.Now()
	// 计算今天的起始和结束时间
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowStart := todayStart.Add(24 * time.Hour)
	if err := db.WithContext(ctx).Table(tableName).
		Where("created_at >= ? AND created_at < ?", todayStart, tomorrowStart).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count today %s: %w", tableName, err)
	}
	return count, nil
}

// getAdminStats 聚合查询管理后台所需的各项统计数据。
func getAdminStats(db *gorm.DB, ctx context.Context) (*Stats, error) {
	s := &Stats{}
	var err error

	// 逐一查询各项统计数据
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

// searchUsers 按关键词搜索用户，支持分页。
// q 为空时返回全部用户列表。
func searchUsers(db *gorm.DB, ctx context.Context, q string, page, pageSize int) ([]UserListItem, int64, error) {
	var items []UserListItem
	var total int64

	// 构建基础查询
	query := db.WithContext(ctx).Table("users")

	// 带关键词的模糊搜索（匹配用户名或昵称）
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("searchUsers count: %w", err)
	}

	// 分页查询
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

// updateUserRole 更新指定用户的角色值。
// 若用户不存在则返回错误。
func updateUserRole(db *gorm.DB, ctx context.Context, userID uint, role uint8) error {
	result := db.WithContext(ctx).Table("users").
		Where("id = ?", userID).
		Update("role", role)
	if result.Error != nil {
		return fmt.Errorf("updateUserRole: %w", result.Error)
	}
	// 检查是否实际影响了行（防止更新不存在的用户）
	if result.RowsAffected == 0 {
		return fmt.Errorf("updateUserRole: user not found")
	}
	return nil
}

// SystemInfo 代表服务器配置和运行状态信息。
type SystemInfo struct {
	GoVersion   string `json:"go_version"`   // Go 运行时版本
	Uptime      string `json:"uptime"`       // 服务器运行时长
	DBConnected bool   `json:"db_connected"` // 数据库连接是否正常
}

// getSystemInfo 收集并返回服务器运行状态信息。
func getSystemInfo(db *gorm.DB, ctx context.Context) *SystemInfo {
	info := &SystemInfo{
		GoVersion: runtime.Version(),
		Uptime:    time.Since(ServerStartTime).Round(time.Second).String(),
	}
	// 通过 Ping 检测数据库连接状态
	if sqlDB, err := db.DB(); err == nil {
		info.DBConnected = sqlDB.PingContext(ctx) == nil
	}
	return info
}

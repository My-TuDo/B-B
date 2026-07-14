// Package follow 定义用户关注数据模型，包含关注实体与请求/响应结构。
package follow

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

// Follow 关注关系实体，映射 follows 表。
type Follow struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`                                 // 记录主键
	FollowerID  uint      `gorm:"uniqueIndex:uk_follower_following;not null" json:"follower_id"`     // 关注者用户 ID
	FollowingID uint      `gorm:"uniqueIndex:uk_follower_following;not null" json:"following_id"`    // 被关注者用户 ID
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`                                        // 关注时间
}

// TableName 返回 follows 表名。
func (Follow) TableName() string {
	return "follows"
}

// FollowResp 关注操作响应体。
type FollowResp struct {
	Following bool `json:"following"` // 当前用户是否正在关注目标
}

// UserListResp 用户列表响应体（粉丝列表 / 关注列表通用）。
type UserListResp struct {
	Items []usermodel.UserBrief `json:"items"` // 用户简要信息列表
	Total int64                 `json:"total"` // 用户总数
}

// ProfileResp 用户主页响应体，含用户信息与统计数据。
type ProfileResp struct {
	User  *usermodel.UserResp `json:"user"`  // 用户详细信息
	Stats ProfileStats        `json:"stats"` // 主页统计数据
}

// ProfileStats 用户主页统计数据。
type ProfileStats struct {
	Videos    int64 `json:"videos"`    // 视频总数
	Followers int64 `json:"followers"` // 粉丝总数
	Following int64 `json:"following"` // 关注总数
}

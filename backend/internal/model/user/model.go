// Package user 定义用户数据模型，包含用户实体、认证请求/响应以及用户信息 DTO。
package user

import "time"

// ==================== Entity ====================

// User 用户实体，映射 users 表。
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`                // 用户主键
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"` // 用户名（唯一）
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`   // 邮箱（唯一）
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`                   // 密码哈希（不序列化）
	Avatar       string    `gorm:"type:varchar(500);default:''" json:"avatar"`            // 头像 URL
	Nickname     string    `gorm:"type:varchar(50);default:''" json:"nickname"`           // 昵称
	Bio          string    `gorm:"type:varchar(500);default:''" json:"bio"`               // 个人简介
	Role         uint8     `gorm:"type:tinyint unsigned;default:1" json:"role"`           // 角色（1 普通用户 / 2 管理员）
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`                            // 注册时间
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`                            // 更新时间
}

// TableName 返回 users 表名。
func (User) TableName() string {
	return "users"
}

// ==================== DTOs ====================

// RegisterReq 用户注册请求体。
type RegisterReq struct {
	Username string `json:"username" validate:"required,username"` // 用户名
	Email    string `json:"email" validate:"required,email"`       // 邮箱
	Password string `json:"password" validate:"required,password"` // 密码
}

// LoginReq 用户登录请求体。
type LoginReq struct {
	Account  string `json:"account" validate:"required"`  // 账号（用户名或邮箱）
	Password string `json:"password" validate:"required"` // 密码
}

// LoginResp 登录成功响应体。
type LoginResp struct {
	ID       uint   `json:"id"`       // 用户主键
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像 URL
}

// UpdateUserReq 更新用户信息请求体。
type UpdateUserReq struct {
	Nickname *string `json:"nickname" validate:"omitempty,max=50"`   // 昵称
	Avatar   *string `json:"avatar" validate:"omitempty,max=500"`    // 头像 URL
	Bio      *string `json:"bio" validate:"omitempty,max=500"`       // 个人简介
}

// UserResp 用户详细信息响应体。
type UserResp struct {
	ID        uint      `json:"id"`         // 用户主键
	Username  string    `json:"username"`   // 用户名
	Nickname  string    `json:"nickname"`   // 昵称
	Avatar    string    `json:"avatar"`     // 头像 URL
	Bio       string    `json:"bio"`        // 个人简介
	CreatedAt time.Time `json:"created_at"` // 注册时间
}

// UserBrief 用户简要信息，用于列表、关联展示等场景。
type UserBrief struct {
	ID       uint   `json:"id"`       // 用户主键
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像 URL
}

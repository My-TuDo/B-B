package user

import "time"

// ==================== Entity ====================

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Avatar       string    `gorm:"type:varchar(500);default:''" json:"avatar"`
	Nickname     string    `gorm:"type:varchar(50);default:''" json:"nickname"`
	Bio          string    `gorm:"type:varchar(500);default:''" json:"bio"`
	Role         uint8     `gorm:"type:tinyint unsigned;default:1" json:"role"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// ==================== DTOs ====================

type RegisterReq struct {
	Username string `json:"username" validate:"required,username"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type LoginReq struct {
	Account  string `json:"account" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResp struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type UpdateUserReq struct {
	Nickname *string `json:"nickname" validate:"omitempty,max=50"`
	Avatar   *string `json:"avatar" validate:"omitempty,max=500"`
	Bio      *string `json:"bio" validate:"omitempty,max=500"`
}

type UserResp struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

type UserBrief struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

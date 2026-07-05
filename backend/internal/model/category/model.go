package category

import "time"

// ==================== Entity ====================

type Category struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`
	Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (Category) TableName() string {
	return "categories"
}

// ==================== DTOs ====================

type CategoryResp struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

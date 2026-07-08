package follow

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

type Follow struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FollowerID  uint      `gorm:"uniqueIndex:uk_follower_following;not null" json:"follower_id"`
	FollowingID uint      `gorm:"uniqueIndex:uk_follower_following;not null" json:"following_id"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
}

func (Follow) TableName() string {
	return "follows"
}

type FollowResp struct {
	Following bool `json:"following"`
}

type UserListResp struct {
	Items []usermodel.UserBrief `json:"items"`
	Total int64                 `json:"total"`
}

type ProfileResp struct {
	User  *usermodel.UserResp `json:"user"`
	Stats ProfileStats        `json:"stats"`
}

type ProfileStats struct {
	Videos    int64 `json:"videos"`
	Followers int64 `json:"followers"`
	Following int64 `json:"following"`
}

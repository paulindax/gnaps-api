package models

import (
	"gorm.io/gorm"
	"time"
)

// NewsComment model generated from database table 'news_comments'
type NewsComment struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Content *string `json:"content" gorm:"column:content"`
	NewsId *int64 `json:"news_id" gorm:"column:news_id"`
	UserId *int64 `json:"user_id" gorm:"column:user_id"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`
	IsApproved *bool `json:"is_approved" gorm:"column:is_approved"`
}

func (NewsComment) TableName() string {
	return "news_comments"
}

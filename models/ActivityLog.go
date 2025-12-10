package models

import (
	"gorm.io/gorm"
	"time"
)

// ActivityLog model generated from database table 'activity_logs'
type ActivityLog struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	UserId       *int64  `json:"user_id" gorm:"column:user_id"`
	Username     *string `json:"username" gorm:"column:username"`
	Role         *string `json:"role" gorm:"column:role"`
	Type         string  `json:"type" gorm:"column:type"`
	Title        string  `json:"title" gorm:"column:title"`
	Description  *string `json:"description" gorm:"column:description"`
	Method       *string `json:"method" gorm:"column:method"`
	Endpoint     *string `json:"endpoint" gorm:"column:endpoint"`
	StatusCode   *int    `json:"status_code" gorm:"column:status_code"`
	Url          *string `json:"url" gorm:"column:url"`
	ResourceType *string `json:"resource_type" gorm:"column:resource_type"`
	ResourceId   *int64  `json:"resource_id" gorm:"column:resource_id"`
	IpAddress    *string `json:"ip_address" gorm:"column:ip_address"`
	UserAgent    *string `json:"user_agent" gorm:"column:user_agent"`
	OwnerType    *string `json:"owner_type" gorm:"column:owner_type"`
	OwnerId      *int64  `json:"owner_id" gorm:"column:owner_id"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

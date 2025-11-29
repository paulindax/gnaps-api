package models

import (
	"gorm.io/gorm"
	"time"
	"gorm.io/datatypes"
)

// New model generated from database table 'news'
type New struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Title *string `json:"title" gorm:"column:title"`
	Content *string `json:"content" gorm:"column:content"`
	ExecutiveId *int64 `json:"executive_id" gorm:"column:executive_id"`
	RegionIds *datatypes.JSON `json:"region_ids" gorm:"column:region_ids"`
	ZoneIds *datatypes.JSON `json:"zone_ids" gorm:"column:zone_ids"`
	SchoolIds *datatypes.JSON `json:"school_ids" gorm:"column:school_ids"`
	SchoolGroupIds *datatypes.JSON `json:"school_group_ids" gorm:"column:school_group_ids"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`
	Excerpt *string `json:"excerpt" gorm:"column:excerpt"`
	ImageUrl *string `json:"image_url" gorm:"column:image_url"`
	Category *string `json:"category" gorm:"column:category"`
	Status *string `json:"status" gorm:"column:status"`
	Featured *bool `json:"featured" gorm:"column:featured"`
	AuthorId *int64 `json:"author_id" gorm:"column:author_id"`
}

func (New) TableName() string {
	return "news"
}

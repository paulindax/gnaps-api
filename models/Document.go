package models

import (
	"time"

	"gorm.io/gorm"
)

// Document model for document vault templates
type Document struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Title       string  `json:"title" gorm:"column:title"`
	Description *string `json:"description,omitempty" gorm:"column:description"`
	Category    *string `json:"category,omitempty" gorm:"column:category"`
	Status      *string `json:"status,omitempty" gorm:"column:status"` // draft, published, archived

	// JSON field storing the document structure
	TemplateData string `json:"template_data" gorm:"column:template_data;type:json"`

	// Permissions and targeting
	CreatedBy  *int64  `json:"created_by,omitempty" gorm:"column:created_by"`
	IsRequired *bool   `json:"is_required,omitempty" gorm:"column:is_required"`
	RegionIds  *string `json:"region_ids,omitempty" gorm:"column:region_ids;type:json"`
	ZoneIds    *string `json:"zone_ids,omitempty" gorm:"column:zone_ids;type:json"`
	GroupIds   *string `json:"group_ids,omitempty" gorm:"column:group_ids;type:json"`
	SchoolIds  *string `json:"school_ids,omitempty" gorm:"column:school_ids;type:json"`

	// Metadata
	Version   *int  `json:"version,omitempty" gorm:"column:version"`
	IsDeleted *bool `json:"is_deleted,omitempty" gorm:"column:is_deleted"`

	// Computed fields
	SubmissionCount int `json:"submission_count,omitempty" gorm:"-"`
}

func (Document) TableName() string {
	return "documents"
}

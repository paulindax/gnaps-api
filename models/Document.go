package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// Document model generated from database table 'documents'
type Document struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Title        string          `json:"title" gorm:"column:title"`
	Description  *string         `json:"description" gorm:"column:description"`
	Category     *string         `json:"category" gorm:"column:category"`
	Status       *string         `json:"status" gorm:"column:status"`
	TemplateData datatypes.JSON  `json:"template_data" gorm:"column:template_data"`
	CreatedBy    int64           `json:"created_by" gorm:"column:created_by"`
	IsRequired   *bool           `json:"is_required" gorm:"column:is_required"`
	RegionIds    *datatypes.JSON `json:"region_ids" gorm:"column:region_ids"`
	ZoneIds      *datatypes.JSON `json:"zone_ids" gorm:"column:zone_ids"`
	GroupIds     *datatypes.JSON `json:"group_ids" gorm:"column:group_ids"`
	SchoolIds    *datatypes.JSON `json:"school_ids" gorm:"column:school_ids"`
	Version      *int            `json:"version" gorm:"column:version"`
	IsDeleted    *bool           `json:"is_deleted" gorm:"column:is_deleted"`
	OwnerType    *string         `json:"owner_type" gorm:"column:owner_type"`
	OwnerId      *int64          `json:"owner_id" gorm:"column:owner_id"`

	// Transient fields (not in database)
	SubmissionCount int `json:"submission_count,omitempty" gorm:"-"`
}

func (Document) TableName() string {
	return "documents"
}

// SetOwner implements the OwnerFieldSetter interface
func (d *Document) SetOwner(ownerType string, ownerID int64) {
	d.OwnerType = &ownerType
	d.OwnerId = &ownerID
}

package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// DocumentSubmission model generated from database table 'document_submissions'
type DocumentSubmission struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	DocumentId  int64          `json:"document_id" gorm:"column:document_id"`
	SchoolId    int64          `json:"school_id" gorm:"column:school_id"`
	SubmittedBy int64          `json:"submitted_by" gorm:"column:submitted_by"`
	FormData    datatypes.JSON `json:"form_data" gorm:"column:form_data"`
	Status      *string        `json:"status" gorm:"column:status"`
	SubmittedAt time.Time      `json:"submitted_at" gorm:"column:submitted_at"`
	ReviewedAt  time.Time      `json:"reviewed_at" gorm:"column:reviewed_at"`
	ReviewedBy  *int64         `json:"reviewed_by" gorm:"column:reviewed_by"`
	ReviewNotes *string        `json:"review_notes" gorm:"column:review_notes"`
	IsDeleted   *bool          `json:"is_deleted" gorm:"column:is_deleted"`
	OwnerType   *string        `json:"owner_type" gorm:"column:owner_type"`
	OwnerId     *int64         `json:"owner_id" gorm:"column:owner_id"`

	// Transient fields (not in database)
	DocumentTitle *string `json:"document_title,omitempty" gorm:"-"`
	SchoolName    *string `json:"school_name,omitempty" gorm:"-"`
	SubmitterName *string `json:"submitter_name,omitempty" gorm:"-"`
}

func (DocumentSubmission) TableName() string {
	return "document_submissions"
}

// SetOwner implements the OwnerFieldSetter interface
func (d *DocumentSubmission) SetOwner(ownerType string, ownerID int64) {
	d.OwnerType = &ownerType
	d.OwnerId = &ownerID
}

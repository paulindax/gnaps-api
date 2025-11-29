package models

import (
	"time"

	"gorm.io/gorm"
)

// DocumentSubmission model for filled document forms
type DocumentSubmission struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	DocumentId  *int64 `json:"document_id,omitempty" gorm:"column:document_id"`
	SchoolId    *int64 `json:"school_id,omitempty" gorm:"column:school_id"`
	SubmittedBy *int64 `json:"submitted_by,omitempty" gorm:"column:submitted_by"`

	// JSON field storing the filled form data
	FormData string `json:"form_data" gorm:"column:form_data;type:json"`

	// Status tracking
	Status      *string `json:"status,omitempty" gorm:"column:status"` // draft, submitted, reviewed, approved, rejected
	SubmittedAt *string `json:"submitted_at,omitempty" gorm:"column:submitted_at"`
	ReviewedAt  *string `json:"reviewed_at,omitempty" gorm:"column:reviewed_at"`
	ReviewedBy  *int64  `json:"reviewed_by,omitempty" gorm:"column:reviewed_by"`
	ReviewNotes *string `json:"review_notes,omitempty" gorm:"column:review_notes"`

	// Metadata
	IsDeleted *bool `json:"is_deleted,omitempty" gorm:"column:is_deleted"`

	// Computed fields
	DocumentTitle *string `json:"document_title,omitempty" gorm:"-"`
	SchoolName    *string `json:"school_name,omitempty" gorm:"-"`
	SubmitterName *string `json:"submitter_name,omitempty" gorm:"-"`
}

func (DocumentSubmission) TableName() string {
	return "document_submissions"
}

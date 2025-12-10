package models

import (
	"time"
)

// DocumentAttachment model generated from database table 'document_attachments'
type DocumentAttachment struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time `json:"created_at"`
	SubmissionId int64     `json:"submission_id" gorm:"column:submission_id"`
	FieldName    string    `json:"field_name" gorm:"column:field_name"`

	FileUrl   string  `json:"file_url" gorm:"column:file_url"`
	FileName  string  `json:"file_name" gorm:"column:file_name"`
	FileType  *string `json:"file_type" gorm:"column:file_type"`
	FileSize  *int64  `json:"file_size" gorm:"column:file_size"`
	OwnerType *string `json:"owner_type" gorm:"column:owner_type"`
	OwnerId   *int64  `json:"owner_id" gorm:"column:owner_id"`
}

func (DocumentAttachment) TableName() string {
	return "document_attachments"
}

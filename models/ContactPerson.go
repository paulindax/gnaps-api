package models

import (
	"gorm.io/gorm"
	"time"
)

// ContactPerson model generated from database table 'contact_persons'
type ContactPerson struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	SchoolId  *int64  `json:"school_id" gorm:"column:school_id"`
	FirstName *string `json:"first_name" gorm:"column:first_name"`
	LastName  *string `json:"last_name" gorm:"column:last_name"`
	Relation  *string `json:"relation" gorm:"column:relation"`
	Email     *string `json:"email" gorm:"column:email"`
	MobileNo  *string `json:"mobile_no" gorm:"column:mobile_no"`
}

func (ContactPerson) TableName() string {
	return "contact_persons"
}

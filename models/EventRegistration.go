package models

import (
	"gorm.io/gorm"
	"time"
)

// EventRegistration model generated from database table 'event_registrations'
type EventRegistration struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	EventId           *int64  `json:"event_id,omitempty" gorm:"column:event_id"`
	SchoolId          *int64  `json:"school_id,omitempty" gorm:"column:school_id"`
	RegisteredBy      *int64  `json:"registered_by,omitempty" gorm:"column:registered_by"`
	PaymentStatus     *string `json:"payment_status,omitempty" gorm:"column:payment_status"`
	PaymentReference  *string `json:"payment_reference,omitempty" gorm:"column:payment_reference"`
	PaymentMethod     *string `json:"payment_method,omitempty" gorm:"column:payment_method"`
	PaymentPhone      *string `json:"payment_phone,omitempty" gorm:"column:payment_phone"`
	RegistrationDate  *string `json:"registration_date,omitempty" gorm:"column:registration_date"`
	NumberOfAttendees *int    `json:"number_of_attendees,omitempty" gorm:"column:number_of_attendees"`
	IsDeleted         *bool   `json:"is_deleted,omitempty" gorm:"column:is_deleted"`

	// Virtual fields for joined data
	SchoolName *string `json:"school_name,omitempty" gorm:"-"`
	EventTitle *string `json:"event_title,omitempty" gorm:"-"`
}

func (EventRegistration) TableName() string {
	return "event_registrations"
}

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

	EventId int64 `json:"event_id" gorm:"column:event_id"`
	SchoolId int64 `json:"school_id" gorm:"column:school_id"`
	RegisteredBy *int64 `json:"registered_by,omitempty" gorm:"column:registered_by"`
	PaymentStatus *string `json:"payment_status" gorm:"column:payment_status"`
	PaymentReference *string `json:"payment_reference" gorm:"column:payment_reference"`
	PaymentMethod *string `json:"payment_method" gorm:"column:payment_method"`
	PaymentPhone *string `json:"payment_phone" gorm:"column:payment_phone"`
	RegistrationDate time.Time `json:"registration_date" gorm:"column:registration_date"`
	NumberOfAttendees *int `json:"number_of_attendees" gorm:"column:number_of_attendees"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`

	// Transient fields (not in database)
	EventTitle *string `json:"event_title,omitempty" gorm:"-"`
	SchoolName *string `json:"school_name,omitempty" gorm:"-"`
}

func (EventRegistration) TableName() string {
	return "event_registrations"
}

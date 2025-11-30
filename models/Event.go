package models

import (
	"time"

	"gorm.io/gorm"
)

// Event model generated from database table 'events'
type Event struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Title                string   `json:"title" gorm:"column:title"`
	Description          *string  `json:"description,omitempty" gorm:"column:description"`
	StartDate            string   `json:"start_date" gorm:"column:start_date"`
	EndDate              *string  `json:"end_date,omitempty" gorm:"column:end_date"`
	CreatedBy            *int64   `json:"created_by,omitempty" gorm:"column:created_by"`
	Location             *string  `json:"location,omitempty" gorm:"column:location"`
	Venue                *string  `json:"venue,omitempty" gorm:"column:venue"`
	IsPaid               *bool    `json:"is_paid,omitempty" gorm:"column:is_paid"`
	Price                *float64 `json:"price,omitempty" gorm:"column:price"`
	MaxAttendees         *int     `json:"max_attendees,omitempty" gorm:"column:max_attendees"`
	RegistrationDeadline *string  `json:"registration_deadline,omitempty" gorm:"column:registration_deadline"`
	Status               *string  `json:"status,omitempty" gorm:"column:status"`
	ImageUrl             *string  `json:"image_url,omitempty" gorm:"column:image_url"`
	RegistrationCode     *string  `json:"registration_code,omitempty" gorm:"column:registration_code;unique"`
	IsDeleted            *bool    `json:"is_deleted,omitempty" gorm:"column:is_deleted"`
	RegisteredCount      int      `json:"registered_count" gorm:"-"`
}

func (Event) TableName() string {
	return "events"
}

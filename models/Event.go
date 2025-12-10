package models

import (
	"gorm.io/gorm"
	"time"
)

// Event model generated from database table 'events'
type Event struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Title                string    `json:"title" gorm:"column:title"`
	Description          *string   `json:"description" gorm:"column:description"`
	StartDate            time.Time `json:"start_date" gorm:"column:start_date"`
	EndDate              time.Time `json:"end_date" gorm:"column:end_date"`
	OrganizationId       *int64    `json:"organization_id" gorm:"column:organization_id"`
	CreatedBy            int64     `json:"created_by" gorm:"column:created_by"`
	Location             *string   `json:"location" gorm:"column:location"`
	Venue                *string   `json:"venue" gorm:"column:venue"`
	IsPaid               *bool     `json:"is_paid" gorm:"column:is_paid"`
	Price                *float64  `json:"price" gorm:"column:price"`
	MaxAttendees         *int      `json:"max_attendees" gorm:"column:max_attendees"`
	RegistrationDeadline time.Time `json:"registration_deadline" gorm:"column:registration_deadline"`
	Status               *string   `json:"status" gorm:"column:status"`
	ImageUrl             *string   `json:"image_url" gorm:"column:image_url"`
	RegistrationCode     *string   `json:"registration_code" gorm:"column:registration_code"`
	IsDeleted            *bool     `json:"is_deleted" gorm:"column:is_deleted"`
	BillId               *int64    `json:"bill_id" gorm:"column:bill_id"`
	OwnerType            *string   `json:"owner_type" gorm:"column:owner_type"`
	OwnerId              *int64    `json:"owner_id" gorm:"column:owner_id"`

	// Transient fields (not in database)
	RegisteredCount int     `json:"registered_count,omitempty" gorm:"-"`
	BillName        *string `json:"bill_name,omitempty" gorm:"-"`
}

func (Event) TableName() string {
	return "events"
}

// SetOwner implements the OwnerFieldSetter interface
func (e *Event) SetOwner(ownerType string, ownerID int64) {
	e.OwnerType = &ownerType
	e.OwnerId = &ownerID
}

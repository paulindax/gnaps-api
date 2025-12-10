package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// School model generated from database table 'schools'
type School struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	MemberNo            string          `json:"member_no" gorm:"column:member_no"`
	JoiningDate         time.Time       `json:"joining_date" gorm:"column:joining_date"`
	Name                string          `json:"name" gorm:"column:name"`
	ZoneId              *int64          `json:"zone_id" gorm:"column:zone_id"`
	DateOfEstablishment time.Time       `json:"date_of_establishment" gorm:"column:date_of_establishment"`
	Address             *string         `json:"address" gorm:"column:address"`
	Location            *string         `json:"location" gorm:"column:location"`
	MobileNo            *string         `json:"mobile_no" gorm:"column:mobile_no"`
	Email               *string         `json:"email" gorm:"column:email"`
	GpsAddress          *string         `json:"gps_address" gorm:"column:gps_address"`
	IsDeleted           *bool           `json:"is_deleted" gorm:"column:is_deleted"`
	UserId              *int64          `json:"user_id" gorm:"column:user_id"`
	SchoolGroupIds      *datatypes.JSON `json:"school_group_ids" gorm:"column:school_group_ids"`

	// Transient fields (not in database)
	Zone           *Zone           `json:"zone,omitempty" gorm:"foreignKey:ZoneId"`
	ContactPersons []ContactPerson `json:"contact_persons,omitempty" gorm:"foreignKey:SchoolId"`
}

func (School) TableName() string {
	return "schools"
}

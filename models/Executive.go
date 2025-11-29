package models

import (
	"gorm.io/gorm"
	"time"
	"gorm.io/datatypes"
)

// Executive model generated from database table 'executives'
type Executive struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	ExecutiveNo *string `json:"executive_no" gorm:"column:executive_no"`
	FirstName *string `json:"first_name" gorm:"column:first_name"`
	MiddleName *string `json:"middle_name" gorm:"column:middle_name"`
	LastName *string `json:"last_name" gorm:"column:last_name"`
	Gender *string `json:"gender" gorm:"column:gender"`
	PositionId *int64 `json:"position_id" gorm:"column:position_id"`
	DateOfBirth time.Time `json:"date_of_birth" gorm:"column:date_of_birth"`
	MobileNo *string `json:"mobile_no" gorm:"column:mobile_no"`
	Email *string `json:"email" gorm:"column:email"`
	PhotoFileName *string `json:"photo_file_name" gorm:"column:photo_file_name"`
	UserId *int64 `json:"user_id" gorm:"column:user_id"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`
	AssignedZoneIds *datatypes.JSON `json:"assigned_zone_ids" gorm:"column:assigned_zone_ids"`
	AssignedRegionsIds *datatypes.JSON `json:"assigned_regions_ids" gorm:"column:assigned_regions_ids"`
}

func (Executive) TableName() string {
	return "executives"
}

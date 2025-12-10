package models

import (
	"gorm.io/gorm"
	"time"
)

// Executive model generated from database table 'executives'
type Executive struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	ExecutiveNo *string   `json:"executive_no" gorm:"column:executive_no"`
	FirstName   *string   `json:"first_name" gorm:"column:first_name"`
	MiddleName  *string   `json:"middle_name" gorm:"column:middle_name"`
	LastName    *string   `json:"last_name" gorm:"column:last_name"`
	Gender      *string   `json:"gender" gorm:"column:gender"`
	PositionId  *int64    `json:"position_id" gorm:"column:position_id"`
	DateOfBirth time.Time `json:"date_of_birth" gorm:"column:date_of_birth"`
	MobileNo    *string   `json:"mobile_no" gorm:"column:mobile_no"`
	Email       *string   `json:"email" gorm:"column:email"`
	ImageUrl    *string   `json:"image_url" gorm:"column:image_url"`
	UserId      *int64    `json:"user_id" gorm:"column:user_id"`
	IsDeleted   *bool     `json:"is_deleted" gorm:"column:is_deleted"`
	Role        *string   `json:"role" gorm:"column:role"`
	RegionId    *int64    `json:"region_id" gorm:"column:region_id"`
	ZoneId      *int64    `json:"zone_id" gorm:"column:zone_id"`
	Status      *string   `json:"status" gorm:"column:status"`
	Bio         *string   `json:"bio" gorm:"column:bio"`

	// Transient fields (not in database)
	PositionName *string `json:"position_name,omitempty" gorm:"-"`
	RegionName   *string `json:"region_name,omitempty" gorm:"-"`
	ZoneName     *string `json:"zone_name,omitempty" gorm:"-"`
}

func (Executive) TableName() string {
	return "executives"
}

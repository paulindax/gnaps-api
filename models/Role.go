package models

import (
	"gorm.io/gorm"
	"time"
)

// Role model generated from database table 'roles'
type Role struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name        *string `json:"name" gorm:"column:name"`
	NameTag     *[]byte `json:"name_tag" gorm:"column:name_tag"`
	Description *string `json:"description" gorm:"column:description"`
	AuthKey     *string `json:"auth_key" gorm:"column:auth_key"`
}

func (Role) TableName() string {
	return "roles"
}

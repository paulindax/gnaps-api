package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// User model generated from database table 'users'
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Username            *string         `json:"username" gorm:"column:username"`
	FirstName           *string         `json:"first_name" gorm:"column:first_name"`
	LastName            *string         `json:"last_name" gorm:"column:last_name"`
	Role                *string         `json:"role" gorm:"column:role"`
	Email               *string         `json:"email" gorm:"column:email"`
	MobileNo            *string         `json:"mobile_no" gorm:"column:mobile_no"`
	IsFirstLogin        *bool           `json:"is_first_login" gorm:"column:is_first_login"`
	IsDeleted           *bool           `json:"is_deleted" gorm:"column:is_deleted"`
	EncryptedPassword   *string         `json:"encrypted_password" gorm:"column:encrypted_password"`
	ResetPasswordToken  *string         `json:"reset_password_token" gorm:"column:reset_password_token"`
	ResetPasswordSentAt time.Time       `json:"reset_password_sent_at" gorm:"column:reset_password_sent_at"`
	RememberCreatedAt   time.Time       `json:"remember_created_at" gorm:"column:remember_created_at"`
	SignInCount         *int            `json:"sign_in_count" gorm:"column:sign_in_count"`
	CurrentSignInAt     time.Time       `json:"current_sign_in_at" gorm:"column:current_sign_in_at"`
	LastSignInAt        time.Time       `json:"last_sign_in_at" gorm:"column:last_sign_in_at"`
	CurrentSignInIp     *string         `json:"current_sign_in_ip" gorm:"column:current_sign_in_ip"`
	LastSignInIp        *string         `json:"last_sign_in_ip" gorm:"column:last_sign_in_ip"`
	AuthyId             *string         `json:"authy_id" gorm:"column:authy_id"`
	AuthyEnabled        *bool           `json:"authy_enabled" gorm:"column:authy_enabled"`
	ScreenLocked        *bool           `json:"screen_locked" gorm:"column:screen_locked"`
	OtpSecretKey        *string         `json:"otp_secret_key" gorm:"column:otp_secret_key"`
	ForceReload         *bool           `json:"force_reload" gorm:"column:force_reload"`
	UserProperties      *datatypes.JSON `json:"user_properties" gorm:"column:user_properties"`
}

func (User) TableName() string {
	return "users"
}

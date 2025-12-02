package models

import (
	"gorm.io/gorm"
	"time"
)

// FinanceAccount model generated from database table 'finance_accounts'
type FinanceAccount struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name *string `json:"name" gorm:"column:name"`
	Code *string `json:"code" gorm:"column:code"`
	Description *string `json:"description" gorm:"column:description"`
	AccountType *string `json:"account_type" gorm:"column:account_type"`
	IsIncome *bool `json:"is_income" gorm:"column:is_income"`
	IsDeleted bool `json:"is_deleted" gorm:"column:is_deleted"`
	ApproverId *int64 `json:"approver_id" gorm:"column:approver_id"`
}

func (FinanceAccount) TableName() string {
	return "finance_accounts"
}

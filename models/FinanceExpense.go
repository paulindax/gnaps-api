package models

import (
	"time"
)

// FinanceExpense model generated from database table 'finance_expenses'
type FinanceExpense struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Title *string `json:"title" gorm:"column:title"`

	Description *string `json:"description" gorm:"column:description"`
	Amount *float64 `json:"amount" gorm:"column:amount"`
	BudgetAccountId *int64 `json:"budget_account_id" gorm:"column:budget_account_id"`
	TransactionDate time.Time `json:"transaction_date" gorm:"column:transaction_date"`
	VoucherNo *string `json:"voucher_no" gorm:"column:voucher_no"`
	Status *string `json:"status" gorm:"column:status"`
	IsPaid *bool `json:"is_paid" gorm:"column:is_paid"`
	IsApproved *bool `json:"is_approved" gorm:"column:is_approved"`
	ApprovedBy *int64 `json:"approved_by" gorm:"column:approved_by"`
	CategoryId *int64 `json:"category_id" gorm:"column:category_id"`
	RejectedReason *string `json:"rejected_reason" gorm:"column:rejected_reason"`
	RejectedBy *int64 `json:"rejected_by" gorm:"column:rejected_by"`
	UserId *int64 `json:"user_id" gorm:"column:user_id"`
	BankFieldId *int64 `json:"bank_field_id" gorm:"column:bank_field_id"`
	BankAccountId *int64 `json:"bank_account_id" gorm:"column:bank_account_id"`
	ChequeNo *string `json:"cheque_no" gorm:"column:cheque_no"`
	FollowUp *bool `json:"follow_up" gorm:"column:follow_up"`
}

func (FinanceExpense) TableName() string {
	return "finance_expenses"
}

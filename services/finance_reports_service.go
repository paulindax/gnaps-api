package services

import (
	"gnaps-api/models"
	"time"

	"gorm.io/gorm"
)

type FinanceReportsService struct {
	db *gorm.DB
}

func NewFinanceReportsService(db *gorm.DB) *FinanceReportsService {
	return &FinanceReportsService{db: db}
}

type MomoPaymentFilters struct {
	Status      string
	SchoolID    int64
	MomoNetwork string
	FromDate    string
	ToDate      string
}

type FinanceTransactionFilters struct {
	SchoolID         int64
	FinanceAccountID int64
	FinanceType      string
	FromDate         string
	ToDate           string
}

type MomoPaymentWithSchool struct {
	models.MomoPayment
	SchoolName string `json:"school_name"`
}

type FinanceTransactionWithDetails struct {
	models.FinanceTransaction
	SchoolName         string `json:"school_name"`
	FinanceAccountName string `json:"finance_account_name"`
}

type MomoPaymentStats struct {
	Total       int64   `json:"total"`
	Successful  int64   `json:"successful"`
	Pending     int64   `json:"pending"`
	Failed      int64   `json:"failed"`
	TotalAmount float64 `json:"total_amount"`
}

type FinanceTransactionStats struct {
	Total        int64   `json:"total"`
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
}

func (s *FinanceReportsService) GetMomoPayments(filters MomoPaymentFilters, page, limit int) ([]MomoPaymentWithSchool, int64, error) {
	var payments []MomoPaymentWithSchool
	var total int64

	query := s.db.Table("momo_payments").
		Select("momo_payments.*, schools.name as school_name").
		Joins("LEFT JOIN schools ON schools.id = momo_payments.school_id").
		Where("momo_payments.is_deleted IS NULL OR momo_payments.is_deleted = ?", false)

	// Apply filters
	if filters.Status != "" {
		query = query.Where("momo_payments.status = ?", filters.Status)
	}
	if filters.SchoolID > 0 {
		query = query.Where("momo_payments.school_id = ?", filters.SchoolID)
	}
	if filters.MomoNetwork != "" {
		query = query.Where("momo_payments.momo_network = ?", filters.MomoNetwork)
	}
	if filters.FromDate != "" {
		fromTime, err := time.Parse("2006-01-02", filters.FromDate)
		if err == nil {
			query = query.Where("momo_payments.created_at >= ?", fromTime)
		}
	}
	if filters.ToDate != "" {
		toTime, err := time.Parse("2006-01-02", filters.ToDate)
		if err == nil {
			// Add 1 day to include the end date
			toTime = toTime.Add(24 * time.Hour)
			query = query.Where("momo_payments.created_at < ?", toTime)
		}
	}

	// Get total count
	countQuery := s.db.Table("momo_payments").
		Where("momo_payments.is_deleted IS NULL OR momo_payments.is_deleted = ?", false)

	if filters.Status != "" {
		countQuery = countQuery.Where("momo_payments.status = ?", filters.Status)
	}
	if filters.SchoolID > 0 {
		countQuery = countQuery.Where("momo_payments.school_id = ?", filters.SchoolID)
	}
	if filters.MomoNetwork != "" {
		countQuery = countQuery.Where("momo_payments.momo_network = ?", filters.MomoNetwork)
	}
	if filters.FromDate != "" {
		fromTime, err := time.Parse("2006-01-02", filters.FromDate)
		if err == nil {
			countQuery = countQuery.Where("momo_payments.created_at >= ?", fromTime)
		}
	}
	if filters.ToDate != "" {
		toTime, err := time.Parse("2006-01-02", filters.ToDate)
		if err == nil {
			toTime = toTime.Add(24 * time.Hour)
			countQuery = countQuery.Where("momo_payments.created_at < ?", toTime)
		}
	}
	countQuery.Count(&total)

	// Pagination
	offset := (page - 1) * limit
	err := query.Order("momo_payments.created_at DESC").Offset(offset).Limit(limit).Scan(&payments).Error

	return payments, total, err
}

func (s *FinanceReportsService) GetMomoPaymentStats() MomoPaymentStats {
	var stats MomoPaymentStats

	s.db.Table("momo_payments").
		Where("is_deleted IS NULL OR is_deleted = ?", false).
		Count(&stats.Total)

	s.db.Table("momo_payments").
		Where("(is_deleted IS NULL OR is_deleted = ?) AND status = ?", false, "success").
		Count(&stats.Successful)

	s.db.Table("momo_payments").
		Where("(is_deleted IS NULL OR is_deleted = ?) AND status IN ?", false, []string{"pending", "created"}).
		Count(&stats.Pending)

	s.db.Table("momo_payments").
		Where("(is_deleted IS NULL OR is_deleted = ?) AND status = ?", false, "failed").
		Count(&stats.Failed)

	var totalAmount *float64
	s.db.Table("momo_payments").
		Where("(is_deleted IS NULL OR is_deleted = ?) AND status = ?", false, "success").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount)

	if totalAmount != nil {
		stats.TotalAmount = *totalAmount
	}

	return stats
}

func (s *FinanceReportsService) GetFinanceTransactions(filters FinanceTransactionFilters, page, limit int) ([]FinanceTransactionWithDetails, int64, error) {
	var transactions []FinanceTransactionWithDetails
	var total int64

	query := s.db.Table("finance_transactions").
		Select("finance_transactions.*, schools.name as school_name, finance_accounts.name as finance_account_name").
		Joins("LEFT JOIN schools ON schools.id = finance_transactions.school_id").
		Joins("LEFT JOIN finance_accounts ON finance_accounts.id = finance_transactions.finance_account_id")

	// Apply filters
	if filters.SchoolID > 0 {
		query = query.Where("finance_transactions.school_id = ?", filters.SchoolID)
	}
	if filters.FinanceAccountID > 0 {
		query = query.Where("finance_transactions.finance_account_id = ?", filters.FinanceAccountID)
	}
	if filters.FinanceType != "" {
		query = query.Where("finance_transactions.finance_type = ?", filters.FinanceType)
	}
	if filters.FromDate != "" {
		fromTime, err := time.Parse("2006-01-02", filters.FromDate)
		if err == nil {
			query = query.Where("finance_transactions.transaction_date >= ?", fromTime)
		}
	}
	if filters.ToDate != "" {
		toTime, err := time.Parse("2006-01-02", filters.ToDate)
		if err == nil {
			toTime = toTime.Add(24 * time.Hour)
			query = query.Where("finance_transactions.transaction_date < ?", toTime)
		}
	}

	// Get total count
	countQuery := s.db.Table("finance_transactions")
	if filters.SchoolID > 0 {
		countQuery = countQuery.Where("school_id = ?", filters.SchoolID)
	}
	if filters.FinanceAccountID > 0 {
		countQuery = countQuery.Where("finance_account_id = ?", filters.FinanceAccountID)
	}
	if filters.FinanceType != "" {
		countQuery = countQuery.Where("finance_type = ?", filters.FinanceType)
	}
	if filters.FromDate != "" {
		fromTime, err := time.Parse("2006-01-02", filters.FromDate)
		if err == nil {
			countQuery = countQuery.Where("transaction_date >= ?", fromTime)
		}
	}
	if filters.ToDate != "" {
		toTime, err := time.Parse("2006-01-02", filters.ToDate)
		if err == nil {
			toTime = toTime.Add(24 * time.Hour)
			countQuery = countQuery.Where("transaction_date < ?", toTime)
		}
	}
	countQuery.Count(&total)

	// Pagination
	offset := (page - 1) * limit
	err := query.Order("finance_transactions.transaction_date DESC").Offset(offset).Limit(limit).Scan(&transactions).Error

	return transactions, total, err
}

func (s *FinanceReportsService) GetFinanceTransactionStats() FinanceTransactionStats {
	var stats FinanceTransactionStats

	s.db.Table("finance_transactions").Count(&stats.Total)

	// Calculate income (finance_type = 'income' or based on finance_account.is_income)
	var totalIncome *float64
	s.db.Table("finance_transactions").
		Joins("LEFT JOIN finance_accounts ON finance_accounts.id = finance_transactions.finance_account_id").
		Where("finance_accounts.is_income = ?", true).
		Select("COALESCE(SUM(finance_transactions.amount), 0)").
		Scan(&totalIncome)

	if totalIncome != nil {
		stats.TotalIncome = *totalIncome
	}

	// Calculate expense (finance_type = 'expense' or based on finance_account.is_income = false)
	var totalExpense *float64
	s.db.Table("finance_transactions").
		Joins("LEFT JOIN finance_accounts ON finance_accounts.id = finance_transactions.finance_account_id").
		Where("finance_accounts.is_income = ?", false).
		Select("COALESCE(SUM(finance_transactions.amount), 0)").
		Scan(&totalExpense)

	if totalExpense != nil {
		stats.TotalExpense = *totalExpense
	}

	return stats
}

package controllers

import (
	"fmt"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type FinanceReportsController struct {
	financeReportsService *services.FinanceReportsService
}

func NewFinanceReportsController(financeReportsService *services.FinanceReportsService) *FinanceReportsController {
	return &FinanceReportsController{
		financeReportsService: financeReportsService,
	}
}

func (c *FinanceReportsController) Handle(action string, ctx *fiber.Ctx) error {
	switch action {
	case "momo-payments":
		return c.getMomoPayments(ctx)
	case "momo-payments-stats":
		return c.getMomoPaymentStats(ctx)
	case "transactions":
		return c.getFinanceTransactions(ctx)
	case "transactions-stats":
		return c.getFinanceTransactionStats(ctx)
	default:
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

func (c *FinanceReportsController) getMomoPayments(ctx *fiber.Ctx) error {
	// Check user role - only system_admin and national_admin can access
	role, ok := ctx.Locals("role").(string)
	if !ok || (role != "system_admin" && role != "national_admin") {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied. Only system admins and national admins can view finance reports.",
		})
	}

	// Parse query parameters
	filters := services.MomoPaymentFilters{
		Status:      ctx.Query("status"),
		MomoNetwork: ctx.Query("momo_network"),
		FromDate:    ctx.Query("from_date"),
		ToDate:      ctx.Query("to_date"),
	}

	if schoolID := ctx.Query("school_id"); schoolID != "" {
		if id, err := strconv.ParseInt(schoolID, 10, 64); err == nil {
			filters.SchoolID = id
		}
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	payments, total, err := c.financeReportsService.GetMomoPayments(filters, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch momo payments: %v", err),
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return ctx.JSON(fiber.Map{
		"data":        payments,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (c *FinanceReportsController) getMomoPaymentStats(ctx *fiber.Ctx) error {
	// Check user role - only system_admin and national_admin can access
	role, ok := ctx.Locals("role").(string)
	if !ok || (role != "system_admin" && role != "national_admin") {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied. Only system admins and national admins can view finance reports.",
		})
	}

	stats := c.financeReportsService.GetMomoPaymentStats()
	return ctx.JSON(stats)
}

func (c *FinanceReportsController) getFinanceTransactions(ctx *fiber.Ctx) error {
	// Check user role - only system_admin and national_admin can access
	role, ok := ctx.Locals("role").(string)
	if !ok || (role != "system_admin" && role != "national_admin") {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied. Only system admins and national admins can view finance reports.",
		})
	}

	// Parse query parameters
	filters := services.FinanceTransactionFilters{
		FinanceType: ctx.Query("finance_type"),
		FromDate:    ctx.Query("from_date"),
		ToDate:      ctx.Query("to_date"),
	}

	if schoolID := ctx.Query("school_id"); schoolID != "" {
		if id, err := strconv.ParseInt(schoolID, 10, 64); err == nil {
			filters.SchoolID = id
		}
	}

	if accountID := ctx.Query("finance_account_id"); accountID != "" {
		if id, err := strconv.ParseInt(accountID, 10, 64); err == nil {
			filters.FinanceAccountID = id
		}
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	transactions, total, err := c.financeReportsService.GetFinanceTransactions(filters, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch finance transactions: %v", err),
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return ctx.JSON(fiber.Map{
		"data":        transactions,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (c *FinanceReportsController) getFinanceTransactionStats(ctx *fiber.Ctx) error {
	// Check user role - only system_admin and national_admin can access
	role, ok := ctx.Locals("role").(string)
	if !ok || (role != "system_admin" && role != "national_admin") {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied. Only system admins and national admins can view finance reports.",
		})
	}

	stats := c.financeReportsService.GetFinanceTransactionStats()
	return ctx.JSON(stats)
}

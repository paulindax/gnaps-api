package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type FinanceAccountsController struct {
	accountService *services.FinanceAccountService
}

func NewFinanceAccountsController(accountService *services.FinanceAccountService) *FinanceAccountsController {
	return &FinanceAccountsController{
		accountService: accountService,
	}
}

func (f *FinanceAccountsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return f.list(c)
	case "show":
		return f.show(c)
	case "create":
		return f.create(c)
	case "update":
		return f.update(c)
	case "delete":
		return f.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (f *FinanceAccountsController) list(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if code := c.Query("code"); code != "" {
		filters["code"] = code
	}
	if accountType := c.Query("account_type"); accountType != "" {
		filters["account_type"] = accountType
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	accounts, total, err := f.accountService.ListAccounts(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve finance accounts",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve finance accounts",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": accounts,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (f *FinanceAccountsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	accountId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	account, err := f.accountService.GetAccountByID(uint(accountId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": account})
}

func (f *FinanceAccountsController) create(c *fiber.Ctx) error {
	var account models.FinanceAccount
	if err := c.BodyParser(&account); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := f.accountService.CreateAccount(&account); err != nil {
		if err.Error() == "finance account with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Finance account created successfully",
		"flash_message": fiber.Map{
			"msg":  "Finance account created successfully",
			"type": "success",
		},
		"data": account,
	})
}

func (f *FinanceAccountsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	accountId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.FinanceAccount
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Build updates map
	updates := make(map[string]interface{})
	if updateData.Name != nil {
		updates["name"] = *updateData.Name
	}
	if updateData.Code != nil {
		updates["code"] = *updateData.Code
	}
	if updateData.Description != nil {
		updates["description"] = *updateData.Description
	}
	if updateData.AccountType != nil {
		updates["account_type"] = *updateData.AccountType
	}
	if updateData.IsIncome != nil {
		updates["is_income"] = *updateData.IsIncome
	}
	if updateData.ApproverId != nil {
		updates["approver_id"] = *updateData.ApproverId
	}

	if err := f.accountService.UpdateAccount(uint(accountId), updates); err != nil {
		if err.Error() == "finance account not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "finance account with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated account
	account, _ := f.accountService.GetAccountByID(uint(accountId))

	return c.JSON(fiber.Map{
		"message": "Finance account updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Finance account updated successfully",
			"type": "success",
		},
		"data": account,
	})
}

func (f *FinanceAccountsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	accountId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := f.accountService.DeleteAccount(uint(accountId)); err != nil {
		if err.Error() == "finance account not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Finance account deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Finance account deleted successfully",
			"type": "success",
		},
	})
}

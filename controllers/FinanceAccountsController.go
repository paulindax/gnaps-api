package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
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
	// Owner-based actions
	case "ownerList":
		return f.ownerList(c)
	case "ownerShow":
		return f.ownerShow(c)
	case "ownerCreate":
		return f.ownerCreate(c)
	case "ownerUpdate":
		return f.ownerUpdate(c)
	case "ownerDelete":
		return f.ownerDelete(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (f *FinanceAccountsController) list(c *fiber.Ctx) error {
	// Get owner context for filtering
	ownerCtx := utils.GetOwnerContext(c)

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

	accounts, total, err := f.accountService.ListAccountsWithOwner(filters, page, limit, ownerCtx)
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
	ownerCtx := utils.GetOwnerContext(c)

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

	account, err := f.accountService.GetAccountByIDWithOwner(uint(accountId), ownerCtx)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Finance account not found or access denied"})
	}

	return c.JSON(fiber.Map{"data": account})
}

func (f *FinanceAccountsController) create(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	var account models.FinanceAccount
	if err := c.BodyParser(&account); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := f.accountService.CreateAccountWithOwner(&account, ownerCtx); err != nil {
		if err.Error() == financeAccountSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
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
	ownerCtx := utils.GetOwnerContext(c)

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

	if err := f.accountService.UpdateAccountWithOwner(uint(accountId), updates, ownerCtx); err != nil {
		if err.Error() == financeAccountSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		if err.Error() == "finance account with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(404).JSON(fiber.Map{"error": "Finance account not found or access denied"})
	}

	// Get updated account
	account, _ := f.accountService.GetAccountByIDWithOwner(uint(accountId), ownerCtx)

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
	ownerCtx := utils.GetOwnerContext(c)

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

	if err := f.accountService.DeleteAccountWithOwner(uint(accountId), ownerCtx); err != nil {
		if err.Error() == financeAccountSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(404).JSON(fiber.Map{"error": "Finance account not found or access denied"})
	}

	return c.JSON(fiber.Map{
		"message": "Finance account deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Finance account deleted successfully",
			"type": "success",
		},
	})
}

// ============================================
// Owner-based methods for data filtering
// ============================================

const financeAccountSystemAdminError = "system admin cannot modify data in owner-based tables (view only)"

func (f *FinanceAccountsController) ownerList(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

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

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	accounts, total, err := f.accountService.ListAccountsWithOwner(filters, page, limit, ownerCtx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve finance accounts",
			"details": err.Error(),
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

func (f *FinanceAccountsController) ownerShow(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

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

	account, err := f.accountService.GetAccountByIDWithOwner(uint(accountId), ownerCtx)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Finance account not found or access denied"})
	}

	return c.JSON(fiber.Map{"data": account})
}

func (f *FinanceAccountsController) ownerCreate(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	var account models.FinanceAccount
	if err := c.BodyParser(&account); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := f.accountService.CreateAccountWithOwner(&account, ownerCtx); err != nil {
		if err.Error() == financeAccountSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
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

func (f *FinanceAccountsController) ownerUpdate(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

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

	if err := f.accountService.UpdateAccountWithOwner(uint(accountId), updates, ownerCtx); err != nil {
		if err.Error() == financeAccountSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		if err.Error() == "finance account with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(404).JSON(fiber.Map{"error": "Finance account not found or access denied"})
	}

	account, _ := f.accountService.GetAccountByIDWithOwner(uint(accountId), ownerCtx)

	return c.JSON(fiber.Map{
		"message": "Finance account updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Finance account updated successfully",
			"type": "success",
		},
		"data": account,
	})
}

func (f *FinanceAccountsController) ownerDelete(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

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

	if err := f.accountService.DeleteAccountWithOwner(uint(accountId), ownerCtx); err != nil {
		if err.Error() == financeAccountSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(404).JSON(fiber.Map{"error": "Finance account not found or access denied"})
	}

	return c.JSON(fiber.Map{
		"message": "Finance account deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Finance account deleted successfully",
			"type": "success",
		},
	})
}

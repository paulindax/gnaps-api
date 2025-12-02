package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type BillParticularsController struct {
	particularService *services.BillParticularService
}

func NewBillParticularsController(particularService *services.BillParticularService) *BillParticularsController {
	return &BillParticularsController{
		particularService: particularService,
	}
}

func (b *BillParticularsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return b.list(c)
	case "show":
		return b.show(c)
	case "create":
		return b.create(c)
	case "update":
		return b.update(c)
	case "delete":
		return b.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (b *BillParticularsController) list(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if financeAccountId := c.Query("finance_account_id"); financeAccountId != "" {
		if id, err := strconv.ParseInt(financeAccountId, 10, 64); err == nil {
			filters["finance_account_id"] = id
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	particulars, total, err := b.particularService.ListParticulars(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve bill particulars",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve bill particulars",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": particulars,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (b *BillParticularsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	particularId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	particular, err := b.particularService.GetParticularByID(uint(particularId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": particular})
}

func (b *BillParticularsController) create(c *fiber.Ctx) error {
	var particular models.BillParticular
	if err := c.BodyParser(&particular); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := b.particularService.CreateParticular(&particular); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Bill particular created successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill particular created successfully",
			"type": "success",
		},
		"data": particular,
	})
}

func (b *BillParticularsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	particularId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.BillParticular
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
	if updateData.Priority != nil {
		updates["priority"] = *updateData.Priority
	}
	if updateData.FinanceAccountId != nil {
		updates["finance_account_id"] = *updateData.FinanceAccountId
	}
	if updateData.IsArrears != nil {
		updates["is_arrears"] = *updateData.IsArrears
	}

	if err := b.particularService.UpdateParticular(uint(particularId), updates); err != nil {
		if err.Error() == "bill particular not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated particular
	particular, _ := b.particularService.GetParticularByID(uint(particularId))

	return c.JSON(fiber.Map{
		"message": "Bill particular updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill particular updated successfully",
			"type": "success",
		},
		"data": particular,
	})
}

func (b *BillParticularsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	particularId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := b.particularService.DeleteParticular(uint(particularId)); err != nil {
		if err.Error() == "bill particular not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Bill particular deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill particular deleted successfully",
			"type": "success",
		},
	})
}

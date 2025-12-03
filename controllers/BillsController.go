package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type BillsController struct {
	billService *services.BillService
}

func NewBillsController(billService *services.BillService) *BillsController {
	return &BillsController{
		billService: billService,
	}
}

func (b *BillsController) Handle(action string, c *fiber.Ctx) error {
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
	case "items":
		return b.getItems(c)
	case "create-item":
		return b.createItem(c)
	case "update-item":
		return b.updateItem(c)
	case "delete-item":
		return b.deleteItem(c)
	case "item-assignments":
		return b.getItemAssignments(c)
	case "create-assignments":
		return b.createAssignments(c)
	case "delete-assignment":
		return b.deleteAssignment(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

// Bill CRUD operations
func (b *BillsController) list(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if academicYear := c.Query("academic_year"); academicYear != "" {
		filters["academic_year"] = academicYear
	}
	if term := c.Query("term"); term != "" {
		filters["term"] = term
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	bills, total, err := b.billService.ListBills(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve bills",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve bills",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": bills,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (b *BillsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	billId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	bill, err := b.billService.GetBillByID(uint(billId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": bill})
}

func (b *BillsController) create(c *fiber.Ctx) error {
	var bill models.Bill
	if err := c.BodyParser(&bill); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := b.billService.CreateBill(&bill); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Bill created successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill created successfully",
			"type": "success",
		},
		"data": bill,
	})
}

func (b *BillsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	billId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.Bill
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
	if updateData.Description != nil {
		updates["description"] = *updateData.Description
	}

	if err := b.billService.UpdateBill(uint(billId), updates); err != nil {
		if err.Error() == "bill not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated bill
	bill, _ := b.billService.GetBillByID(uint(billId))

	return c.JSON(fiber.Map{
		"message": "Bill updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill updated successfully",
			"type": "success",
		},
		"data": bill,
	})
}

func (b *BillsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	billId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := b.billService.DeleteBill(uint(billId)); err != nil {
		if err.Error() == "bill not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Bill deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill deleted successfully",
			"type": "success",
		},
	})
}

// Bill Item operations
func (b *BillsController) getItems(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("bill_id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Bill ID is required"})
	}

	billId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid Bill ID"})
	}

	// Check if pagination parameters are provided
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	search := c.Query("search")

	// If pagination parameters are provided, use pagination
	if pageStr != "" || limitStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}

		items, total, err := b.billService.GetBillItemsByBillIDWithPagination(uint(billId), page, limit, search)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to retrieve bill items",
				"details": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data": items,
			"pagination": fiber.Map{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
	}

	// Default behavior without pagination for backward compatibility
	items, err := b.billService.GetBillItemsByBillID(uint(billId))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve bill items",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"data": items})
}

func (b *BillsController) createItem(c *fiber.Ctx) error {
	var billItem models.BillItem
	if err := c.BodyParser(&billItem); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := b.billService.CreateBillItem(&billItem); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Bill item created successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill item created successfully",
			"type": "success",
		},
		"data": billItem,
	})
}

func (b *BillsController) updateItem(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	itemId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.BillItem
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Build updates map
	updates := make(map[string]interface{})
	if updateData.BillParticularId != nil {
		updates["bill_particular_id"] = *updateData.BillParticularId
	}
	if updateData.Amount != nil {
		updates["amount"] = *updateData.Amount
	}

	if err := b.billService.UpdateBillItem(uint(itemId), updates); err != nil {
		if err.Error() == "bill item not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated bill item
	billItem, _ := b.billService.GetBillItemByID(uint(itemId))

	return c.JSON(fiber.Map{
		"message": "Bill item updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill item updated successfully",
			"type": "success",
		},
		"data": billItem,
	})
}

func (b *BillsController) deleteItem(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	itemId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := b.billService.DeleteBillItem(uint(itemId)); err != nil {
		if err.Error() == "bill item not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Bill item deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Bill item deleted successfully",
			"type": "success",
		},
	})
}

// Assignment operations
func (b *BillsController) getItemAssignments(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("bill_item_id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Bill Item ID is required"})
	}

	itemId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid Bill Item ID"})
	}

	assignments, err := b.billService.GetAssignmentsByBillItemID(uint(itemId))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve assignments",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"data": assignments})
}

func (b *BillsController) createAssignments(c *fiber.Ctx) error {
	var request struct {
		BillItemId uint                      `json:"bill_item_id"`
		Assignments []models.BillAssignment  `json:"assignments"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Set the bill item ID for all assignments
	for i := range request.Assignments {
		request.Assignments[i].BillItemId = &request.BillItemId
	}

	if err := b.billService.CreateAssignments(request.Assignments); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Assignments created successfully",
		"flash_message": fiber.Map{
			"msg":  "Assignments created successfully",
			"type": "success",
		},
		"data": request.Assignments,
	})
}

func (b *BillsController) deleteAssignment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	assignmentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := b.billService.DeleteAssignment(uint(assignmentId)); err != nil {
		if err.Error() == "assignment not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Assignment deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Assignment deleted successfully",
			"type": "success",
		},
	})
}

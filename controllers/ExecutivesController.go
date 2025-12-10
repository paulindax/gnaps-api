package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ExecutivesController struct {
	executiveService *services.ExecutiveService
}

func NewExecutivesController(executiveService *services.ExecutiveService) *ExecutivesController {
	return &ExecutivesController{
		executiveService: executiveService,
	}
}

func (e *ExecutivesController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return e.list(c)
	case "show":
		return e.show(c)
	case "create":
		return e.create(c)
	case "update":
		return e.update(c)
	case "delete":
		return e.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (e *ExecutivesController) list(c *fiber.Ctx) error {
	// Get owner context for role-based filtering
	ownerCtx := utils.GetOwnerContext(c)
	if ownerCtx == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized - owner context not found"})
	}

	// Parse filters from query params
	filters := make(map[string]interface{})

	// Search filter (searches name, email, executive_no)
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if positionID := c.Query("position_id"); positionID != "" {
		filters["position_id"] = positionID
	}
	if gender := c.Query("gender"); gender != "" {
		filters["gender"] = gender
	}
	if firstName := c.Query("first_name"); firstName != "" {
		filters["first_name"] = firstName
	}
	if lastName := c.Query("last_name"); lastName != "" {
		filters["last_name"] = lastName
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if executiveNo := c.Query("executive_no"); executiveNo != "" {
		filters["executive_no"] = executiveNo
	}
	if email := c.Query("email"); email != "" {
		filters["email"] = email
	}
	if mobileNo := c.Query("mobile_no"); mobileNo != "" {
		filters["mobile_no"] = mobileNo
	}
	// New filters for role-based management
	if role := c.Query("role"); role != "" {
		filters["role"] = role
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if regionID := c.Query("region_id"); regionID != "" {
		filters["region_id"] = regionID
	}
	if zoneID := c.Query("zone_id"); zoneID != "" {
		filters["zone_id"] = zoneID
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	executives, total, err := e.executiveService.ListExecutivesWithRole(filters, page, limit, ownerCtx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve executives",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve executives",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": executives,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (e *ExecutivesController) show(c *fiber.Ctx) error {
	// Get owner context for role-based filtering
	ownerCtx := utils.GetOwnerContext(c)
	if ownerCtx == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized - owner context not found"})
	}

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	executiveId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	executive, err := e.executiveService.GetExecutiveByIDWithRole(uint(executiveId), ownerCtx)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": executive})
}

func (e *ExecutivesController) create(c *fiber.Ctx) error {
	var executive models.Executive
	if err := c.BodyParser(&executive); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := e.executiveService.CreateExecutive(&executive); err != nil {
		if err.Error() == "executive with this executive_no already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "executive with this email already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "invalid gender value" {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Executive created successfully",
		"flash_message": fiber.Map{
			"msg":  "Executive created successfully",
			"type": "success",
		},
		"data": executive,
	})
}

func (e *ExecutivesController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	executiveId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.Executive
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Build updates map
	updates := make(map[string]interface{})
	if updateData.FirstName != nil {
		updates["first_name"] = *updateData.FirstName
	}
	if updateData.LastName != nil {
		updates["last_name"] = *updateData.LastName
	}
	if updateData.ExecutiveNo != nil {
		updates["executive_no"] = *updateData.ExecutiveNo
	}
	if updateData.Email != nil {
		updates["email"] = *updateData.Email
	}
	if updateData.MobileNo != nil {
		updates["mobile_no"] = *updateData.MobileNo
	}
	if updateData.Gender != nil {
		updates["gender"] = *updateData.Gender
	}
	if updateData.PositionId != nil {
		updates["position_id"] = *updateData.PositionId
	}
	// New fields for role-based management
	if updateData.Role != nil {
		updates["role"] = *updateData.Role
	}
	if updateData.RegionId != nil {
		updates["region_id"] = *updateData.RegionId
	}
	if updateData.ZoneId != nil {
		updates["zone_id"] = *updateData.ZoneId
	}
	if updateData.Status != nil {
		updates["status"] = *updateData.Status
	}
	if updateData.Bio != nil {
		updates["bio"] = *updateData.Bio
	}

	if err := e.executiveService.UpdateExecutive(uint(executiveId), updates); err != nil {
		if err.Error() == "executive not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "executive with this executive_no already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "executive with this email already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "invalid gender value" {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated executive
	executive, _ := e.executiveService.GetExecutiveByID(uint(executiveId))

	return c.JSON(fiber.Map{
		"message": "Executive updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Executive updated successfully",
			"type": "success",
		},
		"data": executive,
	})
}

func (e *ExecutivesController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	executiveId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := e.executiveService.DeleteExecutive(uint(executiveId)); err != nil {
		if err.Error() == "executive not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Executive deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Executive deleted successfully",
			"type": "success",
		},
	})
}

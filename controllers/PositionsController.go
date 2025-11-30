package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PositionsController struct {
	positionService *services.PositionService
}

func NewPositionsController(positionService *services.PositionService) *PositionsController {
	return &PositionsController{
		positionService: positionService,
	}
}

func (p *PositionsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return p.list(c)
	case "show":
		return p.show(c)
	case "create":
		return p.create(c)
	case "update":
		return p.update(c)
	case "delete":
		return p.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (p *PositionsController) list(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	positions, total, err := p.positionService.ListPositions(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve positions",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve positions",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": positions,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (p *PositionsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	positionId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	position, err := p.positionService.GetPositionByID(uint(positionId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": position})
}

func (p *PositionsController) create(c *fiber.Ctx) error {
	var position models.Position
	if err := c.BodyParser(&position); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := p.positionService.CreatePosition(&position); err != nil {
		if err.Error() == "position with this name already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Position created successfully",
		"flash_message": fiber.Map{
			"msg":  "Position created successfully",
			"type": "success",
		},
		"data": position,
	})
}

func (p *PositionsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	positionId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.Position
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

	if err := p.positionService.UpdatePosition(uint(positionId), updates); err != nil {
		if err.Error() == "position not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "position with this name already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated position
	position, _ := p.positionService.GetPositionByID(uint(positionId))

	return c.JSON(fiber.Map{
		"message": "Position updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Position updated successfully",
			"type": "success",
		},
		"data": position,
	})
}

func (p *PositionsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	positionId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := p.positionService.DeletePosition(uint(positionId)); err != nil {
		if err.Error() == "position not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "position is in use by executives" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Position deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Position deleted successfully",
			"type": "success",
		},
	})
}

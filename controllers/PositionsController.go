package controllers

import (
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PositionsController struct {
}

func init() {
	RegisterController("positions", &PositionsController{})
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
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// list retrieves all non-deleted positions
func (p *PositionsController) list(c *fiber.Ctx) error {
	var positions []models.Position

	// Query parameters for filtering
	query := DB.Where("is_deleted = ?", false)

	// General search parameter (searches name)
	search := c.Query("search")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ?", searchPattern)
	}

	// Optional search by name (for backward compatibility)
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Position{}).Count(&total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&positions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve positions",
			"details": result.Error.Error(),
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

// show retrieves a single position by ID
func (p *PositionsController) show(c *fiber.Ctx) error {
	// Get ID from URL params or query string
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
			"flash_message": fiber.Map{
				"msg":  "ID is required",
				"type": "error",
			},
		})
	}

	var position models.Position
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&position)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Position not found",
			"flash_message": fiber.Map{
				"msg":  "Position not found",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": position,
	})
}

// create creates a new position
func (p *PositionsController) create(c *fiber.Ctx) error {
	var position models.Position

	if err := c.BodyParser(&position); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Invalid request body",
				"type": "error",
			},
		})
	}

	// Validate required fields
	if position.Name == nil || *position.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name is required",
			"flash_message": fiber.Map{
				"msg":  "Name is required",
				"type": "error",
			},
		})
	}

	// Check if name already exists
	var existing models.Position
	if err := DB.Where("name = ? AND is_deleted = ?", position.Name, false).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Position with this name already exists",
			"flash_message": fiber.Map{
				"msg":  "Position with this name already exists",
				"type": "error",
			},
		})
	}

	// Set default values
	falseVal := false
	position.IsDeleted = &falseVal

	result := DB.Create(&position)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create position",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to create position",
				"type": "error",
			},
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Position created successfully",
		"data":    position,
		"flash_message": fiber.Map{
			"msg":  "Position created successfully",
			"type": "success",
		},
	})
}

// update updates an existing position
func (p *PositionsController) update(c *fiber.Ctx) error {
	// Get ID from URL params or query string
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
			"flash_message": fiber.Map{
				"msg":  "ID is required",
				"type": "error",
			},
		})
	}

	var position models.Position
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&position)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Position not found",
			"flash_message": fiber.Map{
				"msg":  "Position not found",
				"type": "error",
			},
		})
	}

	var updateData models.Position
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Invalid request body",
				"type": "error",
			},
		})
	}

	// Check if name is being changed and if new name already exists
	if updateData.Name != nil && *updateData.Name != *position.Name {
		var existing models.Position
		if err := DB.Where("name = ? AND id != ? AND is_deleted = ?", updateData.Name, id, false).First(&existing).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Position with this name already exists",
				"flash_message": fiber.Map{
					"msg":  "Position with this name already exists",
					"type": "error",
				},
			})
		}
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if updateData.Name != nil {
		updates["name"] = updateData.Name
	}

	result = DB.Model(&position).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update position",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to update position",
				"type": "error",
			},
		})
	}

	// Fetch updated record
	DB.First(&position, id)

	return c.JSON(fiber.Map{
		"message": "Position updated successfully",
		"data":    position,
		"flash_message": fiber.Map{
			"msg":  "Position updated successfully",
			"type": "success",
		},
	})
}

// delete soft deletes a position
func (p *PositionsController) delete(c *fiber.Ctx) error {
	// Get ID from URL params or query string
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
			"flash_message": fiber.Map{
				"msg":  "ID is required",
				"type": "error",
			},
		})
	}

	var position models.Position
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&position)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Position not found",
			"flash_message": fiber.Map{
				"msg":  "Position not found",
				"type": "error",
			},
		})
	}

	// Check if position is being used by any executives
	var executiveCount int64
	DB.Model(&models.Executive{}).Where("position_id = ? AND is_deleted = ?", id, 0).Count(&executiveCount)

	if executiveCount > 0 {
		return c.Status(409).JSON(fiber.Map{
			"error":   "Cannot delete position",
			"details": fmt.Sprintf("Position is currently assigned to %d executive(s)", executiveCount),
			"flash_message": fiber.Map{
				"msg":  fmt.Sprintf("Cannot delete position - currently assigned to %d executive(s)", executiveCount),
				"type": "error",
			},
		})
	}

	// Soft delete by setting is_deleted flag
	trueVal := true
	result = DB.Model(&position).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete position",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to delete position",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Position deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Position deleted successfully",
			"type": "success",
		},
	})
}

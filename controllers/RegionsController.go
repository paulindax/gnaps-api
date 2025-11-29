package controllers

import (
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type RegionsController struct {
}

func init() {
	RegisterController("regions", &RegionsController{})
}

func (r *RegionsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return r.list(c)
	case "show":
		return r.show(c)
	case "create":
		return r.create(c)
	case "update":
		return r.update(c)
	case "delete":
		return r.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// list retrieves all non-deleted regions
func (r *RegionsController) list(c *fiber.Ctx) error {
	var regions []models.Region

	// Query parameters for filtering
	query := DB.Where("is_deleted = ?", false)

	// General search parameter (searches both name and code)
	search := c.Query("search")
	fmt.Printf("[RegionsController] Search parameter: '%s'\n", search)
	if search != "" {
		searchPattern := "%" + search + "%"
		fmt.Printf("[RegionsController] Applying search with pattern: %s\n", searchPattern)
		query = query.Where("name LIKE ? OR code LIKE ?", searchPattern, searchPattern)
	}

	// Optional search by name (for backward compatibility)
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Optional filter by code (for backward compatibility)
	if code := c.Query("code"); code != "" {
		query = query.Where("code = ?", code)
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Region{}).Count(&total)
	fmt.Printf("[RegionsController] Total matching regions: %d\n", total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&regions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve regions",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve regions",
				"type": "error",
			},
		})
	}

	fmt.Printf("[RegionsController] Returning %d regions (page %d, limit %d)\n", len(regions), page, limit)
	return c.JSON(fiber.Map{
		"data": regions,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// show retrieves a single region by ID
func (r *RegionsController) show(c *fiber.Ctx) error {
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

	var region models.Region
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&region)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Region not found",
			"flash_message": fiber.Map{
				"msg":  "Region not found",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": region,
	})
}

// create creates a new region
func (r *RegionsController) create(c *fiber.Ctx) error {
	var region models.Region

	if err := c.BodyParser(&region); err != nil {
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
	if region.Name == nil || *region.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name is required",
			"flash_message": fiber.Map{
				"msg":  "Name is required",
				"type": "error",
			},
		})
	}

	if region.Code == nil || *region.Code == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Code is required",
			"flash_message": fiber.Map{
				"msg":  "Code is required",
				"type": "error",
			},
		})
	}

	// Check if code already exists
	var existing models.Region
	if err := DB.Where("code = ? AND is_deleted = ?", region.Code, false).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Region with this code already exists",
			"flash_message": fiber.Map{
				"msg":  "Region with this code already exists",
				"type": "error",
			},
		})
	}

	// Set default values
	falseVal := false
	region.IsDeleted = &falseVal

	result := DB.Create(&region)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create region",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to create region",
				"type": "error",
			},
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Region created successfully",
		"data":    region,
		"flash_message": fiber.Map{
			"msg":  "Region created successfully",
			"type": "success",
		},
	})
}

// update updates an existing region
func (r *RegionsController) update(c *fiber.Ctx) error {
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

	var region models.Region
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&region)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Region not found",
			"flash_message": fiber.Map{
				"msg":  "Region not found",
				"type": "error",
			},
		})
	}

	var updateData models.Region
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

	// Check if code is being changed and if new code already exists
	if updateData.Code != nil && *updateData.Code != *region.Code {
		var existing models.Region
		if err := DB.Where("code = ? AND id != ? AND is_deleted = ?", updateData.Code, id, false).First(&existing).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Region with this code already exists",
				"flash_message": fiber.Map{
					"msg":  "Region with this code already exists",
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
	if updateData.Code != nil {
		updates["code"] = updateData.Code
	}

	result = DB.Model(&region).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update region",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to update region",
				"type": "error",
			},
		})
	}

	// Fetch updated record
	DB.First(&region, id)

	return c.JSON(fiber.Map{
		"message": "Region updated successfully",
		"data":    region,
		"flash_message": fiber.Map{
			"msg":  "Region updated successfully",
			"type": "success",
		},
	})
}

// delete soft deletes a region
func (r *RegionsController) delete(c *fiber.Ctx) error {
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

	var region models.Region
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&region)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Region not found",
			"flash_message": fiber.Map{
				"msg":  "Region not found",
				"type": "error",
			},
		})
	}

	// Soft delete by setting is_deleted flag
	trueVal := true
	result = DB.Model(&region).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete region",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to delete region",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Region deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Region deleted successfully",
			"type": "success",
		},
	})
}

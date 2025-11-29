package controllers

import (
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ZonesController struct {
}

func init() {
	RegisterController("zones", &ZonesController{})
}

func (z *ZonesController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return z.list(c)
	case "show":
		return z.show(c)
	case "create":
		return z.create(c)
	case "update":
		return z.update(c)
	case "delete":
		return z.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// list retrieves all non-deleted zones
func (z *ZonesController) list(c *fiber.Ctx) error {
	var zones []models.Zone

	// Query parameters for filtering
	query := DB.Where("is_deleted = ?", false).Order("region_id ASC, name ASC")

	// Optional filter by region_id
	if regionID := c.Query("region_id"); regionID != "" {
		query = query.Where("region_id = ?", regionID)
	}

	// General search parameter (searches both name and code)
	search := c.Query("search")
	if search != "" {
		searchPattern := "%" + search + "%"
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
	query.Model(&models.Zone{}).Count(&total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&zones)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve zones",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve zones",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": zones,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// show retrieves a single zone by ID
func (z *ZonesController) show(c *fiber.Ctx) error {
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

	var zone models.Zone
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&zone)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Zone not found",
			"flash_message": fiber.Map{
				"msg":  "Zone not found",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": zone,
	})
}

// create creates a new zone
func (z *ZonesController) create(c *fiber.Ctx) error {
	var zone models.Zone

	if err := c.BodyParser(&zone); err != nil {
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
	if zone.Name == nil || *zone.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name is required",
			"flash_message": fiber.Map{
				"msg":  "Name is required",
				"type": "error",
			},
		})
	}

	if zone.Code == nil || *zone.Code == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Code is required",
			"flash_message": fiber.Map{
				"msg":  "Code is required",
				"type": "error",
			},
		})
	}

	if zone.RegionId == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Region ID is required",
			"flash_message": fiber.Map{
				"msg":  "Region ID is required",
				"type": "error",
			},
		})
	}

	// Verify that the region exists
	var region models.Region
	if err := DB.Where("id = ? AND is_deleted = ?", zone.RegionId, false).First(&region).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid region ID - Region does not exist",
			"flash_message": fiber.Map{
				"msg":  "Invalid region ID - Region does not exist",
				"type": "error",
			},
		})
	}

	// Check if code already exists within the same region
	var existing models.Zone
	if err := DB.Where("code = ? AND region_id = ? AND is_deleted = ?", zone.Code, zone.RegionId, false).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Zone with this code already exists in this region",
			"flash_message": fiber.Map{
				"msg":  "Zone with this code already exists in this region",
				"type": "error",
			},
		})
	}

	// Set default values
	falseVal := false
	zone.IsDeleted = &falseVal

	result := DB.Create(&zone)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create zone",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to create zone",
				"type": "error",
			},
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Zone created successfully",
		"data":    zone,
		"flash_message": fiber.Map{
			"msg":  "Zone created successfully",
			"type": "success",
		},
	})
}

// update updates an existing zone
func (z *ZonesController) update(c *fiber.Ctx) error {
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

	var zone models.Zone
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&zone)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Zone not found",
			"flash_message": fiber.Map{
				"msg":  "Zone not found",
				"type": "error",
			},
		})
	}

	var updateData models.Zone
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

	// If region_id is being changed, verify the new region exists
	if updateData.RegionId != nil && *updateData.RegionId != *zone.RegionId {
		var region models.Region
		if err := DB.Where("id = ? AND is_deleted = ?", updateData.RegionId, false).First(&region).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid region ID - Region does not exist",
				"flash_message": fiber.Map{
					"msg":  "Invalid region ID - Region does not exist",
					"type": "error",
				},
			})
		}
	}

	// Check if code is being changed and if new code already exists in the region
	regionIDForCheck := zone.RegionId
	if updateData.RegionId != nil {
		regionIDForCheck = updateData.RegionId
	}

	if updateData.Code != nil && *updateData.Code != *zone.Code {
		var existing models.Zone
		if err := DB.Where("code = ? AND region_id = ? AND id != ? AND is_deleted = ?", updateData.Code, regionIDForCheck, id, false).First(&existing).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Zone with this code already exists in this region",
				"flash_message": fiber.Map{
					"msg":  "Zone with this code already exists in this region",
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
	if updateData.RegionId != nil {
		updates["region_id"] = updateData.RegionId
	}

	result = DB.Model(&zone).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update zone",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to update zone",
				"type": "error",
			},
		})
	}

	// Fetch updated record
	DB.First(&zone, id)

	return c.JSON(fiber.Map{
		"message": "Zone updated successfully",
		"data":    zone,
		"flash_message": fiber.Map{
			"msg":  "Zone updated successfully",
			"type": "success",
		},
	})
}

// delete soft deletes a zone
func (z *ZonesController) delete(c *fiber.Ctx) error {
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

	var zone models.Zone
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&zone)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Zone not found",
			"flash_message": fiber.Map{
				"msg":  "Zone not found",
				"type": "error",
			},
		})
	}

	// Soft delete by setting is_deleted flag
	trueVal := true
	result = DB.Model(&zone).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete zone",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to delete zone",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Zone deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Zone deleted successfully",
			"type": "success",
		},
	})
}

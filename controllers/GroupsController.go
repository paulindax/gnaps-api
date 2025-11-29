package controllers

import (
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type GroupsController struct {
}

func init() {
	RegisterController("groups", &GroupsController{})
}

func (g *GroupsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return g.list(c)
	case "show":
		return g.show(c)
	case "create":
		return g.create(c)
	case "update":
		return g.update(c)
	case "delete":
		return g.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// list retrieves all non-deleted groups
func (g *GroupsController) list(c *fiber.Ctx) error {
	var groups []models.SchoolGroup

	// Query parameters for filtering
	query := DB.Where("is_deleted = ?", false)

	// Optional filter by zone_id
	if zoneID := c.Query("zone_id"); zoneID != "" {
		query = query.Where("zone_id = ?", zoneID)
	}

	// General search parameter (searches name and description)
	search := c.Query("search")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR (description IS NOT NULL AND description LIKE ?)", searchPattern, searchPattern)
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
	query.Model(&models.SchoolGroup{}).Count(&total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&groups)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve groups",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve groups",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": groups,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// show retrieves a single group by ID
func (g *GroupsController) show(c *fiber.Ctx) error {
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

	var group models.SchoolGroup
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&group)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Group not found",
			"flash_message": fiber.Map{
				"msg":  "Group not found",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": group,
	})
}

// create creates a new group
func (g *GroupsController) create(c *fiber.Ctx) error {
	var group models.SchoolGroup

	if err := c.BodyParser(&group); err != nil {
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
	if group.Name == nil || *group.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name is required",
			"flash_message": fiber.Map{
				"msg":  "Name is required",
				"type": "error",
			},
		})
	}

	// if group.ZoneId == nil {
	// 	return c.Status(400).JSON(fiber.Map{
	// 		"error": "Zone ID is required",
	// 		"flash_message": fiber.Map{
	// 			"msg":  "Zone ID is required",
	// 			"type": "error",
	// 		},
	// 	})
	// }

	// Verify that the zone exists
	// var zone models.Zone
	// if err := DB.Where("id = ? AND is_deleted = ?", group.ZoneId, false).First(&zone).Error; err != nil {
	// 	return c.Status(400).JSON(fiber.Map{
	// 		"error": "Invalid zone ID - Zone does not exist",
	// 		"flash_message": fiber.Map{
	// 			"msg":  "Invalid zone ID - Zone does not exist",
	// 			"type": "error",
	// 		},
	// 	})
	// }

	// Check if name already exists within the same zone
	var existing models.SchoolGroup
	if err := DB.Where("name = ?  AND is_deleted = ?", group.Name, false).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Group with this name already exists in this zone",
			"flash_message": fiber.Map{
				"msg":  "Group with this name already exists in this zone",
				"type": "error",
			},
		})
	}

	// Set default values
	//falseVal := false
	group.IsDeleted = false

	result := DB.Create(&group)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create group",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to create group",
				"type": "error",
			},
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Group created successfully",
		"data":    group,
		"flash_message": fiber.Map{
			"msg":  "Group created successfully",
			"type": "success",
		},
	})
}

// update updates an existing group
func (g *GroupsController) update(c *fiber.Ctx) error {
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

	var group models.SchoolGroup
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&group)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Group not found",
			"flash_message": fiber.Map{
				"msg":  "Group not found",
				"type": "error",
			},
		})
	}

	var updateData models.SchoolGroup
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

	// If zone_id is being changed, verify the new zone exists
	// if updateData.ZoneId != nil && (group.ZoneId == nil || *updateData.ZoneId != *group.ZoneId) {
	// 	var zone models.Zone
	// 	if err := DB.Where("id = ? AND is_deleted = ?", updateData.ZoneId, false).First(&zone).Error; err != nil {
	// 		return c.Status(400).JSON(fiber.Map{
	// 			"error": "Invalid zone ID - Zone does not exist",
	// 			"flash_message": fiber.Map{
	// 				"msg":  "Invalid zone ID - Zone does not exist",
	// 				"type": "error",
	// 			},
	// 		})
	// 	}
	// }

	// Check if name is being changed and if new name already exists in the zone
	// zoneIDForCheck := group.ZoneId
	// if updateData.ZoneId != nil {
	// 	zoneIDForCheck = updateData.ZoneId
	// }

	if updateData.Name != nil && (group.Name == nil || *updateData.Name != *group.Name) {
		var existing models.SchoolGroup
		if err := DB.Where("name = ? AND id != ? AND is_deleted = ?", updateData.Name, id, false).First(&existing).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Group with this name already exists in this zone",
				"flash_message": fiber.Map{
					"msg":  "Group with this name already exists in this zone",
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
	// if updateData.ZoneId != nil {
	// 	updates["zone_id"] = updateData.ZoneId
	// }
	if updateData.Description != nil {
		updates["description"] = updateData.Description
	}

	result = DB.Model(&group).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update group",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to update group",
				"type": "error",
			},
		})
	}

	// Fetch updated record
	DB.First(&group, id)

	return c.JSON(fiber.Map{
		"message": "Group updated successfully",
		"data":    group,
		"flash_message": fiber.Map{
			"msg":  "Group updated successfully",
			"type": "success",
		},
	})
}

// delete soft deletes a group
func (g *GroupsController) delete(c *fiber.Ctx) error {
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

	var group models.SchoolGroup
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&group)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Group not found",
			"flash_message": fiber.Map{
				"msg":  "Group not found",
				"type": "error",
			},
		})
	}

	// Soft delete by setting is_deleted flag
	trueVal := true
	result = DB.Model(&group).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete group",
			"details": result.Error.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to delete group",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Group deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Group deleted successfully",
			"type": "success",
		},
	})
}

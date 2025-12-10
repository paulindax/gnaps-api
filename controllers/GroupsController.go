package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type GroupsController struct {
	groupService *services.GroupService
}

func NewGroupsController(groupService *services.GroupService) *GroupsController {
	return &GroupsController{
		groupService: groupService,
	}
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
	// Owner-based actions
	case "ownerList":
		return g.ownerList(c)
	case "ownerShow":
		return g.ownerShow(c)
	case "ownerCreate":
		return g.ownerCreate(c)
	case "ownerUpdate":
		return g.ownerUpdate(c)
	case "ownerDelete":
		return g.ownerDelete(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (g *GroupsController) list(c *fiber.Ctx) error {
	// Get owner context for filtering
	ownerCtx := utils.GetOwnerContext(c)

	// Parse filters from query params
	filters := make(map[string]interface{})
	if zoneID := c.Query("zone_id"); zoneID != "" {
		filters["zone_id"] = zoneID
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Use owner filtering to ensure users only see groups they have access to
	groups, total, err := g.groupService.ListGroupsWithOwner(filters, page, limit, ownerCtx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve groups",
			"details": err.Error(),
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

func (g *GroupsController) show(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	groupId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	group, err := g.groupService.GetGroupByIDWithOwner(uint(groupId), ownerCtx)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Group not found or access denied"})
	}

	return c.JSON(fiber.Map{"data": group})
}

func (g *GroupsController) create(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	var group models.SchoolGroup
	if err := c.BodyParser(&group); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := g.groupService.CreateGroupWithOwner(&group, ownerCtx); err != nil {
		if err.Error() == groupSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Group created successfully",
		"flash_message": fiber.Map{
			"msg":  "Group created successfully",
			"type": "success",
		},
		"data": group,
	})
}

func (g *GroupsController) update(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	groupId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.SchoolGroup
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
	if updateData.ZoneId != nil {
		updates["zone_id"] = *updateData.ZoneId
	}

	if err := g.groupService.UpdateGroupWithOwner(uint(groupId), updates, ownerCtx); err != nil {
		if err.Error() == groupSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(404).JSON(fiber.Map{"error": "Group not found or access denied"})
	}

	// Get updated group
	group, _ := g.groupService.GetGroupByIDWithOwner(uint(groupId), ownerCtx)

	return c.JSON(fiber.Map{
		"message": "Group updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Group updated successfully",
			"type": "success",
		},
		"data": group,
	})
}

func (g *GroupsController) delete(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	groupId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := g.groupService.DeleteGroupWithOwner(uint(groupId), ownerCtx); err != nil {
		if err.Error() == groupSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(404).JSON(fiber.Map{"error": "Group not found or access denied"})
	}

	return c.JSON(fiber.Map{
		"message": "Group deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Group deleted successfully",
			"type": "success",
		},
	})
}

// ============================================
// Owner-based methods for data filtering
// ============================================

const groupSystemAdminError = "system admin cannot modify data in owner-based tables (view only)"

func (g *GroupsController) ownerList(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	filters := make(map[string]interface{})
	if zoneID := c.Query("zone_id"); zoneID != "" {
		filters["zone_id"] = zoneID
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	groups, total, err := g.groupService.ListGroupsWithOwner(filters, page, limit, ownerCtx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve groups",
			"details": err.Error(),
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

func (g *GroupsController) ownerShow(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	groupId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	group, err := g.groupService.GetGroupByIDWithOwner(uint(groupId), ownerCtx)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Group not found or access denied"})
	}

	return c.JSON(fiber.Map{"data": group})
}

func (g *GroupsController) ownerCreate(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	var group models.SchoolGroup
	if err := c.BodyParser(&group); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := g.groupService.CreateGroupWithOwner(&group, ownerCtx); err != nil {
		if err.Error() == groupSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Group created successfully",
		"flash_message": fiber.Map{
			"msg":  "Group created successfully",
			"type": "success",
		},
		"data": group,
	})
}

func (g *GroupsController) ownerUpdate(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	groupId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.SchoolGroup
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
	if updateData.Description != nil {
		updates["description"] = *updateData.Description
	}
	if updateData.ZoneId != nil {
		updates["zone_id"] = *updateData.ZoneId
	}

	if err := g.groupService.UpdateGroupWithOwner(uint(groupId), updates, ownerCtx); err != nil {
		if err.Error() == groupSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(404).JSON(fiber.Map{"error": "Group not found or access denied"})
	}

	group, _ := g.groupService.GetGroupByIDWithOwner(uint(groupId), ownerCtx)

	return c.JSON(fiber.Map{
		"message": "Group updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Group updated successfully",
			"type": "success",
		},
		"data": group,
	})
}

func (g *GroupsController) ownerDelete(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	groupId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := g.groupService.DeleteGroupWithOwner(uint(groupId), ownerCtx); err != nil {
		if err.Error() == groupSystemAdminError {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return c.Status(404).JSON(fiber.Map{"error": "Group not found or access denied"})
	}

	return c.JSON(fiber.Map{
		"message": "Group deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Group deleted successfully",
			"type": "success",
		},
	})
}

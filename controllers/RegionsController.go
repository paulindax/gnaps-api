package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type RegionsController struct {
	regionService *services.RegionService
}

func NewRegionsController(regionService *services.RegionService) *RegionsController {
	return &RegionsController{
		regionService: regionService,
	}
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
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (r *RegionsController) list(c *fiber.Ctx) error {
	// Get owner context for role-based filtering
	ownerCtx := utils.GetOwnerContext(c)
	if ownerCtx == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized - owner context not found"})
	}

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

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	regions, total, err := r.regionService.ListRegionsWithRole(filters, page, limit, ownerCtx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve regions",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve regions",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": regions,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (r *RegionsController) show(c *fiber.Ctx) error {
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

	regionId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	region, err := r.regionService.GetRegionByIDWithRole(uint(regionId), ownerCtx)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": region})
}

func (r *RegionsController) create(c *fiber.Ctx) error {
	var region models.Region
	if err := c.BodyParser(&region); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := r.regionService.CreateRegion(&region); err != nil {
		if err.Error() == "region with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Region created successfully",
		"flash_message": fiber.Map{
			"msg":  "Region created successfully",
			"type": "success",
		},
		"data": region,
	})
}

func (r *RegionsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	regionId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.Region
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

	if err := r.regionService.UpdateRegion(uint(regionId), updates); err != nil {
		if err.Error() == "region not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "region with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated region
	region, _ := r.regionService.GetRegionByID(uint(regionId))

	return c.JSON(fiber.Map{
		"message": "Region updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Region updated successfully",
			"type": "success",
		},
		"data": region,
	})
}

func (r *RegionsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	regionId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := r.regionService.DeleteRegion(uint(regionId)); err != nil {
		if err.Error() == "region not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Region deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Region deleted successfully",
			"type": "success",
		},
	})
}

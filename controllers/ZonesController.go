package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ZonesController struct {
	zoneService *services.ZoneService
}

func NewZonesController(zoneService *services.ZoneService) *ZonesController {
	return &ZonesController{
		zoneService: zoneService,
	}
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
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (z *ZonesController) list(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if regionID := c.Query("region_id"); regionID != "" {
		filters["region_id"] = regionID
	}
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

	zones, total, err := z.zoneService.ListZones(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve zones",
			"details": err.Error(),
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

func (z *ZonesController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	zoneId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	zone, err := z.zoneService.GetZoneByID(uint(zoneId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": zone})
}

func (z *ZonesController) create(c *fiber.Ctx) error {
	var zone models.Zone
	if err := c.BodyParser(&zone); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := z.zoneService.CreateZone(&zone); err != nil {
		if err.Error() == "zone with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Zone created successfully",
		"flash_message": fiber.Map{
			"msg":  "Zone created successfully",
			"type": "success",
		},
		"data": zone,
	})
}

func (z *ZonesController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	zoneId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.Zone
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
	if updateData.RegionId != nil {
		updates["region_id"] = *updateData.RegionId
	}

	if err := z.zoneService.UpdateZone(uint(zoneId), updates); err != nil {
		if err.Error() == "zone not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "zone with this code already exists" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated zone
	zone, _ := z.zoneService.GetZoneByID(uint(zoneId))

	return c.JSON(fiber.Map{
		"message": "Zone updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Zone updated successfully",
			"type": "success",
		},
		"data": zone,
	})
}

func (z *ZonesController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	zoneId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := z.zoneService.DeleteZone(uint(zoneId)); err != nil {
		if err.Error() == "zone not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Zone deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Zone deleted successfully",
			"type": "success",
		},
	})
}

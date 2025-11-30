package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ContactPersonsController struct {
	contactPersonService *services.ContactPersonService
}

func NewContactPersonsController(contactPersonService *services.ContactPersonService) *ContactPersonsController {
	return &ContactPersonsController{
		contactPersonService: contactPersonService,
	}
}

func (cp *ContactPersonsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return cp.list(c)
	case "show":
		return cp.show(c)
	case "create":
		return cp.create(c)
	case "update":
		return cp.update(c)
	case "delete":
		return cp.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (cp *ContactPersonsController) list(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if schoolID := c.Query("school_id"); schoolID != "" {
		filters["school_id"] = schoolID
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
	if relation := c.Query("relation"); relation != "" {
		filters["relation"] = relation
	}
	if email := c.Query("email"); email != "" {
		filters["email"] = email
	}
	if mobileNo := c.Query("mobile_no"); mobileNo != "" {
		filters["mobile_no"] = mobileNo
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	contactPersons, total, err := cp.contactPersonService.ListContactPersons(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve contact persons",
			"details": err.Error(),
			"flash_message": fiber.Map{
				"msg":  "Failed to retrieve contact persons",
				"type": "error",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": contactPersons,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (cp *ContactPersonsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	contactPersonId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	contactPerson, err := cp.contactPersonService.GetContactPersonByID(uint(contactPersonId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": contactPerson})
}

func (cp *ContactPersonsController) create(c *fiber.Ctx) error {
	var contactPerson models.ContactPerson
	if err := c.BodyParser(&contactPerson); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := cp.contactPersonService.CreateContactPerson(&contactPerson); err != nil {
		if err.Error() == "school not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "contact person with this email already exists for this school" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Contact person created successfully",
		"flash_message": fiber.Map{
			"msg":  "Contact person created successfully",
			"type": "success",
		},
		"data": contactPerson,
	})
}

func (cp *ContactPersonsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	contactPersonId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.ContactPerson
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
	if updateData.Email != nil {
		updates["email"] = *updateData.Email
	}
	if updateData.MobileNo != nil {
		updates["mobile_no"] = *updateData.MobileNo
	}
	if updateData.Relation != nil {
		updates["relation"] = *updateData.Relation
	}
	if updateData.SchoolId != nil {
		updates["school_id"] = *updateData.SchoolId
	}

	if err := cp.contactPersonService.UpdateContactPerson(uint(contactPersonId), updates); err != nil {
		if err.Error() == "contact person not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "school not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "contact person with this email already exists for this school" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated contact person
	contactPerson, _ := cp.contactPersonService.GetContactPersonByID(uint(contactPersonId))

	return c.JSON(fiber.Map{
		"message": "Contact person updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Contact person updated successfully",
			"type": "success",
		},
		"data": contactPerson,
	})
}

func (cp *ContactPersonsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	contactPersonId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := cp.contactPersonService.DeleteContactPerson(uint(contactPersonId)); err != nil {
		if err.Error() == "contact person not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Contact person deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Contact person deleted successfully",
			"type": "success",
		},
	})
}

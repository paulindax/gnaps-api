package controllers

import (
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ContactPersonsController struct {
}

func init() {
	RegisterController("contact_persons", &ContactPersonsController{})
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
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// list retrieves all contact persons
func (cp *ContactPersonsController) list(c *fiber.Ctx) error {
	var contactPersons []models.ContactPerson

	// Base query
	query := DB.Model(&models.ContactPerson{})

	// Optional filter by school_id
	if schoolID := c.Query("school_id"); schoolID != "" {
		query = query.Where("school_id = ?", schoolID)
	}

	// Optional search by first name
	if firstName := c.Query("first_name"); firstName != "" {
		query = query.Where("first_name LIKE ?", "%"+firstName+"%")
	}

	// Optional search by last name
	if lastName := c.Query("last_name"); lastName != "" {
		query = query.Where("last_name LIKE ?", "%"+lastName+"%")
	}

	// Optional search by full name (searches both first and last name)
	if name := c.Query("name"); name != "" {
		query = query.Where("CONCAT(first_name, ' ', last_name) LIKE ?", "%"+name+"%")
	}

	// Optional filter by relation
	if relation := c.Query("relation"); relation != "" {
		query = query.Where("relation = ?", relation)
	}

	// Optional filter by email
	if email := c.Query("email"); email != "" {
		query = query.Where("email = ?", email)
	}

	// Optional filter by mobile_no
	if mobileNo := c.Query("mobile_no"); mobileNo != "" {
		query = query.Where("mobile_no = ?", mobileNo)
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Count(&total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&contactPersons)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve contact persons",
			"details": result.Error.Error(),
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

// show retrieves a single contact person by ID
func (cp *ContactPersonsController) show(c *fiber.Ctx) error {
	// Get ID from URL params or query string
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var contactPerson models.ContactPerson
	result := DB.First(&contactPerson, id)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Contact person not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": contactPerson,
	})
}

// create creates a new contact person
func (cp *ContactPersonsController) create(c *fiber.Ctx) error {
	var contactPerson models.ContactPerson

	if err := c.BodyParser(&contactPerson); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if contactPerson.SchoolId == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "School ID is required",
		})
	}

	if contactPerson.FirstName == nil || *contactPerson.FirstName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "First name is required",
		})
	}

	if contactPerson.LastName == nil || *contactPerson.LastName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Last name is required",
		})
	}

	// Verify that the school exists
	var school models.School
	if err := DB.Where("id = ? AND is_deleted = ?", contactPerson.SchoolId, false).First(&school).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid school ID - School does not exist",
		})
	}

	// Check if email already exists for this school (if provided)
	if contactPerson.Email != nil && *contactPerson.Email != "" {
		var existing models.ContactPerson
		if err := DB.Where("email = ? AND school_id = ?", contactPerson.Email, contactPerson.SchoolId).First(&existing).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Contact person with this email already exists for this school",
			})
		}
	}

	result := DB.Create(&contactPerson)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create contact person",
			"details": result.Error.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Contact person created successfully",
		"data":    contactPerson,
	})
}

// update updates an existing contact person
func (cp *ContactPersonsController) update(c *fiber.Ctx) error {
	// Get ID from URL params or query string
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var contactPerson models.ContactPerson
	result := DB.First(&contactPerson, id)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Contact person not found",
		})
	}

	var updateData models.ContactPerson
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// If school_id is being changed, verify the new school exists
	if updateData.SchoolId != nil {
		if contactPerson.SchoolId == nil || *updateData.SchoolId != *contactPerson.SchoolId {
			var school models.School
			if err := DB.Where("id = ? AND is_deleted = ?", updateData.SchoolId, false).First(&school).Error; err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": "Invalid school ID - School does not exist",
				})
			}
		}
	}

	// Determine which school_id to use for email check
	schoolIDForCheck := contactPerson.SchoolId
	if updateData.SchoolId != nil {
		schoolIDForCheck = updateData.SchoolId
	}

	// Check if email is being changed and if new email already exists for the school
	if updateData.Email != nil && *updateData.Email != "" {
		if contactPerson.Email == nil || *updateData.Email != *contactPerson.Email {
			var existingEmail models.ContactPerson
			if err := DB.Where("email = ? AND school_id = ? AND id != ?", updateData.Email, schoolIDForCheck, id).First(&existingEmail).Error; err == nil {
				return c.Status(409).JSON(fiber.Map{
					"error": "Contact person with this email already exists for this school",
				})
			}
		}
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if updateData.SchoolId != nil {
		updates["school_id"] = updateData.SchoolId
	}
	if updateData.FirstName != nil {
		updates["first_name"] = updateData.FirstName
	}
	if updateData.LastName != nil {
		updates["last_name"] = updateData.LastName
	}
	if updateData.Relation != nil {
		updates["relation"] = updateData.Relation
	}
	if updateData.Email != nil {
		updates["email"] = updateData.Email
	}
	if updateData.MobileNo != nil {
		updates["mobile_no"] = updateData.MobileNo
	}

	result = DB.Model(&contactPerson).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update contact person",
			"details": result.Error.Error(),
		})
	}

	// Fetch updated record
	DB.First(&contactPerson, id)

	return c.JSON(fiber.Map{
		"message": "Contact person updated successfully",
		"data":    contactPerson,
	})
}

// delete soft deletes a contact person
func (cp *ContactPersonsController) delete(c *fiber.Ctx) error {
	// Get ID from URL params or query string
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var contactPerson models.ContactPerson
	result := DB.First(&contactPerson, id)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Contact person not found",
		})
	}

	// Soft delete using GORM's built-in soft delete (DeletedAt)
	result = DB.Delete(&contactPerson)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete contact person",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Contact person deleted successfully",
	})
}

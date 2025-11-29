package controllers

import (
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ExecutivesController struct {
}

func init() {
	RegisterController("executives", &ExecutivesController{})
}

func (e *ExecutivesController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return e.list(c)
	case "show":
		return e.show(c)
	case "create":
		return e.create(c)
	case "update":
		return e.update(c)
	case "delete":
		return e.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// list retrieves all non-deleted executives
func (e *ExecutivesController) list(c *fiber.Ctx) error {
	var executives []models.Executive

	// Query parameters for filtering
	query := DB.Where("is_deleted = ?", false)

	// Optional filter by position_id
	if positionID := c.Query("position_id"); positionID != "" {
		query = query.Where("position_id = ?", positionID)
	}

	// Optional filter by gender
	if gender := c.Query("gender"); gender != "" {
		query = query.Where("gender = ?", gender)
	}

	// Optional search by first name
	if firstName := c.Query("first_name"); firstName != "" {
		query = query.Where("first_name LIKE ?", "%"+firstName+"%")
	}

	// Optional search by last name
	if lastName := c.Query("last_name"); lastName != "" {
		query = query.Where("last_name LIKE ?", "%"+lastName+"%")
	}

	// Optional search by full name
	if name := c.Query("name"); name != "" {
		query = query.Where("CONCAT(first_name, ' ', IFNULL(middle_name, ''), ' ', last_name) LIKE ?", "%"+name+"%")
	}

	// Optional filter by executive_no
	if executiveNo := c.Query("executive_no"); executiveNo != "" {
		query = query.Where("executive_no LIKE ?", "%"+executiveNo+"%")
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
	query.Model(&models.Executive{}).Count(&total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&executives)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve executives",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": executives,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// show retrieves a single executive by ID
func (e *ExecutivesController) show(c *fiber.Ctx) error {
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

	var executive models.Executive
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&executive)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Executive not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": executive,
	})
}

// create creates a new executive
func (e *ExecutivesController) create(c *fiber.Ctx) error {
	var executive models.Executive

	if err := c.BodyParser(&executive); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if executive.ExecutiveNo == nil || *executive.ExecutiveNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Executive number is required",
		})
	}

	if executive.FirstName == nil || *executive.FirstName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "First name is required",
		})
	}

	if executive.LastName == nil || *executive.LastName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Last name is required",
		})
	}

	// Check if executive_no already exists
	var existing models.Executive
	if err := DB.Where("executive_no = ? AND is_deleted = ?", executive.ExecutiveNo, false).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Executive with this executive number already exists",
		})
	}

	// Check if email already exists (if provided)
	if executive.Email != nil && *executive.Email != "" {
		var existingEmail models.Executive
		if err := DB.Where("email = ? AND is_deleted = ?", executive.Email, false).First(&existingEmail).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "Executive with this email already exists",
			})
		}
	}

	// Validate gender if provided
	if executive.Gender != nil && *executive.Gender != "" {
		validGenders := []string{"Male", "Female", "Other"}
		isValidGender := false
		for _, g := range validGenders {
			if *executive.Gender == g {
				isValidGender = true
				break
			}
		}
		if !isValidGender {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid gender. Must be one of: Male, Female, Other",
			})
		}
	}

	// Set default values
	falseVal := false
	executive.IsDeleted = &falseVal

	result := DB.Create(&executive)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create executive",
			"details": result.Error.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Executive created successfully",
		"data":    executive,
	})
}

// update updates an existing executive
func (e *ExecutivesController) update(c *fiber.Ctx) error {
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

	var executive models.Executive
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&executive)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Executive not found",
		})
	}

	var updateData models.Executive
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Check if executive_no is being changed and if new executive_no already exists
	if updateData.ExecutiveNo != nil && *updateData.ExecutiveNo != "" {
		if executive.ExecutiveNo == nil || *updateData.ExecutiveNo != *executive.ExecutiveNo {
			var existing models.Executive
			if err := DB.Where("executive_no = ? AND id != ? AND is_deleted = ?", updateData.ExecutiveNo, id, false).First(&existing).Error; err == nil {
				return c.Status(409).JSON(fiber.Map{
					"error": "Executive with this executive number already exists",
				})
			}
		}
	}

	// Check if email is being changed and if new email already exists
	if updateData.Email != nil && *updateData.Email != "" {
		if executive.Email == nil || *updateData.Email != *executive.Email {
			var existingEmail models.Executive
			if err := DB.Where("email = ? AND id != ? AND is_deleted = ?", updateData.Email, id, false).First(&existingEmail).Error; err == nil {
				return c.Status(409).JSON(fiber.Map{
					"error": "Executive with this email already exists",
				})
			}
		}
	}

	// Validate gender if being changed
	if updateData.Gender != nil && *updateData.Gender != "" {
		validGenders := []string{"Male", "Female", "Other"}
		isValidGender := false
		for _, g := range validGenders {
			if *updateData.Gender == g {
				isValidGender = true
				break
			}
		}
		if !isValidGender {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid gender. Must be one of: Male, Female, Other",
			})
		}
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if updateData.ExecutiveNo != nil {
		updates["executive_no"] = updateData.ExecutiveNo
	}
	if updateData.FirstName != nil {
		updates["first_name"] = updateData.FirstName
	}
	if updateData.MiddleName != nil {
		updates["middle_name"] = updateData.MiddleName
	}
	if updateData.LastName != nil {
		updates["last_name"] = updateData.LastName
	}
	if updateData.Gender != nil {
		updates["gender"] = updateData.Gender
	}
	if updateData.PositionId != nil {
		updates["position_id"] = updateData.PositionId
	}
	if !updateData.DateOfBirth.IsZero() {
		updates["date_of_birth"] = updateData.DateOfBirth
	}
	if updateData.MobileNo != nil {
		updates["mobile_no"] = updateData.MobileNo
	}
	if updateData.Email != nil {
		updates["email"] = updateData.Email
	}
	if updateData.PhotoFileName != nil {
		updates["photo_file_name"] = updateData.PhotoFileName
	}
	if updateData.UserId != nil {
		updates["user_id"] = updateData.UserId
	}
	if updateData.AssignedZoneIds != nil {
		updates["assigned_zone_ids"] = updateData.AssignedZoneIds
	}
	if updateData.AssignedRegionsIds != nil {
		updates["assigned_regions_ids"] = updateData.AssignedRegionsIds
	}

	result = DB.Model(&executive).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update executive",
			"details": result.Error.Error(),
		})
	}

	// Fetch updated record
	DB.First(&executive, id)

	return c.JSON(fiber.Map{
		"message": "Executive updated successfully",
		"data":    executive,
	})
}

// delete soft deletes an executive
func (e *ExecutivesController) delete(c *fiber.Ctx) error {
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

	var executive models.Executive
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&executive)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Executive not found",
		})
	}

	// Soft delete by setting is_deleted flag
	trueVal := true
	result = DB.Model(&executive).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete executive",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Executive deleted successfully",
	})
}

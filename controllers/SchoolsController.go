package controllers

import (
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SchoolsController struct {
}

func init() {
	RegisterController("schools", &SchoolsController{})
}

func (s *SchoolsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return s.list(c)
	case "show":
		return s.show(c)
	case "create":
		return s.create(c)
	case "update":
		return s.update(c)
	case "delete":
		return s.delete(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// list retrieves all non-deleted schools
func (s *SchoolsController) list(c *fiber.Ctx) error {
	var schools []models.School

	// Query parameters for filtering
	query := DB.Where("is_deleted = ?", false)

	// Optional filter by zone_id
	if zoneID := c.Query("zone_id"); zoneID != "" {
		query = query.Where("zone_id = ?", zoneID)
	}

	// Optional search by name
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Optional filter by member_no
	if memberNo := c.Query("member_no"); memberNo != "" {
		query = query.Where("member_no LIKE ?", "%"+memberNo+"%")
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
	query.Model(&models.School{}).Count(&total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&schools)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve schools",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": schools,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// show retrieves a single school by ID
func (s *SchoolsController) show(c *fiber.Ctx) error {
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

	var school models.School
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&school)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "School not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": school,
	})
}

// create creates a new school
func (s *SchoolsController) create(c *fiber.Ctx) error {
	var school models.School

	if err := c.BodyParser(&school); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if school.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	if school.MemberNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Member number is required",
		})
	}

	// Verify that the zone exists if zone_id is provided
	if school.ZoneId != nil {
		var zone models.Zone
		if err := DB.Where("id = ? AND is_deleted = ?", school.ZoneId, false).First(&zone).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid zone ID - Zone does not exist",
			})
		}
	}

	// Check if member_no already exists
	var existing models.School
	if err := DB.Where("member_no = ? AND is_deleted = ?", school.MemberNo, false).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "School with this member number already exists",
		})
	}

	// Check if email already exists (if provided)
	if school.Email != nil && *school.Email != "" {
		var existingEmail models.School
		if err := DB.Where("email = ? AND is_deleted = ?", school.Email, false).First(&existingEmail).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "School with this email already exists",
			})
		}
	}

	// Set default values
	falseVal := false
	school.IsDeleted = &falseVal

	result := DB.Create(&school)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create school",
			"details": result.Error.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "School created successfully",
		"data":    school,
	})
}

// update updates an existing school
func (s *SchoolsController) update(c *fiber.Ctx) error {
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

	var school models.School
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&school)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "School not found",
		})
	}

	var updateData models.School
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// If zone_id is being changed, verify the new zone exists
	if updateData.ZoneId != nil {
		if school.ZoneId == nil || *updateData.ZoneId != *school.ZoneId {
			var zone models.Zone
			if err := DB.Where("id = ? AND is_deleted = ?", updateData.ZoneId, false).First(&zone).Error; err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": "Invalid zone ID - Zone does not exist",
				})
			}
		}
	}

	// Check if member_no is being changed and if new member_no already exists
	if updateData.MemberNo != "" && updateData.MemberNo != school.MemberNo {
		var existing models.School
		if err := DB.Where("member_no = ? AND id != ? AND is_deleted = ?", updateData.MemberNo, id, false).First(&existing).Error; err == nil {
			return c.Status(409).JSON(fiber.Map{
				"error": "School with this member number already exists",
			})
		}
	}

	// Check if email is being changed and if new email already exists
	if updateData.Email != nil && *updateData.Email != "" {
		if school.Email == nil || *updateData.Email != *school.Email {
			var existingEmail models.School
			if err := DB.Where("email = ? AND id != ? AND is_deleted = ?", updateData.Email, id, false).First(&existingEmail).Error; err == nil {
				return c.Status(409).JSON(fiber.Map{
					"error": "School with this email already exists",
				})
			}
		}
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if updateData.Name != "" {
		updates["name"] = updateData.Name
	}
	if updateData.MemberNo != "" {
		updates["member_no"] = updateData.MemberNo
	}
	if updateData.ZoneId != nil {
		updates["zone_id"] = updateData.ZoneId
	}
	if !updateData.JoiningDate.IsZero() {
		updates["joining_date"] = updateData.JoiningDate
	}
	if !updateData.DateOfEstablishment.IsZero() {
		updates["date_of_establishment"] = updateData.DateOfEstablishment
	}
	if updateData.Address != nil {
		updates["address"] = updateData.Address
	}
	if updateData.Location != nil {
		updates["location"] = updateData.Location
	}
	if updateData.MobileNo != nil {
		updates["mobile_no"] = updateData.MobileNo
	}
	if updateData.Email != nil {
		updates["email"] = updateData.Email
	}
	if updateData.GpsAddress != nil {
		updates["gps_address"] = updateData.GpsAddress
	}
	if updateData.UserId != nil {
		updates["user_id"] = updateData.UserId
	}

	result = DB.Model(&school).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update school",
			"details": result.Error.Error(),
		})
	}

	// Fetch updated record
	DB.First(&school, id)

	return c.JSON(fiber.Map{
		"message": "School updated successfully",
		"data":    school,
	})
}

// delete soft deletes a school
func (s *SchoolsController) delete(c *fiber.Ctx) error {
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

	var school models.School
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&school)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "School not found",
		})
	}

	// Soft delete by setting is_deleted flag
	trueVal := true
	result = DB.Model(&school).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete school",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "School deleted successfully",
	})
}

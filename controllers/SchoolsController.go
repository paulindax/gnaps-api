package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SchoolsController struct {
	schoolService *services.SchoolService
}

func NewSchoolsController(schoolService *services.SchoolService) *SchoolsController {
	return &SchoolsController{
		schoolService: schoolService,
	}
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
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

func (s *SchoolsController) list(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if zoneID := c.Query("zone_id"); zoneID != "" {
		filters["zone_id"] = zoneID
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if memberNo := c.Query("member_no"); memberNo != "" {
		filters["member_no"] = memberNo
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

	schools, total, err := s.schoolService.ListSchools(filters, page, limit)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to retrieve schools")
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

func (s *SchoolsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	schoolId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	school, err := s.schoolService.GetSchoolByID(uint(schoolId))
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return c.JSON(fiber.Map{"data": school})
}

func (s *SchoolsController) create(c *fiber.Ctx) error {
	var school models.School
	if err := c.BodyParser(&school); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := s.schoolService.CreateSchool(&school); err != nil {
		// Handle specific error types
		errMsg := err.Error()
		if errMsg == "school with this member number already exists" ||
			errMsg == "school with this email already exists" {
			return utils.ConflictResponse(c, errMsg)
		}
		return utils.ValidationErrorResponse(c, errMsg)
	}

	return utils.SuccessResponseWithStatus(c, 201, school, "School created successfully")
}

func (s *SchoolsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	schoolId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	var updateData models.School
	if err := c.BodyParser(&updateData); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Build updates map
	updates := make(map[string]interface{})
	if updateData.Name != "" {
		updates["name"] = updateData.Name
	}
	if updateData.MemberNo != "" {
		updates["member_no"] = updateData.MemberNo
	}
	if updateData.ZoneId != nil {
		updates["zone_id"] = *updateData.ZoneId
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
		updates["email"] = *updateData.Email
	}
	if updateData.GpsAddress != nil {
		updates["gps_address"] = updateData.GpsAddress
	}
	if updateData.UserId != nil {
		updates["user_id"] = updateData.UserId
	}

	if err := s.schoolService.UpdateSchool(uint(schoolId), updates); err != nil {
		errMsg := err.Error()
		if errMsg == "school not found" {
			return utils.NotFoundResponse(c, errMsg)
		}
		if errMsg == "school with this member number already exists" ||
			errMsg == "school with this email already exists" {
			return utils.ConflictResponse(c, errMsg)
		}
		return utils.ValidationErrorResponse(c, errMsg)
	}

	// Get updated school
	school, _ := s.schoolService.GetSchoolByID(uint(schoolId))

	return utils.SuccessResponse(c, school, "School updated successfully")
}

func (s *SchoolsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	schoolId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	if err := s.schoolService.DeleteSchool(uint(schoolId)); err != nil {
		if err.Error() == "school not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, nil, "School deleted successfully")
}

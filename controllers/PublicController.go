package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PublicController struct {
	regionRepo        *repositories.RegionRepository
	zoneRepo          *repositories.ZoneRepository
	schoolRepo        *repositories.SchoolRepository
	contactPersonRepo *repositories.ContactPersonRepository
	db                *gorm.DB
}

// NewPublicController creates a new instance of PublicController
func NewPublicController(
	regionRepo *repositories.RegionRepository,
	zoneRepo *repositories.ZoneRepository,
	schoolRepo *repositories.SchoolRepository,
	contactPersonRepo *repositories.ContactPersonRepository,
	db *gorm.DB,
) *PublicController {
	return &PublicController{
		regionRepo:        regionRepo,
		zoneRepo:          zoneRepo,
		schoolRepo:        schoolRepo,
		contactPersonRepo: contactPersonRepo,
		db:                db,
	}
}

// Handle routes the action to the appropriate handler method
func (p *PublicController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "regions":
		return p.listRegions(c)
	case "zones":
		return p.listZones(c)
	case "register":
		return p.registerSchool(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// listRegions returns all regions (for school registration dropdown)
func (p *PublicController) listRegions(c *fiber.Ctx) error {
	filters := make(map[string]interface{})
	regions, _, err := p.regionRepo.List(filters, 1, 100)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to load regions")
	}
	return c.JSON(fiber.Map{
		"data": regions,
	})
}

// listZones returns all zones (for school registration dropdown)
func (p *PublicController) listZones(c *fiber.Ctx) error {
	filters := make(map[string]interface{})
	zones, _, err := p.zoneRepo.List(filters, 1, 1000)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to load zones")
	}
	return c.JSON(fiber.Map{
		"data": zones,
	})
}

// SchoolRegistrationRequest represents the request body for school registration
type SchoolRegistrationRequest struct {
	Name                string               `json:"name"`
	ZoneID              int64                `json:"zone_id"`
	Address             string               `json:"address"`
	Location            string               `json:"location"`
	MobileNo            string               `json:"mobile_no"`
	Email               string               `json:"email"`
	GpsAddress          string               `json:"gps_address"`
	DateOfEstablishment string               `json:"date_of_establishment"`
	ContactPersons      []ContactPersonInput `json:"contact_persons"`
}

type ContactPersonInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	MobileNo  string `json:"mobile_no"`
	Relation  string `json:"relation"`
}

// registerSchool handles public school registration
func (p *PublicController) registerSchool(c *fiber.Ctx) error {
	var req SchoolRegistrationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Validate required fields
	if req.Name == "" {
		return utils.ValidationErrorResponse(c, "School name is required")
	}
	if req.ZoneID == 0 {
		return utils.ValidationErrorResponse(c, "Zone is required")
	}
	if len(req.ContactPersons) == 0 {
		return utils.ValidationErrorResponse(c, "At least one contact person is required")
	}

	// Validate contact person has required fields
	hasValidContact := false
	for _, cp := range req.ContactPersons {
		if cp.FirstName != "" && cp.LastName != "" && cp.MobileNo != "" {
			hasValidContact = true
			break
		}
	}
	if !hasValidContact {
		return utils.ValidationErrorResponse(c, "Contact person must have first name, last name, and phone number")
	}

	// Verify zone exists
	zone, err := p.zoneRepo.FindByID(uint(req.ZoneID))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid zone selected")
	}

	// Generate member number
	memberNo, err := p.schoolRepo.GetNextMemberNoForZone(req.ZoneID)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to generate member number")
	}

	// Create school - Name and MemberNo are non-pointer strings in the model
	school := &models.School{
		Name:     req.Name,
		ZoneId:   &req.ZoneID,
		MemberNo: memberNo,
	}

	// Set optional fields
	if req.Address != "" {
		school.Address = &req.Address
	}
	if req.Location != "" {
		school.Location = &req.Location
	}
	if req.MobileNo != "" {
		school.MobileNo = &req.MobileNo
	}
	if req.Email != "" {
		school.Email = &req.Email
	}
	if req.GpsAddress != "" {
		school.GpsAddress = &req.GpsAddress
	}

	// Parse date of establishment if provided
	if req.DateOfEstablishment != "" {
		parsedDate, err := time.Parse("2006-01-02", req.DateOfEstablishment)
		if err == nil {
			school.DateOfEstablishment = parsedDate
		}
	}

	// Set joining date to today
	school.JoiningDate = time.Now()

	// Set is_deleted to false
	isDeleted := false
	school.IsDeleted = &isDeleted

	// Create the school
	if err := p.schoolRepo.Create(school); err != nil {
		return utils.ServerErrorResponse(c, "Failed to register school. Please try again.")
	}

	// Create contact persons
	schoolID := int64(school.ID)
	for _, cpInput := range req.ContactPersons {
		if cpInput.FirstName == "" || cpInput.LastName == "" {
			continue
		}

		contactPerson := &models.ContactPerson{
			SchoolId:  &schoolID,
			FirstName: &cpInput.FirstName,
			LastName:  &cpInput.LastName,
		}
		if cpInput.Email != "" {
			contactPerson.Email = &cpInput.Email
		}
		if cpInput.MobileNo != "" {
			contactPerson.MobileNo = &cpInput.MobileNo
		}
		if cpInput.Relation != "" {
			contactPerson.Relation = &cpInput.Relation
		}

		if err := p.contactPersonRepo.Create(contactPerson); err != nil {
			// Log but don't fail - school is already created
			fmt.Printf("Failed to create contact person for school %d: %v\n", school.ID, err)
		}
	}

	zoneName := ""
	if zone.Name != nil {
		zoneName = *zone.Name
	}

	return utils.SuccessResponseWithStatus(c, 201, fiber.Map{
		"school_id": school.ID,
		"member_no": memberNo,
		"zone_name": zoneName,
		"status":    "pending",
	}, "School registration submitted successfully! Your application is pending review.")
}

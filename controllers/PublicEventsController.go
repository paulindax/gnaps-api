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

type PublicEventsController struct {
	eventRepo        *repositories.EventRepository
	registrationRepo *repositories.RegistrationRepository
	schoolRepo       *repositories.SchoolRepository
	db               *gorm.DB
}

// NewPublicEventsController creates a new instance of PublicEventsController
func NewPublicEventsController(
	eventRepo *repositories.EventRepository,
	registrationRepo *repositories.RegistrationRepository,
	schoolRepo *repositories.SchoolRepository,
	db *gorm.DB,
) *PublicEventsController {
	return &PublicEventsController{
		eventRepo:        eventRepo,
		registrationRepo: registrationRepo,
		schoolRepo:       schoolRepo,
		db:               db,
	}
}

// Handle routes the action to the appropriate handler method
func (p *PublicEventsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "view":
		return p.viewEvent(c)
	case "register":
		return p.registerForEvent(c)
	case "schools":
		return p.searchSchools(c)
	case "school-balance":
		return p.getSchoolBalance(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// viewEvent returns event details by registration code (no auth required)
func (p *PublicEventsController) viewEvent(c *fiber.Ctx) error {
	code := c.Params("id")
	if code == "" {
		return utils.ValidationErrorResponse(c, "Registration code is required")
	}

	event, err := p.eventRepo.FindByCode(code)
	if err != nil {
		return utils.NotFoundResponse(c, "Event not found or registration link is invalid")
	}

	// Check if event is active
	if event.Status != nil && *event.Status != "upcoming" && *event.Status != "active" {
		return utils.ErrorResponse(c, 400, "Event registration is not currently available")
	}

	// Get registration count for this event
	count, err := p.registrationRepo.CountByEvent(event.ID)
	if err == nil {
		event.RegisteredCount = int(count)
	}

	return c.JSON(event)
}

// registerForEvent creates a new event registration (no auth required)
func (p *PublicEventsController) registerForEvent(c *fiber.Ctx) error {
	code := c.Params("id")
	if code == "" {
		return utils.ValidationErrorResponse(c, "Registration code is required")
	}

	// Get event by code
	event, err := p.eventRepo.FindByCode(code)
	if err != nil {
		return utils.NotFoundResponse(c, "Event not found or registration link is invalid")
	}

	// Check if event is active
	if event.Status != nil && *event.Status != "upcoming" && *event.Status != "active" {
		return utils.ErrorResponse(c, 400, "Event registration is not currently available")
	}

	// Parse registration data
	var registration models.EventRegistration
	if err := c.BodyParser(&registration); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Validate required fields
	if registration.SchoolId == 0 {
		return utils.ValidationErrorResponse(c, "School selection is required")
	}

	// Verify school exists
	_, err = p.schoolRepo.FindByID(uint(registration.SchoolId))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid school selected")
	}

	// Set event ID
	registration.EventId = int64(event.ID)

	// Set registration date
	registration.RegistrationDate = time.Now()

	// Set payment status
	if event.IsPaid != nil && *event.IsPaid {
		// For paid events, validate payment info
		if registration.PaymentMethod == nil || *registration.PaymentMethod == "" {
			return utils.ValidationErrorResponse(c, "Payment method is required for paid events")
		}
		if registration.PaymentPhone == nil || *registration.PaymentPhone == "" {
			return utils.ValidationErrorResponse(c, "Mobile money number is required for paid events")
		}

		pending := "pending"
		registration.PaymentStatus = &pending
	} else {
		// For free events, set payment status to confirmed
		confirmed := "confirmed"
		registration.PaymentStatus = &confirmed
	}

	// Set default number of attendees if not provided
	if registration.NumberOfAttendees == nil {
		defaultAttendees := 1
		registration.NumberOfAttendees = &defaultAttendees
	}

	// Create registration
	if err := p.registrationRepo.Create(&registration); err != nil {
		return utils.ServerErrorResponse(c, "Failed to create registration. Please try again.")
	}

	return utils.SuccessResponseWithStatus(c, 201, registration, "Registration submitted successfully! We look forward to seeing you at the event.")
}

// searchSchools searches for schools by keyword (no auth required)
func (p *PublicEventsController) searchSchools(c *fiber.Ctx) error {
	keyword := c.Query("search")
	if keyword == "" {
		return c.JSON([]models.School{})
	}

	schools, err := p.schoolRepo.Search(keyword, 20)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to search schools. Please try again.")
	}

	return c.JSON(schools)
}

// getSchoolBalance returns the last unpaid bill balance for a school (no auth required)
func (p *PublicEventsController) getSchoolBalance(c *fiber.Ctx) error {
	schoolID := c.Params("id")
	if schoolID == "" {
		return utils.ValidationErrorResponse(c, "School ID is required")
	}

	// Get the last unpaid bill for the school
	var schoolBill models.SchoolBill
	err := p.db.Where("school_id = ? AND (is_paid IS NULL OR is_paid = ?)", schoolID, false).
		Order("created_at DESC").
		First(&schoolBill).Error

	if err != nil {
		// If no bill found, return zero balance
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{
				"balance": 0,
				"has_balance": false,
			})
		}
		return utils.ServerErrorResponse(c, "Failed to retrieve school balance")
	}

	balance := 0.0
	if schoolBill.Balance != nil {
		balance = *schoolBill.Balance
	}

	return c.JSON(fiber.Map{
		"balance": balance,
		"has_balance": balance > 0,
		"bill_id": schoolBill.ID,
	})
}

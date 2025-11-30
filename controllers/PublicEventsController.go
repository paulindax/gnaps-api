package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PublicEventsController struct {
	eventRepo        *repositories.EventRepository
	registrationRepo *repositories.RegistrationRepository
	schoolRepo       *repositories.SchoolRepository
}

// NewPublicEventsController creates a new instance of PublicEventsController
func NewPublicEventsController(
	eventRepo *repositories.EventRepository,
	registrationRepo *repositories.RegistrationRepository,
	schoolRepo *repositories.SchoolRepository,
) *PublicEventsController {
	return &PublicEventsController{
		eventRepo:        eventRepo,
		registrationRepo: registrationRepo,
		schoolRepo:       schoolRepo,
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
		return c.Status(400).JSON(fiber.Map{
			"error": "registration code is required",
		})
	}

	event, err := p.eventRepo.FindByCode(code)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "event not found or registration link is invalid",
		})
	}

	// Check if event is active
	if event.Status != nil && *event.Status != "upcoming" && *event.Status != "active" {
		return c.Status(400).JSON(fiber.Map{
			"error": "event registration is not currently available",
		})
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
		return c.Status(400).JSON(fiber.Map{
			"error": "registration code is required",
		})
	}

	// Get event by code
	event, err := p.eventRepo.FindByCode(code)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "event not found or registration link is invalid",
		})
	}

	// Check if event is active
	if event.Status != nil && *event.Status != "upcoming" && *event.Status != "active" {
		return c.Status(400).JSON(fiber.Map{
			"error": "event registration is not currently available",
		})
	}

	// Parse registration data
	var registration models.EventRegistration
	if err := c.BodyParser(&registration); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if registration.SchoolId == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "school_id is required",
		})
	}

	// Verify school exists
	_, err = p.schoolRepo.FindByID(uint(*registration.SchoolId))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid school_id",
		})
	}

	// Set event ID
	eventID := int64(event.ID)
	registration.EventId = &eventID

	// Set registration date
	now := time.Now().Format("2006-01-02 15:04:05")
	registration.RegistrationDate = &now

	// Set payment status
	if event.IsPaid != nil && *event.IsPaid {
		// For paid events, validate payment info
		if registration.PaymentMethod == nil || *registration.PaymentMethod == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "payment_method is required for paid events",
			})
		}
		if registration.PaymentPhone == nil || *registration.PaymentPhone == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "payment_phone is required for paid events",
			})
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
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to create registration",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    registration,
		"flash_message": fiber.Map{
			"msg":  "Registration submitted successfully",
			"type": "success",
		},
	})
}

// searchSchools searches for schools by keyword (no auth required)
func (p *PublicEventsController) searchSchools(c *fiber.Ctx) error {
	keyword := c.Query("search")
	if keyword == "" {
		return c.JSON([]models.School{})
	}

	schools, err := p.schoolRepo.Search(keyword, 20)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to search schools",
		})
	}

	return c.JSON(schools)
}

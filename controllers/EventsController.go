package controllers

import (
	"fmt"
	"gnaps-api/models"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type EventsController struct {
}

func init() {
	RegisterController("events", &EventsController{})
	RegisterController("event-registrations", &EventsController{})
}

func (e *EventsController) Handle(action string, c *fiber.Ctx) error {
	controller := c.Params("controller")

	// Route to appropriate handler based on controller name
	if controller == "event-registrations" {
		return e.handleEventRegistrations(action, c)
	}

	// Check if this is a nested resource route (e.g., /events/:id/registrations or /events/:id/register)
	// In this case, 'action' will be the event ID and 'id' param will be the actual action
	id := c.Params("id")
	if id != "" {
		// Check if action is numeric (event ID)
		if eventId, err := strconv.ParseUint(action, 10, 64); err == nil {
			// This is a nested route like /events/:id/registrations or /events/:id/register
			switch id {
			case "registrations":
				return e.getEventRegistrations(c, uint(eventId))
			case "register":
				return e.registerForEvent(c, uint(eventId))
			default:
				// Fall through to normal handling (e.g., /events/show/:id)
			}
		}
	}

	// Handle standard event actions
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

func (e *EventsController) handleEventRegistrations(action string, c *fiber.Ctx) error {
	switch action {
	case "my":
		return e.getMyRegistrations(c)
	default:
		// Handle actions with ID parameter
		id := c.Params("id")
		if id == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "registration ID is required",
			})
		}

		registrationId, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid registration ID",
			})
		}

		switch action {
		case "payment":
			return e.updatePaymentStatus(c, uint(registrationId))
		default:
			// DELETE /event-registrations/:id
			if c.Method() == "DELETE" {
				return e.cancelRegistration(c, uint(registrationId))
			}
			return c.Status(404).JSON(fiber.Map{
				"error": fmt.Sprintf("unknown action %s", action),
			})
		}
	}
}

// ==================== EVENT ENDPOINTS ====================

// List all events with pagination and filters
func (e *EventsController) list(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status")
	isPaid := c.Query("is_paid")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	offset := (page - 1) * limit

	query := DB.Model(&models.Event{}).Where("is_deleted = ?", false)

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if isPaid != "" {
		isPaidBool := isPaid == "true"
		query = query.Where("is_paid = ?", isPaidBool)
	}
	if fromDate != "" {
		query = query.Where("start_date >= ?", fromDate)
	}
	if toDate != "" {
		query = query.Where("start_date <= ?", toDate)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get events
	var events []models.Event
	if err := query.Order("start_date DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch events",
		})
	}

	// Get registered count for each event
	for i := range events {
		var count int64
		DB.Model(&models.EventRegistration{}).
			Where("event_id = ? AND is_deleted = ?", events[i].ID, false).
			Count(&count)
		events[i].RegisteredCount = int(count)
	}

	return c.JSON(fiber.Map{
		"data": events,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// Show single event
func (e *EventsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	eventId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid event ID",
		})
	}

	var event models.Event
	if err := DB.Where("id = ? AND is_deleted = ?", eventId, false).First(&event).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "event not found",
		})
	}

	// Get registered count
	var count int64
	DB.Model(&models.EventRegistration{}).
		Where("event_id = ? AND is_deleted = ?", event.ID, false).
		Count(&count)
	event.RegisteredCount = int(count)

	return c.JSON(event)
}

// Create new event
func (e *EventsController) create(c *fiber.Ctx) error {
	userId, _ := c.Locals("user_id").(uint)

	var event models.Event
	if err := c.BodyParser(&event); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Set created_by to current user
	createdBy := int64(userId)
	event.CreatedBy = &createdBy

	// Set defaults
	isDeleted := false
	event.IsDeleted = &isDeleted

	if event.Status == nil {
		status := "draft"
		event.Status = &status
	}

	if err := DB.Create(&event).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to create event",
		})
	}

	return c.Status(201).JSON(event)
}

// Update existing event
func (e *EventsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	eventId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid event ID",
		})
	}

	var event models.Event
	if err := DB.Where("id = ? AND is_deleted = ?", eventId, false).First(&event).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "event not found",
		})
	}

	var updates models.Event
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Update fields
	if err := DB.Model(&event).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to update event",
		})
	}

	return c.JSON(event)
}

// Delete event (soft delete)
func (e *EventsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	eventId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid event ID",
		})
	}

	isDeleted := true
	if err := DB.Model(&models.Event{}).Where("id = ?", eventId).Update("is_deleted", isDeleted).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to delete event",
		})
	}

	return c.Status(204).Send(nil)
}

// ==================== EVENT REGISTRATION ENDPOINTS ====================

// Register school for event (POST /events/:id/register)
func (e *EventsController) registerForEvent(c *fiber.Ctx, eventId uint) error {
	userId, _ := c.Locals("user_id").(uint)

	var registration models.EventRegistration
	if err := c.BodyParser(&registration); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Set event_id and registered_by
	eventIdInt := int64(eventId)
	registeredBy := int64(userId)
	registration.EventId = &eventIdInt
	registration.RegisteredBy = &registeredBy

	// Set defaults
	now := time.Now().Format("2006-01-02 15:04:05")
	registration.RegistrationDate = &now

	isDeleted := false
	registration.IsDeleted = &isDeleted

	if registration.PaymentStatus == nil {
		status := "pending"
		registration.PaymentStatus = &status
	}

	// Check if school already registered
	var existing models.EventRegistration
	err := DB.Where("event_id = ? AND school_id = ? AND is_deleted = ?", eventId, registration.SchoolId, false).First(&existing).Error
	if err == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "school already registered for this event",
		})
	}

	// Check max attendees
	var event models.Event
	if err := DB.Where("id = ?", eventId).First(&event).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "event not found",
		})
	}

	if event.MaxAttendees != nil && *event.MaxAttendees > 0 {
		var count int64
		DB.Model(&models.EventRegistration{}).
			Where("event_id = ? AND is_deleted = ?", eventId, false).
			Count(&count)
		if int(count) >= *event.MaxAttendees {
			return c.Status(400).JSON(fiber.Map{
				"error": "event is full",
			})
		}
	}

	if err := DB.Create(&registration).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to register for event",
		})
	}

	return c.Status(201).JSON(registration)
}

// Get event registrations (GET /events/:id/registrations)
func (e *EventsController) getEventRegistrations(c *fiber.Ctx, eventId uint) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	query := DB.Model(&models.EventRegistration{}).
		Where("event_id = ? AND is_deleted = ?", eventId, false)

	// Get total count
	var total int64
	query.Count(&total)

	// Get registrations with school names
	var registrations []models.EventRegistration
	if err := query.Order("registration_date DESC").Offset(offset).Limit(limit).Find(&registrations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch registrations",
		})
	}

	// Populate school names
	for i := range registrations {
		var school models.School
		if err := DB.Where("id = ?", registrations[i].SchoolId).First(&school).Error; err == nil {
			registrations[i].SchoolName = &school.Name
		}
	}

	return c.JSON(fiber.Map{
		"data": registrations,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// Get my registrations (GET /event-registrations/my)
func (e *EventsController) getMyRegistrations(c *fiber.Ctx) error {
	userId, _ := c.Locals("user_id").(uint)

	// Get user's school_id
	var user models.User
	if err := DB.Where("id = ?", userId).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	query := DB.Model(&models.EventRegistration{}).
		Where("registered_by = ? AND is_deleted = ?", userId, false)

	// Get total count
	var total int64
	query.Count(&total)

	// Get registrations
	var registrations []models.EventRegistration
	if err := query.Order("registration_date DESC").Offset(offset).Limit(limit).Find(&registrations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch registrations",
		})
	}

	// Populate event titles and school names
	for i := range registrations {
		var event models.Event
		if err := DB.Where("id = ?", registrations[i].EventId).First(&event).Error; err == nil {
			registrations[i].EventTitle = &event.Title
		}

		var school models.School
		if err := DB.Where("id = ?", registrations[i].SchoolId).First(&school).Error; err == nil {
			registrations[i].SchoolName = &school.Name
		}
	}

	return c.JSON(fiber.Map{
		"data": registrations,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// Cancel registration (DELETE /event-registrations/:id)
func (e *EventsController) cancelRegistration(c *fiber.Ctx, registrationId uint) error {
	isDeleted := true
	if err := DB.Model(&models.EventRegistration{}).Where("id = ?", registrationId).Update("is_deleted", isDeleted).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to cancel registration",
		})
	}

	return c.Status(204).Send(nil)
}

// Update payment status (PUT /event-registrations/:id/payment)
func (e *EventsController) updatePaymentStatus(c *fiber.Ctx, registrationId uint) error {
	var body struct {
		PaymentStatus    string  `json:"payment_status"`
		PaymentReference *string `json:"payment_reference"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	updates := map[string]interface{}{
		"payment_status": body.PaymentStatus,
	}

	if body.PaymentReference != nil {
		updates["payment_reference"] = *body.PaymentReference
	}

	if err := DB.Model(&models.EventRegistration{}).Where("id = ?", registrationId).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to update payment status",
		})
	}

	var registration models.EventRegistration
	DB.Where("id = ?", registrationId).First(&registration)

	return c.JSON(registration)
}

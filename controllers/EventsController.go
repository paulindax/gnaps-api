package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type EventsController struct {
	eventService  *services.EventService
	schoolService *services.SchoolService
}

// NewEventsController creates a new EventsController with dependencies
func NewEventsController(eventService *services.EventService, schoolService *services.SchoolService) *EventsController {
	return &EventsController{
		eventService:  eventService,
		schoolService: schoolService,
	}
}

func (e *EventsController) Handle(action string, c *fiber.Ctx) error {
	controller := c.Params("controller")

	// Handle public events routes (no auth required)
	if controller == "public-events" {
		return e.handlePublicEvents(action, c)
	}

	// Route to appropriate handler based on controller name
	if controller == "event-registrations" {
		return e.handleEventRegistrations(action, c)
	}

	// Check if this is a nested resource route
	id := c.Params("id")
	if id != "" {
		if eventId, err := strconv.ParseUint(action, 10, 64); err == nil {
			switch id {
			case "registrations":
				return e.getEventRegistrations(c, uint(eventId))
			case "register":
				return e.registerForEvent(c, uint(eventId))
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
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

func (e *EventsController) handlePublicEvents(action string, c *fiber.Ctx) error {
	switch action {
	case "view":
		return e.getEventByCode(c)
	case "register":
		return e.publicRegisterForEvent(c)
	case "schools":
		return e.searchSchools(c)
	default:
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

func (e *EventsController) handleEventRegistrations(action string, c *fiber.Ctx) error {
	switch action {
	case "my":
		return e.getMyRegistrations(c)
	default:
		id := c.Params("id")
		if id == "" {
			return utils.ValidationErrorResponse(c, "registration ID is required")
		}

		registrationId, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return utils.ValidationErrorResponse(c, "invalid registration ID")
		}

		switch action {
		case "payment":
			return e.updatePaymentStatus(c, uint(registrationId))
		default:
			if c.Method() == "DELETE" {
				return e.cancelRegistration(c, uint(registrationId))
			}
			return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
		}
	}
}

// ==================== EVENT ENDPOINTS ====================

func (e *EventsController) list(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if isPaid := c.Query("is_paid"); isPaid != "" {
		filters["is_paid"] = isPaid == "true"
	}
	if fromDate := c.Query("from_date"); fromDate != "" {
		filters["from_date"] = fromDate
	}
	if toDate := c.Query("to_date"); toDate != "" {
		filters["to_date"] = toDate
	}

	events, total, err := e.eventService.ListEvents(filters, page, limit)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to fetch events")
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

func (e *EventsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	eventId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid event ID")
	}

	event, err := e.eventService.GetEventByID(uint(eventId))
	if err != nil {
		return utils.NotFoundResponse(c, "Event not found")
	}

	return c.JSON(event)
}

func (e *EventsController) create(c *fiber.Ctx) error {
	userId, _ := c.Locals("user_id").(uint)

	var event models.Event
	if err := c.BodyParser(&event); err != nil {
		log.Println("Error parsing request body:", err)
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := e.eventService.CreateEvent(&event, userId); err != nil {
		return utils.ServerErrorResponse(c, "Failed to create event")
	}

	return utils.SuccessResponseWithStatus(c, 201, event, "Event created successfully")
}

func (e *EventsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	eventId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid event ID")
	}

	var updates models.Event
	if err := c.BodyParser(&updates); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Convert to map for partial updates
	updateMap := make(map[string]interface{})
	if updates.Title != "" {
		updateMap["title"] = updates.Title
	}
	if updates.Description != nil {
		updateMap["description"] = updates.Description
	}
	if updates.StartDate != "" {
		updateMap["start_date"] = updates.StartDate
	}
	if updates.EndDate != nil {
		updateMap["end_date"] = updates.EndDate
	}
	if updates.Location != nil {
		updateMap["location"] = updates.Location
	}
	if updates.Venue != nil {
		updateMap["venue"] = updates.Venue
	}
	if updates.IsPaid != nil {
		updateMap["is_paid"] = updates.IsPaid
	}
	if updates.Price != nil {
		updateMap["price"] = updates.Price
	}
	if updates.MaxAttendees != nil {
		updateMap["max_attendees"] = updates.MaxAttendees
	}
	if updates.RegistrationDeadline != nil {
		updateMap["registration_deadline"] = updates.RegistrationDeadline
	}
	if updates.Status != nil {
		updateMap["status"] = updates.Status
	}
	if updates.ImageUrl != nil {
		updateMap["image_url"] = updates.ImageUrl
	}

	if err := e.eventService.UpdateEvent(uint(eventId), updateMap); err != nil {
		return utils.ServerErrorResponse(c, err.Error())
	}

	// Get updated event
	event, _ := e.eventService.GetEventByID(uint(eventId))
	return utils.SuccessResponse(c, event, "Event updated successfully")
}

func (e *EventsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	eventId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid event ID")
	}

	if err := e.eventService.DeleteEvent(uint(eventId)); err != nil {
		return utils.ServerErrorResponse(c, "Failed to delete event")
	}

	return c.Status(204).Send(nil)
}

// ==================== REGISTRATION ENDPOINTS ====================

func (e *EventsController) registerForEvent(c *fiber.Ctx, eventId uint) error {
	userId, _ := c.Locals("user_id").(uint)

	var registration models.EventRegistration
	if err := c.BodyParser(&registration); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Get event first to get registration code
	event, err := e.eventService.GetEventByID(eventId)
	if err != nil {
		return utils.NotFoundResponse(c, "Event not found")
	}

	if err := e.eventService.RegisterForEvent(*event.RegistrationCode, &registration, &userId); err != nil {
		return utils.ValidationErrorResponse(c, err.Error())
	}

	return utils.SuccessResponseWithStatus(c, 201, registration, "Registration successful")
}

func (e *EventsController) getEventRegistrations(c *fiber.Ctx, eventId uint) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	registrations, total, err := e.eventService.GetEventRegistrations(eventId, page, limit)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to fetch registrations")
	}

	// Populate school names
	for i := range registrations {
		if registrations[i].SchoolId != nil {
			school, err := e.schoolService.GetSchoolByID(uint(*registrations[i].SchoolId))
			if err == nil {
				registrations[i].SchoolName = &school.Name
			}
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

func (e *EventsController) getMyRegistrations(c *fiber.Ctx) error {
	userId, _ := c.Locals("user_id").(uint)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	registrations, total, err := e.eventService.GetMyRegistrations(userId, page, limit)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to fetch registrations")
	}

	// Populate event titles and school names
	for i := range registrations {
		if registrations[i].EventId != nil {
			event, err := e.eventService.GetEventByID(uint(*registrations[i].EventId))
			if err == nil {
				registrations[i].EventTitle = &event.Title
			}
		}
		if registrations[i].SchoolId != nil {
			school, err := e.schoolService.GetSchoolByID(uint(*registrations[i].SchoolId))
			if err == nil {
				registrations[i].SchoolName = &school.Name
			}
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

func (e *EventsController) cancelRegistration(c *fiber.Ctx, registrationId uint) error {
	if err := e.eventService.CancelRegistration(registrationId); err != nil {
		return utils.ServerErrorResponse(c, "Failed to cancel registration")
	}

	return c.Status(204).Send(nil)
}

func (e *EventsController) updatePaymentStatus(c *fiber.Ctx, registrationId uint) error {
	var body struct {
		PaymentStatus    string  `json:"payment_status"`
		PaymentReference *string `json:"payment_reference"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	reference := ""
	if body.PaymentReference != nil {
		reference = *body.PaymentReference
	}

	if err := e.eventService.UpdatePaymentStatus(registrationId, body.PaymentStatus, reference); err != nil {
		return utils.ServerErrorResponse(c, "Failed to update payment status")
	}

	return utils.SuccessResponse(c, fiber.Map{"payment_status": body.PaymentStatus}, "Payment status updated successfully")
}

// ==================== PUBLIC ENDPOINTS (NO AUTH) ====================

func (e *EventsController) getEventByCode(c *fiber.Ctx) error {
	code := c.Params("id")

	event, err := e.eventService.GetEventByCode(code)
	if err != nil {
		return utils.NotFoundResponse(c, "Event not found")
	}

	return c.JSON(event)
}

func (e *EventsController) publicRegisterForEvent(c *fiber.Ctx) error {
	code := c.Params("id")

	var registration models.EventRegistration
	if err := c.BodyParser(&registration); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := e.eventService.RegisterForEvent(code, &registration, nil); err != nil {
		return utils.ValidationErrorResponse(c, err.Error())
	}

	return utils.SuccessResponseWithStatus(c, 201, registration, "Registration successful")
}

func (e *EventsController) searchSchools(c *fiber.Ctx) error {
	keyword := c.Query("search", "")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	schools, err := e.schoolService.Search(keyword, limit)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to search schools")
	}

	return c.JSON(schools)
}

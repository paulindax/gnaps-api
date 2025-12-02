package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"strings"
	"time"
)

type EventService struct {
	eventRepo *repositories.EventRepository
	regRepo   *repositories.RegistrationRepository
}

func NewEventService(eventRepo *repositories.EventRepository, regRepo *repositories.RegistrationRepository) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		regRepo:   regRepo,
	}
}

// CreateEvent creates a new event with auto-generated registration code
func (s *EventService) CreateEvent(event *models.Event, userId uint) error {
	// Set created_by
	event.CreatedBy = int64(userId)

	// Set defaults
	isDeleted := false
	event.IsDeleted = &isDeleted

	if event.Status == nil {
		status := "draft"
		event.Status = &status
	}

	// Generate unique registration code
	code := s.generateUniqueRegistrationCode()
	event.RegistrationCode = &code

	return s.eventRepo.Create(event)
}

// GetEventByID retrieves an event with registration count
func (s *EventService) GetEventByID(id uint) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Add registered count
	event.RegisteredCount = int(s.eventRepo.GetRegisteredCount(id))
	return event, nil
}

// GetEventByCode retrieves an event by registration code
func (s *EventService) GetEventByCode(code string) (*models.Event, error) {
	event, err := s.eventRepo.FindByCode(code)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Add registered count
	event.RegisteredCount = int(s.eventRepo.GetRegisteredCount(event.ID))
	return event, nil
}

// ListEvents retrieves paginated events with filters
func (s *EventService) ListEvents(filters map[string]interface{}, page, limit int) ([]models.Event, int64, error) {
	events, total, err := s.eventRepo.List(filters, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Add registered counts
	for i := range events {
		events[i].RegisteredCount = int(s.eventRepo.GetRegisteredCount(events[i].ID))
	}

	return events, total, nil
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(id uint, updates map[string]interface{}) error {
	// Verify event exists
	_, err := s.eventRepo.FindByID(id)
	if err != nil {
		return errors.New("event not found")
	}

	return s.eventRepo.Update(id, updates)
}

// DeleteEvent soft deletes an event
func (s *EventService) DeleteEvent(id uint) error {
	return s.eventRepo.Delete(id)
}

// RegisterForEvent handles event registration with validation
func (s *EventService) RegisterForEvent(eventCode string, registration *models.EventRegistration, userId *uint) error {
	// Get event
	event, err := s.eventRepo.FindByCode(eventCode)
	if err != nil {
		return errors.New("event not found")
	}

	// Validate registration
	if err := s.validateRegistration(event, registration); err != nil {
		return err
	}

	// Set event_id
	registration.EventId = int64(event.ID)

	// Set registered_by if provided (for authenticated requests)
	if userId != nil {
		userIdInt64 := int64(*userId)
		registration.RegisteredBy = &userIdInt64
	}

	// Set defaults
	registration.RegistrationDate = time.Now()

	isDeleted := false
	registration.IsDeleted = &isDeleted

	// Set payment status based on whether event is paid
	if event.IsPaid != nil && *event.IsPaid {
		status := "pending"
		registration.PaymentStatus = &status
	} else {
		status := "confirmed"
		registration.PaymentStatus = &status
	}

	return s.regRepo.Create(registration)
}

// validateRegistration validates event registration
func (s *EventService) validateRegistration(event *models.Event, registration *models.EventRegistration) error {
	// Check if school already registered
	if registration.SchoolId != 0 {
		exists, _ := s.regRepo.SchoolAlreadyRegistered(event.ID, registration.SchoolId)
		if exists {
			return errors.New("school already registered for this event")
		}
	}

	// Check max attendees
	if event.MaxAttendees != nil && *event.MaxAttendees > 0 {
		count, _ := s.regRepo.CountByEvent(event.ID)
		if int(count) >= *event.MaxAttendees {
			return errors.New("event is full")
		}
	}

	return nil
}

// GetEventRegistrations retrieves registrations for an event
func (s *EventService) GetEventRegistrations(eventID uint, page, limit int) ([]models.EventRegistration, int64, error) {
	return s.regRepo.ListByEvent(eventID, page, limit)
}

// GetMyRegistrations retrieves registrations for a user
func (s *EventService) GetMyRegistrations(userID uint, page, limit int) ([]models.EventRegistration, int64, error) {
	return s.regRepo.ListByUser(userID, page, limit)
}

// UpdatePaymentStatus updates the payment status of a registration
func (s *EventService) UpdatePaymentStatus(registrationID uint, status, reference string) error {
	return s.regRepo.UpdatePaymentStatus(registrationID, status, reference)
}

// CancelRegistration soft deletes a registration
func (s *EventService) CancelRegistration(registrationID uint) error {
	return s.regRepo.Delete(registrationID)
}

// generateRegistrationCode generates a random 8-character hex code
func (s *EventService) generateRegistrationCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	code := hex.EncodeToString(bytes)
	return strings.ToUpper(code)
}

// generateUniqueRegistrationCode ensures uniqueness
func (s *EventService) generateUniqueRegistrationCode() string {
	for {
		code := s.generateRegistrationCode()
		if !s.eventRepo.CodeExists(code) {
			return code
		}
	}
}

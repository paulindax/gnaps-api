package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
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

// GetEventByID retrieves an event with registration count and bill name
func (s *EventService) GetEventByID(id uint) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Add registered count
	event.RegisteredCount = int(s.eventRepo.GetRegisteredCount(id))

	// Populate bill name if bill_id exists
	if event.BillId != nil && *event.BillId > 0 {
		event.BillName = s.eventRepo.GetBillName(*event.BillId)
	}

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

	// Populate bill name if bill_id exists
	if event.BillId != nil && *event.BillId > 0 {
		event.BillName = s.eventRepo.GetBillName(*event.BillId)
	}

	return event, nil
}

// ListEvents retrieves paginated events with filters
func (s *EventService) ListEvents(filters map[string]interface{}, page, limit int) ([]models.Event, int64, error) {
	events, total, err := s.eventRepo.List(filters, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Add registered counts and bill names
	for i := range events {
		events[i].RegisteredCount = int(s.eventRepo.GetRegisteredCount(events[i].ID))
		// Populate bill name if bill_id exists
		if events[i].BillId != nil && *events[i].BillId > 0 {
			events[i].BillName = s.eventRepo.GetBillName(*events[i].BillId)
		}
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

// ============================================
// Owner-based methods for data filtering
// ============================================

// CreateEventWithOwner creates a new event with owner context
func (s *EventService) CreateEventWithOwner(event *models.Event, userId uint, ownerCtx *utils.OwnerContext) error {
	event.CreatedBy = int64(userId)

	isDeleted := false
	event.IsDeleted = &isDeleted

	if event.Status == nil {
		status := "draft"
		event.Status = &status
	}

	code := s.generateUniqueRegistrationCode()
	event.RegistrationCode = &code

	return s.eventRepo.CreateWithOwner(event, ownerCtx)
}

// GetEventByIDWithOwner retrieves an event with owner filtering
func (s *EventService) GetEventByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.Event, error) {
	event, err := s.eventRepo.FindByIDWithOwner(id, ownerCtx)
	if err != nil {
		return nil, err
	}

	event.RegisteredCount = int(s.eventRepo.GetRegisteredCount(id))

	if event.BillId != nil && *event.BillId > 0 {
		event.BillName = s.eventRepo.GetBillName(*event.BillId)
	}

	return event, nil
}

// ListEventsWithOwner retrieves paginated events with owner filtering
func (s *EventService) ListEventsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Event, int64, error) {
	events, total, err := s.eventRepo.ListWithOwner(filters, page, limit, ownerCtx)
	if err != nil {
		return nil, 0, err
	}

	for i := range events {
		events[i].RegisteredCount = int(s.eventRepo.GetRegisteredCount(events[i].ID))
		if events[i].BillId != nil && *events[i].BillId > 0 {
			events[i].BillName = s.eventRepo.GetBillName(*events[i].BillId)
		}
	}

	return events, total, nil
}

// UpdateEventWithOwner updates an event with owner verification
func (s *EventService) UpdateEventWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	return s.eventRepo.UpdateWithOwner(id, updates, ownerCtx)
}

// DeleteEventWithOwner soft deletes an event with owner verification
func (s *EventService) DeleteEventWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	return s.eventRepo.DeleteWithOwner(id, ownerCtx)
}

// RegisterForEvent handles event registration with validation
// If a school is already registered, it updates the existing registration
func (s *EventService) RegisterForEvent(eventCode string, registration *models.EventRegistration, userId *uint) error {
	// Get event
	event, err := s.eventRepo.FindByCode(eventCode)
	if err != nil {
		return errors.New("event not found")
	}

	// Debug logging
	println("DEBUG: RegisterForEvent - eventCode:", eventCode, "event.ID:", event.ID, "registration.SchoolId:", registration.SchoolId)

	// Check if school already registered
	if registration.SchoolId != 0 {
		println("DEBUG: Checking for existing registration - event.ID:", event.ID, "schoolId:", registration.SchoolId)
		existingReg, err := s.regRepo.FindByEventAndSchool(event.ID, registration.SchoolId)
		println("DEBUG: FindByEventAndSchool result - err:", err, "existingReg != nil:", existingReg != nil)
		if err == nil && existingReg != nil {
			// School already registered - update the existing registration
			println("DEBUG: Found existing registration, updating...")
			return s.updateExistingRegistration(existingReg, registration, userId)
		}
		println("DEBUG: No existing registration found, creating new...")
	}

	// Validate new registration (check max attendees)
	if err := s.validateNewRegistration(event); err != nil {
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

// updateExistingRegistration updates an existing school registration with new data
// This also reactivates previously cancelled registrations
func (s *EventService) updateExistingRegistration(existing *models.EventRegistration, newData *models.EventRegistration, userId *uint) error {
	updates := make(map[string]interface{})

	// Reset soft delete flags to reactivate cancelled registrations
	updates["is_deleted"] = false
	updates["deleted_at"] = nil

	// Update number of attendees if provided
	if newData.NumberOfAttendees != nil {
		updates["number_of_attendees"] = *newData.NumberOfAttendees
	}

	// Update payment details if provided
	if newData.PaymentMethod != nil {
		updates["payment_method"] = *newData.PaymentMethod
	}
	if newData.PaymentPhone != nil {
		updates["payment_phone"] = *newData.PaymentPhone
	}

	// Update registered_by if provided
	if userId != nil {
		updates["registered_by"] = int64(*userId)
	}

	// Update registration date to now
	updates["registration_date"] = time.Now()

	return s.regRepo.Update(existing.ID, updates)
}

// validateNewRegistration validates a new event registration (checks max attendees only)
func (s *EventService) validateNewRegistration(event *models.Event) error {
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

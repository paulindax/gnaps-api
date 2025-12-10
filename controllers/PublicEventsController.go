package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/services"
	"gnaps-api/utils"
	"gnaps-api/workers"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PublicEventsController struct {
	eventRepo        *repositories.EventRepository
	registrationRepo *repositories.RegistrationRepository
	schoolRepo       *repositories.SchoolRepository
	db               *gorm.DB
	paymentService   *services.MomoPaymentService
	paymentWorker    *workers.PaymentWorker
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

// SetPaymentDependencies sets the payment service and worker (called after initialization)
func (p *PublicEventsController) SetPaymentDependencies(paymentService *services.MomoPaymentService, paymentWorker *workers.PaymentWorker) {
	p.paymentService = paymentService
	p.paymentWorker = paymentWorker
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
	case "initiate-payment":
		return p.initiatePayment(c)
	case "payment-status":
		return p.checkPaymentStatus(c)
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
// If a school is already registered, it updates the existing registration
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

	// Check if school is already registered for this event
	existingReg, err := p.registrationRepo.FindByEventAndSchool(event.ID, registration.SchoolId)
	if err == nil && existingReg != nil {
		// School already registered - update the existing registration
		updates := map[string]interface{}{
			"is_deleted":        false,
			"deleted_at":        nil,
			"registration_date": time.Now(),
		}

		// Update number of attendees if provided
		if registration.NumberOfAttendees != nil {
			updates["number_of_attendees"] = *registration.NumberOfAttendees
		}

		// Update payment details if provided
		if registration.PaymentMethod != nil {
			updates["payment_method"] = *registration.PaymentMethod
		}
		if registration.PaymentPhone != nil {
			updates["payment_phone"] = *registration.PaymentPhone
		}

		if err := p.registrationRepo.Update(existingReg.ID, updates); err != nil {
			return utils.ServerErrorResponse(c, "Failed to update registration. Please try again.")
		}

		// Return the updated registration
		existingReg.RegistrationDate = time.Now()
		if registration.NumberOfAttendees != nil {
			existingReg.NumberOfAttendees = registration.NumberOfAttendees
		}
		return utils.SuccessResponse(c, existingReg, "Registration updated successfully!")
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
// If bill_id is provided, only returns schools that have a school_bill for that bill
func (p *PublicEventsController) searchSchools(c *fiber.Ctx) error {
	keyword := c.Query("search")
	if keyword == "" {
		return c.JSON([]models.School{})
	}

	// Check if bill_id filter is provided
	billIDStr := c.Query("bill_id")
	if billIDStr != "" {
		billID, err := strconv.ParseInt(billIDStr, 10, 64)
		if err == nil && billID > 0 {
			// Search with bill filter
			schools, err := p.schoolRepo.SearchWithBill(keyword, billID, 20)
			if err != nil {
				return utils.ServerErrorResponse(c, "Failed to search schools. Please try again.")
			}
			return c.JSON(schools)
		}
	}

	// Search without bill filter
	schools, err := p.schoolRepo.Search(keyword, 20)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to search schools. Please try again.")
	}

	return c.JSON(schools)
}

// getSchoolBalance returns the school's bill balance (no auth required)
// If bill_id query param is provided, checks balance for that specific bill
// Otherwise, checks balance for the latest bill
func (p *PublicEventsController) getSchoolBalance(c *fiber.Ctx) error {
	schoolID := c.Params("id")
	if schoolID == "" {
		return utils.ValidationErrorResponse(c, "School ID is required")
	}

	var targetBill models.Bill

	// Check if a specific bill_id is provided (from event)
	billIDStr := c.Query("bill_id")
	if billIDStr != "" {
		billID, err := strconv.ParseInt(billIDStr, 10, 64)
		if err == nil && billID > 0 {
			// Get the specific bill
			err = p.db.Where("id = ? AND is_deleted = ?", billID, false).First(&targetBill).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return c.JSON(fiber.Map{
						"balance":     0,
						"has_balance": false,
						"blocked":     false,
					})
				}
				return utils.ServerErrorResponse(c, "Failed to retrieve bill information")
			}
		}
	}

	// If no specific bill_id provided, get the latest bill
	if targetBill.ID == 0 {
		err := p.db.Where("is_deleted = ? AND (is_approved IS NULL OR is_approved = ?)", false, true).
			Order("created_at DESC").
			First(&targetBill).Error

		if err != nil {
			// If no bills exist at all, allow registration without balance
			if err == gorm.ErrRecordNotFound {
				return c.JSON(fiber.Map{
					"balance":     0,
					"has_balance": false,
					"blocked":     false,
				})
			}
			return utils.ServerErrorResponse(c, "Failed to retrieve bill information")
		}
	}

	// Check if school has a school_bill for the target bill
	var schoolBill models.SchoolBill
	err := p.db.Where("school_id = ? AND bill_id = ?", schoolID, targetBill.ID).
		First(&schoolBill).Error

	if err != nil {
		// If no school bill found, block registration
		if err == gorm.ErrRecordNotFound {
			billName := ""
			if targetBill.Name != nil {
				billName = *targetBill.Name
			}
			return c.JSON(fiber.Map{
				"balance":     0,
				"has_balance": false,
				"blocked":     true,
				"message":     fmt.Sprintf("Your school does not have a bill for '%s'. Please contact your Zone Admin to have your school added to the billing system.", billName),
			})
		}
		return utils.ServerErrorResponse(c, "Failed to retrieve school balance")
	}

	// Return the balance for the school's bill
	balance := 0.0
	if schoolBill.Balance != nil {
		balance = *schoolBill.Balance
	}

	return c.JSON(fiber.Map{
		"balance":     balance,
		"has_balance": balance > 0,
		"blocked":     false,
		"bill_id":     schoolBill.ID,
		"bill_name":   targetBill.Name,
	})
}

// initiatePayment initiates a MoMo payment for event registration (no auth required)
// Payment must succeed BEFORE registration is created
func (p *PublicEventsController) initiatePayment(c *fiber.Ctx) error {
	if p.paymentService == nil {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "Payment service not available",
		})
	}

	code := c.Params("id")
	if code == "" {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "Registration code is required",
		})
	}

	// Get event by code
	event, err := p.eventRepo.FindByCode(code)
	if err != nil {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "Event not found or registration link is invalid",
		})
	}

	var req struct {
		Amount            float64 `json:"amount"`
		PhoneNumber       string  `json:"phone_number"`
		Network           string  `json:"network"`
		SchoolID          int64   `json:"school_id"`
		SchoolName        string  `json:"school_name"`
		NumberOfAttendees int     `json:"number_of_attendees"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Amount <= 0 {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "Amount must be greater than 0",
		})
	}
	if req.PhoneNumber == "" {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "Phone number is required",
		})
	}
	if req.Network == "" {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "Network is required",
		})
	}
	if req.SchoolID == 0 {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": "School ID is required",
		})
	}

	// Default number of attendees to 1
	if req.NumberOfAttendees <= 0 {
		req.NumberOfAttendees = 1
	}

	// Create payment request - PayeeID will be set to event ID for now
	// Registration will be created on successful payment
	paymentReq := services.MomoPaymentRequest{
		Amount:            req.Amount,
		PhoneNumber:       req.PhoneNumber,
		Network:           req.Network,
		FeeName:           "Event Registration Fee",
		PayeeID:           int64(event.ID), // Use event ID, not registration ID
		PayeeType:         "EventPayment",  // Different type to indicate payment-first flow
		SchoolID:          req.SchoolID,
		SchoolName:        req.SchoolName,
		NumberOfAttendees: req.NumberOfAttendees,
		EventCode:         code,
	}

	// Initiate payment
	resp, err := p.paymentService.InitiatePayment(paymentReq)
	if err != nil {
		return c.JSON(fiber.Map{
			"error":   true,
			"message": resp.Message,
		})
	}

	// Enqueue background job to process payment with Hubtel
	if p.paymentWorker != nil {
		if err := p.paymentWorker.EnqueuePaymentProcess(resp.PaymentTransactionID); err != nil {
			// Log error but don't fail the request - process synchronously as fallback
			fmt.Printf("Failed to enqueue payment processing: %v\n", err)
			go p.paymentService.ProcessPaymentWithHubtel(resp.PaymentTransactionID)
		}
	} else {
		// Process synchronously if worker not available
		go p.paymentService.ProcessPaymentWithHubtel(resp.PaymentTransactionID)
	}

	return c.JSON(fiber.Map{
		"error":                  resp.Error,
		"message":                resp.Message,
		"payment_transaction_id": resp.PaymentTransactionID,
	})
}

// checkPaymentStatus checks the status of a MoMo payment (no auth required)
func (p *PublicEventsController) checkPaymentStatus(c *fiber.Ctx) error {
	if p.paymentService == nil {
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Payment service not available",
		})
	}

	id := c.Params("id")
	paymentID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid payment ID",
		})
	}

	status, err := p.paymentService.CheckPaymentStatus(uint(paymentID))
	if err != nil {
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":         status.Status,
		"bank_status":    status.BankStatus,
		"trans_status":   status.TransStatus,
		"message":        status.Message,
		"transaction_id": status.TransactionID,
	})
}

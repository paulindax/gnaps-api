package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
	"gnaps-api/workers"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SmsController struct {
	smsService *services.SmsService
	db         *gorm.DB
}

func NewSmsController(smsService *services.SmsService, db *gorm.DB) *SmsController {
	return &SmsController{
		smsService: smsService,
		db:         db,
	}
}

// getOwnerInfoFromContext determines owner_type and owner_id based on authenticated user
func (s *SmsController) getOwnerInfoFromContext(c *fiber.Ctx) (string, int64, error) {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	switch userRole {
	case "system_admin", "national_admin":
		return "National", 1, nil

	case "region_admin":
		var executive models.Executive
		if err := s.db.Where("user_id = ? AND is_deleted = ?", userId, false).First(&executive).Error; err != nil {
			return "", 0, fmt.Errorf("executive not found for user")
		}
		if executive.RegionId == nil || *executive.RegionId == 0 {
			return "", 0, fmt.Errorf("no region assigned to executive")
		}
		return "Region", *executive.RegionId, nil

	case "zone_admin":
		var executive models.Executive
		if err := s.db.Where("user_id = ? AND is_deleted = ?", userId, false).First(&executive).Error; err != nil {
			return "", 0, fmt.Errorf("executive not found for user")
		}
		if executive.ZoneId == nil || *executive.ZoneId == 0 {
			return "", 0, fmt.Errorf("no zone assigned to executive")
		}
		return "Zone", *executive.ZoneId, nil

	default:
		return "", 0, fmt.Errorf("unauthorized role for messaging: %s", userRole)
	}
}

func (s *SmsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "send":
		return s.send(c)
	case "send_bulk":
		return s.sendBulk(c)
	case "available_units":
		return s.availableUnits(c)
	default:
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

// SendSmsRequest represents the request body for sending SMS
type SendSmsRequest struct {
	Message   string `json:"message"`
	Recipient string `json:"recipient"`
	Free      bool   `json:"free"`
}

// SendBulkSmsRequest represents the request body for sending bulk SMS
type SendBulkSmsRequest struct {
	Message    string   `json:"message"`
	Recipients []string `json:"recipients"`
	Free       bool     `json:"free"`
}

// send sends an SMS to a single recipient
func (s *SmsController) send(c *fiber.Ctx) error {
	var req SendSmsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Validate required fields
	if req.Message == "" {
		return utils.ValidationErrorResponse(c, "Message is required")
	}
	if req.Recipient == "" {
		return utils.ValidationErrorResponse(c, "Recipient is required")
	}

	// Get owner info from JWT context
	ownerType, ownerID, err := s.getOwnerInfoFromContext(c)
	if err != nil {
		return utils.ForbiddenResponse(c, err.Error())
	}

	// Send SMS
	payload := workers.SmsSendPayload{
		Message:   req.Message,
		Recipient: req.Recipient,
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Free:      req.Free,
	}

	if err := s.smsService.SendSMS(payload); err != nil {
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, fiber.Map{
		"recipient": req.Recipient,
		"message":   req.Message,
	}, "SMS sent successfully")
}

// sendBulk sends SMS to multiple recipients
func (s *SmsController) sendBulk(c *fiber.Ctx) error {
	var req SendBulkSmsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Validate required fields
	if req.Message == "" {
		return utils.ValidationErrorResponse(c, "Message is required")
	}
	if len(req.Recipients) == 0 {
		return utils.ValidationErrorResponse(c, "At least one recipient is required")
	}

	// Get owner info from JWT context
	ownerType, ownerID, err := s.getOwnerInfoFromContext(c)
	if err != nil {
		return utils.ForbiddenResponse(c, err.Error())
	}

	// Send bulk SMS
	payload := workers.SmsBulkSendPayload{
		Message:    req.Message,
		Recipients: req.Recipients,
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		Free:       req.Free,
	}

	if err := s.smsService.SendBulkSMS(payload); err != nil {
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, fiber.Map{
		"recipients_count": len(req.Recipients),
		"message":          req.Message,
	}, "Bulk SMS sent successfully")
}

// availableUnits returns the available SMS units for the current user's organization
func (s *SmsController) availableUnits(c *fiber.Ctx) error {
	// Get owner info from JWT context
	ownerType, ownerID, err := s.getOwnerInfoFromContext(c)
	if err != nil {
		return utils.ForbiddenResponse(c, err.Error())
	}

	units, err := s.smsService.GetAvailableUnits(ownerType, ownerID)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to get available units: "+err.Error())
	}

	return utils.SuccessResponse(c, fiber.Map{
		"owner_type":      ownerType,
		"owner_id":        ownerID,
		"available_units": units,
	}, "Available units retrieved successfully")
}

package controllers

import (
	"fmt"
	"gnaps-api/services"
	"gnaps-api/utils"
	"gnaps-api/workers"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PaymentsController struct {
	paymentService *services.MomoPaymentService
	paymentWorker  *workers.PaymentWorker
}

// NewPaymentsController creates a new PaymentsController with dependencies
func NewPaymentsController(paymentService *services.MomoPaymentService, paymentWorker *workers.PaymentWorker) *PaymentsController {
	return &PaymentsController{
		paymentService: paymentService,
		paymentWorker:  paymentWorker,
	}
}

func (p *PaymentsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "initiate":
		return p.initiatePayment(c)
	case "status":
		return p.checkStatus(c)
	case "callback":
		return p.handleCallback(c)
	default:
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

// initiatePayment handles payment initiation
func (p *PaymentsController) initiatePayment(c *fiber.Ctx) error {
	var req services.MomoPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{
			"error":         true,
			"message":       "Invalid request body",
			"flash_message": fiber.Map{"msg": "Invalid request body", "type": "error"},
		})
	}

	// Validate required fields
	if req.Amount <= 0 {
		return c.JSON(fiber.Map{
			"error":         true,
			"message":       "Amount must be greater than 0",
			"flash_message": fiber.Map{"msg": "Amount must be greater than 0", "type": "error"},
		})
	}
	if req.PhoneNumber == "" {
		return c.JSON(fiber.Map{
			"error":         true,
			"message":       "Phone number is required",
			"flash_message": fiber.Map{"msg": "Phone number is required", "type": "error"},
		})
	}
	if req.Network == "" {
		return c.JSON(fiber.Map{
			"error":         true,
			"message":       "Network is required",
			"flash_message": fiber.Map{"msg": "Network is required", "type": "error"},
		})
	}

	// Set default fee name if not provided
	if req.FeeName == "" {
		req.FeeName = "Event Registration Fee"
	}

	// Initiate payment
	resp, err := p.paymentService.InitiatePayment(req)
	if err != nil {
		return c.JSON(fiber.Map{
			"error":                  true,
			"message":                resp.Message,
			"flash_message":          fiber.Map{"msg": resp.Message, "type": "error"},
			"payment_transaction_id": nil,
		})
	}

	// Enqueue background job to process payment with Hubtel
	if p.paymentWorker != nil {
		if err := p.paymentWorker.EnqueuePaymentProcess(resp.PaymentTransactionID); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to enqueue payment processing: %v\n", err)
			// Process synchronously as fallback
			go p.paymentService.ProcessPaymentWithHubtel(resp.PaymentTransactionID)
		}
	} else {
		// Process synchronously if worker not available
		go p.paymentService.ProcessPaymentWithHubtel(resp.PaymentTransactionID)
	}

	return c.JSON(fiber.Map{
		"error":                  resp.Error,
		"message":                resp.Message,
		"flash_message":          fiber.Map{"msg": resp.Message, "type": "success"},
		"payment_transaction_id": resp.PaymentTransactionID,
	})
}

// checkStatus checks payment status
func (p *PaymentsController) checkStatus(c *fiber.Ctx) error {
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

// handleCallback handles Hubtel callback
func (p *PaymentsController) handleCallback(c *fiber.Ctx) error {
	// Get callback parameters
	actionType := c.Query("action_type")
	reference := c.Query("ref")

	if actionType != "verify_hubtel_pay" && actionType != "verify_fidelity_pay" {
		return c.Status(400).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid action type",
		})
	}

	// Parse callback body
	var callbackData struct {
		Status        string `json:"Status"`
		TransactionId string `json:"TransactionId"`
		ResponseCode  string `json:"ResponseCode"`
		Data          struct {
			TransactionId   string `json:"TransactionId"`
			ClientReference string `json:"ClientReference"`
			Status          string `json:"Status"`
		} `json:"Data"`
	}

	if err := c.BodyParser(&callbackData); err != nil {
		// Try to get status from query params
		status := c.Query("status", "")
		transactionID := c.Query("transaction_id", "")
		if status == "" {
			status = callbackData.Status
		}
		if transactionID == "" {
			transactionID = callbackData.TransactionId
		}
		if err := p.paymentService.HandleHubtelCallback(reference, status, transactionID); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
	} else {
		// Use parsed body
		status := callbackData.Status
		if status == "" {
			status = callbackData.Data.Status
		}
		transactionID := callbackData.TransactionId
		if transactionID == "" {
			transactionID = callbackData.Data.TransactionId
		}
		if err := p.paymentService.HandleHubtelCallback(reference, status, transactionID); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Callback processed successfully",
	})
}

package controllers

import (
	"fmt"
	"gnaps-api/services"
	"gnaps-api/workers"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SchoolPaymentsController struct {
	schoolBillService  *services.SchoolBillService
	momoPaymentService *services.MomoPaymentService
	paymentWorker      *workers.PaymentWorker
}

func NewSchoolPaymentsController(
	schoolBillService *services.SchoolBillService,
	momoPaymentService *services.MomoPaymentService,
	paymentWorker *workers.PaymentWorker,
) *SchoolPaymentsController {
	return &SchoolPaymentsController{
		schoolBillService:  schoolBillService,
		momoPaymentService: momoPaymentService,
		paymentWorker:      paymentWorker,
	}
}

func (s *SchoolPaymentsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "record":
		return s.record(c)
	case "status":
		return s.status(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

// record records a payment for a school bill
func (s *SchoolPaymentsController) record(c *fiber.Ctx) error {
	var req services.PaymentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if req.SchoolID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "school_id is required",
		})
	}
	if req.SchoolBillID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "school_bill_id is required",
		})
	}
	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "amount must be greater than 0",
		})
	}
	if req.PaymentMode == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "payment_mode is required",
		})
	}
	if req.PaymentDate == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "payment_date is required",
		})
	}

	// Get user ID from context (set by auth middleware)
	if userId := c.Locals("user_id"); userId != nil {
		if id, ok := userId.(int64); ok {
			req.UserID = id
		}
	}

	// Handle MoMo payments
	if req.PaymentMode == "MoMo" {
		// Validate MoMo-specific fields
		if req.MomoNumber == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "momo_number is required for MoMo payments",
			})
		}
		if req.MomoNetwork == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "momo_network is required for MoMo payments",
			})
		}

		// Get school name and bill name for payment description
		_, billName, err := s.schoolBillService.GetBalance(req.SchoolBillID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "School bill not found",
			})
		}

		// Check balance before proceeding
		balance, _, _ := s.schoolBillService.GetBalance(req.SchoolBillID)
		if req.Amount > balance {
			return c.Status(400).JSON(fiber.Map{
				"error": "Payment amount exceeds outstanding balance",
			})
		}

		// Initiate MoMo payment
		momoReq := services.MomoPaymentRequest{
			Amount:      req.Amount,
			PhoneNumber: req.MomoNumber,
			Network:     req.MomoNetwork,
			FeeName:     fmt.Sprintf("School Bill: %s", billName),
			PayeeID:     int64(req.SchoolBillID),
			PayeeType:   "SchoolBillPayment",
			UserID:      &req.UserID,
			SchoolID:    req.SchoolID,
			SchoolName:  req.SchoolName,
		}

		resp, err := s.momoPaymentService.InitiatePayment(momoReq)
		if err != nil || resp.Error {
			errorMsg := "Failed to initiate payment"
			if resp != nil {
				errorMsg = resp.Message
			}
			return c.Status(400).JSON(fiber.Map{
				"error": errorMsg,
				"flash_message": fiber.Map{
					"msg":  errorMsg,
					"type": "error",
				},
			})
		}

		return c.Status(201).JSON(fiber.Map{
			"message":        "MoMo payment initiated. Please approve on your phone.",
			"payment_method": "MoMo",
			"flash_message": fiber.Map{
				"msg":  "MoMo payment initiated. Please approve on your phone.",
				"type": "info",
			},
			"data": fiber.Map{
				"payment_transaction_id": resp.PaymentTransactionID,
				"status":                 "pending",
			},
		})
	}

	// Handle Cash payments
	transaction, err := s.schoolBillService.RecordPayment(req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
			"flash_message": fiber.Map{
				"msg":  err.Error(),
				"type": "error",
			},
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":        "Payment recorded successfully",
		"payment_method": "Cash",
		"flash_message": fiber.Map{
			"msg":  "Payment recorded successfully",
			"type": "success",
		},
		"data": fiber.Map{
			"transaction_id": transaction.ID,
			"receipt_no":     transaction.ReceiptNo,
			"amount":         transaction.Amount,
			"status":         "successful",
		},
	})
}

// status checks the status of a MoMo payment
func (s *SchoolPaymentsController) status(c *fiber.Ctx) error {
	// Get payment_id from params or query
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}
	if id == "" {
		id = c.Query("payment_transaction_id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "payment_transaction_id is required",
		})
	}

	paymentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid payment_transaction_id",
		})
	}

	status, err := s.momoPaymentService.CheckPaymentStatus(uint(paymentId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": err.Error(),
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

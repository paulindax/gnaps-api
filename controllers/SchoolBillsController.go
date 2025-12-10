package controllers

import (
	"fmt"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SchoolBillsController struct {
	schoolBillService *services.SchoolBillService
}

func NewSchoolBillsController(schoolBillService *services.SchoolBillService) *SchoolBillsController {
	return &SchoolBillsController{
		schoolBillService: schoolBillService,
	}
}

func (s *SchoolBillsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return s.list(c)
	case "balance":
		return s.balance(c)
	case "particulars":
		return s.particulars(c)
	case "payment-history":
		return s.paymentHistory(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

// list returns all bills for a school
func (s *SchoolBillsController) list(c *fiber.Ctx) error {
	// Get school_id from query params
	schoolIdStr := c.Query("school_id")
	if schoolIdStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "school_id is required",
		})
	}

	schoolId, err := strconv.ParseInt(schoolIdStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid school_id",
		})
	}

	schoolBills, err := s.schoolBillService.GetSchoolBills(schoolId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve school bills",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": schoolBills,
	})
}

// balance returns the balance for a specific school bill
func (s *SchoolBillsController) balance(c *fiber.Ctx) error {
	// Get school_bill_id from params or query
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "school bill ID is required",
		})
	}

	schoolBillId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid school bill ID",
		})
	}

	balance, billName, err := s.schoolBillService.GetBalance(uint(schoolBillId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"balance":   balance,
		"bill_name": billName,
	})
}

// particulars returns the billing particulars for a school bill
func (s *SchoolBillsController) particulars(c *fiber.Ctx) error {
	// Get school_bill_id from params or query
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "school bill ID is required",
		})
	}

	schoolBillId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid school bill ID",
		})
	}

	particulars, err := s.schoolBillService.GetSchoolBillingParticulars(uint(schoolBillId))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve particulars",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": particulars,
	})
}

// paymentHistory returns the payment history for a school
func (s *SchoolBillsController) paymentHistory(c *fiber.Ctx) error {
	// Get school_id from query params
	schoolIdStr := c.Query("school_id")
	if schoolIdStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "school_id is required",
		})
	}

	schoolId, err := strconv.ParseInt(schoolIdStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid school_id",
		})
	}

	// Get pagination params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	payments, total, err := s.schoolBillService.GetPaymentHistory(schoolId, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve payment history",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  payments,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

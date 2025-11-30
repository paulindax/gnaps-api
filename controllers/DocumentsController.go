package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type DocumentsController struct {
	documentService *services.DocumentService
}

func NewDocumentsController(documentService *services.DocumentService) *DocumentsController {
	return &DocumentsController{documentService: documentService}
}

func (d *DocumentsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return d.list(c)
	case "show":
		return d.show(c)
	case "create":
		return d.create(c)
	case "update":
		return d.update(c)
	case "delete":
		return d.delete(c)
	case "submit":
		return d.submitDocument(c)
	case "submissions":
		return d.getSubmissions(c)
	case "submission":
		return d.getSubmission(c)
	case "updateSubmission":
		return d.updateSubmission(c)
	case "deleteSubmission":
		return d.deleteSubmission(c)
	case "reviewSubmission":
		return d.reviewSubmission(c)
	default:
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

// List all documents
func (d *DocumentsController) list(c *fiber.Ctx) error {
	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}

	documents, total, err := d.documentService.ListDocuments(filters, page, limit)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to fetch documents")
	}

	return c.JSON(fiber.Map{
		"data":  documents,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Show a single document
func (d *DocumentsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	documentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	document, err := d.documentService.GetDocumentByID(uint(documentId))
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return c.JSON(document)
}

// Create a new document
func (d *DocumentsController) create(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(uint)
	if !ok {
		return utils.UnauthorizedResponse(c, "Unauthorized")
	}

	var document models.Document
	if err := c.BodyParser(&document); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := d.documentService.CreateDocument(&document, userId); err != nil {
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponseWithStatus(c, 201, document, "Document created successfully")
}

// Update a document
func (d *DocumentsController) update(c *fiber.Ctx) error {
	id := c.Params("id")
	documentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	var updates models.Document
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
	if updates.Category != nil {
		updateMap["category"] = updates.Category
	}
	if updates.Status != nil {
		updateMap["status"] = updates.Status
	}
	if updates.TemplateData != "" {
		updateMap["template_data"] = updates.TemplateData
	}

	if err := d.documentService.UpdateDocument(uint(documentId), updateMap); err != nil {
		if err.Error() == "document not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	// Fetch updated document
	document, _ := d.documentService.GetDocumentByID(uint(documentId))

	return utils.SuccessResponse(c, document, "Document updated successfully")
}

// Delete a document (soft delete)
func (d *DocumentsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	documentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	if err := d.documentService.DeleteDocument(uint(documentId)); err != nil {
		if err.Error() == "document not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, nil, "Document deleted successfully")
}

// Submit a document (create submission)
func (d *DocumentsController) submitDocument(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(uint)
	if !ok {
		return utils.UnauthorizedResponse(c, "Unauthorized")
	}

	var submission models.DocumentSubmission
	if err := c.BodyParser(&submission); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := d.documentService.CreateSubmission(&submission, userId); err != nil {
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponseWithStatus(c, 201, submission, "Document submitted successfully")
}

// Get all submissions for a document or school
func (d *DocumentsController) getSubmissions(c *fiber.Ctx) error {
	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Build filters
	filters := make(map[string]interface{})
	if documentId := c.Query("document_id"); documentId != "" {
		filters["document_id"] = documentId
	}
	if schoolId := c.Query("school_id"); schoolId != "" {
		filters["school_id"] = schoolId
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	submissions, total, err := d.documentService.ListSubmissions(filters, page, limit)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to fetch submissions")
	}

	return c.JSON(fiber.Map{
		"data":  submissions,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Get a single submission
func (d *DocumentsController) getSubmission(c *fiber.Ctx) error {
	id := c.Params("id")
	submissionId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	submission, err := d.documentService.GetSubmissionByID(uint(submissionId))
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return c.JSON(submission)
}

// Update a submission
func (d *DocumentsController) updateSubmission(c *fiber.Ctx) error {
	id := c.Params("id")
	submissionId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	var updates models.DocumentSubmission
	if err := c.BodyParser(&updates); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Convert to map for partial updates
	updateMap := make(map[string]interface{})
	if updates.DocumentId != nil {
		updateMap["document_id"] = updates.DocumentId
	}
	if updates.SchoolId != nil {
		updateMap["school_id"] = updates.SchoolId
	}
	if updates.FormData != "" {
		updateMap["form_data"] = updates.FormData
	}
	if updates.Status != nil {
		updateMap["status"] = updates.Status
	}

	if err := d.documentService.UpdateSubmission(uint(submissionId), updateMap); err != nil {
		if err.Error() == "submission not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	// Fetch updated submission
	submission, _ := d.documentService.GetSubmissionByID(uint(submissionId))

	return utils.SuccessResponse(c, submission, "Submission updated successfully")
}

// Delete a submission (soft delete)
func (d *DocumentsController) deleteSubmission(c *fiber.Ctx) error {
	id := c.Params("id")
	submissionId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	if err := d.documentService.DeleteSubmission(uint(submissionId)); err != nil {
		if err.Error() == "submission not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, nil, "Submission deleted successfully")
}

// Review a submission
func (d *DocumentsController) reviewSubmission(c *fiber.Ctx) error {
	id := c.Params("id")
	submissionId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	userId, ok := c.Locals("user_id").(uint)
	if !ok {
		return utils.UnauthorizedResponse(c, "Unauthorized")
	}

	var review struct {
		Status      string  `json:"status"`
		ReviewNotes *string `json:"review_notes"`
	}

	if err := c.BodyParser(&review); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := d.documentService.ReviewSubmission(uint(submissionId), review.Status, review.ReviewNotes, userId); err != nil {
		if err.Error() == "submission not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	// Fetch updated submission
	submission, _ := d.documentService.GetSubmissionByID(uint(submissionId))

	return utils.SuccessResponse(c, submission, "Submission reviewed successfully")
}

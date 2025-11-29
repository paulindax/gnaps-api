package controllers

import (
	"fmt"
	"gnaps-api/models"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type DocumentsController struct {
}

func init() {
	RegisterController("documents", &DocumentsController{})
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
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// List all documents
func (d *DocumentsController) list(c *fiber.Ctx) error {
	var documents []models.Document

	query := DB.Where("is_deleted = ? OR is_deleted IS NULL", false)

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by category
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Document{}).Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&documents).Error; err != nil {
		log.Println("Error fetching documents:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch documents"})
	}

	// Get submission counts for each document
	for i := range documents {
		var count int64
		DB.Model(&models.DocumentSubmission{}).
			Where("document_id = ? AND (is_deleted = ? OR is_deleted IS NULL)", documents[i].ID, false).
			Count(&count)
		documents[i].SubmissionCount = int(count)
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

	var document models.Document
	if err := DB.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false).First(&document).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "document not found"})
	}

	// Get submission count
	var count int64
	DB.Model(&models.DocumentSubmission{}).
		Where("document_id = ? AND (is_deleted = ? OR is_deleted IS NULL)", document.ID, false).
		Count(&count)
	document.SubmissionCount = int(count)

	return c.JSON(document)
}

// Create a new document
func (d *DocumentsController) create(c *fiber.Ctx) error {
	userId, _ := c.Locals("user_id").(uint)

	var document models.Document
	if err := c.BodyParser(&document); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	createdBy := int64(userId)
	document.CreatedBy = &createdBy

	if err := DB.Create(&document).Error; err != nil {
		log.Println("Error creating document:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to create document"})
	}

	return c.Status(201).JSON(document)
}

// Update a document
func (d *DocumentsController) update(c *fiber.Ctx) error {
	id := c.Params("id")

	var document models.Document
	if err := DB.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false).First(&document).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "document not found"})
	}

	var updates models.Document
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Update fields
	if err := DB.Model(&document).Updates(updates).Error; err != nil {
		log.Println("Error updating document:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to update document"})
	}

	return c.JSON(document)
}

// Delete a document (soft delete)
func (d *DocumentsController) delete(c *fiber.Ctx) error {
	id := c.Params("id")

	var document models.Document
	if err := DB.Where("id = ?", id).First(&document).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "document not found"})
	}

	isDeleted := true
	if err := DB.Model(&document).Update("is_deleted", isDeleted).Error; err != nil {
		log.Println("Error deleting document:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete document"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "document deleted successfully"})
}

// Submit a document (create submission)
func (d *DocumentsController) submitDocument(c *fiber.Ctx) error {
	userId, _ := c.Locals("user_id").(uint)

	var submission models.DocumentSubmission
	if err := c.BodyParser(&submission); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	submittedBy := int64(userId)
	submission.SubmittedBy = &submittedBy

	if err := DB.Create(&submission).Error; err != nil {
		log.Println("Error creating submission:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to create submission"})
	}

	return c.Status(201).JSON(submission)
}

// Get all submissions for a document or school
func (d *DocumentsController) getSubmissions(c *fiber.Ctx) error {
	var submissions []models.DocumentSubmission

	query := DB.Where("is_deleted = ? OR is_deleted IS NULL", false)

	// Filter by document ID
	if documentId := c.Query("document_id"); documentId != "" {
		query = query.Where("document_id = ?", documentId)
	}

	// Filter by school ID
	if schoolId := c.Query("school_id"); schoolId != "" {
		query = query.Where("school_id = ?", schoolId)
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.DocumentSubmission{}).Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&submissions).Error; err != nil {
		log.Println("Error fetching submissions:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch submissions"})
	}

	// Enrich with related data
	for i := range submissions {
		// Get document title
		var document models.Document
		if err := DB.Select("title").Where("id = ?", submissions[i].DocumentId).First(&document).Error; err == nil {
			submissions[i].DocumentTitle = &document.Title
		}

		// Get school name
		var school models.School
		if err := DB.Select("name").Where("id = ?", submissions[i].SchoolId).First(&school).Error; err == nil {
			submissions[i].SchoolName = &school.Name
		}

		// Get submitter name
		var user models.User
		if err := DB.Select("username").Where("id = ?", submissions[i].SubmittedBy).First(&user).Error; err == nil {
			submissions[i].SubmitterName = user.Username
		}
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

	var submission models.DocumentSubmission
	if err := DB.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false).First(&submission).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "submission not found"})
	}

	// Get document title
	var document models.Document
	if err := DB.Select("title").Where("id = ?", submission.DocumentId).First(&document).Error; err == nil {
		submission.DocumentTitle = &document.Title
	}

	// Get school name
	var school models.School
	if err := DB.Select("name").Where("id = ?", submission.SchoolId).First(&school).Error; err == nil {
		submission.SchoolName = &school.Name
	}

	return c.JSON(submission)
}

// Update a submission
func (d *DocumentsController) updateSubmission(c *fiber.Ctx) error {
	id := c.Params("id")

	var submission models.DocumentSubmission
	if err := DB.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false).First(&submission).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "submission not found"})
	}

	var updates models.DocumentSubmission
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := DB.Model(&submission).Updates(updates).Error; err != nil {
		log.Println("Error updating submission:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to update submission"})
	}

	return c.JSON(submission)
}

// Delete a submission (soft delete)
func (d *DocumentsController) deleteSubmission(c *fiber.Ctx) error {
	id := c.Params("id")

	var submission models.DocumentSubmission
	if err := DB.Where("id = ?", id).First(&submission).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "submission not found"})
	}

	isDeleted := true
	if err := DB.Model(&submission).Update("is_deleted", isDeleted).Error; err != nil {
		log.Println("Error deleting submission:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete submission"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "submission deleted successfully"})
}

// Review a submission
func (d *DocumentsController) reviewSubmission(c *fiber.Ctx) error {
	id := c.Params("id")
	userId, _ := c.Locals("user_id").(uint)

	var submission models.DocumentSubmission
	if err := DB.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false).First(&submission).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "submission not found"})
	}

	var review struct {
		Status      string  `json:"status"`
		ReviewNotes *string `json:"review_notes"`
	}

	if err := c.BodyParser(&review); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	reviewedBy := int64(userId)
	submission.Status = &review.Status
	submission.ReviewNotes = review.ReviewNotes
	submission.ReviewedBy = &reviewedBy

	if err := DB.Model(&submission).Updates(submission).Error; err != nil {
		log.Println("Error reviewing submission:", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to review submission"})
	}

	return c.JSON(submission)
}

package controllers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MediaController struct {
}

func init() {
	RegisterController("media", &MediaController{})
}

func (m *MediaController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "upload":
		return m.upload(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// Upload handles file uploads
func (m *MediaController) upload(c *fiber.Ctx) error {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "no file uploaded",
		})
	}

	// Validate file type (images, videos, documents, PDFs)
	contentType := file.Header.Get("Content-Type")
	allowedTypes := []string{
		// Images
		"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/svg+xml",
		// Videos
		"video/mp4", "video/webm", "video/ogg", "video/quicktime",
		// Documents
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv",
	}
	if !contains(allowedTypes, contentType) {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid file type. Allowed types: images, videos, PDFs, and documents",
		})
	}

	// Validate file size based on content type
	var maxSize int64
	if contentType == "video/mp4" || contentType == "video/webm" || contentType == "video/ogg" || contentType == "video/quicktime" {
		maxSize = int64(50 * 1024 * 1024) // 50MB for videos
	} else if contentType == "application/pdf" || strings.HasPrefix(contentType, "application/vnd.") || strings.HasPrefix(contentType, "application/msword") {
		maxSize = int64(10 * 1024 * 1024) // 10MB for documents/PDFs
	} else {
		maxSize = int64(5 * 1024 * 1024) // 5MB for images and others
	}

	if file.Size > maxSize {
		sizeMB := maxSize / (1024 * 1024)
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("file size exceeds %dMB limit", sizeMB),
		})
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to create uploads directory",
		})
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), generateRandomString(10), ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to open uploaded file",
		})
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to save file",
		})
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to save file",
		})
	}

	// Return file URL
	fileURL := fmt.Sprintf("/uploads/%s", filename)

	return c.JSON(fiber.Map{
		"success": true,
		"url":     fileURL,
		"filename": filename,
	})
}

// Helper function to check if a string exists in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to generate random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return strings.ToLower(string(b))
}

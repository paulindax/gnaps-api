package controllers

import (
	"fmt"
	"gnaps-api/services"

	"github.com/gofiber/fiber/v2"
)

type MediaController struct {
	mediaService *services.MediaService
}

// NewMediaController creates a new instance of MediaController
func NewMediaController(mediaService *services.MediaService) *MediaController {
	return &MediaController{
		mediaService: mediaService,
	}
}

// Handle routes the action to the appropriate handler method
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

// upload handles file uploads
func (m *MediaController) upload(c *fiber.Ctx) error {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "no file uploaded",
		})
	}

	// Call the service to upload the file
	uploadResponse, err := m.mediaService.UploadFile(file)
	if err != nil {
		// Return appropriate error response
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
			"flash_message": fiber.Map{
				"msg":  err.Error(),
				"type": "error",
			},
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"success":  true,
		"url":      uploadResponse.URL,
		"filename": uploadResponse.Filename,
		"flash_message": fiber.Map{
			"msg":  "File uploaded successfully",
			"type": "success",
		},
	})
}

package controllers

import (
	"fmt"
	"gnaps-api/services"
	"gnaps-api/utils"

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
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

// upload handles file uploads
func (m *MediaController) upload(c *fiber.Ctx) error {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ValidationErrorResponse(c, "No file uploaded")
	}

	// Call the service to upload the file
	uploadResponse, err := m.mediaService.UploadFile(file)
	if err != nil {
		return utils.ValidationErrorResponse(c, err.Error())
	}

	// Return success response
	return utils.SuccessResponse(c, fiber.Map{
		"url":      uploadResponse.URL,
		"filename": uploadResponse.Filename,
	}, "File uploaded successfully")
}

package utils

import "github.com/gofiber/fiber/v2"

// FlashMessage represents a flash message to be displayed to the user
type FlashMessage struct {
	Msg  string `json:"msg"`
	Type string `json:"type"` // success, error, warning, info
}

// SuccessResponse returns a standardized success response with optional flash message
func SuccessResponse(c *fiber.Ctx, data interface{}, message string) error {
	response := fiber.Map{
		"success": true,
		"data":    data,
	}

	if message != "" {
		response["flash_message"] = FlashMessage{
			Msg:  message,
			Type: "success",
		}
	}

	return c.JSON(response)
}

// SuccessResponseWithStatus returns a success response with custom status code
func SuccessResponseWithStatus(c *fiber.Ctx, status int, data interface{}, message string) error {
	response := fiber.Map{
		"success": true,
		"data":    data,
	}

	if message != "" {
		response["flash_message"] = FlashMessage{
			Msg:  message,
			Type: "success",
		}
	}

	return c.Status(status).JSON(response)
}

// ErrorResponse returns a standardized error response with flash message
func ErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "error",
		},
	})
}

// WarningResponse returns a standardized warning response
func WarningResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"warning": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "warning",
		},
	})
}

// InfoResponse returns a standardized info response
func InfoResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"info": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "info",
		},
	})
}

// ValidationErrorResponse returns a validation error response with flash message
func ValidationErrorResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "error",
		},
	})
}

// NotFoundResponse returns a not found error response with flash message
func NotFoundResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "error",
		},
	})
}

// UnauthorizedResponse returns an unauthorized error response with flash message
func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "error",
		},
	})
}

// ForbiddenResponse returns a forbidden error response with flash message
func ForbiddenResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"error": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "error",
		},
	})
}

// ServerErrorResponse returns a server error response with flash message
func ServerErrorResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "error",
		},
	})
}

// ConflictResponse returns a conflict error response with flash message
func ConflictResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"error": message,
		"flash_message": FlashMessage{
			Msg:  message,
			Type: "error",
		},
	})
}

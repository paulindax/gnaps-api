package controllers

import (
	"fmt"
	"gnaps-api/services"

	"github.com/gofiber/fiber/v2"
)

type ChatController struct {
	chatService *services.ChatService
}

func NewChatController(chatService *services.ChatService) *ChatController {
	return &ChatController{chatService: chatService}
}

func (c *ChatController) Handle(action string, ctx *fiber.Ctx) error {
	switch action {
	case "message":
		return c.handleMessage(ctx)
	case "health":
		return c.healthCheck(ctx)
	default:
		return ctx.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

func (c *ChatController) handleMessage(ctx *fiber.Ctx) error {
	var request services.ChatRequest

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if request.Message == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Message is required",
		})
	}

	response, err := c.chatService.GetResponse(request)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to process message",
		})
	}

	return ctx.JSON(fiber.Map{
		"data": response,
	})
}

func (c *ChatController) healthCheck(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"status":  "ok",
		"service": "Adesua360 Chat Service",
	})
}

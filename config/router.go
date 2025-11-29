package config

import (
	"fmt"
	"gnaps-api/controllers"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func DynamicControllerDispatcher(c *fiber.Ctx) error {
	ctrlName := c.Params("controller")
	action := c.Params("action")
	authCtrl := strings.HasPrefix(c.Path(), "/api/auth/")

	if authCtrl {
		ctrlName = "auth"
	}

	ctrl, ok := controllers.GetController(ctrlName)
	if !ok {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("controller %s undefined", ctrlName),
		})
	}

	return ctrl.Handle(action, c)
}

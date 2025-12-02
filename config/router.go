package config

import (
	"fmt"
	"gnaps-api/controllers"
	"gnaps-api/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func DynamicControllerDispatcher(c *fiber.Ctx) error {
	ctrlName := c.Params("controller")
	action := c.Params("action")
	authCtrl := strings.HasPrefix(c.Path(), "/api/auth/")
	pubEvents := strings.HasPrefix(c.Path(), "/api/public-events/")

	if pubEvents {
		ctrlName = "public-events"
	}

	if authCtrl {
		ctrlName = "auth"
	}

	ctrl, ok := controllers.GetController(ctrlName)
	if !ok {
		return utils.NotFoundResponse(c, fmt.Sprintf("Controller '%s' not found", ctrlName))
	}

	return ctrl.Handle(action, c)
}

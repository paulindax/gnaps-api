package controllers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type UsersController struct {
}

func init() {
	RegisterController("users", &UsersController{})
}

func (u *UsersController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return u.list(c)
	case "create":
		return u.create(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

func (u *UsersController) list(c *fiber.Ctx) error {
	return c.JSON(jsonData)
}

func (u *UsersController) create(c *fiber.Ctx) error {

	return c.JSON(jsonData)
}

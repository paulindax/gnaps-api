package controllers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ApplicationController interface {
	Handle(action string, c *fiber.Ctx) error
}

var jsonData = fiber.Map{}

// DB is the global database connection for controllers
var DB *gorm.DB

// global registry
var registry = make(map[string]ApplicationController)

// RegisterController is called by each controller's init()
func RegisterController(name string, ctrl ApplicationController) {
	registry[name] = ctrl
}

// GetController fetches a registered controller by name
func GetController(name string) (ApplicationController, bool) {
	ctrl, ok := registry[name]
	return ctrl, ok
}

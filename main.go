package main

import (
	"log"
	"os"
	"path/filepath"

	"gnaps-api/config"
	"gnaps-api/controllers"
	"gnaps-api/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database connection
	config.ConnectDb()

	// Set the database connection for controllers
	controllers.DB = config.DBConn

	// Generate models from database tables (only if GENERATE_MODELS=true)
	if os.Getenv("GENERATE_MODELS") == "true" {
		modelsDir := filepath.Join(".", "models")
		if err := config.GenerateModelsMiddleware(config.DBConn, modelsDir); err != nil {
			log.Printf("Warning: Failed to generate models: %v", err)
		}
	}

	// Initialize a new Fiber app
	app := fiber.New()

	// Global middleware
	app.Use(logger.New()) // Request logging
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("FRONTEND_URL"),
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))

	// API routes
	api := app.Group("/api")

	// Public routes (no authentication required)
	public := api.Group("")
	public.All("/auth/:action", config.DynamicControllerDispatcher)

	// Protected routes (authentication required)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth) // Apply JWT middleware
	protected.All("/:controller/:action/:id", config.DynamicControllerDispatcher)
	protected.All("/:controller/:action", config.DynamicControllerDispatcher)

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "GNAPS API is running",
		})
	})

	// Serve static files (uploaded images)
	app.Static("/uploads", "./uploads")

	// Start the server on port 3010
	log.Fatal(app.Listen(":3010"))
}

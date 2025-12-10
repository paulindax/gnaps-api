package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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

	// Initialize dependency injection for refactored controllers
	config.InitializeControllers(config.DBConn)

	// Start the payment worker in a goroutine
	go func() {
		if err := config.PaymentWorker.StartWorker(config.MomoPaymentService.ProcessPaymentWithHubtel); err != nil {
			log.Printf("Payment worker error: %v", err)
		}
	}()

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
	public.All("/public-events/:action/:id", config.DynamicControllerDispatcher)
	public.All("/public-events/:action", config.DynamicControllerDispatcher)
	public.All("/public/:action", config.DynamicControllerDispatcher) // Public school registration

	// Protected routes (authentication required)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth)                           // Apply JWT middleware
	protected.Use(middleware.AttachOwnerContext(config.DBConn)) // Attach owner context for executive users
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

	// Start the server on configured port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3020"
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start server in goroutine
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	log.Printf("Server started on port %s. Press Ctrl+C to stop.", port)

	// Wait for interrupt signal
	<-quit
	log.Println("\nShutting down server...")

	// Shutdown Fiber server
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	// Close payment worker
	if config.PaymentWorker != nil {
		log.Println("Closing payment worker...")
		if err := config.PaymentWorker.Close(); err != nil {
			log.Printf("Error closing payment worker: %v", err)
		}
	}

	log.Println("Server stopped gracefully")
}

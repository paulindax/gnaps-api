package middleware

import (
	"gnaps-api/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTAuth middleware validates JWT tokens and attaches user info to context
func JWTAuth(c *fiber.Ctx) error {
	// Get Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing authorization header",
		})
	}

	// Extract token from "Bearer <token>" format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid authorization header format, expected: Bearer <token>",
		})
	}

	tokenString := parts[1]

	// Validate token
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "invalid or expired token",
			"details": err.Error(),
		})
	}

	// Attach user info to context for downstream handlers
	c.Locals("user_id", claims.UserID)
	c.Locals("email", claims.Email)
	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)

	// Continue to next handler
	return c.Next()
}

// OptionalJWTAuth middleware validates JWT if present, but doesn't require it
func OptionalJWTAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// No token provided, continue without user info
		return c.Next()
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		tokenString := parts[1]
		claims, err := utils.ValidateJWT(tokenString)
		if err == nil {
			// Valid token, attach user info
			c.Locals("user_id", claims.UserID)
			c.Locals("email", claims.Email)
			c.Locals("username", claims.Username)
			c.Locals("role", claims.Role)
		}
	}

	return c.Next()
}

// RequireRole middleware checks if user has specific role
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user role not found in context",
			})
		}

		// Check if user role is in allowed roles
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions",
		})
	}
}

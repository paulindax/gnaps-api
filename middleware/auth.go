package middleware

import (
	"gnaps-api/models"
	"gnaps-api/utils"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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

// AttachOwnerContext middleware fetches executive info and attaches owner context
// This should be used after JWTAuth middleware for routes that need owner-based filtering
// For system_admin users (may not be executive), creates a system admin context for view-only access
// For other admin roles without executive records, creates a fallback context based on JWT role
func AttachOwnerContext(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user_id from context (set by JWTAuth)
		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			// Log debug info for troubleshooting
			rawUserID := c.Locals("user_id")
			log.Printf("[AttachOwnerContext] user_id type assertion failed. Raw value: %v, Type: %T", rawUserID, rawUserID)
			// No user ID, continue without owner context
			return c.Next()
		}

		// Get role from JWT context
		userRole, _ := c.Locals("role").(string)
		log.Printf("[AttachOwnerContext] Processing user_id=%d, role=%s", userID, userRole)

		// Check if user is system_admin (may not be an executive)
		if userRole == utils.RoleSystemAdmin {
			// Create system admin owner context (view only, no create permission)
			ownerCtx := &utils.OwnerContext{
				Role:   utils.RoleSystemAdmin,
				UserID: userID,
				// No OwnerType/OwnerID - system admin sees all but cannot create
			}
			utils.SetOwnerContext(c, ownerCtx)
			log.Printf("[AttachOwnerContext] Created system_admin context for user_id=%d", userID)
			return c.Next()
		}

		// Find executive by user_id
		var executive models.Executive
		err := db.Where("user_id = ? AND (is_deleted = ? OR is_deleted IS NULL)", userID, false).First(&executive).Error
		if err != nil {
			log.Printf("[AttachOwnerContext] No executive record found for user_id=%d, creating fallback context", userID)
			// User is not an executive - create fallback owner context based on JWT role
			// This ensures admin users can still access data even without executive record
			switch userRole {
			case utils.RoleNationalAdmin:
				ownerCtx := &utils.OwnerContext{
					Role:      utils.RoleNationalAdmin,
					UserID:    userID,
					OwnerType: utils.OwnerTypeNational,
					OwnerID:   utils.DefaultNationalOwnerID,
				}
				utils.SetOwnerContext(c, ownerCtx)
				log.Printf("[AttachOwnerContext] Created national_admin context for user_id=%d", userID)
			case utils.RoleRegionAdmin, utils.RoleZoneAdmin:
				// For region/zone admin without executive record, create limited context
				// They will have view-only access similar to system admin
				ownerCtx := &utils.OwnerContext{
					Role:   userRole,
					UserID: userID,
					// No specific owner - will need executive record for proper filtering
				}
				utils.SetOwnerContext(c, ownerCtx)
				log.Printf("[AttachOwnerContext] Created %s context for user_id=%d", userRole, userID)
			case utils.RoleSchoolAdmin:
				// School admin without executive record - create school-level context
				ownerCtx := &utils.OwnerContext{
					Role:   utils.RoleSchoolAdmin,
					UserID: userID,
					// School admin access will be determined by school_id from user record
				}
				utils.SetOwnerContext(c, ownerCtx)
				log.Printf("[AttachOwnerContext] Created school_admin context for user_id=%d", userID)
			default:
				// For any other authenticated role, create basic context
				// This ensures authenticated users always have an owner context
				log.Printf("[AttachOwnerContext] Creating default context for unknown role=%s, user_id=%d", userRole, userID)
				ownerCtx := &utils.OwnerContext{
					Role:   userRole,
					UserID: userID,
				}
				utils.SetOwnerContext(c, ownerCtx)
			}
			// Continue to next handler
			return c.Next()
		}

		// Get role from executive
		role := ""
		if executive.Role != nil {
			role = *executive.Role
		}

		log.Printf("[AttachOwnerContext] Found executive record for user_id=%d, executive_role=%s", userID, role)

		// Create owner context based on executive role
		ownerCtx := utils.GetOwnerContextFromExecutive(
			role,
			executive.RegionId,
			executive.ZoneId,
			userID,
		)

		// Attach owner context to fiber context
		utils.SetOwnerContext(c, ownerCtx)

		// Also attach executive info for convenience
		c.Locals("executive_id", executive.ID)
		c.Locals("executive_role", role)

		return c.Next()
	}
}

// RequireOwnerContext middleware ensures owner context is present and valid
func RequireOwnerContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ownerCtx := utils.GetOwnerContext(c)
		if ownerCtx == nil || !ownerCtx.IsValid() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "owner context required but not found or invalid",
			})
		}
		return c.Next()
	}
}

// RequireExecutiveRole middleware checks if user is an executive with specific roles
func RequireExecutiveRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ownerCtx := utils.GetOwnerContext(c)
		if ownerCtx == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "executive context not found",
			})
		}

		// Check if executive role is in allowed roles
		for _, role := range allowedRoles {
			if ownerCtx.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient executive permissions",
		})
	}
}

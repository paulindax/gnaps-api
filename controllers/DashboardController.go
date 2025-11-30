package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"

	"github.com/gofiber/fiber/v2"
)

type DashboardController struct {
	dashboardService *services.DashboardService
}

// NewDashboardController creates a new instance of DashboardController
func NewDashboardController(dashboardService *services.DashboardService) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

// Handle routes requests to appropriate action handlers
func (d *DashboardController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "stats":
		return d.stats(c)
	case "overview":
		return d.overview(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// stats returns role-based dashboard statistics
func (d *DashboardController) stats(c *fiber.Ctx) error {
	// Get user info from JWT context (set by auth middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	role, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User role not found",
		})
	}

	// Fetch full user details from database
	var user models.User
	if err := DB.Where("id = ? AND is_deleted = ?", userID, false).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var stats fiber.Map
	var err error

	switch role {
	case "system_admin":
		stats = d.dashboardService.GetSystemAdminStats()
	case "national_admin":
		stats = d.dashboardService.GetNationalAdminStats()
	case "regional_admin":
		// Find executive record for this user to get assigned regions
		var executive models.Executive
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err != nil {
			c.Set("X-Flash-Error", "Executive profile not found for this user")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Executive profile not found for this user",
			})
		}
		// Get the first assigned region (or you could iterate through all)
		// For simplicity, we'll use the first region if multiple are assigned
		stats, err = d.dashboardService.GetRegionalAdminStatsFromExecutive(&executive)
		if err != nil {
			c.Set("X-Flash-Error", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	case "zone_admin":
		// Find executive record for this user to get assigned zones
		var executive models.Executive
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err != nil {
			c.Set("X-Flash-Error", "Executive profile not found for this user")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Executive profile not found for this user",
			})
		}
		stats, err = d.dashboardService.GetZoneAdminStatsFromExecutive(&executive)
		if err != nil {
			c.Set("X-Flash-Error", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	case "school_user":
		// Find school for this user
		var school models.School
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, false).First(&school).Error; err != nil {
			c.Set("X-Flash-Error", "School not found for this user")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "School not found for this user",
			})
		}
		stats = d.dashboardService.GetSchoolUserStats(fmt.Sprintf("%d", school.ID))
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid role: %s", role),
		})
	}

	// Add user info to response
	stats["user"] = fiber.Map{
		"id":         user.ID,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"role":       user.Role,
	}

	return c.JSON(fiber.Map{
		"role": role,
		"data": stats,
	})
}

// overview returns recent activities and news based on role
func (d *DashboardController) overview(c *fiber.Ctx) error {
	// Get user info from JWT context (set by auth middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	role, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User role not found",
		})
	}

	// Get overview data from service
	overview := d.dashboardService.GetOverview(userID, role)

	return c.JSON(overview)
}

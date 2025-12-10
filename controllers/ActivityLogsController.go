package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ActivityLogsController struct {
	activityLogService *services.ActivityLogService
}

func NewActivityLogsController(activityLogService *services.ActivityLogService) *ActivityLogsController {
	return &ActivityLogsController{
		activityLogService: activityLogService,
	}
}

func (a *ActivityLogsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "list":
		return a.list(c)
	case "show":
		return a.show(c)
	case "create":
		return a.create(c)
	case "batch":
		return a.batch(c)
	case "stats":
		return a.stats(c)
	case "user":
		return a.userActivities(c)
	case "recent":
		return a.recent(c)
	case "cleanup":
		return a.cleanup(c)
	default:
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

// list returns a paginated list of all activities (system admin only)
func (a *ActivityLogsController) list(c *fiber.Ctx) error {
	// Check if user is system admin
	role, ok := c.Locals("role").(string)
	if !ok || role != "system_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "insufficient permissions - system admin only"})
	}

	// Parse filters from query params
	filters := make(map[string]interface{})

	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = uint(id)
		}
	}

	if activityType := c.Query("type"); activityType != "" {
		filters["type"] = activityType
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		if id, err := strconv.ParseUint(resourceID, 10, 64); err == nil {
			filters["resource_id"] = uint(id)
		}
	}

	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			filters["from_date"] = t
		}
	}

	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse("2006-01-02", toDate); err == nil {
			// Add 1 day to include the entire end date
			filters["to_date"] = t.AddDate(0, 0, 1)
		}
	}

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	activities, total, err := a.activityLogService.ListActivities(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve activities",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": activities,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// show returns a single activity log by ID
func (a *ActivityLogsController) show(c *fiber.Ctx) error {
	// Check if user is system admin
	role, ok := c.Locals("role").(string)
	if !ok || role != "system_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "insufficient permissions - system admin only"})
	}

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	activityID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	activity, err := a.activityLogService.GetActivityByID(uint(activityID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": activity})
}

// create logs a new activity
func (a *ActivityLogsController) create(c *fiber.Ctx) error {
	// Get user info from context
	userID, _ := c.Locals("user_id").(uint)
	username, _ := c.Locals("username").(string)
	role, _ := c.Locals("role").(string)

	var activity models.ActivityLog
	if err := c.BodyParser(&activity); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Set user info from JWT context
	userIDInt64 := int64(userID)
	activity.UserId = &userIDInt64
	activity.Username = &username
	activity.Role = &role

	// Set IP and user agent
	ip := c.IP()
	userAgent := c.Get("User-Agent")
	activity.IpAddress = &ip
	activity.UserAgent = &userAgent

	if err := a.activityLogService.LogActivity(&activity); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Activity logged successfully",
		"data":    activity,
	})
}

// batch logs multiple activities at once
func (a *ActivityLogsController) batch(c *fiber.Ctx) error {
	// Get user info from context
	userID, _ := c.Locals("user_id").(uint)
	username, _ := c.Locals("username").(string)
	role, _ := c.Locals("role").(string)

	var activities []models.ActivityLog
	if err := c.BodyParser(&activities); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Set user info from JWT context for all activities
	ip := c.IP()
	userAgent := c.Get("User-Agent")
	userIDInt64 := int64(userID)

	for i := range activities {
		activities[i].UserId = &userIDInt64
		activities[i].Username = &username
		activities[i].Role = &role
		activities[i].IpAddress = &ip
		activities[i].UserAgent = &userAgent
	}

	if err := a.activityLogService.LogBatch(activities); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": fmt.Sprintf("Successfully logged %d activities", len(activities)),
	})
}

// stats returns activity statistics (system admin only)
func (a *ActivityLogsController) stats(c *fiber.Ctx) error {
	// Check if user is system admin
	role, ok := c.Locals("role").(string)
	if !ok || role != "system_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "insufficient permissions - system admin only"})
	}

	days, _ := strconv.Atoi(c.Query("days", "7"))

	stats, err := a.activityLogService.GetActivityStats(days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve statistics",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"data": stats})
}

// userActivities returns activities for a specific user or current user
func (a *ActivityLogsController) userActivities(c *fiber.Ctx) error {
	currentUserID, _ := c.Locals("user_id").(uint)
	currentRole, _ := c.Locals("role").(string)

	// Get requested user ID (defaults to current user)
	requestedUserID := currentUserID
	if idParam := c.Params("id"); idParam != "" {
		if id, err := strconv.ParseUint(idParam, 10, 64); err == nil {
			requestedUserID = uint(id)
		}
	} else if idQuery := c.Query("user_id"); idQuery != "" {
		if id, err := strconv.ParseUint(idQuery, 10, 64); err == nil {
			requestedUserID = uint(id)
		}
	}

	// Non-admin users can only view their own activities
	if currentRole != "system_admin" && requestedUserID != currentUserID {
		return c.Status(403).JSON(fiber.Map{"error": "insufficient permissions"})
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	activities, total, err := a.activityLogService.ListUserActivities(requestedUserID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve activities",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": activities,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// recent returns recent activities (last N hours)
func (a *ActivityLogsController) recent(c *fiber.Ctx) error {
	// Check if user is system admin
	role, ok := c.Locals("role").(string)
	if !ok || role != "system_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "insufficient permissions - system admin only"})
	}

	hours, _ := strconv.Atoi(c.Query("hours", "24"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))

	activities, total, err := a.activityLogService.ListRecentActivities(hours, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve activities",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": activities,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// cleanup removes old activity logs (system admin only)
func (a *ActivityLogsController) cleanup(c *fiber.Ctx) error {
	// Check if user is system admin
	role, ok := c.Locals("role").(string)
	if !ok || role != "system_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "insufficient permissions - system admin only"})
	}

	// Retention period in days (default 90 days)
	retentionDays, _ := strconv.Atoi(c.Query("retention_days", "90"))

	deleted, err := a.activityLogService.CleanupOldActivities(retentionDays)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Successfully deleted %d old activity logs", deleted),
		"deleted": deleted,
	})
}

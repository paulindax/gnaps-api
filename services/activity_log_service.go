package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"time"
)

// Activity type constants
const (
	ActivityTypeCreate  = "create"
	ActivityTypeUpdate  = "update"
	ActivityTypeDelete  = "delete"
	ActivityTypeView    = "view"
	ActivityTypeAPICall = "api_call"
	ActivityTypeLogin   = "login"
	ActivityTypeLogout  = "logout"
	ActivityTypeExport  = "export"
)

type ActivityLogService struct {
	activityLogRepo *repositories.ActivityLogRepository
}

func NewActivityLogService(activityLogRepo *repositories.ActivityLogRepository) *ActivityLogService {
	return &ActivityLogService{activityLogRepo: activityLogRepo}
}

// LogActivity creates a new activity log entry
func (s *ActivityLogService) LogActivity(activity *models.ActivityLog) error {
	if activity.Type == "" {
		return errors.New("activity type is required")
	}
	if activity.Title == "" {
		return errors.New("activity title is required")
	}
	if activity.UserId == nil || *activity.UserId == 0 {
		return errors.New("user_id is required")
	}

	return s.activityLogRepo.Create(activity)
}

// LogBatch creates multiple activity log entries at once
func (s *ActivityLogService) LogBatch(activities []models.ActivityLog) error {
	// Validate each activity
	for i, activity := range activities {
		if activity.Type == "" {
			return errors.New("activity type is required for all entries")
		}
		if activity.Title == "" {
			return errors.New("activity title is required for all entries")
		}
		if activity.UserId == nil || *activity.UserId == 0 {
			return errors.New("user_id is required for all entries")
		}
		activities[i] = activity
	}

	return s.activityLogRepo.CreateBatch(activities)
}

// GetActivityByID retrieves a single activity log by ID
func (s *ActivityLogService) GetActivityByID(id uint) (*models.ActivityLog, error) {
	activity, err := s.activityLogRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("activity log not found")
	}
	return activity, nil
}

// ListActivities returns a paginated list of activities with optional filters
func (s *ActivityLogService) ListActivities(filters map[string]interface{}, page, limit int) ([]models.ActivityLog, int64, error) {
	return s.activityLogRepo.List(filters, page, limit)
}

// ListUserActivities returns activities for a specific user
func (s *ActivityLogService) ListUserActivities(userID uint, page, limit int) ([]models.ActivityLog, int64, error) {
	return s.activityLogRepo.ListByUser(userID, page, limit)
}

// ListRecentActivities returns activities from the last N hours
func (s *ActivityLogService) ListRecentActivities(hours int, page, limit int) ([]models.ActivityLog, int64, error) {
	return s.activityLogRepo.ListRecent(hours, page, limit)
}

// GetActivityStats returns activity statistics for the dashboard
func (s *ActivityLogService) GetActivityStats(days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 7 // Default to last 7 days
	}
	return s.activityLogRepo.GetActivityStats(days)
}

// CleanupOldActivities removes activities older than the specified number of days
func (s *ActivityLogService) CleanupOldActivities(retentionDays int) (int64, error) {
	if retentionDays < 7 {
		return 0, errors.New("retention period must be at least 7 days")
	}
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return s.activityLogRepo.DeleteOlderThan(cutoffDate)
}

// Helper functions for creating specific activity types

// LogCreate logs a resource creation activity
func (s *ActivityLogService) LogCreate(userID uint, username, role, resourceType string, resourceID uint, title, description string) error {
	userIDInt64 := int64(userID)
	resourceIDInt64 := int64(resourceID)
	activity := &models.ActivityLog{
		UserId:       &userIDInt64,
		Username:     &username,
		Role:         &role,
		Type:         ActivityTypeCreate,
		Title:        title,
		Description:  &description,
		ResourceType: &resourceType,
		ResourceId:   &resourceIDInt64,
	}
	return s.activityLogRepo.Create(activity)
}

// LogUpdate logs a resource update activity
func (s *ActivityLogService) LogUpdate(userID uint, username, role, resourceType string, resourceID uint, title, description string) error {
	userIDInt64 := int64(userID)
	resourceIDInt64 := int64(resourceID)
	activity := &models.ActivityLog{
		UserId:       &userIDInt64,
		Username:     &username,
		Role:         &role,
		Type:         ActivityTypeUpdate,
		Title:        title,
		Description:  &description,
		ResourceType: &resourceType,
		ResourceId:   &resourceIDInt64,
	}
	return s.activityLogRepo.Create(activity)
}

// LogDelete logs a resource deletion activity
func (s *ActivityLogService) LogDelete(userID uint, username, role, resourceType string, resourceID uint, title, description string) error {
	userIDInt64 := int64(userID)
	resourceIDInt64 := int64(resourceID)
	activity := &models.ActivityLog{
		UserId:       &userIDInt64,
		Username:     &username,
		Role:         &role,
		Type:         ActivityTypeDelete,
		Title:        title,
		Description:  &description,
		ResourceType: &resourceType,
		ResourceId:   &resourceIDInt64,
	}
	return s.activityLogRepo.Create(activity)
}

// LogView logs a resource view activity
func (s *ActivityLogService) LogView(userID uint, username, role, resourceType string, resourceID uint, title string) error {
	userIDInt64 := int64(userID)
	resourceIDInt64 := int64(resourceID)
	activity := &models.ActivityLog{
		UserId:       &userIDInt64,
		Username:     &username,
		Role:         &role,
		Type:         ActivityTypeView,
		Title:        title,
		ResourceType: &resourceType,
		ResourceId:   &resourceIDInt64,
	}
	return s.activityLogRepo.Create(activity)
}

// LogAPICall logs an API call activity
func (s *ActivityLogService) LogAPICall(userID uint, username, role, method, endpoint string, statusCode int, ipAddress, userAgent string) error {
	userIDInt64 := int64(userID)
	title := method + " " + endpoint
	activity := &models.ActivityLog{
		UserId:     &userIDInt64,
		Username:   &username,
		Role:       &role,
		Type:       ActivityTypeAPICall,
		Title:      title,
		Method:     &method,
		Endpoint:   &endpoint,
		StatusCode: &statusCode,
		IpAddress:  &ipAddress,
		UserAgent:  &userAgent,
	}
	return s.activityLogRepo.Create(activity)
}

// LogLogin logs a user login activity
func (s *ActivityLogService) LogLogin(userID uint, username, role, ipAddress, userAgent string) error {
	userIDInt64 := int64(userID)
	activity := &models.ActivityLog{
		UserId:    &userIDInt64,
		Username:  &username,
		Role:      &role,
		Type:      ActivityTypeLogin,
		Title:     "User logged in",
		IpAddress: &ipAddress,
		UserAgent: &userAgent,
	}
	return s.activityLogRepo.Create(activity)
}

// LogLogout logs a user logout activity
func (s *ActivityLogService) LogLogout(userID uint, username, role string) error {
	userIDInt64 := int64(userID)
	activity := &models.ActivityLog{
		UserId:   &userIDInt64,
		Username: &username,
		Role:     &role,
		Type:     ActivityTypeLogout,
		Title:    "User logged out",
	}
	return s.activityLogRepo.Create(activity)
}

// LogExport logs an export activity
func (s *ActivityLogService) LogExport(userID uint, username, role, resourceType, format, title string) error {
	userIDInt64 := int64(userID)
	description := "Exported " + resourceType + " as " + format
	activity := &models.ActivityLog{
		UserId:       &userIDInt64,
		Username:     &username,
		Role:         &role,
		Type:         ActivityTypeExport,
		Title:        title,
		Description:  &description,
		ResourceType: &resourceType,
	}
	return s.activityLogRepo.Create(activity)
}

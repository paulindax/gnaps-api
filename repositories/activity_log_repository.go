package repositories

import (
	"gnaps-api/models"
	"time"

	"gorm.io/gorm"
)

type ActivityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

func (r *ActivityLogRepository) Create(activity *models.ActivityLog) error {
	return r.db.Create(activity).Error
}

func (r *ActivityLogRepository) CreateBatch(activities []models.ActivityLog) error {
	if len(activities) == 0 {
		return nil
	}
	return r.db.Create(&activities).Error
}

func (r *ActivityLogRepository) FindByID(id uint) (*models.ActivityLog, error) {
	var activity models.ActivityLog
	err := r.db.First(&activity, id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityLogRepository) List(filters map[string]interface{}, page, limit int) ([]models.ActivityLog, int64, error) {
	var activities []models.ActivityLog
	var total int64

	query := r.db.Model(&models.ActivityLog{})

	// Apply filters
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}

	if activityType, ok := filters["type"]; ok {
		query = query.Where("type = ?", activityType)
	}

	if resourceType, ok := filters["resource_type"]; ok {
		query = query.Where("resource_type = ?", resourceType)
	}

	if resourceID, ok := filters["resource_id"]; ok {
		query = query.Where("resource_id = ?", resourceID)
	}

	// Date range filters
	if fromDate, ok := filters["from_date"]; ok {
		query = query.Where("created_at >= ?", fromDate)
	}

	if toDate, ok := filters["to_date"]; ok {
		query = query.Where("created_at <= ?", toDate)
	}

	// Search in title and description
	if search, ok := filters["search"]; ok {
		searchTerm := "%" + search.(string) + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	// Search by username
	if username, ok := filters["username"]; ok {
		usernameTerm := "%" + username.(string) + "%"
		query = query.Where("username LIKE ?", usernameTerm)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities).Error

	return activities, total, err
}

// ListByUser returns all activities for a specific user
func (r *ActivityLogRepository) ListByUser(userID uint, page, limit int) ([]models.ActivityLog, int64, error) {
	filters := map[string]interface{}{"user_id": userID}
	return r.List(filters, page, limit)
}

// ListRecent returns activities from the last N hours
func (r *ActivityLogRepository) ListRecent(hours int, page, limit int) ([]models.ActivityLog, int64, error) {
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	filters := map[string]interface{}{"from_date": cutoff}
	return r.List(filters, page, limit)
}

// CountByType returns the count of activities by type for a given time period
func (r *ActivityLogRepository) CountByType(fromDate, toDate time.Time) (map[string]int64, error) {
	type Result struct {
		Type  string
		Count int64
	}

	var results []Result
	err := r.db.Model(&models.ActivityLog{}).
		Select("type, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", fromDate, toDate).
		Group("type").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Type] = r.Count
	}

	return counts, nil
}

// CountByUser returns the count of activities per user for a given time period
func (r *ActivityLogRepository) CountByUser(fromDate, toDate time.Time, limit int) ([]map[string]interface{}, error) {
	var results []struct {
		UserID   uint
		Username string
		Count    int64
	}

	err := r.db.Model(&models.ActivityLog{}).
		Select("user_id, username, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", fromDate, toDate).
		Group("user_id, username").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	var output []map[string]interface{}
	for _, r := range results {
		output = append(output, map[string]interface{}{
			"user_id":  r.UserID,
			"username": r.Username,
			"count":    r.Count,
		})
	}

	return output, nil
}

// DeleteOlderThan removes activities older than the specified date
func (r *ActivityLogRepository) DeleteOlderThan(date time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", date).Delete(&models.ActivityLog{})
	return result.RowsAffected, result.Error
}

// GetActivityStats returns activity statistics for dashboard
func (r *ActivityLogRepository) GetActivityStats(days int) (map[string]interface{}, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	// Get total count
	var totalCount int64
	r.db.Model(&models.ActivityLog{}).Where("created_at >= ?", startDate).Count(&totalCount)

	// Get count by type
	countByType, err := r.CountByType(startDate, now)
	if err != nil {
		return nil, err
	}

	// Get top users
	topUsers, err := r.CountByUser(startDate, now, 10)
	if err != nil {
		return nil, err
	}

	// Get count by day
	type DayCount struct {
		Date  string
		Count int64
	}
	var dailyCounts []DayCount
	r.db.Model(&models.ActivityLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Find(&dailyCounts)

	return map[string]interface{}{
		"total_count":   totalCount,
		"count_by_type": countByType,
		"top_users":     topUsers,
		"daily_counts":  dailyCounts,
	}, nil
}

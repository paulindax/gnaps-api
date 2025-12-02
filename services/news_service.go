package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"

	"gorm.io/datatypes"
)

type NewsService struct {
	newsRepo    *repositories.NewsRepository
	commentRepo *repositories.CommentRepository
	userRepo    *repositories.UserRepository
}

func NewNewsService(newsRepo *repositories.NewsRepository, commentRepo *repositories.CommentRepository, userRepo *repositories.UserRepository) *NewsService {
	return &NewsService{
		newsRepo:    newsRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
	}
}

// ==================== HELPER FUNCTIONS ====================

// containsInJSON checks if a value exists in a JSON array
func containsInJSON(jsonData *datatypes.JSON, value int64) bool {
	if jsonData == nil {
		return false
	}

	var arr []int64
	if err := json.Unmarshal(*jsonData, &arr); err != nil {
		return false
	}

	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// GetAccessibleEntities gets user's accessible entity IDs based on role
func (s *NewsService) GetAccessibleEntities(userId uint, userRole string) (regionIds []int64, zoneIds []int64, groupIds []int64, schoolIds []int64, err error) {
	// Get user details
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Parse user_properties to get assigned entities
	var userProps map[string]interface{}
	if user.UserProperties != nil {
		if err := json.Unmarshal(*user.UserProperties, &userProps); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	// Extract entity IDs based on role
	switch userRole {
	case "system_admin", "national_admin":
		// These roles can see all news
		return nil, nil, nil, nil, nil

	case "regional_admin":
		// Get region_id from user properties
		if regionId, ok := userProps["region_id"].(float64); ok {
			regionIds = append(regionIds, int64(regionId))

			// Get all zones under this region
			zones, _ := s.newsRepo.GetZonesByRegion(int64(regionId))
			for _, zone := range zones {
				zoneIds = append(zoneIds, int64(zone.ID))

				// Get all schools under these zones
				schools, _ := s.newsRepo.GetSchoolsByZone(int64(zone.ID))
				for _, school := range schools {
					schoolIds = append(schoolIds, int64(school.ID))
				}
			}
		}

	case "zone_admin":
		// Get zone_id from user properties
		if zoneId, ok := userProps["zone_id"].(float64); ok {
			zoneIds = append(zoneIds, int64(zoneId))

			// Get all schools under this zone
			schools, _ := s.newsRepo.GetSchoolsByZone(int64(zoneId))
			for _, school := range schools {
				schoolIds = append(schoolIds, int64(school.ID))
			}
		}

	case "school_user":
		// Get school_id from user properties
		if schoolId, ok := userProps["school_id"].(float64); ok {
			schoolIds = append(schoolIds, int64(schoolId))
		}
	}

	return regionIds, zoneIds, groupIds, schoolIds, nil
}

// CanAccessNews checks if user can access a news item
func (s *NewsService) CanAccessNews(newsItem models.New, regionIds, zoneIds, groupIds, schoolIds []int64, userRole string) bool {
	// System and National admins can see all news
	if userRole == "system_admin" || userRole == "national_admin" {
		return true
	}

	// Check if news is targeted to user's accessible entities
	// News is accessible if ANY of the targeting matches

	// Check regions
	if len(regionIds) > 0 && newsItem.RegionIds != nil {
		for _, id := range regionIds {
			if containsInJSON(newsItem.RegionIds, id) {
				return true
			}
		}
	}

	// Check zones
	if len(zoneIds) > 0 && newsItem.ZoneIds != nil {
		for _, id := range zoneIds {
			if containsInJSON(newsItem.ZoneIds, id) {
				return true
			}
		}
	}

	// Check schools
	if len(schoolIds) > 0 && newsItem.SchoolIds != nil {
		for _, id := range schoolIds {
			if containsInJSON(newsItem.SchoolIds, id) {
				return true
			}
		}
	}

	return false
}

// ValidateTargeting validates news targeting based on user role
func (s *NewsService) ValidateTargeting(userId uint, userRole string, newsItem *models.New) error {
	// Get user details
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return fmt.Errorf("failed to get user details")
	}

	var userProps map[string]interface{}
	if user.UserProperties != nil {
		if err := json.Unmarshal(*user.UserProperties, &userProps); err != nil {
			return fmt.Errorf("failed to parse user properties")
		}
	}

	switch userRole {
	case "system_admin", "national_admin":
		// Can target any regions, zones, groups, schools
		return nil

	case "regional_admin":
		// Can only target their region and zones/schools under it
		userRegionId, ok := userProps["region_id"].(float64)
		if !ok {
			return fmt.Errorf("regional admin must have a region assigned")
		}

		// Validate region targeting
		if newsItem.RegionIds != nil {
			var regionIds []int64
			if err := json.Unmarshal(*newsItem.RegionIds, &regionIds); err == nil {
				for _, id := range regionIds {
					if id != int64(userRegionId) {
						return fmt.Errorf("you can only target your assigned region")
					}
				}
			}
		}

		// Validate zone targeting
		if newsItem.ZoneIds != nil {
			var zoneIds []int64
			if err := json.Unmarshal(*newsItem.ZoneIds, &zoneIds); err == nil {
				for _, zoneId := range zoneIds {
					zone, err := s.newsRepo.GetZoneByID(zoneId)
					if err != nil || zone.RegionId == nil || *zone.RegionId != int64(userRegionId) {
						return fmt.Errorf("you can only target zones under your region")
					}
				}
			}
		}

		// Validate school targeting
		if newsItem.SchoolIds != nil {
			var schoolIds []int64
			if err := json.Unmarshal(*newsItem.SchoolIds, &schoolIds); err == nil {
				for _, schoolId := range schoolIds {
					school, err := s.newsRepo.GetSchoolByID(schoolId)
					if err != nil {
						return fmt.Errorf("invalid school ID")
					}
					if school.ZoneId != nil {
						zone, err := s.newsRepo.GetZoneByID(*school.ZoneId)
						if err != nil || zone.RegionId == nil || *zone.RegionId != int64(userRegionId) {
							return fmt.Errorf("you can only target schools under your region")
						}
					}
				}
			}
		}

	case "zone_admin":
		// Can only target their zone and schools under it
		userZoneId, ok := userProps["zone_id"].(float64)
		if !ok {
			return fmt.Errorf("zone admin must have a zone assigned")
		}

		// Cannot target regions
		if newsItem.RegionIds != nil {
			return fmt.Errorf("zone admins cannot target regions")
		}

		// Validate zone targeting
		if newsItem.ZoneIds != nil {
			var zoneIds []int64
			if err := json.Unmarshal(*newsItem.ZoneIds, &zoneIds); err == nil {
				for _, id := range zoneIds {
					if id != int64(userZoneId) {
						return fmt.Errorf("you can only target your assigned zone")
					}
				}
			}
		}

		// Validate school targeting
		if newsItem.SchoolIds != nil {
			var schoolIds []int64
			if err := json.Unmarshal(*newsItem.SchoolIds, &schoolIds); err == nil {
				for _, schoolId := range schoolIds {
					school, err := s.newsRepo.GetSchoolByID(schoolId)
					if err != nil || school.ZoneId == nil || *school.ZoneId != int64(userZoneId) {
						return fmt.Errorf("you can only target schools under your zone")
					}
				}
			}
		}

	default:
		return fmt.Errorf("invalid user role")
	}

	return nil
}

// ==================== NEWS OPERATIONS ====================

// ListNews retrieves news with role-based filtering
func (s *NewsService) ListNews(userId uint, userRole string, filters map[string]interface{}, page, limit int) ([]models.New, int64, error) {
	// Filter by status for non-admins
	if userRole != "system_admin" && userRole != "national_admin" {
		filters["status"] = "published"
	}

	// Get user's accessible entities
	regionIds, zoneIds, groupIds, schoolIds, err := s.GetAccessibleEntities(userId, userRole)
	if err != nil {
		return nil, 0, err
	}

	// Get all news matching filters (without pagination)
	allNews, err := s.newsRepo.ListAll(filters)
	if err != nil {
		return nil, 0, err
	}

	// Filter by role-based access
	// Initialize as empty slice (not nil) to ensure JSON marshals as [] not null
	accessibleNews := make([]models.New, 0)
	for _, item := range allNews {
		if s.CanAccessNews(item, regionIds, zoneIds, groupIds, schoolIds, userRole) {
			accessibleNews = append(accessibleNews, item)
		}
	}

	// Apply pagination to filtered results
	total := int64(len(accessibleNews))
	start := (page - 1) * limit
	end := start + limit

	if start >= len(accessibleNews) {
		return []models.New{}, total, nil
	}

	if end > len(accessibleNews) {
		end = len(accessibleNews)
	}

	return accessibleNews[start:end], total, nil
}

// GetNewsByID retrieves a single news item with access check
func (s *NewsService) GetNewsByID(id uint, userId uint, userRole string) (*models.New, error) {
	newsItem, err := s.newsRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("news not found")
	}

	// Check role-based access
	regionIds, zoneIds, groupIds, schoolIds, err := s.GetAccessibleEntities(userId, userRole)
	if err != nil {
		return nil, err
	}

	if !s.CanAccessNews(*newsItem, regionIds, zoneIds, groupIds, schoolIds, userRole) {
		return nil, errors.New("you do not have permission to view this news")
	}

	return newsItem, nil
}

// CreateNews creates a new news item with validation
func (s *NewsService) CreateNews(news *models.New, userId uint, userRole string) error {
	// Only admins can create news
	if userRole == "school_user" {
		return errors.New("you do not have permission to create news")
	}

	// Validate required fields
	if news.Title == nil || *news.Title == "" {
		return errors.New("title is required")
	}
	if news.Content == nil || *news.Content == "" {
		return errors.New("content is required")
	}

	// Set author
	authorId := int64(userId)
	news.AuthorId = &authorId

	// Validate targeting based on role
	if err := s.ValidateTargeting(userId, userRole, news); err != nil {
		return err
	}

	// Set default values
	falseVal := false
	news.IsDeleted = &falseVal

	if news.Status == nil {
		defaultStatus := "draft"
		news.Status = &defaultStatus
	}

	if news.Featured == nil {
		news.Featured = &falseVal
	}

	return s.newsRepo.Create(news)
}

// UpdateNews updates an existing news item with validation
func (s *NewsService) UpdateNews(id uint, updates map[string]interface{}, userId uint, userRole string) error {
	// Only admins can update news
	if userRole == "school_user" {
		return errors.New("you do not have permission to update news")
	}

	// Verify news exists
	_, err := s.newsRepo.FindByID(id)
	if err != nil {
		return errors.New("news not found")
	}

	// If targeting is being updated, validate it
	var newsItem models.New
	if regionIds, ok := updates["region_ids"]; ok {
		newsItem.RegionIds = regionIds.(*datatypes.JSON)
	}
	if zoneIds, ok := updates["zone_ids"]; ok {
		newsItem.ZoneIds = zoneIds.(*datatypes.JSON)
	}
	if schoolIds, ok := updates["school_ids"]; ok {
		newsItem.SchoolIds = schoolIds.(*datatypes.JSON)
	}

	if newsItem.RegionIds != nil || newsItem.ZoneIds != nil || newsItem.SchoolIds != nil {
		if err := s.ValidateTargeting(userId, userRole, &newsItem); err != nil {
			return err
		}
	}

	return s.newsRepo.Update(id, updates)
}

// DeleteNews soft deletes a news item
func (s *NewsService) DeleteNews(id uint, userRole string) error {
	// Only admins can delete news
	if userRole == "school_user" {
		return errors.New("you do not have permission to delete news")
	}

	// Verify news exists
	_, err := s.newsRepo.FindByID(id)
	if err != nil {
		return errors.New("news not found")
	}

	return s.newsRepo.Delete(id)
}

// ==================== COMMENT OPERATIONS ====================

// ListComments retrieves comments with filters and pagination
func (s *NewsService) ListComments(filters map[string]interface{}, page, limit int) ([]models.NewsComment, int64, error) {
	return s.commentRepo.List(filters, page, limit)
}

// GetCommentByID retrieves a single comment by ID
func (s *NewsService) GetCommentByID(id uint) (*models.NewsComment, error) {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("comment not found")
	}
	return comment, nil
}

// CreateComment creates a new comment with validation
func (s *NewsService) CreateComment(comment *models.NewsComment) error {
	// Validate required fields
	if comment.Content == nil || *comment.Content == "" {
		return errors.New("content is required")
	}
	if comment.NewsId == nil {
		return errors.New("news ID is required")
	}

	// Verify that the news item exists
	exists, err := s.commentRepo.VerifyNewsExists(*comment.NewsId)
	if err != nil || !exists {
		return errors.New("invalid news ID - News does not exist")
	}

	// Verify that the user exists if user_id is provided
	if comment.UserId != nil {
		exists, err := s.commentRepo.VerifyUserExists(*comment.UserId)
		if err != nil || !exists {
			return errors.New("invalid user ID - User does not exist")
		}
	}

	// Set default values
	falseVal := false
	comment.IsDeleted = &falseVal
	comment.IsApproved = &falseVal // Comments start as unapproved

	return s.commentRepo.Create(comment)
}

// UpdateComment updates an existing comment
func (s *NewsService) UpdateComment(id uint, updates map[string]interface{}) error {
	// Verify comment exists
	_, err := s.commentRepo.FindByID(id)
	if err != nil {
		return errors.New("comment not found")
	}

	return s.commentRepo.Update(id, updates)
}

// DeleteComment soft deletes a comment
func (s *NewsService) DeleteComment(id uint) error {
	// Verify comment exists
	_, err := s.commentRepo.FindByID(id)
	if err != nil {
		return errors.New("comment not found")
	}

	return s.commentRepo.Delete(id)
}

// ApproveComment approves a comment
func (s *NewsService) ApproveComment(id uint) error {
	// Verify comment exists
	_, err := s.commentRepo.FindByID(id)
	if err != nil {
		return errors.New("comment not found")
	}

	return s.commentRepo.Approve(id)
}

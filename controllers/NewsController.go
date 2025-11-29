package controllers

import (
	"encoding/json"
	"fmt"
	"gnaps-api/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

type NewsController struct {
}

func init() {
	RegisterController("news", &NewsController{})
}

func (n *NewsController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	// News actions
	case "list":
		return n.list(c)
	case "show":
		return n.show(c)
	case "create":
		return n.create(c)
	case "update":
		return n.update(c)
	case "delete":
		return n.delete(c)
	// News Comments actions
	case "list_comments":
		return n.listComments(c)
	case "show_comment":
		return n.showComment(c)
	case "create_comment":
		return n.createComment(c)
	case "update_comment":
		return n.updateComment(c)
	case "delete_comment":
		return n.deleteComment(c)
	case "approve_comment":
		return n.approveComment(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// ==================== NEWS ENDPOINTS ====================

// Helper function to check if a value exists in a JSON array
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

// Helper function to get user's accessible entity IDs based on role
func (n *NewsController) getAccessibleEntities(c *fiber.Ctx) (regionIds []int64, zoneIds []int64, groupIds []int64, schoolIds []int64, err error) {
	// Get user info from context (set by auth middleware)
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	// Get user details
	var user models.User
	if err := DB.Where("id = ?", userId).First(&user).Error; err != nil {
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
			var zones []models.Zone
			DB.Where("region_id = ? AND is_deleted = ?", int64(regionId), false).Find(&zones)
			for _, zone := range zones {
				zoneIds = append(zoneIds, int64(zone.ID))

				// Get all schools under these zones
				var schools []models.School
				DB.Where("zone_id = ? AND is_deleted = ?", int64(zone.ID), false).Find(&schools)
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
			var schools []models.School
			DB.Where("zone_id = ? AND is_deleted = ?", int64(zoneId), false).Find(&schools)
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

// Helper function to filter news based on user's accessible entities
func (n *NewsController) canAccessNews(newsItem models.New, regionIds, zoneIds, groupIds, schoolIds []int64, userRole string) bool {
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

	// Check groups
	// if len(groupIds) > 0 && newsItem.GroupIds != nil {
	// 	for _, id := range groupIds {
	// 		if containsInJSON(newsItem.GroupIds, id) {
	// 			return true
	// 		}
	// 	}
	// }

	// Check schools
	if len(schoolIds) > 0 && newsItem.SchoolIds != nil {
		for _, id := range schoolIds {
			if containsInJSON(newsItem.SchoolIds, id) {
				return true
			}
		}
	}

	// If no targeting is set on the news, it's visible to all
	// if newsItem.RegionIds == nil && newsItem.ZoneIds == nil && newsItem.GroupIds == nil && newsItem.SchoolIds == nil {
	// 	return true
	// }

	return false
}

// list retrieves news with role-based filtering
func (n *NewsController) list(c *fiber.Ctx) error {
	var news []models.New
	userRole, _ := c.Locals("role").(string)

	// Get user's accessible entities
	regionIds, zoneIds, groupIds, schoolIds, err := n.getAccessibleEntities(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get user entities",
			"details": err.Error(),
		})
	}

	// Base query
	query := DB.Where("is_deleted = ?", false)

	// Filter by status - only show published news to non-admins
	if userRole != "system_admin" && userRole != "national_admin" {
		query = query.Where("status = ?", "published")
	}

	// Optional filters from query params
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if featured := c.Query("featured"); featured != "" {
		query = query.Where("featured = ?", featured == "true" || featured == "1")
	}

	if executiveID := c.Query("executive_id"); executiveID != "" {
		query = query.Where("executive_id = ?", executiveID)
	}

	if title := c.Query("title"); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	if content := c.Query("content"); content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	// Get all news matching filters
	var allNews []models.New
	query.Order("created_at DESC").Find(&allNews)

	// Filter by role-based access
	for _, item := range allNews {
		if n.canAccessNews(item, regionIds, zoneIds, groupIds, schoolIds, userRole) {
			news = append(news, item)
		}
	}

	// Apply pagination to filtered results
	total := int64(len(news))
	start := offset
	end := offset + limit

	if start > len(news) {
		news = []models.New{}
	} else {
		if end > len(news) {
			end = len(news)
		}
		news = news[start:end]
	}

	return c.JSON(fiber.Map{
		"data": news,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// show retrieves a single news item by ID with role-based access check
func (n *NewsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	userRole, _ := c.Locals("role").(string)

	var newsItem models.New
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&newsItem)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "News not found",
		})
	}

	// Check role-based access
	regionIds, zoneIds, groupIds, schoolIds, err := n.getAccessibleEntities(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to verify access",
			"details": err.Error(),
		})
	}

	if !n.canAccessNews(newsItem, regionIds, zoneIds, groupIds, schoolIds, userRole) {
		return c.Status(403).JSON(fiber.Map{
			"error": "You do not have permission to view this news",
		})
	}

	return c.JSON(fiber.Map{
		"data": newsItem,
	})
}

// create creates a new news item with role-based targeting validation
func (n *NewsController) create(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	// Only admins can create news
	if userRole == "school_user" {
		return c.Status(403).JSON(fiber.Map{
			"error": "You do not have permission to create news",
		})
	}

	var newsItem models.New

	if err := c.BodyParser(&newsItem); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if newsItem.Title == nil || *newsItem.Title == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if newsItem.Content == nil || *newsItem.Content == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	// Set author
	authorId := int64(userId)
	newsItem.AuthorId = &authorId

	// Validate targeting based on role
	if err := n.validateTargeting(c, &newsItem); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Set default values
	falseVal := false
	newsItem.IsDeleted = &falseVal

	if newsItem.Status == nil {
		defaultStatus := "draft"
		newsItem.Status = &defaultStatus
	}

	if newsItem.Featured == nil {
		newsItem.Featured = &falseVal
	}

	result := DB.Create(&newsItem)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create news",
			"details": result.Error.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "News created successfully",
		"flash_message": fiber.Map{
			"msg":  "News created successfully",
			"type": "success",
		},
		"data": newsItem,
	})
}

// validateTargeting validates news targeting based on user role
func (n *NewsController) validateTargeting(c *fiber.Ctx, newsItem *models.New) error {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	// Get user details
	var user models.User
	if err := DB.Where("id = ?", userId).First(&user).Error; err != nil {
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
					var zone models.Zone
					if err := DB.Where("id = ? AND region_id = ?", zoneId, int64(userRegionId)).First(&zone).Error; err != nil {
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
					var school models.School
					var zone models.Zone
					if err := DB.Where("id = ?", schoolId).First(&school).Error; err != nil {
						return fmt.Errorf("invalid school ID")
					}
					if school.ZoneId != nil {
						if err := DB.Where("id = ? AND region_id = ?", *school.ZoneId, int64(userRegionId)).First(&zone).Error; err != nil {
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
					var school models.School
					if err := DB.Where("id = ? AND zone_id = ?", schoolId, int64(userZoneId)).First(&school).Error; err != nil {
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

// update updates an existing news item with role-based validation
func (n *NewsController) update(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)

	// Only admins can update news
	if userRole == "school_user" {
		return c.Status(403).JSON(fiber.Map{
			"error": "You do not have permission to update news",
		})
	}

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var newsItem models.New
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&newsItem)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "News not found",
		})
	}

	var updateData models.New
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate targeting if being updated
	if updateData.RegionIds != nil || updateData.ZoneIds != nil || updateData.SchoolIds != nil {
		if err := n.validateTargeting(c, &updateData); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	// Update fields
	updates := make(map[string]interface{})
	if updateData.Title != nil {
		updates["title"] = updateData.Title
	}
	if updateData.Content != nil {
		updates["content"] = updateData.Content
	}
	if updateData.Excerpt != nil {
		updates["excerpt"] = updateData.Excerpt
	}
	if updateData.ImageUrl != nil {
		updates["image_url"] = updateData.ImageUrl
	}
	if updateData.Category != nil {
		updates["category"] = updateData.Category
	}
	if updateData.Status != nil {
		updates["status"] = updateData.Status
	}
	if updateData.Featured != nil {
		updates["featured"] = updateData.Featured
	}
	if updateData.ExecutiveId != nil {
		updates["executive_id"] = updateData.ExecutiveId
	}
	if updateData.RegionIds != nil {
		updates["region_ids"] = updateData.RegionIds
	}
	if updateData.ZoneIds != nil {
		updates["zone_ids"] = updateData.ZoneIds
	}
	// if updateData.GroupIds != nil {
	// 	updates["group_ids"] = updateData.GroupIds
	// }
	if updateData.SchoolIds != nil {
		updates["school_ids"] = updateData.SchoolIds
	}

	result = DB.Model(&newsItem).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update news",
			"details": result.Error.Error(),
		})
	}

	// Fetch updated record
	DB.First(&newsItem, id)

	return c.JSON(fiber.Map{
		"message": "News updated successfully",
		"flash_message": fiber.Map{
			"msg":  "News updated successfully",
			"type": "success",
		},
		"data": newsItem,
	})
}

// delete soft deletes a news item
func (n *NewsController) delete(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)

	// Only admins can delete news
	if userRole == "school_user" {
		return c.Status(403).JSON(fiber.Map{
			"error": "You do not have permission to delete news",
		})
	}

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var newsItem models.New
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&newsItem)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "News not found",
		})
	}

	// Soft delete
	trueVal := true
	result = DB.Model(&newsItem).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete news",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "News deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "News deleted successfully",
			"type": "success",
		},
	})
}

// ==================== NEWS COMMENTS ENDPOINTS ====================

// listComments retrieves all comments (optionally filtered by news_id)
func (n *NewsController) listComments(c *fiber.Ctx) error {
	var comments []models.NewsComment

	// Query parameters for filtering
	query := DB.Where("is_deleted = ?", false)

	// Optional filter by news_id
	if newsID := c.Query("news_id"); newsID != "" {
		query = query.Where("news_id = ?", newsID)
	}

	// Optional filter by user_id
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Optional filter by approval status
	if isApproved := c.Query("is_approved"); isApproved != "" {
		approvedVal := isApproved == "true" || isApproved == "1"
		query = query.Where("is_approved = ?", approvedVal)
	}

	// Optional search by content
	if content := c.Query("content"); content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.NewsComment{}).Count(&total)

	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&comments)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve comments",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": comments,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// showComment retrieves a single comment by ID
func (n *NewsController) showComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var comment models.NewsComment
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&comment)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": comment,
	})
}

// createComment creates a new comment on a news item
func (n *NewsController) createComment(c *fiber.Ctx) error {
	var comment models.NewsComment

	if err := c.BodyParser(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if comment.Content == nil || *comment.Content == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	if comment.NewsId == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "News ID is required",
		})
	}

	// Verify that the news item exists
	var newsItem models.New
	if err := DB.Where("id = ? AND is_deleted = ?", comment.NewsId, false).First(&newsItem).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid news ID - News does not exist",
		})
	}

	// Verify that the user exists if user_id is provided
	if comment.UserId != nil {
		var user models.User
		if err := DB.Where("id = ? AND is_deleted = ?", comment.UserId, false).First(&user).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid user ID - User does not exist",
			})
		}
	}

	// Set default values
	falseVal := false
	comment.IsDeleted = &falseVal
	comment.IsApproved = &falseVal // Comments start as unapproved

	result := DB.Create(&comment)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create comment",
			"details": result.Error.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Comment created successfully",
		"flash_message": fiber.Map{
			"msg":  "Comment posted successfully. It will appear after approval.",
			"type": "success",
		},
		"data": comment,
	})
}

// updateComment updates an existing comment
func (n *NewsController) updateComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var comment models.NewsComment
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&comment)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	var updateData models.NewsComment
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if updateData.Content != nil {
		updates["content"] = updateData.Content
	}
	if updateData.IsApproved != nil {
		updates["is_approved"] = updateData.IsApproved
	}

	result = DB.Model(&comment).Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update comment",
			"details": result.Error.Error(),
		})
	}

	// Fetch updated record
	DB.First(&comment, id)

	return c.JSON(fiber.Map{
		"message": "Comment updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Comment updated successfully",
			"type": "success",
		},
		"data": comment,
	})
}

// deleteComment soft deletes a comment
func (n *NewsController) deleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var comment models.NewsComment
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&comment)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	// Soft delete
	trueVal := true
	result = DB.Model(&comment).Update("is_deleted", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete comment",
			"details": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Comment deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Comment deleted successfully",
			"type": "success",
		},
	})
}

// approveComment approves a comment
func (n *NewsController) approveComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var comment models.NewsComment
	result := DB.Where("id = ? AND is_deleted = ?", id, false).First(&comment)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	// Set is_approved to true
	trueVal := true
	result = DB.Model(&comment).Update("is_approved", &trueVal)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to approve comment",
			"details": result.Error.Error(),
		})
	}

	// Fetch updated record
	DB.First(&comment, id)

	return c.JSON(fiber.Map{
		"message": "Comment approved successfully",
		"flash_message": fiber.Map{
			"msg":  "Comment approved successfully",
			"type": "success",
		},
		"data": comment,
	})
}

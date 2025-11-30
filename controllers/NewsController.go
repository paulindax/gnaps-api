package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type NewsController struct {
	newsService *services.NewsService
}

func NewNewsController(newsService *services.NewsService) *NewsController {
	return &NewsController{
		newsService: newsService,
	}
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
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown action %s", action)})
	}
}

// ==================== NEWS ENDPOINTS ====================

func (n *NewsController) list(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	// Parse filters from query params
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if featured := c.Query("featured"); featured != "" {
		filters["featured"] = featured == "true" || featured == "1"
	}
	if executiveID := c.Query("executive_id"); executiveID != "" {
		filters["executive_id"] = executiveID
	}
	if title := c.Query("title"); title != "" {
		filters["title"] = title
	}
	if content := c.Query("content"); content != "" {
		filters["content"] = content
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	news, total, err := n.newsService.ListNews(userId, userRole, filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get user entities",
			"details": err.Error(),
		})
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

func (n *NewsController) show(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	newsItem, err := n.newsService.GetNewsByID(uint(newsId), userId, userRole)
	if err != nil {
		if err.Error() == "news not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": newsItem})
}

func (n *NewsController) create(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	var newsItem models.New
	if err := c.BodyParser(&newsItem); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := n.newsService.CreateNews(&newsItem, userId, userRole); err != nil {
		if err.Error() == "you do not have permission to create news" {
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
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

func (n *NewsController) update(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.New
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Build updates map
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
	if updateData.SchoolIds != nil {
		updates["school_ids"] = updateData.SchoolIds
	}

	if err := n.newsService.UpdateNews(uint(newsId), updates, userId, userRole); err != nil {
		if err.Error() == "news not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "you do not have permission to update news" {
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated news item
	newsItem, _ := n.newsService.GetNewsByID(uint(newsId), userId, userRole)

	return c.JSON(fiber.Map{
		"message": "News updated successfully",
		"flash_message": fiber.Map{
			"msg":  "News updated successfully",
			"type": "success",
		},
		"data": newsItem,
	})
}

func (n *NewsController) delete(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := n.newsService.DeleteNews(uint(newsId), userRole); err != nil {
		if err.Error() == "news not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "you do not have permission to delete news" {
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
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

func (n *NewsController) listComments(c *fiber.Ctx) error {
	// Parse filters from query params
	filters := make(map[string]interface{})
	if newsID := c.Query("news_id"); newsID != "" {
		filters["news_id"] = newsID
	}
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if isApproved := c.Query("is_approved"); isApproved != "" {
		approvedVal := isApproved == "true" || isApproved == "1"
		filters["is_approved"] = approvedVal
	}
	if content := c.Query("content"); content != "" {
		filters["content"] = content
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	comments, total, err := n.newsService.ListComments(filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve comments",
			"details": err.Error(),
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

func (n *NewsController) showComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	comment, err := n.newsService.GetCommentByID(uint(commentId))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": comment})
}

func (n *NewsController) createComment(c *fiber.Ctx) error {
	var comment models.NewsComment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := n.newsService.CreateComment(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
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

func (n *NewsController) updateComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	var updateData models.NewsComment
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Build updates map
	updates := make(map[string]interface{})
	if updateData.Content != nil {
		updates["content"] = updateData.Content
	}
	if updateData.IsApproved != nil {
		updates["is_approved"] = updateData.IsApproved
	}

	if err := n.newsService.UpdateComment(uint(commentId), updates); err != nil {
		if err.Error() == "comment not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated comment
	comment, _ := n.newsService.GetCommentByID(uint(commentId))

	return c.JSON(fiber.Map{
		"message": "Comment updated successfully",
		"flash_message": fiber.Map{
			"msg":  "Comment updated successfully",
			"type": "success",
		},
		"data": comment,
	})
}

func (n *NewsController) deleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := n.newsService.DeleteComment(uint(commentId)); err != nil {
		if err.Error() == "comment not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Comment deleted successfully",
		"flash_message": fiber.Map{
			"msg":  "Comment deleted successfully",
			"type": "success",
		},
	})
}

func (n *NewsController) approveComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ID"})
	}

	if err := n.newsService.ApproveComment(uint(commentId)); err != nil {
		if err.Error() == "comment not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated comment
	comment, _ := n.newsService.GetCommentByID(uint(commentId))

	return c.JSON(fiber.Map{
		"message": "Comment approved successfully",
		"flash_message": fiber.Map{
			"msg":  "Comment approved successfully",
			"type": "success",
		},
		"data": comment,
	})
}

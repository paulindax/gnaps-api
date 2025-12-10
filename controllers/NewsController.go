package controllers

import (
	"fmt"
	"gnaps-api/models"
	"gnaps-api/services"
	"gnaps-api/utils"
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
	// Owner-based news actions (uses executive owner context)
	case "owner_list":
		return n.ownerList(c)
	case "owner_show":
		return n.ownerShow(c)
	case "owner_create":
		return n.ownerCreate(c)
	case "owner_update":
		return n.ownerUpdate(c)
	case "owner_delete":
		return n.ownerDelete(c)
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
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
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
		return utils.ServerErrorResponse(c, "Failed to retrieve news")
	}

	// Ensure news is never null (always return empty array instead of null)
	if news == nil {
		news = []models.New{}
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
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	newsItem, err := n.newsService.GetNewsByID(uint(newsId), userId, userRole)
	if err != nil {
		if err.Error() == "news not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ForbiddenResponse(c, err.Error())
	}

	return c.JSON(fiber.Map{"data": newsItem})
}

func (n *NewsController) create(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	var newsItem models.New
	if err := c.BodyParser(&newsItem); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := n.newsService.CreateNews(&newsItem, userId, userRole); err != nil {
		if err.Error() == "you do not have permission to create news" {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return utils.ValidationErrorResponse(c, err.Error())
	}

	return utils.SuccessResponseWithStatus(c, 201, newsItem, "News created successfully")
}

func (n *NewsController) update(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)
	userId, _ := c.Locals("user_id").(uint)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	var updateData models.New
	if err := c.BodyParser(&updateData); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
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
			return utils.NotFoundResponse(c, err.Error())
		}
		if err.Error() == "you do not have permission to update news" {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return utils.ValidationErrorResponse(c, err.Error())
	}

	// Get updated news item
	newsItem, _ := n.newsService.GetNewsByID(uint(newsId), userId, userRole)

	return utils.SuccessResponse(c, newsItem, "News updated successfully")
}

func (n *NewsController) delete(c *fiber.Ctx) error {
	userRole, _ := c.Locals("role").(string)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	if err := n.newsService.DeleteNews(uint(newsId), userRole); err != nil {
		if err.Error() == "news not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		if err.Error() == "you do not have permission to delete news" {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, nil, "News deleted successfully")
}

// ==================== OWNER-BASED NEWS ENDPOINTS ====================
// These endpoints use executive owner context for data filtering
// Zone admins see zone data, Region admins see region data, National admins see all

func (n *NewsController) ownerList(c *fiber.Ctx) error {
	// Get owner context from middleware
	ownerCtx := utils.GetOwnerContext(c)

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
	if title := c.Query("title"); title != "" {
		filters["title"] = title
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	news, total, err := n.newsService.ListNewsWithOwner(filters, page, limit, ownerCtx)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to retrieve news")
	}

	if news == nil {
		news = []models.New{}
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

func (n *NewsController) ownerShow(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	newsItem, err := n.newsService.GetNewsByIDWithOwner(uint(newsId), ownerCtx)
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return c.JSON(fiber.Map{"data": newsItem})
}

func (n *NewsController) ownerCreate(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	var newsItem models.New
	if err := c.BodyParser(&newsItem); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := n.newsService.CreateNewsWithOwner(&newsItem, ownerCtx); err != nil {
		if err.Error() == "system admin cannot modify data in owner-based tables (view only)" {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return utils.ValidationErrorResponse(c, err.Error())
	}

	return utils.SuccessResponseWithStatus(c, 201, newsItem, "News created successfully")
}

func (n *NewsController) ownerUpdate(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	var updateData models.New
	if err := c.BodyParser(&updateData); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
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

	if err := n.newsService.UpdateNewsWithOwner(uint(newsId), updates, ownerCtx); err != nil {
		if err.Error() == "system admin cannot modify data in owner-based tables (view only)" {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return utils.NotFoundResponse(c, err.Error())
	}

	// Get updated news item
	newsItem, _ := n.newsService.GetNewsByIDWithOwner(uint(newsId), ownerCtx)

	return utils.SuccessResponse(c, newsItem, "News updated successfully")
}

func (n *NewsController) ownerDelete(c *fiber.Ctx) error {
	ownerCtx := utils.GetOwnerContext(c)

	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	newsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	if err := n.newsService.DeleteNewsWithOwner(uint(newsId), ownerCtx); err != nil {
		if err.Error() == "system admin cannot modify data in owner-based tables (view only)" {
			return utils.ForbiddenResponse(c, err.Error())
		}
		return utils.NotFoundResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, nil, "News deleted successfully")
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
		return utils.ServerErrorResponse(c, "Failed to retrieve comments")
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
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	comment, err := n.newsService.GetCommentByID(uint(commentId))
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return c.JSON(fiber.Map{"data": comment})
}

func (n *NewsController) createComment(c *fiber.Ctx) error {
	var comment models.NewsComment
	if err := c.BodyParser(&comment); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := n.newsService.CreateComment(&comment); err != nil {
		return utils.ValidationErrorResponse(c, err.Error())
	}

	return utils.SuccessResponseWithStatus(c, 201, comment, "Comment posted successfully. It will appear after approval.")
}

func (n *NewsController) updateComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	var updateData models.NewsComment
	if err := c.BodyParser(&updateData); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
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
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	// Get updated comment
	comment, _ := n.newsService.GetCommentByID(uint(commentId))

	return utils.SuccessResponse(c, comment, "Comment updated successfully")
}

func (n *NewsController) deleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	if err := n.newsService.DeleteComment(uint(commentId)); err != nil {
		if err.Error() == "comment not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, nil, "Comment deleted successfully")
}

func (n *NewsController) approveComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		return utils.ValidationErrorResponse(c, "ID is required")
	}

	commentId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid ID")
	}

	if err := n.newsService.ApproveComment(uint(commentId)); err != nil {
		if err.Error() == "comment not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.ServerErrorResponse(c, err.Error())
	}

	// Get updated comment
	comment, _ := n.newsService.GetCommentByID(uint(commentId))

	return utils.SuccessResponse(c, comment, "Comment approved successfully")
}

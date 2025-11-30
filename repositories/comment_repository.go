package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment
func (r *CommentRepository) Create(comment *models.NewsComment) error {
	return r.db.Create(comment).Error
}

// FindByID retrieves a comment by ID
func (r *CommentRepository) FindByID(id uint) (*models.NewsComment, error) {
	var comment models.NewsComment
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// List retrieves all comments with filters and pagination
func (r *CommentRepository) List(filters map[string]interface{}, page, limit int) ([]models.NewsComment, int64, error) {
	var comments []models.NewsComment
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Apply filters
	if newsID, ok := filters["news_id"]; ok {
		query = query.Where("news_id = ?", newsID)
	}
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if isApproved, ok := filters["is_approved"]; ok {
		query = query.Where("is_approved = ?", isApproved)
	}
	if content, ok := filters["content"]; ok {
		query = query.Where("content LIKE ?", "%"+content.(string)+"%")
	}

	// Count total before pagination
	query.Model(&models.NewsComment{}).Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// Update updates a comment
func (r *CommentRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.NewsComment{}).Where("id = ?", id).Updates(updates).Error
}

// Delete soft deletes a comment
func (r *CommentRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.NewsComment{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

// Approve approves a comment
func (r *CommentRepository) Approve(id uint) error {
	trueVal := true
	return r.db.Model(&models.NewsComment{}).Where("id = ?", id).Update("is_approved", &trueVal).Error
}

// VerifyNewsExists checks if a news item exists
func (r *CommentRepository) VerifyNewsExists(newsID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.New{}).Where("id = ? AND is_deleted = ?", newsID, false).Count(&count).Error
	return count > 0, err
}

// VerifyUserExists checks if a user exists
func (r *CommentRepository) VerifyUserExists(userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("id = ? AND is_deleted = ?", userID, false).Count(&count).Error
	return count > 0, err
}

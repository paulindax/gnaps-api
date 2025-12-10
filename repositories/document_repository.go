package repositories

import (
	"gnaps-api/models"
	"gnaps-api/utils"

	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Document methods

func (r *DocumentRepository) FindDocumentByID(id uint) (*models.Document, error) {
	var document models.Document
	err := r.db.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) ListDocuments(filters map[string]interface{}, page, limit int) ([]models.Document, int64, error) {
	var documents []models.Document
	var total int64

	query := r.db.Where("is_deleted = ? OR is_deleted IS NULL", false)

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	query.Model(&models.Document{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&documents).Error

	return documents, total, err
}

func (r *DocumentRepository) CreateDocument(document *models.Document) error {
	return r.db.Create(document).Error
}

func (r *DocumentRepository) UpdateDocument(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

func (r *DocumentRepository) DeleteDocument(id uint) error {
	isDeleted := true
	return r.db.Model(&models.Document{}).Where("id = ?", id).Update("is_deleted", isDeleted).Error
}

func (r *DocumentRepository) GetSubmissionCount(documentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.DocumentSubmission{}).
		Where("document_id = ? AND (is_deleted = ? OR is_deleted IS NULL)", documentID, false).
		Count(&count).Error
	return count, err
}

// DocumentSubmission methods

func (r *DocumentRepository) FindSubmissionByID(id uint) (*models.DocumentSubmission, error) {
	var submission models.DocumentSubmission
	err := r.db.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false).First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *DocumentRepository) ListSubmissions(filters map[string]interface{}, page, limit int) ([]models.DocumentSubmission, int64, error) {
	var submissions []models.DocumentSubmission
	var total int64

	query := r.db.Where("is_deleted = ? OR is_deleted IS NULL", false)

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	query.Model(&models.DocumentSubmission{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&submissions).Error

	return submissions, total, err
}

func (r *DocumentRepository) CreateSubmission(submission *models.DocumentSubmission) error {
	return r.db.Create(submission).Error
}

func (r *DocumentRepository) UpdateSubmission(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.DocumentSubmission{}).Where("id = ?", id).Updates(updates).Error
}

func (r *DocumentRepository) DeleteSubmission(id uint) error {
	isDeleted := true
	return r.db.Model(&models.DocumentSubmission{}).Where("id = ?", id).Update("is_deleted", isDeleted).Error
}

// Helper methods for enrichment

func (r *DocumentRepository) GetDocumentTitle(documentID *int64) (*string, error) {
	if documentID == nil {
		return nil, nil
	}
	var document models.Document
	err := r.db.Select("title").Where("id = ?", *documentID).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document.Title, nil
}

func (r *DocumentRepository) GetSchoolName(schoolID *int64) (*string, error) {
	if schoolID == nil {
		return nil, nil
	}
	var school models.School
	err := r.db.Select("name").Where("id = ?", *schoolID).First(&school).Error
	if err != nil {
		return nil, err
	}
	return &school.Name, nil
}

func (r *DocumentRepository) GetUserName(userID *int64) (*string, error) {
	if userID == nil {
		return nil, nil
	}
	var user models.User
	err := r.db.Select("username").Where("id = ?", *userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user.Username, nil
}

// ============================================
// Owner-based methods for data filtering
// ============================================

// CreateDocumentWithOwner creates a new document with owner fields automatically set
func (r *DocumentRepository) CreateDocumentWithOwner(document *models.Document, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	if ownerCtx != nil && ownerCtx.IsValid() {
		ownerType, ownerID := ownerCtx.GetOwnerValues()
		document.SetOwner(ownerType, ownerID)
	}
	return r.db.Create(document).Error
}

// FindDocumentByIDWithOwner retrieves a document by ID with owner filtering
func (r *DocumentRepository) FindDocumentByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.Document, error) {
	var document models.Document
	query := r.db.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	err := query.First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// ListDocumentsWithOwner retrieves documents with filters, pagination, and owner filtering
func (r *DocumentRepository) ListDocumentsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Document, int64, error) {
	var documents []models.Document
	var total int64

	query := r.db.Model(&models.Document{}).Where("is_deleted = ? OR is_deleted IS NULL", false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&documents).Error

	return documents, total, err
}

// UpdateDocumentWithOwner updates a document with owner verification
func (r *DocumentRepository) UpdateDocumentWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.Document{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	result := query.Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// DeleteDocumentWithOwner soft deletes a document with owner verification
func (r *DocumentRepository) DeleteDocumentWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.Document{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	isDeleted := true
	result := query.Update("is_deleted", isDeleted)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// CreateSubmissionWithOwner creates a submission with owner verification
func (r *DocumentRepository) CreateSubmissionWithOwner(submission *models.DocumentSubmission, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	if ownerCtx != nil && ownerCtx.IsValid() {
		ownerType, ownerID := ownerCtx.GetOwnerValues()
		submission.SetOwner(ownerType, ownerID)
	}
	return r.db.Create(submission).Error
}

// FindSubmissionByIDWithOwner retrieves a submission by ID with owner filtering
func (r *DocumentRepository) FindSubmissionByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.DocumentSubmission, error) {
	var submission models.DocumentSubmission
	query := r.db.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	err := query.First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// ListSubmissionsWithOwner retrieves submissions with filters, pagination, and owner filtering
func (r *DocumentRepository) ListSubmissionsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.DocumentSubmission, int64, error) {
	var submissions []models.DocumentSubmission
	var total int64

	query := r.db.Model(&models.DocumentSubmission{}).Where("is_deleted = ? OR is_deleted IS NULL", false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&submissions).Error

	return submissions, total, err
}

// UpdateSubmissionWithOwner updates a submission with owner verification
func (r *DocumentRepository) UpdateSubmissionWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.DocumentSubmission{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	result := query.Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// DeleteSubmissionWithOwner soft deletes a submission with owner verification
func (r *DocumentRepository) DeleteSubmissionWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.DocumentSubmission{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	isDeleted := true
	result := query.Update("is_deleted", isDeleted)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

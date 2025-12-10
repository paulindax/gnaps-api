package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
)

type DocumentService struct {
	documentRepo *repositories.DocumentRepository
}

func NewDocumentService(documentRepo *repositories.DocumentRepository) *DocumentService {
	return &DocumentService{documentRepo: documentRepo}
}

// Document methods

func (s *DocumentService) GetDocumentByID(id uint) (*models.Document, error) {
	document, err := s.documentRepo.FindDocumentByID(id)
	if err != nil {
		return nil, errors.New("document not found")
	}

	// Get submission count
	count, err := s.documentRepo.GetSubmissionCount(id)
	if err == nil {
		document.SubmissionCount = int(count)
	}

	return document, nil
}

func (s *DocumentService) ListDocuments(filters map[string]interface{}, page, limit int) ([]models.Document, int64, error) {
	documents, total, err := s.documentRepo.ListDocuments(filters, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Get submission counts for each document
	for i := range documents {
		count, err := s.documentRepo.GetSubmissionCount(documents[i].ID)
		if err == nil {
			documents[i].SubmissionCount = int(count)
		}
	}

	return documents, total, nil
}

func (s *DocumentService) CreateDocument(document *models.Document, userID uint) error {
	// Set created_by
	document.CreatedBy = int64(userID)

	return s.documentRepo.CreateDocument(document)
}

func (s *DocumentService) UpdateDocument(id uint, updates map[string]interface{}) error {
	// Verify document exists
	_, err := s.documentRepo.FindDocumentByID(id)
	if err != nil {
		return errors.New("document not found")
	}

	return s.documentRepo.UpdateDocument(id, updates)
}

func (s *DocumentService) DeleteDocument(id uint) error {
	// Verify document exists
	_, err := s.documentRepo.FindDocumentByID(id)
	if err != nil {
		return errors.New("document not found")
	}

	return s.documentRepo.DeleteDocument(id)
}

// DocumentSubmission methods

func (s *DocumentService) GetSubmissionByID(id uint) (*models.DocumentSubmission, error) {
	submission, err := s.documentRepo.FindSubmissionByID(id)
	if err != nil {
		return nil, errors.New("submission not found")
	}

	// Enrich with related data
	s.enrichSubmission(submission)

	return submission, nil
}

func (s *DocumentService) ListSubmissions(filters map[string]interface{}, page, limit int) ([]models.DocumentSubmission, int64, error) {
	submissions, total, err := s.documentRepo.ListSubmissions(filters, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Enrich with related data
	for i := range submissions {
		s.enrichSubmission(&submissions[i])
	}

	return submissions, total, nil
}

func (s *DocumentService) CreateSubmission(submission *models.DocumentSubmission, userID uint) error {
	// Set submitted_by
	submission.SubmittedBy = int64(userID)

	return s.documentRepo.CreateSubmission(submission)
}

func (s *DocumentService) UpdateSubmission(id uint, updates map[string]interface{}) error {
	// Verify submission exists
	_, err := s.documentRepo.FindSubmissionByID(id)
	if err != nil {
		return errors.New("submission not found")
	}

	return s.documentRepo.UpdateSubmission(id, updates)
}

func (s *DocumentService) DeleteSubmission(id uint) error {
	// Verify submission exists
	_, err := s.documentRepo.FindSubmissionByID(id)
	if err != nil {
		return errors.New("submission not found")
	}

	return s.documentRepo.DeleteSubmission(id)
}

func (s *DocumentService) ReviewSubmission(id uint, status string, reviewNotes *string, userID uint) error {
	// Verify submission exists
	_, err := s.documentRepo.FindSubmissionByID(id)
	if err != nil {
		return errors.New("submission not found")
	}

	reviewedBy := int64(userID)
	updates := map[string]interface{}{
		"status":       status,
		"review_notes": reviewNotes,
		"reviewed_by":  reviewedBy,
	}

	return s.documentRepo.UpdateSubmission(id, updates)
}

// Helper methods

func (s *DocumentService) enrichSubmission(submission *models.DocumentSubmission) {
	// Get document title
	if title, err := s.documentRepo.GetDocumentTitle(&submission.DocumentId); err == nil && title != nil {
		submission.DocumentTitle = title
	}

	// Get school name
	if name, err := s.documentRepo.GetSchoolName(&submission.SchoolId); err == nil && name != nil {
		submission.SchoolName = name
	}

	// Get submitter name
	if name, err := s.documentRepo.GetUserName(&submission.SubmittedBy); err == nil && name != nil {
		submission.SubmitterName = name
	}
}

// ============================================
// Owner-based methods for data filtering
// ============================================

// Document owner methods

func (s *DocumentService) GetDocumentByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.Document, error) {
	document, err := s.documentRepo.FindDocumentByIDWithOwner(id, ownerCtx)
	if err != nil {
		return nil, err
	}

	// Get submission count
	count, err := s.documentRepo.GetSubmissionCount(id)
	if err == nil {
		document.SubmissionCount = int(count)
	}

	return document, nil
}

func (s *DocumentService) ListDocumentsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Document, int64, error) {
	documents, total, err := s.documentRepo.ListDocumentsWithOwner(filters, page, limit, ownerCtx)
	if err != nil {
		return nil, 0, err
	}

	// Get submission counts for each document
	for i := range documents {
		count, err := s.documentRepo.GetSubmissionCount(documents[i].ID)
		if err == nil {
			documents[i].SubmissionCount = int(count)
		}
	}

	return documents, total, nil
}

func (s *DocumentService) CreateDocumentWithOwner(document *models.Document, userID uint, ownerCtx *utils.OwnerContext) error {
	document.CreatedBy = int64(userID)
	return s.documentRepo.CreateDocumentWithOwner(document, ownerCtx)
}

func (s *DocumentService) UpdateDocumentWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	return s.documentRepo.UpdateDocumentWithOwner(id, updates, ownerCtx)
}

func (s *DocumentService) DeleteDocumentWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	return s.documentRepo.DeleteDocumentWithOwner(id, ownerCtx)
}

// Submission owner methods

func (s *DocumentService) GetSubmissionByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.DocumentSubmission, error) {
	submission, err := s.documentRepo.FindSubmissionByIDWithOwner(id, ownerCtx)
	if err != nil {
		return nil, err
	}

	s.enrichSubmission(submission)
	return submission, nil
}

func (s *DocumentService) ListSubmissionsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.DocumentSubmission, int64, error) {
	submissions, total, err := s.documentRepo.ListSubmissionsWithOwner(filters, page, limit, ownerCtx)
	if err != nil {
		return nil, 0, err
	}

	for i := range submissions {
		s.enrichSubmission(&submissions[i])
	}

	return submissions, total, nil
}

func (s *DocumentService) CreateSubmissionWithOwner(submission *models.DocumentSubmission, userID uint, ownerCtx *utils.OwnerContext) error {
	submission.SubmittedBy = int64(userID)
	return s.documentRepo.CreateSubmissionWithOwner(submission, ownerCtx)
}

func (s *DocumentService) UpdateSubmissionWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	return s.documentRepo.UpdateSubmissionWithOwner(id, updates, ownerCtx)
}

func (s *DocumentService) DeleteSubmissionWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	return s.documentRepo.DeleteSubmissionWithOwner(id, ownerCtx)
}

func (s *DocumentService) ReviewSubmissionWithOwner(id uint, status string, reviewNotes *string, userID uint, ownerCtx *utils.OwnerContext) error {
	reviewedBy := int64(userID)
	updates := map[string]interface{}{
		"status":       status,
		"review_notes": reviewNotes,
		"reviewed_by":  reviewedBy,
	}
	return s.documentRepo.UpdateSubmissionWithOwner(id, updates, ownerCtx)
}

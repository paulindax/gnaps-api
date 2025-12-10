package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
)

type BillParticularService struct {
	particularRepo *repositories.BillParticularRepository
}

func NewBillParticularService(particularRepo *repositories.BillParticularRepository) *BillParticularService {
	return &BillParticularService{particularRepo: particularRepo}
}

func (s *BillParticularService) GetParticularByID(id uint) (*models.BillParticular, error) {
	particular, err := s.particularRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("bill particular not found")
	}
	return particular, nil
}

func (s *BillParticularService) ListParticulars(filters map[string]interface{}, page, limit int) ([]models.BillParticular, int64, error) {
	return s.particularRepo.List(filters, page, limit)
}

func (s *BillParticularService) CreateParticular(particular *models.BillParticular) error {
	// Validate required fields
	if particular.Name == nil || *particular.Name == "" {
		return errors.New("name is required")
	}

	// Set defaults
	isDeleted := false
	particular.IsDeleted = &isDeleted

	// If priority is not set, set it to max + 1
	if particular.Priority == nil {
		maxPriority, err := s.particularRepo.GetMaxPriority()
		if err != nil {
			return err
		}
		newPriority := maxPriority + 1
		particular.Priority = &newPriority
	}

	return s.particularRepo.Create(particular)
}

func (s *BillParticularService) UpdateParticular(id uint, updates map[string]interface{}) error {
	// Verify particular exists
	_, err := s.particularRepo.FindByID(id)
	if err != nil {
		return errors.New("bill particular not found")
	}

	return s.particularRepo.Update(id, updates)
}

func (s *BillParticularService) DeleteParticular(id uint) error {
	_, err := s.particularRepo.FindByID(id)
	if err != nil {
		return errors.New("bill particular not found")
	}

	return s.particularRepo.Delete(id)
}

// ============================================
// Owner-based methods for data filtering
// ============================================

func (s *BillParticularService) GetParticularByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.BillParticular, error) {
	return s.particularRepo.FindByIDWithOwner(id, ownerCtx)
}

func (s *BillParticularService) ListParticularsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.BillParticular, int64, error) {
	return s.particularRepo.ListWithOwner(filters, page, limit, ownerCtx)
}

func (s *BillParticularService) CreateParticularWithOwner(particular *models.BillParticular, ownerCtx *utils.OwnerContext) error {
	// Validate required fields
	if particular.Name == nil || *particular.Name == "" {
		return errors.New("name is required")
	}

	// Set defaults
	isDeleted := false
	particular.IsDeleted = &isDeleted

	// If priority is not set, set it to max + 1
	if particular.Priority == nil {
		maxPriority, err := s.particularRepo.GetMaxPriority()
		if err != nil {
			return err
		}
		newPriority := maxPriority + 1
		particular.Priority = &newPriority
	}

	return s.particularRepo.CreateWithOwner(particular, ownerCtx)
}

func (s *BillParticularService) UpdateParticularWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	return s.particularRepo.UpdateWithOwner(id, updates, ownerCtx)
}

func (s *BillParticularService) DeleteParticularWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	return s.particularRepo.DeleteWithOwner(id, ownerCtx)
}

package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
)

type ExecutiveService struct {
	executiveRepo *repositories.ExecutiveRepository
}

func NewExecutiveService(executiveRepo *repositories.ExecutiveRepository) *ExecutiveService {
	return &ExecutiveService{executiveRepo: executiveRepo}
}

func (s *ExecutiveService) GetExecutiveByID(id uint) (*models.Executive, error) {
	executive, err := s.executiveRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("executive not found")
	}
	return executive, nil
}

func (s *ExecutiveService) ListExecutives(filters map[string]interface{}, page, limit int) ([]models.Executive, int64, error) {
	return s.executiveRepo.List(filters, page, limit)
}

func (s *ExecutiveService) CreateExecutive(executive *models.Executive) error {
	// Validate required fields
	if executive.ExecutiveNo == nil || *executive.ExecutiveNo == "" {
		return errors.New("executive number is required")
	}
	if executive.FirstName == nil || *executive.FirstName == "" {
		return errors.New("first name is required")
	}
	if executive.LastName == nil || *executive.LastName == "" {
		return errors.New("last name is required")
	}

	// Check if executive_no already exists
	exists, err := s.executiveRepo.ExecutiveNoExists(*executive.ExecutiveNo, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("executive with this executive number already exists")
	}

	// Check if email already exists (if provided)
	if executive.Email != nil && *executive.Email != "" {
		exists, err := s.executiveRepo.EmailExists(*executive.Email, nil)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("executive with this email already exists")
		}
	}

	// Validate gender if provided
	if executive.Gender != nil && *executive.Gender != "" {
		validGenders := []string{"Male", "Female", "Other"}
		isValid := false
		for _, g := range validGenders {
			if *executive.Gender == g {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New("invalid gender. Must be one of: Male, Female, Other")
		}
	}

	// Set defaults
	falseVal := false
	executive.IsDeleted = &falseVal

	return s.executiveRepo.Create(executive)
}

func (s *ExecutiveService) UpdateExecutive(id uint, updates map[string]interface{}) error {
	// Verify executive exists
	executive, err := s.executiveRepo.FindByID(id)
	if err != nil {
		return errors.New("executive not found")
	}

	// Check if executive_no is being changed
	if executiveNo, ok := updates["executive_no"]; ok {
		executiveNoPtr := executiveNo.(*string)
		if executiveNoPtr != nil && *executiveNoPtr != "" {
			if executive.ExecutiveNo == nil || *executiveNoPtr != *executive.ExecutiveNo {
				exists, err := s.executiveRepo.ExecutiveNoExists(*executiveNoPtr, &id)
				if err != nil {
					return err
				}
				if exists {
					return errors.New("executive with this executive number already exists")
				}
			}
		}
	}

	// Check if email is being changed
	if email, ok := updates["email"]; ok {
		emailPtr := email.(*string)
		if emailPtr != nil && *emailPtr != "" {
			if executive.Email == nil || *emailPtr != *executive.Email {
				exists, err := s.executiveRepo.EmailExists(*emailPtr, &id)
				if err != nil {
					return err
				}
				if exists {
					return errors.New("executive with this email already exists")
				}
			}
		}
	}

	// Validate gender if being changed
	if gender, ok := updates["gender"]; ok {
		genderPtr := gender.(*string)
		if genderPtr != nil && *genderPtr != "" {
			validGenders := []string{"Male", "Female", "Other"}
			isValid := false
			for _, g := range validGenders {
				if *genderPtr == g {
					isValid = true
					break
				}
			}
			if !isValid {
				return errors.New("invalid gender. Must be one of: Male, Female, Other")
			}
		}
	}

	return s.executiveRepo.Update(id, updates)
}

func (s *ExecutiveService) DeleteExecutive(id uint) error {
	_, err := s.executiveRepo.FindByID(id)
	if err != nil {
		return errors.New("executive not found")
	}

	return s.executiveRepo.Delete(id)
}

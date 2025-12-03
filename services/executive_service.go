package services

import (
	"errors"
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"time"
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
	if executive.FirstName == nil || *executive.FirstName == "" {
		return errors.New("first name is required")
	}
	if executive.LastName == nil || *executive.LastName == "" {
		return errors.New("last name is required")
	}
	if executive.Email == nil || *executive.Email == "" {
		return errors.New("email is required")
	}
	if executive.Role == nil || *executive.Role == "" {
		return errors.New("role is required")
	}

	// Validate role
	validRoles := []string{"national_admin", "region_admin", "zone_admin"}
	isValidRole := false
	for _, r := range validRoles {
		if *executive.Role == r {
			isValidRole = true
			break
		}
	}
	if !isValidRole {
		return errors.New("invalid role. Must be one of: national_admin, region_admin, zone_admin")
	}

	// Validate role-based assignments
	if *executive.Role == "region_admin" && (executive.RegionId == nil || *executive.RegionId == 0) {
		return errors.New("region is required for region admin")
	}
	if *executive.Role == "zone_admin" {
		if executive.RegionId == nil || *executive.RegionId == 0 {
			return errors.New("region is required for zonal admin")
		}
		if executive.ZoneId == nil || *executive.ZoneId == 0 {
			return errors.New("zone is required for zonal admin")
		}
	}

	// Generate executive number if not provided
	if executive.ExecutiveNo == nil || *executive.ExecutiveNo == "" {
		execNo := fmt.Sprintf("EXEC-%d", time.Now().UnixNano()/1000000)
		executive.ExecutiveNo = &execNo
	} else {
		// Check if executive_no already exists
		exists, err := s.executiveRepo.ExecutiveNoExists(*executive.ExecutiveNo, nil)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("executive with this executive number already exists")
		}
	}

	// Check if email already exists
	exists, err := s.executiveRepo.EmailExists(*executive.Email, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("executive with this email already exists")
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

	if executive.Status == nil || *executive.Status == "" {
		activeStatus := "active"
		executive.Status = &activeStatus
	}

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
		if executiveNoStr, isStr := executiveNo.(string); isStr && executiveNoStr != "" {
			if executive.ExecutiveNo == nil || executiveNoStr != *executive.ExecutiveNo {
				exists, err := s.executiveRepo.ExecutiveNoExists(executiveNoStr, &id)
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
		if emailStr, isStr := email.(string); isStr && emailStr != "" {
			if executive.Email == nil || emailStr != *executive.Email {
				exists, err := s.executiveRepo.EmailExists(emailStr, &id)
				if err != nil {
					return err
				}
				if exists {
					return errors.New("executive with this email already exists")
				}
			}
		}
	}

	// Validate role if being changed
	if role, ok := updates["role"]; ok {
		if roleStr, isStr := role.(string); isStr && roleStr != "" {
			validRoles := []string{"national_admin", "region_admin", "zone_admin"}
			isValid := false
			for _, r := range validRoles {
				if roleStr == r {
					isValid = true
					break
				}
			}
			if !isValid {
				return errors.New("invalid role. Must be one of: national_admin, region_admin, zone_admin")
			}
		}
	}

	// Validate status if being changed
	if status, ok := updates["status"]; ok {
		if statusStr, isStr := status.(string); isStr && statusStr != "" {
			validStatuses := []string{"active", "inactive"}
			isValid := false
			for _, s := range validStatuses {
				if statusStr == s {
					isValid = true
					break
				}
			}
			if !isValid {
				return errors.New("invalid status. Must be one of: active, inactive")
			}
		}
	}

	// Validate gender if being changed
	if gender, ok := updates["gender"]; ok {
		if genderStr, isStr := gender.(string); isStr && genderStr != "" {
			validGenders := []string{"Male", "Female", "Other"}
			isValid := false
			for _, g := range validGenders {
				if genderStr == g {
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

package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
)

type SchoolService struct {
	schoolRepo *repositories.SchoolRepository
}

func NewSchoolService(schoolRepo *repositories.SchoolRepository) *SchoolService {
	return &SchoolService{schoolRepo: schoolRepo}
}

// Search searches for schools by keyword
func (s *SchoolService) Search(keyword string, limit int) ([]models.School, error) {
	return s.schoolRepo.Search(keyword, limit)
}

// GetSchoolByID retrieves a school by ID
func (s *SchoolService) GetSchoolByID(id uint) (*models.School, error) {
	school, err := s.schoolRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("school not found")
	}
	return school, nil
}

// ListSchools retrieves paginated schools with filters
func (s *SchoolService) ListSchools(filters map[string]interface{}, page, limit int) ([]models.School, int64, error) {
	return s.schoolRepo.List(filters, page, limit)
}

// CreateSchool creates a new school with validation
func (s *SchoolService) CreateSchool(school *models.School) error {
	// Validate required fields
	if school.Name == "" {
		return errors.New("name is required")
	}
	if school.MemberNo == "" {
		return errors.New("member number is required")
	}

	// Verify that the zone exists if zone_id is provided
	if school.ZoneId != nil {
		exists, err := s.schoolRepo.VerifyZoneExists(*school.ZoneId)
		if err != nil || !exists {
			return errors.New("invalid zone ID - Zone does not exist")
		}
	}

	// Check if member_no already exists
	exists, err := s.schoolRepo.MemberNoExists(school.MemberNo, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("school with this member number already exists")
	}

	// Check if email already exists (if provided)
	if school.Email != nil && *school.Email != "" {
		exists, err := s.schoolRepo.EmailExists(*school.Email, nil)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("school with this email already exists")
		}
	}

	// Set defaults
	isDeleted := false
	school.IsDeleted = &isDeleted

	return s.schoolRepo.Create(school)
}

// UpdateSchool updates an existing school with validation
func (s *SchoolService) UpdateSchool(id uint, updates map[string]interface{}) error {
	// Verify school exists
	school, err := s.schoolRepo.FindByID(id)
	if err != nil {
		return errors.New("school not found")
	}

	// If zone_id is being changed, verify the new zone exists
	if zoneId, ok := updates["zone_id"]; ok {
		zoneIdVal := zoneId.(int64)
		if school.ZoneId == nil || zoneIdVal != *school.ZoneId {
			exists, err := s.schoolRepo.VerifyZoneExists(zoneIdVal)
			if err != nil || !exists {
				return errors.New("invalid zone ID - Zone does not exist")
			}
		}
	}

	// Check if member_no is being changed and if new member_no already exists
	if memberNo, ok := updates["member_no"]; ok {
		memberNoStr := memberNo.(string)
		if memberNoStr != "" && memberNoStr != school.MemberNo {
			exists, err := s.schoolRepo.MemberNoExists(memberNoStr, &id)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("school with this member number already exists")
			}
		}
	}

	// Check if email is being changed and if new email already exists
	if email, ok := updates["email"]; ok {
		emailStr := email.(string)
		if emailStr != "" {
			if school.Email == nil || emailStr != *school.Email {
				exists, err := s.schoolRepo.EmailExists(emailStr, &id)
				if err != nil {
					return err
				}
				if exists {
					return errors.New("school with this email already exists")
				}
			}
		}
	}

	return s.schoolRepo.Update(id, updates)
}

// DeleteSchool soft deletes a school
func (s *SchoolService) DeleteSchool(id uint) error {
	// Verify school exists
	_, err := s.schoolRepo.FindByID(id)
	if err != nil {
		return errors.New("school not found")
	}

	return s.schoolRepo.Delete(id)
}

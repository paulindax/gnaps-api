package services

import (
	"errors"
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
)

type SchoolService struct {
	schoolRepo *repositories.SchoolRepository
	userRepo   *repositories.UserRepository
}

func NewSchoolService(schoolRepo *repositories.SchoolRepository, userRepo *repositories.UserRepository) *SchoolService {
	return &SchoolService{
		schoolRepo: schoolRepo,
		userRepo:   userRepo,
	}
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

	// Create user account for the school if email is provided
	if school.Email != nil && *school.Email != "" {
		mobileNo := ""
		if school.MobileNo != nil {
			mobileNo = *school.MobileNo
		}

		user, err := s.userRepo.CreateUserForSchool(
			school.Name,
			*school.Email,
			mobileNo,
			school.MemberNo, // Password will be member_no + "123"
		)
		if err != nil {
			return fmt.Errorf("failed to create user account: %v", err)
		}

		// Set the user_id on the school
		userID := int64(user.ID)
		school.UserId = &userID
	}

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

	// Update the associated user if school has a user_id
	if school.UserId != nil && *school.UserId > 0 {
		userUpdates := make(map[string]interface{})

		// Sync relevant fields to user
		if name, ok := updates["name"]; ok {
			userUpdates["first_name"] = name // School name maps to user first_name
		}
		if email, ok := updates["email"]; ok {
			userUpdates["email"] = email
		}
		if mobileNo, ok := updates["mobile_no"]; ok {
			userUpdates["mobile_no"] = mobileNo
		}
		if memberNo, ok := updates["member_no"]; ok {
			userUpdates["username"] = memberNo // Member number maps to username
		}

		// Only update user if there are changes to sync
		if len(userUpdates) > 0 {
			if err := s.userRepo.Update(uint(*school.UserId), userUpdates); err != nil {
				return fmt.Errorf("failed to update user account: %v", err)
			}
		}
	}

	return s.schoolRepo.Update(id, updates)
}

// DeleteSchool soft deletes a school
func (s *SchoolService) DeleteSchool(id uint) error {
	// Verify school exists
	school, err := s.schoolRepo.FindByID(id)
	if err != nil {
		return errors.New("school not found")
	}

	// Also soft delete the associated user
	if school.UserId != nil && *school.UserId > 0 {
		if err := s.userRepo.Delete(uint(*school.UserId)); err != nil {
			return fmt.Errorf("failed to delete user account: %v", err)
		}
	}

	return s.schoolRepo.Delete(id)
}

// GetNextMemberNoForZone returns the next member number for a zone
func (s *SchoolService) GetNextMemberNoForZone(zoneID int64) (string, error) {
	return s.schoolRepo.GetNextMemberNoForZone(zoneID)
}

// ============================================
// Role-Based Filtering Methods
// ============================================

// ListSchoolsWithRole returns schools filtered by role-based access
func (s *SchoolService) ListSchoolsWithRole(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.School, int64, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	return s.schoolRepo.ListWithRoleFilter(filters, page, limit, regionID, zoneID)
}

// GetSchoolByIDWithRole returns a school if accessible by the user's role
func (s *SchoolService) GetSchoolByIDWithRole(id uint, ownerCtx *utils.OwnerContext) (*models.School, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	school, err := s.schoolRepo.FindByIDWithRoleFilter(id, regionID, zoneID)
	if err != nil {
		return nil, errors.New("school not found or access denied")
	}
	return school, nil
}

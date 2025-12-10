package services

import (
	"errors"
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
	"time"
)

type ExecutiveService struct {
	executiveRepo *repositories.ExecutiveRepository
	userRepo      *repositories.UserRepository
}

func NewExecutiveService(executiveRepo *repositories.ExecutiveRepository, userRepo *repositories.UserRepository) *ExecutiveService {
	return &ExecutiveService{
		executiveRepo: executiveRepo,
		userRepo:      userRepo,
	}
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

	// Create user account for the executive
	mobileNo := ""
	if executive.MobileNo != nil {
		mobileNo = *executive.MobileNo
	}

	user, err := s.userRepo.CreateUserForExecutive(
		*executive.FirstName,
		*executive.LastName,
		*executive.Email,
		mobileNo,
		*executive.Role,
		*executive.ExecutiveNo, // Password will be executive_no + "123"
	)
	if err != nil {
		return fmt.Errorf("failed to create user account: %v", err)
	}

	// Set the user_id on the executive
	userID := int64(user.ID)
	executive.UserId = &userID

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

	// Update the associated user if executive has a user_id
	if executive.UserId != nil && *executive.UserId > 0 {
		userUpdates := make(map[string]interface{})

		// Sync relevant fields to user
		if firstName, ok := updates["first_name"]; ok {
			userUpdates["first_name"] = firstName
		}
		if lastName, ok := updates["last_name"]; ok {
			userUpdates["last_name"] = lastName
		}
		if email, ok := updates["email"]; ok {
			userUpdates["email"] = email
		}
		if mobileNo, ok := updates["mobile_no"]; ok {
			userUpdates["mobile_no"] = mobileNo
		}
		if role, ok := updates["role"]; ok {
			userUpdates["role"] = role
		}

		// Only update user if there are changes to sync
		if len(userUpdates) > 0 {
			if err := s.userRepo.Update(uint(*executive.UserId), userUpdates); err != nil {
				return fmt.Errorf("failed to update user account: %v", err)
			}
		}
	}

	return s.executiveRepo.Update(id, updates)
}

func (s *ExecutiveService) DeleteExecutive(id uint) error {
	executive, err := s.executiveRepo.FindByID(id)
	if err != nil {
		return errors.New("executive not found")
	}

	// Also soft delete the associated user
	if executive.UserId != nil && *executive.UserId > 0 {
		if err := s.userRepo.Delete(uint(*executive.UserId)); err != nil {
			return fmt.Errorf("failed to delete user account: %v", err)
		}
	}

	return s.executiveRepo.Delete(id)
}

// ============================================
// Role-Based Filtering Methods
// ============================================

// ListExecutivesWithRole returns executives filtered by role-based access
func (s *ExecutiveService) ListExecutivesWithRole(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Executive, int64, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	return s.executiveRepo.ListWithRoleFilter(filters, page, limit, regionID, zoneID)
}

// GetExecutiveByIDWithRole returns an executive if accessible by the user's role
func (s *ExecutiveService) GetExecutiveByIDWithRole(id uint, ownerCtx *utils.OwnerContext) (*models.Executive, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	executive, err := s.executiveRepo.FindByIDWithRoleFilter(id, regionID, zoneID)
	if err != nil {
		return nil, errors.New("executive not found or access denied")
	}
	return executive, nil
}

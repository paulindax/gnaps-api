package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
)

type ContactPersonService struct {
	contactPersonRepo *repositories.ContactPersonRepository
}

func NewContactPersonService(contactPersonRepo *repositories.ContactPersonRepository) *ContactPersonService {
	return &ContactPersonService{contactPersonRepo: contactPersonRepo}
}

func (s *ContactPersonService) GetContactPersonByID(id uint) (*models.ContactPerson, error) {
	contactPerson, err := s.contactPersonRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("contact person not found")
	}
	return contactPerson, nil
}

func (s *ContactPersonService) ListContactPersons(filters map[string]interface{}, page, limit int) ([]models.ContactPerson, int64, error) {
	return s.contactPersonRepo.List(filters, page, limit)
}

func (s *ContactPersonService) CreateContactPerson(contactPerson *models.ContactPerson) error {
	// Validate required fields
	if contactPerson.SchoolId == nil {
		return errors.New("school ID is required")
	}
	if contactPerson.FirstName == nil || *contactPerson.FirstName == "" {
		return errors.New("first name is required")
	}
	if contactPerson.LastName == nil || *contactPerson.LastName == "" {
		return errors.New("last name is required")
	}

	// Verify school exists
	exists, err := s.contactPersonRepo.VerifySchoolExists(*contactPerson.SchoolId)
	if err != nil || !exists {
		return errors.New("invalid school ID - School does not exist")
	}

	// Check if email already exists for this school (if provided)
	if contactPerson.Email != nil && *contactPerson.Email != "" {
		exists, err := s.contactPersonRepo.EmailExistsForSchool(*contactPerson.Email, *contactPerson.SchoolId, nil)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("contact person with this email already exists for this school")
		}
	}

	return s.contactPersonRepo.Create(contactPerson)
}

func (s *ContactPersonService) UpdateContactPerson(id uint, updates map[string]interface{}) error {
	// Verify contact person exists
	contactPerson, err := s.contactPersonRepo.FindByID(id)
	if err != nil {
		return errors.New("contact person not found")
	}

	// If school_id is being changed, verify the new school exists
	schoolIDForCheck := contactPerson.SchoolId
	if schoolID, ok := updates["school_id"]; ok {
		schoolIDVal := schoolID.(int64)
		if contactPerson.SchoolId == nil || schoolIDVal != *contactPerson.SchoolId {
			exists, err := s.contactPersonRepo.VerifySchoolExists(schoolIDVal)
			if err != nil || !exists {
				return errors.New("invalid school ID - School does not exist")
			}
			schoolIDForCheck = &schoolIDVal
		}
	}

	// Check if email is being changed
	if email, ok := updates["email"]; ok {
		emailPtr := email.(*string)
		if emailPtr != nil && *emailPtr != "" {
			if contactPerson.Email == nil || *emailPtr != *contactPerson.Email {
				exists, err := s.contactPersonRepo.EmailExistsForSchool(*emailPtr, *schoolIDForCheck, &id)
				if err != nil {
					return err
				}
				if exists {
					return errors.New("contact person with this email already exists for this school")
				}
			}
		}
	}

	return s.contactPersonRepo.Update(id, updates)
}

func (s *ContactPersonService) DeleteContactPerson(id uint) error {
	_, err := s.contactPersonRepo.FindByID(id)
	if err != nil {
		return errors.New("contact person not found")
	}

	return s.contactPersonRepo.Delete(id)
}

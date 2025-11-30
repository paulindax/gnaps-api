package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type ContactPersonRepository struct {
	db *gorm.DB
}

func NewContactPersonRepository(db *gorm.DB) *ContactPersonRepository {
	return &ContactPersonRepository{db: db}
}

func (r *ContactPersonRepository) FindByID(id uint) (*models.ContactPerson, error) {
	var contactPerson models.ContactPerson
	err := r.db.First(&contactPerson, id).Error
	if err != nil {
		return nil, err
	}
	return &contactPerson, nil
}

func (r *ContactPersonRepository) List(filters map[string]interface{}, page, limit int) ([]models.ContactPerson, int64, error) {
	var contactPersons []models.ContactPerson
	var total int64

	query := r.db.Model(&models.ContactPerson{})

	// Apply filters
	for key, value := range filters {
		if key == "first_name" || key == "last_name" {
			query = query.Where(key+" LIKE ?", "%"+value.(string)+"%")
		} else if key == "name" {
			query = query.Where("CONCAT(first_name, ' ', last_name) LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&contactPersons).Error

	return contactPersons, total, err
}

func (r *ContactPersonRepository) Create(contactPerson *models.ContactPerson) error {
	return r.db.Create(contactPerson).Error
}

func (r *ContactPersonRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.ContactPerson{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ContactPersonRepository) Delete(id uint) error {
	var contactPerson models.ContactPerson
	return r.db.Delete(&contactPerson, id).Error
}

func (r *ContactPersonRepository) EmailExistsForSchool(email string, schoolID int64, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.ContactPerson{}).Where("email = ? AND school_id = ?", email, schoolID)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *ContactPersonRepository) VerifySchoolExists(schoolID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.School{}).Where("id = ? AND is_deleted = ?", schoolID, false).Count(&count).Error
	return count > 0, err
}

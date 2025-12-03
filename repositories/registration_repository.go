package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type RegistrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Create(registration *models.EventRegistration) error {
	return r.db.Create(registration).Error
}

func (r *RegistrationRepository) FindByID(id uint) (*models.EventRegistration, error) {
	var registration models.EventRegistration
	err := r.db.Where("id = ? AND COALESCE(is_deleted, false) = ?", id, false).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *RegistrationRepository) SchoolAlreadyRegistered(eventID uint, schoolID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.EventRegistration{}).
		Where("event_id = ? AND school_id = ? AND COALESCE(is_deleted, false) = ?", eventID, schoolID, false).
		Count(&count).Error
	return count > 0, err
}

func (r *RegistrationRepository) CountByEvent(eventID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.EventRegistration{}).
		Where("event_id = ? AND COALESCE(is_deleted, false) = ?", eventID, false).
		Count(&count).Error
	return count, err
}

func (r *RegistrationRepository) ListByEvent(eventID uint, page, limit int) ([]models.EventRegistration, int64, error) {
	var registrations []models.EventRegistration
	var total int64

	query := r.db.Model(&models.EventRegistration{}).
		Where("event_id = ? AND COALESCE(is_deleted, false) = ?", eventID, false)

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("registration_date DESC").Find(&registrations).Error

	return registrations, total, err
}

func (r *RegistrationRepository) ListByUser(userID uint, page, limit int) ([]models.EventRegistration, int64, error) {
	var registrations []models.EventRegistration
	var total int64

	query := r.db.Model(&models.EventRegistration{}).
		Where("registered_by = ? AND COALESCE(is_deleted, false) = ?", userID, false)

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("registration_date DESC").Find(&registrations).Error

	return registrations, total, err
}

func (r *RegistrationRepository) UpdatePaymentStatus(id uint, status, reference string) error {
	updates := map[string]interface{}{
		"payment_status": status,
	}
	if reference != "" {
		updates["payment_reference"] = reference
	}
	return r.db.Model(&models.EventRegistration{}).Where("id = ?", id).Updates(updates).Error
}

func (r *RegistrationRepository) Delete(id uint) error {
	return r.db.Model(&models.EventRegistration{}).Where("id = ?", id).Update("is_deleted", true).Error
}

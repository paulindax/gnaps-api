package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *EventRepository) FindByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) FindByCode(code string) (*models.Event, error) {
	var event models.Event
	err := r.db.Where("registration_code = ? AND is_deleted = ?", code, false).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) List(filters map[string]interface{}, page, limit int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	query := r.db.Model(&models.Event{}).Where("is_deleted = ?", false)

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	// Date range filters
	if fromDate, ok := filters["from_date"]; ok {
		query = query.Where("start_date >= ?", fromDate)
		delete(filters, "from_date")
	}
	if toDate, ok := filters["to_date"]; ok {
		query = query.Where("start_date <= ?", toDate)
		delete(filters, "to_date")
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("start_date DESC").Find(&events).Error

	return events, total, err
}

func (r *EventRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Event{}).Where("id = ?", id).Updates(updates).Error
}

func (r *EventRepository) UpdateStruct(event *models.Event) error {
	return r.db.Model(event).Updates(event).Error
}

func (r *EventRepository) Delete(id uint) error {
	return r.db.Model(&models.Event{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *EventRepository) CodeExists(code string) bool {
	var count int64
	r.db.Model(&models.Event{}).Where("registration_code = ?", code).Count(&count)
	return count > 0
}

func (r *EventRepository) GetRegisteredCount(eventID uint) int64 {
	var count int64
	r.db.Model(&models.EventRegistration{}).
		Where("event_id = ? AND is_deleted = ?", eventID, false).
		Count(&count)
	return count
}

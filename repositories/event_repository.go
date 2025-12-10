package repositories

import (
	"gnaps-api/models"
	"gnaps-api/utils"

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
		Where("event_id = ? AND COALESCE(is_deleted, false) = ?", eventID, false).
		Count(&count)
	return count
}

// GetBillName retrieves the bill name for a given bill_id
func (r *EventRepository) GetBillName(billID int64) *string {
	var bill struct {
		Name *string `gorm:"column:name"`
	}
	err := r.db.Table("bills").Where("id = ? AND is_deleted = ?", billID, false).First(&bill).Error
	if err != nil {
		return nil
	}
	return bill.Name
}

// ============================================
// Owner-based methods for data filtering
// ============================================

// CreateWithOwner creates a new event with owner fields automatically set
// Returns ErrSystemAdminCannotWrite if system_admin tries to create owner-based data
func (r *EventRepository) CreateWithOwner(event *models.Event, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	if ownerCtx != nil && ownerCtx.IsValid() {
		ownerType, ownerID := ownerCtx.GetOwnerValues()
		event.SetOwner(ownerType, ownerID)
	}
	return r.db.Create(event).Error
}

// FindByIDWithOwner retrieves an event by ID with owner filtering
func (r *EventRepository) FindByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.Event, error) {
	var event models.Event
	query := r.db.Where("id = ? AND is_deleted = ?", id, false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	err := query.First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ListWithOwner retrieves events with filters, pagination, and owner filtering
func (r *EventRepository) ListWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	query := r.db.Model(&models.Event{}).Where("is_deleted = ?", false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	// Date range filters - process and remove before general filter loop
	if fromDate, ok := filters["from_date"]; ok {
		query = query.Where("start_date >= ?", fromDate)
		delete(filters, "from_date")
	}
	if toDate, ok := filters["to_date"]; ok {
		query = query.Where("start_date <= ?", toDate)
		delete(filters, "to_date")
	}

	// Apply other filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("start_date DESC").Find(&events).Error

	return events, total, err
}

// UpdateWithOwner updates an event with owner verification
// Returns ErrSystemAdminCannotWrite if system_admin tries to update
func (r *EventRepository) UpdateWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.Event{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	result := query.Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// DeleteWithOwner soft deletes an event with owner verification
// Returns ErrSystemAdminCannotWrite if system_admin tries to delete
func (r *EventRepository) DeleteWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.Event{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	result := query.Update("is_deleted", true)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

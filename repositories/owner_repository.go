package repositories

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/utils"

	"gorm.io/gorm"
)

// ErrSystemAdminCannotWrite is returned when system_admin tries to write (create/update/delete) owner-based data
var ErrSystemAdminCannotWrite = errors.New("system admin cannot modify data in owner-based tables (view only)")

// ErrSystemAdminCannotCreate is an alias for backwards compatibility
var ErrSystemAdminCannotCreate = ErrSystemAdminCannotWrite

// OwnerRepository provides helper methods for owner-based data filtering
type OwnerRepository struct {
	db *gorm.DB
}

// NewOwnerRepository creates a new OwnerRepository
func NewOwnerRepository(db *gorm.DB) *OwnerRepository {
	return &OwnerRepository{db: db}
}

// ApplyOwnerFilter applies owner filtering to a GORM query based on the owner context
// Returns the modified query
// If ownerCtx is nil or national admin, no filter is applied (access all)
func (r *OwnerRepository) ApplyOwnerFilter(query *gorm.DB, ownerCtx *utils.OwnerContext) *gorm.DB {
	if ownerCtx == nil {
		return query
	}

	filter := ownerCtx.GetOwnerFilter()
	if filter == nil {
		// National admin - no filter
		return query
	}

	// Apply owner filter
	return query.Where("owner_type = ? AND owner_id = ?", filter.OwnerType, filter.OwnerID)
}

// ApplyOwnerFilterWithTable applies owner filtering with explicit table name
// Useful for joins where column names need to be qualified
func (r *OwnerRepository) ApplyOwnerFilterWithTable(query *gorm.DB, tableName string, ownerCtx *utils.OwnerContext) *gorm.DB {
	if ownerCtx == nil {
		return query
	}

	filter := ownerCtx.GetOwnerFilter()
	if filter == nil {
		// National admin - no filter
		return query
	}

	// Apply owner filter with table prefix
	return query.Where(tableName+".owner_type = ? AND "+tableName+".owner_id = ?", filter.OwnerType, filter.OwnerID)
}

// SetOwnerFields sets owner_type and owner_id on a map for creating/updating records
func (r *OwnerRepository) SetOwnerFields(data map[string]interface{}, ownerCtx *utils.OwnerContext) map[string]interface{} {
	if ownerCtx == nil || !ownerCtx.IsValid() {
		return data
	}

	ownerType, ownerID := ownerCtx.GetOwnerValues()
	data["owner_type"] = ownerType
	data["owner_id"] = ownerID
	return data
}

// OwnerFieldSetter is an interface for models that have owner fields
type OwnerFieldSetter interface {
	SetOwner(ownerType string, ownerID int64)
}

// SetOwnerOnModel sets owner fields on a model that implements OwnerFieldSetter
func (r *OwnerRepository) SetOwnerOnModel(model OwnerFieldSetter, ownerCtx *utils.OwnerContext) {
	if ownerCtx == nil || !ownerCtx.IsValid() {
		return
	}
	ownerType, ownerID := ownerCtx.GetOwnerValues()
	model.SetOwner(ownerType, ownerID)
}

// OwnerQueryBuilder provides a fluent interface for building owner-filtered queries
type OwnerQueryBuilder struct {
	db       *gorm.DB
	ownerCtx *utils.OwnerContext
}

// NewOwnerQueryBuilder creates a new OwnerQueryBuilder
func NewOwnerQueryBuilder(db *gorm.DB, ownerCtx *utils.OwnerContext) *OwnerQueryBuilder {
	return &OwnerQueryBuilder{
		db:       db,
		ownerCtx: ownerCtx,
	}
}

// Model starts a query on a specific model with owner filtering applied
func (b *OwnerQueryBuilder) Model(model interface{}) *gorm.DB {
	query := b.db.Model(model)
	return ApplyOwnerFilterToQuery(query, b.ownerCtx)
}

// Table starts a query on a specific table with owner filtering applied
func (b *OwnerQueryBuilder) Table(name string) *gorm.DB {
	query := b.db.Table(name)
	return ApplyOwnerFilterToQuery(query, b.ownerCtx)
}

// ApplyOwnerFilterToQuery is a standalone function to apply owner filtering
func ApplyOwnerFilterToQuery(query *gorm.DB, ownerCtx *utils.OwnerContext) *gorm.DB {
	if ownerCtx == nil {
		return query
	}

	filter := ownerCtx.GetOwnerFilter()
	if filter == nil {
		// National admin - no filter
		return query
	}

	return query.Where("owner_type = ? AND owner_id = ?", filter.OwnerType, filter.OwnerID)
}

// CreateWithOwner creates a record with owner fields automatically set
// Returns ErrSystemAdminCannotCreate if system_admin tries to create owner-based data
func CreateWithOwner(db *gorm.DB, model interface{}, ownerCtx *utils.OwnerContext) error {
	// Check if user can create owner-based data
	if ownerCtx != nil && !ownerCtx.CanCreateOwnerData() {
		return ErrSystemAdminCannotCreate
	}

	if setter, ok := model.(OwnerFieldSetter); ok && ownerCtx != nil && ownerCtx.IsValid() {
		ownerType, ownerID := ownerCtx.GetOwnerValues()
		setter.SetOwner(ownerType, ownerID)
	}
	return db.Create(model).Error
}

// GetOwnerValues is a helper to get owner values from context
func GetOwnerValues(ownerCtx *utils.OwnerContext) (ownerType string, ownerID int64) {
	if ownerCtx == nil || !ownerCtx.IsValid() {
		return "", 0
	}
	return ownerCtx.GetOwnerValues()
}

// CanWrite checks if the owner context allows writing data (create/update/delete)
// Returns error if system_admin (view only)
func CanWrite(ownerCtx *utils.OwnerContext) error {
	if ownerCtx != nil && !ownerCtx.CanWriteOwnerData() {
		return ErrSystemAdminCannotWrite
	}
	return nil
}

// CanCreate is an alias for CanWrite for backwards compatibility
func CanCreate(ownerCtx *utils.OwnerContext) error {
	return CanWrite(ownerCtx)
}

// UpdateWithOwner updates a record with owner verification
// Returns ErrSystemAdminCannotWrite if system_admin tries to update
// Returns gorm.ErrRecordNotFound if record not found or user doesn't have access
func UpdateWithOwner(db *gorm.DB, model interface{}, id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	// Check if user can write owner-based data
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := db.Model(model).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	result := query.Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// DeleteWithOwner soft deletes a record with owner verification
// Returns ErrSystemAdminCannotWrite if system_admin tries to delete
// Uses is_deleted field for soft delete
func DeleteWithOwner(db *gorm.DB, model interface{}, id uint, ownerCtx *utils.OwnerContext) error {
	// Check if user can write owner-based data
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := db.Model(model).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	trueVal := true
	result := query.Update("is_deleted", &trueVal)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// FindByIDWithOwner retrieves a single record by ID with owner filtering
func FindByIDWithOwner(db *gorm.DB, dest interface{}, id uint, ownerCtx *utils.OwnerContext) error {
	query := db.Where("id = ? AND (is_deleted = ? OR is_deleted IS NULL)", id, false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)
	return query.First(dest).Error
}

// ListWithOwner retrieves records with owner filtering and pagination
func ListWithOwner(db *gorm.DB, dest interface{}, baseQuery *gorm.DB, page, limit int, ownerCtx *utils.OwnerContext) (int64, error) {
	var total int64

	// Apply owner filter
	query := ApplyOwnerFilterToQuery(baseQuery, ownerCtx)

	// Count total
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(dest).Error

	return total, err
}

// OwnerInfo holds owner_type and owner_id from an entity
type OwnerInfo struct {
	OwnerType string
	OwnerID   int64
}

// GetOwnerFromPayee looks up ownership from a payee entity (for MomoPayment)
// Payee can be: SchoolBill, EventRegistration, Event
// Returns nil if ownership cannot be determined
func GetOwnerFromPayee(db *gorm.DB, payeeType string, payeeID int64) *OwnerInfo {
	if payeeID == 0 {
		return nil
	}

	switch payeeType {
	case "SchoolBillPayment", "SchoolBill":
		// For school bill payment, get ownership from the Bill (via SchoolBill.BillId)
		var schoolBill models.SchoolBill
		if err := db.Select("bill_id, owner_type, owner_id").First(&schoolBill, payeeID).Error; err != nil {
			return nil
		}
		// First try to get ownership from SchoolBill itself
		if schoolBill.OwnerType != nil && schoolBill.OwnerId != nil && *schoolBill.OwnerId > 0 {
			return &OwnerInfo{OwnerType: *schoolBill.OwnerType, OwnerID: *schoolBill.OwnerId}
		}
		// Fall back to Bill's ownership
		if schoolBill.BillId != nil {
			var bill models.Bill
			if err := db.Select("owner_type, owner_id").First(&bill, *schoolBill.BillId).Error; err == nil {
				if bill.OwnerType != nil && bill.OwnerId != nil && *bill.OwnerId > 0 {
					return &OwnerInfo{OwnerType: *bill.OwnerType, OwnerID: *bill.OwnerId}
				}
			}
		}

	case "EventRegistration":
		// Get ownership from EventRegistration itself, or fall back to Event
		var registration models.EventRegistration
		if err := db.Select("event_id, owner_type, owner_id").First(&registration, payeeID).Error; err != nil {
			return nil
		}
		// First try to get ownership from EventRegistration itself
		if registration.OwnerType != nil && registration.OwnerId != nil && *registration.OwnerId > 0 {
			return &OwnerInfo{OwnerType: *registration.OwnerType, OwnerID: *registration.OwnerId}
		}
		// Fall back to Event's ownership
		var event models.Event
		if err := db.Select("owner_type, owner_id").First(&event, registration.EventId).Error; err == nil {
			if event.OwnerType != nil && event.OwnerId != nil && *event.OwnerId > 0 {
				return &OwnerInfo{OwnerType: *event.OwnerType, OwnerID: *event.OwnerId}
			}
		}

	case "EventPayment":
		// EventPayment uses Event ownership (registration is created after payment)
		// payeeID in this case might be a temporary ID or 0, so we can't look it up yet
		// The ownership will be derived from the Event when creating the payment
		return nil

	case "Event":
		// Direct event lookup
		var event models.Event
		if err := db.Select("owner_type, owner_id").First(&event, payeeID).Error; err != nil {
			return nil
		}
		if event.OwnerType != nil && event.OwnerId != nil && *event.OwnerId > 0 {
			return &OwnerInfo{OwnerType: *event.OwnerType, OwnerID: *event.OwnerId}
		}
	}

	return nil
}

// GetOwnerFromFinance looks up ownership from a finance entity (for FinanceTransaction)
// Finance can be: SchoolBill, EventRegistration, Event, Bill
// Returns nil if ownership cannot be determined
func GetOwnerFromFinance(db *gorm.DB, financeType string, financeID int64) *OwnerInfo {
	if financeID == 0 {
		return nil
	}

	switch financeType {
	case "SchoolBill", "SchoolBillPayment":
		// Get ownership from SchoolBill or its Bill
		var schoolBill models.SchoolBill
		if err := db.Select("bill_id, owner_type, owner_id").First(&schoolBill, financeID).Error; err != nil {
			return nil
		}
		// First try to get ownership from SchoolBill itself
		if schoolBill.OwnerType != nil && schoolBill.OwnerId != nil && *schoolBill.OwnerId > 0 {
			return &OwnerInfo{OwnerType: *schoolBill.OwnerType, OwnerID: *schoolBill.OwnerId}
		}
		// Fall back to Bill's ownership
		if schoolBill.BillId != nil {
			var bill models.Bill
			if err := db.Select("owner_type, owner_id").First(&bill, *schoolBill.BillId).Error; err == nil {
				if bill.OwnerType != nil && bill.OwnerId != nil && *bill.OwnerId > 0 {
					return &OwnerInfo{OwnerType: *bill.OwnerType, OwnerID: *bill.OwnerId}
				}
			}
		}

	case "EventRegistration":
		// Get ownership from EventRegistration or Event
		var registration models.EventRegistration
		if err := db.Select("event_id, owner_type, owner_id").First(&registration, financeID).Error; err != nil {
			return nil
		}
		// First try to get ownership from EventRegistration itself
		if registration.OwnerType != nil && registration.OwnerId != nil && *registration.OwnerId > 0 {
			return &OwnerInfo{OwnerType: *registration.OwnerType, OwnerID: *registration.OwnerId}
		}
		// Fall back to Event's ownership
		var event models.Event
		if err := db.Select("owner_type, owner_id").First(&event, registration.EventId).Error; err == nil {
			if event.OwnerType != nil && event.OwnerId != nil && *event.OwnerId > 0 {
				return &OwnerInfo{OwnerType: *event.OwnerType, OwnerID: *event.OwnerId}
			}
		}

	case "Event":
		// Direct event lookup
		var event models.Event
		if err := db.Select("owner_type, owner_id").First(&event, financeID).Error; err != nil {
			return nil
		}
		if event.OwnerType != nil && event.OwnerId != nil && *event.OwnerId > 0 {
			return &OwnerInfo{OwnerType: *event.OwnerType, OwnerID: *event.OwnerId}
		}

	case "Bill":
		// Direct bill lookup
		var bill models.Bill
		if err := db.Select("owner_type, owner_id").First(&bill, financeID).Error; err != nil {
			return nil
		}
		if bill.OwnerType != nil && bill.OwnerId != nil && *bill.OwnerId > 0 {
			return &OwnerInfo{OwnerType: *bill.OwnerType, OwnerID: *bill.OwnerId}
		}
	}

	return nil
}

// GetOwnerFromEvent looks up ownership from an Event by code
// Returns nil if ownership cannot be determined
func GetOwnerFromEvent(db *gorm.DB, eventCode string) *OwnerInfo {
	if eventCode == "" {
		return nil
	}

	var event models.Event
	if err := db.Select("owner_type, owner_id").Where("registration_code = ?", eventCode).First(&event).Error; err != nil {
		return nil
	}

	if event.OwnerType != nil && event.OwnerId != nil && *event.OwnerId > 0 {
		return &OwnerInfo{OwnerType: *event.OwnerType, OwnerID: *event.OwnerId}
	}

	return nil
}

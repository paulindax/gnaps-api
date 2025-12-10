package repositories

import (
	"gnaps-api/models"
	"gnaps-api/utils"

	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) FindByID(id uint) (*models.SchoolGroup, error) {
	var group models.SchoolGroup
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) List(filters map[string]interface{}, page, limit int) ([]models.SchoolGroup, int64, error) {
	var groups []models.SchoolGroup
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Handle search filter (searches name and description)
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ? OR (description IS NOT NULL AND description LIKE ?)", searchPattern, searchPattern)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		if key == "name" {
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Model(&models.SchoolGroup{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&groups).Error

	return groups, total, err
}

func (r *GroupRepository) Create(group *models.SchoolGroup) error {
	return r.db.Create(group).Error
}

func (r *GroupRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.SchoolGroup{}).Where("id = ?", id).Updates(updates).Error
}

func (r *GroupRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.SchoolGroup{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *GroupRepository) VerifyZoneExists(zoneID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Zone{}).Where("id = ? AND is_deleted = ?", zoneID, false).Count(&count).Error
	return count > 0, err
}

// ============================================
// Owner-based methods for data filtering
// ============================================

// CreateWithOwner creates a new group with owner fields automatically set
func (r *GroupRepository) CreateWithOwner(group *models.SchoolGroup, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	if ownerCtx != nil && ownerCtx.IsValid() {
		ownerType, ownerID := ownerCtx.GetOwnerValues()
		group.SetOwner(ownerType, ownerID)
	}
	return r.db.Create(group).Error
}

// FindByIDWithOwner retrieves a group by ID with owner filtering
func (r *GroupRepository) FindByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.SchoolGroup, error) {
	var group models.SchoolGroup
	query := r.db.Where("id = ? AND is_deleted = ?", id, false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	err := query.First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListWithOwner retrieves groups with filters, pagination, and owner filtering
func (r *GroupRepository) ListWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.SchoolGroup, int64, error) {
	var groups []models.SchoolGroup
	var total int64

	query := r.db.Model(&models.SchoolGroup{}).Where("is_deleted = ?", false)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	// Handle search filter
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ? OR (description IS NOT NULL AND description LIKE ?)", searchPattern, searchPattern)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		if key == "name" {
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&groups).Error

	return groups, total, err
}

// UpdateWithOwner updates a group with owner verification
func (r *GroupRepository) UpdateWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.SchoolGroup{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	result := query.Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// DeleteWithOwner soft deletes a group with owner verification
func (r *GroupRepository) DeleteWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	if err := CanWrite(ownerCtx); err != nil {
		return err
	}

	query := r.db.Model(&models.SchoolGroup{}).Where("id = ?", id)
	query = ApplyOwnerFilterToQuery(query, ownerCtx)

	trueVal := true
	result := query.Update("is_deleted", &trueVal)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type ExecutiveRepository struct {
	db *gorm.DB
}

func NewExecutiveRepository(db *gorm.DB) *ExecutiveRepository {
	return &ExecutiveRepository{db: db}
}

func (r *ExecutiveRepository) FindByID(id uint) (*models.Executive, error) {
	var executive models.Executive
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&executive).Error
	if err != nil {
		return nil, err
	}
	// Populate computed fields
	r.populateComputedFields(&executive)
	return &executive, nil
}

func (r *ExecutiveRepository) List(filters map[string]interface{}, page, limit int) ([]models.Executive, int64, error) {
	var executives []models.Executive
	var total int64

	query := r.db.Where("executives.is_deleted = ?", false)

	// Handle search filter (searches name and email)
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where(
			"executives.first_name LIKE ? OR executives.last_name LIKE ? OR executives.executive_no LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		switch key {
		case "first_name", "last_name", "executive_no":
			query = query.Where("executives."+key+" LIKE ?", "%"+value.(string)+"%")
		case "name":
			query = query.Where("executives.first_name LIKE ? OR executives.last_name LIKE ?", "%"+value.(string)+"%", "%"+value.(string)+"%")
		case "role", "status", "region_id", "zone_id", "position_id":
			query = query.Where("executives."+key+" = ?", value)
		default:
			query = query.Where("executives."+key+" = ?", value)
		}
	}

	query.Model(&models.Executive{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("executives.created_at DESC").Find(&executives).Error
	if err != nil {
		return nil, 0, err
	}

	// Populate computed fields for each executive
	for i := range executives {
		r.populateComputedFields(&executives[i])
	}

	return executives, total, nil
}

// populateComputedFields populates position_name, region_name, zone_name
func (r *ExecutiveRepository) populateComputedFields(executive *models.Executive) {
	// Get position name
	if executive.PositionId != nil {
		var position models.Position
		if err := r.db.Where("id = ?", *executive.PositionId).First(&position).Error; err == nil {
			executive.PositionName = position.Name
		}
	}

	// Get region name
	if executive.RegionId != nil {
		var region models.Region
		if err := r.db.Where("id = ?", *executive.RegionId).First(&region).Error; err == nil {
			executive.RegionName = region.Name
		}
	}

	// Get zone name
	if executive.ZoneId != nil {
		var zone models.Zone
		if err := r.db.Where("id = ?", *executive.ZoneId).First(&zone).Error; err == nil {
			executive.ZoneName = zone.Name
		}
	}
}

func (r *ExecutiveRepository) Create(executive *models.Executive) error {
	return r.db.Create(executive).Error
}

func (r *ExecutiveRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Executive{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ExecutiveRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.Executive{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *ExecutiveRepository) ExecutiveNoExists(executiveNo string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Executive{}).Where("executive_no = ? AND is_deleted = ?", executiveNo, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *ExecutiveRepository) EmailExists(email string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Executive{}).Where("email = ? AND is_deleted = ?", email, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

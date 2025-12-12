package repositories

import (
	"fmt"
	"gnaps-api/models"
	"strings"

	"gorm.io/gorm"
)

type SchoolRepository struct {
	db *gorm.DB
}

func NewSchoolRepository(db *gorm.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) Search(keyword string, limit int) ([]models.School, error) {
	var schools []models.School
	query := r.db.Model(&models.School{}).Where("is_deleted = ?", false)

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	err := query.Limit(limit).Order("name ASC").Find(&schools).Error
	return schools, err
}

// SearchWithBill searches for schools that have a school_bill for a specific bill
func (r *SchoolRepository) SearchWithBill(keyword string, billID int64, limit int) ([]models.School, error) {
	var schools []models.School

	// Join with school_bills to filter schools that have the specific bill
	query := r.db.Model(&models.School{}).
		Joins("INNER JOIN school_bills ON school_bills.school_id = schools.id").
		Where("schools.is_deleted = ?", false).
		Where("school_bills.bill_id = ?", billID)

	if keyword != "" {
		query = query.Where("schools.name LIKE ?", "%"+keyword+"%")
	}

	err := query.Limit(limit).Order("schools.name ASC").Find(&schools).Error
	return schools, err
}

func (r *SchoolRepository) FindByID(id uint) (*models.School, error) {
	var school models.School
	err := r.db.Preload("ContactPersons").Preload("Zone").Where("id = ? AND is_deleted = ?", id, false).First(&school).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

func (r *SchoolRepository) List(filters map[string]interface{}, page, limit int) ([]models.School, int64, error) {
	var schools []models.School
	var total int64

	query := r.db.Model(&models.School{}).Where("is_deleted = ?", false)

	// Apply filters
	for key, value := range filters {
		// Handle LIKE queries for name and member_no
		if key == "name" || key == "member_no" {
			query = query.Where(key+" LIKE ?", "%"+value.(string)+"%")
		} else if key == "school_group_id" {
			// Handle JSON array contains for school_group_ids
			// MySQL JSON_CONTAINS syntax
			query = query.Where("JSON_CONTAINS(school_group_ids, ?, '$')", fmt.Sprintf("%v", value))
		} else if key == "region_id" {
			// Handle region_id by filtering through zones
			query = query.Where("zone_id IN (SELECT id FROM zones WHERE region_id = ? AND is_deleted = ?)", value, false)
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("ContactPersons").Preload("Zone").Offset(offset).Limit(limit).Order("created_at DESC").Find(&schools).Error

	return schools, total, err
}

func (r *SchoolRepository) Create(school *models.School) error {
	return r.db.Create(school).Error
}

func (r *SchoolRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.School{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SchoolRepository) Delete(id uint) error {
	return r.db.Model(&models.School{}).Where("id = ?", id).Update("is_deleted", true).Error
}

// MemberNoExists checks if a member number already exists
func (r *SchoolRepository) MemberNoExists(memberNo string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.School{}).Where("member_no = ? AND is_deleted = ?", memberNo, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// EmailExists checks if an email already exists
func (r *SchoolRepository) EmailExists(email string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.School{}).Where("email = ? AND is_deleted = ?", email, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// VerifyZoneExists checks if a zone exists
func (r *SchoolRepository) VerifyZoneExists(zoneID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Zone{}).Where("id = ? AND is_deleted = ?", zoneID, false).Count(&count).Error
	return count > 0, err
}

// GetNextMemberNoForZone returns the next member number for a zone
// Member number format: ZONE_CODE-XXX where XXX is an incrementing number
func (r *SchoolRepository) GetNextMemberNoForZone(zoneID int64) (string, error) {
	// Get the zone to get its code/name
	var zone models.Zone
	if err := r.db.Where("id = ? AND is_deleted = ?", zoneID, false).First(&zone).Error; err != nil {
		return "", err
	}

	// Count existing schools in this zone
	var count int64
	if err := r.db.Model(&models.School{}).Where("zone_id = ? AND is_deleted = ?", zoneID, false).Count(&count).Error; err != nil {
		return "", err
	}

	// Generate next member number: ZONE_CODE-XXX (e.g., ACC-001, KUM-002)
	// Use first 3 letters of zone name as prefix
	prefix := ""
	if zone.Name != nil {
		prefix = *zone.Name
	}
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}

	nextNumber := count + 1
	memberNo := fmt.Sprintf("%s-%03d", strings.ToUpper(prefix), nextNumber)

	return memberNo, nil
}

// ListWithRoleFilter returns schools filtered by role-based access
// - system_admin/national_admin: all schools
// - region_admin: schools in zones within their region
// - zone_admin: schools in their zone
func (r *SchoolRepository) ListWithRoleFilter(filters map[string]interface{}, page, limit int, regionID, zoneID *int64) ([]models.School, int64, error) {
	var schools []models.School
	var total int64

	query := r.db.Model(&models.School{}).Where("is_deleted = ?", false)

	// Apply role-based filtering
	if zoneID != nil {
		// zone_admin can only see schools in their zone
		query = query.Where("zone_id = ?", *zoneID)
	} else if regionID != nil {
		// region_admin can see schools in zones within their region
		query = query.Where("zone_id IN (SELECT id FROM zones WHERE region_id = ? AND is_deleted = ?)", *regionID, false)
	}
	// system_admin and national_admin see all schools (no additional filter)

	// Apply filters
	for key, value := range filters {
		// Handle LIKE queries for name and member_no
		if key == "name" || key == "member_no" {
			query = query.Where(key+" LIKE ?", "%"+value.(string)+"%")
		} else if key == "school_group_id" {
			// Handle JSON array contains for school_group_ids
			query = query.Where("JSON_CONTAINS(school_group_ids, ?, '$')", fmt.Sprintf("%v", value))
		} else if key == "region_id" {
			// Handle region_id by filtering through zones
			query = query.Where("zone_id IN (SELECT id FROM zones WHERE region_id = ? AND is_deleted = ?)", value, false)
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("ContactPersons").Preload("Zone").Offset(offset).Limit(limit).Order("created_at DESC").Find(&schools).Error

	return schools, total, err
}

// FindByIDWithRoleFilter returns a school if accessible by the user's role
func (r *SchoolRepository) FindByIDWithRoleFilter(id uint, regionID, zoneID *int64) (*models.School, error) {
	var school models.School
	query := r.db.Where("id = ? AND is_deleted = ?", id, false)

	// Apply role-based filtering
	if zoneID != nil {
		// zone_admin can only see schools in their zone
		query = query.Where("zone_id = ?", *zoneID)
	} else if regionID != nil {
		// region_admin can see schools in zones within their region
		query = query.Where("zone_id IN (SELECT id FROM zones WHERE region_id = ? AND is_deleted = ?)", *regionID, false)
	}

	err := query.Preload("ContactPersons").Preload("Zone").First(&school).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

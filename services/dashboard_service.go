package services

import (
	"fmt"
	"gnaps-api/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService creates a new instance of DashboardService
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// GetSystemAdminStats returns statistics for system administrators
func (s *DashboardService) GetSystemAdminStats() fiber.Map {
	var (
		totalRegions        int64
		totalZones          int64
		totalSchools        int64
		totalExecutives     int64
		totalNews           int64
		totalUsers          int64
		totalContactPersons int64
		pendingComments     int64
	)

	// Count all entities
	s.db.Model(&models.Region{}).Where("is_deleted = ?", false).Count(&totalRegions)
	s.db.Model(&models.Zone{}).Where("is_deleted = ?", false).Count(&totalZones)
	s.db.Model(&models.School{}).Where("is_deleted = ?", false).Count(&totalSchools)
	s.db.Model(&models.Executive{}).Where("is_deleted = ?", 0).Count(&totalExecutives)
	s.db.Model(&models.New{}).Where("is_deleted = ?", false).Count(&totalNews)
	s.db.Model(&models.User{}).Where("is_deleted = ?", false).Count(&totalUsers)
	s.db.Model(&models.ContactPerson{}).Count(&totalContactPersons)
	s.db.Model(&models.NewsComment{}).Where("is_deleted = ? AND is_approved = ?", false, false).Count(&pendingComments)

	// Get recent schools
	var recentSchools []models.School
	s.db.Where("is_deleted = ?", false).Order("created_at DESC").Limit(5).Find(&recentSchools)

	return fiber.Map{
		"summary": fiber.Map{
			"total_regions":         totalRegions,
			"total_zones":           totalZones,
			"total_schools":         totalSchools,
			"total_executives":      totalExecutives,
			"total_news":            totalNews,
			"total_users":           totalUsers,
			"total_contact_persons": totalContactPersons,
			"pending_comments":      pendingComments,
		},
		"recent_schools": recentSchools,
	}
}

// GetNationalAdminStats returns statistics for national administrators
func (s *DashboardService) GetNationalAdminStats() fiber.Map {
	var (
		totalRegions    int64
		totalZones      int64
		totalSchools    int64
		totalExecutives int64
		totalNews       int64
	)

	// Count all entities (same as system admin but with different perspective)
	s.db.Model(&models.Region{}).Where("is_deleted = ?", false).Count(&totalRegions)
	s.db.Model(&models.Zone{}).Where("is_deleted = ?", false).Count(&totalZones)
	s.db.Model(&models.School{}).Where("is_deleted = ?", false).Count(&totalSchools)
	s.db.Model(&models.Executive{}).Where("is_deleted = ?", 0).Count(&totalExecutives)
	s.db.Model(&models.New{}).Where("is_deleted = ?", false).Count(&totalNews)

	// Get regions with their zone counts
	var regions []models.Region
	s.db.Where("is_deleted = ?", false).Find(&regions)

	regionsWithCounts := []fiber.Map{}
	for _, region := range regions {
		var zoneCount int64
		s.db.Model(&models.Zone{}).Where("region_id = ? AND is_deleted = ?", region.ID, false).Count(&zoneCount)

		var schoolCount int64
		s.db.Table("schools").
			Joins("JOIN zones ON schools.zone_id = zones.id").
			Where("zones.region_id = ? AND schools.is_deleted = ? AND zones.is_deleted = ?", region.ID, false, false).
			Count(&schoolCount)

		regionsWithCounts = append(regionsWithCounts, fiber.Map{
			"id":            region.ID,
			"name":          region.Name,
			"code":          region.Code,
			"total_zones":   zoneCount,
			"total_schools": schoolCount,
		})
	}

	return fiber.Map{
		"summary": fiber.Map{
			"total_regions":    totalRegions,
			"total_zones":      totalZones,
			"total_schools":    totalSchools,
			"total_executives": totalExecutives,
			"total_news":       totalNews,
		},
		"regions": regionsWithCounts,
	}
}

// GetRegionalAdminStats returns statistics for regional administrators
func (s *DashboardService) GetRegionalAdminStats(regionID string) fiber.Map {
	var (
		totalZones      int64
		totalSchools    int64
		totalExecutives int64
	)

	// Get region info
	var region models.Region
	s.db.Where("id = ? AND is_deleted = ?", regionID, false).First(&region)

	// Count zones in this region
	s.db.Model(&models.Zone{}).Where("region_id = ? AND is_deleted = ?", regionID, false).Count(&totalZones)

	// Count schools in zones of this region
	s.db.Table("schools").
		Joins("JOIN zones ON schools.zone_id = zones.id").
		Where("zones.region_id = ? AND schools.is_deleted = ? AND zones.is_deleted = ?", regionID, false, false).
		Count(&totalSchools)

	// Count executives assigned to this region
	s.db.Model(&models.Executive{}).
		Where("region_id = ? AND is_deleted = ?", regionID, false).
		Count(&totalExecutives)

	// Get zones with their school counts
	var zones []models.Zone
	s.db.Where("region_id = ? AND is_deleted = ?", regionID, false).Find(&zones)

	zonesWithCounts := []fiber.Map{}
	for _, zone := range zones {
		var schoolCount int64
		s.db.Model(&models.School{}).Where("zone_id = ? AND is_deleted = ?", zone.ID, false).Count(&schoolCount)

		zonesWithCounts = append(zonesWithCounts, fiber.Map{
			"id":            zone.ID,
			"name":          zone.Name,
			"code":          zone.Code,
			"total_schools": schoolCount,
		})
	}

	return fiber.Map{
		"region": fiber.Map{
			"id":   region.ID,
			"name": region.Name,
			"code": region.Code,
		},
		"summary": fiber.Map{
			"total_zones":      totalZones,
			"total_schools":    totalSchools,
			"total_executives": totalExecutives,
		},
		"zones": zonesWithCounts,
	}
}

// GetZoneAdminStats returns statistics for zone administrators
func (s *DashboardService) GetZoneAdminStats(zoneID string) fiber.Map {
	var (
		totalSchools        int64
		totalContactPersons int64
		totalExecutives     int64
	)

	// Get zone info with region
	var zone models.Zone
	s.db.Where("id = ? AND is_deleted = ?", zoneID, false).First(&zone)

	var region models.Region
	if zone.RegionId != nil {
		s.db.Where("id = ? AND is_deleted = ?", zone.RegionId, false).First(&region)
	}

	// Count schools in this zone
	s.db.Model(&models.School{}).Where("zone_id = ? AND is_deleted = ?", zoneID, false).Count(&totalSchools)

	// Count contact persons in schools of this zone
	s.db.Table("contact_persons").
		Joins("JOIN schools ON contact_persons.school_id = schools.id").
		Where("schools.zone_id = ? AND schools.is_deleted = ?", zoneID, false).
		Count(&totalContactPersons)

	// Count executives assigned to this zone
	s.db.Model(&models.Executive{}).
		Where("zone_id = ? AND is_deleted = ?", zoneID, false).
		Count(&totalExecutives)

	// Get schools in this zone
	var schools []models.School
	s.db.Where("zone_id = ? AND is_deleted = ?", zoneID, false).Order("created_at DESC").Find(&schools)

	return fiber.Map{
		"zone": fiber.Map{
			"id":   zone.ID,
			"name": zone.Name,
			"code": zone.Code,
		},
		"region": fiber.Map{
			"id":   region.ID,
			"name": region.Name,
			"code": region.Code,
		},
		"summary": fiber.Map{
			"total_schools":         totalSchools,
			"total_contact_persons": totalContactPersons,
			"total_executives":      totalExecutives,
		},
		"schools": schools,
	}
}

// GetSchoolUserStats returns statistics for school users
func (s *DashboardService) GetSchoolUserStats(schoolID string) fiber.Map {
	var (
		totalContactPersons int64
		totalComments       int64
	)

	// Get school info
	var school models.School
	s.db.Where("id = ? AND is_deleted = ?", schoolID, false).First(&school)

	// Get zone and region info
	var zone models.Zone
	var region models.Region
	if school.ZoneId != nil {
		s.db.Where("id = ? AND is_deleted = ?", school.ZoneId, false).First(&zone)
		if zone.RegionId != nil {
			s.db.Where("id = ? AND is_deleted = ?", zone.RegionId, false).First(&region)
		}
	}

	// Count contact persons for this school
	s.db.Model(&models.ContactPerson{}).Where("school_id = ?", schoolID).Count(&totalContactPersons)

	// Get contact persons
	var contactPersons []models.ContactPerson
	s.db.Where("school_id = ?", schoolID).Order("created_at DESC").Find(&contactPersons)

	// Count comments by users from this school (if user_id is linked to school)
	s.db.Model(&models.NewsComment{}).
		Where("user_id IN (SELECT id FROM users WHERE id IN (SELECT user_id FROM schools WHERE id = ?))", schoolID).
		Count(&totalComments)

	return fiber.Map{
		"school": fiber.Map{
			"id":                    school.ID,
			"name":                  school.Name,
			"member_no":             school.MemberNo,
			"email":                 school.Email,
			"mobile_no":             school.MobileNo,
			"address":               school.Address,
			"location":              school.Location,
			"gps_address":           school.GpsAddress,
			"date_of_establishment": school.DateOfEstablishment,
			"joining_date":          school.JoiningDate,
		},
		"zone": fiber.Map{
			"id":   zone.ID,
			"name": zone.Name,
			"code": zone.Code,
		},
		"region": fiber.Map{
			"id":   region.ID,
			"name": region.Name,
			"code": region.Code,
		},
		"summary": fiber.Map{
			"total_contact_persons": totalContactPersons,
			"total_comments":        totalComments,
		},
		"contact_persons": contactPersons,
	}
}

// GetRegionalAdminStatsFromExecutive extracts region ID from executive and returns stats
func (s *DashboardService) GetRegionalAdminStatsFromExecutive(executive *models.Executive) (fiber.Map, error) {
	if executive.RegionId == nil || *executive.RegionId == 0 {
		return nil, fmt.Errorf("no region assigned to this executive")
	}

	regionID := fmt.Sprintf("%d", *executive.RegionId)
	return s.GetRegionalAdminStats(regionID), nil
}

// GetZoneAdminStatsFromExecutive extracts zone ID from executive and returns stats
func (s *DashboardService) GetZoneAdminStatsFromExecutive(executive *models.Executive) (fiber.Map, error) {
	if executive.ZoneId == nil || *executive.ZoneId == 0 {
		return nil, fmt.Errorf("no zone assigned to this executive")
	}

	zoneID := fmt.Sprintf("%d", *executive.ZoneId)
	return s.GetZoneAdminStats(zoneID), nil
}

// GetOverview returns recent activities and news based on role and user ID
func (s *DashboardService) GetOverview(userID uint, role string) fiber.Map {
	// Get recent news
	var recentNews []models.New
	newsQuery := s.db.Where("is_deleted = ?", false).Order("created_at DESC").Limit(5)

	// Filter news based on role
	switch role {
	case "region_admin":
		// Get executive's assigned region
		var executive models.Executive
		if err := s.db.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err == nil {
			if executive.RegionId != nil && *executive.RegionId > 0 {
				regionID := fmt.Sprintf("%d", *executive.RegionId)
				newsQuery = newsQuery.Where("JSON_CONTAINS(region_ids, ?)", fmt.Sprintf("[%s]", regionID))
			}
		}
	case "zone_admin":
		// Get executive's assigned zone
		var executive models.Executive
		if err := s.db.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err == nil {
			if executive.ZoneId != nil && *executive.ZoneId > 0 {
				zoneID := fmt.Sprintf("%d", *executive.ZoneId)
				newsQuery = newsQuery.Where("JSON_CONTAINS(zone_ids, ?)", fmt.Sprintf("[%s]", zoneID))
			}
		}
	}

	newsQuery.Find(&recentNews)

	// Get recent comments (approved only)
	var recentComments []models.NewsComment
	s.db.Where("is_deleted = ? AND is_approved = ?", false, true).
		Order("created_at DESC").
		Limit(5).
		Find(&recentComments)

	return fiber.Map{
		"recent_news":     recentNews,
		"recent_comments": recentComments,
	}
}

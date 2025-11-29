package controllers

import (
	"fmt"
	"gnaps-api/models"

	"github.com/gofiber/fiber/v2"
)

type DashboardController struct {
}

func init() {
	RegisterController("dashboard", &DashboardController{})
}

func (d *DashboardController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "stats":
		return d.stats(c)
	case "overview":
		return d.overview(c)
	default:
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("unknown action %s", action),
		})
	}
}

// stats returns role-based dashboard statistics
func (d *DashboardController) stats(c *fiber.Ctx) error {
	// Get user info from JWT context (set by auth middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	role, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User role not found",
		})
	}

	// Fetch full user details from database
	var user models.User
	if err := DB.Where("id = ? AND is_deleted = ?", userID, false).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var stats fiber.Map
	var err error

	switch role {
	case "system_admin":
		stats = d.getSystemAdminStats()
	case "national_admin":
		stats = d.getNationalAdminStats()
	case "regional_admin":
		// Find executive record for this user to get assigned regions
		var executive models.Executive
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Executive profile not found for this user",
			})
		}
		// Get the first assigned region (or you could iterate through all)
		// For simplicity, we'll use the first region if multiple are assigned
		stats, err = d.getRegionalAdminStatsFromExecutive(&executive)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	case "zone_admin":
		// Find executive record for this user to get assigned zones
		var executive models.Executive
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Executive profile not found for this user",
			})
		}
		stats, err = d.getZoneAdminStatsFromExecutive(&executive)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	case "school_user":
		// Find school for this user
		var school models.School
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, false).First(&school).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "School not found for this user",
			})
		}
		stats = d.getSchoolUserStats(fmt.Sprintf("%d", school.ID))
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid role: %s", role),
		})
	}

	// Add user info to response
	stats["user"] = fiber.Map{
		"id":         user.ID,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"role":       user.Role,
	}

	return c.JSON(fiber.Map{
		"role": role,
		"data": stats,
	})
}

// getRegionalAdminStatsFromExecutive extracts region ID from executive and returns stats
func (d *DashboardController) getRegionalAdminStatsFromExecutive(executive *models.Executive) (fiber.Map, error) {
	if executive.AssignedRegionsIds == nil {
		return nil, fmt.Errorf("no regions assigned to this executive")
	}

	// Parse JSON array to get region IDs
	var regionIDs []int64
	if err := executive.AssignedRegionsIds.Scan(&regionIDs); err != nil {
		return nil, fmt.Errorf("failed to parse assigned regions: %w", err)
	}

	if len(regionIDs) == 0 {
		return nil, fmt.Errorf("no regions assigned to this executive")
	}

	// Use the first assigned region
	// In production, you might want to handle multiple regions differently
	regionID := fmt.Sprintf("%d", regionIDs[0])
	return d.getRegionalAdminStats(regionID), nil
}

// getZoneAdminStatsFromExecutive extracts zone ID from executive and returns stats
func (d *DashboardController) getZoneAdminStatsFromExecutive(executive *models.Executive) (fiber.Map, error) {
	if executive.AssignedZoneIds == nil {
		return nil, fmt.Errorf("no zones assigned to this executive")
	}

	// Parse JSON array to get zone IDs
	var zoneIDs []int64
	if err := executive.AssignedZoneIds.Scan(&zoneIDs); err != nil {
		return nil, fmt.Errorf("failed to parse assigned zones: %w", err)
	}

	if len(zoneIDs) == 0 {
		return nil, fmt.Errorf("no zones assigned to this executive")
	}

	// Use the first assigned zone
	// In production, you might want to handle multiple zones differently
	zoneID := fmt.Sprintf("%d", zoneIDs[0])
	return d.getZoneAdminStats(zoneID), nil
}

// overview returns recent activities and news based on role
func (d *DashboardController) overview(c *fiber.Ctx) error {
	// Get user info from JWT context (set by auth middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	role, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User role not found",
		})
	}

	// Get recent news
	var recentNews []models.New
	newsQuery := DB.Where("is_deleted = ?", false).Order("created_at DESC").Limit(5)

	// Filter news based on role
	switch role {
	case "regional_admin":
		// Get executive's assigned regions
		var executive models.Executive
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err == nil {
			if executive.AssignedRegionsIds != nil {
				var regionIDs []int64
				if err := executive.AssignedRegionsIds.Scan(&regionIDs); err == nil && len(regionIDs) > 0 {
					regionID := fmt.Sprintf("%d", regionIDs[0])
					newsQuery = newsQuery.Where("JSON_CONTAINS(region_ids, ?)", fmt.Sprintf("[%s]", regionID))
				}
			}
		}
	case "zone_admin":
		// Get executive's assigned zones
		var executive models.Executive
		if err := DB.Where("user_id = ? AND is_deleted = ?", userID, 0).First(&executive).Error; err == nil {
			if executive.AssignedZoneIds != nil {
				var zoneIDs []int64
				if err := executive.AssignedZoneIds.Scan(&zoneIDs); err == nil && len(zoneIDs) > 0 {
					zoneID := fmt.Sprintf("%d", zoneIDs[0])
					newsQuery = newsQuery.Where("JSON_CONTAINS(zone_ids, ?)", fmt.Sprintf("[%s]", zoneID))
				}
			}
		}
	}

	newsQuery.Find(&recentNews)

	// Get recent comments (approved only)
	var recentComments []models.NewsComment
	DB.Where("is_deleted = ? AND is_approved = ?", false, true).
		Order("created_at DESC").
		Limit(5).
		Find(&recentComments)

	return c.JSON(fiber.Map{
		"recent_news":     recentNews,
		"recent_comments": recentComments,
	})
}

// ==================== ROLE-SPECIFIC STATS FUNCTIONS ====================

// getSystemAdminStats returns statistics for system administrators
func (d *DashboardController) getSystemAdminStats() fiber.Map {
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
	DB.Model(&models.Region{}).Where("is_deleted = ?", false).Count(&totalRegions)
	DB.Model(&models.Zone{}).Where("is_deleted = ?", false).Count(&totalZones)
	DB.Model(&models.School{}).Where("is_deleted = ?", false).Count(&totalSchools)
	DB.Model(&models.Executive{}).Where("is_deleted = ?", 0).Count(&totalExecutives)
	DB.Model(&models.New{}).Where("is_deleted = ?", false).Count(&totalNews)
	DB.Model(&models.User{}).Where("is_deleted = ?", false).Count(&totalUsers)
	DB.Model(&models.ContactPerson{}).Count(&totalContactPersons)
	DB.Model(&models.NewsComment{}).Where("is_deleted = ? AND is_approved = ?", false, false).Count(&pendingComments)

	// Get recent schools
	var recentSchools []models.School
	DB.Where("is_deleted = ?", false).Order("created_at DESC").Limit(5).Find(&recentSchools)

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

// getNationalAdminStats returns statistics for national administrators
func (d *DashboardController) getNationalAdminStats() fiber.Map {
	var (
		totalRegions    int64
		totalZones      int64
		totalSchools    int64
		totalExecutives int64
		totalNews       int64
	)

	// Count all entities (same as system admin but with different perspective)
	DB.Model(&models.Region{}).Where("is_deleted = ?", false).Count(&totalRegions)
	DB.Model(&models.Zone{}).Where("is_deleted = ?", false).Count(&totalZones)
	DB.Model(&models.School{}).Where("is_deleted = ?", false).Count(&totalSchools)
	DB.Model(&models.Executive{}).Where("is_deleted = ?", 0).Count(&totalExecutives)
	DB.Model(&models.New{}).Where("is_deleted = ?", false).Count(&totalNews)

	// Get regions with their zone counts
	var regions []models.Region
	DB.Where("is_deleted = ?", false).Find(&regions)

	regionsWithCounts := []fiber.Map{}
	for _, region := range regions {
		var zoneCount int64
		DB.Model(&models.Zone{}).Where("region_id = ? AND is_deleted = ?", region.ID, false).Count(&zoneCount)

		var schoolCount int64
		DB.Table("schools").
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

// getRegionalAdminStats returns statistics for regional administrators
func (d *DashboardController) getRegionalAdminStats(regionID string) fiber.Map {
	var (
		totalZones      int64
		totalSchools    int64
		totalExecutives int64
	)

	// Get region info
	var region models.Region
	DB.Where("id = ? AND is_deleted = ?", regionID, false).First(&region)

	// Count zones in this region
	DB.Model(&models.Zone{}).Where("region_id = ? AND is_deleted = ?", regionID, false).Count(&totalZones)

	// Count schools in zones of this region
	DB.Table("schools").
		Joins("JOIN zones ON schools.zone_id = zones.id").
		Where("zones.region_id = ? AND schools.is_deleted = ? AND zones.is_deleted = ?", regionID, false, false).
		Count(&totalSchools)

	// Count executives assigned to this region
	DB.Model(&models.Executive{}).
		Where("JSON_CONTAINS(assigned_regions_ids, ?) AND is_deleted = ?", fmt.Sprintf("[%s]", regionID), 0).
		Count(&totalExecutives)

	// Get zones with their school counts
	var zones []models.Zone
	DB.Where("region_id = ? AND is_deleted = ?", regionID, false).Find(&zones)

	zonesWithCounts := []fiber.Map{}
	for _, zone := range zones {
		var schoolCount int64
		DB.Model(&models.School{}).Where("zone_id = ? AND is_deleted = ?", zone.ID, false).Count(&schoolCount)

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

// getZoneAdminStats returns statistics for zone administrators
func (d *DashboardController) getZoneAdminStats(zoneID string) fiber.Map {
	var (
		totalSchools        int64
		totalContactPersons int64
		totalExecutives     int64
	)

	// Get zone info with region
	var zone models.Zone
	DB.Where("id = ? AND is_deleted = ?", zoneID, false).First(&zone)

	var region models.Region
	if zone.RegionId != nil {
		DB.Where("id = ? AND is_deleted = ?", zone.RegionId, false).First(&region)
	}

	// Count schools in this zone
	DB.Model(&models.School{}).Where("zone_id = ? AND is_deleted = ?", zoneID, false).Count(&totalSchools)

	// Count contact persons in schools of this zone
	DB.Table("contact_persons").
		Joins("JOIN schools ON contact_persons.school_id = schools.id").
		Where("schools.zone_id = ? AND schools.is_deleted = ?", zoneID, false).
		Count(&totalContactPersons)

	// Count executives assigned to this zone
	DB.Model(&models.Executive{}).
		Where("JSON_CONTAINS(assigned_zone_ids, ?) AND is_deleted = ?", fmt.Sprintf("[%s]", zoneID), 0).
		Count(&totalExecutives)

	// Get schools in this zone
	var schools []models.School
	DB.Where("zone_id = ? AND is_deleted = ?", zoneID, false).Order("created_at DESC").Find(&schools)

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

// getSchoolUserStats returns statistics for school users
func (d *DashboardController) getSchoolUserStats(schoolID string) fiber.Map {
	var (
		totalContactPersons int64
		totalComments       int64
	)

	// Get school info
	var school models.School
	DB.Where("id = ? AND is_deleted = ?", schoolID, false).First(&school)

	// Get zone and region info
	var zone models.Zone
	var region models.Region
	if school.ZoneId != nil {
		DB.Where("id = ? AND is_deleted = ?", school.ZoneId, false).First(&zone)
		if zone.RegionId != nil {
			DB.Where("id = ? AND is_deleted = ?", zone.RegionId, false).First(&region)
		}
	}

	// Count contact persons for this school
	DB.Model(&models.ContactPerson{}).Where("school_id = ?", schoolID).Count(&totalContactPersons)

	// Get contact persons
	var contactPersons []models.ContactPerson
	DB.Where("school_id = ?", schoolID).Order("created_at DESC").Find(&contactPersons)

	// Count comments by users from this school (if user_id is linked to school)
	DB.Model(&models.NewsComment{}).
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

package utils

import (
	"github.com/gofiber/fiber/v2"
)

// OwnerType constants
const (
	OwnerTypeNational = "national"
	OwnerTypeRegion   = "region"
	OwnerTypeZone     = "zone"
)

// Executive role constants
const (
	RoleSystemAdmin   = "system_admin"
	RoleNationalAdmin = "national_admin"
	RoleRegionAdmin   = "region_admin"
	RoleZoneAdmin     = "zone_admin"
	RoleSchoolAdmin   = "school_admin"
)

// Default national owner ID
const DefaultNationalOwnerID int64 = 1

// OwnerContext holds the ownership information for data filtering and creation
type OwnerContext struct {
	OwnerType string `json:"owner_type"`
	OwnerID   int64  `json:"owner_id"`
	Role      string `json:"role"`
	UserID    uint   `json:"user_id"`
}

// IsValid checks if the owner context has valid values
func (oc *OwnerContext) IsValid() bool {
	return oc.OwnerType != "" && oc.OwnerID > 0
}

// IsSystemAdmin checks if the owner is a system admin
func (oc *OwnerContext) IsSystemAdmin() bool {
	return oc.Role == RoleSystemAdmin
}

// IsNationalAdmin checks if the owner is a national admin
func (oc *OwnerContext) IsNationalAdmin() bool {
	return oc.Role == RoleNationalAdmin
}

// IsRegionAdmin checks if the owner is a region admin
func (oc *OwnerContext) IsRegionAdmin() bool {
	return oc.Role == RoleRegionAdmin
}

// IsZoneAdmin checks if the owner is a zone admin
func (oc *OwnerContext) IsZoneAdmin() bool {
	return oc.Role == RoleZoneAdmin
}

// CanWriteOwnerData checks if this owner context can write (create/update/delete) data in owner-based tables
// System admin cannot write owner-based data (view only)
// National, Region, Zone admins can write data with their respective owner context
func (oc *OwnerContext) CanWriteOwnerData() bool {
	// System admin cannot write owner-based data (view only)
	if oc.IsSystemAdmin() {
		return false
	}
	// Other roles can write if they have valid owner context
	return oc.IsValid()
}

// CanCreateOwnerData is an alias for CanWriteOwnerData for backwards compatibility
func (oc *OwnerContext) CanCreateOwnerData() bool {
	return oc.CanWriteOwnerData()
}

// CanAccessOwner checks if this owner context can access data with the given owner type and ID
// System admin and National admin can access all data
// Region admin can access region and zone data within their region
// Zone admin can only access their zone data
func (oc *OwnerContext) CanAccessOwner(targetOwnerType string, targetOwnerID int64) bool {
	// System admin can access everything (view only)
	if oc.IsSystemAdmin() {
		return true
	}

	// National admin can access everything
	if oc.IsNationalAdmin() {
		return true
	}

	// Region admin can access their region data
	if oc.IsRegionAdmin() {
		if targetOwnerType == OwnerTypeRegion && targetOwnerID == oc.OwnerID {
			return true
		}
		// Region admin can also access zone data within their region (would need zone-region mapping)
		// For now, region admin only accesses region-owned data
		return false
	}

	// Zone admin can only access their zone data
	if oc.IsZoneAdmin() {
		return targetOwnerType == OwnerTypeZone && targetOwnerID == oc.OwnerID
	}

	return false
}

// GetOwnerContextFromExecutive creates an OwnerContext from executive data
func GetOwnerContextFromExecutive(role string, regionID, zoneID *int64, userID uint) *OwnerContext {
	ctx := &OwnerContext{
		Role:   role,
		UserID: userID,
	}

	switch role {
	case RoleNationalAdmin:
		ctx.OwnerType = OwnerTypeNational
		ctx.OwnerID = DefaultNationalOwnerID
	case RoleRegionAdmin:
		ctx.OwnerType = OwnerTypeRegion
		if regionID != nil {
			ctx.OwnerID = *regionID
		}
	case RoleZoneAdmin:
		ctx.OwnerType = OwnerTypeZone
		if zoneID != nil {
			ctx.OwnerID = *zoneID
		}
	}

	return ctx
}

// GetOwnerContext retrieves the OwnerContext from fiber context
func GetOwnerContext(c *fiber.Ctx) *OwnerContext {
	if ctx, ok := c.Locals("owner_context").(*OwnerContext); ok {
		return ctx
	}
	return nil
}

// SetOwnerContext sets the OwnerContext in fiber context
func SetOwnerContext(c *fiber.Ctx, ctx *OwnerContext) {
	c.Locals("owner_context", ctx)
}

// MustGetOwnerContext retrieves the OwnerContext or returns an error response
func MustGetOwnerContext(c *fiber.Ctx) (*OwnerContext, error) {
	ctx := GetOwnerContext(c)
	if ctx == nil || !ctx.IsValid() {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "owner context not found or invalid")
	}
	return ctx, nil
}

// OwnerFilter represents filter parameters for owner-based queries
type OwnerFilter struct {
	OwnerType string
	OwnerID   int64
}

// GetOwnerFilter returns the owner filter based on the context
// For system admin and national admin, returns nil (no filter - access all)
// For region/zone admin, returns their specific filter
func (oc *OwnerContext) GetOwnerFilter() *OwnerFilter {
	// System admin sees all data - no filter (view only)
	if oc.IsSystemAdmin() {
		return nil
	}

	// National admin sees all data - no filter
	if oc.IsNationalAdmin() {
		return nil
	}

	return &OwnerFilter{
		OwnerType: oc.OwnerType,
		OwnerID:   oc.OwnerID,
	}
}

// GetOwnerValues returns owner_type and owner_id for creating new records
func (oc *OwnerContext) GetOwnerValues() (string, int64) {
	return oc.OwnerType, oc.OwnerID
}

// ============================================
// Role-Based Hierarchical Filtering
// For tables without owner_type/owner_id columns
// (zones, regions, schools, executives)
// ============================================

// RoleBasedFilter holds filter info for role-based queries
type RoleBasedFilter struct {
	RegionID *int64 // For region_admin filtering
	ZoneID   *int64 // For zone_admin filtering
	AllData  bool   // True for system_admin and national_admin
}

// GetRoleBasedFilter returns filtering info based on role
// system_admin & national_admin -> AllData = true (see everything)
// region_admin -> RegionID set (filter by region)
// zone_admin -> ZoneID set (filter by zone)
func (oc *OwnerContext) GetRoleBasedFilter() *RoleBasedFilter {
	if oc == nil {
		return &RoleBasedFilter{AllData: true}
	}

	filter := &RoleBasedFilter{}

	switch oc.Role {
	case RoleSystemAdmin, RoleNationalAdmin:
		filter.AllData = true
	case RoleRegionAdmin:
		if oc.OwnerID > 0 {
			filter.RegionID = &oc.OwnerID
		}
	case RoleZoneAdmin:
		if oc.OwnerID > 0 {
			filter.ZoneID = &oc.OwnerID
		}
	default:
		// Unknown role, restrict to nothing
		filter.AllData = false
	}

	return filter
}

// CanViewAllHierarchyData checks if the user can view all hierarchical data
func (oc *OwnerContext) CanViewAllHierarchyData() bool {
	if oc == nil {
		return false
	}
	return oc.IsSystemAdmin() || oc.IsNationalAdmin()
}

// GetRegionIDFilter returns the region ID filter for region_admin
// Returns nil if user can view all data or is zone_admin
func (oc *OwnerContext) GetRegionIDFilter() *int64 {
	if oc == nil || oc.CanViewAllHierarchyData() {
		return nil
	}
	if oc.IsRegionAdmin() && oc.OwnerID > 0 {
		return &oc.OwnerID
	}
	return nil
}

// GetZoneIDFilter returns the zone ID filter for zone_admin
// Returns nil if user can view all data or is region_admin
func (oc *OwnerContext) GetZoneIDFilter() *int64 {
	if oc == nil || oc.CanViewAllHierarchyData() {
		return nil
	}
	if oc.IsZoneAdmin() && oc.OwnerID > 0 {
		return &oc.OwnerID
	}
	return nil
}

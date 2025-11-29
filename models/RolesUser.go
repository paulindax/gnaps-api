package models

// RolesUser model generated from database table 'roles_users'
type RolesUser struct {
	RoleId *int64 `json:"role_id" gorm:"column:role_id"`
	UserId *int64 `json:"user_id" gorm:"column:user_id"`
}

func (RolesUser) TableName() string {
	return "roles_users"
}

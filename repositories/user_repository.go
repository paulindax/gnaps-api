package repositories

import (
	"gnaps-api/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND (is_deleted = ? OR is_deleted IS NULL)", email, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername retrieves a user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ? AND (is_deleted = ? OR is_deleted IS NULL)", username, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// EmailExists checks if an email already exists (optionally excluding a specific user ID)
func (r *UserRepository) EmailExists(email string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.User{}).Where("email = ? AND (is_deleted = ? OR is_deleted IS NULL)", email, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// UsernameExists checks if a username already exists (optionally excluding a specific user ID)
func (r *UserRepository) UsernameExists(username string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.User{}).Where("username = ? AND (is_deleted = ? OR is_deleted IS NULL)", username, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update updates an existing user
func (r *UserRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

// HashPassword hashes a password using bcrypt
func (r *UserRepository) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CreateUserForExecutive creates a user account for an executive
func (r *UserRepository) CreateUserForExecutive(firstName, lastName, email, mobileNo, role, executiveNo string) (*models.User, error) {
	// Generate username from email (part before @)
	username := email
	if atIdx := len(email); atIdx > 0 {
		for i, c := range email {
			if c == '@' {
				username = email[:i]
				break
			}
		}
	}

	// Default password is executive_no + "123"
	defaultPassword := executiveNo + "123"

	// Hash the default password
	hashedPassword, err := r.HashPassword(defaultPassword)
	if err != nil {
		return nil, err
	}

	isFirstLogin := true
	isDeleted := false

	user := &models.User{
		Username:          &username,
		FirstName:         &firstName,
		LastName:          &lastName,
		Email:             &email,
		MobileNo:          &mobileNo,
		Role:              &role,
		EncryptedPassword: &hashedPassword,
		IsFirstLogin:      &isFirstLogin,
		IsDeleted:         &isDeleted,
	}

	if err := r.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateUserForSchool creates a user account for a school
func (r *UserRepository) CreateUserForSchool(schoolName, email, mobileNo, memberNo string) (*models.User, error) {
	// Use member_no as the username for schools
	username := memberNo

	// Default password is member_no + "123"
	defaultPassword := memberNo + "123"

	// Hash the default password
	hashedPassword, err := r.HashPassword(defaultPassword)
	if err != nil {
		return nil, err
	}

	isFirstLogin := true
	isDeleted := false
	role := "school_admin"

	user := &models.User{
		Username:          &username,
		FirstName:         &schoolName, // Use school name as first name
		Email:             &email,
		MobileNo:          &mobileNo,
		Role:              &role,
		EncryptedPassword: &hashedPassword,
		IsFirstLogin:      &isFirstLogin,
		IsDeleted:         &isDeleted,
	}

	if err := r.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

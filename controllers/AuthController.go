package controllers

import (
	"fmt"
	"log"

	"gnaps-api/models"
	"gnaps-api/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct{}

func init() {
	RegisterController("auth", &AuthController{})
}

func (a *AuthController) Handle(action string, c *fiber.Ctx) error {
	switch action {
	case "login":
		return a.login(c)
	case "register":
		return a.register(c)
	case "refresh":
		return a.refresh(c)
	case "me":
		return a.me(c)
	case "logout":
		return a.logout(c)
	default:
		return utils.NotFoundResponse(c, fmt.Sprintf("unknown action %s", action))
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// login handles user authentication and token generation
func (a *AuthController) login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return utils.ValidationErrorResponse(c, "Username and password are required")
	}

	// Find user by username
	var user models.User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return utils.UnauthorizedResponse(c, "Invalid credentials")
	}

	// Check if user is deleted
	if user.IsDeleted != nil && *user.IsDeleted {
		return utils.UnauthorizedResponse(c, "Account has been deleted")
	}

	// Verify password
	if user.EncryptedPassword == nil {
		log.Println("We got to Encrypted")
		return utils.UnauthorizedResponse(c, "Invalid credentials")
	}

	err := bcrypt.CompareHashAndPassword([]byte(*user.EncryptedPassword), []byte(req.Password))
	log.Println("We got to Check Password")
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid credentials")
	}

	// Get user role
	role := "user"
	if user.Role != nil {
		role = *user.Role
	}

	// Get username
	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, req.Username, username, role)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to generate token")
	}

	// Update sign in tracking
	DB.Model(&user).Updates(map[string]interface{}{
		"sign_in_count": user.SignInCount,
	})

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}

// register handles user registration
func (a *AuthController) register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return utils.ValidationErrorResponse(c, "Email and password are required")
	}

	// Check if user already exists
	var existingUser models.User
	if err := DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return utils.ConflictResponse(c, "User with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to hash password")
	}

	hashedPasswordStr := string(hashedPassword)
	role := "user"
	isFirstLogin := true

	// Create user
	user := models.User{
		Username:          &req.Username,
		Email:             &req.Email,
		FirstName:         &req.FirstName,
		LastName:          &req.LastName,
		EncryptedPassword: &hashedPasswordStr,
		Role:              &role,
		IsFirstLogin:      &isFirstLogin,
	}

	if err := DB.Create(&user).Error; err != nil {
		return utils.ServerErrorResponse(c, "Failed to create user")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, req.Email, req.Username, role)
	if err != nil {
		return utils.ServerErrorResponse(c, "Failed to generate token")
	}

	return utils.SuccessResponseWithStatus(c, 201, fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	}, "Registration successful")
}

// refresh generates a new token from an existing valid token
func (a *AuthController) refresh(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.UnauthorizedResponse(c, "Missing authorization header")
	}

	// Extract token
	tokenString := authHeader[7:] // Remove "Bearer " prefix
	newToken, err := utils.RefreshJWT(tokenString)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Failed to refresh token")
	}

	return c.JSON(fiber.Map{
		"token": newToken,
	})
}

// me returns current user info from JWT token
func (a *AuthController) me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return utils.UnauthorizedResponse(c, "User not authenticated")
	}

	// Fetch user from database
	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		return utils.NotFoundResponse(c, "User not found")
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}

// logout handles user logout (client-side token removal)
func (a *AuthController) logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}

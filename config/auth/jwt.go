package auth

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("WARNING: JWT_SECRET is empty!")
	}
	return []byte(secret)
}

func GenerateToken(userID string, role string, orgID *string, schoolID *string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	// if orgID != nil {
	// 	claims["organization_id"] = *orgID
	// }
	// if schoolID != nil {
	// 	claims["school_id"] = *schoolID
	// }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	log.Println("Validating token:", tokenString)
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return getJWTSecret(), nil
	})
}

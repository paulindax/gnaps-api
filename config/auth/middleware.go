package auth

import (
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
)

func JWTProtected() fiber.Handler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	config := jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		SuccessHandler: func(c *fiber.Ctx) error {
			// authHeader := c.Get("Authorization")
			// log.Printf("Authorization Header Sent: %v", authHeader)

			// user := c.Locals("user")
			// log.Printf("All Locals - user: %+v", user)
			// if token, ok := user.(*jwt.Token); ok {
			// 	log.Printf("Token Raw: %v", token.Raw)
			// 	log.Printf("Token Claims: %+v", token.Claims)
			// 	log.Printf("Token Header: %+v", token.Header)
			// }

			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// respond with 401 Unauthorized on token error
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		},
	}
	return jwtware.New(config)
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		token := user.(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userRole := claims["role"].(string)

		for _, role := range roles {
			if role == userRole {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
}

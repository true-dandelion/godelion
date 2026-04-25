package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var JwtSecret = []byte("super-secret-key-change-me") // In production, load from env

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid token"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return JwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		c.Locals("user_id", claims["sub"])
		c.Locals("role", claims["role"])

		return c.Next()
	}
}

func RoleRequired(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole != role && userRole != "admin" { // admin can access everything
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
		}
		return c.Next()
	}
}

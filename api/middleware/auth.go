package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var JwtSecret = []byte("super-secret-key-change-me") // In production, load from env

// Session store for d_delion_id validation (in production, use Redis)
var sessionStore = make(map[string]int64) // delionId -> userId

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

		// Validate X-Delion-Id header
		delionId := c.Get("X-Delion-Id")
		if delionId == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing d_delion_id"})
		}

		// Parse d_delion_id format: userId_timestamp
		parts := strings.Split(delionId, "_")
		if len(parts) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid d_delion_id format"})
		}

		delionUserId := parts[0]
		timestamp, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid d_delion_id timestamp"})
		}

		// Check if timestamp is within 7 days
		if time.Now().UnixMilli()-timestamp > 7*24*60*60*1000 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "d_delion_id expired"})
		}

		// Verify delionId matches JWT user
		jwtUserId := fmt.Sprintf("%v", claims["sub"])
		if delionUserId != jwtUserId {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "d_delion_id mismatch"})
		}

		c.Locals("user_id", claims["sub"])
		c.Locals("role", claims["role"])
		c.Locals("delion_id", delionId)

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

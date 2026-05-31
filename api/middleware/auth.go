package middleware

import (
	"fmt"
	"strings"

	"godelion/session"

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

		// Validate X-Delion-Id header
		delionId := c.Get("X-Delion-Id")
		if delionId == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing d_delion_id"})
		}

		// Check if d_delion_id exists in session store
		sess, exists := session.DelionSessionStore[delionId]
		if !exists {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未登录！"})
		}

		// Check if session expired (7 days)
		if sess.IsExpired() {
			// Clean up expired session
			delete(session.DelionSessionStore, delionId)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "d_delion_id expired"})
		}

		// Verify d_delion_id matches JWT user
		jwtUserId := fmt.Sprintf("%v", claims["sub"])
		sessionUserId := fmt.Sprintf("%v", sess.UserID)
		if sessionUserId != jwtUserId {
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

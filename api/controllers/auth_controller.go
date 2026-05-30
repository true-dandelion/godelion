package controllers

import (
	"fmt"
	"time"

	"godelion/db"
	"godelion/middleware"
	"godelion/models"
	"godelion/session"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "用户名或密码错误"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "用户名或密码错误"})
	}

	// Get session timeout from system config
	var config models.SystemConfig
	db.DB.First(&config)
	timeout := config.SessionTimeout
	if timeout <= 0 {
		timeout = 86400 // default 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Duration(timeout) * time.Second).Unix(),
	})

	tokenString, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not login"})
        }

        // Generate d_delion_id
	delionID := session.GenerateDelionID()
	session.DelionSessionStore[delionID] = &session.DelionSession{
		UserID:    user.ID,
		CreatedAt: time.Now(),
	}

	LogAction(c, "Login", "Auth", "User logged in: "+user.Username)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Login successful",
		"data": fiber.Map{
			"token":       tokenString,
			"d_delion_id": delionID,
			"user": fiber.Map{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
			},
		},
	})
}

func GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	userIDStr := fmt.Sprintf("%v", userID)
	var user models.User
	if err := db.DB.Where("id = ?", userIDStr).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	userIDStr := fmt.Sprintf("%v", userID)
	var user models.User
	if err := db.DB.Where("id = ?", userIDStr).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	var req struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username != "" {
		// Check if username already exists for another user
		var existing models.User
		if err := db.DB.Where("username = ? AND id != ?", req.Username, userIDStr).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username already exists"})
		}
		user.Username = req.Username
	}

	if req.NewPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "旧密码错误"})
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		user.PasswordHash = string(hashedPassword)
	}

	if err := db.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update profile"})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Profile updated successfully",
	})
}

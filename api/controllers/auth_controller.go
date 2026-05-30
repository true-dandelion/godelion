package controllers

import (
	"fmt"
	"strings"
	"time"

	"godelion/db"
	"godelion/middleware"
	"godelion/models"
	"godelion/session"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 400, "message": "请求格式错误"})
	}

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "用户名或密码错误"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "用户名或密码错误"})
	}

	// Check if 2FA is enabled
	var config models.SystemConfig
	db.DB.First(&config)

	if config.TwoFactorEnabled && config.TwoFactorSecret != "" {
		// 2FA enabled: generate temp token for 2FA verification step
		tempToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":  user.ID,
			"role": user.Role,
			"type": "2fa_pending",
			"exp":  time.Now().Add(5 * time.Minute).Unix(), // 5 min to complete 2FA
		})
		tempTokenString, err := tempToken.SignedString(middleware.JwtSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": 500, "message": "登录失败"})
		}

		return c.JSON(fiber.Map{
			"code":    200,
			"message": "需要两步验证",
			"data": fiber.Map{
				"require_2fa": true,
				"temp_token":   tempTokenString,
			},
		})
	}

	// No 2FA: proceed with normal login
	return completeLogin(c, user, config)
}

func completeLogin(c *fiber.Ctx, user models.User, config models.SystemConfig) error {
	timeout := config.SessionTimeout
	if timeout <= 0 {
		timeout = 86400
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Duration(timeout) * time.Second).Unix(),
	})

	tokenString, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": 500, "message": "登录失败"})
	}

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
			"require_2fa": false,
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

// VerifyLogin2FA handles 2FA verification during login
func VerifyLogin2FA(c *fiber.Ctx) error {
	type Req struct {
		TempToken string `json:"temp_token"`
		Code      string `json:"code"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 400, "message": "请求格式错误"})
	}

	// Validate temp token
	tokenString := strings.TrimPrefix(req.TempToken, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return middleware.JwtSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "验证已过期，请重新登录"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "验证已过期，请重新登录"})
	}

	// Check this is a 2fa_pending token
	if claims["type"] != "2fa_pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 400, "message": "无效的验证请求"})
	}

	// Validate TOTP code
	var config models.SystemConfig
	db.DB.First(&config)

	if !config.TwoFactorEnabled || config.TwoFactorSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 400, "message": "两步验证未开启"})
	}

	valid := totp.Validate(req.Code, config.TwoFactorSecret)
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "验证码错误"})
	}

	// Get user and complete login
	userID := fmt.Sprintf("%v", claims["sub"])
	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": 500, "message": "用户不存在"})
	}

	return completeLogin(c, user, config)
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

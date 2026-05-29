package controllers

import (
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"regexp"
	"bytes"

	"godelion/db"
	"godelion/middleware"
	"godelion/models"

	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// GetSystemConfig returns the current system configuration
func GetSystemConfig(c *fiber.Ctx) error {
	var config models.SystemConfig
	result := db.DB.First(&config)
	
	if result.Error != nil {
		// Create default config if not exists
		config = models.SystemConfig{
			PanelName:          "Godelion",
			SessionTimeout:     86400,
			EnableHTTPS:        false,
			HTTPSPort:          443,
			HTTPPort:           8080,
			PasswordExpiryDays: 0,
			PasswordComplexity: false,
			TwoFactorEnabled:   false,
		}
		db.DB.Create(&config)
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    config,
	})
}

// UpdateSystemConfig updates system configuration
func UpdateSystemConfig(c *fiber.Ctx) error {
	var req models.SystemConfig
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var config models.SystemConfig
	result := db.DB.First(&config)
	
	if result.Error != nil {
		// Create new config
		config = req
		db.DB.Create(&config)
	} else {
		// Update existing config
		config.PanelName = req.PanelName
		config.SessionTimeout = req.SessionTimeout
		config.EnableHTTPS = req.EnableHTTPS
		config.HTTPSPort = req.HTTPSPort
		config.HTTPPort = req.HTTPPort
		config.SecureEntrypoint = req.SecureEntrypoint
		config.AuthorizedIPs = req.AuthorizedIPs
		config.DomainBinding = req.DomainBinding
		config.PasswordExpiryDays = req.PasswordExpiryDays
		config.PasswordComplexity = req.PasswordComplexity
		config.TwoFactorEnabled = req.TwoFactorEnabled
		db.DB.Save(&config)
	}

	// Update access control cache
	middleware.SetAccessConfig(config.DomainBinding, config.AuthorizedIPs)

	LogAction(c, "Update", "SystemConfig", "Updated system configuration")

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Configuration updated successfully",
		"data":    config,
	})
}

// ChangeUsername changes the current user's username
func ChangeUsername(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	userIDStr, ok := userID.(string)
	if !ok {
		userIDStr = fmt.Sprintf("%v", userID)
	}

	type Req struct {
		NewUsername string `json:"new_username"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.NewUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username cannot be empty"})
	}

	// Check if username already exists
	var existingUser models.User
	if db.DB.Where("username = ? AND id != ?", req.NewUsername, userIDStr).First(&existingUser).Error == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username already exists"})
	}

	var user models.User
	if db.DB.First(&user, "id = ?", userIDStr).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	user.Username = req.NewUsername
	db.DB.Save(&user)

	LogAction(c, "Update", "User", "Changed username to: "+req.NewUsername)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Username changed successfully",
	})
}

// ChangePassword changes the current user's password
func ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	userIDStr, ok := userID.(string)
	if !ok {
		// JWT may store sub as float64
		userIDStr = fmt.Sprintf("%v", userID)
	}

	type Req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user models.User
	if db.DB.First(&user, "id = ?", userIDStr).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Verify current password
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)) != nil {
		log.Printf("Password mismatch for user %s: hash=%s input=%s", userIDStr, user.PasswordHash, req.CurrentPassword)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Current password is incorrect"})
	}

	// Check password complexity if enabled
	var config models.SystemConfig
	db.DB.First(&config)
	if config.PasswordComplexity {
		if len(req.NewPassword) < 8 || len(req.NewPassword) > 30 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password must be 8-30 characters"})
		}
		hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(req.NewPassword)
		hasDigit := regexp.MustCompile(`[0-9]`).MatchString(req.NewPassword)
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(req.NewPassword)
		count := 0
		if hasLetter { count++ }
		if hasDigit { count++ }
		if hasSpecial { count++ }
		if count < 2 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password must contain at least two of: letters, numbers, special characters"})
		}
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	user.PasswordHash = string(hashedPassword)
	db.DB.Save(&user)

	LogAction(c, "Update", "User", "Changed password")

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Password changed successfully",
	})
}

// GetPasskeys returns user's passkeys
func GetPasskeys(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))

	var passkeys []models.Passkey
	db.DB.Where("user_id = ?", userID).Find(&passkeys)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    passkeys,
	})
}

// CreatePasskey creates a new passkey for the user
func CreatePasskey(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))

	type Req struct {
		Name        string `json:"name"`
		CredentialID string `json:"credential_id"`
		PublicKey   string `json:"public_key"`
		Counter     uint32 `json:"counter"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check max 5 passkeys
	var count int64
	db.DB.Model(&models.Passkey{}).Where("user_id = ?", userID).Count(&count)
	if count >= 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Maximum 5 passkeys allowed"})
	}

	passkey := models.Passkey{
		UserID:       userID,
		Name:         req.Name,
		CredentialID: req.CredentialID,
		PublicKey:    req.PublicKey,
		Counter:      req.Counter,
	}
	db.DB.Create(&passkey)

	LogAction(c, "Create", "Passkey", "Created passkey: "+req.Name)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Passkey created successfully",
		"data":    passkey,
	})
}

// DeletePasskey deletes a passkey
func DeletePasskey(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))
	passkeyID := c.Params("id")

	var passkey models.Passkey
	if db.DB.First(&passkey, "id = ? AND user_id = ?", passkeyID, userID).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Passkey not found"})
	}

	db.DB.Delete(&passkey)

	LogAction(c, "Delete", "Passkey", "Deleted passkey: "+passkey.Name)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Passkey deleted successfully",
	})
}

// Get2FAStatus returns current 2FA status
func Get2FAStatus(c *fiber.Ctx) error {
	var config models.SystemConfig
	db.DB.First(&config)

	enabled := config.TwoFactorEnabled && config.TwoFactorSecret != ""

	return c.JSON(fiber.Map{
		"code": 200,
		"data": fiber.Map{
			"enabled": enabled,
		},
	})
}

// Generate2FASecret generates a new TOTP secret and returns QR code
func Generate2FASecret(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	userIDStr, _ := userID.(string)
	if userIDStr == "" {
		userIDStr = fmt.Sprintf("%v", userID)
	}

	var user models.User
	db.DB.First(&user, "id = ?", userIDStr)

	var config models.SystemConfig
	db.DB.First(&config)

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Godelion",
		AccountName: user.Username,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate 2FA secret"})
	}

	// Store secret temporarily (not enabled until verified)
	config.TwoFactorSecret = key.Secret()
	db.DB.Save(&config)

	// Generate QR code as base64 image
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate QR code"})
	}
	png.Encode(&buf, img)
	qrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return c.JSON(fiber.Map{
		"code": 200,
		"data": fiber.Map{
			"secret":    key.Secret(),
			"qr_code":   "data:image/png;base64," + qrBase64,
			"issuer":    "Godelion",
			"account":   user.Username,
		},
	})
}

// Verify2FA verifies the user's TOTP code to enable 2FA
func Verify2FA(c *fiber.Ctx) error {
	type Req struct {
		Code string `json:"code"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var config models.SystemConfig
	db.DB.First(&config)

	if config.TwoFactorSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Please generate a 2FA secret first"})
	}

	valid := totp.Validate(req.Code, config.TwoFactorSecret)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid verification code"})
	}

	// Enable 2FA
	config.TwoFactorEnabled = true
	db.DB.Save(&config)

	LogAction(c, "Update", "Security", "Enabled 2FA")

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "2FA enabled successfully",
	})
}

// Disable2FA disables 2FA
func Disable2FA(c *fiber.Ctx) error {
	type Req struct {
		Code string `json:"code"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var config models.SystemConfig
	db.DB.First(&config)

	if !config.TwoFactorEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "2FA is not enabled"})
	}

	// Verify code before disabling
	if config.TwoFactorSecret != "" {
		valid := totp.Validate(req.Code, config.TwoFactorSecret)
		if !valid {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid verification code"})
		}
	}

	config.TwoFactorEnabled = false
	config.TwoFactorSecret = ""
	db.DB.Save(&config)

	LogAction(c, "Update", "Security", "Disabled 2FA")

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "2FA disabled successfully",
	})
}
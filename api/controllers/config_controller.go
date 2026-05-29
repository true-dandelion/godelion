package controllers

import (
	"log"
	"regexp"

	"godelion/db"
	"godelion/models"

	"github.com/gofiber/fiber/v2"
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

	LogAction(c, "Update", "SystemConfig", "Updated system configuration")

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Configuration updated successfully",
		"data":    config,
	})
}

// ChangeUsername changes the current user's username
func ChangeUsername(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

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
	if db.DB.Where("username = ? AND id != ?", req.NewUsername, userID).First(&existingUser).Error == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username already exists"})
	}

	var user models.User
	if db.DB.First(&user, "id = ?", userID).Error != nil {
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
	userID := c.Locals("user_id").(string)

	type Req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user models.User
	if db.DB.First(&user, "id = ?", userID).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Verify current password
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)) != nil {
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
	userID := c.Locals("user_id").(string)

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
	userID := c.Locals("user_id").(string)

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
	userID := c.Locals("user_id").(string)
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
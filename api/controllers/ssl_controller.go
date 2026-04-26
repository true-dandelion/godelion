package controllers

import (
        "time"

        "godelion/db"
        "godelion/models"

        "github.com/gofiber/fiber/v2"
        "github.com/google/uuid"
)

// ListSSLCerts returns a list of all SSL certificates (without private keys for security)
func ListSSLCerts(c *fiber.Ctx) error {
        var certs []models.SSLCertificate
        if err := db.DB.Select("id", "domain", "created_at", "updated_at").Find(&certs).Error; err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch certificates"})
        }

        return c.JSON(fiber.Map{
                "code":    200,
                "message": "Success",
                "data":    certs,
        })
}

// CreateSSLCert handles both pasting and uploading (which frontend will convert to string)
func CreateSSLCert(c *fiber.Ctx) error {
        var payload struct {
                Domain      string `json:"domain"`
                CertContent string `json:"cert_content"`
                KeyContent  string `json:"key_content"`
        }

        if err := c.BodyParser(&payload); err != nil {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
        }

        if payload.Domain == "" || payload.CertContent == "" || payload.KeyContent == "" {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Domain, Cert Content and Key Content are required"})
        }

        // Check if domain already has a cert
        var existing models.SSLCertificate
        if err := db.DB.Where("domain = ?", payload.Domain).First(&existing).Error; err == nil {
                // Update existing
                existing.CertContent = payload.CertContent
                existing.KeyContent = payload.KeyContent
                existing.UpdatedAt = time.Now()
                db.DB.Save(&existing)
                return c.JSON(fiber.Map{"code": 200, "message": "Certificate updated successfully", "data": existing})
        }

        cert := models.SSLCertificate{
                ID:          uuid.NewString(),
                Domain:      payload.Domain,
                CertContent: payload.CertContent,
                KeyContent:  payload.KeyContent,
        }

        if err := db.DB.Create(&cert).Error; err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save certificate"})
        }

        return c.JSON(fiber.Map{
                "code":    200,
                "message": "Certificate created successfully",
                "data":    cert,
        })
}

// DeleteSSLCert removes a certificate
func DeleteSSLCert(c *fiber.Ctx) error {
        id := c.Params("id")
        var cert models.SSLCertificate
        if err := db.DB.First(&cert, "id = ?", id).Error; err != nil {
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Certificate not found"})
        }

        if err := db.DB.Delete(&cert).Error; err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete certificate"})
        }

        return c.JSON(fiber.Map{
                "code":    200,
                "message": "Certificate deleted successfully",
        })
}

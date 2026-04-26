package controllers

import (
        "crypto/tls"
        "crypto/x509"
        "encoding/pem"
        "errors"
        "time"

        "godelion/db"
        "godelion/models"

        "github.com/gofiber/fiber/v2"
        "github.com/google/uuid"
)

// ListSSLCerts returns a list of all SSL certificates (without private keys for security)
func ListSSLCerts(c *fiber.Ctx) error {
        var certs []models.SSLCertificate
        if err := db.DB.Select("id", "domain", "issued_at", "expires_at", "created_at", "updated_at").Find(&certs).Error; err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch certificates"})
        }

        return c.JSON(fiber.Map{
                "code":    200,
                "message": "Success",
                "data":    certs,
        })
}

// Helper function to extract issue and expiration dates from PEM encoded cert
func getCertDates(certPEM []byte) (time.Time, time.Time, error) {
        block, _ := pem.Decode(certPEM)
        if block == nil {
                return time.Time{}, time.Time{}, errors.New("failed to parse certificate PEM")
        }
        
        cert, err := x509.ParseCertificate(block.Bytes)
        if err != nil {
                return time.Time{}, time.Time{}, err
        }
        
        return cert.NotBefore, cert.NotAfter, nil
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

        // Validate the certificate pair
        _, err := tls.X509KeyPair([]byte(payload.CertContent), []byte(payload.KeyContent))
        if err != nil {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid certificate or private key: " + err.Error()})
        }

        // Extract dates
        issuedAt, expiresAt, err := getCertDates([]byte(payload.CertContent))
        if err != nil {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Could not parse certificate dates: " + err.Error()})
        }

        // Check if domain already has a cert
        var existing models.SSLCertificate
        if err := db.DB.Where("domain = ?", payload.Domain).First(&existing).Error; err == nil {
                // Update existing
                existing.CertContent = payload.CertContent
                existing.KeyContent = payload.KeyContent
                existing.IssuedAt = issuedAt
                existing.ExpiresAt = expiresAt
                existing.UpdatedAt = time.Now()
                db.DB.Save(&existing)
                return c.JSON(fiber.Map{"code": 200, "message": "Certificate updated successfully", "data": existing})
        }

        cert := models.SSLCertificate{
                ID:          uuid.NewString(),
                Domain:      payload.Domain,
                CertContent: payload.CertContent,
                KeyContent:  payload.KeyContent,
                IssuedAt:    issuedAt,
                ExpiresAt:   expiresAt,
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

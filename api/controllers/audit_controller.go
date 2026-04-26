package controllers

import (
        "time"

        "godelion/db"
        "godelion/models"

        "github.com/gofiber/fiber/v2"
)

// LogAction is a helper function to record an audit log
func LogAction(c *fiber.Ctx, action, resource, details string) {
        userID := ""
        if val := c.Locals("user_id"); val != nil {
                userID = val.(string)
        }

        ip := c.IP()

        log := models.AuditLog{
                UserID:    userID,
                Action:    action,
                Resource:  resource,
                IPAddress: ip,
                Details:   details,
                CreatedAt: time.Now(),
        }

        // Run async to not block the request
        go func() {
                db.DB.Create(&log)
        }()
}

func ListAuditLogs(c *fiber.Ctx) error {
        var logs []models.AuditLog
        // Get the latest 50 logs
        if err := db.DB.Order("created_at desc").Limit(50).Find(&logs).Error; err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch audit logs"})
        }

        return c.JSON(fiber.Map{
                "code":    200,
                "message": "Success",
                "data":    logs,
        })
}

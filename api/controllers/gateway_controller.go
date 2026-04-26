package controllers

import (
	"encoding/json"
	
	"godelion/db"
	"godelion/models"
	"godelion/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateGatewayRule(c *fiber.Ctx) error {
	var payload struct {
		Domain      string `json:"domain"`
		ListenPorts string `json:"listen_ports"`
		TargetURLs  string `json:"target_urls"`
		TLSEnabled  bool   `json:"tls_enabled"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	rule := models.GatewayRule{
		ID:          uuid.NewString(),
		Domain:      payload.Domain,
		ListenPorts: payload.ListenPorts,
		TargetURLs:  payload.TargetURLs,
		TLSEnabled:  payload.TLSEnabled,
	}

	if err := db.DB.Create(&rule).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create rule"})
	}

	services.UpdateProxyRule(rule)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Gateway rule created",
		"data":    rule,
	})
}

func ListGatewayRules(c *fiber.Ctx) error {
	var rules []models.GatewayRule
	if err := db.DB.Find(&rules).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch rules"})
	}

	// We will create a unified response format that includes both domain-based GatewayRules and HostPort mappings
	type UnifiedRule struct {
		ID            string `json:"id"`
		Domain        string `json:"domain"`
		ListenPorts   string `json:"listen_ports"`
		TargetURLs    string `json:"target_urls"`
		TLSEnabled    bool   `json:"tls_enabled"`
		IsPortMapping bool   `json:"is_port_mapping"`
	}

	var unifiedRules []UnifiedRule

	// 1. Add standard domain gateway rules
	for _, r := range rules {
		unifiedRules = append(unifiedRules, UnifiedRule{
			ID:            r.ID,
			Domain:        r.Domain,
			ListenPorts:   r.ListenPorts,
			TargetURLs:    r.TargetURLs,
			TLSEnabled:    r.TLSEnabled,
			IsPortMapping: false,
		})
	}

	// 2. Add container port mappings
	var containers []models.Container
	if err := db.DB.Find(&containers).Error; err == nil {
		for _, container := range containers {
			if container.Ports == "" || container.Ports == "[]" {
				continue
			}

			var ports []services.WorkloadPort
			if err := json.Unmarshal([]byte(container.Ports), &ports); err == nil {
				for _, p := range ports {
					if p.Host != "" && p.Container != "" {
						tlsEnabled := p.Host == "443" // Simple assumption: 443 implies TLS

						unifiedRules = append(unifiedRules, UnifiedRule{
							ID:            "port-" + container.ID + "-" + p.Host,
							Domain:        "*:" + p.Host + " (主机端口代理)",
							ListenPorts:   p.Host,
							TargetURLs:    "Container: " + container.Name + ":" + p.Container,
							TLSEnabled:    tlsEnabled,
							IsPortMapping: true,
						})
					}
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    unifiedRules,
	})
}

func DeleteGatewayRule(c *fiber.Ctx) error {
	id := c.Params("id")
	var rule models.GatewayRule
	if err := db.DB.First(&rule, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rule not found"})
	}

	if err := db.DB.Delete(&rule).Error; err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete rule"})
        }

        services.RemoveProxyRule(rule)

        return c.JSON(fiber.Map{
		"code":    200,
		"message": "Rule deleted",
	})
}

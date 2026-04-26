package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	
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
                ContainerID string `json:"container_id"`
                TargetPort  int    `json:"target_port"`
        }

        if err := c.BodyParser(&payload); err != nil {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
        }

        // Validate port conflicts
        if payload.ListenPorts != "" {
                for _, p := range strings.Split(payload.ListenPorts, ",") {
                        port := strings.TrimSpace(p)
                        if port == "" {
                                continue
                        }
                        isConflict, reason := services.CheckPortConflict(port, payload.Domain, "", "")
                        if isConflict {
                                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "端口 " + port + " 已被 [" + reason + "] 占用"})
                        }
                }
        }

        rule := models.GatewayRule{
		ID:          uuid.NewString(),
		Domain:      payload.Domain,
		ListenPorts: payload.ListenPorts,
		TargetURLs:  payload.TargetURLs,
		TLSEnabled:  payload.TLSEnabled,
		ContainerID: payload.ContainerID,
		TargetPort:  payload.TargetPort,
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
		targetDisplay := r.TargetURLs
		if r.ContainerID != "" {
			var c models.Container
			if err := db.DB.First(&c, "id = ?", r.ContainerID).Error; err == nil {
				targetDisplay = "Container: " + c.Name + " (" + fmt.Sprintf("%d", r.TargetPort) + ")"
			} else {
				targetDisplay = "Container: " + r.ContainerID + " (Not Found)"
			}
		} else if r.TargetURLs == "" && r.TargetPort > 0 {
			targetDisplay = fmt.Sprintf("127.0.0.1:%d", r.TargetPort)
		}

		unifiedRules = append(unifiedRules, UnifiedRule{
			ID:            r.ID,
			Domain:        r.Domain,
			ListenPorts:   r.ListenPorts,
			TargetURLs:    targetDisplay,
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

func UpdateGatewayRule(c *fiber.Ctx) error {
        id := c.Params("id")
        var payload struct {
                Domain      string `json:"domain"`
                ListenPorts string `json:"listen_ports"`
                TargetURLs  string `json:"target_urls"`
                TLSEnabled  bool   `json:"tls_enabled"`
                ContainerID string `json:"container_id"`
                TargetPort  int    `json:"target_port"`
        }

        if err := c.BodyParser(&payload); err != nil {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
        }

        var rule models.GatewayRule
        if err := db.DB.First(&rule, "id = ?", id).Error; err != nil {
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rule not found"})
        }

        // Validate port conflicts
        if payload.ListenPorts != "" {
                for _, p := range strings.Split(payload.ListenPorts, ",") {
                        port := strings.TrimSpace(p)
                        if port == "" {
                                continue
                        }
                        isConflict, reason := services.CheckPortConflict(port, payload.Domain, id, "")
                        if isConflict {
                                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "端口 " + port + " 已被 [" + reason + "] 占用"})
                        }
                }
        }

        // Remove old proxy rule from memory/listeners
        services.RemoveProxyRule(rule)

        rule.Domain = payload.Domain
        rule.ListenPorts = payload.ListenPorts
        rule.TargetURLs = payload.TargetURLs
        rule.TLSEnabled = payload.TLSEnabled
        rule.ContainerID = payload.ContainerID
        rule.TargetPort = payload.TargetPort

        if err := db.DB.Save(&rule).Error; err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update rule"})
        }

        // Apply new proxy rule
        services.UpdateProxyRule(rule)

        return c.JSON(fiber.Map{
                "code":    200,
                "message": "Successfully updated",
                "data":    rule,
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

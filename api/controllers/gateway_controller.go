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
		Domain       string `json:"domain"`
		HTTPPort     string `json:"http_port"`
		HTTPSPort    string `json:"https_port"`
		TargetURLs   string `json:"target_urls"`
		TLSEnabled   bool   `json:"tls_enabled"`
		SSLCertID    string `json:"ssl_cert_id"`
		ContainerID  string `json:"container_id"`
		TargetPort   int    `json:"target_port"`
		RuleType     string `json:"rule_type"`
		RedirectURL  string `json:"redirect_url"`
		RedirectCode int    `json:"redirect_code"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求格式错误"})
	}

	// Default rule type
	if payload.RuleType == "" {
		payload.RuleType = "proxy"
	}
	if payload.RedirectCode == 0 {
		payload.RedirectCode = 301
	}

	// Validate: HTTPS enabled requires https_port and ssl_cert_id
	if payload.TLSEnabled {
		if payload.HTTPSPort == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "开启 HTTPS 时必须填写 HTTPS 端口"})
		}
		if payload.SSLCertID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "开启 HTTPS 时必须选择 SSL 证书"})
		}
	} else {
		if payload.HTTPPort == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "未开启 HTTPS 时必须填写 HTTP 端口"})
		}
	}

	// Validate port conflicts
	portsToCheck := []string{}
	if payload.HTTPPort != "" {
		portsToCheck = append(portsToCheck, payload.HTTPPort)
	}
	if payload.TLSEnabled && payload.HTTPSPort != "" {
		portsToCheck = append(portsToCheck, payload.HTTPSPort)
	}
	for _, port := range portsToCheck {
		port = strings.TrimSpace(port)
		if port == "" {
			continue
		}
		isConflict, reason := services.CheckPortConflict(port, payload.Domain, "", "")
		if isConflict {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "端口 " + port + " 已被 [" + reason + "] 占用"})
		}
	}

	rule := models.GatewayRule{
		ID:           uuid.NewString(),
		Domain:       payload.Domain,
		HTTPPort:     payload.HTTPPort,
		HTTPSPort:    payload.HTTPSPort,
		TargetURLs:   payload.TargetURLs,
		TLSEnabled:   payload.TLSEnabled,
		SSLCertID:    payload.SSLCertID,
		ContainerID:  payload.ContainerID,
		TargetPort:   payload.TargetPort,
		RuleType:     payload.RuleType,
		RedirectURL:  payload.RedirectURL,
		RedirectCode: payload.RedirectCode,
	}

	if err := db.DB.Create(&rule).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "创建规则失败"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "获取规则失败"})
	}

	type UnifiedRule struct {
		ID            string `json:"id"`
		Domain        string `json:"domain"`
		HTTPPort      string `json:"http_port"`
		HTTPSPort     string `json:"https_port"`
		TargetURLs    string `json:"target_urls"`
		TLSEnabled    bool   `json:"tls_enabled"`
		SSLCertID     string `json:"ssl_cert_id"`
		IsPortMapping bool   `json:"is_port_mapping"`
		ContainerID   string `json:"container_id,omitempty"`
		RuleType      string `json:"rule_type"`
		RedirectURL   string `json:"redirect_url,omitempty"`
		RedirectCode  int    `json:"redirect_code,omitempty"`
	}

	var unifiedRules []UnifiedRule

	for _, r := range rules {
		targetDisplay := r.TargetURLs
		if r.RuleType == "redirect" {
			targetDisplay = "→ " + r.RedirectURL + " (" + fmt.Sprintf("%d", r.RedirectCode) + ")"
		} else if r.ContainerID != "" {
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
			HTTPPort:      r.HTTPPort,
			HTTPSPort:     r.HTTPSPort,
			TargetURLs:    targetDisplay,
			TLSEnabled:    r.TLSEnabled,
			SSLCertID:     r.SSLCertID,
			IsPortMapping: false,
			ContainerID:   r.ContainerID,
			RuleType:      r.RuleType,
			RedirectURL:   r.RedirectURL,
			RedirectCode:  r.RedirectCode,
		})
	}

	// Add container port mappings
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
						tlsEnabled := p.Host == "443"

						unifiedRules = append(unifiedRules, UnifiedRule{
							ID:            "port-" + container.ID + "-" + p.Host,
							Domain:        "*:" + p.Host + " (主机端口代理)",
							HTTPPort:      p.Host,
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
		Domain       string `json:"domain"`
		HTTPPort     string `json:"http_port"`
		HTTPSPort    string `json:"https_port"`
		TargetURLs   string `json:"target_urls"`
		TLSEnabled   bool   `json:"tls_enabled"`
		SSLCertID    string `json:"ssl_cert_id"`
		ContainerID  string `json:"container_id"`
		TargetPort   int    `json:"target_port"`
		RuleType     string `json:"rule_type"`
		RedirectURL  string `json:"redirect_url"`
		RedirectCode int    `json:"redirect_code"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求格式错误"})
	}

	if payload.RuleType == "" {
		payload.RuleType = "proxy"
	}
	if payload.RedirectCode == 0 {
		payload.RedirectCode = 301
	}

	var rule models.GatewayRule
	if err := db.DB.First(&rule, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "规则不存在"})
	}

	// Validate: HTTPS enabled requires https_port and ssl_cert_id
	if payload.TLSEnabled {
		if payload.HTTPSPort == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "开启 HTTPS 时必须填写 HTTPS 端口"})
		}
		if payload.SSLCertID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "开启 HTTPS 时必须选择 SSL 证书"})
		}
	} else {
		if payload.HTTPPort == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "未开启 HTTPS 时必须填写 HTTP 端口"})
		}
	}

	// Validate port conflicts
	portsToCheck := []string{}
	if payload.HTTPPort != "" {
		portsToCheck = append(portsToCheck, payload.HTTPPort)
	}
	if payload.TLSEnabled && payload.HTTPSPort != "" {
		portsToCheck = append(portsToCheck, payload.HTTPSPort)
	}
	for _, port := range portsToCheck {
		port = strings.TrimSpace(port)
		if port == "" {
			continue
		}
		isConflict, reason := services.CheckPortConflict(port, payload.Domain, id, "")
		if isConflict {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "端口 " + port + " 已被 [" + reason + "] 占用"})
		}
	}

	// Remove old proxy rule
	services.RemoveProxyRule(rule)

	rule.Domain = payload.Domain
	rule.HTTPPort = payload.HTTPPort
	rule.HTTPSPort = payload.HTTPSPort
	rule.TargetURLs = payload.TargetURLs
	rule.TLSEnabled = payload.TLSEnabled
	rule.SSLCertID = payload.SSLCertID
	rule.ContainerID = payload.ContainerID
	rule.TargetPort = payload.TargetPort
	rule.RuleType = payload.RuleType
	rule.RedirectURL = payload.RedirectURL
	rule.RedirectCode = payload.RedirectCode

	if err := db.DB.Save(&rule).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "更新规则失败"})
	}

	services.UpdateProxyRule(rule)
	LogAction(c, "Update", "Gateway", "Updated gateway rule for: "+rule.Domain)

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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "规则不存在"})
	}

	if err := db.DB.Delete(&rule).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "删除规则失败"})
	}

	services.RemoveProxyRule(rule)
	LogAction(c, "Delete", "Gateway", "Deleted gateway rule for: "+rule.Domain)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Rule deleted",
	})
}

package middleware

import (
	"net"
	"strings"

	"godelion/db"
	"godelion/models"

	"github.com/gofiber/fiber/v2"
)

// AccessControl checks IP whitelist and domain binding
func AccessControl() fiber.Handler {
	return func(c *fiber.Ctx) error {
		config := getSystemConfigForMiddleware()

		// Check domain binding
		if config.DomainBinding != "" {
			host := c.Hostname()
			if host != config.DomainBinding {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Access denied: domain not allowed",
				})
			}
		}

		// Check authorized IPs
		if config.AuthorizedIPs != "" {
			clientIP := c.IP()

			// Extract real IP from proxy headers
			if xff := c.Get("X-Forwarded-For"); xff != "" {
				clientIP = strings.Split(xff, ",")[0]
				clientIP = strings.TrimSpace(clientIP)
			} else if xri := c.Get("X-Real-IP"); xri != "" {
				clientIP = strings.TrimSpace(xri)
			}

			// Always allow localhost (127.0.0.1 and ::1)
			if clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "localhost" {
				return c.Next()
			}

			allowed := false
			ips := strings.Split(config.AuthorizedIPs, ",")
			for _, allowedIP := range ips {
				allowedIP = strings.TrimSpace(allowedIP)
				if allowedIP == "" {
					continue
				}
				if isIPMatch(clientIP, allowedIP) {
					allowed = true
					break
				}
			}

			if !allowed {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Access denied: IP not authorized",
				})
			}
		}

		return c.Next()
	}
}

// isIPMatch checks if a client IP matches an allowed IP or CIDR
func isIPMatch(clientIP, allowedIP string) bool {
	// Exact match
	if clientIP == allowedIP {
		return true
	}

	// CIDR match
	if strings.Contains(allowedIP, "/") {
		_, ipNet, err := net.ParseCIDR(allowedIP)
		if err == nil {
			return ipNet.Contains(net.ParseIP(clientIP))
		}
	}

	return false
}

// getSystemConfigForMiddleware retrieves system config for middleware use
// Defined here to avoid circular imports
func getSystemConfigForMiddleware() struct {
	DomainBinding string
	AuthorizedIPs string
} {
	// Import db and models inline to avoid circular dependency
	// We use a simple approach: store config in a package-level variable
	return struct {
		DomainBinding string
		AuthorizedIPs string
	}{
		DomainBinding: accessDomainBinding,
		AuthorizedIPs: accessAuthorizedIPs,
	}
}

// Package-level config cache (updated by config_controller)
var (
	accessDomainBinding  string
	accessAuthorizedIPs string
)

// SetAccessConfig updates the cached access control config
func SetAccessConfig(domainBinding, authorizedIPs string) {
	accessDomainBinding = domainBinding
	accessAuthorizedIPs = authorizedIPs
}

// LoadAccessConfig loads config from DB into cache on startup
func LoadAccessConfig() {
	var config models.SystemConfig
	if db.DB.First(&config).Error == nil {
		accessDomainBinding = config.DomainBinding
		accessAuthorizedIPs = config.AuthorizedIPs
	}
}

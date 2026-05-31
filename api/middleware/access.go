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

		clientIP := c.IP()
		// Extract real IP from proxy headers
		if xff := c.Get("X-Forwarded-For"); xff != "" {
			clientIP = strings.Split(xff, ",")[0]
			clientIP = strings.TrimSpace(clientIP)
		} else if xri := c.Get("X-Real-IP"); xri != "" {
			clientIP = strings.TrimSpace(xri)
		}

		// Always allow localhost
		isLocalhost := clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "localhost"

		// Check domain binding
		if config.DomainBinding != "" && !isLocalhost {
			host := c.Hostname()
			if host != config.DomainBinding {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Access denied: domain not allowed",
				})
			}
		}

		// Check authorized IPs
		if config.AuthorizedIPs != "" {
			if isLocalhost {
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
	accessDomainBinding    string
	accessAuthorizedIPs    string
	accessSecureEntrypoint string
)

// SetAccessConfig updates the cached access control config
func SetAccessConfig(domainBinding, authorizedIPs, secureEntrypoint string) {
	accessDomainBinding = domainBinding
	accessAuthorizedIPs = authorizedIPs
	accessSecureEntrypoint = secureEntrypoint
}

// LoadAccessConfig loads config from DB into cache on startup
func LoadAccessConfig() {
	var config models.SystemConfig
	if db.DB.First(&config).Error == nil {
		accessDomainBinding = config.DomainBinding
		accessAuthorizedIPs = config.AuthorizedIPs
		accessSecureEntrypoint = config.SecureEntrypoint
	}
}

// GetSecureEntrypoint returns the current secure entrypoint
func GetSecureEntrypoint() string {
	return accessSecureEntrypoint
}

// SecureEntrypointCheck is a middleware that enforces secure entrypoint access
// When secure_entrypoint is set, users must visit the entrypoint URL first
// to get a session cookie before accessing any other page
func SecureEntrypointCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		entrypoint := accessSecureEntrypoint
		if entrypoint == "" {
			return c.Next()
		}

		// Normalize entrypoint: ensure it starts with "/"
		if !strings.HasPrefix(entrypoint, "/") {
			entrypoint = "/" + entrypoint
		}

		path := c.Path()

		// Allow API routes to pass through (they have their own auth)
		if strings.HasPrefix(path, "/sys/v1") {
			return c.Next()
		}

		// Allow static assets (js, css, images, etc.)
		if isStaticAsset(path) {
			return c.Next()
		}

		// If user visits the entrypoint URL, set cookie and redirect to /
		if path == entrypoint {
			c.Cookie(&fiber.Cookie{
				Name:     "godelion_entry",
				Value:    "1",
				Path:     "/",
				HTTPOnly: true,
				MaxAge:   86400, // 24 hours
			})
			return c.Redirect("/")
		}

		// Check if user has the entrypoint cookie
		cookie := c.Cookies("godelion_entry")
		if cookie == "1" {
			return c.Next()
		}

		// No cookie - deny access
		return c.Status(fiber.StatusForbidden).SendString("Access denied")
	}
}

// isStaticAsset checks if the path is a static asset file
func isStaticAsset(path string) bool {
	staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map"}
	for _, ext := range staticExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

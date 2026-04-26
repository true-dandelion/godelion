package services

import (
	"crypto/tls"
	"fmt"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"godelion/db"
	"godelion/models"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type TargetPool struct {
	Targets []string
	Current int
	Mutex   sync.Mutex
}

func (p *TargetPool) Next() string {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if len(p.Targets) == 0 {
		return ""
	}
	t := p.Targets[p.Current]
	p.Current = (p.Current + 1) % len(p.Targets)
	return t
}

var (
	proxyTargetPools = make(map[string]*TargetPool)
	proxyMutex       sync.RWMutex
)

// InitProxy loads gateway rules
func InitProxy() {
	var rules []models.GatewayRule
	db.DB.Find(&rules)
	for _, rule := range rules {
		UpdateProxyRule(rule)
	}
}

func UpdateProxyRule(rule models.GatewayRule) {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	targets := []string{}
	
	// Support both legacy TargetPort and new TargetURLs
	if rule.TargetURLs != "" {
		for _, t := range strings.Split(rule.TargetURLs, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				targets = append(targets, t)
			}
		}
	}
	
	// If container target is specified, we resolve it dynamically.
	// We can inject a special target marker that ProxyHandler will resolve at runtime.
	if rule.ContainerID != "" && rule.TargetPort > 0 {
		targets = append(targets, fmt.Sprintf("@container:%s:%d", rule.ContainerID, rule.TargetPort))
	} else if rule.TargetPort > 0 && rule.TargetURLs == "" {
		// Legacy mapping fallback
		targets = append(targets, fmt.Sprintf("127.0.0.1:%d", rule.TargetPort))
	}

	// Ensure dynamic listeners are started for the new ListenPorts field
	if rule.ListenPorts != "" {
		for _, portStr := range strings.Split(rule.ListenPorts, ",") {
			portStr = strings.TrimSpace(portStr)
			if portStr != "" {
				// Key by domain:port to allow same domain on different ports
				key := rule.Domain + ":" + portStr
				proxyTargetPools[key] = &TargetPool{
					Targets: targets,
					Current: 0,
				}
				// We only ensure dynamic listener if it's not the main web UI port
				if portStr != SystemPort {
					EnsureListenerRunning(portStr)
				}
			}
		}
	} else {
		// Fallback for rules without ListenPorts
		proxyTargetPools[rule.Domain+":80"] = &TargetPool{
			Targets: targets,
			Current: 0,
		}
	}
}

func RemoveProxyRule(rule models.GatewayRule) {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	if rule.ListenPorts != "" {
		for _, portStr := range strings.Split(rule.ListenPorts, ",") {
			portStr = strings.TrimSpace(portStr)
			if portStr != "" {
				delete(proxyTargetPools, rule.Domain+":"+portStr)
				go CheckAndStopUnusedListener(portStr)
			}
		}
	} else {
		delete(proxyTargetPools, rule.Domain+":80")
		go CheckAndStopUnusedListener("80")
	}
}

func ProxyHandler(c *fiber.Ctx) error {
        host := c.Hostname()
        
        // Hostname in fiber might not include the port if standard 80/443
        // We need the port to route correctly. Assuming fiber is on SystemPort or TLS 443
        port := SystemPort
        if c.Protocol() == "https" {
                port = "443"
        }
        
        // If Host header explicitly has a port, extract it
        hostHeader := c.Get("Host")
        if strings.Contains(hostHeader, ":") {
                _, portPart, _ := strings.Cut(hostHeader, ":")
                port = portPart
        }

        key := host + ":" + port

        proxyMutex.RLock()
        pool, exists := proxyTargetPools[key]
        proxyMutex.RUnlock()

	if !exists {
		// Not found, continue to next handler (e.g., Godelion UI)
		return c.Next()
	}

	targetStr := pool.Next()
        if targetStr == "" {
                return c.Status(502).SendString("Bad Gateway: No targets available")
        }

        // Handle dynamic container resolution
        if strings.HasPrefix(targetStr, "@container:") {
                parts := strings.SplitN(targetStr, ":", 3) // @container:uuid:port
                if len(parts) == 3 {
                        containerID := parts[1]
                        containerPort := parts[2]
                        
                        // Find the Docker container ID from the DB UUID
                        var container models.Container
                        if err := db.DB.First(&container, "id = ?", containerID).Error; err == nil && container.DockerID != "" {
                                ip := getContainerIP(container.DockerID)
                                if ip != "" {
                                        targetStr = fmt.Sprintf("http://%s:%s", ip, containerPort)
                                } else {
                                        return c.Status(502).SendString("Bad Gateway: Container IP not found")
                                }
                        } else {
                                return c.Status(502).SendString("Bad Gateway: Container not running or not found")
                        }
                }
        }

        if !strings.HasPrefix(targetStr, "http://") && !strings.HasPrefix(targetStr, "https://") {
                targetStr = "http://" + targetStr
        }

        targetURL, _ := url.Parse(targetStr)
        proxy := httputil.NewSingleHostReverseProxy(targetURL)

        // Use fasthttp adaptor to serve httputil.ReverseProxy
        handler := fasthttpadaptor.NewFastHTTPHandler(proxy)
        handler(c.Context())
        return nil
}

// GetTLSConfig fetches TLS config for domains
func GetTLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			serverName := hello.ServerName
			
			// If no SNI (Server Name Indication) is provided by the client,
			// or it's an IP address access without domain, we cannot strictly match a domain cert.
			// We will try to find a wildcard cert or a default cert if necessary, but for now we reject or find ANY valid cert.
			if serverName == "" {
				// For direct IP access, try to find a gateway rule that has empty or IP domain
				var rule models.GatewayRule
				result := db.DB.Where("tls_enabled = ?", true).First(&rule)
				if result.Error == nil {
					var anyCert models.SSLCertificate
					if rule.SSLCertID != "" {
						db.DB.Where("id = ?", rule.SSLCertID).First(&anyCert)
					} else {
						db.DB.First(&anyCert)
					}
					if anyCert.CertContent != "" {
						cert, err := tls.X509KeyPair([]byte(anyCert.CertContent), []byte(anyCert.KeyContent))
						if err == nil {
							return &cert, nil
						}
					}
				}
				return nil, fmt.Errorf("no SNI server name provided, and no fallback certificate available")
			}

			var rule models.GatewayRule
			result := db.DB.Where("domain = ? AND tls_enabled = ?", serverName, true).First(&rule)
			if result.Error != nil {
				// If exact match fails, see if we have a wildcard cert that matches
				parts := strings.SplitN(serverName, ".", 2)
				if len(parts) == 2 {
					wildcardDomain := "*." + parts[1]
					result = db.DB.Where("domain = ? AND tls_enabled = ?", wildcardDomain, true).First(&rule)
				}
			}
			
			if result.Error != nil {
				return nil, fmt.Errorf("no active gateway rule found for %s", serverName)
			}
			
			var sslCert models.SSLCertificate
			if rule.SSLCertID != "" {
				if err := db.DB.Where("id = ?", rule.SSLCertID).First(&sslCert).Error; err != nil {
					return nil, err
				}
			} else {
				// Try to find a cert by exact domain name
				if err := db.DB.Where("domain = ?", serverName).First(&sslCert).Error; err != nil {
					// If exact domain fails, try wildcard matching for certs
					parts := strings.SplitN(serverName, ".", 2)
					if len(parts) == 2 {
						wildcardDomain := "*." + parts[1]
						err = db.DB.Where("domain = ?", wildcardDomain).First(&sslCert).Error
					}
					
					if err != nil {
						// Fallback to old path based cert if SSLCertificate not found
						cert, err := tls.LoadX509KeyPair(rule.CertPath, rule.KeyPath)
						if err != nil {
							return nil, err
						}
						return &cert, nil
					}
				}
			}
			
			cert, err := tls.X509KeyPair([]byte(sslCert.CertContent), []byte(sslCert.KeyContent))
			if err != nil {
				return nil, err
			}
			return &cert, nil
		},
	}
}

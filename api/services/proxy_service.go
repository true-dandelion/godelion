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
	} else if rule.TargetPort > 0 {
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
				if portStr != "8080" {
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
        // We need the port to route correctly. Assuming fiber is on 8080 or TLS 443
        port := "8080"
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
			var rule models.GatewayRule
			if err := db.DB.Where("domain = ? AND tls_enabled = ?", hello.ServerName, true).First(&rule).Error; err != nil {
				return nil, fmt.Errorf("no certificate found for %s", hello.ServerName)
			}
			cert, err := tls.LoadX509KeyPair(rule.CertPath, rule.KeyPath)
			if err != nil {
				return nil, err
			}
			return &cert, nil
		},
	}
}

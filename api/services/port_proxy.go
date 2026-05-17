package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"godelion/db"
	"godelion/models"
)

var (
        // Map of hostPort -> *http.Server
        proxyServers = make(map[string]*http.Server)
        workloadProxyMutex   sync.Mutex
        
        // SystemPort tracks the main port used by the Godelion API to prevent it from being stopped
        SystemPort = "8080"
)

// A generic structure to hold parsed ports
type WorkloadPort struct {
	Host      string `json:"host"`
	Container string `json:"container"`
}

// LoadAndStartAllProxies should be called on startup to restore proxy listeners for all running workloads
func LoadAndStartAllProxies() {
	var containers []models.Container
	db.DB.Find(&containers)

	for _, c := range containers {
		if c.DockerID == "" || c.Ports == "" || c.Ports == "[]" {
			continue
		}

		// Only start if the container is actually running
		info, err := DockerClient.ContainerInspect(context.Background(), c.DockerID)
		if err == nil && info.State.Running {
			StartProxiesForContainer(c)
		}
	}
}

// StartProxiesForContainer parses the container ports and starts listeners
func StartProxiesForContainer(c models.Container) {
	if c.Ports == "" || c.Ports == "[]" {
		return
	}

	var ports []WorkloadPort
	if err := json.Unmarshal([]byte(c.Ports), &ports); err != nil {
		log.Printf("[Proxy Error] Failed to parse ports for container %s: %v", c.ID, err)
		return
	}

	for _, p := range ports {
		if p.Host != "" && p.Container != "" {
			EnsureListenerRunning(p.Host)
		}
	}
}

// CheckAndStopUnusedListener checks if a port is no longer used by any container or gateway rule, and stops it
func CheckAndStopUnusedListener(port string) {
        if port == SystemPort {
                return // Never stop the main UI port
        }

        // Check if any gateway rule is still using this port
        inUse := false
        var rules []models.GatewayRule
        db.DB.Find(&rules)
        for _, r := range rules {
                for _, p := range strings.Split(r.ListenPorts, ",") {
                        if strings.TrimSpace(p) == port {
                                inUse = true
                                break
                        }
                }
        }

        // Check if any container is still using this port
        if !inUse {
                var containers []models.Container
                db.DB.Find(&containers)
                for _, c := range containers {
                        if c.Ports == "" || c.Ports == "[]" {
                                continue
                        }
                        var ports []WorkloadPort
                        json.Unmarshal([]byte(c.Ports), &ports)
                        for _, p := range ports {
                                if p.Host == port {
                                        inUse = true
                                        break
                                }
                        }
                        if inUse {
                                break
                        }
                }
        }

        if !inUse {
                StopSingleProxy(port)
        }
}

// StopProxiesForContainer triggers a check to stop listeners that are no longer needed
func StopProxiesForContainer(c models.Container) {
        if c.Ports == "" || c.Ports == "[]" {
                return
        }

        var ports []WorkloadPort
        if err := json.Unmarshal([]byte(c.Ports), &ports); err == nil {
                for _, p := range ports {
                        if p.Host != "" {
                                go CheckAndStopUnusedListener(p.Host)
                        }
                }
        }
}

// CheckPortConflict checks if a given port (and optionally domain) is already in use.
// Returns (isConflict bool, conflictReason string)
func CheckPortConflict(port string, domain string, excludeRuleID string, excludeContainerID string) (bool, string) {
        if port == SystemPort {
                return true, "Godelion系统服务"
        }

        // Check containers
        var containers []models.Container
        db.DB.Find(&containers)
        for _, c := range containers {
                if c.ID == excludeContainerID {
                        continue
                }
                if c.Ports != "" && c.Ports != "[]" {
                        var ports []WorkloadPort
                        json.Unmarshal([]byte(c.Ports), &ports)
                        for _, p := range ports {
                                if p.Host == port {
                                        return true, "容器: " + c.Name
                                }
                        }
                }
        }

        // Check gateway rules
        var rules []models.GatewayRule
        db.DB.Find(&rules)
        for _, r := range rules {
                if r.ID == excludeRuleID {
                        continue
                }
                for _, rp := range strings.Split(r.ListenPorts, ",") {
                        rp = strings.TrimSpace(rp)
                        if rp == port {
                                // If checking for a container (domain == ""), ANY gateway rule on this port is a conflict.
                                // Because a container binds the port on all interfaces/domains.
                                // If checking for a gateway rule (domain != ""), conflict only if the SAME domain uses it.
                                if domain == "" || r.Domain == domain {
                                        return true, "中继规则: " + r.Domain
                                }
                        }
                }
        }

        return false, ""
}

func getContainerIP(dockerID string) string {
	info, err := DockerClient.ContainerInspect(context.Background(), dockerID)
	if err != nil {
		return ""
	}
	ip := info.NetworkSettings.IPAddress
	if ip == "" {
		for _, nw := range info.NetworkSettings.Networks {
			ip = nw.IPAddress
			break
		}
	}
	return ip
}

type dynamicProxyHandler struct {
	Port string
}

func (h *dynamicProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        host := r.Host
        domain := host
        if strings.Contains(host, ":") {
                domain, _, _ = strings.Cut(host, ":")
        }

        // The key is domain:port
        key := domain + ":" + h.Port

        // 1. Check Gateway Rules (Domain matching)
        proxyMutex.RLock()
        pool, exists := proxyTargetPools[key]
        proxyMutex.RUnlock()

	if exists {
		if pool == nil || len(pool.Targets) == 0 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(getNiceErrorPage("502", r.Host)))
			return
		}

		// 2. Select target
		targetStr := pool.Next()
		if targetStr != "" {
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
							w.Header().Set("Content-Type", "text/html; charset=utf-8")
							w.WriteHeader(http.StatusBadGateway)
							w.Write([]byte(getNiceErrorPage("502", r.Host)))
							return
						}
					} else {
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
						w.WriteHeader(http.StatusBadGateway)
						w.Write([]byte(getNiceErrorPage("502", r.Host)))
						return
					}
				}
			}

			if !strings.HasPrefix(targetStr, "http://") && !strings.HasPrefix(targetStr, "https://") {
				targetStr = "http://" + targetStr
			}
			
			targetURL, _ := url.Parse(targetStr)
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			
			proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
				rw.Header().Set("Content-Type", "text/html; charset=utf-8")
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(getNiceErrorPage("502", req.Host)))
			}
			
			proxy.ServeHTTP(w, r)
			return
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(getNiceErrorPage("502", r.Host)))
			return
		}
	}

	// 2. Check Container Port Mappings (Fallback Port-based matching)
	var containers []models.Container
	db.DB.Find(&containers)

	for _, c := range containers {
		if c.DockerID == "" || c.Ports == "" || c.Ports == "[]" {
			continue
		}
		// Skip containers that are not running
		info, err := DockerClient.ContainerInspect(context.Background(), c.DockerID)
		if err != nil || !info.State.Running {
			continue
		}
		var ports []WorkloadPort
		json.Unmarshal([]byte(c.Ports), &ports)
		for _, p := range ports {
			if p.Host == h.Port {
				// Match! Route to this container's IP
				ip := getContainerIP(c.DockerID)
				if ip != "" {
					containerPort := p.Container
					if containerPort == "" {
						containerPort = "80"
					}
					targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", ip, containerPort))
					proxy := httputil.NewSingleHostReverseProxy(targetURL)

					proxy.Director = func(req *http.Request) {
						req.URL.Scheme = targetURL.Scheme
						req.URL.Host = targetURL.Host
						req.Header.Set("X-Forwarded-Host", req.Host)
					}

					proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
						rw.Header().Set("Content-Type", "text/html; charset=utf-8")
						rw.WriteHeader(http.StatusBadGateway)
						rw.Write([]byte(getNiceErrorPage("502", req.Host)))
					}

					proxy.ServeHTTP(w, r)
					return
				}
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(getNiceErrorPage("404", r.Host)))
}

// EnsureListenerRunning starts a shared HTTP multiplexer on the specified port if it doesn't exist
func EnsureListenerRunning(port string) {
	workloadProxyMutex.Lock()
	defer workloadProxyMutex.Unlock()

	if _, exists := proxyServers[port]; exists {
		return // Already running
	}

	handler := &dynamicProxyHandler{Port: port}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	proxyServers[port] = server

	go func() {
		log.Printf("[Dynamic Listener] Starting unified proxy listener on port %s", port)
		var err error
		
		// If the user requests 443, we wrap it with TLS
		if port == "443" {
			server.TLSConfig = GetTLSConfig()
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("[Dynamic Listener Error] Port %s failed: %v", port, err)
		}
		
		// Once it exits (due to error or Close), remove from map
		workloadProxyMutex.Lock()
		delete(proxyServers, port)
		workloadProxyMutex.Unlock()
	}()
}

// Keeping this for backward compatibility if ever needed
func StopSingleProxy(hostPort string) {
	workloadProxyMutex.Lock()
	server, exists := proxyServers[hostPort]
	workloadProxyMutex.Unlock()

	if exists {
		log.Printf("[Dynamic Listener] Stopping listener on port %s", hostPort)
		server.Close()
		// Map deletion is handled in the listener's goroutine when Close() returns
	}
}

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

// StopProxiesForContainer - we no longer actively stop listeners since they are multiplexed and shared
func StopProxiesForContainer(c models.Container) {
	// Optional: Could implement reference counting to stop listeners when 0 users, but not strictly necessary
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
		targetStr := pool.Next()
		if targetStr != "" {
			if !strings.HasPrefix(targetStr, "http://") && !strings.HasPrefix(targetStr, "https://") {
				targetStr = "http://" + targetStr
			}
			targetURL, _ := url.Parse(targetStr)
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			
			// Let the container know the original host
			proxy.Director = func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.Header.Set("X-Forwarded-Host", req.Host)
			}
			
			proxy.ServeHTTP(w, r)
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
		var ports []WorkloadPort
		json.Unmarshal([]byte(c.Ports), &ports)
		for _, p := range ports {
			if p.Host == h.Port {
				// Match! Route to this container's IP
				ip := getContainerIP(c.DockerID)
				if ip != "" {
					targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", ip, p.Container))
					proxy := httputil.NewSingleHostReverseProxy(targetURL)
					
					proxy.Director = func(req *http.Request) {
						req.URL.Scheme = targetURL.Scheme
						req.URL.Host = targetURL.Host
						req.Header.Set("X-Forwarded-Host", req.Host)
					}
					
					proxy.ServeHTTP(w, r)
					return
				}
			}
		}
	}

	http.Error(w, "Godelion Proxy: No matching gateway rule or container port mapping found for this port/host", http.StatusNotFound)
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
			workloadProxyMutex.Lock()
			delete(proxyServers, port)
			workloadProxyMutex.Unlock()
		}
	}()
}

// Keeping this for backward compatibility if ever needed
func StopSingleProxy(hostPort string) {
	workloadProxyMutex.Lock()
	defer workloadProxyMutex.Unlock()

	if server, exists := proxyServers[hostPort]; exists {
		log.Printf("[Dynamic Listener] Stopping listener on port %s", hostPort)
		server.Close()
		delete(proxyServers, hostPort)
	}
}

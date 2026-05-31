package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"godelion/controllers"
	"godelion/db"
	"godelion/middleware"
	"godelion/models"
	"godelion/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func main() {
	// Initialize Subsystems
	db.InitDB()
	if err := services.InitDocker(); err != nil {
		log.Printf("Warning: Docker init failed: %v. Core features may be disabled.", err)
	}
	services.InitProxy()
	services.LoadAndStartAllProxies()
	middleware.LoadAccessConfig()

	// Main restart loop
	for {
		// Load config from database
		var config models.SystemConfig
		if err := db.DB.First(&config).Error; err != nil {
			config.Port = 9960
		}
		if config.Port <= 0 {
			config.Port = 9960
		}
		services.SystemPort = fmt.Sprintf("%d", config.Port)
		services.EnableHTTPS = config.EnableHTTPS

		// Load SSL certificate if HTTPS enabled
		if config.EnableHTTPS && config.PanelSSLID != "" {
			var cert models.SSLCertificate
			if err := db.DB.First(&cert, "id = ?", config.PanelSSLID).Error; err == nil {
				services.CertContent = cert.CertContent
				services.KeyContent = cert.KeyContent
			} else {
				log.Printf("[Warning] SSL certificate not found: %s, falling back to HTTP", config.PanelSSLID)
				services.EnableHTTPS = false
			}
		} else {
			services.CertContent = ""
			services.KeyContent = ""
		}

		app := buildApp()

		// Listen for restart signal in background
		restartDone := make(chan struct{})
		go func() {
			<-services.RestartChan
			log.Println("[Restart] Config changed, restarting server...")
			if err := app.Shutdown(); err != nil {
				log.Printf("[Restart] Shutdown error: %v", err)
			}
			close(restartDone)
		}()

		// Start server based on HTTPS config
		if services.EnableHTTPS && services.CertContent != "" && services.KeyContent != "" {
			log.Printf("Starting HTTPS panel on port %s", services.SystemPort)
			cert, err := tls.X509KeyPair([]byte(services.CertContent), []byte(services.KeyContent))
			if err != nil {
				log.Printf("[Error] Failed to load certificate: %v, falling back to HTTP", err)
				if err := app.Listen(":" + services.SystemPort); err != nil {
					log.Printf("Server stopped: %v", err)
				}
			} else {
				ln, err := tls.Listen("tcp", ":"+services.SystemPort, &tls.Config{
					Certificates: []tls.Certificate{cert},
				})
				if err != nil {
					log.Printf("[Error] Failed to start TLS listener: %v", err)
				} else {
					if err := app.Listener(ln); err != nil {
						log.Printf("Server stopped: %v", err)
					}
				}
			}
		} else {
			log.Printf("Starting HTTP panel on port %s", services.SystemPort)
			if err := app.Listen(":" + services.SystemPort); err != nil {
				log.Printf("Server stopped: %v", err)
			}
		}

		// Wait for restart signal to complete, then loop
		<-restartDone
		log.Println("[Restart] Server shutdown complete, restarting...")
	}
}

func buildApp() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             100 * 1024 * 1024, // 100MB limit for file uploads
	})

	app.Use(cors.New())
	app.Use(logger.New())

	// Access control (IP whitelist + domain binding) - blocks everything if denied
	app.Use(middleware.AccessControl())

	// Secure entrypoint check (must be before static files)
	app.Use(middleware.SecureEntrypointCheck())

	// Proxy handler runs before everything else
	app.Use(services.ProxyHandler)

	// Admin / System API routes
	api := app.Group("/sys/v1")

	// Auth (no access control for login)
	api.Post("/auth/login", controllers.Login)
	api.Post("/auth/verify-2fa", controllers.VerifyLogin2FA)

	// Protected routes
	protected := api.Use(middleware.AuthRequired())

	// User Profile
	user := protected.Group("/user")
	user.Get("/profile", controllers.GetProfile)
	user.Put("/profile", controllers.UpdateProfile)
	user.Post("/change-username", controllers.ChangeUsername)
	user.Post("/change-password", controllers.ChangePassword)
	user.Get("/passkeys", controllers.GetPasskeys)
	user.Post("/passkeys", controllers.CreatePasskey)
	user.Delete("/passkeys/:id", controllers.DeletePasskey)

	// System Config
	config := protected.Group("/config")
	config.Get("/", controllers.GetSystemConfig)
	config.Put("/", controllers.UpdateSystemConfig)

	// 2FA
	twofa := protected.Group("/2fa")
	twofa.Get("/status", controllers.Get2FAStatus)
	twofa.Post("/generate", controllers.Generate2FASecret)
	twofa.Post("/verify", controllers.Verify2FA)
	twofa.Post("/disable", controllers.Disable2FA)

	// Workloads
	workloads := protected.Group("/workloads")
	workloads.Get("/", controllers.ListWorkloads)
	workloads.Post("/", controllers.CreateWorkload)
	workloads.Post("/:id/start", controllers.StartWorkload)
	workloads.Post("/:id/stop", controllers.StopWorkload)
	workloads.Get("/:id/logs", controllers.GetWorkloadLogs)
	workloads.Delete("/:id", controllers.DeleteWorkload)
	workloads.Put("/:id", controllers.UpdateWorkload)

	// SSL Certificates
	ssl := protected.Group("/ssl")
	ssl.Get("/", controllers.ListSSLCerts)
	ssl.Post("/", controllers.CreateSSLCert)
	ssl.Delete("/:id", controllers.DeleteSSLCert)

	// Gateways
	gateways := protected.Group("/gateways")
	gateways.Post("/", controllers.CreateGatewayRule)
	gateways.Get("/", controllers.ListGatewayRules)
	gateways.Put("/:id", controllers.UpdateGatewayRule)
	gateways.Delete("/:id", controllers.DeleteGatewayRule)

	// Audit Logs
	audit := protected.Group("/audit")
	audit.Get("/", controllers.ListAuditLogs)

	// Storage
	storage := protected.Group("/storage")
	storage.Post("/upload", controllers.UploadFile)
	storage.Post("/folder", controllers.CreateFolder)
	storage.Post("/move", controllers.MoveFile)
	storage.Post("/extract", controllers.ExtractArchive)
	storage.Get("/list", controllers.ListFiles)
	storage.Get("/read", controllers.ReadFileContent)
	storage.Post("/save", controllers.SaveFileContent)
	storage.Delete("/delete", controllers.DeleteFile)
	storage.Get("/download", controllers.DownloadFile)

	// System
	system := protected.Group("/system")
	system.Get("/docker/status", controllers.GetDockerStatus)
	system.Post("/docker/install", controllers.InstallDocker)
	system.Post("/docker/start", controllers.StartDocker)
	system.Post("/docker/stop", controllers.StopDocker)
	system.Post("/docker/restart", controllers.RestartDocker)
	system.Get("/docker/config", controllers.GetDockerConfig)
	system.Post("/docker/config", controllers.UpdateDockerConfig)
	system.Get("/health", controllers.GetSystemHealth)

	// Serve frontend static files (MUST be after all API routes)
	publicDir := filepath.Join(".", "godelion_public")
	if info, err := os.Stat(publicDir); err == nil && info.IsDir() {
		fs := http.Dir(publicDir)
		fileServer := http.FileServer(fs)

		app.Get("/*", func(c *fiber.Ctx) error {
			path := c.Path()
			if f, err := fs.Open(path); err == nil {
				f.Close()
				handler := fasthttpadaptor.NewFastHTTPHandler(fileServer)
				handler(c.Context())
				return nil
			}
			handler := fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, filepath.Join(publicDir, "index.html"))
			}))
			handler(c.Context())
			return nil
		})
	}

	return app
}

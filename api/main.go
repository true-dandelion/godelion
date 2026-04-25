package main

import (
	"crypto/tls"
	"log"

	"godelion/controllers"
	"godelion/db"
	"godelion/middleware"
	"godelion/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Initialize Subsystems
	db.InitDB()
	if err := services.InitDocker(); err != nil {
		log.Printf("Warning: Docker init failed: %v. Core features may be disabled.", err)
	}
	services.InitProxy()
	services.LoadAndStartAllProxies()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(cors.New())
	app.Use(logger.New())

	// Proxy handler runs before everything else. It checks the host header.
	// If a matching reverse proxy is found, it handles the request and returns.
	app.Use(services.ProxyHandler)

	// Admin / System API routes
	api := app.Group("/sys/v1")

	// Auth
	api.Post("/auth/login", controllers.Login)

	// Protected routes
	protected := api.Use(middleware.AuthRequired())

	// Workloads
	workloads := protected.Group("/workloads")
	workloads.Get("/", controllers.ListWorkloads)
	workloads.Post("/", controllers.CreateWorkload)
	workloads.Post("/:id/start", controllers.StartWorkload)
	workloads.Post("/:id/stop", controllers.StopWorkload)
	workloads.Get("/:id/logs", controllers.GetWorkloadLogs)

	// Gateways
	gateways := protected.Group("/gateways")
	gateways.Post("/", controllers.CreateGatewayRule)
	gateways.Get("/", controllers.ListGatewayRules)
	gateways.Delete("/:id", controllers.DeleteGatewayRule)

	// Storage
	storage := protected.Group("/storage")
	storage.Post("/upload", controllers.UploadFile)
	storage.Post("/folder", controllers.CreateFolder)
	storage.Post("/move", controllers.MoveFile)
	storage.Post("/extract", controllers.ExtractArchive)
	storage.Get("/list", controllers.ListFiles)
	storage.Get("/read", controllers.ReadFileContent)
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

	// TLS config for dynamic certificates
	tlsConfig := services.GetTLSConfig()
	ln, err := tls.Listen("tcp", ":443", tlsConfig)
	if err == nil {
		go func() {
			log.Println("Starting HTTPS server on :443")
			app.Listener(ln)
		}()
	} else {
		log.Printf("Could not start HTTPS listener: %v\n", err)
	}

	log.Println("Starting HTTP server on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

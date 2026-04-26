package controllers

	import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"godelion/db"
	"godelion/models"
	"godelion/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create a helper function to get user dir inside workload_controller too
func getWorkloadUserDir(userID string) string {
	cwd, _ := os.Getwd()
	if filepath.Base(cwd) == "api" {
		cwd = filepath.Dir(cwd)
	}
	return filepath.Join(cwd, "godelion", "user", userID)
}

func CreateWorkload(c *fiber.Ctx) error {
	var req struct {
		Name           string `json:"name"`
		NodeVersion    string `json:"node_version"`
		ProjectDir     string `json:"project_dir"`
		StartCommand   string `json:"start_command"`
		ContainerName  string `json:"container_name"`
		PackageManager string `json:"package_manager"`
		Dependencies   string `json:"dependencies"`
		Ports          []struct {
			Host      string `json:"host"`
			Container string `json:"container"`
		} `json:"ports"`
		ResourceLimits string `json:"resource_limits"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	userID := c.Locals("user_id").(string)

	// Resolve the requested node version
	baseImage := req.NodeVersion
	if baseImage == "" {
		baseImage = "node:24-alpine"
	}

	// We use standard image names directly (e.g. node:24-alpine).
	// It's recommended to configure registry-mirrors in /etc/docker/daemon.json if pulling fails.
	imageName := baseImage
	
	// Pre-generate an ID for the container so we can save it to DB immediately
	containerID := uuid.NewString()

	// Convert ports to JSON string
        for _, p := range req.Ports {
                if p.Host != "" {
                        isConflict, reason := services.CheckPortConflict(p.Host, "", "", "")
                        if isConflict {
                                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "主机端口 " + p.Host + " 已被 [" + reason + "] 占用"})
                        }
                }
        }
        portsJSON, _ := json.Marshal(req.Ports)

	// Save to DB immediately with status 'creating' (handled by the UI as error/stopped initially until Docker catches up)
	dbContainer := models.Container{
		ID:             containerID,
		DockerID:       "", // Not created yet
		Name:           req.Name,
		Image:          imageName + " (Source: " + req.ProjectDir + ")",
		UserID:         userID,
		Ports:          string(portsJSON),
		ResourceLimits: req.ResourceLimits,
	}

	if err := db.DB.Create(&dbContainer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save to db"})
	}

	LogAction(c, "Deploy", "Container", "Deployed container: "+req.Name)

	// Run actual container pulling and creation asynchronously in a goroutine
	go func() {
		// Helper to append logs to DB
		appendLog := func(msg string) {
			log.Printf(msg)
			// Actually let's use a simpler format for DB:
			logLine := fmt.Sprintf("%s\n", msg)
			db.DB.Model(&models.Container{}).Where("id = ?", containerID).
				Update("deployment_logs", gorm.Expr("IFNULL(deployment_logs, '') || ?", logLine))
		}

		appendLog(fmt.Sprintf("[Workload Async] Starting deployment for project '%s' (UUID: %s)", req.Name, containerID))
		
		// Use a background context since the request context will be cancelled when response returns
		ctx := context.Background()

		err := services.PullImage(ctx, imageName)
		if err != nil {
			appendLog(fmt.Sprintf("[Workload Async Error] Failed to pull image for '%s': %v", req.Name, err))
			return
		}
		appendLog(fmt.Sprintf("[Workload Async] Successfully pulled image '%s'", imageName))

		var ports []services.PortMapping
		for _, p := range req.Ports {
			ports = append(ports, services.PortMapping{
				HostPort:      p.Host,
				ContainerPort: p.Container,
			})
		}

		// Mount the selected project directory to /app
		userDir := getWorkloadUserDir(userID)
		cleanVirtualPath := filepath.Clean(filepath.Join("/", req.ProjectDir))
		physicalHostPath := filepath.Join(userDir, cleanVirtualPath)

		volumes := []services.VolumeMapping{
			{
				HostPath:      physicalHostPath,
				ContainerPath: "/app",
			},
		}

		// Command to install dependencies and run start command
		// If the command starts with "node ", we check if manual dependencies are provided
		cmdStr := ""
		if len(req.StartCommand) >= 5 && req.StartCommand[:5] == "node " {
			if req.Dependencies != "" {
				// Replace commas with spaces for npm install
				deps := strings.ReplaceAll(req.Dependencies, ",", " ")
				cmdStr = "npm install " + deps + " && " + req.StartCommand
			} else {
				cmdStr = req.StartCommand
			}
		} else {
			// For standard npm/yarn/pnpm scripts, we install dependencies first then run the script
			cmdStr = req.PackageManager + " install && " + req.PackageManager + " run " + req.StartCommand
		}
		
		containerCmd := []string{"sh", "-c", cmdStr}

		// Actual Docker container creation
		realContainerID, err := services.CreateContainer(ctx, req.ContainerName, imageName, ports, volumes, containerCmd, "/app")
		if err != nil {
			appendLog(fmt.Sprintf("[Workload Async Error] Failed to create container for '%s': %v", req.Name, err))
			return
		}

		// Update the database record with the real Docker container ID so future start/stop commands work
		db.DB.Model(&models.Container{}).Where("id = ?", containerID).Update("docker_id", realContainerID)
		appendLog(fmt.Sprintf("[Workload Async] Successfully mapped UUID %s to Docker ID %s", containerID, realContainerID))

		// Try to start the container immediately
		err = services.StartContainer(ctx, realContainerID)
		if err != nil {
			appendLog(fmt.Sprintf("[Workload Async Error] Failed to start container %s automatically: %v", realContainerID, err))
		} else {
			appendLog(fmt.Sprintf("[Workload Async] Deployment finished successfully for '%s'", req.Name))
			
			// Start the proxy listeners for this container
			var updatedContainer models.Container
			db.DB.First(&updatedContainer, "id = ?", containerID)
			services.StartProxiesForContainer(updatedContainer)
		}
	}()

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Container deployment started in background",
		"data":    dbContainer,
	})
}

func ListWorkloads(c *fiber.Ctx) error {
	var workloads []models.Container
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)

	query := db.DB
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&workloads).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch workloads"})
	}

	for i, w := range workloads {
		// If DockerID is empty, it means the async task hasn't finished replacing it with the Docker container ID
		if w.DockerID == "" {
			workloads[i].Status = "creating"
			continue
		}

		info, err := services.InspectContainer(c.Context(), w.DockerID)
		if err == nil {
			workloads[i].Status = info.State.Status
		} else {
			workloads[i].Status = "error"
		}
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    workloads,
	})
}

func StartWorkload(c *fiber.Ctx) error {
        id := c.Params("id")
        var w models.Container
        if err := db.DB.First(&w, "id = ?", id).Error; err != nil {
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Container not found"})
        }
        if w.DockerID == "" {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Container is still creating"})
        }

        if err := services.StartContainer(c.Context(), w.DockerID); err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
        }

        // Add a start log entry to action logs
        logLine := "=== 服务启动 ===\n"
        db.DB.Model(&models.Container{}).Where("id = ?", w.ID).
                Update("action_logs", gorm.Expr("IFNULL(action_logs, '') || ?", logLine))

        // Start proxy
        services.StartProxiesForContainer(w)

        LogAction(c, "Start", "Container", "Started container: "+w.Name)

        return c.JSON(fiber.Map{"code": 200, "message": "Started"})
}

func StopWorkload(c *fiber.Ctx) error {
        id := c.Params("id")
        var w models.Container
        if err := db.DB.First(&w, "id = ?", id).Error; err != nil {
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Container not found"})
        }
        if w.DockerID == "" {
                return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Container is still creating"})
        }

        if err := services.StopContainer(c.Context(), w.DockerID); err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
        }

        // Add a stop log entry to action logs
        logLine := "\n=== 服务停止 ===\n"
        db.DB.Model(&models.Container{}).Where("id = ?", w.ID).
                Update("action_logs", gorm.Expr("IFNULL(action_logs, '') || ?", logLine))

        // Stop proxy
        services.StopProxiesForContainer(w)

        LogAction(c, "Stop", "Container", "Stopped container: "+w.Name)

        return c.JSON(fiber.Map{"code": 200, "message": "Stopped"})
}

func GetWorkloadLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	var w models.Container
	if err := db.DB.First(&w, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Container not found"})
	}
	
	// If DockerID is empty, Docker container doesn't exist yet (still deploying or failed)
	// Return the async deployment logs from the database
	if w.DockerID == "" {
		logs := w.DeploymentLogs
		if logs == "" {
			logs = "部署进行中，正在拉取镜像或创建容器，请稍后再查看日志..."
		}
		return c.JSON(fiber.Map{
			"code":    200,
			"message": "Success",
			"data":    logs,
		})
	}

	logs, err := services.GetContainerLogs(c.Context(), w.DockerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	
	// If it successfully created, we can prepend the deployment logs to the actual Docker logs
	// so the user sees the full history.
	fullLogs := ""
	if w.DeploymentLogs != "" {
		fullLogs += "=== 部署阶段日志 ===\n" + w.DeploymentLogs + "\n"
	}
	
	fullLogs += "=== 运行阶段日志 ===\n"
	fullLogs += logs

	if w.ActionLogs != "" {
		fullLogs += w.ActionLogs
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    fullLogs,
	})
}

func DeleteWorkload(c *fiber.Ctx) error {
	id := c.Params("id")
	var w models.Container
	if err := db.DB.First(&w, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Container not found"})
	}

	if w.DockerID != "" {
		services.StopContainer(context.Background(), w.DockerID)
		services.RemoveContainer(context.Background(), w.DockerID)
	}

	if err := db.DB.Delete(&w).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete db record"})
	}

	// Stop proxy before deleting
	services.StopProxiesForContainer(w)

	LogAction(c, "Delete", "Container", "Deleted container: "+w.Name)

	return c.JSON(fiber.Map{"code": 200, "message": "Successfully deleted"})
}

func UpdateWorkload(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name  string `json:"name"`
		Ports []struct {
			Host      string `json:"host"`
			Container string `json:"container"`
		} `json:"ports"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	var w models.Container
	if err := db.DB.First(&w, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Container not found"})
	}

	if req.Name != "" {
		w.Name = req.Name
	}
	if req.Ports != nil {
		for _, p := range req.Ports {
			if p.Host != "" {
				isConflict, reason := services.CheckPortConflict(p.Host, "", "", id)
				if isConflict {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "主机端口 " + p.Host + " 已被 [" + reason + "] 占用"})
				}
			}
		}
		portsJSON, _ := json.Marshal(req.Ports)
		w.Ports = string(portsJSON)
	}

	if err := db.DB.Save(&w).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update db record"})
	}

	services.StartProxiesForContainer(w)

	LogAction(c, "Update", "Container", "Updated container config: "+w.Name)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Successfully updated",
		"data":    w,
	})
}

package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
)

func GetDockerStatus(c *fiber.Ctx) error {
	_, err := exec.LookPath("docker")
	installed := err == nil

	running := false
	if installed {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err == nil {
			_, pingErr := cli.Ping(context.Background())
			running = pingErr == nil
			cli.Close()
		}
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data": fiber.Map{
			"installed": installed,
			"running":   running,
		},
	})
}

func InstallDocker(c *fiber.Ctx) error {
	// 使用国内镜像源安装 Docker（例如阿里云镜像或清华大学镜像）
	// get.docker.com 在国内可能会被墙或者连接重置
	cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com | sh -s docker --mirror Aliyun")
	
	// Start the command and wait for it to finish
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Docker 安装失败",
			"error":   string(output),
		})
	}

	// Try to start the docker service just in case it didn't start automatically
	exec.Command("systemctl", "start", "docker").Run()
	// Optionally enable it on boot
	exec.Command("systemctl", "enable", "docker").Run()

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Docker 安装并启动成功",
	})
}

func StartDocker(c *fiber.Ctx) error {
	cmd := exec.Command("systemctl", "start", "docker")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "启动 Docker 失败",
			"error":   string(output),
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Docker 启动成功",
	})
}

func StopDocker(c *fiber.Ctx) error {
	cmd := exec.Command("systemctl", "stop", "docker")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "停止 Docker 失败",
			"error":   string(output),
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Docker 停止成功",
	})
}

func RestartDocker(c *fiber.Ctx) error {
	cmd := exec.Command("systemctl", "restart", "docker")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "重启 Docker 失败",
			"error":   string(output),
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Docker 重启成功",
	})
}

func GetDockerConfig(c *fiber.Ctx) error {
	configPath := "/etc/docker/daemon.json"
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return c.JSON(fiber.Map{
			"code":    200,
			"message": "Success",
			"data":    "{\n}", // Return empty JSON object if file doesn't exist
		})
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "读取 daemon.json 失败",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    string(content),
	})
}

func UpdateDockerConfig(c *fiber.Ctx) error {
	var req struct {
		Config string `json:"config"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate JSON format
	if !json.Valid([]byte(req.Config)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "无效的 JSON 格式",
		})
	}

	configPath := "/etc/docker/daemon.json"
	dir := filepath.Dir(configPath)
	
	// Create directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	if err := ioutil.WriteFile(configPath, []byte(req.Config), 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "保存 daemon.json 失败",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Docker 配置已更新",
	})
}

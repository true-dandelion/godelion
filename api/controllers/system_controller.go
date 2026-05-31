package controllers

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"godelion/db"
	"godelion/models"

	"github.com/docker/docker/api/types/container"
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
	downloadURL := "https://sh.shaoxin.top:1111/shaoxin/Dnld/docker-29.5.0.tgz"
	tmpDir := "/tmp/docker-install"
	tgzPath := tmpDir + "/docker-29.5.0.tgz"

	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Step 1: Download
	logText := "[1/5] 下载 Docker 二进制包...\n"
	dlCmd := exec.Command("curl", "-fSL", "--progress-bar", "-o", tgzPath, downloadURL)
	dlOut, err := dlCmd.CombinedOutput()
	logText += string(dlOut)
	if err != nil {
		return c.JSON(fiber.Map{"code": 500, "message": "下载失败", "data": fiber.Map{"log": logText}})
	}
	logText += "下载完成\n"

	// Step 2: Extract
	logText += "\n[2/5] 解压安装包...\n"
	extCmd := exec.Command("tar", "xzf", tgzPath, "-C", tmpDir)
	extOut, err := extCmd.CombinedOutput()
	logText += string(extOut)
	if err != nil {
		return c.JSON(fiber.Map{"code": 500, "message": "解压失败", "data": fiber.Map{"log": logText}})
	}
	logText += "解压完成\n"

	// Step 3: Copy binaries
	logText += "\n[3/5] 安装二进制文件到 /usr/bin/...\n"
	bins := []string{"docker", "dockerd", "containerd", "containerd-shim", "containerd-shim-runc-v2", "ctr", "runc", "docker-init", "docker-proxy"}
	for _, bin := range bins {
		src := tmpDir + "/docker/" + bin
		if _, err := os.Stat(src); err == nil {
			out, err := exec.Command("cp", "-f", src, "/usr/bin/"+bin).CombinedOutput()
			if err != nil {
				logText += "安装 " + bin + " 失败: " + string(out) + "\n"
			} else {
				logText += "  ✓ " + bin + "\n"
			}
		}
	}

	// Step 4: Create systemd services
	logText += "\n[4/5] 创建 systemd 服务...\n"

	containerdContent := `[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/bin/containerd
Type=notify
Delegate=yes
KillMode=process
Restart=always
RestartSec=5
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity

[Install]
WantedBy=multi-user.target
`

	serviceContent := `[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target firewalld.service containerd.service
Wants=network-online.target
Requires=docker.socket containerd.service

[Service]
Type=notify
ExecStart=/usr/bin/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutStartSec=0
RestartSec=2
Restart=always
StartLimitBurst=3
StartLimitInterval=60s
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
Delegate=yes
KillMode=process
OOMScoreAdjust=-500

[Install]
WantedBy=multi-user.target
`
	socketContent := `[Unit]
Description=Docker Socket for the API
PartOf=docker.service

[Socket]
ListenStream=/var/run/docker.sock
SocketMode=0660
SocketUser=root
SocketGroup=docker

[Install]
WantedBy=sockets.target
`
	// Step 4.5: Create docker group
	logText += "  创建 docker 用户组...\n"
	groupOut, _ := exec.Command("groupadd", "-f", "docker").CombinedOutput()
	logText += string(groupOut)

	os.WriteFile("/etc/systemd/system/containerd.service", []byte(containerdContent), 0644)
	os.WriteFile("/etc/systemd/system/docker.service", []byte(serviceContent), 0644)
	os.WriteFile("/etc/systemd/system/docker.socket", []byte(socketContent), 0644)
	logText += "  ✓ containerd.service\n"
	logText += "  ✓ docker.service\n"
	logText += "  ✓ docker.socket\n"

	// Step 5: Reload and start
	logText += "\n[5/5] 启动 Docker 服务...\n"
	reloadOut, _ := exec.Command("systemctl", "daemon-reload").CombinedOutput()
	logText += string(reloadOut)
	startOut, _ := exec.Command("systemctl", "start", "containerd").CombinedOutput()
	logText += string(startOut)
	startOut, _ = exec.Command("systemctl", "start", "docker").CombinedOutput()
	logText += string(startOut)
	enableOut, _ := exec.Command("systemctl", "enable", "containerd").CombinedOutput()
	logText += string(enableOut)
	enableOut, _ = exec.Command("systemctl", "enable", "docker").CombinedOutput()
	logText += string(enableOut)

	// Verify
	logText += "\n验证安装...\n"
	verifyOut, _ := exec.Command("docker", "version").CombinedOutput()
	logText += string(verifyOut)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Docker 安装并启动成功",
		"data":    fiber.Map{"log": logText},
	})
}

func StartDocker(c *fiber.Ctx) error {
	// Start docker service first, then socket
	cmd := exec.Command("systemctl", "start", "docker")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "启动 Docker 失败",
			"error":   string(output),
		})
	}
	
	// Start docker.socket after docker service
	exec.Command("systemctl", "start", "docker.socket").Run()
	
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Docker 启动成功",
	})
}

func StopDocker(c *fiber.Ctx) error {
	// Stop docker.socket first to prevent it from restarting docker service
	exec.Command("systemctl", "stop", "docker.socket").Run()
	
	// Then stop docker service
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
	// Stop docker.socket first
	exec.Command("systemctl", "stop", "docker.socket").Run()
	
	// Restart docker service
	cmd := exec.Command("systemctl", "restart", "docker")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "重启 Docker 失败",
			"error":   string(output),
		})
	}
	
	// Start docker.socket after docker service
	exec.Command("systemctl", "start", "docker.socket").Run()
	
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求格式错误"})
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

// GetSystemHealth returns health status of Docker, SSL certs, and gateway rules
func GetSystemHealth(c *fiber.Ctx) error {
	health := fiber.Map{
		"docker":  checkDockerHealth(),
		"ssl":     checkSSLHealth(),
		"gateway": checkGatewayHealth(),
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    health,
	})
}

func checkDockerHealth() fiber.Map {
	status := "offline"
	message := "Docker 未安装"

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fiber.Map{"status": status, "message": message}
	}
	defer cli.Close()

	_, pingErr := cli.Ping(context.Background())
	if pingErr != nil {
		return fiber.Map{"status": "offline", "message": "Docker 未运行"}
	}

	// Count running containers
	containers, _ := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	running := 0
	for _, ct := range containers {
		if ct.State == "running" {
			running++
		}
	}

	return fiber.Map{
		"status":   "online",
		"message":  "运行中",
		"running":  running,
		"total":    len(containers),
	}
}

func checkSSLHealth() fiber.Map {
	var certs []models.SSLCertificate
	db.DB.Find(&certs)

	total := len(certs)
	if total == 0 {
		return fiber.Map{"status": "normal", "message": "暂无证书", "total": 0, "expiring": 0, "expired": 0}
	}

	expiring := 0
	expired := 0
	now := time.Now()

	for _, cert := range certs {
		if cert.CertContent == "" {
			continue
		}

		block, _ := pem.Decode([]byte(cert.CertContent))
		if block == nil {
			continue
		}

		x509Cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		daysLeft := time.Until(x509Cert.NotAfter).Hours() / 24
		if x509Cert.NotAfter.Before(now) {
			expired++
		} else if daysLeft <= 30 {
			expiring++
		}
	}

	status := "normal"
	message := "证书正常"
	if expired > 0 {
		status = "error"
		message = "存在过期证书"
	} else if expiring > 0 {
		status = "warning"
		message = "存在即将过期证书"
	}

	return fiber.Map{
		"status":   status,
		"message":  message,
		"total":    total,
		"expiring": expiring,
		"expired":  expired,
	}
}

func checkGatewayHealth() fiber.Map {
	var rules []models.GatewayRule
	db.DB.Find(&rules)

	total := len(rules)
	if total == 0 {
		return fiber.Map{"status": "normal", "message": "暂无规则", "total": 0, "unreachable": 0}
	}

	unreachable := 0
	for _, rule := range rules {
		if rule.TargetURLs == "" {
			continue
		}

		// Check each target URL
		targets := strings.Split(rule.TargetURLs, ",")
		for _, target := range targets {
			target = strings.TrimSpace(target)
			if target == "" || target[0] == '@' {
				// Dynamic container target, skip direct check
				continue
			}

			// Remove protocol prefix if present
			target = strings.TrimPrefix(target, "http://")
			target = strings.TrimPrefix(target, "https://")

			// Ensure host:port format
			host, port, err := net.SplitHostPort(target)
			if err != nil {
				host = target
				port = "80"
			}

			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 2*time.Second)
			if err != nil {
				unreachable++
			} else {
				conn.Close()
			}
		}
	}

	status := "normal"
	message := "规则正常"
	if unreachable > 0 {
		status = "warning"
		message = "存在不可达目标"
	}

	return fiber.Map{
		"status":      status,
		"message":     message,
		"total":       total,
		"unreachable": unreachable,
	}
}

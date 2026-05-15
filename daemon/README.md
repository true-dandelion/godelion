# Godelion 守护进程使用说明

## 功能概述

守护进程是一个独立的 Go 程序，用于监控 Godelion 主程序的运行状态。当主程序出现任何问题时，守护进程会自动记录日志并尝试重启主程序。

## 主要功能

1. **自动监控**：定期检查主程序是否在运行
2. **自动重启**：当主程序崩溃或异常退出时，自动尝试重启
3. **日志记录**：将所有监控和错误信息记录到日志文件
4. **信号处理**：支持通过信号控制守护进程和主程序
5. **PID 管理**：维护 PID 文件便于管理

## 安装步骤

### 1. 编译和安装

```bash
cd /path/to/godelion/daemon
chmod +x install.sh
sudo ./install.sh
```

### 2. 配置 systemd 服务（可选）

安装脚本会自动创建 systemd 服务文件，也可以手动复制：

```bash
sudo cp godelion-daemon.service /etc/systemd/system/
sudo systemctl daemon-reload
```

## 使用方法

### 方式一：使用 systemd 管理（推荐）

```bash
# 启动守护进程
sudo systemctl start godelion-daemon

# 停止守护进程
sudo systemctl stop godelion-daemon

# 重启守护进程
sudo systemctl restart godelion-daemon

# 查看状态
sudo systemctl status godelion-daemon

# 设置开机自启
sudo systemctl enable godelion-daemon

# 取消开机自启
sudo systemctl disable godelion-daemon
```

### 方式二：直接运行

```bash
cd /root/godelion/bin
./godelion-daemon
```

### 方式三：后台运行

```bash
nohup /root/godelion/bin/godelion-daemon > /dev/null 2>&1 &
```

## 日志管理

### 日志位置

日志文件保存在：`/root/godelion/<日期时间>/logs/`

每次守护进程启动会创建新的日志目录，格式为：`/root/godelion/2026-05-12_15-30-00/logs/`

### 日志文件

- `daemon_YYYY-MM-DD.log`：每日守护进程日志
- `daemon-stdout.log`：systemd stdout 日志
- `daemon-stderr.log`：systemd stderr 日志

### 日志格式

```
[2026-05-12 15:30:00] [INFO] 主程序已启动, PID: 12345
[2026-05-12 15:35:00] [WARN] 检测到主程序 (PID: 12345) 未运行
[2026-05-12 15:35:00] [WARN] 准备重启主程序 (第 1/5 次)
[2026-05-12 15:35:10] [INFO] 主程序已启动, PID: 12346
```

## 信号控制

守护进程支持以下信号：

| 信号 | 功能 |
|------|------|
| SIGINT | 优雅关闭守护进程和主程序 |
| SIGTERM | 优雅关闭守护进程和主程序 |
| SIGUSR1 | 重启主程序 |
| SIGUSR2 | 重新加载配置 |

### 使用示例

```bash
# 重启主程序（不停止守护进程）
kill -USR1 $(cat /root/godelion/daemon.pid)

# 优雅关闭
kill -TERM $(cat /root/godelion/daemon.pid)
```

## 配置说明

可以在 `daemon/main.go` 中修改以下配置：

```go
const (
    MainProgramPath = "/root/godelion/api/main.go"  // 主程序路径
    MainProgramName = "godelion-api"                // 主程序名称
    LogBaseDir = "/root/godelion"                   // 日志基础目录
    CheckInterval = 5 * time.Second                 // 检查间隔
    MaxRestartAttempts = 5                          // 最大重启次数
    RestartDelay = 10 * time.Second                 // 重启延迟
)
```

## 故障排查

### 1. 查看守护进程状态

```bash
sudo systemctl status godelion-daemon
```

### 2. 查看守护进程日志

```bash
# 查看最新日志目录
ls -lt /root/godelion/ | head -1

# 查看日志内容
cat /root/godelion/2026-05-12_15-30-00/logs/daemon_2026-05-12.log
```

### 3. 检查主程序是否在运行

```bash
ps aux | grep godelion-api
```

### 4. 手动启动主程序查看错误

```bash
cd /root/godelion/api
go run main.go
```

## 安全建议

1. **限制权限**：守护进程以 root 用户运行，生产环境建议创建专用用户
2. **日志管理**：定期清理旧日志文件
3. **监控告警**：可以配合监控系统设置告警
4. **防火墙**：确保只开放必要的端口

## 常见问题

### Q: 守护进程启动失败怎么办？
A: 检查日志文件，常见问题包括：
- Go 环境未安装
- 权限不足
- 主程序路径错误

### Q: 主程序反复重启怎么办？
A: 检查主程序日志，可能是配置错误或依赖问题。守护进程最多重启 5 次后会自动退出。

### Q: 如何完全卸载守护进程？
A: 
```bash
sudo systemctl stop godelion-daemon
sudo systemctl disable godelion-daemon
sudo rm /etc/systemd/system/godelion-daemon.service
sudo systemctl daemon-reload
rm -rf /root/godelion/bin/godelion-daemon
```

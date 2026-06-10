# Godelion - Go & Docker Container Management System

[中文](#chinese) | [English](#english)

# Godelion

---

<a id="chinese"></a>
## 中文

> Docker 容器管理面板 + 内置反向代理网关。

**Godelion** 是一个轻量级的自托管 Web 面板，可在同一仪表盘中管理 Docker 容器和配置反向代理路由。前端基于 Vue 3，后端基于 Go（Fiber），使用 SQLite 存储，无需外部数据库依赖。

### 功能

- **容器（工作负载）管理** — 通过 Web UI 部署、启动、停止和删除 Docker 容器。内置 Node.js、Python、Go、PHP、静态站点（Nginx）、C/C++ 及通用二进制等运行环境的预配置默认值。
- **反向代理网关** — 创建基于域名的代理规则，支持轮询负载均衡、TLS 终止、HTTP→HTTPS 重定向及自定义 301/302 重定向。
- **SSL 证书管理** — 上传和管理 PEM 格式的证书与密钥对，用于代理域名及面板本身。
- **文件管理器** — 浏览、上传、下载、编辑、创建文件夹、移动和解压归档文件（zip/tgz），所有操作限定在用户独立存储目录内。
- **Docker 守护进程控制** — 从 tar 包安装 Docker、启动/停止/重启 Docker 守护进程、在 UI 中编辑 `/etc/docker/daemon.json`。
- **认证与安全** — 基于 JWT 的认证，可选 TOTP 双因素认证、通行密钥支持、IP 白名单（CIDR）、域名绑定、安全入口点强制访问、会话管理、密码复杂度与过期策略。
- **审计日志** — 所有管理操作均被记录并可在面板中查看。
- **系统监控** — 仪表盘展示容器数量、网关规则状态、SSL 证书过期时间、Docker 运行状态及系统整体健康度。
- **守护进程（自动重启）** — 配套的守护进程监控 API 进程，崩溃时自动重启（最多尝试 5 次）。

### 技术栈

| 层级 | 技术 |
|---|---|
| 前端 | Vue 3, TypeScript, Element Plus, Tailwind CSS, Pinia, Vite |
| 后端 | Go 1.25, Fiber v2 |
| 数据库 | SQLite（纯 Go 实现，无需 CGO） via GORM |
| 容器 | Docker Engine SDK for Go |
| 认证 | JWT (golang-jwt), bcrypt, TOTP (pquerna/otp) |
| 构建 | pnpm（前端），Go 工具链（后端） |

### 快速开始

#### 环境要求

- Go 1.25+
- pnpm
- Docker（使用容器管理功能时需要）
- Linux（推荐，附带 systemd 单元文件）

#### 开发模式

```bash
# 终端 1：启动 Go API 服务
cd api
go run main.go

# 终端 2：启动 Vue 开发服务器
pnpm install
pnpm run dev
```

或使用提供的脚本：

```bash
./start-godelion.sh
```

#### 生产构建

```bash
# 构建前端
pnpm run build

# 将 dist/ 目录复制到 api/godelion_public/
# API 会自动服务 godelion_public/ 下的静态文件

# 构建 Go 后端
cd api
go build -o godelion main.go

# 运行
./godelion
```

#### systemd 安装

```bash
sudo cp godelion.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now godelion
```

#### 守护进程（可选）

`daemon/` 目录下提供了一个看门狗守护进程，用于监控并在主 API 崩溃时自动重启。详见 `daemon/README.md`。

```bash
# 构建
cd daemon
go build -o godelion-daemon main.go

# 安装服务
sudo cp godelion-daemon.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now godelion-daemon
```

### 默认凭据

| 用户名 | 密码 |
|---|---|
| `admin` | `admin123` |

**首次登录后请立即修改。**

### 配置

面板运行时配置通过 Web UI（**Settings → Panel Config**）完成。关键配置项：

- **Port** — 面板监听端口（默认：`9960`）
- **HTTPS** — 为面板启用 HTTPS（需要 SSL 证书）
- **Panel SSL Certificate** — 选择已上传的证书用于面板 HTTPS

配置变更即时生效，无需重启系统级服务。

### 项目结构

```
godelion/
├── api/                    # Go 后端
│   ├── main.go             # 入口，Fiber 应用，路由
│   ├── controllers/        # HTTP 处理器（auth, workload, gateway, ssl, config, system, storage, audit）
│   ├── services/           # 业务逻辑（docker, proxy, port proxy）
│   ├── middleware/          # 认证、IP 白名单、域名绑定、安全入口
│   ├── models/             # GORM 模型
│   ├── db/                 # SQLite 初始化与迁移
│   └── session/            # d_delion_id 会话存储
├── src/                    # Vue 3 前端
│   ├── views/              # 页面组件
│   ├── api/                # Axios API 客户端
│   ├── router/             # 路由配置
│   ├── store/              # Pinia 状态管理
│   └── components/         # 通用组件
├── daemon/                 # 看门狗守护进程（纯 Go 标准库）
├── public/                 # 静态资源
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── godelion.service        # systemd 单元文件
```

### 安全建议

- 立即修改默认 `admin` 密码
- 生产环境务必启用 HTTPS
- 使用 IP 白名单限制可访问的信任网络
- 配置安全入口点增加额外访问控制层
- 定期轮换 SSL 证书和管理员凭据

---

[中文](#chinese) | [English](#english)

<a id="english"></a>
## English

> Docker container management panel with built-in reverse proxy gateway.

**Godelion** is a lightweight self-hosted web panel for managing Docker containers and configuring reverse proxy routes — all from a single dashboard. It comes with a Vue 3 frontend and a Go (Fiber) backend, with SQLite storage and zero external database dependencies.

### Features

- **Container (Workload) Management** — Deploy, start, stop, and delete Docker containers via a web UI. Supports Node.js, Python, Go, PHP, static sites (Nginx), C/C++, and general binaries with pre-configured defaults.
- **Reverse Proxy Gateway** — Create domain-based proxy rules with round-robin load balancing, TLS termination, HTTP→HTTPS redirects, and custom 301/302 redirects.
- **SSL Certificate Management** — Upload and manage PEM certificate+key pairs for proxied domains and the panel itself.
- **File Manager** — Browse, upload, download, edit, create folders, move, and extract archives (zip/tgz) in per-user storage.
- **Docker Daemon Control** — Install Docker from tarball, start/stop/restart the Docker daemon, edit `/etc/docker/daemon.json` from the UI.
- **Authentication & Security** — JWT-based auth with optional TOTP 2FA, passkey support, IP whitelisting (CIDR), domain binding, secure entrypoint enforcement, session management, password complexity and expiry policies.
- **Audit Logging** — All administrative actions are recorded and viewable in the panel.
- **System Monitoring** — Dashboard shows container counts, gateway rule status, SSL certificate expiry, Docker health, and overall system health.
- **Daemon (Auto-Restart)** — A companion daemon monitors the API process and restarts it on crash (up to 5 attempts).

### Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3, TypeScript, Element Plus, Tailwind CSS, Pinia, Vite |
| Backend | Go 1.25, Fiber v2 |
| Database | SQLite (pure Go, no CGO) via GORM |
| Container | Docker Engine SDK for Go |
| Auth | JWT (golang-jwt), bcrypt, TOTP (pquerna/otp) |
| Build | pnpm (frontend), Go toolchain (backend) |

### Quick Start

#### Prerequisites

- Go 1.25+
- pnpm
- Docker (for container management features)
- Linux (recommended; systemd units included)

#### Development

```bash
# Terminal 1: Start the Go API server
cd api
go run main.go

# Terminal 2: Start the Vue dev server
pnpm install
pnpm run dev
```

Or use the provided script:

```bash
./start-godelion.sh
```

#### Production Build

```bash
# Build frontend
pnpm run build

# Copy dist/ to api/godelion_public/
# The API serves static files from godelion_public/ automatically.

# Build Go backend
cd api
go build -o godelion main.go

# Run
./godelion
```

#### systemd Installation

```bash
sudo cp godelion.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now godelion
```

#### Daemon (optional)

A watchdog daemon is available under `daemon/` that monitors and auto-restarts the main API if it crashes. See `daemon/README.md` for details.

```bash
# Build
cd daemon
go build -o godelion-daemon main.go

# Install service
sudo cp godelion-daemon.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now godelion-daemon
```

### Default Credentials

| Username | Password |
|---|---|
| `admin` | `admin123` |

**Change these immediately after first login.**

### Configuration

The panel is configured at runtime through the web UI (**Settings → Panel Config**). Key settings:

- **Port** — panel listen port (default: `9960`)
- **HTTPS** — enable HTTPS for the panel itself (requires an SSL certificate)
- **Panel SSL Certificate** — select which uploaded cert to use for panel HTTPS

Configuration changes apply without restarting the OS-level service.

### Project Structure

```
godelion/
├── api/                    # Go backend
│   ├── main.go             # Entry point, Fiber app, routing
│   ├── controllers/        # HTTP handlers
│   ├── services/           # Business logic
│   ├── middleware/          # Auth, IP whitelist, domain binding, secure entrypoint
│   ├── models/             # GORM models
│   ├── db/                 # SQLite initialization and migration
│   └── session/            # d_delion_id session store
├── src/                    # Vue 3 frontend
│   ├── views/              # Page components
│   ├── api/                # Axios API client
│   ├── router/             # Vue Router config
│   ├── store/              # Pinia stores
│   └── components/         # Reusable components
├── daemon/                 # Watchdog daemon (Go stdlib only)
├── public/                 # Static assets
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── godelion.service        # systemd unit
```

### Security

- Change the default `admin` password immediately.
- Enable HTTPS in production.
- Use the IP whitelist feature to restrict access to trusted networks.
- Configure the secure entrypoint for an additional layer of access control.
- Regularly rotate SSL certificates and admin credentials.

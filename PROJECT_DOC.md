# Godelion - Go & Docker Container Management System

## 一、项目概述

### 1.1 项目简介

**Godelion** 是一个轻量级、高性能的 Docker 容器管理系统，提供直观的 Web 界面来管理 Docker 容器、网关代理、SSL 证书和文件存储，从而简化容器化应用的部署和运维工作。

### 1.2 核心特性

| 特性 | 描述 |
|------|------|
| **容器管理** | 创建、启动、停止、删除容器，实时查看日志 |
| **反向代理** | 基于域名的虚拟主机代理，支持多端口监听 |
| **SSL 证书** | 证书上传、管理、自动 TLS 配置 |
| **文件管理** | 上传、下载、移动、解压文件 |
| **用户认证** | JWT 令牌认证，支持管理员角色 |
| **审计日志** | 记录所有用户操作 |

### 1.3 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 前端框架 | Vue | 3.4.15 |
| 前端构建 | Vite | 5.0.12 |
| 前端语言 | TypeScript | 5.3.3 |
| UI 组件 | Element Plus | 2.13.7 |
| 状态管理 | Pinia | 3.0.4 |
| 路由管理 | Vue Router | 4.2.5 |
| 样式框架 | Tailwind CSS | 3.4.1 |
| 后端语言 | Go | 1.25.0 |
| Web 框架 | Fiber | 2.52.12 |
| 数据库 | SQLite | 嵌入式 |
| ORM | GORM | 1.31.1 |
| Docker SDK | docker/docker | 28.5.2 |

---

## 二、系统架构

### 2.1 架构图

```mermaid
flowchart TD
    A[用户浏览器] -->|HTTP/HTTPS| B[反向代理层]
    B -->|路由匹配| C{请求类型}
    C -->|管理界面| D[Vue 3 前端]
    C -->|API 请求| E[Go Fiber 后端]
    C -->|业务流量| F[Docker 容器]
    E -->|Docker SDK| G[Docker 守护进程]
    E -->|GORM| H[(SQLite 数据库)]
    G -->|管理| F
```

### 2.2 请求流程

1. **前端请求**: 用户通过浏览器访问，前端路由匹配
2. **API 调用**: 通过 axios 发起 RESTful API 请求
3. **认证校验**: JWT 令牌验证（除登录外）
4. **业务处理**: 控制器处理请求，调用服务层
5. **数据库操作**: 通过 GORM 进行数据持久化
6. **Docker 交互**: 通过 Docker SDK 操作容器

---

## 三、目录结构

```
godelion/
├── api/                          # 后端 Go 服务
│   ├── controllers/              # 控制器层（处理 HTTP 请求）
│   │   ├── audit_controller.go   # 审计日志
│   │   ├── auth_controller.go    # 用户认证
│   │   ├── gateway_controller.go # 网关规则
│   │   ├── ssl_controller.go     # SSL 证书
│   │   ├── storage_controller.go # 文件存储
│   │   ├── system_controller.go  # 系统管理
│   │   └── workload_controller.go# 容器管理
│   ├── db/                       # 数据库连接
│   │   └── db.go                 # SQLite 初始化
│   ├── middleware/               # 中间件
│   │   └── auth.go               # JWT 认证中间件
│   ├── models/                   # 数据模型
│   │   └── models.go             # GORM 模型定义
│   ├── services/                 # 服务层
│   │   ├── docker_service.go     # Docker 操作封装
│   │   ├── port_proxy.go         # 端口代理服务
│   │   └── proxy_service.go      # 反向代理服务
│   ├── go.mod                    # Go 依赖管理
│   ├── go.sum                    # 依赖校验
│   └── main.go                   # 应用入口
├── public/                       # 静态资源
│   ├── error.html                # 错误页面模板
│   └── favicon.svg               # 网站图标
├── src/                          # 前端 Vue 代码
│   ├── api/                      # API 接口定义
│   │   └── index.ts              # 统一 API 导出
│   ├── assets/                   # 静态资源
│   │   └── vue.svg               # Vue 图标
│   ├── components/               # 公共组件
│   │   └── Empty.vue             # 空状态组件
│   ├── composables/              # 组合式函数
│   │   └── useTheme.ts           # 主题切换
│   ├── lib/                      # 工具库
│   │   └── utils.ts              # 工具函数
│   ├── pages/                    # 页面组件
│   │   └── HomePage.vue          # 首页
│   ├── router/                   # 路由配置
│   │   └── index.ts              # 路由定义
│   ├── store/                    # 状态管理
│   │   └── user.ts               # 用户状态
│   ├── utils/                    # 工具函数
│   │   └── request.ts            # axios 封装
│   ├── views/                    # 视图层
│   │   ├── docker/               # Docker 管理
│   │   │   ├── Config.vue        # Docker 配置
│   │   │   └── Status.vue        # Docker 状态
│   │   ├── layout/               # 布局组件
│   │   │   └── index.vue         # 主布局
│   │   ├── Container.vue         # 容器管理页面
│   │   ├── Dashboard.vue         # 仪表盘
│   │   ├── File.vue              # 文件管理
│   │   ├── Login.vue             # 登录页面
│   │   ├── Network.vue           # 网络中继
│   │   ├── SSL.vue               # SSL 证书
│   │   └── Settings.vue          # 设置页面
│   ├── App.vue                   # 根组件
│   ├── main.ts                   # 入口文件
│   ├── style.css                 # 全局样式
│   └── vite-env.d.ts             # Vite 类型声明
├── .gitignore                    # Git 忽略配置
├── index.html                    # HTML 模板
├── package.json                  # 前端依赖
├── pnpm-lock.yaml               # pnpm 锁定文件
├── postcss.config.js            # PostCSS 配置
├── README.md                    # 项目说明
├── start-godelion.sh            # 启动脚本
├── tailwind.config.js           # Tailwind 配置
├── tsconfig.json                # TypeScript 配置
└── vite.config.ts               # Vite 配置
```

---

## 四、核心功能模块

### 4.1 用户认证模块

#### 4.1.1 功能说明

- 用户登录验证（用户名/密码）
- JWT 令牌生成与验证
- 用户信息查询与修改
- 密码修改

#### 4.1.2 API 接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 登录 | POST | `/sys/v1/auth/login` | 用户登录 |
| 获取用户信息 | GET | `/sys/v1/user/profile` | 获取当前用户信息 |
| 更新用户信息 | PUT | `/sys/v1/user/profile` | 更新用户信息和密码 |

#### 4.1.3 数据模型

```go
type User struct {
    ID             string         `gorm:"primaryKey"`
    Username       string         `gorm:"uniqueIndex;not null"`
    PasswordHash   string         `gorm:"not null"`
    Role           string         `gorm:"not null"` // "admin" or "user"
    SessionTimeout int
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### 4.2 容器管理模块

#### 4.2.1 功能说明

- 创建容器（支持 Node.js 项目部署）
- 启动/停止容器
- 查看容器日志
- 删除容器
- 更新容器配置

#### 4.2.2 API 接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 列表容器 | GET | `/sys/v1/workloads` | 获取容器列表 |
| 创建容器 | POST | `/sys/v1/workloads` | 创建新容器 |
| 启动容器 | POST | `/sys/v1/workloads/:id/start` | 启动指定容器 |
| 停止容器 | POST | `/sys/v1/workloads/:id/stop` | 停止指定容器 |
| 获取日志 | GET | `/sys/v1/workloads/:id/logs` | 获取容器日志 |
| 删除容器 | DELETE | `/sys/v1/workloads/:id` | 删除容器 |
| 更新容器 | PUT | `/sys/v1/workloads/:id` | 更新容器配置 |

#### 4.2.3 数据模型

```go
type Container struct {
    ID             string `gorm:"primaryKey"`
    DockerID       string // 实际 Docker 容器 ID
    Name           string
    Image          string
    UserID         string
    Ports          string `gorm:"type:text"` // JSON 序列化的端口映射
    ResourceLimits string
    DeploymentLogs string `gorm:"type:text"` // 部署阶段日志
    ActionLogs     string `gorm:"type:text"` // 操作日志
    Status         string `gorm:"-"`         // 运行状态（动态获取）
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### 4.3 网关代理模块

#### 4.3.1 功能说明

- 基于域名的反向代理
- 支持多端口监听
- 支持 TLS/SSL 配置
- 动态后端目标（支持容器 ID 引用）

#### 4.3.2 API 接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 列表规则 | GET | `/sys/v1/gateways` | 获取网关规则列表 |
| 创建规则 | POST | `/sys/v1/gateways` | 创建新规则 |
| 更新规则 | PUT | `/sys/v1/gateways/:id` | 更新规则 |
| 删除规则 | DELETE | `/sys/v1/gateways/:id` | 删除规则 |

#### 4.3.3 数据模型

```go
type GatewayRule struct {
    ID          string `gorm:"primaryKey"`
    Domain      string `gorm:"not null"`
    ListenPorts string // 监听端口，逗号分隔
    TargetURLs  string // 目标 URL，逗号分隔
    TargetPort  int    // 旧版字段（兼容）
    ContainerID string // 关联容器 ID
    TLSEnabled  bool
    SSLCertID   string // SSL 证书 ID
    CertPath    string
    KeyPath     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 4.4 SSL 证书模块

#### 4.4.1 功能说明

- 证书上传与验证
- 证书有效期管理
- 证书关联到网关规则

#### 4.4.2 API 接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 列表证书 | GET | `/sys/v1/ssl` | 获取证书列表 |
| 创建证书 | POST | `/sys/v1/ssl` | 上传新证书 |
| 删除证书 | DELETE | `/sys/v1/ssl/:id` | 删除证书 |

#### 4.4.3 数据模型

```go
type SSLCertificate struct {
    ID          string `gorm:"primaryKey"`
    Domain      string `gorm:"not null"`
    CertContent string `gorm:"type:text"`
    KeyContent  string `gorm:"type:text"`
    IssuedAt    time.Time
    ExpiresAt   time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 4.5 文件存储模块

#### 4.5.1 功能说明

- 文件上传
- 文件下载
- 文件列表
- 文件夹创建
- 文件移动
- 压缩包解压

#### 4.5.2 API 接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 上传文件 | POST | `/sys/v1/storage/upload` | 上传文件 |
| 创建文件夹 | POST | `/sys/v1/storage/folder` | 创建文件夹 |
| 移动文件 | POST | `/sys/v1/storage/move` | 移动文件 |
| 解压文件 | POST | `/sys/v1/storage/extract` | 解压归档文件 |
| 列表文件 | GET | `/sys/v1/storage/list` | 获取文件列表 |
| 读取文件 | GET | `/sys/v1/storage/read` | 读取文件内容 |
| 删除文件 | DELETE | `/sys/v1/storage/delete` | 删除文件 |
| 下载文件 | GET | `/sys/v1/storage/download` | 下载文件 |

### 4.6 系统管理模块

#### 4.6.1 功能说明

- Docker 状态检测
- Docker 安装
- Docker 启动/停止/重启
- Docker 配置管理（daemon.json）

#### 4.6.2 API 接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 获取状态 | GET | `/sys/v1/system/docker/status` | 获取 Docker 状态 |
| 安装 Docker | POST | `/sys/v1/system/docker/install` | 安装 Docker |
| 启动 Docker | POST | `/sys/v1/system/docker/start` | 启动 Docker 服务 |
| 停止 Docker | POST | `/sys/v1/system/docker/stop` | 停止 Docker 服务 |
| 重启 Docker | POST | `/sys/v1/system/docker/restart` | 重启 Docker 服务 |
| 获取配置 | GET | `/sys/v1/system/docker/config` | 获取 Docker 配置 |
| 更新配置 | POST | `/sys/v1/system/docker/config` | 更新 Docker 配置 |

### 4.7 审计日志模块

#### 4.7.1 功能说明

- 记录所有用户操作
- 支持查询操作日志

#### 4.7.2 API 接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 列表日志 | GET | `/sys/v1/audit` | 获取审计日志列表 |

#### 4.7.3 数据模型

```go
type AuditLog struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    string
    Action    string
    Resource  string
    IPAddress string
    Details   string
    CreatedAt time.Time
}
```

---

## 五、数据库设计

### 5.1 数据库关系图

```mermaid
erDiagram
    USER ||--o{ CONTAINER : owns
    CONTAINER ||--o{ GATEWAY_RULE : has
    GATEWAY_RULE ||--o{ SSL_CERTIFICATE : uses
    USER ||--o{ AUDIT_LOG : creates

    USER {
        string id PK
        string username UK
        string password_hash
        string role
        int session_timeout
        datetime created_at
        datetime updated_at
    }

    CONTAINER {
        string id PK
        string docker_id
        string name
        string image
        string user_id FK
        text ports
        string resource_limits
        text deployment_logs
        text action_logs
        datetime created_at
        datetime updated_at
    }

    GATEWAY_RULE {
        string id PK
        string domain
        string listen_ports
        string target_urls
        int target_port
        string container_id FK
        bool tls_enabled
        string ssl_cert_id FK
        string cert_path
        string key_path
        datetime created_at
        datetime updated_at
    }

    SSL_CERTIFICATE {
        string id PK
        string domain
        text cert_content
        text key_content
        datetime issued_at
        datetime expires_at
        datetime created_at
        datetime updated_at
    }

    AUDIT_LOG {
        uint id PK
        string user_id FK
        string action
        string resource
        string ip_address
        string details
        datetime created_at
    }
```

---

## 六、API 接口详细文档

### 6.1 认证接口

#### 6.1.1 登录

**请求:**
```http
POST /sys/v1/auth/login
Content-Type: application/json

{
    "username": "admin",
    "password": "admin123"
}
```

**成功响应:**
```json
{
    "code": 200,
    "message": "Login successful",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
            "id": "u_admin",
            "username": "admin",
            "role": "admin"
        }
    }
}
```

#### 6.1.2 获取用户信息

**请求:**
```http
GET /sys/v1/user/profile
Authorization: Bearer <token>
```

**成功响应:**
```json
{
    "code": 200,
    "message": "Success",
    "data": {
        "id": "u_admin",
        "username": "admin",
        "role": "admin"
    }
}
```

#### 6.1.3 更新用户信息

**请求:**
```http
PUT /sys/v1/user/profile
Authorization: Bearer <token>
Content-Type: application/json

{
    "username": "newadmin",
    "old_password": "admin123",
    "new_password": "newpassword"
}
```

**成功响应:**
```json
{
    "code": 200,
    "message": "Profile updated successfully"
}
```

### 6.2 容器管理接口

#### 6.2.1 创建容器

**请求:**
```http
POST /sys/v1/workloads
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "My Node App",
    "node_version": "node:24-alpine",
    "project_dir": "projects/myapp",
    "start_command": "npm run dev",
    "container_name": "myapp-container",
    "package_manager": "npm",
    "dependencies": "express,axios",
    "ports": [
        {"host": "3000", "container": "3000"}
    ],
    "resource_limits": "{}"
}
```

**成功响应:**
```json
{
    "code": 200,
    "message": "Container deployment started in background",
    "data": {
        "id": "uuid-xxx",
        "name": "My Node App",
        "image": "node:24-alpine (Source: projects/myapp)",
        ...
    }
}
```

#### 6.2.2 列表容器

**请求:**
```http
GET /sys/v1/workloads
Authorization: Bearer <token>
```

**成功响应:**
```json
{
    "code": 200,
    "message": "Success",
    "data": [
        {
            "id": "uuid-xxx",
            "docker_id": "abc123...",
            "name": "My Node App",
            "image": "node:24-alpine",
            "status": "running",
            ...
        }
    ]
}
```

### 6.3 网关代理接口

#### 6.3.1 创建网关规则

**请求:**
```http
POST /sys/v1/gateways
Authorization: Bearer <token>
Content-Type: application/json

{
    "domain": "example.com",
    "listen_ports": "80,443",
    "target_urls": "127.0.0.1:3000",
    "tls_enabled": true,
    "ssl_cert_id": "cert-uuid"
}
```

**成功响应:**
```json
{
    "code": 200,
    "message": "Gateway rule created",
    "data": {
        "id": "uuid-xxx",
        "domain": "example.com",
        ...
    }
}
```

### 6.4 SSL 证书接口

#### 6.4.1 上传证书

**请求:**
```http
POST /sys/v1/ssl
Authorization: Bearer <token>
Content-Type: application/json

{
    "domain": "example.com",
    "cert_content": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "key_content": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
}
```

**成功响应:**
```json
{
    "code": 200,
    "message": "Certificate created successfully",
    "data": {
        "id": "uuid-xxx",
        "domain": "example.com",
        "issued_at": "2024-01-01T00:00:00Z",
        "expires_at": "2025-01-01T00:00:00Z"
    }
}
```

---

## 七、部署与运行

### 7.1 环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | >= 1.25 | 后端开发语言 |
| Node.js | >= 20 | 前端开发环境 |
| Docker | >= 20 | 容器运行时 |
| pnpm | >= 8 | 包管理器 |

### 7.2 快速开始

#### 7.2.1 后端启动

```bash
cd api
go mod download
go run main.go
```

后端服务将在 `http://localhost:8080` 启动。

#### 7.2.2 前端启动

```bash
pnpm install
pnpm run dev
```

前端服务将在 `http://localhost:5173` 启动。

#### 7.2.3 一键启动

```bash
./start-godelion.sh
```

### 7.3 访问方式

1. **前端界面**: http://localhost:5173
2. **API 接口**: http://localhost:8080/sys/v1/
3. **默认账号**: admin / admin123

### 7.4 生产部署

#### 7.4.1 构建前端

```bash
pnpm run build
```

构建产物将输出到 `dist` 目录。

#### 7.4.2 构建后端

```bash
cd api
go build -o godelion-api .
```

#### 7.4.3 环境变量配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | 服务端口 | 8080 |
| JWT_SECRET | JWT 密钥 | super-secret-key-change-me |
| DATABASE_PATH | 数据库路径 | ./godelion/godelion.db |

---

## 八、安全注意事项

### 8.1 密钥管理

- JWT 密钥应通过环境变量配置，不应硬编码
- 数据库密码应加密存储（当前使用 bcrypt）

### 8.2 权限控制

- 管理员（admin）可以访问所有资源
- 普通用户（user）只能访问自己创建的容器

### 8.3 输入验证

- 所有 API 输入都应进行验证
- 文件上传应限制大小和类型
- 防止路径遍历攻击

### 8.4 SSL/TLS

- 生产环境应启用 HTTPS
- SSL 证书应定期更新

---

## 九、扩展建议

### 9.1 功能扩展

- [ ] 镜像管理（拉取、删除镜像）
- [ ] 网络管理（创建、配置 Docker 网络）
- [ ] 数据卷管理
- [ ] 容器编排（Docker Compose 支持）
- [ ] 定时任务
- [ ] WebSocket 实时日志

### 9.2 性能优化

- [ ] Redis 缓存支持
- [ ] 数据库连接池优化
- [ ] 请求限流
- [ ] 异步任务队列

### 9.3 安全增强

- [ ] 双因素认证
- [ ] 操作审计增强
- [ ] IP 白名单
- [ ] 容器资源隔离

---

## 十、版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0.0 | 2024 | 初始版本，基础容器管理功能 |

---

## 十一、Docker 应用部署指南

### 11.1 概述

Godelion 支持部署和管理多种类型的 Docker 容器应用。系统通过 Docker SDK 与 Docker 引擎交互，实现容器的创建、启动、停止、删除和日志查看等功能。

### 11.2 支持的应用类型

| 应用类型 | 推荐镜像 | 说明 |
|---------|---------|------|
| **Node.js** | `node:24-alpine` | 轻量级 Alpine 版本 |
| **Python** | `python:3.12-alpine` | Python 3.12 运行环境 |
| **Go** | `golang:1.22-alpine` | Go 编译运行环境 |
| **PHP** | `php:8.3-apache` | PHP + Apache 环境 |
| **静态网页** | `nginx:alpine` | Nginx 静态文件服务器 |
| **C** | `gcc:latest` | GCC 编译环境 |
| **C++** | `gcc:latest` | G++ 编译环境 |
| **二进制程序** | `alpine:latest` | 极简 Linux 环境 |

### 11.3 通用部署流程

1. **准备项目文件**：通过文件管理模块上传项目文件
2. **创建容器**：选择应用类型，配置参数并创建容器
3. **自动部署**：系统自动拉取镜像、安装依赖（如需要）、启动应用
4. **代理配置**：配置网关规则，通过域名访问应用

### 11.4 各类应用部署指南

#### 11.4.1 Node.js 应用

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "我的 Express 应用" |
| `runtime_type` | 是 | 运行时类型 | "nodejs" |
| `image` | 否 | Docker 镜像 | "node:24-alpine" |
| `project_dir` | 是 | 项目目录 | "projects/my-express-app" |
| `start_command` | 是 | 启动命令 | "dev" 或 "node server.js" |
| `container_name` | 否 | Docker 容器名称 | "my-express-container" |
| `package_manager` | 否 | 包管理器 | "npm" |
| `dependencies` | 否 | 额外依赖（逗号分隔） | "express,cors,axios" |
| `ports` | 是 | 端口映射 | `[{"host":"3000","container":"3000"}]` |

**部署示例：**

```json
{
    "name": "Express API 服务",
    "runtime_type": "nodejs",
    "image": "node:20-alpine",
    "project_dir": "projects/express-api",
    "start_command": "start",
    "container_name": "express-api-container",
    "package_manager": "npm",
    "ports": [
        {"host": "3000", "container": "3000"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`node:20-alpine`
2. 安装依赖：`npm install`
3. 启动应用：`npm run start`

---

#### 11.4.2 Python 应用

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "Flask API" |
| `runtime_type` | 是 | 运行时类型 | "python" |
| `image` | 否 | Docker 镜像 | "python:3.12-alpine" |
| `project_dir` | 是 | 项目目录 | "projects/flask-api" |
| `start_command` | 是 | 启动命令 | "python app.py" |
| `container_name` | 否 | Docker 容器名称 | "flask-container" |
| `requirements_file` | 否 | 依赖文件 | "requirements.txt" |
| `ports` | 是 | 端口映射 | `[{"host":"5000","container":"5000"}]` |

**部署示例：**

```json
{
    "name": "Flask API 服务",
    "runtime_type": "python",
    "image": "python:3.12-alpine",
    "project_dir": "projects/flask-api",
    "start_command": "python app.py",
    "container_name": "flask-container",
    "requirements_file": "requirements.txt",
    "ports": [
        {"host": "5000", "container": "5000"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`python:3.12-alpine`
2. 安装依赖：`pip install -r requirements.txt`
3. 启动应用：`python app.py`

---

#### 11.4.3 Go 应用

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "Go Web 服务" |
| `runtime_type` | 是 | 运行时类型 | "go" |
| `image` | 否 | Docker 镜像 | "golang:1.22-alpine" |
| `project_dir` | 是 | 项目目录 | "projects/go-app" |
| `build_command` | 否 | 编译命令 | "go build -o app main.go" |
| `start_command` | 是 | 启动命令 | "./app" |
| `container_name` | 否 | Docker 容器名称 | "go-container" |
| `ports` | 是 | 端口映射 | `[{"host":"8080","container":"8080"}]` |

**部署示例：**

```json
{
    "name": "Go Web 服务",
    "runtime_type": "go",
    "image": "golang:1.22-alpine",
    "project_dir": "projects/go-app",
    "build_command": "go build -o app main.go",
    "start_command": "./app",
    "container_name": "go-container",
    "ports": [
        {"host": "8080", "container": "8080"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`golang:1.22-alpine`
2. 编译代码：`go build -o app main.go`
3. 启动应用：`./app`

---

#### 11.4.4 PHP 应用

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "PHP 网站" |
| `runtime_type` | 是 | 运行时类型 | "php" |
| `image` | 否 | Docker 镜像 | "php:8.3-apache" |
| `project_dir` | 是 | 项目目录 | "projects/php-website" |
| `container_name` | 否 | Docker 容器名称 | "php-container" |
| `ports` | 是 | 端口映射 | `[{"host":"8000","container":"80"}]` |

**部署示例：**

```json
{
    "name": "PHP 网站",
    "runtime_type": "php",
    "image": "php:8.3-apache",
    "project_dir": "projects/php-website",
    "container_name": "php-container",
    "ports": [
        {"host": "8000", "container": "80"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`php:8.3-apache`
2. 挂载项目目录到 Apache 文档根目录
3. 自动启动 Apache 服务

---

#### 11.4.5 静态网页

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "我的博客" |
| `runtime_type` | 是 | 运行时类型 | "static" |
| `image` | 否 | Docker 镜像 | "nginx:alpine" |
| `project_dir` | 是 | 项目目录 | "projects/static-site" |
| `container_name` | 否 | Docker 容器名称 | "nginx-container" |
| `ports` | 是 | 端口映射 | `[{"host":"8080","container":"80"}]` |

**部署示例：**

```json
{
    "name": "静态博客网站",
    "runtime_type": "static",
    "image": "nginx:alpine",
    "project_dir": "projects/static-site",
    "container_name": "nginx-container",
    "ports": [
        {"host": "8080", "container": "80"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`nginx:alpine`
2. 挂载项目目录到 `/usr/share/nginx/html`
3. 自动启动 Nginx 服务

---

#### 11.4.6 二进制程序

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "二进制服务" |
| `runtime_type` | 是 | 运行时类型 | "binary" |
| `image` | 否 | Docker 镜像 | "alpine:latest" |
| `project_dir` | 是 | 项目目录 | "projects/binary-app" |
| `start_command` | 是 | 启动命令 | "./myapp" |
| `container_name` | 否 | Docker 容器名称 | "binary-container" |
| `ports` | 是 | 端口映射 | `[{"host":"9000","container":"9000"}]` |

**部署示例：**

```json
{
    "name": "二进制网络服务",
    "runtime_type": "binary",
    "image": "alpine:latest",
    "project_dir": "projects/binary-app",
    "start_command": "./myapp",
    "container_name": "binary-container",
    "ports": [
        {"host": "9000", "container": "9000"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`alpine:latest`
2. 设置文件执行权限（如需要）
3. 启动二进制程序：`./myapp`

---

#### 11.4.7 C 应用

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "C Web 服务" |
| `runtime_type` | 是 | 运行时类型 | "c" |
| `image` | 否 | Docker 镜像 | "gcc:latest" |
| `project_dir` | 是 | 项目目录 | "projects/c-app" |
| `build_command` | 否 | 编译命令 | "gcc -o app main.c" |
| `start_command` | 是 | 启动命令 | "./app" |
| `container_name` | 否 | Docker 容器名称 | "c-container" |
| `ports` | 是 | 端口映射 | `[{"host":"8080","container":"8080"}]` |

**部署示例：**

```json
{
    "name": "C HTTP 服务",
    "runtime_type": "c",
    "image": "gcc:latest",
    "project_dir": "projects/c-app",
    "build_command": "gcc -o server server.c -lpthread",
    "start_command": "./server",
    "container_name": "c-server",
    "ports": [
        {"host": "8080", "container": "8080"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`gcc:latest`
2. 编译代码：`gcc -o server server.c -lpthread`
3. 启动应用：`./server`

---

#### 11.4.8 C++ 应用

**创建容器参数：**

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `name` | 是 | 应用名称 | "C++ 服务" |
| `runtime_type` | 是 | 运行时类型 | "cpp" |
| `image` | 否 | Docker 镜像 | "gcc:latest" |
| `project_dir` | 是 | 项目目录 | "projects/cpp-app" |
| `build_command` | 否 | 编译命令 | "g++ -o app main.cpp" |
| `start_command` | 是 | 启动命令 | "./app" |
| `container_name` | 否 | Docker 容器名称 | "cpp-container" |
| `ports` | 是 | 端口映射 | `[{"host":"9000","container":"9000"}]` |

**部署示例：**

```json
{
    "name": "C++ 网络服务",
    "runtime_type": "cpp",
    "image": "gcc:latest",
    "project_dir": "projects/cpp-app",
    "build_command": "g++ -std=c++17 -o app main.cpp -lpthread",
    "start_command": "./app",
    "container_name": "cpp-server",
    "ports": [
        {"host": "9000", "container": "9000"}
    ]
}
```

**执行的命令序列：**
1. 拉取镜像：`gcc:latest`
2. 编译代码：`g++ -std=c++17 -o app main.cpp -lpthread`
3. 启动应用：`./app`

### 11.5 容器部署技术细节

#### 11.5.1 目录挂载

系统将主机上的项目目录挂载到容器内的对应目录：

| 应用类型 | 容器内挂载目录 |
|---------|--------------|
| Node.js, Python, Go, C, C++, 二进制程序 | `/app` |
| PHP (Apache) | `/var/www/html` |
| 静态网页 (Nginx) | `/usr/share/nginx/html` |

这样可以实现：
- 代码热更新（取决于应用是否支持）
- 持久化存储项目文件
- 无需重新构建镜像即可修改代码

#### 11.5.2 容器配置

| 配置项 | 值 | 说明 |
|-------|-----|------|
| 工作目录 | `/app` (除 PHP/静态外) | 容器内工作目录 |
| 重启策略 | `unless-stopped` | 除非手动停止，否则自动重启 |
| TTY | `false` | 不分配伪终端 |
| 网络模式 | 默认 bridge | 与宿主机在同一网络 |

#### 11.5.3 异步部署流程

容器创建采用异步处理，流程如下：

```mermaid
flowchart TD
    A[接收部署请求] --> B[保存容器记录到数据库]
    B --> C[返回部署开始响应]
    C --> D[异步拉取 Docker 镜像]
    D --> E[创建并启动容器]
    E --> F[更新数据库记录 Docker ID]
    F --> G[启动网关代理]
```

可以通过日志接口查看部署进度和运行日志。

### 11.6 应用访问配置

部署完成后，可以通过以下方式访问应用：

#### 11.6.1 通过主机端口直接访问

如果配置了端口映射，可以通过 `http://localhost:{host_port}` 直接访问。

#### 11.6.2 通过网关代理访问（推荐）

1. 在网关模块创建规则
2. 配置域名和监听端口
3. 关联到容器的内部端口
4. 可选择配置 SSL 证书启用 HTTPS

### 11.7 日志查看

系统提供完整的日志查看功能：

| 日志类型 | 说明 |
|---------|------|
| 部署日志 | 镜像拉取、容器创建过程 |
| 运行日志 | 应用运行时的 stdout/stderr |
| 操作日志 | 容器启动、停止等操作记录 |

### 11.8 常见问题

#### Q: 镜像拉取失败怎么办？
A: 可以在系统配置中配置 Docker 镜像加速器，编辑 `/etc/docker/daemon.json` 添加镜像源。

#### Q: 如何使用自定义 Docker 镜像？
A: 在 `image` 参数中直接输入完整的镜像名称，如 `my-custom-image:latest`。

#### Q: 应用需要环境变量怎么办？
A: 可以在项目目录中创建 `.env` 文件，应用可以读取该文件获取环境变量。

#### Q: 如何持久化应用数据？
A: 应用数据保存在挂载的主机目录中，只要不删除项目文件，数据就会保留。

#### Q: Go 应用能否不编译直接运行？
A: 可以，使用 `go run main.go` 作为启动命令即可，但建议编译后运行以获得更好的性能。

---

## 十二、附录：默认用户

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |

> **注意**: 首次启动时，系统会自动创建默认管理员用户。请在生产环境中立即修改默认密码。
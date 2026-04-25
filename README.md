# Godelion - Go & Docker Container Management System

[English](#english) | [中文](#中文)

---

<a name="english"></a>
## 🇬🇧 English

### Overview
**Godelion** is a lightweight, high-performance Docker container management system. It aims to provide an intuitive web-based interface for managing Docker containers, images, networks, and volumes, simplifying the deployment and operation of containerized applications.

### Current Progress
We have completed the foundational scaffolding and core structural design of the project.

- [x] **Project Initialization**: Frontend and backend project structures created.
- [x] **Frontend Architecture**: Built with Vue 3 + TypeScript + Vite, integrated with Tailwind CSS for styling, and Vue Router for routing.
- [x] **Backend Architecture**: Built with Go, providing RESTful API services. Basic database connection and middleware setups are in place.
- [x] **Core Scripting**: Created `start-godelion.sh` for one-click concurrent startup of both frontend and backend services.
- [x] **Version Control**: Cleaned up unnecessary files (e.g., `node_modules`, binary builds, editor configurations) and successfully pushed the clean codebase to GitHub.
- [ ] **Docker API Integration**: Implementing Go backend communication with the Docker Daemon.
- [ ] **Frontend Dashboard**: Developing pages for container status, image management, and network configuration.

### System Architecture
The project adopts a modern **Frontend-Backend Separation** architecture, with a clear flow of requests from the user to the underlying Docker containers:

```text
Visitor (User Browser)
 └──> Frontend (Vue 3 + Vite)
       └──> Proxy (Go API Service)
             └──> Docker Daemon (Host Machine)
                   └──> Target Docker Container
```

1. **Frontend (Web UI)**
   - **Tech Stack**: Vue 3 (Composition API), TypeScript, Vite, Tailwind CSS.
   - **Role**: Provides the user interface for monitoring and managing Docker resources visually.
   - **Directory**: `src/` (Vue components, views, routers, store), `public/` (static assets).

2. **Backend (API Service)**
   - **Tech Stack**: Go (Golang).
   - **Role**: Handles business logic, interacts with the Docker engine via Docker SDK for Go, and provides RESTful APIs for the frontend.
   - **Directory**: `api/` (contains `controllers`, `models`, `services`, `middleware`, `db`, etc.).

3. **Data Storage**
   - **Tech Stack**: SQLite (currently indicated by `godelion.db`).
   - **Role**: Stores system configurations, user accounts, and operational logs locally without needing a heavy database server.

4. **Startup & Deployment**
   - **Script**: `start-godelion.sh` manages the simultaneous execution of the Vite dev server and the Go API server.

---

<a name="中文"></a>
## 🇨🇳 中文

### 项目简介
**Godelion** 是一个轻量级、高性能的 Docker 容器管理系统。它旨在提供一个直观的 Web 界面来管理 Docker 容器、镜像、网络和数据卷，从而简化容器化应用的部署和运维工作。

### 目前进度
目前我们已经完成了项目的基础脚手架搭建和核心架构设计。

- [x] **项目初始化**：完成前端和后端项目目录结构的创建。
- [x] **前端架构**：基于 Vue 3 + TypeScript + Vite 搭建，集成了 Tailwind CSS 进行样式开发，并配置了 Vue Router。
- [x] **后端架构**：基于 Go 语言构建，提供 RESTful API 服务，已完成基础的数据库连接和中间件配置。
- [x] **核心脚本**：编写了 `start-godelion.sh` 脚本，支持一键同时启动前后端服务。
- [x] **版本控制**：清理了冗余文件（如 `node_modules`、编译的二进制文件、编辑器配置等），并将纯净的代码库成功推送到 GitHub。
- [ ] **Docker API 集成**：开发 Go 后端与 Docker 守护进程的通信逻辑。
- [ ] **前端控制台**：开发容器状态监控、镜像管理、网络配置等页面。

### 系统架构
项目采用现代化的**前后端分离**架构。整体请求流程从用户侧出发，通过代理层最终访问到 Docker 容器，其树状访问流程如下：

```text
访问者 (用户浏览器)
 └──> 前端 (Vue 3 + Vite)
       └──> 代理服务 (Go API 后端)
             └──> Docker 守护进程 (宿主机)
                   └──> 目标 Docker 容器
```

1. **前端 (Web UI)**
   - **技术栈**：Vue 3 (Composition API), TypeScript, Vite, Tailwind CSS。
   - **职责**：提供用户界面，用于可视化地监控和管理 Docker 资源。
   - **目录**：`src/`（Vue 组件、视图、路由、状态管理），`public/`（静态资源）。

2. **后端 (API Service)**
   - **技术栈**：Go (Golang)。
   - **职责**：处理核心业务逻辑，通过 Docker SDK for Go 与宿主机的 Docker 引擎进行交互，并为前端提供 RESTful API 接口。
   - **目录**：`api/`（包含 `controllers`, `models`, `services`, `middleware`, `db` 等模块）。

3. **数据存储**
   - **技术栈**：SQLite（目前通过 `godelion.db` 文件体现）。
   - **职责**：在本地轻量级存储系统配置、用户账号和操作日志，无需额外部署大型数据库服务。

4. **启动与部署**
   - **脚本管理**：通过 `start-godelion.sh` 统一管理 Vite 开发服务器和 Go API 服务器的并发启动。

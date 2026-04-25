#!/bin/bash

# ==============================================================================
# Godelion 启动脚本
# 该脚本用于同时启动 Vue 前端 (Vite) 和 Go 后端服务。
# 支持在接收到中断信号 (Ctrl+C) 时，自动清理并关闭这两个进程。
# ==============================================================================

# 获取脚本所在的当前目录绝对路径，无论在哪执行都能正确定位到项目根目录
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# 定义颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}        启动 Godelion 系统服务...       ${NC}"
echo -e "${BLUE}========================================${NC}"

# 初始化 PID 变量
FRONTEND_PID=""
BACKEND_PID=""

# 清理函数：捕获到退出信号时执行，优雅关闭后台进程
cleanup() {
    echo -e "\n${RED}正在关闭服务...${NC}"
    if [ -n "$FRONTEND_PID" ]; then
        echo "停止前端服务 (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID 2>/dev/null
    fi
    if [ -n "$BACKEND_PID" ]; then
        echo "停止后端服务 (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null
    fi
    echo -e "${GREEN}服务已全部关闭。${NC}"
    exit 0
}

# 注册信号捕获 (SIGINT, SIGTERM)
trap cleanup SIGINT SIGTERM

# 1. 启动 Go 后端
echo -e "${GREEN}[1/2] 正在启动 Go 后端服务...${NC}"
cd "$PROJECT_ROOT/api" || { echo -e "${RED}错误：找不到 $PROJECT_ROOT/api 目录。${NC}"; exit 1; }
go run main.go &
BACKEND_PID=$!
echo "后端服务已在后台运行 (PID: $BACKEND_PID)"

# 稍微等待一下，确保后端初始化完成（如数据库等）
sleep 2

# 2. 启动 Vue 前端
echo -e "${GREEN}[2/2] 正在启动 Vue 前端服务...${NC}"
cd "$PROJECT_ROOT" || { echo -e "${RED}错误：找不到 $PROJECT_ROOT 目录。${NC}"; exit 1; }

# 检查是否安装了 pnpm，如果没有则回退使用 npm
if command -v pnpm &> /dev/null; then
    pnpm run dev --host &
else
    echo -e "${BLUE}未检测到 pnpm，将回退使用 npm...${NC}"
    npm run dev --host &
fi

FRONTEND_PID=$!
echo "前端服务已在后台运行 (PID: $FRONTEND_PID)"

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Godelion 系统已成功启动！${NC}"
echo -e "项目根目录: $PROJECT_ROOT"
echo -e "前端访问地址: http://localhost:5173 (或 Vite 分配的其他端口)"
echo -e "后端 API 地址: 默认监听相应的端口"
echo -e "${RED}按 Ctrl+C 停止所有服务。${NC}"
echo -e "${BLUE}========================================${NC}"

# 等待后台进程，防止脚本直接退出
wait $FRONTEND_PID $BACKEND_PID

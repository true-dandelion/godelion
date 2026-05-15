#!/bin/bash

# ==============================================================================
# Godelion 守护进程安装和启动脚本
# 该脚本用于编译、安装和启动 Godelion 守护进程
# ==============================================================================

# 获取脚本所在的当前目录绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
DAEMON_DIR="$SCRIPT_DIR"
INSTALL_DIR="/root/godelion"

# 定义颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}     Godelion 守护进程安装程序         ${NC}"
echo -e "${BLUE}========================================${NC}"

# 编译守护进程
echo -e "${GREEN}[1/3] 正在编译守护进程...${NC}"
cd "$DAEMON_DIR" || { echo -e "${RED}错误：找不到 $DAEMON_DIR 目录。${NC}"; exit 1; }

# 初始化 Go 模块
if [ ! -f "go.mod" ]; then
    echo -e "${YELLOW}初始化 Go 模块...${NC}"
    go mod init godelion-daemon
fi

# 下载依赖
echo -e "${YELLOW}下载依赖...${NC}"
go mod tidy

# 编译
echo -e "${YELLOW}编译中...${NC}"
go build -o godelion-daemon main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}编译失败！${NC}"
    exit 1
fi

echo -e "${GREEN}编译成功！${NC}"

# 创建安装目录
echo -e "${GREEN}[2/3] 正在安装守护进程...${NC}"
mkdir -p "$INSTALL_DIR/bin"
mkdir -p "$INSTALL_DIR/logs"

# 复制守护进程
cp godelion-daemon "$INSTALL_DIR/bin/"
chmod +x "$INSTALL_DIR/bin/godelion-daemon"

echo -e "${GREEN}守护进程已安装到: $INSTALL_DIR/bin/godelion-daemon${NC}"

# 创建 systemd 服务文件
echo -e "${GREEN}[3/3] 正在配置 systemd 服务...${NC}"

SERVICE_FILE="/etc/systemd/system/godelion-daemon.service"

if [ -f "$SERVICE_FILE" ]; then
    echo -e "${YELLOW}systemd 服务文件已存在，是否覆盖？${NC}"
    read -p "输入 y 覆盖，其他键跳过: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}跳过 systemd 配置。${NC}"
    else
        sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=Godelion Daemon Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/godelion-daemon
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        echo -e "${GREEN}systemd 服务文件已创建: $SERVICE_FILE${NC}"
        
        # 重新加载 systemd
        echo -e "${YELLOW}重新加载 systemd 配置...${NC}"
        sudo systemctl daemon-reload
    fi
else
    sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=Godelion Daemon Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/godelion-daemon
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    echo -e "${GREEN}systemd 服务文件已创建: $SERVICE_FILE${NC}"
    
    # 重新加载 systemd
    echo -e "${YELLOW}重新加载 systemd 配置...${NC}"
    sudo systemctl daemon-reload
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}安装完成！${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "使用方法："
echo ""
echo -e "${YELLOW}方式一：直接运行守护进程${NC}"
echo -e "  cd $INSTALL_DIR/bin"
echo -e "  ./godelion-daemon"
echo ""
echo -e "${YELLOW}方式二：使用 systemd 管理${NC}"
echo -e "  启动服务:   sudo systemctl start godelion-daemon"
echo -e "  停止服务:   sudo systemctl stop godelion-daemon"
echo -e "  重启服务:   sudo systemctl restart godelion-daemon"
echo -e "  查看状态:   sudo systemctl status godelion-daemon"
echo -e "  设置开机自启: sudo systemctl enable godelion-daemon"
echo ""
echo -e "${YELLOW}方式三：手动启动${NC}"
echo -e "  后台运行: nohup $INSTALL_DIR/bin/godelion-daemon > /dev/null 2>&1 &"
echo ""
echo -e "${YELLOW}日志位置：${NC}"
echo -e "  $INSTALL_DIR/<日期时间>/logs/"
echo ""
echo -e "${YELLOW}PID 文件：${NC}"
echo -e "  $INSTALL_DIR/daemon.pid"
echo ""
echo -e "${BLUE}========================================${NC}"

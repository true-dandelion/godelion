#!/bin/bash

echo "=== Docker 卸载脚本 ==="

# Step 1: Stop and disable services
echo "[1/4] 停止 Docker 服务..."
systemctl stop docker.socket 2>/dev/null
systemctl stop docker.service 2>/dev/null
systemctl stop containerd.service 2>/dev/null
systemctl disable docker.service 2>/dev/null
systemctl disable docker.socket 2>/dev/null
systemctl disable containerd.service 2>/dev/null

# Step 2: Remove systemd service files
echo "[2/4] 删除 systemd 服务文件..."
rm -f /etc/systemd/system/docker.service
rm -f /etc/systemd/system/docker.socket
rm -f /etc/systemd/system/containerd.service
rm -rf /etc/systemd/system/docker.service.wants
rm -rf /etc/systemd/system/multi-user.target.wants/docker.service
systemctl daemon-reload

# Step 3: Remove binaries
echo "[3/4] 删除二进制文件..."
for bin in docker dockerd containerd containerd-shim containerd-shim-runc-v2 ctr runc docker-init docker-proxy; do
    if [ -f "/usr/bin/$bin" ]; then
        rm -f "/usr/bin/$bin"
        echo "  ✓ 已删除 /usr/bin/$bin"
    fi
done

# Step 4: Clean up data
echo "[4/4] 清理数据..."
rm -rf /var/lib/docker
rm -rf /var/lib/containerd
rm -rf /etc/docker
rm -rf /tmp/docker-install
rm -f /var/run/docker.sock
rm -f /var/run/docker.pid

echo ""
echo "=== Docker 卸载完成 ==="

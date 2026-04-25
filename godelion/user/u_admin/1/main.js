// 导入依赖
const express = require('express');
const app = express();

// 获取客户端真实 IP（支持代理、CDN、NAT）
app.set('trust proxy', true);

// 主接口：返回 IP + NAT + 网络信息
app.get('/ipinfo', (req, res) => {
  // 1. 获取客户端真实 IP（优先从代理头读取）
  const clientIp = 
    req.headers['x-forwarded-for']?.split(',')[0].trim() || // 多层代理/CDN
    req.ip || 
    req.connection.remoteAddress || 
    '未知';

  // 2. 获取 NAT 出口 IP（公网出口地址）
  const natPublicIp = req.connection.remoteAddress;

  // 3. 端口
  const clientPort = req.connection.remotePort;

  // 4. 请求协议
  const protocol = req.protocol;

  // 5. 完整请求信息
  const result = {
    code: 200,
    msg: "获取成功",
    data: {
      // 核心 IP 信息
      client_real_ip: clientIp,
      nat_public_ip: natPublicIp,
      client_port: clientPort,
      
      // 请求环境
      protocol: protocol,
      method: req.method,
      host: req.headers.host,
      
      // 原始请求头（可查看代理信息）
      headers: {
        x_forwarded_for: req.headers['x-forwarded-for'] || '无',
        x_real_ip: req.headers['x-real-ip'] || '无',
        user_agent: req.headers['user-agent'] || '无'
      }
    }
  };

  // 返回 JSON
  res.json(result);
});

// 启动服务
const PORT = 3000;
app.listen(PORT, () => {
  console.log(`服务运行中：http://127.0.0.1:${PORT}/ipinfo`);
});
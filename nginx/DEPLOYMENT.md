# Docker Registry Nginx 部署指南

本文档介绍如何使用 Nginx 作为反向代理来为 Docker Registry 提供 HTTPS 访问。

## 概述

- **域名**: test.arctanp.top
- **后端服务**: Docker Registry (localhost:5000)
- **协议**: HTTPS (SSL/TLS)
- **代理服务器**: Nginx

## 部署步骤

### 1. 前置条件

确保以下条件已满足：
- 域名 `test.arctanp.top` 已解析到当前服务器的公网 IP
- Docker Registry 服务运行在本地 5000 端口
- 已安装 Nginx 和 Certbot

### 2. 安装依赖

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install nginx certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install nginx certbot python3-certbot-nginx
```

### 3. 部署 Nginx 配置

```bash
# 复制配置文件到 Nginx 配置目录
sudo cp nginx/registry.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/registry.conf /etc/nginx/sites-enabled/

# 测试配置文件语法
sudo nginx -t

# 重新加载 Nginx
sudo systemctl reload nginx
```

### 4. 申请 SSL 证书

```bash
# 运行 SSL 设置脚本
./nginx/setup-ssl.sh

# 或者手动申请证书
sudo certbot --nginx -d test.arctanp.top
```

### 5. 更新 SSL 证书路径

证书申请成功后，更新 `nginx/registry.conf` 中的 SSL 证书路径：

```nginx
ssl_certificate /etc/letsencrypt/live/test.arctanp.top/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/test.arctanp.top/privkey.pem;
```

### 6. 启动服务

```bash
# 启动 Docker Registry
cd /path/to/my_docker_registry
./registry &

# 重启 Nginx
sudo systemctl restart nginx

# 设置开机自启
sudo systemctl enable nginx
```

## 配置详解

### Nginx 配置特性

- **SSL/TLS**: 强制 HTTPS，HTTP 自动重定向
- **客户端上传**: 支持无限制的请求体大小 (`client_max_body_size 0`)
- **超时设置**: 适合大文件传输的超时配置
- **代理优化**: 禁用缓冲以支持流式传输
- **Docker 头**: 正确传递 Docker Registry 特定的 HTTP 头

### 关键配置项

```nginx
# 允许大文件上传 (Docker layers 可能很大)
client_max_body_size 0;

# 支持流式传输
proxy_buffering off;
proxy_request_buffering off;

# Docker Registry 特定头
proxy_pass_header Docker-Content-Digest;
proxy_pass_header Docker-Distribution-Api-Version;
```

## 使用方法

### 1. 配置 Docker 客户端

```bash
# 登录到私有 registry
docker login test.arctanp.top

# 推送镜像
docker tag myimage:latest test.arctanp.top/myimage:latest
docker push test.arctanp.top/myimage:latest

# 拉取镜像
docker pull test.arctanp.top/myimage:latest
```

### 2. API 访问

```bash
# 检查 registry 状态
curl https://test.arctanp.top/v2/

# 列出仓库
curl https://test.arctanp.top/v2/_catalog

# 查看镜像标签
curl https://test.arctanp.top/v2/myimage/tags/list
```

## 故障排除

### 1. 检查服务状态

```bash
# 检查 Docker Registry
curl http://localhost:5000/v2/

# 检查 Nginx 状态
sudo systemctl status nginx

# 查看 Nginx 日志
sudo tail -f /var/log/nginx/registry.error.log
sudo tail -f /var/log/nginx/registry.access.log
```

### 2. 常见问题

**问题**: SSL 证书错误
**解决**: 检查域名解析，重新申请证书

**问题**: 502 Bad Gateway
**解决**: 确认 Docker Registry 在 5000 端口运行

**问题**: 413 Request Entity Too Large
**解决**: 确认 `client_max_body_size 0;` 配置正确

### 3. 证书自动续期

Let's Encrypt 证书有效期 90 天，自动续期已配置：

```bash
# 检查 cron 任务
sudo crontab -l | grep certbot

# 手动测试续期
sudo certbot renew --dry-run
```

## 安全建议

1. **防火墙**: 只开放 80, 443 端口
2. **访问控制**: 根据需要配置 IP 白名单
3. **监控**: 设置日志监控和告警
4. **备份**: 定期备份 registry 数据和配置

## 监控

Nginx 状态页面：
```bash
curl https://test.arctanp.top/nginx_status
```

输出示例：
```
Active connections: 2
server accepts handled requests
 1000 1000 2000
Reading: 0 Writing: 1 Waiting: 1
```
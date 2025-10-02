#!/bin/bash

# Docker Registry + Cloudflare 自动部署脚本
set -e

echo "=== Docker Registry + Cloudflare 部署脚本 ==="

# 检查域名参数
DOMAIN=${1:-test.arctanp.top}
REGISTRY_PORT=${2:-5000}

echo "域名: $DOMAIN"
echo "Registry 端口: $REGISTRY_PORT"

# 检查是否以 root 权限运行
if [[ $EUID -ne 0 ]]; then
   echo "请以 root 权限运行此脚本 (sudo ./deploy-cloudflare.sh)"
   exit 1
fi

echo "1. 检查依赖..."
# 检查 nginx
if ! command -v nginx &> /dev/null; then
    echo "安装 nginx..."
    apt update && apt install -y nginx
fi

echo "2. 检查 DNS 解析..."
RESOLVED_IP=$(dig +short $DOMAIN 2>/dev/null | head -n1)
if [ -z "$RESOLVED_IP" ]; then
    echo "❌ 域名 $DOMAIN 未解析，请先配置 Cloudflare DNS"
    echo "运行 ./nginx/cloudflare-setup.sh 查看配置指南"
    exit 1
fi

echo "✅ DNS 解析: $DOMAIN -> $RESOLVED_IP"

# 检查是否使用 Cloudflare 代理
SERVER_IP=$(curl -s ifconfig.me)
if [ "$RESOLVED_IP" = "$SERVER_IP" ]; then
    echo "🔘 检测到仅 DNS 模式"
    USE_CLOUDFLARE_PROXY=false
else
    echo "🟠 检测到 Cloudflare 代理模式"
    USE_CLOUDFLARE_PROXY=true
fi

echo "3. 备份现有配置..."
if [ -f "/etc/nginx/sites-enabled/registry.conf" ]; then
    cp /etc/nginx/sites-enabled/registry.conf /etc/nginx/sites-enabled/registry.conf.backup.$(date +%Y%m%d%H%M%S)
fi

echo "4. 部署 Nginx 配置..."

if [ "$USE_CLOUDFLARE_PROXY" = true ]; then
    echo "使用 Cloudflare 代理配置..."
    # 更新配置文件中的域名
    sed "s/test\.arctanp\.top/$DOMAIN/g" nginx/registry-cloudflare.conf > /tmp/registry.conf
    sed -i "s/127\.0\.0\.1:5000/127.0.0.1:$REGISTRY_PORT/g" /tmp/registry.conf
    
    # 生成 Cloudflare Origin Certificate (自签名，供参考)
    echo "📋 请在 Cloudflare 控制台生成 Origin Certificate:"
    echo "   1. 访问 SSL/TLS -> Origin Server"
    echo "   2. 创建 Origin Certificate"
    echo "   3. 下载证书并保存为:"
    echo "      /etc/ssl/certs/cloudflare-origin.pem"
    echo "      /etc/ssl/private/cloudflare-origin.key"
    echo ""
else
    echo "使用标准配置..."
    # 使用标准配置文件
    sed "s/test\.arctanp\.top/$DOMAIN/g" nginx/registry.conf > /tmp/registry.conf
    sed -i "s/127\.0\.0\.1:5000/127.0.0.1:$REGISTRY_PORT/g" /tmp/registry.conf
fi

# 复制配置文件
cp /tmp/registry.conf /etc/nginx/sites-available/registry.conf
ln -sf /etc/nginx/sites-available/registry.conf /etc/nginx/sites-enabled/registry.conf

# 删除默认站点
if [ -f "/etc/nginx/sites-enabled/default" ]; then
    rm -f /etc/nginx/sites-enabled/default
fi

echo "5. 测试 Nginx 配置..."
nginx -t

echo "6. 重新加载 Nginx..."
systemctl reload nginx

echo "7. 检查 Docker Registry 服务..."
if ! curl -f http://127.0.0.1:$REGISTRY_PORT/v2/ &> /dev/null; then
    echo "⚠️  Docker Registry 未在端口 $REGISTRY_PORT 运行"
    echo "请手动启动: ./registry &"
fi

if [ "$USE_CLOUDFLARE_PROXY" = false ]; then
    echo "8. 申请 Let's Encrypt SSL 证书..."
    if command -v certbot &> /dev/null; then
        certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@$DOMAIN || echo "SSL 证书申请失败，请手动申请"
    else
        echo "安装 certbot..."
        apt install -y certbot python3-certbot-nginx
        certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@$DOMAIN || echo "SSL 证书申请失败，请手动申请"
    fi
    
    echo "9. 设置证书自动续期..."
    crontab -l 2>/dev/null | grep -q 'certbot renew' || (crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | crontab -
else
    echo "8. Cloudflare SSL 配置..."
    echo "   请确保已正确配置 Origin Certificate"
    echo "   或在 Cloudflare 中设置 SSL 模式为 'Flexible'"
fi

echo "10. 启用服务开机自启..."
systemctl enable nginx

echo ""
echo "=== 部署完成! ==="
echo "访问地址: https://$DOMAIN"
echo "API 端点: https://$DOMAIN/v2/"

if [ "$USE_CLOUDFLARE_PROXY" = true ]; then
    echo "代理模式: Cloudflare 已代理"
    echo "SSL 提供: Cloudflare Universal SSL"
else
    echo "代理模式: 仅 DNS"
    echo "SSL 提供: Let's Encrypt"
fi

echo ""
echo "测试命令:"
echo "  curl https://$DOMAIN/v2/"
echo "  docker login $DOMAIN"
echo ""
echo "日志位置:"
echo "  Nginx: /var/log/nginx/registry.*.log"
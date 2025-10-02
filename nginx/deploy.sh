# Docker Registry 快速部署脚本
#!/bin/bash

set -e

echo "=== Docker Registry + Nginx 自动部署脚本 ==="

# 检查域名参数
DOMAIN=${1:-test.arctanp.top}
REGISTRY_PORT=${2:-5000}

echo "域名: $DOMAIN"
echo "Registry 端口: $REGISTRY_PORT"

# 检查是否以 root 权限运行
if [[ $EUID -ne 0 ]]; then
   echo "请以 root 权限运行此脚本 (sudo ./deploy.sh)"
   exit 1
fi

echo "1. 检查依赖..."
# 检查 nginx
if ! command -v nginx &> /dev/null; then
    echo "安装 nginx..."
    apt update && apt install -y nginx
fi

# 检查 certbot
if ! command -v certbot &> /dev/null; then
    echo "安装 certbot..."
    apt install -y certbot python3-certbot-nginx
fi

echo "2. 备份现有配置..."
if [ -f "/etc/nginx/sites-enabled/registry.conf" ]; then
    cp /etc/nginx/sites-enabled/registry.conf /etc/nginx/sites-enabled/registry.conf.backup.$(date +%Y%m%d%H%M%S)
fi

echo "3. 部署 Nginx 配置..."
# 更新配置文件中的域名
sed "s/test\.arctanp\.top/$DOMAIN/g" nginx/registry.conf > /tmp/registry.conf
sed -i "s/127\.0\.0\.1:5000/127.0.0.1:$REGISTRY_PORT/g" /tmp/registry.conf

# 复制配置文件
cp /tmp/registry.conf /etc/nginx/sites-available/registry.conf
ln -sf /etc/nginx/sites-available/registry.conf /etc/nginx/sites-enabled/registry.conf

# 删除默认站点 (如果存在)
if [ -f "/etc/nginx/sites-enabled/default" ]; then
    rm -f /etc/nginx/sites-enabled/default
fi

echo "4. 测试 Nginx 配置..."
nginx -t

echo "5. 重新加载 Nginx..."
systemctl reload nginx

echo "6. 检查 Docker Registry 服务..."
if ! curl -f http://127.0.0.1:$REGISTRY_PORT/v2/ &> /dev/null; then
    echo "警告: Docker Registry 未在端口 $REGISTRY_PORT 运行"
    echo "请手动启动: ./registry &"
fi

echo "7. 申请 SSL 证书..."
# 检查域名解析
if ! nslookup $DOMAIN | grep -q "$(curl -s ifconfig.me)"; then
    echo "警告: 域名 $DOMAIN 可能未正确解析到当前服务器"
    echo "请确保域名解析后再申请 SSL 证书"
else
    echo "申请 SSL 证书..."
    certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@$DOMAIN || echo "SSL 证书申请失败，请手动申请"
fi

echo "8. 设置证书自动续期..."
crontab -l 2>/dev/null | grep -q 'certbot renew' || (crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | crontab -

echo "9. 启用服务开机自启..."
systemctl enable nginx

echo ""
echo "=== 部署完成! ==="
echo "访问地址: https://$DOMAIN"
echo "API 端点: https://$DOMAIN/v2/"
echo ""
echo "测试命令:"
echo "  curl https://$DOMAIN/v2/"
echo "  docker login $DOMAIN"
echo ""
echo "日志位置:"
echo "  Nginx: /var/log/nginx/registry.*.log"
echo "  SSL: /var/log/letsencrypt/letsencrypt.log"
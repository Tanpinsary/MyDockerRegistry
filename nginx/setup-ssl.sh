# Let's Encrypt SSL 证书自动获取脚本
#!/bin/bash

# 使用 Certbot 为 test.arctanp.top 申请免费的 SSL 证书

# 安装 certbot (如果未安装)
# Ubuntu/Debian: sudo apt update && sudo apt install certbot python3-certbot-nginx
# CentOS/RHEL: sudo yum install certbot python3-certbot-nginx

# 获取证书 (需要确保域名已解析到当前服务器)
sudo certbot --nginx -d test.arctanp.top

# 设置自动续期
sudo crontab -l | grep -q 'certbot renew' || (crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | sudo crontab -

echo "SSL 证书配置完成！"
echo "证书将存储在: /etc/letsencrypt/live/test.arctanp.top/"
echo "nginx 配置需要更新证书路径为:"
echo "  ssl_certificate /etc/letsencrypt/live/test.arctanp.top/fullchain.pem;"
echo "  ssl_certificate_key /etc/letsencrypt/live/test.arctanp.top/privkey.pem;"
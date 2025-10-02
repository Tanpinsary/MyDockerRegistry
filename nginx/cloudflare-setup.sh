#!/bin/bash

# Cloudflare DNS 配置指南脚本
# 为 test.arctanp.top 配置 DNS 解析

set -e

echo "=== Cloudflare DNS 配置指南 ==="
echo ""

# 获取服务器公网 IP
echo "🔍 检测服务器信息..."
SERVER_IP=$(curl -s ifconfig.me)
echo "服务器公网 IP: $SERVER_IP"
echo ""

echo "📋 Cloudflare DNS 配置步骤："
echo ""

echo "1. 登录 Cloudflare 控制台:"
echo "   https://dash.cloudflare.com/"
echo ""

echo "2. 选择域名 'arctanp.top'"
echo ""

echo "3. 进入 DNS 管理页面"
echo ""

echo "4. 添加 A 记录:"
echo "   ┌─────────────────────────────────────┐"
echo "   │ 类型: A                             │"
echo "   │ 名称: test                          │"
echo "   │ IPv4 地址: $SERVER_IP          │"
echo "   │ TTL: Auto                           │"
echo "   │ 代理状态: 🟠 已代理 (推荐)          │"
echo "   └─────────────────────────────────────┘"
echo ""

echo "5. 代理状态选择："
echo "   🟠 已代理 (橙色云朵) - 启用 CDN + 安全防护"
echo "   🔘 仅 DNS (灰色云朵) - 直接解析到服务器"
echo ""

echo "💡 推荐使用 '已代理' 模式，获得以下优势："
echo "   ✅ Cloudflare CDN 加速"
echo "   ✅ DDoS 防护"
echo "   ✅ 免费 SSL 证书"
echo "   ✅ 隐藏真实服务器 IP"
echo ""

echo "6. 等待 DNS 生效 (2-5分钟)"
echo ""

echo "7. 验证 DNS 解析:"
echo "   nslookup test.arctanp.top"
echo "   dig test.arctanp.top"
echo ""

echo "🔧 配置完成后运行部署脚本:"
if [ -f "nginx/deploy-cloudflare.sh" ]; then
    echo "   sudo ./nginx/deploy-cloudflare.sh"
else
    echo "   sudo ./nginx/deploy.sh"
fi
echo ""

# 实时检查 DNS 解析状态
echo "🕐 实时检查 DNS 解析状态..."
echo "按 Ctrl+C 停止检查"
echo ""

while true; do
    RESOLVED_IP=$(dig +short test.arctanp.top 2>/dev/null | head -n1)
    CURRENT_TIME=$(date '+%H:%M:%S')
    
    if [ -n "$RESOLVED_IP" ]; then
        if [[ "$RESOLVED_IP" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            if [ "$RESOLVED_IP" = "$SERVER_IP" ]; then
                echo "[$CURRENT_TIME] ✅ DNS 已生效: test.arctanp.top -> $RESOLVED_IP (直接解析)"
            else
                echo "[$CURRENT_TIME] 🟠 DNS 已生效: test.arctanp.top -> $RESOLVED_IP (Cloudflare 代理)"
            fi
        else
            echo "[$CURRENT_TIME] ⏳ DNS 解析中..."
        fi
    else
        echo "[$CURRENT_TIME] ❌ DNS 未解析到 IP 地址"
    fi
    
    sleep 10
done
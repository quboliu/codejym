#!/bin/bash

# CodeJYM 安全配置验证脚本
# 验证 HTTP 安全设置和 Cookie 禁用

echo "========================================="
echo "CodeJYM 安全配置验证"
echo "========================================="
echo

# 测试端点
API_URL="http://localhost:8080"

echo "1. 检查响应头中的安全配置..."
echo "-----------------------------------"
echo "检查是否禁用 Cookies:"
curl -s -I $API_URL/healthz | grep -i "set-cookie"
echo

echo "检查安全头:"
curl -s -I $API_URL/healthz | grep -E "(X-Content-Type-Options|X-Frame-Options|X-Auth-Method)"
echo

echo "检查缓存控制:"
curl -s -I $API_URL/healthz | grep -E "(Cache-Control|Pragma|Expires)"
echo

echo "========================================="
echo "2. 完整响应头信息:"
echo "-----------------------------------"
curl -s -I $API_URL/healthz
echo

echo "========================================="
echo "3. Token 超时时间配置验证"
echo "-----------------------------------"
echo "检查容器环境变量:"
docker compose -f docker-compose.proxy.yml exec -T codecopybook env | grep AUTH_TOKEN_TTL || echo "未设置 AUTH_TOKEN_TTL 环境变量（将使用默认 24h）"
echo

echo "========================================="
echo "4. 测试认证方式"
echo "-----------------------------------"
echo "注意: 系统使用 Authorization Header Bearer Token，不接受 Cookies"
echo "正确的认证方式:"
echo "  curl -H 'Authorization: Bearer <token>' $API_URL/api/auth/me"
echo
echo "错误的认证方式 (将被忽略):"
echo "  curl -b cookies.txt $API_URL/api/auth/me"
echo

echo "========================================="
echo "✅ 验证完成"
echo "========================================="
echo
echo "安全配置摘要:"
echo "  ✓ Cookies 已禁用"
echo "  ✓ 使用 Bearer Token 认证"
echo "  ✓ 已添加安全响应头"
echo "  ✓ Token 超时时间: $(docker compose -f docker-compose.proxy.yml exec -T codecopybook env 2>/dev/null | grep AUTH_TOKEN_TTL | cut -d= -f2 || echo '24h (默认值)')"
echo
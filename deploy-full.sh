#!/bin/bash

set -e

echo "================================"
echo "  CodeJYM 全功能一键部署脚本"
echo "  （包含域名访问功能）"
echo "================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装${NC}"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose 是否可用
if ! command -v docker compose &> /dev/null; then
    echo -e "${RED}错误: Docker Compose 未安装${NC}"
    echo "请安装 Docker Compose 或升级 Docker 到包含 compose 插件的版本"
    exit 1
fi

echo -e "${GREEN}✓${NC} Docker 和 Docker Compose 检查通过"
echo ""

# 停止并清理旧容器
echo -e "${YELLOW}[1/6] 清理旧容器...${NC}"
docker compose down -v 2>/dev/null || true
echo -e "${GREEN}✓${NC} 清理完成"
echo ""

# 构建镜像
echo -e "${YELLOW}[2/6] 构建 Docker 镜像...${NC}"
echo "这可能需要几分钟时间，请耐心等待..."
docker compose build
if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 镜像构建失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} 镜像构建完成"
echo ""

# 启动完整服务（包含反向代理）
echo -e "${YELLOW}[3/6] 启动完整服务（PostgreSQL + 应用 + Nginx）...${NC}"
docker compose -f docker-compose.proxy.yml up -d
if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 服务启动失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} 服务启动成功"
echo ""

# 等待服务就绪
echo -e "${YELLOW}[4/6] 等待服务就绪...${NC}"
echo "正在检查 PostgreSQL 和应用服务启动状态..."

# 检查 PostgreSQL
MAX_RETRIES=60
RETRY_COUNT=0
echo "正在检查 PostgreSQL..."
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker compose -f docker-compose.proxy.yml ps | grep -q "postgres.*Up.*healthy"; then
        echo -e "${GREEN}✓${NC} PostgreSQL 服务健康"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $((RETRY_COUNT % 5)) -eq 0 ]; then
        echo -n "."
    fi
    sleep 1
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo ""
    echo -e "${YELLOW}警告: PostgreSQL 健康检查超时，但服务可能仍在启动中${NC}"
    sleep 5
else
    sleep 3  # 额外等待确保应用完全启动
fi

# 检查应用服务
RETRY_COUNT=0
echo "正在检查应用服务..."
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 | grep -q "200\|304"; then
        echo -e "${GREEN}✓${NC} 应用服务可访问"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $((RETRY_COUNT % 5)) -eq 0 ]; then
        echo -n "."
    fi
    sleep 1
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo ""
    echo -e "${YELLOW}警告: 应用服务响应超时${NC}"
    docker compose -f docker-compose.proxy.yml ps
    echo ""
    echo "应用日志:"
    docker compose -f docker-compose.proxy.yml logs --tail=20 codecopybook
    echo ""
else
    echo ""
    echo -e "${GREEN}✓${NC} 所有服务就绪"
fi
echo ""

# 显示服务信息
echo -e "${YELLOW}[5/6] 部署完成！${NC}"
echo ""
echo "================================"
echo "  服务信息"
echo "================================"
echo ""
echo -e "${BLUE}本地访问:${NC}       http://localhost:8080"
echo -e "${BLUE}域名访问:${NC}       http://jiezispace.com"
echo -e "${BLUE}www 访问:${NC}      http://www.jiezispace.com"
echo -e "${BLUE}后端 API:${NC}      http://localhost:8080/api"
echo -e "${BLUE}PostgreSQL:${NC}    localhost:5432"
echo ""
echo "================================"
echo "  数据库配置"
echo "================================"
echo ""
echo -e "${BLUE}主机:${NC}           postgres"
echo -e "${BLUE}端口:${NC}           5432"
echo -e "${BLUE}数据库:${NC}         codecopybook"
echo -e "${BLUE}用户名:${NC}         codecopy"
echo -e "${BLUE}密码:${NC}           codecopy123"
echo ""
echo "================================"
echo "  容器状态"
echo "================================"
echo ""
docker compose -f docker-compose.proxy.yml ps
echo ""

# 显示管理命令
echo -e "${YELLOW}[6/6] 管理命令${NC}"
echo ""
echo -e "${BLUE}查看服务状态:${NC}    docker compose -f docker-compose.proxy.yml ps"
echo -e "${BLUE}查看应用日志:${NC}    docker compose -f docker-compose.proxy.yml logs -f codecopybook"
echo -e "${BLUE}查看数据库日志:${NC}  docker compose -f docker-compose.proxy.yml logs -f postgres"
echo -e "${BLUE}查看Nginx日志:${NC}   docker compose -f docker-compose.proxy.yml logs -f nginx"
echo -e "${BLUE}重启所有服务:${NC}    docker compose -f docker-compose.proxy.yml restart"
echo -e "${BLUE}停止所有服务:${NC}    docker compose -f docker-compose.proxy.yml down"
echo -e "${BLUE}完全清理:${NC}        docker compose -f docker-compose.proxy.yml down -v"
echo ""
echo "================================"
echo ""

# 验证部署
echo -e "${GREEN}正在验证部署...${NC}"
echo ""

# 测试本地访问
if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} 本地访问正常 (http://localhost:8080)"
else
    echo -e "${RED}✗${NC} 本地访问失败"
fi

# 测试 API 访问
if curl -s http://localhost:8080/api/auth/me > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} API 访问正常 (http://localhost:8080/api)"
else
    echo -e "${RED}✗${NC} API 访问失败"
fi

# 检查服务状态
RUNNING_SERVICES=$(docker compose -f docker-compose.proxy.yml ps --services --filter "status=running" | wc -l)
TOTAL_SERVICES=$(docker compose -f docker-compose.proxy.yml ps --services | wc -l)

echo -e "${GREEN}✓${NC} 运行服务数量: $RUNNING_SERVICES / $TOTAL_SERVICES"
echo ""

echo "================================"
echo ""
echo -e "${GREEN}🎉 部署成功！${NC}"
echo ""
echo -e "${BLUE}访问地址：${NC}"
echo "  • 本地: http://localhost:8080"
echo "  • 域名: http://jiezispace.com"
echo ""
echo -e "${BLUE}现在你可以：${NC}"
echo "  1. 在浏览器中打开 http://jiezispace.com"
echo "  2. 或访问 http://localhost:8080"
echo ""
echo "================================"
echo ""

#!/bin/bash

set -e

echo "================================"
echo "  CodeJYM 一键部署脚本"
echo "================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
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
echo -e "${YELLOW}[1/5] 清理旧容器...${NC}"
docker compose down -v 2>/dev/null || true
echo -e "${GREEN}✓${NC} 清理完成"
echo ""

# 构建镜像
echo -e "${YELLOW}[2/5] 构建 Docker 镜像...${NC}"
echo "这可能需要几分钟时间，请耐心等待..."
docker compose build
if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 镜像构建失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} 镜像构建完成"
echo ""

# 启动服务
echo -e "${YELLOW}[3/5] 启动服务...${NC}"
docker compose up -d
if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 服务启动失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} 服务启动成功"
echo ""

# 等待服务就绪
echo -e "${YELLOW}[4/5] 等待服务就绪...${NC}"
echo "正在等待 PostgreSQL 和应用服务启动..."
sleep 5

# 检查服务状态
MAX_RETRIES=60
RETRY_COUNT=0
echo "正在检查 PostgreSQL 健康状态..."
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker compose ps | grep -q "postgres.*Up.*healthy"; then
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
    echo "等待额外时间以确保服务完全启动..."
    sleep 5
fi

# 检查应用是否可访问
echo "正在检查应用服务状态..."
RETRY_COUNT=0
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
    echo -e "${YELLOW}警告: 应用服务响应超时，尝试查看详细状态...${NC}"
    docker compose ps
    echo ""
    echo "应用日志 (最近20行):"
    docker compose logs --tail=20 codecopybook
    echo ""
    echo -e "${YELLOW}尝试强制继续部署...${NC}"
else
    echo ""
    echo -e "${GREEN}✓${NC} 服务就绪"
fi
echo ""

# 显示服务信息
echo -e "${YELLOW}[5/5] 部署完成！${NC}"
echo ""
echo "================================"
echo "  服务信息"
echo "================================"
echo ""
echo -e "${GREEN}前端地址:${NC}     http://localhost:8080"
echo -e "${GREEN}后端 API:${NC}     http://localhost:8080/api"
echo -e "${GREEN}PostgreSQL:${NC}   localhost:5432"
echo ""
echo "================================"
echo "  数据库配置"
echo "================================"
echo ""
echo -e "${GREEN}用户名:${NC} codecopy"
echo -e "${GREEN}密码:${NC} codecopy123"
echo -e "${GREEN}数据库:${NC} codecopybook"
echo ""
echo "================================"
echo "  管理命令"
echo "================================"
echo ""
echo "查看服务状态:    docker compose ps"
echo "查看应用日志:    docker compose logs -f codecopybook"
echo "查看数据库日志:  docker compose logs -f postgres"
echo "重启服务:        docker compose restart"
echo "停止服务:        docker compose down"
echo "完全清理:        docker compose down -v"
echo ""
echo "================================"
echo ""

#!/bin/bash

set -e

# 切换到项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

echo "================================"
echo "  CodeJYM 本地部署脚本"
echo "  （仅本地开发使用）"
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
echo "  工作目录: $PROJECT_ROOT"
echo ""

# 停止并清理旧容器
echo -e "${YELLOW}[1/6] 停止并清理旧容器...${NC}"
docker compose -f config/docker-compose.yml down 2>/dev/null || true
echo -e "${GREEN}✓${NC} 旧容器已停止"
echo ""

# 清理旧镜像
echo -e "${YELLOW}[2/6] 清理旧镜像（确保重新构建）...${NC}"
docker rmi codecopybook:local 2>/dev/null || echo "  无旧镜像需要清理"
echo -e "${GREEN}✓${NC} 清理完成"
echo ""

# 构建镜像
echo -e "${YELLOW}[3/6] 构建 Docker 镜像（前端 + 后端）...${NC}"
echo "  这将会："
echo "  • 构建最新的前端（Vue 3 + TypeScript + Vite）"
echo "  • 构建最新的后端（Go 1.24）"
echo "  • 打包成生产镜像"
echo ""
echo "  这可能需要几分钟时间，请耐心等待..."
echo ""
docker compose -f config/docker-compose.yml build --no-cache
if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 镜像构建失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} 镜像构建完成"
echo ""

# 启动服务
echo -e "${YELLOW}[4/6] 启动服务...${NC}"
echo "  • PostgreSQL 数据库"
echo "  • CodeJYM 应用（前端+后端）"
echo ""
docker compose -f config/docker-compose.yml up -d
if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 服务启动失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} 服务启动成功"
echo ""

# 等待服务就绪
echo -e "${YELLOW}[5/6] 等待服务就绪...${NC}"

# 检查 PostgreSQL
MAX_RETRIES=60
RETRY_COUNT=0
echo "正在检查 PostgreSQL..."
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker compose -f config/docker-compose.yml ps postgres | grep -q "Up.*healthy"; then
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
    echo -e "${YELLOW}警告: PostgreSQL 健康检查超时${NC}"
    sleep 5
else
    sleep 3
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
    docker compose -f config/docker-compose.yml ps
    echo ""
    echo "应用日志:"
    docker compose -f config/docker-compose.yml logs --tail=20 codecopybook
    echo ""
else
    echo ""
    echo -e "${GREEN}✓${NC} 所有服务就绪"
fi
echo ""

# 验证部署
echo -e "${YELLOW}[6/6] 验证部署...${NC}"
echo ""

if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} 本地访问正常"
else
    echo -e "${RED}✗${NC} 本地访问失败"
fi

if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} API 访问正常"
else
    echo -e "${RED}✗${NC} API 访问失败"
fi

RUNNING_SERVICES=$(docker compose -f config/docker-compose.yml ps --services --filter "status=running" | wc -l)
TOTAL_SERVICES=$(docker compose -f config/docker-compose.yml ps --services | wc -l)

echo -e "${GREEN}✓${NC} 运行服务数量: $RUNNING_SERVICES / $TOTAL_SERVICES"
echo ""

# 显示服务信息
echo "================================"
echo "  📱 访问地址"
echo "================================"
echo ""
echo -e "${BLUE}本地访问:${NC}       http://localhost:8080"
echo ""
echo "================================"
echo "  🔌 API 端点"
echo "================================"
echo ""
echo -e "${BLUE}健康检查:${NC}       http://localhost:8080/api/health"
echo -e "${BLUE}用户注册:${NC}       http://localhost:8080/api/auth/signup"
echo -e "${BLUE}用户登录:${NC}       http://localhost:8080/api/auth/login"
echo -e "${BLUE}训练组列表:${NC}     http://localhost:8080/api/assets"
echo ""
echo "================================"
echo "  🗄️  数据库配置"
echo "================================"
echo ""
echo -e "${BLUE}主机:${NC}           postgres (容器内) / localhost:5432 (主机)"
echo -e "${BLUE}端口:${NC}           5432"
echo -e "${BLUE}数据库:${NC}         codecopybook"
echo -e "${BLUE}用户名:${NC}         codecopy"
echo -e "${BLUE}密码:${NC}           codecopy"
echo ""
echo "================================"
echo "  📊 容器状态"
echo "================================"
echo ""
docker compose -f config/docker-compose.yml ps
echo ""

# 显示管理命令
echo "================================"
echo "  🛠️  管理命令"
echo "================================"
echo ""
echo -e "${BLUE}查看服务状态:${NC}"
echo "  docker compose -f config/docker-compose.yml ps"
echo ""
echo -e "${BLUE}查看日志:${NC}"
echo "  docker compose -f config/docker-compose.yml logs -f codecopybook  # 应用日志"
echo "  docker compose -f config/docker-compose.yml logs -f postgres      # 数据库日志"
echo ""
echo -e "${BLUE}重启服务:${NC}"
echo "  docker compose -f config/docker-compose.yml restart               # 重启所有"
echo "  docker compose -f config/docker-compose.yml restart codecopybook  # 仅重启应用"
echo ""
echo -e "${BLUE}停止服务:${NC}"
echo "  docker compose -f config/docker-compose.yml stop                  # 停止所有"
echo "  docker compose -f config/docker-compose.yml down                  # 停止并删除容器"
echo "  docker compose -f config/docker-compose.yml down -v               # 完全清理（含数据）"
echo ""
echo -e "${BLUE}更新部署:${NC}"
echo "  ./deploy.sh                          # 重新运行此脚本"
echo ""
echo "================================"
echo ""

echo -e "${GREEN}🎉 部署成功！${NC}"
echo ""
echo -e "${BLUE}现在你可以：${NC}"
echo ""
echo "  1. 在浏览器中打开："
echo "     • http://localhost:8080"
echo ""
echo "  2. 注册新用户并体验新功能："
echo "     • ✨ 默认训练组自动创建"
echo "     • 📝 重命名和删除训练组"
echo "     • 📁 创建文件夹"
echo "     • 🗂️  文件树右键菜单"
echo "     • 📂 收起/展开训练组和文件列表"
echo ""
echo "  3. 查看实时日志："
echo "     docker compose -f config/docker-compose.yml logs -f"
echo ""
echo "================================"
echo ""


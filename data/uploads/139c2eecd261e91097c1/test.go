#!/usr/bin/env bash

# Docker 常用命令清单示例脚本（仅打印常用命令）

echo "=== Docker 基础 ==="
echo "docker --version"
echo "docker info"

echo "=== 镜像相关 ==="
echo "docker images"
echo "docker search <keyword>"
echo "docker pull <image>:<tag>"
echo "docker rmi <image_id>"
echo "docker tag <image_id> <new_name>:<tag>"

echo "=== 容器管理 ==="
echo "docker ps"
echo "docker ps -a"
echo "docker start <container_id>"
echo "docker stop <container_id>"
echo "docker restart <container_id>"
echo "docker rm <container_id>"

echo "=== 运行容器 ==="
echo "docker run -it <image> bash"
echo "docker run -d <image>"
echo "docker run -p 8080:80 <image>"
echo "docker run -v /host:/container <image>"

echo "=== 容器内部操作 ==="
echo "docker exec -it <container_id> bash"
echo "docker exec -it <container_id> sh"

echo "=== 日志与检查 ==="
echo "docker logs <container_id>"
echo "docker logs -f <container_id>"
echo "docker inspect <container_id>"

echo "=== 导入导出镜像 ==="
echo "docker save -o image.tar <image>"
echo "docker load -i image.tar"

echo "=== 导入导出容器 ==="
echo "docker export -o container.tar <container>"
echo "cat container.tar | docker import - <image>:tag"

echo "=== 清理 ==="
echo "docker system prune"
echo "docker system prune -a"

echo "=== 构建镜像 ==="
echo "docker build -t myapp:latest ."

echo "=== 网络相关 ==="
echo "docker network ls"
echo "docker network create <name>"
echo "docker network inspect <name>"

echo "=== 资源监控 ==="
echo "docker stats"

echo "=== 进入容器文件变化 ==="
echo "docker diff <container_id>"

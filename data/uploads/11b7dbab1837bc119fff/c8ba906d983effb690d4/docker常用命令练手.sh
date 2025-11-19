#!/usr/bin/env bash
# ============================================================
# Docker 常用场景命令清单（Cheat Sheet）
# 保存为：docker-cheatsheet.sh
# 说明：
#   - 这是“命令 + 注释”的清单脚本，并不是给你一次性跑完的。
#   - 建议：将本文件当成笔记，按需复制某一段命令到终端执行。
#   - 若要统一修改默认镜像、容器名称，可以改下面的变量。
# ============================================================

# 一些常用变量（按需修改）
IMAGE_NAME="nginx:latest"
CONTAINER_NAME="my-nginx"
NETWORK_NAME="my-network"
VOLUME_NAME="my-data"
COMPOSE_FILE="docker-compose.yml"

# ------------------------------------------------------------
# 0. 基础信息与帮助
# ------------------------------------------------------------

# 场景：确认 Docker 是否安装成功、版本信息
docker version           # 显示客户端 + 服务端版本
docker info              # 显示 Docker 详细信息（存储驱动、Cgroup、镜像数等）

# 场景：不知道某个子命令怎么用
docker --help            # 查看 Docker 顶层帮助
docker ps --help         # 查看 docker ps 的帮助
docker run --help        # 查看 docker run 的帮助

# ------------------------------------------------------------
# 1. 镜像管理：拉取、查看、删除
# ------------------------------------------------------------

# 场景：拉取一个镜像（例如 nginx）
docker pull "${IMAGE_NAME}"

# 场景：列出本地所有镜像
docker images            # 等价于 docker image ls

# 场景：搜索镜像（从 Docker Hub）
docker search nginx      # 会返回匹配 nginx 的公共镜像列表

# 场景：删除镜像
# 注意：删除前需要确保没有容器在使用该镜像，否则会失败
docker rmi nginx:latest          # 删除指定标签的镜像
docker rmi IMAGE_ID              # 也可以用 IMAGE ID 删除

# 场景：清理悬挂镜像（<none> 标签）
docker image prune                # 删除未被使用的悬挂镜像（谨慎）

# ------------------------------------------------------------
# 2. 容器生命周期：创建、启动、停止、删除
# ------------------------------------------------------------

# 场景：以交互方式运行一个临时容器（用完即删）
docker run --rm -it alpine:latest sh
# 说明：
#   --rm   退出后自动删除容器
#   -it   交互式终端
#   alpine:latest   使用 alpine 镜像
#   sh    启动容器后执行的命令

# 场景：后台启动一个 nginx 容器并映射端口
docker run -d \
  --name "${CONTAINER_NAME}" \
  -p 8080:80 \
  "${IMAGE_NAME}"
# 说明：
#   -d       后台运行
#   --name   为容器起一个名字
#   -p 8080:80  将宿主机的 8080 端口映射到容器的 80 端口

# 场景：查看正在运行的容器
docker ps                      # 默认只显示运行中容器
docker ps -a                   # 显示所有容器（包括已停止）

# 场景：停止、启动、重启容器
docker stop "${CONTAINER_NAME}"
docker start "${CONTAINER_NAME}"
docker restart "${CONTAINER_NAME}"

# 场景：删除容器
docker rm "${CONTAINER_NAME}"      # 删除已停止的容器
docker rm -f "${CONTAINER_NAME}"   # 强制删除（会先 stop 再删，谨慎）

# ------------------------------------------------------------
# 3. 容器日志与排错
# ------------------------------------------------------------

# 场景：查看容器日志（实时滚动）
docker logs "${CONTAINER_NAME}"        # 查看当前日志
docker logs -f "${CONTAINER_NAME}"     # 持续滚动日志（类似 tail -f）
docker logs -f --tail=100 "${CONTAINER_NAME}"  # 只看最后 100 行并跟随

# 场景：查看容器详细信息（包括挂载、网络、环境变量等）
docker inspect "${CONTAINER_NAME}"     # 输出是 JSON

# 场景：查看容器资源占用（CPU、内存、网络、IO）
docker stats                           # 类似 top，默认显示所有运行中容器
docker stats "${CONTAINER_NAME}"

# ------------------------------------------------------------
# 4. 进入容器、执行命令
# ------------------------------------------------------------

# 场景：进入一个正在运行的容器执行 bash
docker exec -it "${CONTAINER_NAME}" bash
# 若容器中没有 bash（比如基于 alpine），可以用 sh
# docker exec -it "${CONTAINER_NAME}" sh

# 场景：在容器中执行一次性命令
docker exec "${CONTAINER_NAME}" ls -al /   # 非交互执行命令

# ------------------------------------------------------------
# 5. 数据持久化：卷和绑定挂载
# ------------------------------------------------------------

# 场景：创建一个命名卷
docker volume create "${VOLUME_NAME}"

# 场景：查看本地所有卷
docker volume ls

# 场景：删除卷（卷中数据会一并删除，谨慎）
docker volume rm "${VOLUME_NAME}"
docker volume prune      # 删除所有未被使用的卷（谨慎）

# 场景：使用卷挂载数据（适合数据库等持久化场景）
docker run -d \
  --name my-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -v "${VOLUME_NAME}":/var/lib/postgresql/data \
  -p 5432:5432 \
  postgres:16
# 说明：
#   -v 卷名:容器内路径  使用 Docker 卷做持久化

# 场景：使用宿主机目录做绑定挂载（方便直接看文件）
docker run -d \
  --name "${CONTAINER_NAME}" \
  -p 8080:80 \
  -v "$(pwd)"/html:/usr/share/nginx/html:ro \
  "${IMAGE_NAME}"
# 说明：
#   -v 宿主机路径:容器路径:ro  只读挂载，避免容器修改宿主机文件

# ------------------------------------------------------------
# 6. 文件拷贝：宿主机 <-> 容器
# ------------------------------------------------------------

# 场景：从宿主机复制文件到容器
docker cp ./local-file.txt "${CONTAINER_NAME}":/tmp/local-file.txt

# 场景：从容器复制文件到宿主机
docker cp "${CONTAINER_NAME}":/etc/nginx/nginx.conf ./nginx.conf

# ------------------------------------------------------------
# 7. 网络：自定义网络、多容器互联
# ------------------------------------------------------------

# 场景：查看网络
docker network ls

# 场景：创建一个自定义 bridge 网络
docker network create "${NETWORK_NAME}"

# 场景：以自定义网络运行多个容器，使用容器名互相访问
docker run -d \
  --name app1 \
  --network "${NETWORK_NAME}" \
  nginx:latest

docker run -d \
  --name app2 \
  --network "${NETWORK_NAME}" \
  curlimages/curl:latest sleep 3600

# 场景：在 app2 容器内通过 app1 主机名访问 nginx
docker exec -it app2 curl http://app1

# 场景：将已有容器连接到网络 / 从网络断开
docker network connect "${NETWORK_NAME}" "${CONTAINER_NAME}"
docker network disconnect "${NETWORK_NAME}" "${CONTAINER_NAME}"

# ------------------------------------------------------------
# 8. 构建镜像 & 标记 & 推送
# ------------------------------------------------------------

# 场景：在当前目录根据 Dockerfile 构建镜像
docker build -t my-image:latest .

# 场景：给已有镜像打一个新标签（比如推送到私有仓库）
docker tag my-image:latest registry.example.com/my-namespace/my-image:latest

# 场景：登录镜像仓库
docker login registry.example.com

# 场景：推送镜像到仓库
docker push registry.example.com/my-namespace/my-image:latest

# ------------------------------------------------------------
# 9. 镜像 & 容器导出/导入（备份）
# ------------------------------------------------------------

# 场景：保存镜像到 tar 文件（备份或离线传输）
docker save -o my-image.tar my-image:latest

# 场景：从 tar 文件加载镜像
docker load -i my-image.tar

# 场景：导出容器的文件系统（不包含历史、层信息）
docker export -o my-container.tar "${CONTAINER_NAME}"

# 场景：从导出的 tar 创建新的镜像
cat my-container.tar | docker import - my-new-image:latest

# ------------------------------------------------------------
# 10. 资源清理：容器、镜像、网络、卷
# ------------------------------------------------------------

# 场景：删除所有已停止的容器
docker container prune      # 谨慎，确认提示后清理所有已停止的容器

# 场景：删除未使用的镜像、网络、卷（不影响正在使用的）
docker system prune         # 较安全的清理

# 场景：更激进的清理（包括所有未使用的镜像）
# 非常谨慎使用，可能删除你之后还要用的镜像
docker system prune -a

# 场景：查看磁盘使用情况
docker system df            # 类似 "docker 磁盘空间使用情况" 报告

# ------------------------------------------------------------
# 11. docker compose 常用命令（v2 语法：docker compose）
# ------------------------------------------------------------

# 场景：启动 compose 中定义的所有服务（后台运行）
docker compose -f "${COMPOSE_FILE}" up -d

# 场景：查看 compose 服务的日志
docker compose -f "${COMPOSE_FILE}" logs
docker compose -f "${COMPOSE_FILE}" logs -f       # 实时跟随

# 场景：停止并删除 compose 创建的容器、网络等
docker compose -f "${COMPOSE_FILE}" down

# 场景：只重启某一个服务（比如 web）
docker compose -f "${COMPOSE_FILE}" restart web

# 场景：只查看某一个服务的日志
docker compose -f "${COMPOSE_FILE}" logs -f web

# ------------------------------------------------------------
# 12. 调试信息：快速查看某类对象的 ID 列表
# ------------------------------------------------------------

# 场景：只想拿到容器 ID 列表
docker ps -aq

# 场景：只想拿到镜像 ID 列表
docker images -q

# 场景：一键删除所有已停止容器（非常谨慎）
# docker rm $(docker ps -aq)

# 场景：一键删除所有镜像（超级危险，一般不要用）
# docker rmi $(docker images -q)

# ============================================================
# 结束
# 说明：
#   - 建议：把这一份脚本当成“带注释的命令速查表”，
#     每次操作前先看注释再决定要不要执行。
# ============================================================

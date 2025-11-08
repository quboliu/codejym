## 部署指南

本项目采用前后端分离架构：React/Vite SPA + Go API 服务。部署方式分为“纯软件部署”（直接在宿主机构建并运行）与“容器部署”（Docker/Compose）。两种模式均由 `scripts/deploy.sh` 一键脚本支持。

---

### 1. 目录结构 & 产物

```
frontend/   # React 源码，Vite 构建输出 dist/
backend/    # Go 服务，负责 API、文件存储、会话管理
data/       # 运行时上传与会话数据（可自定义）
```

运行时必须让后端可读写 `DATA_DIR`，并通过 `FRONTEND_DIR` 指向已构建好的 `frontend/dist`，这样 Go 服务会同时托管 API 与静态资源。

关键环境变量：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ADDR` | HTTP 监听地址 | `:8080` |
| `DATA_DIR` | 元数据 + 上传文件目录 | `./data` |
| `FRONTEND_DIR` | 前端静态资源目录 | `./frontend/dist` |

---

### 2. 一键部署脚本

```bash
scripts/deploy.sh [local|docker]
```

#### `local`（纯软件）
1. 安装/更新前端依赖并执行 `npm run build`。
2. 编译 Go 服务到 `bin/codecopybook`。
3. 创建数据目录（默认 `./data`）。
4. 以当前终端前台方式启动服务（可通过 `Ctrl+C` 结束）。

可通过环境变量覆盖默认行为，例如：

```bash
ADDR=:9090 DATA_DIR=/srv/codecopybook-data scripts/deploy.sh local
```

#### `docker`
1. 使用根目录下 `Dockerfile` 进行多阶段构建。
2. `docker compose up -d` 运行服务，自动挂载本地 `./data` 到容器 `/data` 以持久化上传。

停止/查看：

```bash
docker compose logs -f
docker compose down
```

---

### 3. 手动部署步骤（可选）

**纯软件：**
1. `cd frontend && npm install && npm run build`
2. `cd backend && go build -o ../bin/codecopybook ./cmd/server`
3. `DATA_DIR=./data FRONTEND_DIR=../frontend/dist ./bin/codecopybook -addr :8080`

**容器：**
1. `docker compose build`
2. `docker compose up -d`

---

### 4. 生产化建议
- 使用 Nginx/Caddy 在前面做 TLS 终端与反向代理。
- 将 `DATA_DIR` 指向独立磁盘或对象存储挂载，定期备份 `assets.json` 与 `sessions.json`。
- 根据需要在 Go 服务前添加鉴权（当前版本默认本地/信任环境）。
- 配置 `LOG_LEVEL`（后续可扩展）以及上传大小限制以满足企业策略。

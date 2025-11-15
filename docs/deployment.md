## 部署指南

本项目采用前后端分离架构：React/Vite SPA + Go API 服务，配合 PostgreSQL 保存账号/素材/会话数据。`scripts/deploy.sh` 既支持纯软件部署（local）也支持 Docker/Compose（docker），默认会自动构建前端和后端。

---

### 1. 目录结构 & 运行所需文件

```
frontend/   # React 源码，Vite 构建输出 dist/
backend/    # Go 服务，负责 API、鉴权、文件存储
data/       # 运行时上传与缓存（持久化挂载）
```

运行时需保证：
- `DATA_DIR` 指向可读写目录，包含 `uploads/`。
- `FRONTEND_DIR` 指向打包后的 `frontend/dist`。
- 提供可访问的 PostgreSQL（`DATABASE_URL`）和 JWT 签名密钥（`AUTH_SECRET`）。服务启动会自动执行数据库迁移。

关键环境变量：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ADDR` | HTTP 监听地址 | `:8080` |
| `DATA_DIR` | 上传/缓存目录 | `./data` |
| `FRONTEND_DIR` | 前端静态资源目录 | `./frontend/dist` |
| `DATABASE_URL` | PostgreSQL 连接串 | 自动回退至 `postgres://codecopy:codecopy@localhost:5432/codecopybook?sslmode=disable`（当脚本为你拉起 Docker PG 时） |
| `AUTH_SECRET` | HS256 签名密钥 | 若未设置，脚本会自动生成随机值 |

---

### 2. 一键部署脚本

```bash
scripts/deploy.sh [local|docker]
```

#### `local`
1. 构建前端：`npm install && npm run build`
2. 构建后端：`go build -o bin/codecopybook ./cmd/server`
3. 如未手动提供数据库，脚本会使用 Docker Compose 自动拉起 `postgres` 服务，并将 `DATABASE_URL` 指向 `localhost:${POSTGRES_PORT}`；否则可自行设置 `DATABASE_URL`
4. 若 `AUTH_SECRET` 未设置，脚本会随机生成；也可通过环境变量显式指定
5. 前台启动服务（`Ctrl+C` 结束）

示例（使用现有 Postgres）：

```bash
export DATABASE_URL=postgres://user:pass@localhost:5432/codecopybook?sslmode=disable
export AUTH_SECRET=$(openssl rand -hex 32)
ADDR=:9090 DATA_DIR=/srv/codecopybook-data scripts/deploy.sh local
```

#### `docker`
1. `docker compose build`
2. `docker compose up -d`

Compose 会同时拉起 `postgres` 与 `codecopybook` 两个服务：
- 默认数据库账号/密码均为 `codecopy`，库名 `codecopybook`
- Postgres 端口会映射到宿主 `localhost:${POSTGRES_PORT:-5432}`
- `DATABASE_URL` / `AUTH_SECRET` 可通过 `.env` 或运行时环境覆盖
- 宿主 `./data` 会绑定到容器 `/data`

常用命令：

```bash
docker compose logs -f
docker compose down
```

---

### 3. 手动部署（可选）

**纯软件：**
1. `cd frontend && npm install && npm run build`
2. `cd backend && go build -o ../bin/codecopybook ./cmd/server`
3. `DATA_DIR=./data FRONTEND_DIR=../frontend/dist DATABASE_URL=... AUTH_SECRET=... ./bin/codecopybook -addr :8080`

---

### 4. 数据迁移（从旧版 JSON 升级）

旧版本存储于 `data/assets.json` / `data/sessions.json`。升级至 PostgreSQL 后，可使用新命令导入历史记录：

```bash
cd backend
go run ./cmd/migratejson \
  -db "$DATABASE_URL" \
  -data ../data \
  -user-email you@example.com \
  -user-name "Your Name" \
  -user-password "initial-password"
```

- 若 `user-email` 已存在，则无需 `-user-password`，迁移数据将附加到该用户。
- 工具会把 `data/uploads/<assetId>` 重新移动到 `data/uploads/<userId>/<assetId>`，请确保磁盘容量充足。
- 迁移完成后即可使用登录入口访问历史素材与会话。

---

### 5. 生产化建议
- 使用 Nginx/Caddy 做 TLS 终端与反向代理。
- 将 `DATA_DIR` 挂载至独立磁盘/对象存储，并定期备份 PostgreSQL。
- `AUTH_SECRET` 应由安全随机源生成（≥32 字节）并妥善保管。
- 根据环境开启数据库监控、自动备份与连接池告警。
- 配置上传大小限制、审计日志与额外的访问控制策略。

# CodeJYM - 代码学习平台

## 🚀 一键部署（包含域名访问）

现在你可以使用一个脚本启动所有服务，包括反向代理：

```bash
./deploy-full.sh
```

这将自动：
1. ✅ 清理旧容器
2. ✅ 构建镜像
3. ✅ 启动 PostgreSQL + 应用 + Nginx
4. ✅ 等待服务就绪
5. ✅ 验证部署成功

---

## 📋 访问地址

### 本地访问
- **应用**：http://localhost:8080
- **API**：http://localhost:8080/api
- **数据库**：localhost:5432

### 通过 Nginx 反向代理
- **应用**：http://localhost
- **API**：http://localhost/api

### 你的域名
- **jiezispace.com**：通过 Nginx 代理访问
- **www.jiezispace.com**：通过 Nginx 代理访问

---

## 🎯 快速开始

### 方式 1：全功能部署（推荐）
```bash
./deploy-full.sh
```
包含：PostgreSQL + 应用 + Nginx 反向代理

### 方式 2：本地部署（仅本地访问）
```bash
./deploy.sh
```
包含：PostgreSQL + 应用

---

## 📦 服务架构

```
┌─────────────────────────────────────────────────┐
│                                                 │
│  ┌──────────────┐      ┌──────────────┐        │
│  │   浏览器      │──────▶   Nginx      │        │
│  │              │      │  (反向代理)   │        │
│  └──────────────┘      └──────┬───────┘        │
│                                 │                │
│                         ┌───────▼───────┐        │
│                         │   应用服务     │        │
│                         │ (codecopybook) │        │
│                         └───────┬───────┘        │
│                                 │                │
│                         ┌───────▼───────┐        │
│                         │  PostgreSQL   │        │
│                         │    数据库      │        │
│                         └────────────────┘        │
│                                                 │
│  端口：                                         │
│    - 80:   Nginx                               │
│    - 8080: 应用                                │
│    - 5432: 数据库                              │
└─────────────────────────────────────────────────┘
```

---

## 🔧 管理命令

### 查看服务状态
```bash
docker compose -f docker-compose.proxy.yml ps
```

### 查看日志
```bash
# 应用日志
docker compose -f docker-compose.proxy.yml logs -f codecopybook

# 数据库日志
docker compose -f docker-compose.proxy.yml logs -f postgres

# Nginx 日志
docker compose -f docker-compose.proxy.yml logs -f nginx
```

### 重启服务
```bash
docker compose -f docker-compose.proxy.yml restart
```

### 停止服务
```bash
docker compose -f docker-compose.proxy.yml down
```

### 完全清理（删除所有数据）
```bash
docker compose -f docker-compose.proxy.yml down -v
```

---

## 🌐 域名配置

### DNS 设置
```
类型: A
名称: @
值: YOUR_SERVER_IP
TTL: 600

类型: A
名称: www
值: YOUR_SERVER_IP
TTL: 600
```

### Nginx 配置
已自动配置 jiezispace.com：
- `nginx.conf` - Nginx 配置文件

### 访问测试
```bash
# 测试本地访问
curl http://localhost:8080

# 测试 Nginx 代理
curl http://localhost

# 测试 API
curl http://localhost/api/auth/me
```

---

## 🔐 数据库信息

- **主机**：postgres
- **端口**：5432
- **数据库**：codecopybook
- **用户名**：codecopy
- **密码**：codecopy123

---

## 📚 完整文档

- `QUICK_START.md` - 快速开始指南
- `DOMAIN_SETUP_JIEZISPACE.md` - 详细域名配置指南
- `DEPLOYMENT_FIX.md` - 问题修复说明

---

## ✅ 验证部署

部署完成后，以下所有测试都应该通过：

```bash
# 1. 测试本地访问
curl http://localhost:8080

# 2. 测试 Nginx 代理
curl http://localhost

# 3. 测试 API
curl http://localhost:8080/api/auth/me

# 4. 检查服务状态
docker compose -f docker-compose.proxy.yml ps
```

---

## 🎉 完成！

你现在可以通过以下方式访问 CodeJYM：

1. **本地**：http://localhost:8080
2. **Nginx**：http://localhost
3. **域名**：http://jiezispace.com（需要 DNS 配置）

享受你的代码学习平台！💻✨

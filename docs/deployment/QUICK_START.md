# 🚀 CodeJYM 快速开始指南

## 你的域名：jiezispace.com

---

## 一、本地部署（立即可用）

```bash
# 1. 一键部署
./scripts/deploy.sh

# 2. 访问
# http://localhost:8080
```

---

## 二、域名部署（推荐）

### 选项 1：使用 Docker Compose + Nginx（最简单）

```bash
# 1. 停止本地服务
docker compose -f config/docker-compose.yml down

# 2. 启动完整服务（包含 Nginx 反向代理）
docker compose -f config/docker-compose.proxy.yml up -d

# 3. 访问
# http://jiezispace.com
# http://www.jiezispace.com
```

### 选项 2：使用 Caddy（自动 HTTPS）

**步骤 1**：修改 `docker-compose.proxy.yml`
```yaml
# 注释掉 nginx 服务
# nginx:
#   ...

# 取消注释 caddy 服务
caddy:
  image: caddy:alpine
  restart: unless-stopped
  ports:
    - "80:80"
    - "443:443"
  volumes:
    - ./Caddyfile:/etc/caddy/Caddyfile:ro
    - caddy_data:/data
    - caddy_config:/config
  depends_on:
    - codecopybook
```

**步骤 2**：启动服务
```bash
docker compose -f config/docker-compose.proxy.yml up -d
```

**步骤 3**：访问
```bash
https://jiezispace.com  # 自动 HTTPS
```

---

## 三、系统级 Nginx（可选）

如果希望 Nginx 运行在宿主机系统上：

```bash
# 1. 安装 Nginx
sudo apt update
sudo apt install -y nginx

# 2. 复制配置
sudo cp config/nginx.conf /etc/nginx/sites-available/jiezispace.com
sudo ln -s /etc/nginx/sites-available/jiezispace.com /etc/nginx/sites-enabled/

# 3. 测试并重启
sudo nginx -t
sudo systemctl restart nginx

# 4. 访问
# http://jiezispace.com
```

---

## 验证部署

```bash
# 检查服务状态
docker compose ps

# 查看日志
docker compose logs -f codecopybook

# 测试访问
curl http://localhost:8080/healthz
curl http://jiezispace.com
```

---

## 管理命令

```bash
# 查看服务
docker compose -f config/docker-compose.yml ps

# 查看日志
docker compose -f config/docker-compose.yml logs -f codecopybook
docker compose -f config/docker-compose.yml logs -f postgres

# 重启服务
docker compose -f config/docker-compose.yml restart

# 停止服务
docker compose -f config/docker-compose.yml down

# 完全清理
docker compose -f config/docker-compose.yml down -v
```

---

## 故障排除

### 问题：域名访问返回 502

**解决**：
```bash
# 1. 检查服务状态
docker compose ps

# 2. 检查应用日志
docker compose logs codecopybook

# 3. 重启服务
docker compose restart
```

### 问题：端口被占用

**解决**：
```bash
# 检查端口占用
sudo lsof -i :80

# 停止冲突服务
sudo systemctl stop apache2
```

---

## 访问地址

✅ **本地**：
- http://localhost:8080

✅ **域名**：
- http://jiezispace.com
- http://www.jiezispace.com

✅ **HTTPS**（使用 Caddy 或配置 SSL）：
- https://jiezispace.com
- https://www.jiezispace.com

---

## 下一步

1. **配置 HTTPS**（生产环境必需）
   - 查看 `DOMAIN_SETUP_JIEZISPACE.md` 了解详细步骤

2. **优化配置**
   - 修改数据库密码
   - 配置防火墙
   - 设置备份策略

3. **监控**
   - 配置日志轮转
   - 设置监控告警

---

📚 **完整文档**：
- `DOMAIN_SETUP_JIEZISPACE.md` - 详细域名配置指南
- `DEPLOYMENT_FIX.md` - 问题修复说明
- `README_DEPLOYMENT.md` - 部署指南

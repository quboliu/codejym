# 域名访问和日志问题解决方案

## 问题 1：PostgreSQL 日志错误

### 症状
```
FATAL: database "codecopy" does not exist
```

### 原因
Docker Compose 的 PostgreSQL 健康检查配置不当，每 5 秒产生错误日志。

### 解决方案 ✅
已修复：修改了 `docker-compose.yml` 中的健康检查配置：
```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready"]  # 简化为只检查服务器状态
  interval: 10s                       # 增加到 10 秒
  timeout: 5s
  retries: 10
  start_period: 30s                   # 给启动更多时间
```

验证：
```bash
docker compose logs postgres 2>&1 | grep "codecopy.*does not exist"
# 应该没有输出（错误已消失）
```

---

## 问题 2：域名访问返回 502 错误

### 症状
- 配置了域名 A 记录指向本机 IP
- 访问 `域名:8080` 返回 **HTTP ERROR 502**
- 本地访问 `localhost:8080` 正常

### 原因
502 Bad Gateway 表示反向代理无法连接到后端服务器。Docker 应用的端口映射只绑定到 localhost，外部无法直接访问。

### 解决方案

#### 方案一：使用 Nginx 反向代理（推荐）

**1. 安装 Nginx**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y nginx

# CentOS/RHEL
sudo yum install -y nginx
```

**2. 配置 Nginx**
编辑 `/etc/nginx/sites-available/codejym`（或使用 `nginx.conf`）：
```nginx
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;  # 替换为你的域名

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**3. 启用配置**
```bash
# 创建软链接
sudo ln -s /etc/nginx/sites-available/codejym /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重启 Nginx
sudo systemctl restart nginx
```

**4. 访问测试**
```
http://your-domain.com  # 不需要加端口号
```

#### 方案二：使用 Caddy（更简单，自动 HTTPS）

**1. 安装 Caddy**
```bash
# Ubuntu/Debian
sudo apt install -y caddy

# 或使用 Docker
docker pull caddy:alpine
```

**2. 配置 Caddyfile**
创建 `/etc/caddy/Caddyfile`：
```
your-domain.com, www.your-domain.com {
    reverse_proxy localhost:8080
}
```

**3. 启动 Caddy**
```bash
sudo systemctl start caddy
sudo systemctl enable caddy
```

**4. 访问测试**
```
https://your-domain.com  # 自动 HTTPS
```

#### 方案三：使用 Docker Compose（已配置好）

使用提供的 `docker-compose.proxy.yml`：

```bash
# 停止当前服务
docker compose down

# 使用新的配置启动（包含反向代理）
docker compose -f docker-compose.proxy.yml up -d

# 访问域名（不需要端口号）
http://your-domain.com
```

---

## 完整部署流程

### 1. 克隆项目并部署
```bash
git clone <your-repo>
cd CodeJYM

# 一键部署
./deploy.sh
```

### 2. 配置 Nginx（或 Caddy）
```bash
# 使用 nginx.conf
sudo cp nginx.conf /etc/nginx/sites-available/codejym
sudo ln -s /etc/nginx/sites-available/codejym /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# 或使用 Caddy
sudo cp Caddyfile /etc/caddy/Caddyfile
sudo systemctl start caddy
```

### 3. 配置域名 DNS
```
类型: A
名称: @ (或 www)
值: YOUR_SERVER_IP
TTL: 600
```

### 4. 访问测试
```
http://your-domain.com  # 应该看到应用首页
```

### 5. 获取 SSL 证书（HTTPS）
```bash
# 使用 Caddy（自动）
# Caddy 会自动获取 Let's Encrypt 证书

# 使用 Nginx + Certbot
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

---

## 故障排除

### 502 错误持续
```bash
# 1. 检查服务状态
docker compose ps

# 2. 检查应用日志
docker compose logs codecopybook

# 3. 检查 Nginx 日志
sudo tail -f /var/log/nginx/error.log

# 4. 测试本地连接
curl http://localhost:8080
```

### 端口被占用
```bash
# 检查端口占用
sudo lsof -i :80
sudo lsof -i :8080

# 停止冲突服务
sudo systemctl stop apache2  # 如果安装了 Apache
```

### 防火墙配置
```bash
# Ubuntu UFW
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# CentOS firewalld
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

---

## 使用指南

### 一键部署脚本
```bash
# 基础部署（仅本地访问）
./deploy.sh

# 完整部署（包含反向代理）
docker compose -f docker-compose.proxy.yml up -d
```

### 管理命令
```bash
# 查看状态
docker compose ps

# 查看日志
docker compose logs -f codecopybook
docker compose logs -f postgres

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 完全清理
docker compose down -v
```

---

## 总结

✅ **问题 1**：PostgreSQL 日志错误 - 已通过优化健康检查配置解决

✅ **问题 2**：域名访问 502 错误 - 通过配置 Nginx 或 Caddy 反向代理解决

现在可以通过 `http://your-domain.com` 正常访问应用！

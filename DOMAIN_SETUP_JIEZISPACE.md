# jiezispace.com 域名配置指南

## 你的域名：jiezispace.com

### 方案一：使用 Nginx（推荐）

#### 1. 安装 Nginx
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y nginx

# CentOS/RHEL
sudo yum install -y nginx
```

#### 2. 配置 Nginx
```bash
# 复制配置文件
sudo cp nginx.conf /etc/nginx/sites-available/jiezispace.com

# 创建软链接
sudo ln -s /etc/nginx/sites-available/jiezispace.com /etc/nginx/sites-enabled/

# 删除默认配置（可选）
sudo rm /etc/nginx/sites-enabled/default

# 测试配置
sudo nginx -t

# 重启 Nginx
sudo systemctl restart nginx
sudo systemctl enable nginx
```

#### 3. 访问测试
```
http://jiezispace.com
https://jiezispace.com （需要配置 SSL）
```

---

### 方案二：使用 Caddy（更简单，自动 HTTPS）

#### 1. 安装 Caddy
```bash
# Ubuntu/Debian
sudo apt install -y caddy

# 或使用 Docker
docker pull caddy:alpine
```

#### 2. 配置 Caddyfile
```bash
# 复制配置文件
sudo cp Caddyfile /etc/caddy/Caddyfile

# 启动 Caddy
sudo systemctl start caddy
sudo systemctl enable caddy
```

#### 3. 访问测试
```
https://jiezispace.com  # 自动 HTTPS
```

---

### 方案三：使用 Docker Compose（已配置）

#### 1. 停止当前服务
```bash
docker compose down
```

#### 2. 使用新的配置启动（包含 Nginx）
```bash
docker compose -f docker-compose.proxy.yml up -d
```

#### 3. 访问测试
```
http://jiezispace.com
```

#### 4. 如果需要 Caddy，修改 docker-compose.proxy.yml：
- 注释掉 nginx 服务
- 取消注释 caddy 服务

---

## DNS 配置确认

确保你的域名 DNS 设置正确：

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

---

## SSL 证书配置（HTTPS）

### 使用 Caddy（推荐，自动）
Caddy 会自动获取 Let's Encrypt 证书

### 使用 Nginx + Certbot
```bash
# 安装 Certbot
sudo apt install -y certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d jiezispace.com -d www.jiezispace.com

# 自动续期
sudo systemctl enable certbot.timer
```

---

## 验证部署

### 检查服务状态
```bash
# 检查 Docker 服务
docker compose ps

# 检查 Nginx/Caddy
sudo systemctl status nginx
# 或
sudo systemctl status caddy
```

### 测试访问
```bash
# 本地测试
curl http://localhost:8080

# 域名测试
curl http://jiezispace.com
curl -I http://jiezispace.com
```

### 查看日志
```bash
# 应用日志
docker compose logs -f codecopybook
docker compose logs -f postgres

# Nginx 日志
sudo tail -f /var/log/nginx/jiezispace_error.log

# Caddy 日志
sudo journalctl -u caddy -f
```

---

## 故障排除

### 1. 502 错误
```bash
# 检查应用是否运行
docker compose ps

# 检查应用日志
docker compose logs codecopybook

# 检查端口是否监听
ss -tuln | grep 8080
```

### 2. 域名无法访问
```bash
# 检查 DNS 解析
nslookup jiezispace.com
dig jiezispace.com

# 检查防火墙
sudo ufw status
sudo firewall-cmd --list-all
```

### 3. 端口占用
```bash
# 检查 80 端口占用
sudo lsof -i :80
sudo lsof -i :443

# 停止冲突服务
sudo systemctl stop apache2
```

---

## 完成后的访问地址

✅ **本地访问**：
- http://localhost:8080

✅ **域名访问**：
- http://jiezispace.com
- http://www.jiezispace.com

✅ **HTTPS 访问**（配置 SSL 后）：
- https://jiezispace.com
- https://www.jiezispace.com

---

## 推荐配置

**最简单方案**：
1. 使用 Caddy（自动 HTTPS）
2. 配置 Docker Compose：
   ```bash
   docker compose -f docker-compose.proxy.yml up -d
   ```

**生产环境推荐**：
1. 使用 Nginx + Certbot
2. 配置负载均衡（如果需要）
3. 配置防火墙和安全策略

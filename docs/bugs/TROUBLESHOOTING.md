# CodeJYM 部署运维故障排查手册

## 目录
1. [PostgreSQL 日志错误问题](#1-postgresql-日志错误问题)
2. [Docker 容器权限问题](#2-docker-容器权限问题)
3. [Nginx 反向代理 502 错误](#3-nginx-反向代理-502-错误)
4. [健康检查配置问题](#4-健康检查配置问题)
5. [Nginx 健康检查配置问题](#5-nginx-健康检查配置问题)
6. [PostgreSQL 主机冲突](#6-postgresql-主机冲突)
7. [域名访问和 DNS 传播验证问题](#7-域名访问和-dns-传播验证问题)
8. [其他服务配置问题](#8-其他服务配置问题)

---

## 1. PostgreSQL 日志错误问题

### 问题描述
```
FATAL: database "codecopy" does not exist
```
每 15 秒重复出现，但应用和数据库连接正常。

### 排查方法
```bash
# 查看PostgreSQL日志
docker compose logs postgres

# 检查数据库列表
docker compose exec postgres psql -U codecopy -d codecopybook -c "\l"

# 检查健康检查配置
cat docker-compose.yml | grep -A 5 "healthcheck"

# 手动测试健康检查命令
docker compose exec postgres pg_isready -U codecopy
```

### 根因定位
PostgreSQL 健康检查命令 `pg_isready -U codecopy` 会默认尝试连接**同名数据库** `codecopy`，但实际数据库名称是 `codecopybook`，导致错误。

### 解决方法
修改 `docker-compose.yml` 和 `docker-compose.proxy.yml` 中的健康检查配置：
```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U codecopy -d codecopybook || pg_isready"]
  interval: 15s
  timeout: 5s
  retries: 10
```

---

## 2. Docker 容器权限问题

### 问题描述
```
前端错误：failed to prepare storage
```
文件上传功能失败，无法在数据目录创建子目录。

### 排查方法
```bash
# 查看应用日志
docker compose logs codecopybook

# 检查容器内用户ID
docker compose exec codecopybook id

# 检查挂载目录权限
docker compose exec codecopybook ls -ld /data

# 测试在数据目录创建文件
docker compose exec codecopybook sh -c "mkdir -p /data/test && rm -rf /data/test"

# 检查宿主目录权限
ls -ld /opt/codejym/data
```

### 根因定位
容器内 `app` 用户默认 UID=100，但宿主机 `/data` 目录所有者 UID=1001，权限不匹配导致写入失败。

### 解决方法
修改 `Dockerfile` 第 21 行，设置应用用户 UID：
```dockerfile
# 修改前
RUN addgroup -S app && adduser -S app -G app

# 修改后
RUN addgroup -S app && adduser -S app -G app -u 1001
```

---

## 3. Nginx 反向代理 502 错误

### 问题描述
访问域名返回：
```
HTTP ERROR 502
```
但直接访问应用正常。

### 排查方法
```bash
# 检查服务状态
docker compose ps

# 查看Nginx错误日志
docker compose logs nginx

# 测试容器间网络连通性
docker compose exec nginx sh -c "ping -c 2 codecopybook"

# 检查Nginx配置
docker compose exec nginx sh -c "cat /etc/nginx/conf.d/default.conf"

# 测试容器内访问
docker compose exec nginx sh -c "curl -I http://localhost:8080"
```

### 根因定位
Nginx 配置中使用 `proxy_pass http://localhost:8080;`，但在 Docker 容器内部，`localhost` 指向的是 Nginx 容器本身，无法访问到运行在 `codecopybook` 容器中的应用服务。

### 解决方法
修改 `nginx.conf` 中的代理配置：
```nginx
# 修改前
proxy_pass http://localhost:8080;

# 修改后
proxy_pass http://codecopybook:8080;
```

需要修改的 location 块：
- `/` - 主应用
- `/*.js|*.css|*.png` - 静态文件
- `/healthz` - 健康检查

---

## 4. 健康检查配置问题

### 问题描述
```
dependency failed to start: container codejym-postgres-1 has no healthcheck configured
```
服务启动失败，应用无法依赖数据库启动。

### 排查方法
```bash
# 检查配置文件
cat docker-compose.proxy.yml | grep -B 2 -A 3 "healthcheck"

# 查看服务依赖
cat docker-compose.proxy.yml | grep -A 3 "depends_on"
```

### 根因定位
为了消除 PostgreSQL 日志错误，曾临时禁用了健康检查 (`disable: true`)，但应用服务使用了 `depends_on` 条件依赖健康检查，导致启动失败。

### 解决方法
恢复 PostgreSQL 健康检查配置：
```yaml
postgres:
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U codecopy -d codecopybook || pg_isready"]
    interval: 10s
    timeout: 5s
    retries: 10
```

---

### 问题描述 2
```
Nginx health check: (health: starting) 持续显示不更新
```
Nginx 容器健康检查失败，显示 `health: starting` 状态。

### 排查方法
```bash
# 检查 Nginx 健康检查日志
docker inspect codejym-nginx-1 | grep -A 10 "Health"

# 手动测试健康检查命令
docker exec codejym-nginx-1 sh -c "wget -qO- http://localhost:8080 >/dev/null 2>&1 && echo OK || echo Failed"

# 验证容器间连通性
docker exec codejym-nginx-1 sh -c "wget -qO- http://codecopybook:8080 >/dev/null 2>&1 && echo OK || echo Failed"
```

### 根因定位
Nginx 健康检查命令使用了 `localhost:8080`，但在 Docker 容器内部，`localhost` 指向 Nginx 容器本身，无法访问运行在 `codecopybook` 容器中的应用服务。

### 解决方法
修改 `docker-compose.proxy.yml` 中的 Nginx 健康检查配置：
```yaml
nginx:
  healthcheck:
    test: ["CMD-SHELL", "wget -qO- http://codecopybook:8080 >/dev/null 2>&1 || exit 1"]
    interval: 10s
    timeout: 5s
    retries: 10

# 应用配置
docker compose -f docker-compose.proxy.yml up -d --force-recreate nginx
```

---

## 6. PostgreSQL 主机冲突

### 问题描述
```
FATAL: role "root" does not exist
```
持续每 15 秒出现，错误来源不是 Docker 容器。

### 排查方法
```bash
# 检查系统PostgreSQL进程
ps aux | grep postgres

# 检查端口占用
ss -tlnp | grep 5432

# 检查容器内PostgreSQL端口
docker compose exec postgres psql -U codecopy -d codecopybook -c "SELECT inet_server_addr(), server_port;"

# 查看所有PostgreSQL连接
docker compose exec postgres psql -U codecopy -d codecopybook -c "SELECT pid, usename, datname FROM pg_stat_activity;"
```

### 根因定位
宿主机系统存在独立的 PostgreSQL 服务在运行（端口 5432），与 Docker PostgreSQL 端口冲突。外部监控工具或定时任务尝试连接宿主 PostgreSQL 时因配置不正确而报错。

### 解决方法
将 Docker PostgreSQL 映射到不同端口避免冲突：
```yaml
postgres:
  ports:
    - "127.0.0.1:5433:5432"  # 改为5433端口
```

---

## 7. 域名访问和 DNS 传播验证问题

### 问题描述
1. 现象 1：DNS 解析域名返回 Cloudflare IP，但浏览器访问显示 "Registered at spaceship" 页面
2. 现象 2：Cloudflare 设置为 Proxied 后，无法看到真实 IP，不确定是否配置正确

### 排查方法
```bash
# 1. 查询本地 DNS 解析（可能受缓存影响）
nslookup jiezispace.com

# 2. 查询权威 DNS 服务器（Cloudflare）
dig @1.1.1.1 jiezispace.com

# 3. 查询权威 DNS 服务器（Google）
dig @8.8.8.8 jiezispace.com

# 4. 查询权威 DNS 完整信息
dig @1.1.1.1 jiezispace.com ANY +noall +answer

# 5. 使用不同 DNS 服务器验证
dig @1.1.1.1 jiezispace.com +short
dig @8.8.8.8 jiezispace.com +short

# 6. 测试服务器直接访问
curl -I http://142.171.88.196
curl -H "Host: jiezispace.com" http://142.171.88.196

# 7. 测试本地 Nginx 代理
curl -I http://localhost

# 8. 检查服务器端口监听状态
ss -tlnp | grep :80
ss -tlnp | grep :8080

# 9. 验证 DNS 传播状态（在线工具）
# https://www.whatsmydns.net/
# 输入域名查看全球 DNS 服务器查询结果

# 10. 检查 Cloudflare 面板状态
# 访问 https://dash.cloudflare.com/
# 确认 DNS 面板中 A 记录配置
```

### 根因定位

#### 问题 1：DNS 传播未完成
**现象**：
- 权威 DNS 查询显示：`172.67.222.11` 和 `104.21.25.25`
- 没有记录指向服务器 IP：`142.171.88.196`

**原因**：
- 刚设置 DNS 记录，需要全球传播时间
- DNS 根服务器同步需要 5-30 分钟

#### 问题 2：Cloudflare Proxied 模式特性
**现象**：
- DNS 查询返回 Cloudflare IP（非真实 IP）
- 不确定配置是否生效

**原因**：
- Cloudflare Proxied 模式（橙色云朵）会隐藏真实 IP
- 这是正常的安全特性，查询结果应该显示 Cloudflare 边缘节点 IP
- 流量实际会转发到真实服务器，但外部查询看不到

**验证方法**：
- Cloudflare 权威 DNS 查询应该显示新的 IP（142.171.88.196）
- 但由于 Proxied，显示的是 Cloudflare 代理 IP（不是真实 IP）
- 正确的验证方法是浏览器访问看是否返回应用

### 解决方法

#### 1. 确认 Cloudflare 配置
在 Cloudflare 控制面板检查：
```
登录 https://dash.cloudflare.com/
→ 选择域名 jiezispace.com
→ DNS 设置
→ 确认记录：
  - Type: A, Name: @, IPv4: 142.171.88.196, Proxy: ☑️ (橙色云朵)
  - Type: CNAME, Name: www, Target: jiezispace.com, Proxy: ☑️ (橙色云朵)
```

#### 2. 验证 DNS 传播
```bash
# 权威 DNS 查询（应该看到 142.171.88.196）
dig @1.1.1.1 jiezispace.com

# 如果看到多个 A 记录，删除旧的
# 只保留指向 142.171.88.196 的记录
```

#### 3. 强制刷新本地 DNS
**Windows**：
```cmd
ipconfig /flushdns
```

**macOS**：
```bash
sudo dscacheutil -flushcache
```

**Linux**：
```bash
sudo systemctl restart systemd-resolved
# 或
sudo service network-manager restart
```

#### 4. 等待时间表
```
5 分钟内：   20% 地区生效
10 分钟内：  50% 地区生效
15 分钟内：  70% 地区生效
20 分钟内：  90% 地区生效
30 分钟内：  99% 地区生效
```

#### 5. 验证是否生效
```bash
# 权威 DNS 应该显示新 IP（可能因 Proxied 隐藏）
dig @1.1.1.1 jiezispace.com

# 浏览器访问测试
http://jiezispace.com
# 应该看到应用页面，不是 "Registered at spaceship"
```

### 重要提示
**Cloudflare Proxied 模式行为**：
- ✅ 正常：DNS 查询显示 Cloudflare IP（隐藏真实 IP）
- ✅ 正常：权威 DNS 应该显示新配置（142.171.88.196）
- ❌ 错误：权威 DNS 仍显示旧记录（传播未完成）
- ✅ 生效标志：浏览器访问显示应用页面

---

## 8. 其他服务配置问题

### 问题描述
构建错误、数据库连接失败等。

### 排查方法
```bash
# 构建镜像
docker compose build

# 查看镜像列表
docker images

# 检查容器状态
docker compose ps

# 查看所有日志
docker compose logs -f

# 验证服务连通性
docker network inspect codejym_network
```

### 根因与解决方法

#### 7.1 Go 版本不匹配
```bash
# 错误
#30 0.760 go: golang:1.22-alpine: unknown tag

# 解决：修改Dockerfile
FROM golang:1.24-alpine  # 升级到1.24
```

#### 7.2 前端构建缺少 index.html
```bash
# 错误
ERROR 404 Not Found - The requested resource doesn't exist.

# 解决：修改Dockerfile
COPY frontend/index.html ./
```

#### 7.3 数据库 URL 配置
```bash
# 检查应用环境变量
docker compose exec codecopybook env | grep DATABASE

# 确认DATABASE_URL格式
DATABASE_URL=postgres://codecopy:codecopy123@postgres:5432/codecopybook?sslmode=disable
```

---

## 常用排查命令速查

### 服务管理
```bash
# 启动服务
docker compose up -d

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 完全清理
docker compose down -v

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f [service_name]
```

### 网络排查
```bash
# 检查容器网络
docker network ls
docker network inspect codejym_network

# 容器间连通性测试
docker compose exec [container1] ping [container2]
docker compose exec [container1] curl http://[container2]:8080
```

### DNS 排查
```bash
# 域名解析
nslookup jiezispace.com

# 权威DNS查询
dig @1.1.1.1 jiezispace.com
dig @8.8.8.8 jiezispace.com

# 本地DNS查询
cat /etc/resolv.conf
```

### 性能监控
```bash
# 查看资源使用
docker stats

# 查看进程
docker compose exec [service] ps aux

# 查看端口监听
ss -tlnp
netstat -tlnp
```

---

## 最佳实践

1. **日志分析**
   - 优先查看服务日志：`docker compose logs -f [service]`
   - 分析错误模式：时间戳、频率、错误代码

2. **网络排查**
   - 使用 `ping` 测试连通性
   - 使用 `curl` 测试 HTTP 服务
   - 使用 `dig` 查询 DNS

3. **权限管理**
   - 容器用户 UID 与宿主保持一致
   - 挂载目录权限适当开放

4. **健康检查**
   - 避免复杂的健康检查命令
   - 明确指定数据库和用户
   - 合理设置超时和重试次数

5. **域名配置**
   - 确认 A 记录指向正确 IP
   - 等待 DNS 传播完成（5-30 分钟）
   - 使用权威 DNS 验证配置

---

## 总结

本手册记录了 CodeJYM 部署过程中遇到的主要问题和解决方案。核心原则：
- 详细排查，精确定位根因
- 基于证据做出判断
- 优先解决关键路径问题
- 记录所有操作命令和工具

通过系统化的排查方法，可以快速定位和解决类似问题。

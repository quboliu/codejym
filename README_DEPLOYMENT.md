# CodeJYM 部署指南

## 快速开始

### 1. 一键部署（本地访问）
```bash
./deploy.sh
```
访问：http://localhost:8080

### 2. 域名访问部署
```bash
# 使用包含反向代理的配置
docker compose -f docker-compose.proxy.yml up -d
```
访问：http://your-domain.com

## 问题修复

### ✅ 问题 1：PostgreSQL 日志错误
**症状**：日志显示 `FATAL: database "codecopy" does not exist`

**解决**：已修复健康检查配置为 `pg_isready`，错误日志已消除。

### ✅ 问题 2：域名访问 502 错误
**症状**：访问域名:8080 返回 502 错误

**解决**：配置 Nginx 或 Caddy 反向代理，详细步骤请查看 `DOMAIN_SETUP.md`

## 文件说明

- `deploy.sh` - 一键部署脚本
- `docker-compose.yml` - 基础 Docker Compose 配置
- `docker-compose.proxy.yml` - 包含反向代理的配置
- `nginx.conf` - Nginx 反向代理配置模板
- `Caddyfile` - Caddy 反向代理配置模板
- `DOMAIN_SETUP.md` - 详细的域名配置指南

## 联系方式

有问题请查看相关文档或提交 Issue。

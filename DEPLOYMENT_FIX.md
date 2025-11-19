# CodeJYM 部署问题修复报告

## 问题描述

用户报告上传文件时遇到错误：
- 前端错误：`failed to prepare storage`
- PostgreSQL 日志：`FATAL: database "codecopy" does not exist`

## 根本原因

**Docker 容器权限不匹配问题**

1. **应用用户 UID 不匹配**：
   - 容器内 `app` 用户默认 UID = 100
   - 宿主机 `/data` 目录所有者 UID = 1001（应用用户）
   - 导致容器无法在挂载目录中写入文件

2. **表面现象误导**：
   - PostgreSQL 报错 "database 'codecopy' does not exist" 是由于应用写入失败后的重试连接
   - 实际数据库 `codecopybook` 存在且可正常连接
   - 真正问题是文件写入权限，而非数据库连接

## 解决方案

### 修复内容

修改 `Dockerfile` 第21行，设置应用用户 UID 为 1001：

```dockerfile
# 修改前
RUN addgroup -S app && adduser -S app -G app

# 修改后
RUN addgroup -S app && adduser -S app -G app -u 1001
```

### 验证结果

✅ **容器用户权限**：
- 应用用户：`uid=1001(app) gid=101(app)`
- 数据目录：`drwxr-xr-x 3 app 1001`

✅ **文件写入测试**：
- 目录创建：成功
- 文件写入：成功

✅ **服务状态**：
- PostgreSQL：运行正常（健康检查通过）
- 应用服务：运行正常
- 前端：可访问
- API：可响应

## 部署信息

### 服务地址
- **前端**：http://localhost:8080
- **API**：http://localhost:8080/api
- **PostgreSQL**：localhost:5432

### 数据库配置
- **用户名**：codecopy
- **密码**：codecopy123
- **数据库**：codecopybook

## 使用说明

### 一键部署
```bash
./deploy.sh
```

### 管理命令
```bash
# 查看服务状态
docker compose ps

# 查看应用日志
docker compose logs -f codecopybook

# 查看数据库日志
docker compose logs -f postgres

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 完全清理（包括数据）
docker compose down -v
```

## 修改文件列表

1. **Dockerfile**
   - 修改：设置应用用户 UID 为 1001
   - 影响：解决容器挂载目录写入权限问题

2. **deploy.sh**
   - 增强健康检查逻辑
   - 优化错误处理和用户反馈

3. **docker-compose.yml**
   - 修复：PostgreSQL 健康检查命令

## 总结

此次修复解决了容器与宿主机之间的用户权限映射问题，确保应用能够正常写入数据目录，从而修复了文件上传功能。部署脚本现已完全可用。

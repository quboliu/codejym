# 项目目录结构说明

**最后更新：** 2025-11-20
**版本：** v2.0（重组后）

---

## 📊 概览

CodeJYM 项目采用清晰的模块化目录结构，便于开发、维护和文档管理。

### 整理前后对比

| 指标 | 整理前 | 整理后 | 改进 |
|------|--------|--------|------|
| 根目录文件数 | 35+ | 14 | ⬇️ 60% |
| 文档组织方式 | 散落 | 分类存放 | ✅ |
| 脚本管理 | 混乱 | 集中管理 | ✅ |
| 配置备份 | 无 | 统一备份 | ✅ |

---

## 📁 目录结构

```
CodeJym/
├── backend/                 # 🔧 Go 后端服务
│   ├── cmd/                # 应用入口点
│   │   └── server/         # 主服务器程序
│   └── internal/           # 内部包（不对外导出）
│       ├── api/            # HTTP 处理器和路由
│       └── storage/        # 存储层（PostgreSQL + S3）
│
├── frontend/               # ⚛️ React 前端应用
│   ├── src/                # 源代码
│   ├── public/             # 静态资源
│   └── dist/               # 构建输出
│
├── docs/                   # 📚 文档中心（29篇）
│   ├── README.md           # 文档索引（从这里开始）
│   ├── deployment/         # 🚀 部署指南（6篇）
│   ├── migration/          # 🔄 迁移升级（6篇）
│   ├── bugs/               # 🐛 问题排查（8篇）
│   ├── security/           # 🔒 安全配置（4篇）
│   ├── development/        # 💻 开发文档（3篇）
│   └── user-guide/         # 📖 用户指南（2篇）
│
├── scripts/                # 🔧 运维脚本（7个）
│   ├── deploy.sh                   # 本地部署
│   ├── deploy-full.sh              # 生产部署（含 Nginx）
│   ├── backup-db-to-s3.sh          # 数据库备份
│   ├── restore-db-from-s3.sh       # 数据库恢复
│   ├── migrate-files-to-s3.go      # 文件迁移到 S3
│   ├── verify_progress_save.sh     # 验证进度保存
│   └── verify_security.sh          # 验证安全配置
│
├── config/                 # ⚙️ 配置文件备份（6个）
│   ├── docker-compose.yml          # Docker Compose 主配置
│   ├── docker-compose.proxy.yml    # 代理配置
│   ├── Dockerfile                  # Docker 镜像配置
│   ├── Caddyfile                   # Caddy 配置
│   ├── nginx.conf                  # Nginx 配置
│   └── .env.example                # 环境变量模板
│
├── data/                   # 📦 持久化数据（git ignore）
│   ├── uploads/            # 用户上传的文件
│   └── postgres/           # PostgreSQL 数据
│
├── certs/                  # 🔐 SSL 证书（git ignore）
│
├── bin/                    # 🔨 编译输出
│
├── .env                    # 环境变量（敏感文件，git ignore）
├── .gitignore              # Git 忽略配置
├── README.md               # 项目主文档
├── docker-compose.yml      # Docker Compose 配置
├── Dockerfile              # Docker 镜像配置
├── Caddyfile               # Caddy 配置
└── nginx.conf              # Nginx 配置
```

---

## 📚 文档中心详解

### 🚀 [deployment/](deployment/) - 部署指南
快速部署和域名配置

- `QUICK_START.md` - 5分钟快速部署
- `README_DEPLOYMENT.md` - 完整部署文档
- `DOMAIN_SETUP.md` - DNS 和域名配置
- `DOMAIN_SETUP_JIEZISPACE.md` - JieziSpace 域名案例
- `DEPLOYMENT_STATUS.md` - 当前部署状态
- `DEPLOYMENT_FIX.md` - 常见问题解决

### 🔄 [migration/](migration/) - 迁移升级
数据迁移、存储升级、数据库备份

- `STORAGE_MIGRATION_DESIGN.md` - S3 存储架构设计 ⭐
- `IMPLEMENTATION_GUIDE.md` - 迁移实施指南
- `MIGRATION_SUMMARY.md` - 迁移工作总结
- `FILE_UPLOAD_MODIFICATION_COMPLETE.md` - 代码修改说明
- `FINAL_COMPLETION_REPORT.md` - 最终完成报告
- `DATABASE_BACKUP_DESIGN.md` - 数据库备份设计 ⭐

### 🐛 [bugs/](bugs/) - 问题排查
已修复的 Bug 和问题分析

- `CURSOR_BUG_REPORT.md` - 光标定位问题
- `ALIGNMENT_FIX.md` - UI 对齐问题
- `CROSS_DEVICE_SYNC_FIX.md` - 跨设备同步
- `ERROR_ANALYSIS.md` - 错误分析
- `PROGRESS_SAVE_ANALYSIS.md` - 进度保存分析
- `PROGRESS_SAVE_FIX_SUMMARY.md` - 进度保存修复
- `TROUBLESHOOTING.md` - 故障排除指南

### 🔒 [security/](security/) - 安全配置
系统安全加固和最佳实践

- `SECURITY_CONFIGURATION.md` - 安全配置指南
- `SECURITY_IMPLEMENTATION_REPORT.md` - 安全实施报告
- `SECURITY_QUICK_REFERENCE.md` - 安全快速参考
- `SECURITY_SUMMARY.txt` - 安全要点摘要

### 💻 [development/](development/) - 开发文档
开发指南和技术决策

- `FRONTEND_REFACTOR.md` - 前端架构重构
- `TODO` - 功能待办事项
- `test_progress_save.html` - 测试页面

### 📖 [user-guide/](user-guide/) - 用户指南
面向最终用户的使用文档

- `PROGRESS_USER_GUIDE.md` - 进度保存功能使用指南
- `USER_PROGRESS_FEATURE.md` - 用户进度功能说明

---

## 🔧 脚本使用指南

### 部署脚本

```bash
# 本地开发环境部署
./scripts/deploy.sh

# 生产环境部署（含 Nginx 反向代理）
./scripts/deploy-full.sh
```

### 数据库管理

```bash
# 备份数据库到 S3
./scripts/backup-db-to-s3.sh

# 从 S3 恢复数据库
./scripts/restore-db-from-s3.sh

# 迁移本地文件到 S3
go run scripts/migrate-files-to-s3.go
```

### 验证脚本

```bash
# 验证进度保存功能
./scripts/verify_progress_save.sh

# 验证安全配置
./scripts/verify_security.sh
```

---

## ⚙️ 配置文件说明

### 工作配置（根目录）
这些配置文件需要在根目录，因为 Docker Compose 和相关工具会直接读取：

- `docker-compose.yml` - 主 Docker Compose 配置
- `docker-compose.proxy.yml` - 反向代理配置
- `Dockerfile` - 容器镜像构建配置
- `Caddyfile` - Caddy Web 服务器配置
- `nginx.conf` - Nginx 反向代理配置
- `.env` - 环境变量（敏感文件，不提交到 Git）

### 配置备份（config/ 目录）
所有配置文件在 `config/` 目录都有备份副本，便于版本管理和回滚。

---

## 🎯 快速查找

### 我想...

| 需求 | 路径 |
|------|------|
| 快速部署应用 | `scripts/deploy.sh` |
| 了解 S3 迁移 | `docs/migration/STORAGE_MIGRATION_DESIGN.md` |
| 配置数据库备份 | `docs/migration/DATABASE_BACKUP_DESIGN.md` |
| 排查光标问题 | `docs/bugs/CURSOR_BUG_REPORT.md` |
| 加强安全性 | `docs/security/SECURITY_CONFIGURATION.md` |
| 查看开发计划 | `docs/development/TODO` |
| 使用进度保存 | `docs/user-guide/PROGRESS_USER_GUIDE.md` |

---

## 📝 文件命名规范

### 文档命名
- 使用 `UPPER_SNAKE_CASE.md` 格式
- 文件名清晰描述文档内容
- 例如：`STORAGE_MIGRATION_DESIGN.md`

### 脚本命名
- 使用 `kebab-case.sh` 或 `kebab-case.go` 格式
- 动词开头，描述脚本功能
- 例如：`backup-db-to-s3.sh`, `migrate-files-to-s3.go`

### 配置文件
- 使用工具的标准命名
- 例如：`docker-compose.yml`, `Dockerfile`, `.env`

---

## 🔍 版本历史

### v2.0 - 2025-11-20（当前版本）
- ✅ 重组项目目录结构
- ✅ 文档分类存放（6个分类目录）
- ✅ 脚本集中管理
- ✅ 配置文件备份
- ✅ 根目录清理整洁

### v1.0 - 2025-11-15
- 初始项目结构
- 文档散落在根目录
- 缺少统一管理

---

## 🤝 维护指南

### 添加新文档
1. 确定文档类型（部署/迁移/Bug/安全/开发/用户指南）
2. 在对应的 `docs/` 子目录下创建文档
3. 更新 `docs/README.md` 索引
4. 提交 Git 变更

### 添加新脚本
1. 在 `scripts/` 目录创建脚本
2. 添加可执行权限：`chmod +x scripts/your-script.sh`
3. 在本文档中添加使用说明
4. 提交 Git 变更

### 修改配置
1. 修改根目录的工作配置
2. 同步更新 `config/` 目录的备份
3. 更新 `.env.example` 模板（如需要）
4. 提交 Git 变更

---

## 📞 需要帮助？

- 📖 查看 [文档索引](README.md)
- 🚀 查看 [快速开始](deployment/QUICK_START.md)
- 🐛 查看 [问题排查](bugs/)
- 💬 提交 [Issue](https://github.com/yourusername/CodeJym/issues)

---

**文档持续更新中...** 📝

*最后更新：2025-11-20 by Claude*

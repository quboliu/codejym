# CodeJYM - 代码学习平台

> 一个帮助你通过临摹代码来提升编程技能的 Web 应用

## 🚀 快速开始

### 本地开发
```bash
# 启动所有服务（PostgreSQL + 应用）
./scripts/deploy.sh

# 访问应用
open http://localhost:8080
```

### 生产部署（含 Nginx 反向代理）
```bash
# 全功能部署（PostgreSQL + 应用 + Nginx）
./scripts/deploy-full.sh

# 通过域名访问
open http://jiezispace.com
```

---

## 📦 项目结构

```
CodeJym/
├── backend/              # Go 后端服务
│   ├── cmd/             # 应用入口
│   └── internal/        # 内部包
│       ├── api/         # HTTP 处理器
│       └── storage/     # 存储层（PostgreSQL + S3）
├── frontend/            # React 前端应用
├── scripts/             # 部署和运维脚本
│   ├── deploy.sh
│   ├── deploy-full.sh
│   ├── backup-db-to-s3.sh
│   ├── restore-db-from-s3.sh
│   └── migrate-files-to-s3.go
├── config/              # 配置文件备份
│   ├── docker-compose.yml
│   ├── Dockerfile
│   ├── Caddyfile
│   └── nginx.conf
└── docs/                # 📚 完整文档
    ├── deployment/      # 部署指南
    ├── migration/       # 迁移和升级
    ├── bugs/            # Bug 报告
    ├── security/        # 安全配置
    ├── development/     # 开发文档
    └── user-guide/      # 用户指南
```

---

## 📚 文档导航

### 🚀 部署相关
- [快速开始指南](docs/deployment/QUICK_START.md) - 5分钟部署
- [完整部署文档](docs/deployment/README_DEPLOYMENT.md) - 详细步骤
- [域名配置指南](docs/deployment/DOMAIN_SETUP.md) - DNS 和 SSL 设置
- [部署状态报告](docs/deployment/DEPLOYMENT_STATUS.md)

### 🔄 迁移和升级
- [S3 存储迁移设计](docs/migration/STORAGE_MIGRATION_DESIGN.md) - 完整架构设计
- [迁移实施指南](docs/migration/IMPLEMENTATION_GUIDE.md) - 分步操作手册
- [迁移总结](docs/migration/MIGRATION_SUMMARY.md) - 已完成工作和后续步骤
- [数据库备份设计](docs/migration/DATABASE_BACKUP_DESIGN.md) - 备份策略

### 🔒 安全配置
- [安全配置指南](docs/security/SECURITY_CONFIGURATION.md)
- [安全实施报告](docs/security/SECURITY_IMPLEMENTATION_REPORT.md)
- [安全快速参考](docs/security/SECURITY_QUICK_REFERENCE.md)

### 🐛 问题排查
- [已修复的 Bug](docs/bugs/) - 完整的问题分析和解决方案
- [光标定位问题](docs/bugs/CURSOR_BUG_REPORT.md)
- [跨设备同步问题](docs/bugs/CROSS_DEVICE_SYNC_FIX.md)

### 💻 开发文档
- [前端重构文档](docs/development/FRONTEND_REFACTOR.md)
- [开发待办事项](docs/development/TODO)

### 📖 用户指南
- [进度保存功能使用指南](docs/user-guide/PROGRESS_USER_GUIDE.md)

---

## ⚙️ 环境配置

### 环境变量（.env）
```bash
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=codecopy
DB_PASSWORD=codecopy123
DB_NAME=codecopybook

# 存储配置（支持 local 或 s3）
STORAGE_TYPE=local              # 本地存储
# STORAGE_TYPE=s3               # S3 对象存储

# S3 配置（使用 S3 时需要）
S3_ENDPOINT=https://s3.bitiful.net
S3_REGION=cn-east-1
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=codejym-uploads
S3_BACKUP_BUCKET=codejym-backups
```

详细配置说明请参考 [.env.example](.env.example)

---

## 🔧 管理命令

### Docker 服务管理
```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f codecopybook    # 应用日志
docker compose logs -f postgres        # 数据库日志

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 完全清理（删除所有数据）
docker compose down -v
```

### 数据库备份
```bash
# 备份到 S3
./scripts/backup-db-to-s3.sh

# 从 S3 恢复
./scripts/restore-db-from-s3.sh
```

---

## 🏗️ 技术栈

### 后端
- **语言**：Go 1.21+
- **Web 框架**：标准库 net/http
- **数据库**：PostgreSQL 15
- **存储**：本地文件系统 / S3 兼容对象存储
- **容器化**：Docker + Docker Compose

### 前端
- **框架**：React 18
- **语言**：TypeScript
- **构建工具**：Vite
- **样式**：CSS Modules

---

## 📊 系统架构

```
┌─────────────────────────────────────────────────┐
│                    用户                          │
│                     ↓                            │
│              ┌──────────────┐                    │
│              │    Nginx     │  (可选反向代理)    │
│              │  Port: 80    │                    │
│              └──────┬───────┘                    │
│                     ↓                            │
│              ┌──────────────┐                    │
│              │  前端 (React) │                   │
│              │  Port: 8080  │                    │
│              └──────┬───────┘                    │
│                     ↓                            │
│              ┌──────────────┐                    │
│              │ 后端 (Go API) │                   │
│              │  Port: 8080  │                    │
│              └──────┬───────┘                    │
│                     ↓                            │
│         ┌───────────┴──────────┐                 │
│         ↓                      ↓                 │
│  ┌─────────────┐      ┌────────────────┐        │
│  │ PostgreSQL  │      │  文件存储      │        │
│  │  Port: 5432 │      │ (Local / S3)   │        │
│  └─────────────┘      └────────────────┘        │
└─────────────────────────────────────────────────┘
```

---

## 🌟 核心特性

- ✅ **代码临摹练习** - 通过临摹代码提升编程能力
- ✅ **语法高亮** - 支持 20+ 编程语言
- ✅ **进度保存** - 自动保存练习进度
- ✅ **文件管理** - 支持上传 ZIP、单文件、粘贴代码
- ✅ **灵活存储** - 支持本地存储和 S3 对象存储
- ✅ **自动备份** - 数据库定时备份到 S3
- ✅ **跨设备同步** - 用户数据云端存储

---

## 🔗 访问地址

### 本地开发
- 应用：http://localhost:8080
- API：http://localhost:8080/api
- 数据库：localhost:5432

### 生产环境
- 应用：http://jiezispace.com
- API：http://jiezispace.com/api

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

详细开发文档请参考 [开发文档](docs/development/)

---

## 📄 许可证

MIT License

---

## 📞 联系方式

- 项目仓库：[GitHub](https://github.com/yourusername/CodeJym)
- 问题反馈：[Issues](https://github.com/yourusername/CodeJym/issues)

---

**享受你的代码学习之旅！** 💻✨

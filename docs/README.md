# CodeJYM 文档中心

欢迎来到 CodeJYM 的文档中心！这里包含了所有的技术文档、部署指南、问题排查和开发文档。

---

## 📚 文档分类

### 🚀 [部署相关](deployment/)
快速部署和域名配置指南

| 文档 | 说明 | 难度 |
|------|------|------|
| [快速开始](deployment/QUICK_START.md) | 5分钟快速部署指南 | ⭐ |
| [完整部署文档](deployment/README_DEPLOYMENT.md) | 详细的部署步骤和说明 | ⭐⭐ |
| [域名配置](deployment/DOMAIN_SETUP.md) | DNS 和 SSL 配置指南 | ⭐⭐ |
| [JieziSpace 域名配置](deployment/DOMAIN_SETUP_JIEZISPACE.md) | 特定域名配置案例 | ⭐⭐ |
| [部署状态报告](deployment/DEPLOYMENT_STATUS.md) | 当前部署状态和版本信息 | ⭐ |
| [部署问题修复](deployment/DEPLOYMENT_FIX.md) | 常见部署问题解决方案 | ⭐⭐ |

---

### 🔄 [迁移和升级](migration/)
数据迁移、存储升级和数据库备份

| 文档 | 说明 | 难度 |
|------|------|------|
| [S3 存储迁移设计](migration/STORAGE_MIGRATION_DESIGN.md) | 完整的存储架构设计文档 | ⭐⭐⭐ |
| [迁移实施指南](migration/IMPLEMENTATION_GUIDE.md) | 分步操作手册 | ⭐⭐⭐ |
| [迁移总结](migration/MIGRATION_SUMMARY.md) | 已完成工作和后续步骤 | ⭐ |
| [文件上传修改说明](migration/FILE_UPLOAD_MODIFICATION_COMPLETE.md) | 代码修改详细说明 | ⭐⭐ |
| [完成报告](migration/FINAL_COMPLETION_REPORT.md) | 迁移工作最终报告 | ⭐ |
| [数据库备份设计](migration/DATABASE_BACKUP_DESIGN.md) | 数据库备份策略和实施方案 | ⭐⭐⭐ |

**关键内容：**
- 📦 从本地存储迁移到 S3 对象存储
- 🔧 存储抽象层设计（支持 Local/S3 无缝切换）
- 💾 自动化数据库备份到 S3
- 🔄 历史数据批量迁移工具

---

### 🔒 [安全配置](security/)
系统安全加固和最佳实践

| 文档 | 说明 | 难度 |
|------|------|------|
| [安全配置指南](security/SECURITY_CONFIGURATION.md) | 完整的安全配置步骤 | ⭐⭐ |
| [安全实施报告](security/SECURITY_IMPLEMENTATION_REPORT.md) | 已实施的安全措施 | ⭐ |
| [安全快速参考](security/SECURITY_QUICK_REFERENCE.md) | 常用安全命令和检查清单 | ⭐ |
| [安全总结](security/SECURITY_SUMMARY.txt) | 安全要点摘要 | ⭐ |

**覆盖内容：**
- 🔐 认证和授权（JWT Token）
- 🚫 禁用 Cookies（使用 Bearer Token）
- 🛡️ HTTPS 和安全头配置
- 📝 敏感数据保护（.env、密钥管理）

---

### 🐛 [问题排查](bugs/)
已修复的 Bug 和问题分析报告

| 文档 | 说明 | 类型 |
|------|------|------|
| [光标定位问题](bugs/CURSOR_BUG_REPORT.md) | 完整的光标 Bug 排查报告 | 🐛 Bug 修复 |
| [对齐问题修复](bugs/ALIGNMENT_FIX.md) | UI 对齐问题解决方案 | 🐛 Bug 修复 |
| [跨设备同步问题](bugs/CROSS_DEVICE_SYNC_FIX.md) | 进度同步问题修复 | 🐛 Bug 修复 |
| [错误分析](bugs/ERROR_ANALYSIS.md) | 系统错误分析和解决 | 📊 分析报告 |
| [进度保存分析](bugs/PROGRESS_SAVE_ANALYSIS.md) | 进度保存功能分析 | 📊 分析报告 |
| [进度保存修复总结](bugs/PROGRESS_SAVE_FIX_SUMMARY.md) | 进度保存问题修复总结 | 🐛 Bug 修复 |

**价值：**
- 📖 学习如何系统性地排查和修复问题
- 🔍 了解常见问题的根本原因
- ✅ 避免重复遇到相同问题

---

### 💻 [开发文档](development/)
开发指南和技术决策

| 文档 | 说明 | 类型 |
|------|------|------|
| [前端重构文档](development/FRONTEND_REFACTOR.md) | 前端架构重构说明 | 📐 架构设计 |
| [开发待办事项](development/TODO) | 未来功能和改进计划 | 📝 计划 |
| [进度保存测试](development/test_progress_save.html) | 测试页面 | 🧪 测试 |

---

### 📖 [用户指南](user-guide/)
面向最终用户的使用指南

| 文档 | 说明 |
|------|------|
| [进度保存功能使用指南](user-guide/PROGRESS_USER_GUIDE.md) | 如何使用进度保存功能 |

---

## 🔍 快速查找

### 我想...

#### 🚀 部署应用
→ 先看 [快速开始](deployment/QUICK_START.md)
→ 遇到问题看 [部署问题修复](deployment/DEPLOYMENT_FIX.md)

#### 🔧 配置 S3 对象存储
→ 看 [S3 存储迁移设计](migration/STORAGE_MIGRATION_DESIGN.md)
→ 按照 [迁移实施指南](migration/IMPLEMENTATION_GUIDE.md) 操作

#### 💾 设置数据库备份
→ 看 [数据库备份设计](migration/DATABASE_BACKUP_DESIGN.md)
→ 使用 `/scripts/backup-db-to-s3.sh` 脚本

#### 🔒 加强安全性
→ 看 [安全配置指南](security/SECURITY_CONFIGURATION.md)
→ 参考 [安全快速参考](security/SECURITY_QUICK_REFERENCE.md)

#### 🐛 排查问题
→ 看 [问题排查](bugs/) 目录
→ 查找类似的问题报告

#### 💻 开始开发
→ 看 [前端重构文档](development/FRONTEND_REFACTOR.md)
→ 查看 [开发待办事项](development/TODO)

---

## 📊 文档统计

- **总文档数**：27 篇
- **部署文档**：6 篇
- **迁移文档**：6 篇
- **安全文档**：4 篇
- **Bug 报告**：6 篇
- **开发文档**：3 篇
- **用户指南**：1 篇

---

## 🆕 最近更新

| 日期 | 文档 | 变更 |
|------|------|------|
| 2025-11-20 | S3 存储迁移 | 完成所有迁移代码实现 |
| 2025-11-19 | 数据库备份 | 添加自动化备份脚本 |
| 2025-11-18 | 光标定位修复 | 修复光标定位 Bug |
| 2025-11-17 | 安全配置 | 实施安全加固措施 |
| 2025-11-16 | 进度保存 | 修复跨设备同步问题 |

---

## 💡 文档贡献

如果你发现文档有误或需要补充：

1. 在对应文档目录下创建或修改文档
2. 更新本索引文件（docs/README.md）
3. 提交 Pull Request

---

## 📞 需要帮助？

- 查看根目录 [README.md](../README.md) 了解项目概况
- 查看 [快速开始](deployment/QUICK_START.md) 快速上手
- 提交 [Issue](https://github.com/yourusername/CodeJym/issues) 报告问题

---

**文档持续更新中...** 📝

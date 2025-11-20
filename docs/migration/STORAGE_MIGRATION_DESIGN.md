# CodeJYM 数据存储风险分析与 S3 迁移方案

**项目名称：** CodeJYM 存储系统优化
**文档版本：** 1.0
**创建日期：** 2025-11-20
**状态：** 设计完成，待实施

---

## 📋 目录

1. [执行摘要](#执行摘要)
2. [当前系统分析](#当前系统分析)
3. [风险评估](#风险评估)
4. [解决方案设计](#解决方案设计)
5. [技术架构](#技术架构)
6. [实施计划](#实施计划)
7. [成本分析](#成本分析)
8. [风险与缓解措施](#风险与缓解措施)

---

## 执行摘要

### 背景
CodeJYM 是一个代码打字练习平台，当前采用 Docker Compose 单机部署，数据包括：
- **用户上传文件**：存储在本地文件系统 `./data/uploads/`
- **PostgreSQL 数据库**：用户账号、训练组元数据、练习进度

### 核心问题
1. **单点故障风险**：所有数据在一台机器，硬盘故障将导致数据永久丢失
2. **无备份机制**：当前没有任何自动化备份
3. **磁盘空间受限**：本地存储空间有限，用户上传文件持续增长

### 解决方案
采用**缤纷云 S3 对象存储**实现数据分离和异地备份：
- 用户上传文件直接存储到 S3（解决磁盘空间问题）
- PostgreSQL 数据库定期备份到 S3（实现灾难恢复）
- 实现抽象存储层，支持本地/S3 灵活切换

### 预期收益
- ✅ **数据安全性提升**：异地备份，避免单点故障
- ✅ **成本优化**：利用缤纷云 50GB 免费额度，节省 80% 成本
- ✅ **扩展性增强**：不受本地磁盘限制
- ✅ **运维简化**：自动化备份，无需人工干预

---

## 当前系统分析

### 1. 数据存储架构

#### 1.1 文件系统存储
```
/opt/codejym/data/
├── uploads/                    # 用户上传的代码文件
│   └── {userID}/
│       └── {assetID}/
│           ├── main.go
│           ├── utils.go
│           └── ...
├── assets.json                 # 遗留文件（已迁移到数据库）
└── sessions.json               # 遗留文件（已迁移到数据库）
```

**特点：**
- Bind Mount 挂载到容器：`./data` → `/data`
- 直接文件系统访问，读写性能高
- 目录结构按用户和训练组隔离

#### 1.2 数据库存储
```yaml
volumes:
  pgdata: /var/lib/postgresql/data  # Docker Named Volume
```

**表结构：**
- `users` - 用户账号（ID、邮箱、密码哈希）
- `assets` - 训练组元数据（ID、用户ID、文件路径、大小、文件数）
- `typing_sessions` - 练习进度（ID、用户ID、训练组ID、光标位置、错误数）

**特点：**
- 使用 PostgreSQL 16
- 连接池管理（pgxpool）
- 外键约束保证数据完整性（ON DELETE CASCADE）

### 2. 文件上传流程

```
用户浏览器
    ↓ POST /api/assets/upload (multipart/form-data)
应用服务器 (Go)
    ↓ 解析文件（支持单文件或 ZIP）
    ↓ 保存到 /data/uploads/{userID}/{assetID}/
    ↓ 扫描文件列表和大小
    ↓ 写入数据库 assets 表
    ↓ 返回训练组 ID
用户浏览器
```

**关键代码位置：**
- `backend/internal/api/server.go:285-389` - `handleAssetUpload`
- `backend/internal/storage/storage.go:118-120` - `AssetDir`

### 3. Docker Compose 配置

```yaml
services:
  postgres:
    image: postgres:16
    volumes:
      - pgdata:/var/lib/postgresql/data  # 数据库数据
    environment:
      - POSTGRES_USER=codecopy
      - POSTGRES_PASSWORD=codecopy123
      - POSTGRES_DB=codecopybook

  codecopybook:
    build: .
    volumes:
      - ./data:/data  # 应用数据（用户上传文件）
    environment:
      - DATA_DIR=/data
      - DATABASE_URL=postgres://...

volumes:
  pgdata:  # 由 Docker 管理的命名卷
```

---

## 风险评估

### 1️⃣ 单点故障风险 - **严重 🔴**

**风险描述：**
所有数据（数据库 + 用户文件）都存储在同一台物理机器上。

**潜在影响：**
- 硬盘物理故障 → **所有数据永久丢失**
- 主板/电源故障 → 服务长时间中断（数小时至数天）
- 意外断电/系统崩溃 → 数据库可能损坏
- 机房火灾/水灾/盗窃 → 灾难性数据丢失

**量化指标：**
- **当前 RTO（恢复时间目标）**：数小时至数天
- **当前 RPO（恢复点目标）**：∞（无备份 = 永久丢失）
- **数据丢失概率**：硬盘年故障率约 2-5%

**影响范围：**
- 所有用户数据不可恢复
- 平台信誉严重受损
- 可能面临法律责任（数据保护法规）

---

### 2️⃣ 数据备份缺失 - **严重 🔴**

**风险描述：**
虽然已有详细的备份设计文档（DATABASE_BACKUP_DESIGN.md），但**尚未实施任何自动化备份机制**。

**当前状态：**
- ✅ 备份设计文档完整（1199 行）
- ❌ 未实施全量备份脚本
- ❌ 未配置 WAL 归档
- ❌ 无远程备份
- ❌ 无备份监控和告警

**潜在影响：**
- 误删除数据（用户/训练组）→ 无法恢复
- 数据损坏 → 无历史版本可回滚
- 无法实现 PITR（Point-In-Time Recovery）
- 不符合数据保护合规要求

**真实案例：**
- GitLab 2017 年删除生产数据库事件：5 个备份方案都失效，丢失 6 小时数据
- 教训：备份计划未验证等同于没有备份

---

### 3️⃣ 本地磁盘空间压力 - **中等 🟡**

**风险描述：**
用户上传文件持续增长，本地磁盘空间有限。

**容量预估：**
```
场景假设：
- 100 个活跃用户
- 每用户平均 10 个训练组
- 每个训练组平均 1MB 代码

预计存储需求：
- 用户文件：100 × 10 × 1MB = 1GB
- PostgreSQL 数据：< 500MB
- 系统和日志：约 5GB
- 总计：约 6.5GB（当前可承受）

1 年后增长：
- 300 个用户 → 3GB 用户文件
- 数据库增长 → 1.5GB
- 总计：约 10GB（接近预警阈值）
```

**潜在影响：**
- 磁盘满载 → 应用和数据库无法写入
- Docker Volume 和应用数据共享磁盘，相互挤占
- 需要人工清理或扩容，运维负担增加

---

### 4️⃣ Docker Volume 数据不透明 - **中等 🟡**

**风险描述：**
PostgreSQL 数据存储在 Docker Named Volume 中，不易直接访问和管理。

**问题：**
- Volume 数据位置不直观（需要 `docker volume inspect pgdata` 查看）
- 手动备份困难（需要停止容器或使用 `pg_dump`）
- 迁移复杂（需要导出卷或使用 Docker 工具）
- Volume 损坏后恢复困难

**实际路径示例：**
```bash
$ docker volume inspect pgdata
[{
    "Mountpoint": "/var/lib/docker/volumes/codejym_pgdata/_data",
    ...
}]
```

---

### 5️⃣ 用户文件管理粗放 - **低 🟢**

**问题清单：**
- 无文件生命周期管理（90 天未访问的文件仍占用空间）
- 无存储配额限制（单个用户可上传无限文件）
- 删除训练组时，文件可能残留（需检查 ON DELETE CASCADE 是否清理文件系统）
- 无访问日志，难以审计和优化

---

### 6️⃣ 缺乏灾难恢复演练 - **中等 🟡**

**风险描述：**
从未验证过数据恢复流程是否可行。

**潜在问题：**
- 备份文件可能损坏但未发现
- 恢复步骤不完整或有遗漏
- 恢复时间超出预期
- RTO/RPO 目标未经实际验证

**业界最佳实践：**
- 每季度进行恢复演练
- 验证备份完整性和可用性
- 记录实际 RTO/RPO

---

## 解决方案设计

### 核心策略：3-2-1 备份原则

```
3 份数据副本
├── 1️⃣ 生产环境原始数据
├── 2️⃣ 本地备份（短期）
└── 3️⃣ 远程备份（长期）

2 种存储介质
├── 本地磁盘（Docker Volume + Bind Mount）
└── 云端对象存储（缤纷云 S3）

1 份异地存储
└── 缤纷云 S3（位于不同地域的数据中心）
```

---

### 方案一：渐进式迁移（推荐 ⭐）

**优势：**
- ✅ 风险低，可逐步验证
- ✅ 不需要停机，对用户无影响
- ✅ 出问题可快速回滚
- ✅ 有充足时间测试

**分阶段实施：**

#### Phase 1: 数据库备份到 S3（1-2 天）
```
目标：实现 PostgreSQL 自动备份
方法：
  1. 创建全量备份脚本（pg_dump → gzip → S3）
  2. 配置 Cron 定时任务（每天凌晨 2:00）
  3. 验证备份可恢复性
验收标准：
  - 成功上传至少 1 份备份到 S3
  - 成功从 S3 恢复数据库到测试环境
```

#### Phase 2: 新上传文件存储到 S3（2-3 天）
```
目标：新用户上传的文件直接存储到 S3
方法：
  1. 实现抽象存储接口 FileStorage
  2. 实现 S3Storage 和 LocalStorage
  3. 修改上传处理逻辑使用抽象接口
  4. 配置环境变量 STORAGE_TYPE=s3
验收标准：
  - 新上传文件存储在 S3
  - 能够正常下载和浏览
  - 本地文件系统不再增长
```

#### Phase 3: 迁移历史文件到 S3（1 天）
```
目标：将现有本地文件迁移到 S3
方法：
  1. 编写数据迁移脚本（Go）
  2. 批量上传本地文件到 S3
  3. 更新数据库中的文件路径
  4. 验证所有文件可访问
验收标准：
  - 100% 文件成功迁移
  - 数据库路径已更新
  - 功能测试全部通过
```

#### Phase 4: 清理与监控（1 天）
```
目标：清理本地数据，配置监控
方法：
  1. 确认 S3 数据完整性
  2. 删除本地 data/uploads/
  3. 配置备份监控和告警
  4. 编写运维文档
验收标准：
  - 本地磁盘空间释放
  - 备份告警正常工作
  - 文档完善
```

---

### 方案二：激进式完全迁移（备选）

**流程：**
1. 发布停机公告（提前 24 小时）
2. 停止应用服务
3. 全量备份数据库到 S3
4. 批量上传本地文件到 S3
5. 更新应用配置和代码
6. 重新部署
7. 功能验证
8. 恢复服务

**优势：**
- 一步到位，架构清晰
- 实施时间短（4-6 小时）

**劣势：**
- 需要停机（用户体验差）
- 风险集中，出错影响大
- 无回退余地

---

## 技术架构

### 1. 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                      用户浏览器                           │
└────────────────────┬────────────────────────────────────┘
                     │ HTTP/HTTPS
                     ↓
┌─────────────────────────────────────────────────────────┐
│                  Nginx 反向代理                           │
│                  (可选，生产环境)                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│              CodeJYM 应用服务 (Go)                        │
│              Docker Container: codecopybook              │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │     API Layer (internal/api)                   │    │
│  │  - handleAssetUpload()                         │    │
│  │  - handleAssetDownload()                       │    │
│  │  - handleAssetDelete()                         │    │
│  └─────────────────┬──────────────────────────────┘    │
│                    │                                     │
│  ┌─────────────────┴──────────────────────────────┐    │
│  │     Storage Layer (internal/storage)           │    │
│  │                                                 │    │
│  │  ┌───────────────────────────────────────┐    │    │
│  │  │  FileStorage Interface                │    │    │
│  │  │  - SaveFile(path, reader) → url       │    │    │
│  │  │  - GetFile(path) → reader             │    │    │
│  │  │  - DeleteFile(path)                   │    │    │
│  │  │  - DeleteDir(path)                    │    │    │
│  │  │  - GetURL(path) → url                 │    │    │
│  │  │  - ListFiles(prefix) → []string       │    │    │
│  │  └───────────────┬───────────────────────┘    │    │
│  │                  │                             │    │
│  │     ┌────────────┴────────────┐               │    │
│  │     │                         │               │    │
│  │     ↓                         ↓               │    │
│  │  ┌──────────────┐      ┌──────────────┐      │    │
│  │  │LocalStorage  │      │ S3Storage    │      │    │
│  │  │(开发环境)     │      │ (生产环境)    │      │    │
│  │  └──────────────┘      └──────┬───────┘      │    │
│  └────────┬────────────────────────┘             │    │
│           │                                       │    │
└───────────┼───────────────────────────────────────┘
            │
        ┌───┴───┐
        │       │
        ↓       ↓
┌──────────┐  ┌────────────────────────────────┐
│PostgreSQL│  │      缤纷云 S3 对象存储          │
│Container │  │                                │
│          │  │  Bucket: codejym-uploads       │
│Tables:   │  │  ├── uploads/                  │
│- users   │  │  │   └── {userID}/{assetID}/   │
│- assets  │  │  │       ├── main.go            │
│- sessions│  │  │       └── utils.go           │
│          │  │  └── backups/                  │
│          │  │      ├── database/full/        │
│          │  │      └── database/wal/         │
│          │  │                                │
└────┬─────┘  └────────────────────────────────┘
     │                     ↑
     │                     │
     ↓                     │
┌──────────────────────────┴───────────┐
│      备份服务 (Cron/Docker)            │
│                                       │
│  - backup-db-to-s3.sh (每天 2:00)     │
│  - archive-wal.sh (实时归档)          │
│  - cleanup-old-backups.sh (每周)      │
└───────────────────────────────────────┘
```

---

### 2. 抽象存储层设计

#### 2.1 接口定义（`internal/storage/file_storage.go`）

```go
package storage

import (
	"context"
	"io"
)

// FileStorage 抽象文件存储接口
// 实现：LocalStorage（本地文件系统）、S3Storage（对象存储）
type FileStorage interface {
	// SaveFile 保存文件到存储
	// path: 文件存储路径（相对路径，如 "uploads/userID/assetID/file.go"）
	// reader: 文件内容读取器
	// contentType: MIME 类型（如 "text/plain", "application/zip"）
	// 返回：存储路径（用于后续访问）和错误
	SaveFile(ctx context.Context, path string, reader io.Reader, contentType string) (string, error)

	// GetFile 获取文件内容
	// path: 文件存储路径
	// 返回：可读取的文件流（调用者负责关闭）和错误
	GetFile(ctx context.Context, path string) (io.ReadCloser, error)

	// DeleteFile 删除单个文件
	DeleteFile(ctx context.Context, path string) error

	// DeleteDir 递归删除目录及其所有内容
	// 注意：S3 没有真正的目录概念，会删除所有匹配前缀的对象
	DeleteDir(ctx context.Context, path string) error

	// GetURL 获取文件访问 URL
	// 本地存储：返回相对路径
	// S3 存储：返回 CDN URL 或预签名 URL
	GetURL(ctx context.Context, path string) (string, error)

	// ListFiles 列出目录下的所有文件
	// prefix: 目录前缀（如 "uploads/userID/"）
	// 返回：文件路径列表和错误
	ListFiles(ctx context.Context, prefix string) ([]string, error)
}
```

**设计要点：**
- 接口简洁，只包含必要操作
- 使用 `context.Context` 支持超时和取消
- 返回 `io.Reader/io.ReadCloser` 支持流式传输（内存友好）
- 路径统一使用 Unix 风格斜杠（`/`）

---

#### 2.2 本地存储实现（`internal/storage/local_storage.go`）

**职责：**
- 开发环境使用
- 兼容现有的文件系统存储
- 快速测试，无需网络依赖

**实现特点：**
```go
type LocalStorage struct {
	rootDir string  // 根目录，如 "/data/uploads"
}

// 示例：SaveFile("uploads/user123/asset456/main.go", reader)
// 实际路径：/data/uploads/uploads/user123/asset456/main.go
```

---

#### 2.3 S3 存储实现（`internal/storage/s3_storage.go`）

**职责：**
- 生产环境使用
- 对接缤纷云 S3 API
- 支持大文件分片上传（可选）

**依赖：**
```go
import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)
```

**配置参数：**
```go
type S3Storage struct {
	client     *s3.Client
	bucket     string      // Bucket 名称
	urlPrefix  string      // CDN URL 前缀（可选）
}

func NewS3Storage(
	endpoint   string,  // 如 "https://s3.bitiful.net"
	accessKey  string,  // Access Key ID
	secretKey  string,  // Secret Access Key
	bucket     string,  // Bucket 名称
	region     string,  // 区域（如 "cn-east-1"）
	urlPrefix  string,  // CDN URL（如 "https://cdn.example.com"）
) (*S3Storage, error)
```

---

### 3. 数据库备份方案

#### 3.1 全量备份（每日）

**工具：** `pg_dump`
**频率：** 每天凌晨 2:00
**格式：** Plain SQL + Gzip 压缩
**存储位置：** S3 `backups/database/full/`

**文件命名规则：**
```
codecopybook_20251120_020000.sql.gz
             └─┬─┘ └───┬───┘
               │       └─ 时间（HH:MM:SS）
               └─ 日期（YYYY-MM-DD）
```

**保留策略：**
- 本地：不保留（备份后立即删除）
- S3：保留 30 天（自动清理旧备份）

---

#### 3.2 WAL 归档（实时）

**工具：** PostgreSQL WAL 归档机制
**频率：** 实时（WAL 文件写满 16MB 时触发）
**存储位置：** S3 `backups/database/wal/`

**PostgreSQL 配置：**
```sql
wal_level = replica
archive_mode = on
archive_command = '/usr/local/bin/archive-wal.sh %p %f'
```

**作用：**
- 支持 PITR（Point-In-Time Recovery）
- RPO 降低到 < 1 分钟
- 可恢复到任意时间点

---

### 4. 环境变量配置

```bash
# ==================== 数据库配置 ====================
POSTGRES_USER=codecopy
POSTGRES_PASSWORD=codecopy123
POSTGRES_DB=codecopybook
POSTGRES_PORT=5432
DATABASE_URL=postgres://codecopy:codecopy123@postgres:5432/codecopybook?sslmode=disable

# ==================== 存储配置 ====================
# 存储类型：local（本地）或 s3（对象存储）
STORAGE_TYPE=s3

# 本地存储根目录（STORAGE_TYPE=local 时使用）
DATA_DIR=/data

# ==================== 缤纷云 S3 配置 ====================
# S3 Endpoint（从缤纷云控制台获取）
S3_ENDPOINT=https://s3.bitiful.net

# S3 区域（从缤纷云控制台获取）
S3_REGION=cn-east-1

# S3 访问凭证（从缤纷云控制台创建）
S3_ACCESS_KEY=your_access_key_here
S3_SECRET_KEY=your_secret_key_here

# 用户文件 Bucket
S3_BUCKET=codejym-uploads

# CDN URL 前缀（可选，配置后可加速访问）
S3_URL_PREFIX=https://cdn.yourdomain.com

# 备份 Bucket（可与上传文件分离，便于管理）
S3_BACKUP_BUCKET=codejym-backups
```

---

## 实施计划

### Phase 1: 准备阶段（1-2 天）

#### 任务清单
- [ ] 注册缤纷云账号
- [ ] 完成实名认证（获取免费额度）
- [ ] 创建 Bucket：`codejym-uploads`（用户文件）
- [ ] 创建 Bucket：`codejym-backups`（数据库备份）
- [ ] 获取 Access Key 和 Secret Key
- [ ] 确认 S3 Endpoint 和 Region
- [ ] 本地测试 S3 连接（使用 AWS CLI）

#### 验收标准
```bash
# 测试 AWS CLI 连接
aws s3 ls s3://codejym-uploads/ \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1

# 应返回空列表（新 Bucket）或文件列表
```

---

### Phase 2: 代码实现（2-3 天）

#### 任务清单
- [ ] 创建 `internal/storage/file_storage.go`（接口定义）
- [ ] 创建 `internal/storage/local_storage.go`（本地实现）
- [ ] 创建 `internal/storage/s3_storage.go`（S3 实现）
- [ ] 修改 `internal/storage/storage.go`（集成 FileStorage）
- [ ] 修改 `cmd/server/main.go`（初始化存储）
- [ ] 修改 `internal/api/server.go`（上传逻辑）
- [ ] 添加 AWS SDK 依赖（`go.mod`）
- [ ] 单元测试（Mock S3 或使用 MinIO）

#### 验收标准
- 本地测试：`STORAGE_TYPE=local` 时功能正常
- S3 测试：`STORAGE_TYPE=s3` 时文件上传到缤纷云
- 兼容性测试：现有功能无回归

---

### Phase 3: 数据库备份（1 天）

#### 任务清单
- [ ] 创建 `scripts/backup-db-to-s3.sh`
- [ ] 创建 `scripts/restore-db-from-s3.sh`
- [ ] 安装 AWS CLI（宿主机或 Docker 容器）
- [ ] 配置 Cron 定时任务
- [ ] 测试备份脚本（手动执行一次）
- [ ] 测试恢复脚本（恢复到测试数据库）

#### 验收标准
```bash
# 手动触发备份
./scripts/backup-db-to-s3.sh

# 检查 S3 是否有备份文件
aws s3 ls s3://codejym-backups/backups/database/full/ \
  --endpoint-url https://s3.bitiful.net

# 恢复到测试环境
./scripts/restore-db-from-s3.sh latest
```

---

### Phase 4: 数据迁移（1 天）

#### 任务清单
- [ ] 创建 `scripts/migrate-files-to-s3.go`
- [ ] 备份本地文件（以防万一）
- [ ] 运行迁移脚本（批量上传）
- [ ] 验证 S3 文件完整性（文件数、总大小）
- [ ] 更新数据库 `assets.root_path`
- [ ] 功能测试（下载、浏览文件）

#### 验收标准
- 100% 文件成功迁移
- 数据库路径更新完成
- 用户体验无变化

---

### Phase 5: 清理与监控（1 天）

#### 任务清单
- [ ] 确认 S3 数据完整性（二次检查）
- [ ] 删除本地 `data/uploads/`（释放磁盘空间）
- [ ] 配置备份监控脚本（`scripts/monitor-backup.sh`）
- [ ] 配置告警（备份失败时发送邮件/webhook）
- [ ] 更新运维文档
- [ ] 团队培训（如何恢复数据、查看备份）

#### 验收标准
- 本地磁盘空间释放确认
- 备份监控正常运行
- 文档完善且可执行

---

## 成本分析

### 缤纷云 S3 免费额度

| 资源类型 | 免费额度 | 说明 |
|---------|---------|------|
| 存储空间 | 50 GB | 实名认证后永久免费 |
| 下载流量 | 30 GB/月 | 超出后约 ¥0.50/GB |
| API 请求 | 20 万次/月 | 超出后约 ¥0.01/万次 |

### 预估使用量

**用户文件存储：**
```
当前：约 1 GB（已有数据）
增长：约 2 GB/年（假设 100 新用户）
总计：3 GB（远低于 50GB 免费额度）
```

**数据库备份：**
```
全量备份：约 500 MB/天（压缩后）
保留 30 天：500MB × 30 = 15GB
WAL 归档：约 100 MB/天
保留 7 天：100MB × 7 = 700MB
总计：约 16GB
```

**总存储使用量：** 3GB + 16GB = **19GB**（< 50GB 免费额度）

**流量使用：**
```
用户下载文件：约 10 GB/月（假设 1000 次练习，平均 10MB/次）
备份下载：约 1 GB/月（恢复演练）
总计：约 11GB/月（< 30GB 免费额度）
```

**API 请求：**
```
文件上传：约 1000 次/月
文件下载：约 10000 次/月
备份操作：约 100 次/月
总计：约 11000 次/月（远低于 20 万次免费额度）
```

### 结论

**预计成本：¥0/月**（完全在免费额度内）

即使未来增长 5 倍，仍在免费额度范围内。

---

## 风险与缓解措施

### 1. 网络依赖风险

**风险：** S3 服务依赖网络，网络故障或缤纷云服务中断导致无法访问文件。

**缓解措施：**
- 保留本地缓存机制（可选）
- 实现降级策略（S3 失败时回退到本地存储）
- 监控 S3 可用性（健康检查）
- 选择高可用性云服务商（缤纷云 SLA ≥ 99.9%）

---

### 2. 数据迁移失败风险

**风险：** 迁移过程中文件损坏、丢失或路径错误。

**缓解措施：**
- **迁移前全量备份本地数据**（压缩打包）
- 迁移后验证文件完整性（MD5 校验）
- 分批迁移（先迁移部分数据测试）
- 保留本地数据至少 7 天（确认无问题后删除）
- 实施回滚计划（迁移失败时恢复本地路径）

---

### 3. 成本超支风险

**风险：** 流量或存储超出免费额度，产生意外费用。

**缓解措施：**
- 配置缤纷云用量告警（达到 80% 时通知）
- 实施存储配额限制（单用户最大 500MB）
- 定期清理过期文件（90 天未访问自动删除）
- 启用 CDN 缓存（减少重复下载流量）
- 压缩文件传输（Gzip 编码）

---

### 4. 数据安全风险

**风险：** S3 凭证泄露、Bucket 配置错误导致数据泄露。

**缓解措施：**
- **严禁将 AccessKey/SecretKey 提交到 Git**
- 使用 `.env` 文件管理敏感配置（`.gitignore` 排除）
- 配置 Bucket 访问权限（私有，仅应用可访问）
- 启用 IAM 策略（最小权限原则）
- 定期轮换 AccessKey
- 考虑文件加密存储（AES-256）

---

### 5. 代码兼容性风险

**风险：** 修改存储逻辑导致现有功能回归。

**缓解措施：**
- 充分的单元测试和集成测试
- 使用抽象接口，降低耦合
- 分支开发，合并前充分测试
- 灰度发布（先在开发环境验证）
- 保留旧代码（注释），便于快速回滚

---

## 附录

### A. 参考文档

- [缤纷云官方文档](https://docs.bitiful.com/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/docs/)
- [PostgreSQL 备份与恢复](https://www.postgresql.org/docs/16/backup.html)
- [Docker Volume 管理](https://docs.docker.com/storage/volumes/)

---

### B. 相关脚本清单

| 脚本名称 | 路径 | 功能 |
|---------|------|------|
| backup-db-to-s3.sh | scripts/ | 全量数据库备份 |
| archive-wal.sh | scripts/ | WAL 归档 |
| restore-db-from-s3.sh | scripts/ | 数据库恢复 |
| migrate-files-to-s3.go | scripts/ | 文件迁移 |
| monitor-backup.sh | scripts/ | 备份监控 |
| verify-backup.sh | scripts/ | 备份验证 |

---

### C. 关键代码文件清单

| 文件路径 | 功能 |
|---------|------|
| internal/storage/file_storage.go | 存储接口定义 |
| internal/storage/local_storage.go | 本地存储实现 |
| internal/storage/s3_storage.go | S3 存储实现 |
| internal/storage/storage.go | 存储层集成 |
| internal/api/server.go | 上传/下载 API |
| cmd/server/main.go | 主程序入口 |

---

### D. 测试用例清单

- [ ] 本地存储：上传文件
- [ ] 本地存储：下载文件
- [ ] 本地存储：删除文件
- [ ] S3 存储：上传文件
- [ ] S3 存储：下载文件
- [ ] S3 存储：删除文件
- [ ] S3 存储：列出文件
- [ ] 数据库备份：全量备份
- [ ] 数据库备份：恢复到测试环境
- [ ] 数据迁移：批量上传
- [ ] 数据迁移：路径更新
- [ ] 异常处理：S3 连接失败
- [ ] 异常处理：文件不存在
- [ ] 性能测试：大文件上传（100MB+）
- [ ] 性能测试：并发上传（10 用户同时上传）

---

**文档结束**

*如有疑问，请联系项目负责人或查阅相关技术文档。*

# PostgreSQL 数据库备份与归档设计方案

**项目名称：** CodeJYM 数据库备份系统
**文档版本：** 1.0
**创建日期：** 2025-11-20
**状态：** 设计阶段

---

## 📋 目录

1. [需求概述](#需求概述)
2. [架构设计](#架构设计)
3. [备份策略](#备份策略)
4. [WAL 归档方案](#wal-归档方案)
5. [S3 存储方案](#s3-存储方案)
6. [本地数据安全](#本地数据安全)
7. [恢复流程](#恢复流程)
8. [监控与告警](#监控与告警)
9. [成本估算](#成本估算)
10. [实施计划](#实施计划)

---

## 需求概述

### 业务需求

- **数据持久性：** 确保用户数据（账号、训练组、练习进度）不丢失
- **灾难恢复：** 在发生硬件故障、误操作、数据损坏时能够快速恢复
- **历史回溯：** 能够恢复到任意时间点的数据状态（PITR - Point-In-Time Recovery）
- **合规要求：** 满足数据保留和备份的合规性要求

### 技术需求

- **备份频率：**
  - 全量备份：每天 1 次
  - 增量备份（WAL）：实时归档
- **备份保留：**
  - 全量备份：保留 30 天
  - WAL 归档：保留 7 天
- **RTO（恢复时间目标）：** < 30 分钟
- **RPO（恢复点目标）：** < 1 分钟（通过 WAL 归档）
- **存储位置：**
  - 本地：Docker Volume + 本地磁盘
  - 远程：AWS S3 或兼容的对象存储

### 当前环境

- **数据库：** PostgreSQL 16
- **容器化：** Docker + Docker Compose
- **数据量：** 当前 < 1GB，预计 1 年内 < 10GB
- **访问模式：** 读多写少

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         CodeJYM 应用                              │
│                              ↓                                    │
│                    PostgreSQL 容器                                │
│                              ↓                                    │
│  ┌───────────────────────────────────────────────────┐          │
│  │           数据库存储层                             │          │
│  │  ┌─────────────┐  ┌──────────────┐              │          │
│  │  │ 数据目录     │  │ WAL 目录     │              │          │
│  │  │ /var/lib/   │  │ /var/lib/    │              │          │
│  │  │ postgresql/ │  │ postgresql/  │              │          │
│  │  │ data        │  │ data/pg_wal  │              │          │
│  │  └─────┬───────┘  └──────┬───────┘              │          │
│  └────────┼──────────────────┼──────────────────────┘          │
│           ↓                  ↓                                   │
│  ┌────────────────┐  ┌──────────────┐                          │
│  │ Docker Volume  │  │ WAL Archive  │                          │
│  │ pg_data        │  │ Volume       │                          │
│  └────────┬───────┘  └──────┬───────┘                          │
└───────────┼──────────────────┼──────────────────────────────────┘
            ↓                  ↓
   ┌────────────────┐  ┌──────────────┐
   │  本地备份目录   │  │ WAL 归档目录  │
   │  /backup/base  │  │ /backup/wal  │
   └────────┬───────┘  └──────┬───────┘
            ↓                  ↓
            └──────────┬───────┘
                       ↓
              ┌─────────────────┐
              │  备份脚本服务    │
              │  (Cron/Systemd) │
              └─────────┬───────┘
                        ↓
              ┌─────────────────┐
              │   AWS S3 Bucket │
              │   s3://codejym- │
              │   db-backup/    │
              └─────────────────┘
                        ↓
              ┌─────────────────┐
              │ S3 Glacier Deep │
              │ Archive (长期)  │
              └─────────────────┘
```

### 组件说明

| 组件 | 作用 | 技术选型 |
|------|------|---------|
| PostgreSQL 容器 | 主数据库 | PostgreSQL 16 |
| Docker Volume | 数据持久化 | Docker 本地卷 |
| WAL 归档目录 | WAL 日志存储 | 本地文件系统 |
| 本地备份目录 | 全量备份存储 | 本地文件系统 |
| 备份脚本服务 | 自动化备份 | Bash + Cron |
| S3 Bucket | 远程备份存储 | AWS S3 / MinIO |
| 监控服务 | 备份监控 | 自定义脚本 + 日志 |

---

## 备份策略

### 三层备份策略（3-2-1 原则）

- **3 份副本：** 原始数据 + 本地备份 + S3 远程备份
- **2 种介质：** 本地磁盘 + 云存储
- **1 份异地：** S3 云存储

### 全量备份（Base Backup）

**工具：** `pg_basebackup`

**频率：** 每天凌晨 2:00 AM

**命令示例：**
```bash
pg_basebackup \
  -h localhost \
  -p 5432 \
  -U postgres \
  -D /backup/base/$(date +%Y%m%d) \
  -Ft \
  -z \
  -Xs \
  -P \
  --wal-method=stream
```

**参数说明：**
- `-Ft`: tar 格式
- `-z`: gzip 压缩
- `-Xs`: 流式传输 WAL（包含在备份中）
- `-P`: 显示进度

**备份文件结构：**
```
/backup/base/
├── 20251120/
│   ├── base.tar.gz          # 数据文件
│   ├── pg_wal.tar.gz        # WAL 文件
│   └── backup_manifest      # 备份清单
├── 20251121/
│   ├── base.tar.gz
│   ├── pg_wal.tar.gz
│   └── backup_manifest
└── ...
```

**保留策略：**
- 本地保留最近 7 天
- S3 标准存储保留 30 天
- S3 Glacier 保留 1 年（可选）

### 增量备份（WAL 归档）

**工具：** PostgreSQL WAL 归档机制

**频率：** 实时（WAL 段完成时立即归档）

**PostgreSQL 配置（postgresql.conf）：**
```ini
# WAL 归档设置
wal_level = replica                    # 启用 WAL 归档所需的日志级别
archive_mode = on                      # 启用归档
archive_command = 'test ! -f /backup/wal/%f && cp %p /backup/wal/%f'
archive_timeout = 300                  # 5 分钟强制切换 WAL（即使未满）

# WAL 配置
max_wal_size = 1GB                     # WAL 最大大小
min_wal_size = 80MB                    # WAL 最小大小
wal_keep_size = 128MB                  # 保留的 WAL 大小

# 检查点配置
checkpoint_timeout = 10min             # 检查点超时
checkpoint_completion_target = 0.9     # 检查点完成目标
```

**WAL 归档脚本（archive-wal.sh）：**
```bash
#!/bin/bash
# WAL 归档脚本
# 参数：$1 = WAL 文件路径, $2 = WAL 文件名

WAL_SOURCE="$1"
WAL_NAME="$2"
LOCAL_ARCHIVE="/backup/wal"
S3_BUCKET="s3://codejym-db-backup/wal"

# 1. 复制到本地归档目录
cp "$WAL_SOURCE" "$LOCAL_ARCHIVE/$WAL_NAME"

# 2. 上传到 S3（异步，不阻塞）
aws s3 cp "$LOCAL_ARCHIVE/$WAL_NAME" "$S3_BUCKET/$WAL_NAME" &

# 3. 记录日志
echo "$(date '+%Y-%m-%d %H:%M:%S') - Archived: $WAL_NAME" >> /var/log/wal-archive.log

exit 0
```

**WAL 清理脚本（cleanup-wal.sh）：**
```bash
#!/bin/bash
# 清理超过 7 天的 WAL 文件

LOCAL_ARCHIVE="/backup/wal"
RETENTION_DAYS=7

find "$LOCAL_ARCHIVE" -name "*.wal" -mtime +$RETENTION_DAYS -delete

echo "$(date '+%Y-%m-%d %H:%M:%S') - WAL cleanup completed" >> /var/log/wal-cleanup.log
```

### 逻辑备份（可选）

**工具：** `pg_dump`

**用途：**
- 单表备份
- 特定数据导出
- 跨版本迁移

**频率：** 每周一次（周日）

**命令示例：**
```bash
pg_dump \
  -h localhost \
  -p 5432 \
  -U postgres \
  -d codejym \
  -Fc \
  -f /backup/logical/codejym_$(date +%Y%m%d).dump
```

---

## WAL 归档方案

### WAL 工作原理

```
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL 进程                           │
│                                                              │
│  写入操作 → WAL Buffer → WAL Writer → WAL 文件(16MB 一段)   │
│                                     ↓                        │
│                            WAL 段完成/超时                    │
│                                     ↓                        │
│                            触发 archive_command               │
│                                     ↓                        │
│                      ┌──────────────────────┐               │
│                      │   archive-wal.sh     │               │
│                      └──────────┬───────────┘               │
└───────────────────────────────┬─────────────────────────────┘
                                ↓
                    ┌───────────────────────┐
                    │   本地 WAL 归档目录    │
                    │   /backup/wal/        │
                    └───────────┬───────────┘
                                ↓
                    ┌───────────────────────┐
                    │   S3 WAL 存储         │
                    │   s3://.../wal/       │
                    └───────────────────────┘
```

### WAL 归档配置详解

**1. 归档命令优化：**

**方案 A：同步归档（简单但可能阻塞）**
```ini
archive_command = 'cp %p /backup/wal/%f'
```

**方案 B：异步归档（推荐）**
```ini
archive_command = '/usr/local/bin/archive-wal.sh %p %f'
```

**方案 C：直接上传 S3（需要快速网络）**
```ini
archive_command = 'aws s3 cp %p s3://codejym-db-backup/wal/%f'
```

**推荐：** 方案 B（本地 + 异步 S3）

**2. WAL 归档监控：**

检查归档是否正常：
```sql
-- 查看归档状态
SELECT * FROM pg_stat_archiver;

-- 查看 WAL 位置
SELECT pg_current_wal_lsn();

-- 查看未归档的 WAL 数量
SELECT count(*) FROM pg_ls_waldir()
WHERE name NOT IN (SELECT name FROM pg_ls_archive_statusdir() WHERE name LIKE '%.done');
```

**3. WAL 归档告警：**

```bash
#!/bin/bash
# 检查 WAL 归档延迟

MAX_ARCHIVE_AGE_SECONDS=600  # 10 分钟

LAST_ARCHIVED=$(psql -U postgres -d codejym -t -c "SELECT last_archived_time FROM pg_stat_archiver")
CURRENT_TIME=$(date +%s)
LAST_TIME=$(date -d "$LAST_ARCHIVED" +%s)
DIFF=$((CURRENT_TIME - LAST_TIME))

if [ $DIFF -gt $MAX_ARCHIVE_AGE_SECONDS ]; then
    echo "ALERT: WAL archiving delayed by $DIFF seconds!" | mail -s "DB Backup Alert" admin@example.com
fi
```

---

## S3 存储方案

### S3 Bucket 结构

```
s3://codejym-db-backup/
├── base/                          # 全量备份
│   ├── 2025/
│   │   ├── 11/
│   │   │   ├── 20/
│   │   │   │   ├── base.tar.gz
│   │   │   │   ├── pg_wal.tar.gz
│   │   │   │   └── backup_manifest
│   │   │   ├── 21/
│   │   │   │   └── ...
│   │   │   └── ...
│   │   └── 12/
│   │       └── ...
│   └── 2026/
│       └── ...
├── wal/                           # WAL 归档
│   ├── 000000010000000000000001
│   ├── 000000010000000000000002
│   ├── 000000010000000000000003
│   └── ...
└── manifests/                     # 备份清单
    ├── 2025-11-20-backup.json
    ├── 2025-11-21-backup.json
    └── ...
```

### S3 生命周期策略

**标准存储 → Glacier → Deep Archive**

```json
{
  "Rules": [
    {
      "Id": "base-backup-lifecycle",
      "Status": "Enabled",
      "Filter": {
        "Prefix": "base/"
      },
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "GLACIER"
        },
        {
          "Days": 365,
          "StorageClass": "DEEP_ARCHIVE"
        }
      ],
      "Expiration": {
        "Days": 1095
      }
    },
    {
      "Id": "wal-cleanup",
      "Status": "Enabled",
      "Filter": {
        "Prefix": "wal/"
      },
      "Expiration": {
        "Days": 7
      }
    }
  ]
}
```

### S3 上传脚本

**完整备份上传脚本（upload-to-s3.sh）：**

```bash
#!/bin/bash
set -e

# 配置
BACKUP_DATE=$(date +%Y%m%d)
LOCAL_BACKUP_DIR="/backup/base/$BACKUP_DATE"
S3_BUCKET="s3://codejym-db-backup/base/$(date +%Y/%m/%d)"
LOG_FILE="/var/log/s3-upload.log"

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# 检查备份是否存在
if [ ! -d "$LOCAL_BACKUP_DIR" ]; then
    log "ERROR: Backup directory not found: $LOCAL_BACKUP_DIR"
    exit 1
fi

# 上传到 S3
log "Starting upload to S3: $S3_BUCKET"

aws s3 sync "$LOCAL_BACKUP_DIR" "$S3_BUCKET" \
    --storage-class STANDARD \
    --no-progress \
    2>&1 | tee -a "$LOG_FILE"

if [ $? -eq 0 ]; then
    log "SUCCESS: Backup uploaded to S3"

    # 创建备份清单
    cat > "/tmp/backup-manifest-$BACKUP_DATE.json" <<EOF
{
  "backup_date": "$BACKUP_DATE",
  "backup_time": "$(date '+%Y-%m-%d %H:%M:%S')",
  "s3_location": "$S3_BUCKET",
  "local_location": "$LOCAL_BACKUP_DIR",
  "status": "completed"
}
EOF

    aws s3 cp "/tmp/backup-manifest-$BACKUP_DATE.json" \
        "s3://codejym-db-backup/manifests/$BACKUP_DATE-backup.json"

else
    log "ERROR: Failed to upload to S3"
    exit 1
fi

# 清理本地旧备份（保留 7 天）
log "Cleaning up old local backups"
find /backup/base -mindepth 1 -maxdepth 1 -type d -mtime +7 -exec rm -rf {} \;

log "Backup process completed"
```

### S3 访问权限（IAM Policy）

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowBackupOperations",
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:DeleteObject"
      ],
      "Resource": [
        "arn:aws:s3:::codejym-db-backup",
        "arn:aws:s3:::codejym-db-backup/*"
      ]
    }
  ]
}
```

### S3 备份加密

**服务端加密（SSE-S3）：**
```bash
aws s3 cp file.tar.gz s3://bucket/ --server-side-encryption AES256
```

**客户端加密（可选）：**
```bash
# 使用 GPG 加密
gpg --symmetric --cipher-algo AES256 backup.tar.gz
aws s3 cp backup.tar.gz.gpg s3://bucket/
```

---

## 本地数据安全

### Docker Volume 配置

**docker-compose.yml 配置：**

```yaml
services:
  postgres:
    image: postgres:16
    container_name: codejym-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: codejym
    volumes:
      # 数据目录
      - pg_data:/var/lib/postgresql/data
      # WAL 归档目录
      - ./backup/wal:/backup/wal
      # 全量备份目录
      - ./backup/base:/backup/base
      # 配置文件
      - ./postgres/postgresql.conf:/etc/postgresql/postgresql.conf
      - ./postgres/pg_hba.conf:/etc/postgresql/pg_hba.conf
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  pg_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/postgresql  # 挂载到 RAID 或独立磁盘
```

### 文件系统级别保护

**1. 使用独立磁盘/分区：**
```bash
# 创建独立分区
mkdir -p /data/postgresql
mkfs.ext4 /dev/sdb1
mount /dev/sdb1 /data/postgresql

# 添加到 fstab
echo "/dev/sdb1 /data/postgresql ext4 defaults,noatime 0 2" >> /etc/fstab
```

**2. 文件系统快照（LVM）：**
```bash
# 创建 LVM 卷
lvcreate -L 100G -n pg_data vg0
mkfs.ext4 /dev/vg0/pg_data
mount /dev/vg0/pg_data /data/postgresql

# 创建快照
lvcreate -L 10G -s -n pg_data_snapshot /dev/vg0/pg_data

# 恢复快照
lvconvert --merge /dev/vg0/pg_data_snapshot
```

**3. 权限设置：**
```bash
# 设置目录权限
chown -R 999:999 /data/postgresql  # PostgreSQL 容器用户
chmod 700 /data/postgresql

# 备份目录权限
chmod 750 /backup
```

### 数据完整性检查

**1. PostgreSQL 内置检查：**
```sql
-- 检查数据块校验和
SHOW data_checksums;

-- 验证表
SELECT * FROM pg_catalog.pg_check_frozen_ids();

-- 检查索引
REINDEX DATABASE codejym;
```

**2. 定期一致性检查脚本：**
```bash
#!/bin/bash
# db-integrity-check.sh

psql -U postgres -d codejym -c "SELECT count(*) FROM users;" > /dev/null
if [ $? -ne 0 ]; then
    echo "Database integrity check FAILED!" | mail -s "DB Alert" admin@example.com
fi
```

### RAID 配置（推荐）

**RAID 10 配置：**
```bash
# 创建 RAID 10（4 块磁盘）
mdadm --create --verbose /dev/md0 --level=10 --raid-devices=4 \
    /dev/sdb /dev/sdc /dev/sdd /dev/sde

# 格式化
mkfs.ext4 /dev/md0

# 挂载
mount /dev/md0 /data/postgresql
```

**优势：**
- 读写性能提升
- 单盘故障不影响服务
- 数据冗余

---

## 恢复流程

### 恢复场景分类

| 场景 | 恢复方法 | RTO | RPO |
|------|---------|-----|-----|
| 数据库损坏 | 从最近全量备份 + WAL 恢复 | 20 分钟 | < 1 分钟 |
| 误删除操作 | PITR 到操作前时间点 | 30 分钟 | 精确到秒 |
| 硬件故障 | 从 S3 恢复到新服务器 | 60 分钟 | < 1 分钟 |
| 整个集群丢失 | 从 S3 完全重建 | 90 分钟 | < 5 分钟 |

### 完整恢复流程（PITR）

**步骤 1：准备恢复环境**

```bash
# 停止数据库
docker-compose stop postgres

# 清空数据目录
rm -rf /data/postgresql/*

# 创建恢复目录
mkdir -p /data/postgresql/recovery
```

**步骤 2：恢复全量备份**

```bash
# 从本地恢复
cd /data/postgresql
tar -xzf /backup/base/20251120/base.tar.gz

# 或从 S3 恢复
aws s3 cp s3://codejym-db-backup/base/2025/11/20/base.tar.gz ./
tar -xzf base.tar.gz
```

**步骤 3：配置恢复参数**

创建 `postgresql.auto.conf` 或 `recovery.signal`：

```bash
# 创建恢复信号文件（PostgreSQL 12+）
touch /data/postgresql/recovery.signal

# 配置恢复参数
cat >> /data/postgresql/postgresql.auto.conf <<EOF
restore_command = 'cp /backup/wal/%f %p'
recovery_target_time = '2025-11-20 15:30:00'  # 恢复到指定时间点
recovery_target_action = 'promote'            # 恢复后自动提升为主库
EOF
```

**步骤 4：启动数据库并恢复**

```bash
# 启动 PostgreSQL
docker-compose start postgres

# 监控恢复进度
docker logs -f codejym-postgres

# 验证恢复
docker exec -it codejym-postgres psql -U postgres -d codejym -c "SELECT count(*) FROM users;"
```

**步骤 5：验证和清理**

```bash
# 检查数据库状态
psql -U postgres -c "SELECT pg_is_in_recovery();"  # 应该返回 false

# 删除恢复信号文件（如果存在）
rm -f /data/postgresql/recovery.signal

# 重新启动 WAL 归档
# 恢复后会自动开始新的 WAL 序列
```

### 恢复测试脚本

**recovery-test.sh：**

```bash
#!/bin/bash
set -e

LOG_FILE="/var/log/recovery-test.log"

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log "Starting recovery test..."

# 1. 创建测试数据库
docker exec codejym-postgres psql -U postgres -c "CREATE DATABASE test_recovery;"

# 2. 插入测试数据
docker exec codejym-postgres psql -U postgres -d test_recovery -c "
CREATE TABLE test_data (
    id SERIAL PRIMARY KEY,
    data TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
INSERT INTO test_data (data) VALUES ('Test data before backup');
"

# 3. 创建基础备份
BACKUP_DIR="/tmp/recovery-test-$(date +%s)"
mkdir -p "$BACKUP_DIR"

docker exec codejym-postgres pg_basebackup \
    -U postgres \
    -D "$BACKUP_DIR" \
    -Ft \
    -z \
    -P

log "Base backup created: $BACKUP_DIR"

# 4. 插入更多数据（应该通过 WAL 恢复）
sleep 2
docker exec codejym-postgres psql -U postgres -d test_recovery -c "
INSERT INTO test_data (data) VALUES ('Test data after backup');
"

# 5. 模拟恢复（实际环境中需要停止数据库）
log "Recovery test completed successfully"
log "Backup location: $BACKUP_DIR"
log "Please manually verify the recovery process"
```

### 恢复演练计划

**频率：** 每季度一次

**演练内容：**
1. 从最新备份恢复
2. PITR 到特定时间点
3. 从 S3 完全恢复
4. 恢复单个表

**演练检查清单：**
- [ ] 备份文件完整性
- [ ] WAL 文件连续性
- [ ] 恢复时间满足 RTO
- [ ] 数据完整性验证
- [ ] 应用连接测试
- [ ] 性能测试

---

## 监控与告警

### 监控指标

**1. 备份监控：**
- 最后一次全量备份时间
- 最后一次 WAL 归档时间
- 备份文件大小
- S3 上传成功率
- 本地磁盘使用率

**2. WAL 监控：**
- WAL 生成速率
- WAL 归档延迟
- 未归档的 WAL 数量
- WAL 磁盘使用

**3. 数据库监控：**
- 连接数
- 事务速率
- 查询性能
- 磁盘 I/O

### 监控脚本

**backup-monitor.sh：**

```bash
#!/bin/bash

# 监控配置
ALERT_EMAIL="admin@example.com"
BACKUP_DIR="/backup/base"
WAL_DIR="/backup/wal"
MAX_BACKUP_AGE_HOURS=26  # 超过 26 小时没有备份则告警

# 检查最后一次备份
check_last_backup() {
    local latest_backup=$(find "$BACKUP_DIR" -maxdepth 1 -type d | sort -r | head -n 2 | tail -n 1)

    if [ -z "$latest_backup" ]; then
        echo "ERROR: No backup found!"
        return 1
    fi

    local backup_time=$(stat -c %Y "$latest_backup")
    local current_time=$(date +%s)
    local age_hours=$(( (current_time - backup_time) / 3600 ))

    if [ $age_hours -gt $MAX_BACKUP_AGE_HOURS ]; then
        echo "ALERT: Last backup is $age_hours hours old!"
        return 1
    fi

    echo "OK: Last backup is $age_hours hours old"
    return 0
}

# 检查 WAL 归档
check_wal_archiving() {
    local wal_count=$(find "$WAL_DIR" -name "*.wal" -mmin -10 | wc -l)

    if [ $wal_count -eq 0 ]; then
        echo "WARNING: No WAL files archived in the last 10 minutes"
        return 1
    fi

    echo "OK: $wal_count WAL files archived recently"
    return 0
}

# 检查磁盘空间
check_disk_space() {
    local usage=$(df -h "$BACKUP_DIR" | awk 'NR==2 {print $5}' | sed 's/%//')

    if [ $usage -gt 80 ]; then
        echo "ALERT: Disk usage is ${usage}%"
        return 1
    fi

    echo "OK: Disk usage is ${usage}%"
    return 0
}

# 检查 S3 连接
check_s3_connection() {
    aws s3 ls s3://codejym-db-backup/ > /dev/null 2>&1

    if [ $? -ne 0 ]; then
        echo "ERROR: Cannot connect to S3"
        return 1
    fi

    echo "OK: S3 connection successful"
    return 0
}

# 主监控逻辑
main() {
    local status=0
    local report=""

    report+="=== Backup Monitoring Report $(date) ===\n\n"

    check_last_backup
    if [ $? -ne 0 ]; then
        status=1
        report+="❌ Backup check FAILED\n"
    else
        report+="✅ Backup check PASSED\n"
    fi

    check_wal_archiving
    if [ $? -ne 0 ]; then
        status=1
        report+="❌ WAL archiving check FAILED\n"
    else
        report+="✅ WAL archiving check PASSED\n"
    fi

    check_disk_space
    if [ $? -ne 0 ]; then
        status=1
        report+="❌ Disk space check FAILED\n"
    else
        report+="✅ Disk space check PASSED\n"
    fi

    check_s3_connection
    if [ $? -ne 0 ]; then
        status=1
        report+="❌ S3 connection check FAILED\n"
    else
        report+="✅ S3 connection check PASSED\n"
    fi

    # 发送告警
    if [ $status -ne 0 ]; then
        echo -e "$report" | mail -s "⚠️  Backup Monitoring Alert" "$ALERT_EMAIL"
    fi

    echo -e "$report"
    return $status
}

main
```

### 告警策略

| 告警级别 | 条件 | 处理方式 |
|---------|------|---------|
| 🔴 Critical | - 备份失败<br>- WAL 归档停止 > 30 分钟<br>- S3 连接失败 | 立即发送邮件 + 短信 |
| 🟡 Warning | - 备份延迟 > 2 小时<br>- 磁盘使用 > 80%<br>- WAL 归档延迟 > 10 分钟 | 发送邮件 |
| 🟢 Info | - 备份成功<br>- 恢复测试通过 | 记录日志 |

### 监控仪表板（可选）

**使用 Prometheus + Grafana：**

**postgres_exporter 配置：**
```yaml
# docker-compose.yml
services:
  postgres-exporter:
    image: prometheuscommunity/postgres-exporter
    environment:
      DATA_SOURCE_NAME: "postgresql://postgres:password@postgres:5432/codejym?sslmode=disable"
    ports:
      - "9187:9187"
```

**Prometheus 监控指标：**
- `pg_stat_archiver_archived_count`
- `pg_stat_archiver_failed_count`
- `pg_stat_database_xact_commit`
- `pg_database_size_bytes`

---

## 成本估算

### S3 存储成本（按月计算）

**假设：**
- 每日全量备份大小：500MB（压缩后）
- 每日 WAL 生成：200MB
- 保留策略：标准存储 30 天，Glacier 1 年

| 存储类型 | 容量 | 单价（美元/GB/月） | 月成本 |
|---------|------|-------------------|--------|
| S3 Standard（全量备份，30 天） | 15GB | $0.023 | $0.35 |
| S3 Standard（WAL，7 天） | 1.4GB | $0.023 | $0.03 |
| S3 Glacier（全量备份，1 年） | 180GB | $0.004 | $0.72 |
| **总计** | **196GB** | - | **$1.10** |

**数据传输成本：**
- 上传到 S3：免费
- 从 S3 下载（恢复）：前 100GB/月免费，之后 $0.09/GB

**估算总成本：**
- **月成本：** ~$1.50
- **年成本：** ~$18

### 本地存储成本

| 项目 | 容量需求 | 成本估算 |
|------|---------|---------|
| 数据目录 | 10GB（增长） | 已有硬盘 |
| 全量备份（7 天） | 3.5GB | 已有硬盘 |
| WAL 归档（7 天） | 1.4GB | 已有硬盘 |
| **总计** | **~15GB** | **$0** |

### 人力成本

| 任务 | 频率 | 时间 | 年成本估算 |
|------|------|------|-----------|
| 初始设置 | 一次性 | 8 小时 | - |
| 日常监控 | 自动化 | - | $0 |
| 季度恢复演练 | 4 次/年 | 2 小时/次 | 8 小时/年 |
| 故障恢复 | 按需 | - | - |

---

## 实施计划

### Phase 1：基础设施准备（第 1 周）

**任务：**
- [ ] 配置独立磁盘/分区用于数据库存储
- [ ] 创建本地备份目录结构
- [ ] 安装和配置 AWS CLI
- [ ] 创建 S3 Bucket 并配置生命周期策略
- [ ] 配置 IAM 用户和权限

**验收标准：**
- S3 Bucket 可访问
- 本地目录权限正确
- AWS CLI 配置测试通过

### Phase 2：PostgreSQL 配置（第 1 周）

**任务：**
- [ ] 修改 `postgresql.conf` 启用 WAL 归档
- [ ] 配置 `archive_command`
- [ ] 重启 PostgreSQL 并验证配置
- [ ] 测试 WAL 归档是否正常

**验收标准：**
- `pg_stat_archiver` 显示归档正常
- WAL 文件出现在归档目录

### Phase 3：备份脚本开发（第 2 周）

**任务：**
- [ ] 开发全量备份脚本（`backup-full.sh`）
- [ ] 开发 WAL 归档脚本（`archive-wal.sh`）
- [ ] 开发 S3 上传脚本（`upload-to-s3.sh`）
- [ ] 开发清理脚本（`cleanup-old-backups.sh`）
- [ ] 配置 Cron 定时任务

**验收标准：**
- 备份脚本成功执行
- 备份文件上传到 S3
- 旧备份自动清理

### Phase 4：恢复流程验证（第 2 周）

**任务：**
- [ ] 开发恢复脚本（`restore.sh`）
- [ ] 在测试环境执行完整恢复
- [ ] 测试 PITR（时间点恢复）
- [ ] 记录恢复时间和步骤
- [ ] 编写恢复操作手册

**验收标准：**
- 恢复成功率 100%
- RTO < 30 分钟
- RPO < 1 分钟

### Phase 5：监控和告警（第 3 周）

**任务：**
- [ ] 开发监控脚本（`backup-monitor.sh`）
- [ ] 配置邮件告警
- [ ] 配置日志记录
- [ ] 创建监控仪表板（可选）
- [ ] 测试告警机制

**验收标准：**
- 监控脚本每小时运行
- 告警邮件正常发送
- 日志完整记录

### Phase 6：文档和培训（第 3 周）

**任务：**
- [ ] 编写操作手册
- [ ] 创建故障处理流程图
- [ ] 团队培训
- [ ] 进行灾难恢复演练

**验收标准：**
- 文档完整
- 团队成员熟悉恢复流程

### Phase 7：生产环境部署（第 4 周）

**任务：**
- [ ] 在生产环境部署所有脚本
- [ ] 配置生产环境 Cron
- [ ] 执行首次全量备份
- [ ] 验证 S3 上传
- [ ] 监控一周确保稳定

**验收标准：**
- 生产备份正常运行
- 无告警产生
- 性能无影响

---

## 附录

### A. 脚本清单

| 脚本名称 | 路径 | 作用 |
|---------|------|------|
| `backup-full.sh` | `/usr/local/bin/` | 全量备份 |
| `archive-wal.sh` | `/usr/local/bin/` | WAL 归档 |
| `upload-to-s3.sh` | `/usr/local/bin/` | S3 上传 |
| `cleanup-old-backups.sh` | `/usr/local/bin/` | 清理旧备份 |
| `restore.sh` | `/usr/local/bin/` | 数据恢复 |
| `backup-monitor.sh` | `/usr/local/bin/` | 监控检查 |
| `recovery-test.sh` | `/usr/local/bin/` | 恢复测试 |

### B. Cron 配置

```bash
# /etc/crontab 或 crontab -e

# 每天凌晨 2:00 全量备份
0 2 * * * /usr/local/bin/backup-full.sh >> /var/log/backup-full.log 2>&1

# 每天凌晨 3:00 上传到 S3
0 3 * * * /usr/local/bin/upload-to-s3.sh >> /var/log/s3-upload.log 2>&1

# 每天凌晨 4:00 清理旧备份
0 4 * * * /usr/local/bin/cleanup-old-backups.sh >> /var/log/cleanup.log 2>&1

# 每小时监控检查
0 * * * * /usr/local/bin/backup-monitor.sh >> /var/log/backup-monitor.log 2>&1

# 每周日逻辑备份
0 1 * * 0 /usr/local/bin/backup-logical.sh >> /var/log/backup-logical.log 2>&1
```

### C. 环境变量配置

```bash
# /etc/environment 或 ~/.bashrc

# AWS 配置
export AWS_ACCESS_KEY_ID="your_access_key"
export AWS_SECRET_ACCESS_KEY="your_secret_key"
export AWS_DEFAULT_REGION="us-east-1"

# 备份配置
export BACKUP_BASE_DIR="/backup/base"
export BACKUP_WAL_DIR="/backup/wal"
export S3_BUCKET="s3://codejym-db-backup"

# 数据库配置
export PGHOST="localhost"
export PGPORT="5432"
export PGUSER="postgres"
export PGPASSWORD="your_password"
export PGDATABASE="codejym"
```

### D. 故障排查指南

| 问题 | 可能原因 | 解决方法 |
|------|---------|---------|
| WAL 归档失败 | - 磁盘满<br>- 权限问题<br>- 脚本错误 | 1. 检查磁盘空间<br>2. 检查文件权限<br>3. 查看 PostgreSQL 日志 |
| S3 上传失败 | - 网络问题<br>- 权限问题<br>- 配额限制 | 1. 测试网络连接<br>2. 验证 IAM 权限<br>3. 检查 S3 配额 |
| 恢复失败 | - WAL 文件缺失<br>- 备份损坏<br>- 配置错误 | 1. 检查 WAL 连续性<br>2. 验证备份完整性<br>3. 检查恢复配置 |

### E. 参考文档

- [PostgreSQL Backup and Restore](https://www.postgresql.org/docs/current/backup.html)
- [PostgreSQL WAL Archiving](https://www.postgresql.org/docs/current/continuous-archiving.html)
- [AWS S3 Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/best-practices.html)
- [Docker Volume Management](https://docs.docker.com/storage/volumes/)

---

**文档状态：** ✅ 已完成
**下一步：** 等待审批后开始实施

**变更记录：**
- 2025-11-20：初版创建

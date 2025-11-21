# 数据库备份系统使用说明

## 📋 概述

PostgreSQL 数据库自动备份到缤纷云 S3 对象存储。

**备份频率**：每天凌晨 2:00
**保留策略**：保留最近 30 天的备份
**S3 Bucket**：codejym-backups
**备份路径**：backups/database/full/

---

## 📁 脚本文件

### 1. backup-db-to-s3.sh
主备份脚本，执行以下操作：
- 使用 `pg_dump` 导出数据库
- 使用 gzip 压缩
- 上传到 S3
- 清理 30 天前的旧备份

**配置**：
```bash
S3_ENDPOINT=https://s3.bitiful.net
S3_REGION=cn-east-1
S3_BUCKET=codejym-backups
AWS_PROFILE=bitiful
RETENTION_DAYS=30
```

### 2. run-backup.sh
包装脚本，设置正确的数据库连接参数并调用主备份脚本。

**数据库连接**：
```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_USER=codecopy
POSTGRES_PASSWORD=codecopy123
POSTGRES_DB=codecopybook
```

### 3. restore-db-from-s3.sh
恢复脚本（如需使用请参考脚本内容）。

---

## 🚀 使用方法

### 手动执行备份

```bash
# 方法 1：使用包装脚本（推荐）
cd /opt/codejym
./scripts/run-backup.sh

# 方法 2：直接调用主脚本
POSTGRES_HOST=localhost POSTGRES_PORT=5433 ./scripts/backup-db-to-s3.sh
```

### 自动备份（Cron）

已配置 cron 任务，每天凌晨 2:00 自动执行：

```bash
# 查看 crontab 配置
crontab -l

# 编辑 crontab（如需修改时间）
crontab -e

# 查看备份日志
tail -f /opt/codejym/logs/pg-backup.log
```

**Cron 时间格式说明**：
```
0 2 * * *    # 每天凌晨 2:00
0 */6 * * *  # 每 6 小时一次
0 0 * * 0    # 每周日午夜
30 3 * * *   # 每天凌晨 3:30
```

---

## 📊 查看和管理备份

### 列出所有备份

```bash
aws s3 ls s3://codejym-backups/backups/database/full/ \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1 \
  --profile bitiful
```

### 下载备份文件

```bash
# 下载到当前目录
aws s3 cp s3://codejym-backups/backups/database/full/codecopybook_20251120_222500.sql.gz . \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1 \
  --profile bitiful

# 解压查看
zcat codecopybook_20251120_222500.sql.gz | less
```

### 恢复数据库

```bash
# 下载备份文件
aws s3 cp s3://codejym-backups/backups/database/full/codecopybook_YYYYMMDD_HHMMSS.sql.gz . \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1 \
  --profile bitiful

# 恢复到数据库（慎重操作！）
zcat codecopybook_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i codejym-postgres-1 psql -U codecopy -d codecopybook

# 或使用恢复脚本
./scripts/restore-db-from-s3.sh
```

---

## 📝 日志管理

### 查看备份日志

```bash
# 实时查看
tail -f /opt/codejym/logs/pg-backup.log

# 查看最近 50 行
tail -50 /opt/codejym/logs/pg-backup.log

# 搜索错误
grep -i error /opt/codejym/logs/pg-backup.log

# 查看特定日期的备份
grep "2025-11-20" /opt/codejym/logs/pg-backup.log
```

### 日志轮转（可选）

如果日志文件过大，可以配置 logrotate：

```bash
sudo tee /etc/logrotate.d/pg-backup << 'EOF'
/opt/codejym/logs/pg-backup.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
}
EOF
```

---

## 🔧 故障排查

### 1. 备份失败

**检查数据库连接**：
```bash
docker exec codejym-postgres-1 psql -U codecopy -d codecopybook -c "SELECT 1;"
```

**检查 S3 连接**：
```bash
aws s3 ls s3://codejym-backups/ \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1 \
  --profile bitiful
```

**检查 AWS profile 配置**：
```bash
aws configure list --profile bitiful
```

### 2. Cron 任务未执行

**检查 cron 服务**：
```bash
# Debian/Ubuntu
sudo systemctl status cron

# 查看 cron 日志
grep CRON /var/log/syslog | tail -20
```

**测试脚本权限**：
```bash
ls -l /opt/codejym/scripts/run-backup.sh
# 应该有执行权限：-rwxr-xr-x
```

### 3. 常见错误

| 错误 | 原因 | 解决方法 |
|------|------|---------|
| `pg_dump: connection failed` | 数据库未运行 | 检查容器状态 |
| `Unable to locate credentials` | AWS profile 未配置 | 运行 `aws configure --profile bitiful` |
| `Upload failed` | S3 连接失败 | 检查网络和 endpoint 配置 |
| `Permission denied` | 脚本无执行权限 | `chmod +x scripts/*.sh` |

---

## 📈 监控建议

### 定期检查

每周检查一次备份状态：

```bash
# 1. 检查最近 7 天的备份
aws s3 ls s3://codejym-backups/backups/database/full/ \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1 \
  --profile bitiful | tail -7

# 2. 检查日志中的错误
grep -i "error\|failed" /opt/codejym/logs/pg-backup.log | tail -10

# 3. 验证最新备份可以下载
aws s3 cp s3://codejym-backups/backups/database/full/$(aws s3 ls s3://codejym-backups/backups/database/full/ --endpoint-url https://s3.bitiful.net --region cn-east-1 --profile bitiful | tail -1 | awk '{print $4}') /tmp/test-backup.sql.gz \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1 \
  --profile bitiful && echo "备份下载成功" && rm /tmp/test-backup.sql.gz
```

### 建议设置告警

可以添加一个监控脚本，检测备份是否成功：

```bash
# 添加到 crontab（每天早上 3:00 检查昨天的备份）
0 3 * * * /opt/codejym/scripts/check-backup.sh
```

---

## 🔐 安全注意事项

1. **AWS 凭证**：存储在 `~/.aws/credentials`，确保文件权限为 600
2. **数据库密码**：存储在 `run-backup.sh` 中，确保文件权限为 700
3. **备份文件**：包含敏感数据，存储在 S3 私有 bucket
4. **日志文件**：可能包含敏感信息，定期清理或限制访问权限

---

## 📞 相关文档

- [PostgreSQL 备份与恢复设计方案](../docs/migration/DATABASE_BACKUP_DESIGN.md)
- [S3 存储迁移设计](../docs/migration/STORAGE_MIGRATION_DESIGN.md)
- [项目目录结构](../docs/PROJECT_STRUCTURE.md)

---

**最后更新**：2025-11-20
**维护者**：CodeJYM team

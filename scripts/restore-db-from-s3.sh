#!/bin/bash
# 数据库恢复脚本 - 从缤纷云 S3 恢复

set -euo pipefail

# ==================== 配置 ====================
RESTORE_DIR="/tmp/pg_restore"
BACKUP_FILE="${1:-latest}"  # 备份文件名或 "latest"

# 数据库连接信息
PGHOST="${POSTGRES_HOST:-postgres}"
PGPORT="${POSTGRES_PORT:-5432}"
PGDATABASE="${POSTGRES_DB:-codecopybook}"
PGUSER="${POSTGRES_USER:-codecopy}"
PGPASSWORD="${POSTGRES_PASSWORD:-codecopy123}"

# S3 配置
S3_ENDPOINT="${S3_ENDPOINT:-https://s3.bitiful.net}"
S3_REGION="${S3_REGION:-us-east-1}"
S3_BUCKET="${S3_BACKUP_BUCKET:-codejym-backups}"
S3_PREFIX="backups/database/full"

export PGPASSWORD

# ==================== 函数 ====================
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

error_exit() {
    log "ERROR: $*" >&2
    exit 1
}

# ==================== 主流程 ====================
log "Starting database restore..."

# 检查必要的环境变量
if [[ -z "${S3_ACCESS_KEY:-}" ]] || [[ -z "${S3_SECRET_KEY:-}" ]]; then
    error_exit "S3_ACCESS_KEY and S3_SECRET_KEY must be set"
fi

# 如果指定 latest，获取最新备份
if [[ "$BACKUP_FILE" == "latest" ]]; then
    log "Finding latest backup..."
    BACKUP_FILE=$(AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY}" \
                  AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY}" \
                  aws s3 ls "s3://${S3_BUCKET}/${S3_PREFIX}/" \
                      --endpoint-url "${S3_ENDPOINT}" \
                      --region "${S3_REGION}" \
                  | awk '{print $4}' \
                  | grep -E '^codecopybook_[0-9]{8}_' \
                  | sort -r \
                  | head -n 1)

    if [[ -z "$BACKUP_FILE" ]]; then
        error_exit "No backup found in S3"
    fi

    log "Latest backup: $BACKUP_FILE"
fi

# 创建恢复目录
mkdir -p "$RESTORE_DIR"

# 下载备份
log "Downloading backup from S3: s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}"
if ! AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY}" \
     AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY}" \
     aws s3 cp "s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}" \
         "${RESTORE_DIR}/${BACKUP_FILE}" \
         --endpoint-url "${S3_ENDPOINT}" \
         --region "${S3_REGION}"; then
    error_exit "S3 download failed"
fi

BACKUP_SIZE=$(du -h "${RESTORE_DIR}/${BACKUP_FILE}" | cut -f1)
log "Backup downloaded: ${BACKUP_FILE} (${BACKUP_SIZE})"

# 警告：恢复将覆盖现有数据
log "WARNING: This will DROP and recreate the database ${PGDATABASE}"
log "Press Ctrl+C within 5 seconds to cancel..."
sleep 5

# 解压并恢复
log "Restoring database..."

# 断开所有连接并重建数据库
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d postgres <<EOF
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '${PGDATABASE}' AND pid <> pg_backend_pid();

DROP DATABASE IF EXISTS ${PGDATABASE};
CREATE DATABASE ${PGDATABASE};
EOF

# 恢复数据
if ! gunzip -c "${RESTORE_DIR}/${BACKUP_FILE}" | \
     psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE"; then
    error_exit "Database restore failed"
fi

# 清理
rm -rf "$RESTORE_DIR"
log "Restore directory cleaned"

log "Database restored successfully!"
log "Database: ${PGDATABASE}"
log "From backup: ${BACKUP_FILE}"

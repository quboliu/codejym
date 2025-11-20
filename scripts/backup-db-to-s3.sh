#!/bin/bash
# 数据库全量备份脚本 - 备份到缤纷云 S3

set -euo pipefail

# ==================== 配置 ====================
BACKUP_DIR="/tmp/pg_backup"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="codecopybook_${TIMESTAMP}.sql.gz"
RETENTION_DAYS=30

# 数据库连接信息
PGHOST="${POSTGRES_HOST:-postgres}"
PGPORT="${POSTGRES_PORT:-5432}"
PGDATABASE="${POSTGRES_DB:-codecopybook}"
PGUSER="${POSTGRES_USER:-codecopy}"
PGPASSWORD="${POSTGRES_PASSWORD:-codecopy123}"

# S3 配置（从环境变量读取）
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
log "Starting database backup..."

# 检查必要的环境变量
if [[ -z "${S3_ACCESS_KEY:-}" ]] || [[ -z "${S3_SECRET_KEY:-}" ]]; then
    error_exit "S3_ACCESS_KEY and S3_SECRET_KEY must be set"
fi

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 执行备份并压缩
log "Creating backup: ${BACKUP_FILE}"
if ! pg_dump -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" \
    --format=plain \
    --no-owner \
    --no-acl \
    --verbose \
    2>&1 | gzip > "${BACKUP_DIR}/${BACKUP_FILE}"; then
    error_exit "pg_dump failed"
fi

BACKUP_SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_FILE}" | cut -f1)
log "Backup created: ${BACKUP_FILE} (${BACKUP_SIZE})"

# 上传到 S3
log "Uploading to S3: s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}"
if ! AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY}" \
     AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY}" \
     aws s3 cp "${BACKUP_DIR}/${BACKUP_FILE}" \
         "s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}" \
         --endpoint-url "${S3_ENDPOINT}" \
         --region "${S3_REGION}"; then
    error_exit "S3 upload failed"
fi

log "Upload completed successfully"

# 清理本地备份文件
rm -f "${BACKUP_DIR}/${BACKUP_FILE}"
log "Local backup file removed"

# 清理 S3 上的旧备份（保留最近 N 天）
log "Cleaning up old backups (retention: ${RETENTION_DAYS} days)..."
CUTOFF_DATE=$(date -d "${RETENTION_DAYS} days ago" +%Y%m%d)

AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY}" \
AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY}" \
aws s3 ls "s3://${S3_BUCKET}/${S3_PREFIX}/" \
    --endpoint-url "${S3_ENDPOINT}" \
    --region "${S3_REGION}" \
    | awk '{print $4}' \
    | grep -E '^codecopybook_[0-9]{8}_' \
    | while read -r file; do
        if [[ -z "$file" ]]; then
            continue
        fi

        # 提取日期部分（YYYYMMDD）
        file_date=$(echo "$file" | grep -oP '(?<=codecopybook_)\d{8}')

        if [[ -n "$file_date" ]] && [[ "$file_date" -lt "$CUTOFF_DATE" ]]; then
            log "Deleting old backup: $file (date: $file_date)"
            AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY}" \
            AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY}" \
            aws s3 rm "s3://${S3_BUCKET}/${S3_PREFIX}/${file}" \
                --endpoint-url "${S3_ENDPOINT}" \
                --region "${S3_REGION}" || true
        fi
    done

log "Backup completed successfully!"
log "Backup location: s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}"

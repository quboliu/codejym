#!/bin/bash
# 数据库备份包装脚本 - 用于 cron 或手动执行
# 设置正确的数据库连接参数

set -e

# 切换到脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 执行备份（连接到宿主机上的 postgres 容器）
POSTGRES_HOST=localhost \
POSTGRES_PORT=5433 \
POSTGRES_USER=codecopy \
POSTGRES_PASSWORD=codecopy123 \
POSTGRES_DB=codecopybook \
./backup-db-to-s3.sh "$@"

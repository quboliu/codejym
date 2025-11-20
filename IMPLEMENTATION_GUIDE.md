# CodeJYM 存储迁移实施指南

本文档提供分步操作说明，帮助你将 CodeJYM 项目从本地文件存储迁移到缤纷云 S3 对象存储。

---

## 📋 前置条件检查

在开始之前，请确保：

- [ ] 你已阅读 `STORAGE_MIGRATION_DESIGN.md`（完整设计文档）
- [ ] 项目当前运行正常，无未解决的错误
- [ ] 有权限访问服务器和数据库
- [ ] 了解基本的 Docker 和 Linux 命令

---

## 🚀 实施步骤

### Phase 1: 注册缤纷云并配置（预计 30 分钟）

#### 1.1 注册缤纷云账号

1. 访问缤纷云官网：https://www.bitiful.com/
2. 点击"注册"或"登录"
3. 完成账号注册和实名认证（获取免费额度）

#### 1.2 创建存储 Bucket

1. 登录缤纷云控制台：https://console.bitiful.com/
2. 进入"对象存储"或"S4"服务
3. 创建两个 Bucket：
   - `codejym-uploads`（用于存储用户上传文件）
   - `codejym-backups`（用于存储数据库备份）
4. 记录 Bucket 的访问权限设置（建议：私有）

#### 1.3 获取访问凭证

1. 在控制台中找到"访问密钥"或"AccessKey"管理
2. 创建新的访问密钥对
3. 记录以下信息（⚠️ 安全信息，请妥善保管）：
   - **S3_ENDPOINT**（如：`https://s3.bitiful.net`）
   - **S3_REGION**（如：`cn-east-1`）
   - **S3_ACCESS_KEY**（Access Key ID）
   - **S3_SECRET_KEY**（Secret Access Key）

#### 1.4 测试 S3 连接（可选但推荐）

安装 AWS CLI：
```bash
# Ubuntu/Debian
sudo apt-get install awscli

# macOS
brew install awscli
```

测试连接：
```bash
export AWS_ACCESS_KEY_ID="your_access_key"
export AWS_SECRET_ACCESS_KEY="your_secret_key"

# 列出 Bucket
aws s3 ls \
  --endpoint-url https://s3.bitiful.net \
  --region cn-east-1

# 应该看到你创建的 Bucket 列表
```

---

### Phase 2: 更新配置文件（预计 10 分钟）

#### 2.1 更新 `.env` 文件

编辑 `/opt/codejym/.env`：

```bash
# 修改存储类型（从 local 改为 s3）
STORAGE_TYPE=s3

# 填写缤纷云配置
S3_ENDPOINT=https://s3.bitiful.net
S3_REGION=cn-east-1
S3_ACCESS_KEY=你的_access_key
S3_SECRET_KEY=你的_secret_key
S3_BUCKET=codejym-uploads
S3_BACKUP_BUCKET=codejym-backups
S3_URL_PREFIX=    # 可选，暂时留空
```

#### 2.2 验证配置

检查配置文件：
```bash
cat .env | grep S3
```

确保没有语法错误和空格问题。

---

### Phase 3: 执行数据库备份（预计 10 分钟）

#### 3.1 首次手动备份

在开始迁移前，先备份一次数据库：

```bash
cd /opt/codejym

# 加载环境变量
export $(cat .env | grep -v '^#' | xargs)

# 执行备份
./scripts/backup-db-to-s3.sh
```

#### 3.2 验证备份成功

检查 S3 是否有备份文件：
```bash
aws s3 ls s3://codejym-backups/backups/database/full/ \
  --endpoint-url $S3_ENDPOINT \
  --region $S3_REGION
```

你应该看到类似 `codecopybook_20251120_020000.sql.gz` 的文件。

#### 3.3 测试恢复（可选但强烈推荐）

```bash
# 恢复到测试数据库（不要在生产环境执行！）
export PGDATABASE=codecopybook_test
./scripts/restore-db-from-s3.sh latest
```

---

### Phase 4: 重新构建应用（预计 5-10 分钟）

#### 4.1 停止当前服务

```bash
cd /opt/codejym
docker-compose down
```

#### 4.2 重新构建镜像（包含 AWS SDK）

```bash
docker-compose build --no-cache codecopybook
```

这一步会下载 AWS SDK Go v2 依赖，可能需要几分钟。

#### 4.3 启动服务

```bash
docker-compose up -d
```

#### 4.4 检查日志

```bash
docker-compose logs -f codecopybook
```

你应该看到：
```
using S3 storage: endpoint=https://s3.bitiful.net, bucket=codejym-uploads
```

如果看到错误，检查环境变量配置。

---

### Phase 5: 测试新上传功能（预计 5 分钟）

#### 5.1 上传测试文件

1. 访问 CodeJYM 网站
2. 登录账号
3. 创建新训练组，上传一个代码文件
4. 确认上传成功

#### 5.2 验证文件存储在 S3

```bash
aws s3 ls s3://codejym-uploads/uploads/ --recursive \
  --endpoint-url $S3_ENDPOINT \
  --region $S3_REGION
```

你应该看到新上传的文件。

#### 5.3 测试文件下载

1. 在网站上浏览刚上传的文件
2. 开始打字练习
3. 确认文件内容正确显示

---

### Phase 6: 迁移历史文件（预计根据文件大小而定）

#### 6.1 备份本地文件（以防万一）

```bash
cd /opt/codejym
tar -czf data_uploads_backup_$(date +%Y%m%d).tar.gz data/uploads/
```

#### 6.2 运行迁移脚本

```bash
cd /opt/codejym

# 加载环境变量
export $(cat .env | grep -v '^#' | xargs)

# 运行迁移（Go 脚本需要在容器内执行）
docker-compose exec codecopybook /bin/sh -c "
export DATABASE_URL=$DATABASE_URL
export S3_ENDPOINT=$S3_ENDPOINT
export S3_ACCESS_KEY=$S3_ACCESS_KEY
export S3_SECRET_KEY=$S3_SECRET_KEY
export S3_BUCKET=$S3_BUCKET
export S3_REGION=$S3_REGION
export LOCAL_UPLOADS_DIR=/data/uploads

go run /scripts/migrate-files-to-s3.go
"
```

#### 6.3 监控迁移进度

脚本会输出进度：
```
Progress: 10/100 files uploaded (10.0%), 0 failed
Progress: 20/100 files uploaded (20.0%), 0 failed
...
✅ Migration completed!
```

#### 6.4 验证迁移结果

1. 检查 S3 文件数量：
```bash
aws s3 ls s3://codejym-uploads/uploads/ --recursive \
  --endpoint-url $S3_ENDPOINT \
  --region $S3_REGION \
  | wc -l
```

2. 对比本地文件数量：
```bash
find data/uploads/ -type f | wc -l
```

两者应该相同。

#### 6.5 测试历史文件访问

1. 在网站上打开旧的训练组
2. 浏览文件，开始练习
3. 确认功能正常

---

### Phase 7: 清理本地文件（预计 5 分钟）

⚠️ **重要：只有在完全确认 S3 数据正确后才执行此步骤！**

#### 7.1 最终验证

- [ ] 新文件上传正常
- [ ] 历史文件可以访问
- [ ] 数据库备份正常
- [ ] 所有功能测试通过

#### 7.2 删除本地文件

```bash
cd /opt/codejym

# 最后一次确认
ls -lh data/uploads/

# 删除（不可逆！）
rm -rf data/uploads/*

# 验证磁盘空间释放
df -h
```

#### 7.3 保留备份文件7天

```bash
# 备份文件位置
ls -lh data_uploads_backup_*.tar.gz

# 7天后确认无问题再删除
# rm data_uploads_backup_*.tar.gz
```

---

### Phase 8: 配置定时备份（预计 10 分钟）

#### 8.1 配置 Crontab（宿主机）

```bash
# 编辑 crontab
crontab -e

# 添加以下行（每天凌晨 2 点执行备份）
0 2 * * * cd /opt/codejym && export $(cat .env | xargs) && ./scripts/backup-db-to-s3.sh >> /var/log/pg_backup.log 2>&1
```

#### 8.2 或使用 Docker 定时任务（推荐）

创建备份服务容器（已在设计文档中提供配置）。

#### 8.3 测试定时任务

```bash
# 手动触发一次
./scripts/backup-db-to-s3.sh

# 检查日志
tail -f /var/log/pg_backup.log
```

---

## ✅ 验收清单

迁移完成后，请逐项确认：

### 功能验证
- [ ] 用户可以正常登录
- [ ] 新文件上传成功
- [ ] 新文件可以浏览和练习
- [ ] 历史文件可以访问
- [ ] 历史训练进度保留

### 数据验证
- [ ] S3 中有所有文件
- [ ] 数据库连接正常
- [ ] 数据库有最新备份
- [ ] 文件数量和本地一致

### 系统验证
- [ ] 应用日志无错误
- [ ] 本地磁盘空间释放
- [ ] 定时备份配置正常

---

## 🆘 故障排查

### 问题 1：应用启动失败

**症状：**
```
failed to initialize S3 storage: endpoint cannot be empty
```

**解决方案：**
1. 检查 `.env` 文件中 S3 配置是否正确
2. 确认 docker-compose.yml 中环境变量正确传递
3. 重启服务：`docker-compose restart codecopybook`

---

### 问题 2：文件上传失败

**症状：**
```
s3 storage: failed to upload file: ...
```

**解决方案：**
1. 检查 S3 凭证是否正确
2. 检查网络连接：`curl -I https://s3.bitiful.net`
3. 检查 Bucket 权限设置
4. 降级到本地存储测试：`STORAGE_TYPE=local`

---

### 问题 3：历史文件无法访问

**症状：**
404 Not Found 或文件不存在

**解决方案：**
1. 检查数据库中的 `assets.root_path` 是否已更新为 S3 路径
2. 检查 S3 中文件路径是否正确
3. 运行 SQL 查询确认：
```sql
SELECT id, root_path FROM assets LIMIT 10;
```

---

### 问题 4：迁移脚本失败

**症状：**
部分文件上传失败

**解决方案：**
1. 记录失败的文件列表
2. 手动上传失败的文件
3. 或重新运行迁移脚本（会跳过已存在的文件）

---

### 问题 5：备份失败

**症状：**
```
pg_dump failed
```

**解决方案：**
1. 检查数据库连接：`psql -h postgres -U codecopy -d codecopybook`
2. 检查磁盘空间：`df -h`
3. 检查 PostgreSQL 日志：`docker-compose logs postgres`

---

## 🔄 回滚方案

如果迁移出现严重问题，可以回滚：

### 回滚步骤

1. **停止服务**
```bash
docker-compose down
```

2. **修改配置为本地存储**
```bash
# 编辑 .env
STORAGE_TYPE=local
```

3. **恢复本地文件**（如果已删除）
```bash
tar -xzf data_uploads_backup_*.tar.gz
```

4. **重启服务**
```bash
docker-compose up -d
```

5. **验证功能**
确认所有功能恢复正常。

---

## 📞 获取帮助

如果遇到问题：

1. 查看完整设计文档：`STORAGE_MIGRATION_DESIGN.md`
2. 查看应用日志：`docker-compose logs -f`
3. 查看数据库日志：`docker-compose logs postgres`
4. 检查系统资源：`htop` 或 `docker stats`

---

## 📚 参考资料

- [缤纷云官方文档](https://docs.bitiful.com/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/docs/)
- [PostgreSQL 备份最佳实践](https://www.postgresql.org/docs/current/backup.html)

---

**文档结束 - 祝迁移顺利！** 🎉

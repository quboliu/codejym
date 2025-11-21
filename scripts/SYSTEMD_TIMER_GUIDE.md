# Systemd Timer 定时备份指南

## 📋 概述

使用 systemd timer 替代 cron 进行数据库定时备份。

**优势**：
- ✅ 更好的日志管理（集成 journalctl）
- ✅ 依赖管理（确保 Docker 服务已启动）
- ✅ 失败重试机制
- ✅ 错过的任务会在系统启动后执行（Persistent=true）
- ✅ 更现代化的系统管理方式

**劣势**：
- ❌ 需要 root 权限安装
- ❌ 配置相对复杂

---

## 📁 配置文件

### 1. pg-backup.service
定义备份任务的执行方式。

**位置**：`scripts/pg-backup.service`

**配置说明**：
```ini
[Unit]
Description=PostgreSQL Database Backup to S3
After=network.target docker.service  # 确保网络和 Docker 已启动
Documentation=file:///opt/codejym/scripts/BACKUP_README.md

[Service]
Type=oneshot                         # 单次执行任务
User=codejym                          # 执行用户
Group=docker                         # 用户组
WorkingDirectory=/opt/codejym
ExecStart=/opt/codejym/scripts/run-backup.sh

# 日志输出方案选择
StandardOutput=journal               # 输出到 systemd journal
StandardError=journal

# 或者输出到文件（与 cron 一致）
# StandardOutput=append:/opt/codejym/logs/pg-backup.log
# StandardError=append:/opt/codejym/logs/pg-backup.log

Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

[Install]
WantedBy=multi-user.target
```

### 2. pg-backup.timer
定义备份任务的执行时间。

**位置**：`scripts/pg-backup.timer`

**配置说明**：
```ini
[Unit]
Description=PostgreSQL Database Backup Timer
Documentation=file:///opt/codejym/scripts/BACKUP_README.md
Requires=pg-backup.service           # 依赖 service 文件

[Timer]
OnCalendar=*-*-* 02:00:00            # 每天凌晨 2:00
Persistent=true                       # 错过的任务会补执行

[Install]
WantedBy=timers.target
```

---

## 🚀 启用步骤

### 步骤 1：安装配置文件

```bash
# 复制 service 和 timer 文件到系统目录
sudo cp /opt/codejym/scripts/pg-backup.service /etc/systemd/system/
sudo cp /opt/codejym/scripts/pg-backup.timer /etc/systemd/system/

# 设置正确的权限
sudo chmod 644 /etc/systemd/system/pg-backup.service
sudo chmod 644 /etc/systemd/system/pg-backup.timer
```

### 步骤 2：重载 systemd 配置

```bash
sudo systemctl daemon-reload
```

### 步骤 3：启用并启动 timer

```bash
# 启用 timer（开机自启动）
sudo systemctl enable pg-backup.timer

# 启动 timer
sudo systemctl start pg-backup.timer
```

### 步骤 4：验证配置

```bash
# 查看 timer 状态
sudo systemctl status pg-backup.timer

# 列出所有 timer，找到 pg-backup.timer
sudo systemctl list-timers --all

# 查看下次执行时间
sudo systemctl list-timers pg-backup.timer
```

**预期输出示例**：
```
NEXT                         LEFT          LAST PASSED UNIT              ACTIVATES
Thu 2025-11-21 02:00:00 PST  2h 55min left n/a  n/a    pg-backup.timer   pg-backup.service
```

---

## 🔍 日常管理

### 查看 Timer 状态

```bash
# 查看 timer 状态
sudo systemctl status pg-backup.timer

# 查看 service 状态
sudo systemctl status pg-backup.service

# 查看所有定时器
sudo systemctl list-timers
```

### 查看备份日志

#### 方案 1：使用 journalctl（推荐）

```bash
# 查看最近的备份日志
sudo journalctl -u pg-backup.service

# 查看最近 50 条日志
sudo journalctl -u pg-backup.service -n 50

# 实时查看日志
sudo journalctl -u pg-backup.service -f

# 查看今天的日志
sudo journalctl -u pg-backup.service --since today

# 查看特定日期的日志
sudo journalctl -u pg-backup.service --since "2025-11-20" --until "2025-11-21"

# 查看最近一次执行的日志
sudo journalctl -u pg-backup.service -n 100 --no-pager | grep -A 20 "Starting database backup"
```

#### 方案 2：使用日志文件

如果配置输出到文件（需修改 service 文件）：

```bash
# 查看日志文件
tail -f /opt/codejym/logs/pg-backup.log

# 搜索错误
grep -i error /opt/codejym/logs/pg-backup.log
```

### 手动触发备份

```bash
# 手动执行一次备份（不影响定时器）
sudo systemctl start pg-backup.service

# 查看执行结果
sudo systemctl status pg-backup.service
```

### 修改执行时间

```bash
# 编辑 timer 配置
sudo systemctl edit --full pg-backup.timer

# 修改 OnCalendar 行，例如：
# OnCalendar=*-*-* 03:00:00  # 改为凌晨 3:00

# 重载配置
sudo systemctl daemon-reload

# 重启 timer
sudo systemctl restart pg-backup.timer

# 验证新的执行时间
sudo systemctl list-timers pg-backup.timer
```

### 停止和禁用

```bash
# 停止 timer（但不禁用开机自启动）
sudo systemctl stop pg-backup.timer

# 禁用 timer（取消开机自启动，但不停止当前运行）
sudo systemctl disable pg-backup.timer

# 同时停止并禁用
sudo systemctl disable --now pg-backup.timer

# 验证已停止
sudo systemctl status pg-backup.timer
```

---

## ⚙️ 高级配置

### 1. 添加失败重试

编辑 `pg-backup.service`，添加以下配置：

```ini
[Service]
# ... 现有配置 ...

# 失败后重试
Restart=on-failure
RestartSec=5min
StartLimitBurst=3
StartLimitIntervalSec=1h
```

### 2. 添加超时限制

```ini
[Service]
# ... 现有配置 ...

# 超时配置
TimeoutStartSec=10min
TimeoutStopSec=5min
```

### 3. 发送邮件通知

编辑 `pg-backup.service`，添加：

```ini
[Service]
# ... 现有配置 ...

# 失败时发送邮件
OnFailure=status-email@%n.service
```

需要先配置 `status-email@.service`（参考 systemd 文档）。

### 4. 调整日志级别

```ini
[Service]
# ... 现有配置 ...

# 日志级别
SyslogLevel=info
SyslogIdentifier=pg-backup
```

### 5. 多个执行时间

如果需要每天执行多次：

```ini
[Timer]
# 每天凌晨 2:00
OnCalendar=*-*-* 02:00:00
# 每天下午 2:00
OnCalendar=*-*-* 14:00:00
```

或每 6 小时一次：

```ini
[Timer]
OnCalendar=*-*-* 00/6:00:00  # 0:00, 6:00, 12:00, 18:00
```

---

## 🔧 故障排查

### 1. Timer 未执行

**检查 timer 是否启用**：
```bash
sudo systemctl is-enabled pg-backup.timer
sudo systemctl is-active pg-backup.timer
```

**查看 timer 列表**：
```bash
sudo systemctl list-timers --all | grep pg-backup
```

**检查 timer 配置**：
```bash
sudo systemctl cat pg-backup.timer
```

### 2. Service 执行失败

**查看详细状态**：
```bash
sudo systemctl status pg-backup.service -l
```

**查看完整日志**：
```bash
sudo journalctl -u pg-backup.service -xe
```

**常见错误**：

| 错误 | 原因 | 解决方法 |
|------|------|---------|
| `Failed to start` | 权限问题或路径错误 | 检查 User、Group、WorkingDirectory |
| `Timeout` | 执行时间过长 | 增加 TimeoutStartSec |
| `Exit code 1` | 脚本执行失败 | 查看日志，检查数据库和 S3 连接 |

### 3. 权限问题

**确保用户有 Docker 权限**：
```bash
# 检查用户组
groups codejym

# 应该包含 docker 组
# 如果没有，添加：
sudo usermod -aG docker codejym
```

**脚本权限**：
```bash
# 确保脚本可执行
chmod +x /opt/codejym/scripts/run-backup.sh
chmod +x /opt/codejym/scripts/backup-db-to-s3.sh
```

### 4. 验证配置文件

```bash
# 验证 service 配置
sudo systemd-analyze verify /etc/systemd/system/pg-backup.service

# 验证 timer 配置
sudo systemd-analyze verify /etc/systemd/system/pg-backup.timer

# 如果没有输出，说明配置正确
```

---

## 📊 监控建议

### 创建监控脚本

创建 `/opt/codejym/scripts/check-backup-status.sh`：

```bash
#!/bin/bash
# 检查最近一次备份是否成功

LAST_RUN=$(sudo journalctl -u pg-backup.service --since "24 hours ago" | grep "Backup completed successfully" | tail -1)

if [ -z "$LAST_RUN" ]; then
    echo "WARNING: No successful backup in the last 24 hours"
    exit 1
else
    echo "OK: Last backup was successful"
    echo "$LAST_RUN"
    exit 0
fi
```

### 添加监控 timer

创建 `check-backup.timer`，每天检查备份状态：

```ini
[Timer]
OnCalendar=*-*-* 08:00:00
Persistent=true
```

---

## 🆚 Cron vs Systemd Timer 对比

| 特性 | Cron | Systemd Timer |
|------|------|---------------|
| **易用性** | ✅ 简单 | ❌ 复杂 |
| **日志** | 需要手动重定向 | ✅ 集成 journalctl |
| **依赖管理** | ❌ 无 | ✅ 可以设置依赖 |
| **错过的任务** | ❌ 跳过 | ✅ 可以补执行 |
| **失败重试** | ❌ 无 | ✅ 支持 |
| **权限** | 普通用户 | ✅ 需要 root |
| **监控** | ❌ 困难 | ✅ systemctl status |
| **随机延迟** | ❌ 无 | ✅ RandomizedDelaySec |

---

## 📝 从 Cron 迁移到 Systemd Timer

### 步骤 1：验证 Cron 任务正常

```bash
# 查看当前 cron 配置
crontab -l

# 手动执行确认正常
./scripts/run-backup.sh
```

### 步骤 2：安装 Systemd Timer

按照上面的"启用步骤"安装并启动 timer。

### 步骤 3：并行运行一段时间

保持 cron 和 systemd timer 同时运行，观察 1-2 天。

```bash
# 查看 cron 执行记录
tail -f /opt/codejym/logs/pg-backup.log

# 查看 systemd 执行记录
sudo journalctl -u pg-backup.service -f
```

### 步骤 4：确认无误后移除 Cron

```bash
# 编辑 crontab
crontab -e

# 注释或删除备份任务行
# 0 2 * * * /opt/codejym/scripts/run-backup.sh >> /opt/codejym/logs/pg-backup.log 2>&1

# 验证
crontab -l
```

---

## 🔒 安全最佳实践

### 1. 限制文件权限

```bash
# systemd 配置文件
sudo chmod 644 /etc/systemd/system/pg-backup.*

# 脚本文件
chmod 700 /opt/codejym/scripts/run-backup.sh
```

### 2. 启用安全选项

编辑 `pg-backup.service`，取消注释安全选项：

```ini
[Service]
# ... 现有配置 ...

# 安全选项
PrivateTmp=yes
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/opt/codejym/logs
```

### 3. 审计日志

```bash
# 定期检查执行记录
sudo journalctl -u pg-backup.service --since "7 days ago" | grep -E "Starting|completed|failed"
```

---

## 📚 相关文档

- [Systemd Timer 官方文档](https://www.freedesktop.org/software/systemd/man/systemd.timer.html)
- [Systemd Service 官方文档](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [备份系统使用说明](BACKUP_README.md)
- [数据库备份设计方案](../docs/migration/DATABASE_BACKUP_DESIGN.md)

---

## 📞 快速参考

### 常用命令

```bash
# 安装
sudo cp scripts/pg-backup.{service,timer} /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now pg-backup.timer

# 查看状态
sudo systemctl status pg-backup.timer
sudo systemctl list-timers pg-backup.timer

# 查看日志
sudo journalctl -u pg-backup.service -n 50

# 手动执行
sudo systemctl start pg-backup.service

# 停止
sudo systemctl disable --now pg-backup.timer

# 重载配置
sudo systemctl daemon-reload
sudo systemctl restart pg-backup.timer
```

---

**文档版本**：1.0
**最后更新**：2025-11-20
**维护者**：CodeJYM team

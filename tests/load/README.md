# 进度保存并发压测（Load test: progress-save path）

回答一个问题：**同一时刻有 N 个人在打字练习、每人都要保存/同步进度，我们扛得住吗？**

> 现状（2026-06）：项目此前**没有任何自动化测试**（0 个 `*_test.go`，0 个压测）。
> 这是补上的第一个负载测试基线。

## 被测路径
每个练习者的前端每 **1.2s** 调一次 `PATCH /api/sessions/:id` 保存进度（光标/错误数/用时）。
单次保存在后端是 **3 次数据库往返**：
1. `withAuth` → `GetUserByID`（每个鉴权请求都查一次 user 表，纯开销）
2. `updateSession` → `GetSession`（SELECT）
3. `updateSession` → `UpdateSession`（UPDATE，按主键，无行竞争）

## 怎么跑
```bash
# 1) 起服务（见根目录部署脚本），确认可访问
export BASE_URL=http://localhost:8081

# 2) 预置 N 个虚拟用户（每人 1 个 asset + 1 个 session），生成 /tmp/vus.json
node tests/load/provision-users.mjs 200

# 3) 闭环加压，阶梯式拉到 200 并发，找到饱和点
k6 run tests/load/save-progress.k6.js
```

## SLO（阈值，写在 k6 script 里）
- `http_req_duration` p95 < 200ms，p99 < 500ms
- 错误率 < 1%

## 基线结果（8C 机器、与其他容器共享、Postgres 同机）
| 指标 | 值 |
|---|---|
| 最大吞吐 | ~735 saves/s |
| 200 并发下延迟 | avg 104ms / p95 **242ms（超标）** / p99 275ms |
| 错误率 | 0%（不是报错，而是排队等连接，延迟上升） |
| DB 连接 | 全程顶在 **9**（`pgxpool` 默认 ≈ max(4, CPU 数)，未配置） |

**换算**（每人 0.83 saves/s）：当前架构上限 ≈ **600–880 并发练习者**。
- 100 人 ✅ 轻松（~11% 容量）
- 1000 人 ❌ 超过当前上限，会排队 / 延迟飙升
- 10000 人 ❌ 约 11× over，需架构升级（连接池 + 单请求单查询 + 写合并/批量 + 水平扩展）

瓶颈是**连接池 + 每次保存 3 次查询 + 单实例同步写**，不是行锁竞争（每人写自己的行）。
详见架构升级建议。

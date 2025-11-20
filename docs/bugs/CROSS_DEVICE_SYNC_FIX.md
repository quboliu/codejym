# 跨设备进度同步修复说明

## 问题描述

用户在浏览器 A 上的练习进度，在另一台设备的浏览器 B 上登录后无法看到，进度没有同步。

## 根本原因

**之前的实现问题**：

1. **依赖 localStorage**：前端使用 localStorage 存储 session ID
2. **localStorage 的局限性**：localStorage 是浏览器本地存储，不同设备/浏览器之间无法共享
3. **导致重复创建 session**：在新设备上，因为 localStorage 为空，系统会创建新的 session 而不是获取现有的进度

### 问题流程

```
浏览器 A:
1. 用户开始练习文件 A
2. 创建 session_123
3. session_123 的 ID 存储在浏览器 A 的 localStorage ✅
4. 进度上传到服务器 ✅

浏览器 B (同一用户登录):
1. 用户选择同一文件 A
2. localStorage 为空 ❌
3. 创建新的 session_456 ❌
4. 无法看到 session_123 的进度 ❌
```

## 修复方案

### 核心思路

**改为查询服务器**：不依赖 localStorage，而是向服务器查询该用户对该文件是否已存在 session。

### 修改内容

#### 1. 后端 - Storage 层 (backend/internal/storage/storage.go)

添加新方法根据用户、素材、文件路径查询现有 session：

```go
// GetSessionByAssetAndPath retrieves an existing session by user, asset, and file path.
func (s *Storage) GetSessionByAssetAndPath(ctx context.Context, userID, assetID, relPath string) (*Session, error) {
    sess := &Session{}
    err := s.db.QueryRow(
        ctx,
        `SELECT id, user_id, asset_id, rel_path, cursor, errors, duration_seconds, created_at, updated_at
         FROM typing_sessions WHERE user_id = $1 AND asset_id = $2 AND rel_path = $3
         ORDER BY updated_at DESC LIMIT 1`,
        userID, assetID, relPath,
    ).Scan(&sess.ID, &sess.UserID, &sess.AssetID, &sess.RelPath, &sess.Cursor, &sess.Errors, &sess.DurationSeconds, &sess.CreatedAt, &sess.UpdatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return sess, nil
}
```

**关键点**：
- 使用 `ORDER BY updated_at DESC LIMIT 1` 获取最新的 session
- 如果一个用户对同一文件有多个 session（因为之前的 bug），返回最新的

#### 2. 后端 - API 层 (backend/internal/api/server.go)

添加 GET 方法到 `/api/sessions`：

```go
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
    user := currentUser(r)
    if user == nil {
        writeError(w, http.StatusUnauthorized, "not authorized")
        return
    }
    switch r.Method {
    case http.MethodGet:
        s.querySession(user, w, r)  // 新增：查询现有 session
    case http.MethodPost:
        s.createSession(user, w, r)
    default:
        methodNotAllowed(w, http.MethodGet, http.MethodPost)
    }
}

func (s *Server) querySession(user *storage.User, w http.ResponseWriter, r *http.Request) {
    assetID := r.URL.Query().Get("assetId")
    path := r.URL.Query().Get("path")
    if assetID == "" || path == "" {
        writeError(w, http.StatusBadRequest, "assetId and path query parameters are required")
        return
    }
    session, err := s.store.GetSessionByAssetAndPath(r.Context(), user.ID, assetID, path)
    if err != nil {
        if errors.Is(err, storage.ErrNotFound) {
            writeError(w, http.StatusNotFound, "session not found")
        } else {
            writeError(w, http.StatusInternalServerError, "failed to query session")
        }
        return
    }
    writeJSON(w, http.StatusOK, session)
}
```

**API 端点**：
```
GET /api/sessions?assetId=xxx&path=yyy
```

#### 3. 前端 - API (frontend/src/api.ts)

添加查询 session 的方法：

```typescript
export function querySession(assetId: string, filePath: string) {
  const encoded = encodeURIComponent(filePath);
  return request<Session>(`/api/sessions?assetId=${assetId}&path=${encoded}`);
}
```

#### 4. 前端 - 逻辑 (frontend/src/App.vue)

完全重写 `ensureSession` 函数：

```typescript
// 修复前 - 依赖 localStorage
async function ensureSession(assetId: string, filePath: string) {
  const storageKey = sessionKey(user.value?.id ?? 'anon', assetId, filePath)
  let sessionData: Session | null = null

  const existingId = localStorage.getItem(storageKey)  // ❌ 只在本地查找
  if (existingId) {
    try {
      sessionData = await fetchSession(existingId)
    } catch {
      localStorage.removeItem(storageKey)
    }
  }

  if (!sessionData) {
    sessionData = await createSession(assetId, filePath)
    localStorage.setItem(storageKey, sessionData.id)
  }

  return sessionData
}

// 修复后 - 查询服务器
async function ensureSession(assetId: string, filePath: string) {
  // 首先尝试从服务器查询现有的 session
  try {
    const sessionData = await querySession(assetId, filePath)
    return sessionData  // ✅ 找到现有 session，直接使用
  } catch (err) {
    // 如果没有找到现有的 session（404），创建一个新的
    const error = err as Error
    if (error.message.includes('session not found')) {
      const newSession = await createSession(assetId, filePath)
      return newSession  // ✅ 创建新 session
    }
    // 其他错误则抛出
    throw err
  }
}
```

**关键改进**：
- ✅ 总是先查询服务器是否有现有 session
- ✅ 如果找到，直接使用（跨设备同步）
- ✅ 如果没找到，才创建新的
- ✅ 移除了对 localStorage 的依赖

## 修复后的流程

```
浏览器 A:
1. 用户开始练习文件 A
2. 向服务器查询现有 session → 没有
3. 创建 session_123
4. 进度实时上传到服务器 ✅

浏览器 B (同一用户登录):
1. 用户选择同一文件 A
2. 向服务器查询现有 session → 找到 session_123 ✅
3. 加载 session_123 的进度（cursor, errors, duration） ✅
4. 用户可以从上次的进度继续 ✅
```

## 修改的文件

1. **backend/internal/storage/storage.go** - 添加 `GetSessionByAssetAndPath` 方法
2. **backend/internal/api/server.go** - 添加 GET /api/sessions 端点和 `querySession` 处理函数
3. **frontend/src/api.ts** - 添加 `querySession` API 调用
4. **frontend/src/App.vue** - 重写 `ensureSession` 逻辑，移除 `sessionKey` 函数

## 部署

运行以下命令应用修复：

```bash
cd /opt/codejym
bash ./deploy-full.sh
```

## 测试步骤

### 1. 单设备测试
```bash
1. 在浏览器 A 登录
2. 选择一个文件，练习到 50% 进度
3. 刷新页面
4. 验证：进度应该保持在 50%
```

### 2. 跨设备测试
```bash
1. 在浏览器 A 登录
2. 选择一个文件，练习到 50% 进度
3. 在浏览器 B（或另一台设备）用同一账号登录
4. 选择同一文件
5. 验证：进度应该从 50% 开始，而不是从 0% 开始
```

### 3. 多文件测试
```bash
1. 练习文件 A 到 30%
2. 切换到文件 B，练习到 60%
3. 在另一设备登录
4. 验证：
   - 文件 A 进度 30%
   - 文件 B 进度 60%
```

## 性能考虑

### 当前实现
- 每次选择文件时查询一次服务器
- 如果找到现有 session，直接使用
- 进度定期（每 1.2 秒）同步到服务器

### 可选优化（未实现）

如果发现性能问题，可以考虑：

1. **添加缓存层**：
   ```typescript
   const sessionCache = new Map<string, { session: Session, timestamp: number }>()
   const CACHE_TTL = 60000 // 1 分钟
   ```

2. **使用 localStorage 作为缓存**：
   - 仍然查询服务器，但使用 localStorage 缓存结果
   - 设置短期 TTL（如 5 分钟）
   - 可以减少服务器查询次数

但目前的实现已经足够高效，因为：
- 只在选择文件时查询一次
- 后续进度更新使用已有的 session ID
- 数据库查询有索引支持（`idx_sessions_user_asset`）

## 数据库索引

确保数据库有正确的索引：

```sql
CREATE INDEX IF NOT EXISTS idx_sessions_user_asset
ON typing_sessions(user_id, asset_id);
```

这个索引已经在 `Migrate` 函数中创建，确保查询性能。

## 总结

此次修复的核心是：**将 session 查询逻辑从客户端本地存储转移到服务器端**。这样可以实现真正的跨设备进度同步，符合用户的预期。

**关键优势**：
- ✅ 跨设备进度同步
- ✅ 单一数据源（服务器）
- ✅ 代码更简洁（移除了 localStorage 依赖）
- ✅ 更好的数据一致性

**注意事项**：
- 进度仍然定期（每 1.2 秒）自动保存到服务器
- 用户在任何设备上的最新进度都会被保存
- 如果多设备同时编辑同一文件，后提交的进度会覆盖之前的（这是预期行为）

# 实时保存进度逻辑分析报告

## 📋 当前实现逻辑

### 1. 自动保存机制

```typescript
useEffect(() => {
  if (!session) {
    return;
  }
  // 使用 setInterval 每 1.2 秒保存一次进度
  const interval = window.setInterval(() => {
    patchSession(session.id, {
      cursor: cursorRef.current,
      errors: errorsRef.current,
      durationSeconds: Math.round(elapsedSecondsRef.current),
    }).catch((err) => console.warn('session sync failed', err));
  }, 1200);
  return () => window.clearInterval(interval);
}, [session?.id]);
```

**关键要点**:
- 每 1.2 秒自动保存一次
- 使用 `setInterval` 而非 `setTimeout`
- 使用 `useRef` 存储最新状态值，避免频繁重建
- 只依赖于 `session?.id`，减少不必要的重建

### 2. 状态同步机制

```typescript
// 使用 useRef 存储最新状态
const cursorRef = useRef(cursor);
const errorsRef = useRef(errors);
const elapsedSecondsRef = useRef(elapsedSeconds);

// 同步状态到 ref
useEffect(() => {
  cursorRef.current = cursor;
}, [cursor]);

useEffect(() => {
  errorsRef.current = errors;
}, [errors]);

useEffect(() => {
  elapsedSecondsRef.current = elapsedSeconds;
}, [elapsedSeconds]);
```

### 3. 保存流程

```
1. 用户输入字符 → cursor 状态更新
2. cursor 状态更新 → cursorRef.current 同步更新
3. 每 1.2 秒 → patchSession API 调用
4. API 调用 → 数据库更新 (cursor, errors, duration_seconds)
```

## 🔍 问题分析

### 原始问题
**现象**: PATCH API 调用间隔为 60 秒，而非预期的 1.2 秒

**根因分析**:
```typescript
// 原始问题代码
useEffect(() => {
  const timeout = window.setTimeout(() => {
    patchSession(...);
  }, 1200);
  return () => window.clearTimeout(timeout);
}, [session?.id, cursor, errors, elapsedSeconds]); // ❌ 依赖项过多
```

**问题解释**:
1. `elapsedSeconds` 每秒更新一次 (useState + setInterval)
2. 每次 `elapsedSeconds` 变化 → useEffect 重新执行
3. 重新执行 → 清除之前的 timeout
4. 结果：timeout 永远不会被执行

### 修复方案
```typescript
// 修复后代码
useEffect(() => {
  const interval = window.setInterval(() => {
    // 使用 ref 获取最新值
    patchSession(session.id, {
      cursor: cursorRef.current,
      errors: errorsRef.current,
      durationSeconds: Math.round(elapsedSecondsRef.current),
    });
  }, 1200);
  return () => window.clearInterval(interval);
}, [session?.id]); // ✅ 只依赖 session
```

**修复要点**:
1. 使用 `setInterval` 替代 `setTimeout`
2. 使用 `useRef` 存储最新状态
3. 减少 useEffect 依赖项
4. 避免频繁重建 interval

## 🧪 测试验证方法

### 方法 1: 查看后端日志

```bash
# 监控 PATCH API 调用
docker compose -f docker-compose.proxy.yml logs -f codecopybook | grep PATCH
```

**预期结果**: 每 1.2 秒应看到一次 PATCH 调用

### 方法 2: 查看数据库变化

```bash
# 查看 typing_sessions 表
docker exec codejym-postgres-1 psql -U codecopy -d codecopybook \
  -c "SELECT id, cursor, errors, duration_seconds, updated_at \
      FROM typing_sessions ORDER BY updated_at DESC LIMIT 1;"
```

**验证**:
- `updated_at` 每 1.2 秒更新一次
- `duration_seconds` 持续增长

### 方法 3: 查看浏览器控制台

打开浏览器开发者工具 → Console，查看日志：

```
[DEBUG] 创建进度保存 interval，session: f7e812af55451522a0cc
[DEBUG] 保存进度: {cursor: 5, errors: 0, durationSeconds: 12}
[DEBUG] 保存进度: {cursor: 5, errors: 0, durationSeconds: 13}
...
```

### 方法 4: 模拟用户操作测试

1. 登录应用
2. 上传或选择一个代码文件
3. 点击文件进入练习模式
4. 输入一些字符
5. 观察控制台和数据库变化

## 📊 代码修改记录

### 修改前 (有问题的版本)

```typescript
useEffect(() => {
  if (!session) {
    return;
  }
  const timeout = window.setTimeout(() => {
    patchSession(session.id, {
      cursor,
      errors,
      durationSeconds: Math.round(elapsedSeconds),
    }).catch((err) => console.warn('session sync failed', err));
  }, 1200);
  return () => window.clearTimeout(timeout);
}, [session?.id, cursor, errors, elapsedSeconds]); // ❌ 问题：依赖项太多
```

**问题**:
- 依赖项过多：`session?.id`, `cursor`, `errors`, `elapsedSeconds`
- `elapsedSeconds` 每秒变化 → useEffect 频繁重建
- setTimeout 永远等不到执行就被清除

### 修改后 (修复版本)

```typescript
// 1. 添加 ref 存储状态
const cursorRef = useRef(cursor);
const errorsRef = useRef(errors);
const elapsedSecondsRef = useRef(elapsedSeconds);

// 2. 同步状态到 ref
useEffect(() => {
  cursorRef.current = cursor;
}, [cursor]);

useEffect(() => {
  errorsRef.current = errors;
}, [errors]);

useEffect(() => {
  elapsedSecondsRef.current = elapsedSeconds;
}, [elapsedSeconds]);

// 3. 保存逻辑
useEffect(() => {
  if (!session) {
    return;
  }
  console.log('[DEBUG] 创建进度保存 interval，session:', session.id);
  const interval = window.setInterval(() => {
    console.log('[DEBUG] 保存进度:', {
      cursor: cursorRef.current,
      errors: errorsRef.current,
      durationSeconds: Math.round(elapsedSecondsRef.current),
    });
    patchSession(session.id, {
      cursor: cursorRef.current,
      errors: errorsRef.current,
      durationSeconds: Math.round(elapsedSecondsRef.current),
    }).catch((err) => console.warn('session sync failed', err));
  }, 1200);
  return () => {
    console.log('[DEBUG] 清除进度保存 interval');
    window.clearInterval(interval);
  };
}, [session?.id]); // ✅ 只依赖 session
```

## 🔧 部署状态

**当前状态**:
- ✅ 代码已修复
- ✅ 应用已重新构建
- ✅ 应用已重新部署

**验证时间**: 2025-11-17 10:03

**注意事项**:
- 用户需要进入练习模式才会触发保存逻辑
- 如果用户只是浏览文件列表，不会触发 save
- 必须在练习模式下输入字符或经过时间才会看到效果

## 📝 总结

### 修复内容
1. ✅ 使用 `setInterval` 替代 `setTimeout`
2. ✅ 使用 `useRef` 存储最新状态
3. ✅ 减少 useEffect 依赖项
4. ✅ 添加调试日志

### 预期效果
- PATCH API 调用间隔：1.2 秒
- 数据库更新时间：每 1.2 秒
- 控制台日志：每 1.2 秒输出一次

### 测试建议
1. 进入练习模式
2. 输入一些字符
3. 观察后端日志：`docker compose logs -f | grep PATCH`
4. 观察数据库：`SELECT * FROM typing_sessions ORDER BY updated_at DESC LIMIT 1;`
5. 观察控制台：浏览器开发者工具 Console

---

**如果仍未看到预期效果，请检查**:
1. 是否在练习模式下
2. 是否有 JavaScript 错误
3. session 是否正确创建
4. 网络请求是否成功
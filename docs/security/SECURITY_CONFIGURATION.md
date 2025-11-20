# HTTP 安全配置说明

## 🔒 安全措施总览

本次更新为 CodeJYM 添加了严格的安全配置，主要包括：

1. **禁用 Cookies 认证**
2. **JWT Token 超时控制**
3. **安全响应头**
4. **缓存控制**

---

## 📋 详细说明

### 1. Cookies 禁用 ✅

**当前状态**: 系统**完全不使用 Cookies**进行身份验证

**实现方式**:
- 使用 `Authorization: Bearer <JWT Token>` Header 进行认证
- 服务器响应中明确设置 `Set-Cookie` 头，使任何 Cookies 立即过期
- 响应头 `X-Auth-Method: JWT Bearer Token (no cookies)` 明确说明认证方式

**优势**:
- 避免 CSRF 攻击
- 避免 XSS 攻击窃取 Cookie
- 更符合现代 RESTful API 最佳实践
- 更好的移动端和跨域支持

### 2. JWT Token 超时控制 ⚙️

**配置位置**:
- 环境变量: `AUTH_TOKEN_TTL`
- 默认值: `24小时` (比原来的 30 天更安全)
- 可配置格式:
  - `30m` - 30 分钟
  - `1h` - 1 小时
  - `24h` - 24 小时
  - `7d` - 7 天

**配置方法**:

```yaml
# docker-compose.yml
environment:
  - AUTH_TOKEN_TTL=24h  # Token 有效期 24 小时
```

```bash
# 直接运行
export AUTH_TOKEN_TTL=30m
./server -addr :8080
```

**推荐配置**:
- **开发环境**: `30m` (30 分钟)
- **测试环境**: `24h` (24 小时)
- **生产环境**: `7d` (7 天)

### 3. 安全响应头 🛡️

#### 缓存控制
```
Cache-Control: no-cache, no-store, must-revalidate, private
Pragma: no-cache
Expires: 0
```
- 禁用浏览器缓存，避免敏感数据被缓存

#### 内容安全
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```
- 防止 MIME 类型嗅探
- 禁止页面被嵌入 iframe
- 防止 XSS 攻击
- 控制引用者信息

#### 认证头
```
WWW-Authenticate: Bearer realm="CodeJYM API"
X-Auth-Method: JWT Bearer Token (no cookies)
```
- 明确说明认证方式
- 提示客户端使用 Bearer Token

### 4. 清除现有 Cookies ⚡

服务器会自动清除任何客户端发送的 Cookies:

```
Set-Cookie: Path=/; HttpOnly; Max-Age=0
```

这确保了：
- 客户端无法使用 Cookie 认证
- 任何现有的 Cookie 都立即失效
- 强制使用 Authorization Header

---

## 🔍 验证方法

### 1. 检查响应头
```bash
# 登录并检查响应头
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' \
  -i

# 应该看到：
# Set-Cookie: Path=/; HttpOnly; Max-Age=0
# X-Auth-Method: JWT Bearer Token (no cookies)
# WWW-Authenticate: Bearer realm="CodeJYM API"
```

### 2. 检查 Token 过期时间
```bash
# 解码 JWT Token 查看过期时间
# JWT 格式: header.payload.signature
# 使用在线工具或: https://jwt.io/
```

### 3. 使用 Authorization Header 认证
```bash
# 正确的认证方式
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:8080/api/auth/me

# 错误的认证方式（不会使用 Cookie）
curl -b "session=xxx" http://localhost:8080/api/auth/me
# 将会被忽略
```

---

## 📚 最佳实践

### 客户端开发

#### ✅ 正确做法
```javascript
// 1. 登录后获取 Token
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email, password })
});
const { token } = await response.json();

// 2. 存储 Token（建议使用内存或安全存储）
localStorage.setItem('auth_token', token);

// 3. 后续请求使用 Authorization Header
const authHeader = {
  'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
  'Content-Type': 'application/json'
};

// 4. 处理 Token 过期
try {
  const data = await fetch('/api/auth/me', { headers: authHeader });
} catch (error) {
  if (error.status === 401) {
    // Token 过期，重新登录
    redirectToLogin();
  }
}
```

#### ❌ 错误做法
```javascript
// 不要使用 Cookies
document.cookie = "token=xxx";  // 会被服务器忽略

// 不要在 URL 中传递 Token
fetch(`/api/data?token=${token}`);  // 不安全

// 不要信任客户端 Cookie 认证
fetch('/api/data', {
  credentials: 'include'  // 不需要，因为不使用 Cookie
});
```

### 服务器部署

#### 开发环境
```yaml
environment:
  - AUTH_TOKEN_TTL=30m
```

#### 生产环境
```yaml
environment:
  - AUTH_TOKEN_TTL=7d  # 或更短，如 24h
  - AUTH_SECRET=your-very-secure-secret-key
```

---

## ⚠️ 重要提醒

1. **Token 过期**: Token 过期后需要重新登录
2. **无状态**: 服务器不存储会话状态，所有认证信息在 Token 中
3. **安全传输**: 始终使用 HTTPS 传输 Token
4. **密钥安全**: `AUTH_SECRET` 必须保密，建议使用强随机密钥
5. **不要在前端暴露敏感信息**: 避免在前端代码中硬编码密钥

---

## 🧪 测试场景

### 场景 1: Token 自动过期
```bash
# 1. 登录获取 Token（设置 1 分钟过期）
export AUTH_TOKEN_TTL=1m

# 2. 立即访问 API - 应该成功
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/auth/me

# 3. 等待 1 分钟
sleep 60

# 4. 再次访问 - 应该返回 401 Unauthorized
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/auth/me
```

### 场景 2: Cookies 被忽略
```bash
# 1. 设置 Cookie
curl -c cookies.txt -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# 2. 使用 Cookie 访问 - 应该被忽略，返回 401
curl -b cookies.txt http://localhost:8080/api/auth/me

# 3. 使用 Bearer Token 访问 - 应该成功
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/auth/me
```

---

## 📊 配置变更对比

| 项目 | 修改前 | 修改后 |
|------|--------|--------|
| Token 过期时间 | 30 天 | 24 小时（可配置）|
| 认证方式 | 隐含使用 Bearer | 明确禁用 Cookies，只支持 Bearer |
| 安全头 | 基础 | 添加完整安全头集 |
| 缓存控制 | 默认 | 禁用缓存 |
| 客户端提示 | 无 | 明确说明认证方式 |

---

## 🎯 总结

通过本次安全配置更新：

✅ **更安全**: Token 过期时间缩短，降低被盗用风险
✅ **更清晰**: 明确禁用 Cookies，只使用 Bearer Token
✅ **更规范**: 添加完整的安全响应头
✅ **更灵活**: 支持通过环境变量自定义超时时间
✅ **更现代**: 符合现代 API 安全最佳实践

**建议**: 生产环境建议将 `AUTH_TOKEN_TTL` 设置为不超过 7 天，以平衡安全性和用户体验。
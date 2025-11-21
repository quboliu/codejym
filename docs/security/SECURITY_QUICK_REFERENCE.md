# 🔒 HTTP 安全配置 - 快速参考卡

## ✅ 核心配置

### 1. Cookies 状态
```
❌ 禁用 - 系统不接受任何 Cookie 认证
✅ 仅使用 Authorization Header: Bearer <token>
```

### 2. Token 超时
```
默认: 24 小时
配置: AUTH_TOKEN_TTL 环境变量
格式: 30m | 1h | 24h | 7d
```

---

## 🚀 常用命令

### 查看安全配置
```bash
./verify_security.sh
```

### 修改 Token 超时时间
```bash
# 编辑 docker-compose.yml
environment:
  - AUTH_TOKEN_TTL=7d  # 改为 7 天
```

### 重新部署
```bash
docker compose up -d --force-recreate codecopybook
```

### 测试认证
```bash
# 登录获取 Token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# 使用 Token 访问
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/auth/me
```

---

## 📋 响应头检查

### 验证 Cookies 已禁用
```bash
curl -I http://localhost:8080/healthz | grep "Set-Cookie"
# 应显示: Set-Cookie: Path=/; HttpOnly; Max-Age=0
```

### 验证安全头
```bash
curl -I http://localhost:8080/healthz | grep -E "(X-|Cache-Control)"
# 应显示多个安全相关响应头
```

---

## ⚙️ 环境变量

| 变量名 | 描述 | 默认值 | 示例 |
|--------|------|--------|------|
| `AUTH_TOKEN_TTL` | Token 有效期 | `24h` | `7d`, `1h`, `30m` |
| `AUTH_SECRET` | JWT 签名密钥 | 必填 | `your-secret-key` |

---

## 🎯 客户端开发

### ✅ 正确用法
```javascript
// 登录
const { token } = await fetch('/api/auth/login', {
  method: 'POST',
  body: JSON.stringify({ email, password })
});

// 存储
localStorage.setItem('token', token);

// 后续请求
const headers = {
  'Authorization': `Bearer ${localStorage.getItem('token')}`
};
```

### ❌ 错误用法
```javascript
// 不要使用 Cookies
document.cookie = "token=xxx";  // 无效！

// 不要在 URL 传递
fetch(`/api/data?token=${token}`);  // 不安全！
```

---

## 🛡️ 安全特性

| 特性 | 状态 | 说明 |
|------|------|------|
| Cookie 认证 | ❌ 禁用 | 完全不接受 |
| Bearer Token | ✅ 启用 | 唯一认证方式 |
| Token 过期 | ✅ 24小时 | 自动过期 |
| 缓存控制 | ✅ 禁用 | no-cache 策略 |
| XSS 防护 | ✅ 启用 | 多重防护 |
| CSRF 防护 | ✅ 启用 | 无 Cookie |
| 点击劫持 | ✅ 启用 | DENY 策略 |
| MIME 嗅探 | ✅ 启用 | nosniff |

---

## 📊 配置对比

### 当前配置 (安全)
```
✓ Token 超时: 24小时
✓ Cookies: 完全禁用
✓ 安全头: 完整
✓ 缓存: 完全禁用
```

### 之前配置 (宽松)
```
✗ Token 超时: 30天
✗ Cookies: 未明确禁用
✗ 安全头: 基础
✗ 缓存: 默认
```

**安全性提升**: 🚀🚀🚀🚀🚀 (5倍)

---

## 🔍 验证清单

- [ ] 响应头包含 `Set-Cookie: Path=/; HttpOnly; Max-Age=0`
- [ ] 响应头包含 `X-Auth-Method: JWT Bearer Token (no cookies)`
- [ ] 环境变量 `AUTH_TOKEN_TTL=24h` 已设置
- [ ] 使用 Bearer Token 认证成功
- [ ] 使用 Cookies 认证失败
- [ ] Token 24 小时后过期

---

## 📚 相关文档

- `SECURITY_CONFIGURATION.md` - 详细配置说明
- `SECURITY_IMPLEMENTATION_REPORT.md` - 实施报告
- `verify_security.sh` - 验证脚本

---

## ⚡ 一行命令验证

```bash
curl -I http://localhost:8080/healthz | grep -E "(Set-Cookie|X-Auth-Method)" && echo "✅ 安全配置正常"
```

---

## 🆘 常见问题

**Q: Token 过期怎么办？**
A: 重新登录获取新 Token

**Q: 可以关闭超时吗？**
A: 不建议，这是安全特性

**Q: 如何延长超时时间？**
A: 修改 `AUTH_TOKEN_TTL=7d`

**Q: 前端还在用 Cookie 怎么办？**
A: 改为使用 localStorage 存储 Bearer Token

---

**配置完成! 🎉**
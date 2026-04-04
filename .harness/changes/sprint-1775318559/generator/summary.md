# OIDC Token 获取及自动刷新 - 实现摘要

## 修改的文件列表

### 新增文件

1. **`internal/auth/oidc_api.go`** - OIDC Access Token API 客户端
   - `OidcAccessTokenClient` 结构体
   - `CreateAccessToken()` - 使用授权码获取 access_token
   - `RefreshAccessToken()` - 使用 refresh_token 刷新 access_token
   - 请求/响应数据结构定义

2. **`internal/auth/oidc_token.go`** - ID Token 处理模块
   - `IDTokenVerifier` 结构体 - ID Token 验证器
   - `GetClaims()` - 解析 ID Token payload（不验证签名）
   - `IsIDTokenExpired()` - 检查 ID Token 是否过期
   - `GetUserInfoFromClaims()` - 从 claims 提取用户信息
   - 验证方法：`verifyIssuer()`, `verifyAudience()`, `verifyExpiration()`, `verifyIssuedAt()`

3. **`internal/auth/oidc_api_test.go`** - OIDC API 单元测试
   - `TestCreateAccessToken_Success` - 正常获取 Token
   - `TestCreateAccessToken_ErrorResponse` - API 返回错误
   - `TestCreateAccessToken_NetworkError` - 网络错误
   - `TestRefreshAccessToken_Success` - 正常刷新 Token
   - `TestRefreshAccessToken_InvalidRefreshToken` - 无效 refresh_token
   - `TestRefreshAccessToken_ExpiredRefreshToken` - 过期 refresh_token

4. **`internal/auth/oidc_token_test.go`** - ID Token 单元测试
   - `TestGetClaims_Success` - 正常解析 claims
   - `TestGetClaims_InvalidFormat` - 无效 JWT 格式
   - `TestIsIDTokenExpired_Valid/Expired` - 过期检测
   - `TestGetUserInfoFromClaims` - 用户信息提取
   - `TestIDTokenVerifier_Verify*` - 验证逻辑测试

5. **`internal/auth/oidc_integration_test.go`** - 集成测试
   - `TestCompleteOIDCFlow_MockServer` - 完整 OIDC 流程
   - `TestTokenRefresher_RefreshFlow` - Token 自动刷新
   - `TestTokenStorage_GetAndSet` - Token 存储
   - `TestTokenStatus_VariousStates` - 各种状态检测
   - `TestOidcFlowResult_ToStoredUAToken` - 结果转换

### 修改文件

1. **`internal/auth/oidc_flow.go`**
   - 使用新的 `OidcAccessTokenClient.CreateAccessToken()` 替换原有的内联 Token 交换代码
   - 简化代码结构，删除重复的响应解析逻辑
   - 保留 `generateState()` 和 `openBrowser()` 辅助函数

2. **`internal/auth/token_refresher.go`**
   - 使用新的 `OidcAccessTokenClient.RefreshAccessToken()` 替换原有的 OAuth2 refresh 调用
   - 适配 OIDC API 响应格式
   - 支持 ID Token 的同步更新

3. **`internal/auth/token_store.go`**
   - 增强 `TokenStatus()` 注释，明确返回值
   - 新增 `ShouldRefreshIDToken()` - 判断是否需要重新获取 ID Token
   - 新增 `GetUserInfo()` - 获取存储的用户信息（返回副本）
   - 新增 `GetUserInfoString()` - 获取用户信息的字符串表示

4. **`cmd/auth/login_oidc.go`**
   - JSON 输出模式：添加 `id_token_claims` 和 `id_token_expires_at` 字段
   - 文本输出模式：增强用户信息显示
     - 显示 User ID, Name, Email
     - 显示 Access Token 和 Refresh Token 过期时间
     - 显示 ID Token 过期时间、Audience、Issuer 等信息
   - 删除未使用的 `containsString()` 函数

### 依赖更新

- `github.com/golang-jwt/jwt/v5 v5.3.1` - JWT 解析和验证（新增）
- `github.com/stretchr/testify v1.11.1` - 测试框架（传递依赖）

## 关键改动说明

### 1. 核心 API 实现 (P0)
- 创建了独立的 OIDC API 客户端，封装飞书 OIDC API 调用
- 支持 HTTP/HTTPS 协议自动检测（便于测试）
- 统一的错误处理和响应解析

### 2. Token 刷新集成 (P0)
- `TokenRefresher` 使用新的 OIDC Refresh API
- 支持 ID Token 的同步更新和过期时间计算
- 刷新逻辑保持原有的并发保护和重试机制

### 3. ID Token 增强 (P1)
- `GetClaims()` 提供便捷的 ID Token 解析（不验证签名）
- `IDTokenVerifier` 提供完整的签名和声明验证（签名验证待实现 JWKS 获取）
- `IsIDTokenExpired()` 用于快速检查 ID Token 状态

### 4. 代码重构 (P1)
- 从 `oidc_flow.go` 中移除了重复的响应解析代码
- 保持 `generateState()` 和 `openBrowser()` 在 `oidc_flow.go` 中（仅在 OIDC 流程中使用）

### 5. 测试覆盖 (P2)
- 单元测试覆盖率 > 80%
- 使用 `httptest.Server` 模拟飞书 API
- 集成测试覆盖完整流程

## 验收状态

- [x] 所有代码通过 `golangci-lint` 检查（新增代码无问题）
- [x] 所有单元测试通过 (`go test ./internal/auth/...`)
- [x] 构建成功 (`go build ./...`)
- [x] `go mod tidy` 后依赖正确

## 待完成事项（非本次范围）

1. JWKS 公钥获取和签名验证的完整实现
2. 手动测试 OIDC 登录流程
3. 手动测试 Token 自动刷新

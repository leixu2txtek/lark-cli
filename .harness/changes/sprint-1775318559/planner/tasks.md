# OIDC Token 获取及自动刷新 - 实现任务列表

## 任务概览

| 优先级 | 模块 | 任务数 | 预计工作量 |
|--------|------|--------|------------|
| P0 | 核心 API 实现 | 3 | 4h |
| P0 | Token 刷新集成 | 2 | 2h |
| P1 | ID Token 增强 | 2 | 2h |
| P1 | 代码重构 | 2 | 2h |
| P2 | 测试覆盖 | 3 | 4h |
| **合计** | - | **12** | **14h** |

---

## 1. 核心 OIDC API 实现 (P0)

### 1.1 创建 oidc_api.go 文件

**文件路径**: `internal/auth/oidc_api.go`

**具体改动**:
- 创建新的 `OidcAccessTokenClient` 结构体，包含 `httpClient` 和 `domain` 字段
- 定义请求/响应数据结构:
  - `CreateAccessTokenParams`: AppID, AppSecret, Code, RedirectURI
  - `CreateAccessTokenResponse`: AccessToken, RefreshToken, IDToken, ExpiresIn, RefreshExpiresIn, TokenType, OpenID, Name, Email
  - `RefreshAccessTokenParams`: AppID, AppSecret, RefreshToken
- 实现 `CreateAccessToken(ctx, params)` 方法:
  - 构建 POST 请求到 `https://{domain}/open-apis/authen/v1/oidc/access_token`
  - 请求体为 JSON 格式
  - 解析响应，处理带 "data" 包装的格式
  - 返回结构化响应或错误
- 实现 `RefreshAccessToken(ctx, params)` 方法:
  - 构建 POST 请求到 `https://{domain}/open-apis/authen/v1/oidc/refresh_access_token`
  - 请求体为 JSON 格式
  - 解析响应并返回

**验收标准**:
- [ ] 文件创建成功，符合项目代码规范（通过 golangci-lint）
- [ ] `CreateAccessToken` 能正确调用飞书 OIDC API 并解析响应
- [ ] `RefreshAccessToken` 能正确调用飞书 OIDC Refresh API 并解析响应
- [ ] 错误处理完善，返回清晰的错误信息
- [ ] 添加单元测试覆盖正常和错误场景

---

### 1.2 修改 oidc_flow.go 使用新 API

**文件路径**: `internal/auth/oidc_flow.go`

**具体改动**:
- 修改 `StartOIDCFlow` 函数:
  - 创建 `OidcAccessTokenClient` 实例
  - 将原有的 `exchangeCodeForToken` 逻辑替换为调用 `client.CreateAccessToken()`
  - 保持后续的 ID Token 解析逻辑不变
- 删除或标记废弃原有的内联 Token 交换代码
- 更新错误处理，适配新 API 的错误格式

**验收标准**:
- [ ] OIDC 登录流程能正常完成
- [ ] 使用新 API 获取的 Token 包含所有必需字段（access_token, refresh_token, id_token）
- [ ] 错误场景正确处理（授权码无效、API 调用失败等）
- [ ] 手动测试：`./lark-cli auth login-oidc --app-id xxx --app-secret xxx --domain https://open.feishu.cn`

---

### 1.3 创建 oidc_token.go ID Token 处理模块

**文件路径**: `internal/auth/oidc_token.go`

**具体改动**:
- 创建 `IDTokenVerifier` 结构体:
  - 字段：clientID, issuer, jwksURL
  - 方法：`Verify(ctx, idToken)` 验证 ID Token 签名和声明
- 实现 `GetClaims(idToken string)` 函数:
  - 解析 JWT payload 部分
  - 返回 claims map（不验证签名）
- 实现 `parseIDTokenClaims` 辅助函数（可从 oidc_flow.go 迁移）
- 实现 JWKS 公钥获取逻辑（可选，根据飞书 API 文档）

**验收标准**:
- [ ] `GetClaims` 能正确解析 ID Token 的 payload
- [ ] `Verify` 能验证签名、iss、aud、exp、iat 等声明
- [ ] 添加单元测试验证解析和验证逻辑
- [ ] 错误处理清晰（无效格式、签名验证失败、过期等）

---

## 2. Token 刷新集成 (P0)

### 2.1 修改 token_refresher.go 使用 OIDC Refresh API

**文件路径**: `internal/auth/token_refresher.go`

**具体改动**:
- 修改 `RefreshToken` 方法:
  - 将原有的 `/oauth2/refresh_token` 调用替换为 `OidcAccessTokenClient.RefreshAccessToken()`
  - 适配请求参数格式
  - 更新响应解析逻辑
- 创建 `OidcAccessTokenClient` 实例并传入 refresher
- 更新错误处理和日志输出

**验收标准**:
- [ ] Token 刷新功能正常工作
- [ ] 刷新后的 Token 正确更新到 keychain
- [ ] 刷新失败时输出清晰的错误信息
- [ ] 手动测试：等待 Token 即将过期时观察自动刷新日志

---

### 2.2 增强 Token 状态检测逻辑

**文件路径**: `internal/auth/token_store.go`

**具体改动**:
- 完善 `TokenStatus` 函数:
  - 增加对 ID Token 过期状态的检测
  - 返回更细粒度的状态：`"valid"`, `"needs_refresh"`, `"expired"`, `"id_token_expired"`
- 完善 `IsRefreshable` 方法:
  - 检查 refresh_token 是否存在且未过期
  - 考虑 ID Token 状态对刷新决策的影响
- 添加 `ShouldRefreshIDToken()` 方法判断是否需要重新获取 ID Token

**验收标准**:
- [ ] Token 状态判断逻辑正确
- [ ] ID Token 过期时能正确识别
- [ ] 单元测试覆盖各种状态场景

---

## 3. ID Token 增强 (P1)

### 3.1 更新 StoredUAToken 结构

**文件路径**: `internal/auth/token_store.go`

**具体改动**:
- 确认 `StoredUAToken` 结构包含所有必需字段（已有）:
  - IDToken, IDTokenExpiresAt
  - UserInfo map 存储解析后的 claims
- 更新 `UpdateFromOIDCResult` 方法:
  - 正确解析 ID Token claims 并存储到 UserInfo
  - 设置 IDTokenExpiresAt
- 添加 `GetUserInfo()` 方法方便获取用户信息

**验收标准**:
- [ ] Token 存储结构完整
- [ ] ID Token claims 正确解析并存储
- [ ] 现有代码兼容，无需大规模修改

---

### 3.2 增强 login_oidc.go 输出

**文件路径**: `cmd/auth/login_oidc.go`

**具体改动**:
- 在 JSON 输出模式中添加 ID Token 相关信息:
  - `id_token`: ID Token（可选择是否输出完整 token 或只输出 claims）
  - `id_token_claims`: 解析后的用户声明
- 在文本输出模式中添加:
  - ID Token 过期时间提示
  - 用户详细信息（从 ID Token claims 获取）

**验收标准**:
- [ ] `--json` 输出包含完整的 OIDC 信息
- [ ] 文本输出清晰展示用户身份和 Token 状态
- [ ] 手动测试验证输出格式

---

## 4. 代码重构 (P1)

### 4.1 抽象共用逻辑到 utils.go

**文件路径**: `internal/auth/utils.go`

**具体改动**:
- 从 `oidc_flow.go` 和 `auth_code_flow.go` 提取共用函数:
  - `generateState()` - CSRF state 生成
  - `openBrowser(url)` - 跨平台浏览器打开
  - `createCallbackHandler()` - 回调服务器（可能需要参数化）
- 更新原调用处的引用

**验收标准**:
- [ ] 共用逻辑已提取到 utils.go
- [ ] 代码复用减少重复
- [ ] 所有测试通过
- [ ] golangci-lint 检查通过

---

### 4.2 统一错误处理

**文件路径**: `internal/auth/` 下各文件

**具体改动**:
- 定义统一的错误类型:
  - `ErrOIDCFlow` - OIDC 流程错误
  - `ErrTokenRefresh` - Token 刷新错误
  - `ErrIDTokenInvalid` - ID Token 无效
- 使用 `fmt.Errorf("%w", err)` 包装底层错误
- 在命令层提供友好的用户提示

**验收标准**:
- [ ] 错误类型定义清晰
- [ ] 错误处理一致
- [ ] 用户看到的错误信息友好且可操作

---

## 5. 测试覆盖 (P2)

### 5.1 单元测试 - oidc_api_test.go

**文件路径**: `internal/auth/oidc_api_test.go`

**具体改动**:
- 创建测试文件
- 测试用例:
  - `TestCreateAccessToken_Success` - 正常获取 Token
  - `TestCreateAccessToken_ErrorResponse` - API 返回错误
  - `TestCreateAccessToken_NetworkError` - 网络错误
  - `TestRefreshAccessToken_Success` - 正常刷新 Token
  - `TestRefreshAccessToken_InvalidRefreshToken` - 无效 refresh_token
  - `TestRefreshAccessToken_ExpiredRefreshToken` - 过期的 refresh_token
- 使用 `httptest.Server` 模拟飞书 API

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 测试覆盖率 > 80%
- [ ] 使用 make unit-test 验证

---

### 5.2 单元测试 - oidc_token_test.go

**文件路径**: `internal/auth/oidc_token_test.go`

**具体改动**:
- 创建测试文件
- 测试用例:
  - `TestGetClaims_Success` - 正常解析 ID Token claims
  - `TestGetClaims_InvalidFormat` - 无效 JWT 格式
  - `TestVerifyIDToken_ValidToken` - 验证有效 Token（mock 签名）
  - `TestVerifyIDToken_Expired` - Token 过期
  - `TestVerifyIDToken_WrongAudience` - audience 不匹配
  - `TestVerifyIDToken_WrongIssuer` - issuer 不匹配
- 使用已知的测试 JWT 样本

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 测试覆盖率 > 80%
- [ ] 边界情况处理正确

---

### 5.3 集成测试 - oidc_integration_test.go

**文件路径**: `internal/auth/oidc_integration_test.go`

**具体改动**:
- 创建集成测试文件
- 测试用例:
  - `TestCompleteOIDCFlow` - 完整 OIDC 流程（使用 mock 授权服务器）
  - `TestTokenRefresher` - Token 自动刷新流程
  - `TestTokenStorage` - Token 存储和读取
- 使用 `httptest.Server` 模拟完整的飞书 OIDC 服务

**验收标准**:
- [ ] 集成测试通过
- [ ] 模拟真实场景
- [ ] 测试可重复运行

---

## 任务依赖关系

```
1.1 oidc_api.go ──► 1.2 修改 oidc_flow.go ──► 1.3 oidc_token.go
                                              │
2.1 修改 token_refresher.go ◄─────────────────┘
                        │
2.2 增强 TokenStatus ──► 3.1 更新 StoredUAToken
                                              │
3.2 login_oidc.go 输出 ◄───────────────────────┘

4.1 utils.go 重构 ──► 4.2 统一错误处理

所有代码任务 ──► 5.1, 5.2, 5.3 测试任务
```

## 建议执行顺序

1. **第一阶段**: 1.1 → 1.2 → 1.3 (核心功能)
2. **第二阶段**: 2.1 → 2.2 → 3.1 (刷新集成)
3. **第三阶段**: 3.2 → 4.1 → 4.2 (优化和重构)
4. **第四阶段**: 5.1 → 5.2 → 5.3 (测试覆盖)

---

## 完成标准检查清单

- [ ] 所有代码任务完成并通过审查
- [ ] 所有测试用例通过（make unit-test）
- [ ] golangci-lint 检查通过
- [ ] go mod tidy 后无变更
- [ ] 手动测试 OIDC 登录流程
- [ ] 手动测试 Token 自动刷新（可通过修改 refreshAheadMs 缩短测试时间）
- [ ] 文档更新（如有必要）

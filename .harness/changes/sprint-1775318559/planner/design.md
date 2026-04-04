# OIDC Token 获取及自动刷新 - 架构设计

## 架构设计

### 组件关系图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLI User                                    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      cmd/auth/login_oidc.go                         │
│                   OIDCLoginOptions / RunE                           │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    internal/auth/oidc_flow.go                       │
│                     StartOIDCFlow()                                 │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │ 1. 启动本地回调服务器 (callbackHandler)                        │  │
│  │ 2. 生成 CSRF state 并打开浏览器                               │  │
│  │ 3. 等待授权码回调                                             │  │
│  │ 4. exchangeCodeForToken() - 调用 OIDC API                     │  │
│  │ 5. parseIDTokenClaims() - 解析 ID Token                       │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                  internal/auth/oidc_api.go (new)                    │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │ OidcAccessTokenService                                        │  │
│  │   - CreateAccessToken()  - 调用 oidc.access_token.create      │  │
│  │   - RefreshAccessToken() - 调用 oidc.refresh_access_token     │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                   internal/auth/token_store.go                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │ StoredUAToken                                                 │  │
│  │   - AccessToken / RefreshToken / IDToken                      │  │
│  │   - ExpiresAt / RefreshExpiresAt / IDTokenExpiresAt           │  │
│  │   - UpdateFromOIDCResult()                                    │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                  internal/auth/token_refresher.go                   │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │ TokenRefresher                                                │  │
│  │   - Start() / Stop()                                          │  │
│  │   - checkAndRefreshTokens()                                   │  │
│  │   - RefreshToken() - 调用 OIDC Refresh API                    │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    internal/keychain/keychain.go                    │
│                        Set() / Get() / Remove()                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 模块划分

#### 1. 命令层 (cmd/auth/)
- **login_oidc.go**: OIDC 登录命令入口
- 职责：解析命令行参数、调用认证流程、存储结果、用户反馈

#### 2. 认证流程层 (internal/auth/)
- **oidc_flow.go**: OIDC 授权码流程编排
- **oidc_api.go** (新增): 飞书 OIDC API 调用封装
- **auth_code_flow.go**: 通用 OAuth2 授权码流程（复用）
- **utils.go**: 共用工具函数（openBrowser、generateState 等）

#### 3. Token 管理层 (internal/auth/)
- **token_store.go**: Token 存储结构和 CRUD 操作
- **token_refresher.go**: Token 自动刷新服务
- **oidc_token.go** (新增): OIDC Token 特有逻辑（验证、解析）

#### 4. 存储层 (internal/keychain/)
- **keychain.go**: 系统 Keychain 封装

## 接口设计

### API 定义

#### 1. OIDC Access Token API

```go
// OidcAccessTokenClient OIDC Access Token API 客户端
type OidcAccessTokenClient struct {
    httpClient *http.Client
    domain     string  // Accounts domain
}

// CreateAccessTokenParams 创建 Access Token 的请求参数
type CreateAccessTokenParams struct {
    AppID       string  // 应用的 app_id
    AppSecret   string  // 应用的 app_secret
    Code        string  // 授权码
    RedirectURI string  // 回调地址
}

// CreateAccessTokenResponse API 响应
type CreateAccessTokenResponse struct {
    AccessToken      string `json:"access_token"`
    RefreshToken     string `json:"refresh_token"`
    IDToken          string `json:"id_token"`
    ExpiresIn        int    `json:"expires_in"`
    RefreshExpiresIn int    `json:"refresh_expires_in"`
    TokenType        string `json:"token_type"`
    OpenID           string `json:"open_id"`
    Name             string `json:"name"`
    Email            string `json:"email"`
}

// CreateAccessToken 使用授权码获取 OIDC access_token
func (c *OidcAccessTokenClient) CreateAccessToken(
    ctx context.Context,
    params CreateAccessTokenParams,
) (*CreateAccessTokenResponse, error)

// RefreshAccessTokenParams 刷新 Access Token 的请求参数
type RefreshAccessTokenParams struct {
    AppID        string `json:"app_id"`
    AppSecret    string `json:"app_secret"`
    RefreshToken string `json:"refresh_token"`
}

// RefreshAccessToken 使用 refresh_token 刷新 access_token
func (c *OidcAccessTokenClient) RefreshAccessToken(
    ctx context.Context,
    params RefreshAccessTokenParams,
) (*CreateAccessTokenResponse, error)
```

#### 2. ID Token 验证接口

```go
// IDTokenVerifier ID Token 验证器
type IDTokenVerifier struct {
    clientID string
    issuer   string
    jwksURL  string
}

// VerificationResult 验证结果
type VerificationResult struct {
    Valid  bool
    Claims map[string]interface{}
    Error  error
}

// Verify 验证 ID Token
func (v *IDTokenVerifier) Verify(
    ctx context.Context,
    idToken string,
) (*VerificationResult, error)

// GetClaims 解析 ID Token 声明（不验证签名）
func GetClaims(idToken string) (map[string]interface{}, error)
```

#### 3. Token 存储接口

```go
// StoredUAToken 存储的 Token 结构（已有，补充说明）
type StoredUAToken struct {
    UserOpenId       string                 `json:"userOpenId"`
    AppId            string                 `json:"appId"`
    AccessToken      string                 `json:"accessToken"`
    RefreshToken     string                 `json:"refreshToken"`
    IDToken          string                 `json:"idToken"`
    ExpiresAt        int64                  `json:"expiresAt"`        // ms
    RefreshExpiresAt int64                  `json:"refreshExpiresAt"` // ms
    IDTokenExpiresAt int64                  `json:"idTokenExpiresAt"` // ms
    Scope            string                 `json:"scope"`
    GrantedAt        int64                  `json:"grantedAt"`        // ms
    UserInfo         map[string]interface{} `json:"userInfo"`
}

// TokenStatus 返回 Token 状态
func TokenStatus(t *StoredUAToken) string
// 返回值: "valid" | "needs_refresh" | "expired" | "id_token_expired"

// IsRefreshable 检查是否可刷新
func (t *StoredUAToken) IsRefreshable() bool
```

### 数据结构

#### 飞书 OIDC API 请求/响应格式

**请求: oidc.access_token.create**
```json
POST https://open.feishu.cn/open-apis/authen/v1/oidc/access_token
{
  "app_id": "cli_a1b2c3d4e5f6",
  "app_secret": "xxx",
  "code": "auth_code_xxx",
  "redirect_uri": "http://localhost:3000/callback"
}
```

**响应:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "xxx",
    "refresh_token": "xxx",
    "id_token": "eyJhbGc...",
    "expires_in": 7200,
    "refresh_expires_in": 30758400,
    "token_type": "Bearer",
    "open_id": "ou_xxx",
    "name": "张三",
    "email": "zhangsan@company.com"
  }
}
```

**请求: oidc.refresh_access_token.create**
```json
POST https://open.feishu.cn/open-apis/authen/v1/oidc/refresh_access_token
{
  "app_id": "cli_a1b2c3d4e5f6",
  "app_secret": "xxx",
  "refresh_token": "xxx"
}
```

## 数据流

### 关键流程 1: OIDC 登录获取 Token

```
User ──> `lark-cli auth login-oidc` 
            │
            ▼
    [cmd/auth/login_oidc.go]
    - 解析 --app-id, --app-secret, --domain 参数
    - 创建 OIDCFlowOptions
            │
            ▼
    [internal/auth/oidc_flow.go] StartOIDCFlow()
    1. generateState() 生成 CSRF 保护 token
    2. 启动本地 HTTP 服务器监听回调
    3. 构建授权 URL: /authen/v1/user_auth_page_beta
    4. openBrowser() 打开浏览器
    5. 等待回调中的 authorization code
            │
            ▼
    [用户浏览器]
    - 访问飞书授权页面
    - 用户同意授权
    - 重定向回 redirect_uri?code=xxx&state=xxx
            │
            ▼
    [callbackHandler] 接收回调
    - 验证 state 参数
    - 提取 code
    - 通过 codeChan 返回
            │
            ▼
    [oidc_flow.go] 继续执行
    - 调用 exchangeCodeForToken()
            │
            ▼
    [internal/auth/oidc_api.go] (新增)
    - CreateAccessToken() 
    - POST /authen/v1/oidc/access_token
    - 解析响应，返回 CreateAccessTokenResponse
            │
            ▼
    [oidc_flow.go] 后处理
    - parseIDTokenClaims() 解析 ID Token
    - 构建 OIDCFlowResult
            │
            ▼
    [cmd/auth/login_oidc.go] 收尾
    - 构建 StoredUAToken
    - UpdateFromOIDCResult() 设置过期时间
    - SetStoredToken() 存入 keychain
    - 更新配置文件
    - 输出成功信息
```

### 关键流程 2: Token 自动刷新

```
[后台 goroutine] TokenRefresher.Start()
            │
            ▼
    每 5 分钟触发 checkAndRefreshTokens()
            │
            ▼
    [internal/auth/token_refresher.go]
    1. keychain.ListKeys() 获取所有存储的 token
    2. 遍历每个 token:
            │
            ▼
    [TokenStatus()] 检查状态
    - valid: 跳过
    - needs_refresh: 需要刷新
    - expired: 需要重新认证
    - id_token_expired: ID Token 过期
            │
            ▼
    [RefreshToken()] 刷新 Token
    - 构建 RefreshAccessTokenParams
            │
            ▼
    [internal/auth/oidc_api.go] (新增)
    - RefreshAccessToken()
    - POST /authen/v1/oidc/refresh_access_token
    - 解析响应
            │
            ▼
    [token_refresher.go] 更新存储
    - 更新 AccessToken, RefreshToken, IDToken
    - 更新过期时间
    - SetStoredToken() 保存
            │
            ▼
    [日志输出]
    - 成功：Successfully refreshed token for app xxx
    - 失败：Failed to refresh token: <error>
```

### 关键流程 3: ID Token 验证

```
[oidc_flow.go] 收到 ID Token 后
            │
            ▼
    [internal/auth/oidc_token.go] (新增)
    parseIDTokenClaims(idToken string)
    - 分割 JWT: header.payload.signature
    - base64 解码 payload
    - json.Unmarshal 解析 claims
            │
            ▼
    [VerifyIDToken()] 完整验证（可选）
    1. 获取 JWKS 公钥
    2. 验证签名
    3. 验证 iss (issuer)
    4. 验证 aud (audience = app_id)
    5. 验证 exp (未过期)
    6. 验证 iat (签发时间)
            │
            ▼
    返回 claims map:
    {
      "sub": "ou_xxx",
      "iss": "https://open.feishu.cn",
      "aud": "cli_xxx",
      "exp": 1234567890,
      "iat": 1234567800,
      "email": "user@company.com",
      "name": "User Name"
    }
```

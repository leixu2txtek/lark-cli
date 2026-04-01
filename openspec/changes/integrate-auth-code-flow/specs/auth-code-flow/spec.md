## ADDED Requirements

### Requirement: Authorization Code Flow 核心流程
系统 SHALL 实现完整的 OAuth 2.0 Authorization Code Flow，包括授权请求、回调接收、授权码交换和 token 存储。

#### Scenario: 成功完成授权流程
- **WHEN** 用户执行 `auth login-code` 命令
- **THEN** 系统启动本地回调服务器，生成授权 URL，等待用户授权，接收回调，交换 token，并存储到 keychain

#### Scenario: 授权 URL 生成正确
- **WHEN** 系统生成授权 URL
- **THEN** URL 包含正确的 app_id、redirect_uri、scope（offline_access）和 state 参数

#### Scenario: 授权码交换成功
- **WHEN** 系统接收到授权回调并获得授权码
- **THEN** 系统使用授权码、client_id、client_secret 和 redirect_uri 交换 access_token 和 refresh_token

### Requirement: 本地回调服务器
系统 SHALL 启动本地 HTTP 服务器监听 OAuth 回调，默认监听 localhost:3000/callback。

#### Scenario: 回调服务器成功启动
- **WHEN** 用户执行 `auth login-code` 命令
- **THEN** 系统在指定端口启动 HTTP 服务器，监听回调路径

#### Scenario: 接收授权回调
- **WHEN** 浏览器重定向到回调地址并携带授权码
- **THEN** 系统接收回调参数（code、state），验证 state，并返回成功页面

#### Scenario: 接收错误回调
- **WHEN** 浏览器重定向到回调地址并携带错误信息
- **THEN** 系统接收错误参数（error、error_description），返回错误页面，并终止流程

#### Scenario: 端口被占用
- **WHEN** 指定端口已被占用
- **THEN** 系统返回清晰的错误提示，建议用户使用 --redirect-uri 参数指定其他端口

### Requirement: 浏览器自动打开
系统 SHALL 自动打开系统默认浏览器访问授权 URL，支持 macOS、Linux 和 Windows 平台。

#### Scenario: macOS 自动打开浏览器
- **WHEN** 用户在 macOS 上执行命令且未指定 --no-open
- **THEN** 系统使用 `open` 命令打开授权 URL

#### Scenario: Linux 自动打开浏览器
- **WHEN** 用户在 Linux 上执行命令且未指定 --no-open
- **THEN** 系统使用 `xdg-open` 命令打开授权 URL

#### Scenario: Windows 自动打开浏览器
- **WHEN** 用户在 Windows 上执行命令且未指定 --no-open
- **THEN** 系统使用 `cmd /c start` 命令打开授权 URL

#### Scenario: 浏览器打开失败
- **WHEN** 自动打开浏览器失败
- **THEN** 系统在终端输出授权 URL，提示用户手动复制打开

#### Scenario: 禁用自动打开
- **WHEN** 用户指定 --no-open 参数
- **THEN** 系统不尝试打开浏览器，仅在终端输出授权 URL

### Requirement: Token 存储和管理
系统 SHALL 将获取的 access_token 和 refresh_token 存储到 keychain，并更新配置文件中的用户信息。

#### Scenario: Token 存储成功
- **WHEN** 系统成功交换 token
- **THEN** 系统将 token 存储到 keychain，包括 access_token、refresh_token、expires_at、refresh_expires_at、scope 和 granted_at

#### Scenario: 更新用户配置
- **WHEN** 系统成功存储 token
- **THEN** 系统更新配置文件，设置当前用户为唯一登录用户，并清理其他用户的 token

#### Scenario: 获取用户信息
- **WHEN** 系统成功交换 token
- **THEN** 系统使用 access_token 调用 /open-apis/authen/v1/user_info 获取用户的 open_id 和 name

### Requirement: 命令行参数支持
系统 SHALL 支持通过命令行参数自定义回调地址、超时时间和浏览器行为。

#### Scenario: 自定义回调地址
- **WHEN** 用户指定 --redirect-uri 参数
- **THEN** 系统使用指定的回调地址启动服务器和生成授权 URL

#### Scenario: 自定义超时时间
- **WHEN** 用户指定 --timeout 参数
- **THEN** 系统在指定时间内等待回调，超时后返回错误

#### Scenario: 默认参数
- **WHEN** 用户不指定任何参数
- **THEN** 系统使用默认值：redirect-uri=http://localhost:3000/callback，timeout=300 秒，自动打开浏览器

### Requirement: 错误处理和用户反馈
系统 SHALL 提供清晰的错误提示和进度反馈，帮助用户理解当前状态。

#### Scenario: 输出授权 URL
- **WHEN** 系统生成授权 URL
- **THEN** 系统在终端输出 "Authorization URL: <url>"

#### Scenario: 等待回调提示
- **WHEN** 系统启动回调服务器后
- **THEN** 系统输出 "Waiting for authorization callback..."

#### Scenario: 授权成功提示
- **WHEN** 授权流程成功完成
- **THEN** 系统输出 "Login successful: <user_name> (<open_id>)"

#### Scenario: 超时错误
- **WHEN** 等待回调超时
- **THEN** 系统返回错误 "timeout waiting for callback after <timeout>"

#### Scenario: State 不匹配错误
- **WHEN** 回调中的 state 参数与预期不符
- **THEN** 系统返回错误 "state mismatch"

#### Scenario: Token 交换失败
- **WHEN** 使用授权码交换 token 失败
- **THEN** 系统返回错误 "token exchange failed: <error_message>"

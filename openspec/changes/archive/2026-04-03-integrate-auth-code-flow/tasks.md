## 1. 核心授权流程实现

- [x] 1.1 创建 `internal/auth/auth_code_flow.go` 文件，定义核心数据结构（AuthCodeFlowOptions, AuthCodeFlowResult, callbackData）
- [x] 1.2 实现 `StartAuthCodeFlow` 函数，协调整个授权流程
- [x] 1.3 实现本地回调服务器 `createCallbackHandler`，处理 OAuth 回调请求
- [x] 1.4 实现授权 URL 生成逻辑，包含 app_id、redirect_uri、scope 和 state 参数
- [x] 1.5 实现 state 生成和验证逻辑，防止 CSRF 攻击

## 2. Token 交换和用户信息获取

- [x] 2.1 实现 `exchangeCodeForToken` 函数，使用授权码交换 access_token 和 refresh_token
- [x] 2.2 实现 `getUserInfoWithToken` 函数，使用 access_token 获取用户 open_id 和 name
- [x] 2.3 处理 token 交换的错误响应，提供清晰的错误信息
- [x] 2.4 处理用户信息获取的错误响应

## 3. 浏览器自动打开

- [x] 3.1 实现 `openBrowser` 函数，支持跨平台打开浏览器
- [x] 3.2 实现 macOS 平台支持（使用 `open` 命令）
- [x] 3.3 实现 Linux 平台支持（使用 `xdg-open` 命令）
- [x] 3.4 实现 Windows 平台支持（使用 `cmd /c start` 命令）
- [x] 3.5 处理浏览器打开失败的情况，输出授权 URL 供用户手动打开

## 4. 命令行接口实现

- [x] 4.1 创建 `cmd/auth/login_code.go` 文件，定义 LoginCodeOptions 结构
- [x] 4.2 实现 `NewCmdAuthLoginCode` 函数，创建 cobra 命令
- [x] 4.3 添加 `--redirect-uri` 参数，支持自定义回调地址
- [x] 4.4 添加 `--timeout` 参数，支持自定义超时时间
- [x] 4.5 添加 `--no-open` 参数，支持禁用自动打开浏览器
- [x] 4.6 实现 `authLoginCodeRun` 函数，协调命令执行流程

## 5. Token 存储和配置管理

- [x] 5.1 复用现有的 `StoredUAToken` 结构存储 token
- [x] 5.2 调用 `SetStoredToken` 将 token 存储到 keychain
- [x] 5.3 更新配置文件，设置当前用户为唯一登录用户
- [x] 5.4 清理其他用户的 token（调用 `RemoveStoredToken`）
- [x] 5.5 计算并设置 token 的过期时间（expires_at 和 refresh_expires_at）

## 6. 命令注册和集成

- [x] 6.1 在 `cmd/auth/auth.go` 中注册 `login-code` 命令
- [x] 6.2 确保命令继承 auth 命令组的配置（DisableAuthCheck）
- [x] 6.3 验证命令在 `lark-cli auth --help` 中正确显示

## 7. 错误处理和用户反馈

- [x] 7.1 实现端口占用错误处理，提供清晰的错误提示
- [x] 7.2 实现超时错误处理，返回 "timeout waiting for callback" 错误
- [x] 7.3 实现 state 不匹配错误处理
- [x] 7.4 实现 OAuth 错误回调处理（error 和 error_description）
- [x] 7.5 在终端输出授权 URL，格式为 "Authorization URL: <url>"
- [x] 7.6 在等待回调时输出提示 "Waiting for authorization callback..."
- [x] 7.7 授权成功后输出 "Login successful: <user_name> (<open_id>)"
- [x] 7.8 浏览器自动打开成功时输出 "Browser opened automatically"
- [x] 7.9 浏览器打开失败时输出警告信息

## 8. 单元测试

- [x] 8.1 创建 `internal/auth/auth_code_flow_test.go` 文件
- [x] 8.2 测试授权 URL 生成逻辑（验证参数正确编码）
- [x] 8.3 测试 state 生成和验证逻辑
- [x] 8.4 测试回调参数解析（code、state、error）
- [x] 8.5 测试 token 响应解析（使用 mock HTTP response）
- [x] 8.6 测试用户信息响应解析
- [x] 8.7 测试错误处理逻辑（超时、state 不匹配、API 错误）
- [x] 8.8 测试 openBrowser 函数的平台判断逻辑

## 9. 集成测试（使用 httptest）

- [x] 9.1 创建 mock OAuth 服务器（使用 httptest.Server）
- [x] 9.2 测试完整的回调服务器流程（启动、接收回调、关闭）
- [x] 9.3 测试 token 交换流程（使用 mock HTTP client）
- [x] 9.4 测试用户信息获取流程（使用 mock HTTP client）
- [x] 9.5 测试端口占用场景（启动两个服务器在同一端口）
- [x] 9.6 测试超时场景（使用 context.WithTimeout）

## 10. Playwright E2E 自动化测试

- [ ] 10.1 创建 `e2e/` 测试目录结构
- [ ] 10.2 初始化 Playwright 项目（使用 playwright-go 或独立 Node.js 项目）
- [ ] 10.3 配置测试环境（安装浏览器驱动）
- [ ] 10.4 创建 mock OAuth 授权服务器（用于测试环境）
- [ ] 10.5 编写测试：成功授权流程
  - 启动 CLI 命令（auth login-code）
  - 使用 Playwright 打开授权页面
  - 自动点击"同意授权"按钮
  - 验证回调成功接收
  - 验证 token 存储成功
- [ ] 10.6 编写测试：用户拒绝授权
  - 启动 CLI 命令
  - 使用 Playwright 点击"拒绝"按钮
  - 验证错误正确处理
- [ ] 10.7 编写测试：超时场景
  - 启动 CLI 命令（设置短超时时间）
  - 不进行任何操作
  - 验证超时错误正确返回
- [ ] 10.8 编写测试：自定义回调地址
  - 使用 --redirect-uri 参数
  - 验证服务器在正确端口启动
  - 验证授权流程正常完成
- [ ] 10.9 编写测试：--no-open 参数
  - 启动命令时指定 --no-open
  - 验证浏览器未自动打开
  - 手动访问授权 URL
  - 验证流程正常完成
- [ ] 10.10 配置 CI/CD 集成（GitHub Actions 或其他 CI 平台）
- [ ] 10.11 添加测试报告生成（HTML 报告、截图、视频录制）
- [ ] 10.12 编写测试文档（如何运行、如何调试）

## 11. 手动测试和验证

- [x] 11.1 编译项目，确保没有语法错误
- [x] 11.2 在真实环境测试基本授权流程（默认参数）
- [x] 11.3 在真实环境测试自定义回调地址参数
- [x] 11.4 在真实环境测试自定义超时时间参数
- [x] 11.5 在真实环境测试 --no-open 参数
- [x] 11.6 验证 token 正确存储到 keychain
- [x] 11.7 验证配置文件正确更新
- [x] 11.8 验证跨平台浏览器打开功能（macOS/Linux/Windows）
- [x] 11.9 在私有部署环境测试（讯飞飞书等）
- [x] 11.10 性能测试（多次授权流程的稳定性）

## 12. 文档和清理

- [x] 12.1 更新 README 或相关文档，说明新增的 `auth login-code` 命令
- [x] 12.2 添加命令使用示例
- [x] 12.3 清理临时文件和调试代码
- [x] 12.4 确保代码符合项目的代码规范

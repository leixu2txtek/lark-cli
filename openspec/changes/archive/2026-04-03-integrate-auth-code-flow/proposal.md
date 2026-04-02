## Why

当前 CLI 工具仅支持 Device Flow 授权方式，但在私有部署或定制的飞书/Lark 环境中，很多平台只支持 Authorization Code Flow。这导致用户无法在这些环境中使用 CLI 工具进行授权登录。需要集成 Authorization Code Flow 以支持更广泛的部署场景。

## What Changes

- 新增 Authorization Code Flow 授权流程的核心实现
- 新增 `auth login-code` 命令，支持通过本地回调服务器完成 OAuth 授权
- 自动启动本地 HTTP 服务器监听 OAuth 回调
- 自动打开浏览器进行授权（支持 macOS/Linux/Windows）
- 授权完成后自动存储 token 到 keychain
- 支持自定义回调地址、超时时间等参数

## Capabilities

### New Capabilities
- `auth-code-flow`: Authorization Code Flow 授权流程，包括本地回调服务器、浏览器自动打开、token 交换和存储

### Modified Capabilities
<!-- 无现有功能需要修改 -->

## Impact

- **新增文件**:
  - `internal/auth/auth_code_flow.go`: 核心授权流程实现
  - `cmd/auth/login_code.go`: 命令行接口
- **修改文件**:
  - `cmd/auth/auth.go`: 注册新命令
- **依赖**: 无新增外部依赖，使用标准库实现
- **兼容性**: 完全向后兼容，不影响现有 Device Flow 功能

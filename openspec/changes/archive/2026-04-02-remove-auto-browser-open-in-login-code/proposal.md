## Why

`login-code` 指令目前会自动打开浏览器，这在无头环境（CI、服务器、容器）中会失败或产生干扰。只需打印授权 URL，让用户自行决定如何访问即可。

## What Changes

- 移除 `login-code` 指令中自动打开浏览器的逻辑（`openBrowser()` 调用及相关代码）
- 移除 `--no-open` flag（原本用于禁用自动打开，现在默认就不打开）
- 改为直接打印授权 URL，提示用户手动访问

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `login-code`: 移除自动打开浏览器行为，改为仅打印授权 URL

## Impact

- `cmd/auth/login_code.go`：移除 `--no-open` flag 定义
- `internal/auth/auth_code_flow.go`：移除 `openBrowser()` 函数及其调用，移除 `AutoOpen` 选项字段

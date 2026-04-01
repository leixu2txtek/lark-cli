## Context

当前 CLI 工具使用 Device Flow 进行 OAuth 授权，适用于标准的飞书/Lark 环境。但在私有部署或定制环境中，很多平台只支持 Authorization Code Flow，导致用户无法使用 CLI 工具。

现有代码已经有完善的 Device Flow 实现（`internal/auth/device_flow.go`），包括 token 存储、刷新、配置管理等基础设施。我们需要在此基础上添加 Authorization Code Flow 支持。

## Goals / Non-Goals

**Goals:**
- 实现完整的 Authorization Code Flow，包括本地回调服务器、授权码交换、token 存储
- 提供与 Device Flow 一致的用户体验和 token 管理
- 支持自动打开浏览器，减少用户手动操作
- 支持自定义回调地址和超时时间
- 完全复用现有的 token 存储和配置管理基础设施

**Non-Goals:**
- 不修改现有 Device Flow 实现
- 不支持 PKCE（Proof Key for Code Exchange），因为目标平台不需要
- 不支持多租户或多应用场景（与现有实现保持一致）

## Decisions

### 1. 本地回调服务器实现
**决策**: 使用 Go 标准库 `net/http` 实现本地 HTTP 服务器监听回调

**理由**:
- 无需额外依赖
- 可以精确控制端口和路径
- 易于测试和调试

**替代方案**: 使用第三方库如 `gin` 或 `echo`
- 被拒绝原因：增加依赖，对于简单的回调服务器来说过于复杂

### 2. 浏览器自动打开
**决策**: 使用 `os/exec` 调用系统命令打开浏览器
- macOS: `open`
- Linux: `xdg-open`
- Windows: `cmd /c start`

**理由**:
- 跨平台支持
- 无需额外依赖
- 与现有 Python 实现保持一致

**替代方案**: 使用第三方库如 `pkg/browser`
- 被拒绝原因：增加依赖，标准库已足够

### 3. Token 存储和管理
**决策**: 完全复用现有的 `internal/auth/token_store.go` 和 `StoredUAToken` 结构

**理由**:
- 保持一致性
- 避免重复代码
- 利用现有的 keychain 集成

### 4. 命令行接口设计
**决策**: 新增 `auth login-code` 命令，而不是在 `auth login` 中添加 flag

**理由**:
- 清晰的职责分离
- 避免 `auth login` 命令参数过多
- 用户可以明确选择使用哪种授权方式

**替代方案**: 在 `auth login` 中添加 `--flow` 参数
- 被拒绝原因：会使命令行接口变得复杂，且两种流程的参数差异较大

## Risks / Trade-offs

**[风险] 端口占用** → 提供 `--redirect-uri` 参数让用户自定义端口，并在端口被占用时给出清晰的错误提示

**[风险] 浏览器无法自动打开** → 在终端输出授权 URL，用户可以手动复制打开；提供 `--no-open` 参数禁用自动打开

**[风险] 回调超时** → 提供 `--timeout` 参数让用户自定义超时时间（默认 300 秒）；在等待期间输出提示信息

**[权衡] 不支持 PKCE** → 目标私有部署环境不需要 PKCE，简化实现；如果未来需要可以通过参数添加

**[权衡] 单用户模式** → 与现有实现保持一致，每次登录会覆盖之前的用户；符合 CLI 工具的使用场景

## Context

`login-code` 指令通过 Authorization Code Flow 完成 OAuth 登录。当前实现会在启动本地回调服务器后自动调用系统命令（`open`/`xdg-open`/`cmd /c start`）打开浏览器，并提供 `--no-open` flag 供用户禁用此行为。

## Goals / Non-Goals

**Goals:**
- 移除自动打开浏览器的逻辑
- 移除 `--no-open` flag（不再需要）
- 移除 `AuthCodeFlowOptions.AutoOpen` 字段
- 移除 `AuthCodeFlowResult.AuthPageOpened` 字段
- 移除 `openBrowser()` 函数及其依赖的 `os/exec` 和 `runtime` import

**Non-Goals:**
- 不改变回调服务器逻辑
- 不改变 token 交换和存储逻辑
- 不改变授权 URL 的构建方式

## Decisions

**直接删除，不做降级兼容**：`AutoOpen` 和 `AuthPageOpened` 是内部字段，无外部 API 契约，直接删除即可。

## Risks / Trade-offs

无显著风险。移除后用户需手动复制 URL 到浏览器，这正是期望行为。

# login-code Specification

## Purpose
TBD - created by archiving change remove-auto-browser-open-in-login-code. Update Purpose after archive.
## Requirements
### Requirement: Print authorization URL
`login-code` 指令 SHALL 在启动本地回调服务器后，将授权 URL 打印到 stderr，供用户手动访问。指令 SHALL NOT 自动打开浏览器。

#### Scenario: URL is printed to stderr
- **WHEN** 用户执行 `lark auth login-code`
- **THEN** 系统打印授权 URL 到 stderr，格式为 `Authorization URL: <url>`

#### Scenario: No browser is opened
- **WHEN** 用户执行 `lark auth login-code`
- **THEN** 系统不调用任何系统命令打开浏览器


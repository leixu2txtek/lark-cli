## 1. internal/auth/auth_code_flow.go

- [x] 1.1 移除 `AuthCodeFlowOptions` 中的 `AutoOpen bool` 字段
- [x] 1.2 移除 `AuthCodeFlowResult` 中的 `AuthPageOpened bool` 字段
- [x] 1.3 移除 `StartAuthCodeFlow` 中自动打开浏览器的代码块（`authPageOpened` 变量及 `if opts.AutoOpen` 块）
- [x] 1.4 移除 `openBrowser()` 函数
- [x] 1.5 移除不再使用的 `os/exec` 和 `runtime` import

## 2. cmd/auth/login_code.go

- [x] 2.1 移除 `LoginCodeOptions` 中的 `NoOpen bool` 字段
- [x] 2.2 移除 `--no-open` flag 注册
- [x] 2.3 移除 `flowOpts` 中的 `AutoOpen: !opts.NoOpen` 赋值

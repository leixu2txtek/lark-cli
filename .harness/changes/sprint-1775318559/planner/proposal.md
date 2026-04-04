# OIDC Token 获取及自动刷新实现规划

## 背景

### 为什么需要这个任务

当前 CLI 的 OIDC 认证流程已实现基础功能，但存在以下问题需要解决：

1. **缺少官方 API 对接**: 当前实现使用 `/authen/v2/oauth/token` 端点，但飞书官方 OIDC API 提供专门的 `/authen/v1/oidc/access_token` 端点用于 OIDC 场景，需要对接官方推荐的 API。

2. **Refresh Token 机制不完善**: 现有的 `token_refresher.go` 使用传统的 OAuth2 `refresh_token` 方式，但飞书 OIDC 场景下需要使用专门的 `oidc.refresh_access_token.create` API 进行刷新。

3. **OIDC 特异性处理不足**: 
   - ID Token 的验证和使用未充分实现
   - 缺少对 OIDC 特有声明（claims）的处理
   - 刷新流程未区分 OIDC 和普通 OAuth2 场景

4. **代码复用性差**: OIDC flow 和 Auth Code flow 有大量重复逻辑，需要抽象共用组件。

## 目标

### 期望达成的效果

1. **实现官方 OIDC API 对接**:
   - 使用 `oidc.access_token.create` API 获取 access_token
   - 使用 `oidc.refresh_access_token.create` API 刷新 access_token

2. **完善 Token 自动刷新机制**:
   - 自动检测即将过期的 Token（提前 5 分钟）
   - 使用正确的 API 刷新 OIDC Token
   - 刷新失败时提供清晰的错误提示和降级方案

3. **增强 ID Token 处理**:
   - 实现 ID Token 的完整验证（签名、过期、audience、issuer）
   - 正确解析和存储 ID Token 中的用户声明

4. **提升代码质量**:
   - 抽象 OIDC 和 OAuth2 的共用逻辑
   - 完善单元测试和集成测试
   - 提供清晰的错误处理和日志输出

## 方案选择

### 技术方案对比

#### 方案 A: 完全对接飞书 OIDC 原生 API（推荐）

**实现方式**:
- 使用 `POST /open-apis/authen/v1/oidc/access_token` 获取 Token
- 使用 `POST /open-apis/authen/v1/oidc/refresh_access_token` 刷新 Token
- 请求格式使用飞书特定的 JSON 格式

**优点**:
- 符合飞书官方最佳实践
- 支持 OIDC 特有的功能和声明
- 与飞书其他 OIDC 服务集成更顺畅
- 长期维护性更好

**缺点**:
- 需要修改现有 Token 交换逻辑
- 需要处理 API 格式差异

#### 方案 B: 保持现有实现，仅做 minor 改进

**实现方式**:
- 继续使用 `/authen/v2/oauth/token` 端点
- 仅优化刷新逻辑和错误处理

**优点**:
- 改动最小
- 风险低

**缺点**:
- 不符合 OIDC 标准实践
- 可能错过飞书 OIDC 新特性
- 长期技术债务

#### 方案 C: 双模式支持

**实现方式**:
- 同时支持 OIDC API 和 OAuth2 API
- 根据配置或自动检测选择使用哪个端点

**优点**:
- 向后兼容
- 灵活性高

**缺点**:
- 代码复杂度高
- 维护成本大
- 测试覆盖面要求高

### 选型结论

**选择方案 A**（完全对接飞书 OIDC 原生 API），原因如下：

1. **符合项目定位**: 作为飞书官方 CLI，应使用官方推荐的 API
2. **技术正确性**: OIDC 场景使用 OIDC 专用端点是行业标准实践
3. **可维护性**: 单一明确的实现路径，减少技术债务
4. **功能完整性**: 支持完整的 OIDC 特性，为未来扩展打下基础

考虑到当前用户量和使用场景，方案 A 的改动风险可控，且已有 `login-oidc` 命令作为基础，迁移成本较低。

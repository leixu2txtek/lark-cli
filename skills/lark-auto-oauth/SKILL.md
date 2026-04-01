---
name: lark-auto-oauth
description: Automated OAuth authentication for lark-cli. Use when the user asks to authenticate with Lark, login to Lark, or verify user credentials. Uses Authorization Code Flow by default with automatic browser-based authentication.
---

# Lark Auto OAuth

Automated OAuth authentication using lark-cli built-in commands.

## Quick Authentication

```bash
# One-command authentication (Authorization Code Flow)
lark-cli auth login-code

# Verify authentication
lark-cli auth status --verify
```

## Complete Setup Flow

```bash
# 1. Check current status
lark-cli auth status

# 2. If not configured, initialize
lark-cli config init

# 3. Login (Authorization Code Flow - default)
lark-cli auth login-code

# 4. Verify user info
lark-cli auth status --verify
```

## Authentication Method

Authorization Code Flow (default):

- Browser-based OAuth
- Supports custom domains
- Local callback server (default: http://localhost:3000/callback)
- Auto-opens browser for authentication

**Custom options:**

```bash
# Custom redirect URI
lark-cli auth login-code --redirect-uri http://localhost:8080/callback

# Specific scopes
lark-cli auth login-code --scope "calendar:calendar:readonly im:message:send"

# Manual browser opening
lark-cli auth login-code --no-open

# Custom timeout
lark-cli auth login-code --timeout 600
```

## Verification

```bash
# Check status (local)
lark-cli auth status

# Verify with server
lark-cli auth status --verify

# Check specific scope
lark-cli auth check "calendar:calendar:readonly"
```

## Expected Output

Successful authentication returns:

```json
{
  "identity": "user",
  "userName": "John Doe",
  "userOpenId": "ou_xxx",
  "tokenStatus": "valid",
  "scope": "calendar:calendar:readonly im:message:send ...",
  "verified": true
}
```

## Common Issues

**Config not found**: Run `lark-cli config init`

**Browser failed to open**: Use `--no-open` and manually open the URL

**Timeout**: Increase with `--timeout 600`

**Invalid redirect URI**: Must use localhost or 127.0.0.1

## Reference

See [references/scopes.md](references/scopes.md) for common OAuth scopes.

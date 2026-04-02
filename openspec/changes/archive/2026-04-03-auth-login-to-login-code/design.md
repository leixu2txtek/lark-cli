## Context

The lark-cli currently has two separate authentication flows: `auth login` and `login-code`. The `login-code` command was recently enhanced with Authorization Code Flow using a local callback server and browser automation, which provides a more robust authentication experience. The `auth login` command still uses the older flow without the local callback server approach. This creates inconsistency and prevents skills from reusing the same authentication mechanism.

## Goals / Non-Goals

**Goals:**
- Unify the authentication flow between `auth login` and `login-code` commands
- Enable all skills to reuse the same authentication mechanism
- Maintain backward compatibility for existing users
- Leverage the proven Authorization Code Flow implementation already in `login-code`

**Non-Goals:**
- Change the user-facing interface of the `auth login` command
- Modify the underlying OAuth2 libraries beyond what's needed for consistency
- Refactor unrelated parts of the authentication system

## Decisions

1. **Reuse existing login-code implementation**: Rather than duplicating the Authorization Code Flow logic, we'll refactor the common functionality into shared functions that both commands can use.

2. **Maintain command interface**: The `auth login` command will continue to work the same way from the user's perspective, but will internally use the same flow as `login-code`.

3. **Local callback server approach**: Both commands will use the local callback server method for receiving the authorization code, eliminating the need for manual code copying.

4. **Browser automation**: Both commands will use browser automation to streamline the authentication process, removing the need for users to manually open URLs.

## Risks / Trade-offs

[Risk: Breaking changes to auth login behavior] → Mitigation: Thorough testing to ensure user experience remains consistent
[Risk: Duplicate code between auth login and login-code] → Mitigation: Extract common functionality into shared modules
[Risk: Regression in existing auth functionality] → Mitigation: Comprehensive testing of both auth login and login-code flows
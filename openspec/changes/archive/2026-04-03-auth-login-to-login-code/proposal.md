## Why

Currently, the `auth login` command uses a different authentication flow than `login-code`, which creates inconsistency and prevents skills from automatically reusing the same authentication mechanism. By standardizing the `auth login` command to use the same `login-code` flow, all skills can benefit from a unified authentication approach.

## What Changes

- Modify the `auth login` command to use the same Authorization Code Flow with local callback server and browser automation as implemented in `login-code`
- Remove the current automatic browser opening behavior from the existing `auth login` command
- Maintain backward compatibility for existing users while adopting the improved authentication method
- Ensure all existing functionality remains the same from the user's perspective

## Capabilities

### Modified Capabilities
- `auth-login`: Updating the authentication flow implementation to use the same Authorization Code Flow as login-code command

## Impact

- The `auth login` command implementation will be updated to use the same underlying authentication mechanism as `login-code`
- All skills that depend on authentication will be able to reuse the same authentication flow
- Improved consistency across authentication commands in the CLI
- Better user experience with standardized authentication flow
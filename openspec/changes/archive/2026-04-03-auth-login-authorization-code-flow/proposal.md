## Why

The current auth login command presents users with multiple domain and permission options that are unnecessary for our specific use case. This creates a confusing and complex UX for a simple authorization flow. We need to streamline the process to directly reach the authorization code stage.

## What Changes

- Modify the auth login command to bypass the domain and permission selection steps
- Simplify the flow to go directly to the authorization code phase
- Remove unnecessary interactive prompts that don't add value for our specific use case
- Maintain backward compatibility for other authentication flows

## Capabilities

### New Capabilities
- `direct-auth-code`: Capability for direct authorization code flow without extra prompts

### Modified Capabilities
- `auth-login`: Modify existing auth login flow to support streamlined option

## Impact

- Affected code: cmd/auth/login.go and related authentication handlers
- Simpler user experience for authorization code flow
- Potentially modified command-line interface for auth command
- Updated documentation for the simplified flow
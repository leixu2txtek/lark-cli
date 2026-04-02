## Context

The current auth login command in lark-cli provides multiple domain and permission options which creates a complex user experience for what should be a simple authorization flow. Based on the recent migration to authorization code flow (as seen in the commit history), we now need to streamline the command to bypass unnecessary interactive prompts and go directly to the authorization code stage for certain use cases.

## Goals / Non-Goals

**Goals:**
- Simplify the auth login flow by bypassing domain and permission selection steps
- Provide a direct path to the authorization code phase for appropriate use cases
- Maintain backward compatibility with existing authentication methods
- Improve user experience by reducing unnecessary prompts

**Non-Goals:**
- Completely overhaul the entire authentication system
- Modify other authentication flows beyond the login command
- Add new authentication methods (only streamline existing ones)

## Decisions

1. **Command Option Approach**: Add a new flag (e.g., `--direct-code` or `--skip-selection`) to the auth login command that bypasses the domain/permission selection prompts and goes directly to the authorization code flow.

2. **Conditional Logic**: Implement conditional logic in the auth login command to detect when the direct code flow should be used, skipping the interactive domain and permission selection.

3. **Preserve Existing Flow**: Keep the existing interactive flow as the default behavior to maintain backward compatibility.

4. **Authorization Code Endpoint**: Direct the simplified flow to the existing authorization code endpoint that was implemented in the previous migration, ensuring consistency with the new authentication approach.

## Risks / Trade-offs

[Risk: Reduced flexibility] → Mitigation: Keep the default behavior as the full interactive flow, only offering the simplified path as an option
[Risk: Confusion for new users] → Mitigation: Provide clear documentation and help text explaining when to use the simplified flow
[Risk: Breaking changes] → Mitigation: Ensure default behavior remains the same, only adding new options
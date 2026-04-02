## MODIFIED Requirements

### Requirement: Auth login uses Authorization Code Flow with local callback
The auth login command SHALL use the same Authorization Code Flow with local callback server and browser automation as the login-code command.

#### Scenario: Successful authentication via auth login
- **WHEN** user runs `lark auth login` command
- **THEN** the CLI starts a local callback server on a random port
- **AND** the CLI opens the authorization URL in the browser automatically
- **AND** the CLI waits for the callback with the authorization code
- **AND** the CLI exchanges the authorization code for access and refresh tokens
- **AND** the CLI stores the tokens securely for future API calls

#### Scenario: Failed authentication via auth login
- **WHEN** user cancels the authentication in the browser or the callback fails
- **THEN** the CLI stops the local callback server
- **AND** the CLI displays an appropriate error message
- **AND** no tokens are stored

### Requirement: Auth login maintains backward compatibility
The auth login command SHALL maintain the same user-facing interface while using the updated authentication implementation internally.

#### Scenario: User runs auth login command as before
- **WHEN** user runs `lark auth login` command
- **THEN** the command behaves the same way from user's perspective as before
- **AND** the authentication happens seamlessly using the new flow
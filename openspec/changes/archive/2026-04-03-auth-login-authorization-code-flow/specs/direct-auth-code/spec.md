## ADDED Requirements

### Requirement: Direct authorization code flow
The system SHALL provide a direct authorization code flow that skips domain and permission selection steps.

#### Scenario: User initiates direct auth code flow
- **WHEN** user runs `lark-cli auth login --direct-code`
- **THEN** the system bypasses domain and permission selection prompts
- **AND** proceeds directly to the authorization code stage

#### Scenario: User completes direct auth code flow
- **WHEN** user completes the direct authorization code flow
- **THEN** the system follows the same authorization code process as the standard flow
- **AND** the resulting authentication is functionally equivalent to the standard flow

## MODIFIED Requirements

### Requirement: Auth login command options
The auth login command SHOULD provide an option to skip interactive domain and permission selection for streamlined authentication.

#### Scenario: User runs standard auth login
- **WHEN** user runs `lark-cli auth login` without --direct-code flag
- **THEN** the system behaves as it currently does with domain and permission selection

#### Scenario: User runs auth login with direct code flag
- **WHEN** user runs `lark-cli auth login --direct-code`
- **THEN** the system skips the domain and permission selection steps
- **AND** proceeds directly to the authorization code phase
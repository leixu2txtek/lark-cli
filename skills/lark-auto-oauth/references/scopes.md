# Lark OAuth Scopes Reference

Common OAuth scopes for lark-cli authentication, organized by domain.

## Recommended Scopes

The `--recommend` flag automatically selects these commonly used scopes:

```
calendar:calendar:readonly
im:message:send
im:message:readonly
contact:user:readonly
drive:drive:readonly
docs:doc:readonly
```

## Domain-Specific Scopes

### Calendar

```
calendar:calendar:readonly          # Read calendar events
calendar:calendar                   # Full calendar access
```

### Messenger (IM)

```
im:message:send                     # Send messages
im:message:readonly                 # Read messages
im:chat:readonly                    # Read chat info
im:chat                             # Full chat access
```

### Docs

```
docs:doc:readonly                   # Read documents
docs:doc                            # Full document access
```

### Drive

```
drive:drive:readonly                # Read files
drive:drive                         # Full drive access
```

### Base

```
bitable:app:readonly                # Read base apps
bitable:app                         # Full base access
```

### Sheets

```
sheets:spreadsheet:readonly         # Read spreadsheets
sheets:spreadsheet                  # Full spreadsheet access
```

### Contact

```
contact:user:readonly               # Read user info
contact:contact:readonly            # Read contacts
```

### Tasks

```
task:task:readonly                  # Read tasks
task:task                           # Full task access
```

### Mail

```
mail:mail:readonly                  # Read emails
mail:mail:send                      # Send emails
```

### Meetings

```
vc:meeting:readonly                 # Read meeting info
```

## Scope Format

Scopes follow the pattern: `domain:resource:permission`

- **domain**: Service area (calendar, im, docs, etc.)
- **resource**: Specific resource type
- **permission**: Access level (readonly, send, etc.)

## Multiple Scopes

Specify multiple scopes separated by spaces:

```bash
lark-cli auth login-code --scope "calendar:calendar:readonly im:message:send docs:doc:readonly"
```

## Checking Available Scopes

List all scopes enabled for your app:

```bash
lark-cli auth scopes
```

## Verifying Scope Access

Check if a specific scope is granted:

```bash
lark-cli auth check "calendar:calendar:readonly"
```

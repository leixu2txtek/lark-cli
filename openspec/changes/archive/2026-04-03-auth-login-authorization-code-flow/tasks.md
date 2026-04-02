## 1. Command Line Interface Updates

- [x] 1.1 Add --direct-code flag to auth login command
- [x] 1.2 Update command help text to document the new flag
- [x] 1.3 Implement conditional logic to detect the --direct-code flag

## 2. Authentication Flow Modifications

- [x] 2.1 Modify auth login command to skip domain selection when --direct-code is used
- [x] 2.2 Modify auth login command to skip permission selection when --direct-code is used
- [x] 2.3 Ensure the authorization code flow continues properly from the direct path

## 3. Testing

- [x] 3.1 Create unit tests for the new --direct-code flag behavior
- [x] 3.2 Test that the direct code flow works as expected
- [x] 3.3 Verify that the default behavior remains unchanged

## 4. Documentation and Verification

- [x] 4.1 Update user documentation to reflect the new option
- [x] 4.2 Verify backward compatibility is maintained
- [x] 4.3 Test complete auth flow with the new flag
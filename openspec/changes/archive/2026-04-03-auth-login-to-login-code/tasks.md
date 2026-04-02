## 1. Refactor Authentication Logic

- [x] 1.1 Extract common authentication functions from login-code command
- [x] 1.2 Create shared authentication module for both auth login and login-code
- [x] 1.3 Update auth login command to use shared authentication module

## 2. Update Command Implementation

- [x] 2.1 Modify auth login command to use Authorization Code Flow with local callback
- [x] 2.2 Ensure auth login uses browser automation like login-code
- [x] 2.3 Test auth login command maintains backward compatibility

## 3. Testing and Validation

- [x] 3.1 Test auth login command with new implementation
- [x] 3.2 Verify both auth login and login-code use same authentication flow
- [x] 3.3 Confirm skills can reuse authentication mechanism
- [x] 3.4 Run integration tests to ensure no regressions
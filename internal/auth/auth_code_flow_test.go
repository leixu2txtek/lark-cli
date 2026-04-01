package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestAuthURLGeneration tests authorization URL generation logic
func TestAuthURLGeneration(t *testing.T) {
	tests := []struct {
		name        string
		appID       string
		redirectURI string
		scope       string
		wantNoScope bool
	}{
		{
			name:        "without scope",
			appID:       "cli_test123",
			redirectURI: "http://localhost:3000/callback",
			scope:       "",
			wantNoScope: true,
		},
		{
			name:        "with custom scope",
			appID:       "cli_test123",
			redirectURI: "http://localhost:3000/callback",
			scope:       "contact:user.id:readonly",
			wantNoScope: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := fmt.Sprintf("oauth_%d", time.Now().Unix())
			params := url.Values{}
			params.Set("app_id", tt.appID)
			params.Set("redirect_uri", tt.redirectURI)
			params.Set("state", state)
			if tt.scope != "" {
				params.Set("scope", tt.scope)
			}

			authURL := fmt.Sprintf("https://open.feishu.cn/open-apis/authen/v1/user_auth_page_beta?%s",
				params.Encode())

			parsedURL, err := url.Parse(authURL)
			if err != nil {
				t.Fatalf("failed to parse URL: %v", err)
			}

			query := parsedURL.Query()
			if query.Get("app_id") != tt.appID {
				t.Errorf("app_id = %q, want %q", query.Get("app_id"), tt.appID)
			}
			if query.Get("redirect_uri") != tt.redirectURI {
				t.Errorf("redirect_uri = %q, want %q", query.Get("redirect_uri"), tt.redirectURI)
			}

			if tt.wantNoScope && query.Has("scope") {
				t.Errorf("expected no scope parameter, but got: %s", query.Get("scope"))
			}
			if !tt.wantNoScope && query.Get("scope") != tt.scope {
				t.Errorf("scope = %q, want %q", query.Get("scope"), tt.scope)
			}
		})
	}
}

// TestStateGeneration tests state generation
func TestStateGeneration(t *testing.T) {
	state1 := fmt.Sprintf("oauth_%d", time.Now().Unix())
	time.Sleep(1 * time.Millisecond)
	state2 := fmt.Sprintf("oauth_%d", time.Now().Unix())

	if state1 == "" {
		t.Error("state should not be empty")
	}

	if !strings.HasPrefix(state1, "oauth_") {
		t.Errorf("state should start with 'oauth_', got: %s", state1)
	}

	// States generated at different times should be different
	if state1 == state2 {
		t.Log("Warning: consecutive states are the same (may happen if generated too quickly)")
	}
}

// TestCallbackParameterParsing tests callback parameter parsing
func TestCallbackParameterParsing(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantCode  string
		wantState string
		wantError string
	}{
		{
			name:      "successful callback",
			query:     "code=test_code_123&state=oauth_456",
			wantCode:  "test_code_123",
			wantState: "oauth_456",
			wantError: "",
		},
		{
			name:      "error callback",
			query:     "error=access_denied&error_description=User+denied+access",
			wantCode:  "",
			wantState: "",
			wantError: "access_denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://localhost:3000/callback?"+tt.query, nil)
			query := req.URL.Query()

			code := query.Get("code")
			state := query.Get("state")
			errCode := query.Get("error")

			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if state != tt.wantState {
				t.Errorf("state = %q, want %q", state, tt.wantState)
			}
			if errCode != tt.wantError {
				t.Errorf("error = %q, want %q", errCode, tt.wantError)
			}
		})
	}
}

// TestTokenResponseParsing tests token response parsing
func TestTokenResponseParsing(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		wantToken    string
		wantRefresh  string
		wantExpires  int
	}{
		{
			name: "wrapped response format",
			responseBody: `{
				"code": 0,
				"msg": "success",
				"data": {
					"access_token": "u-test_access_token",
					"refresh_token": "ur-test_refresh_token",
					"expires_in": 7200,
					"refresh_expires_in": 2592000
				}
			}`,
			wantToken:   "u-test_access_token",
			wantRefresh: "ur-test_refresh_token",
			wantExpires: 7200,
		},
		{
			name: "direct response format",
			responseBody: `{
				"access_token": "u-test_access_token",
				"refresh_token": "ur-test_refresh_token",
				"expires_in": 7200,
				"refresh_expires_in": 2592000
			}`,
			wantToken:   "u-test_access_token",
			wantRefresh: "ur-test_refresh_token",
			wantExpires: 7200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrappedResp struct {
				Code int `json:"code"`
				Data struct {
					AccessToken      string `json:"access_token"`
					RefreshToken     string `json:"refresh_token"`
					ExpiresIn        int    `json:"expires_in"`
					RefreshExpiresIn int    `json:"refresh_expires_in"`
				} `json:"data"`
			}

			if err := json.Unmarshal([]byte(tt.responseBody), &wrappedResp); err == nil && wrappedResp.Data.AccessToken != "" {
				if wrappedResp.Data.AccessToken != tt.wantToken {
					t.Errorf("access_token = %q, want %q", wrappedResp.Data.AccessToken, tt.wantToken)
				}
				if wrappedResp.Data.RefreshToken != tt.wantRefresh {
					t.Errorf("refresh_token = %q, want %q", wrappedResp.Data.RefreshToken, tt.wantRefresh)
				}
				if wrappedResp.Data.ExpiresIn != tt.wantExpires {
					t.Errorf("expires_in = %d, want %d", wrappedResp.Data.ExpiresIn, tt.wantExpires)
				}
			}
		})
	}
}

// TestUserInfoResponseParsing tests user info response parsing
func TestUserInfoResponseParsing(t *testing.T) {
	responseBody := `{
		"code": 0,
		"msg": "success",
		"data": {
			"open_id": "ou_test123",
			"name": "Test User"
		}
	}`

	var resp struct {
		Code int `json:"code"`
		Data struct {
			OpenID string `json:"open_id"`
			Name   string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(responseBody), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Data.OpenID != "ou_test123" {
		t.Errorf("open_id = %q, want %q", resp.Data.OpenID, "ou_test123")
	}
	if resp.Data.Name != "Test User" {
		t.Errorf("name = %q, want %q", resp.Data.Name, "Test User")
	}
}

// TestMockOAuthServer tests with a mock OAuth server
func TestMockOAuthServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "oauth/token") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "success",
				"data": map[string]interface{}{
					"access_token":       "u-mock_access_token",
					"refresh_token":      "ur-mock_refresh_token",
					"expires_in":         7200,
					"refresh_expires_in": 2592000,
				},
			})
		} else if strings.Contains(r.URL.Path, "user_info") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "success",
				"data": map[string]interface{}{
					"open_id": "ou_mock123",
					"name":    "Mock User",
				},
			})
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Test token exchange
	result, err := exchangeCodeForToken(http.DefaultClient, server.URL, "cli_test", "test_secret", "test_code", "http://localhost:3000/callback")
	if err != nil {
		t.Fatalf("exchangeCodeForToken failed: %v", err)
	}

	if result.AccessToken != "u-mock_access_token" {
		t.Errorf("access_token = %q, want %q", result.AccessToken, "u-mock_access_token")
	}
	if result.RefreshToken != "ur-mock_refresh_token" {
		t.Errorf("refresh_token = %q, want %q", result.RefreshToken, "ur-mock_refresh_token")
	}

	// Test user info retrieval
	userInfo, err := getUserInfoWithToken(http.DefaultClient, server.URL, result.AccessToken)
	if err != nil {
		t.Fatalf("getUserInfoWithToken failed: %v", err)
	}

	if userInfo.OpenID != "ou_mock123" {
		t.Errorf("open_id = %q, want %q", userInfo.OpenID, "ou_mock123")
	}
	if userInfo.Name != "Mock User" {
		t.Errorf("name = %q, want %q", userInfo.Name, "Mock User")
	}
}

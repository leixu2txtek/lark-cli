package auth

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larksuite/cli/internal/core"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestRefreshCredentialOIDCRefreshUpdatesAndPreservesOIDCFields(t *testing.T) {
	now := time.Now().UnixMilli()
	stored := &StoredUAToken{
		UserOpenId:       "ou_oidc_refresh",
		AppId:            "test_app",
		AuthProtocol:     "legacy_oidc",
		AccessToken:      "old_access_token",
		RefreshToken:     "old_refresh_token",
		IDToken:          "old.id.token",
		ExpiresAt:        now + 10*time.Minute.Milliseconds(),
		RefreshExpiresAt: now + time.Minute.Milliseconds(),
		IDTokenExpiresAt: now - time.Second.Milliseconds(),
		Scope:            "openid profile",
		GrantedAt:        now - time.Hour.Milliseconds(),
		UserInfo: map[string]interface{}{
			"email": "old@example.com",
			"name":  "Old User",
		},
	}
	t.Cleanup(func() {
		_ = RemoveStoredToken(stored.AppId, stored.UserOpenId)
	})

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/open-apis/auth/v3/app_access_token/internal":
			return jsonResponse(`{"code":0,"msg":"success","app_access_token":"app_token","app_access_token_expire":7200}`), nil
		case "/open-apis/authen/v1/oidc/refresh_access_token":
			return jsonResponse(`{"code":0,"msg":"success","data":{"access_token":"new_access_token","refresh_token":"new_refresh_token","id_token":"` + testValidIDToken + `","expires_in":7200,"refresh_expires_in":3600,"token_type":"Bearer"}}`), nil
		default:
			t.Fatalf("unexpected request path: %s", req.URL.Path)
			return nil, nil
		}
	})}

	var errOut bytes.Buffer
	store := NewCredentialStore(&memoryCredentials{values: map[string]string{}}, t.TempDir())
	refreshed, err := refreshCredential(context.Background(), client, UATCallOptions{
		UserOpenId: stored.UserOpenId,
		AppId:      stored.AppId,
		AppSecret:  "test_secret",
		Domain:     core.BrandFeishu,
		ErrOut:     &errOut,
	}, store, stored)
	if err != nil {
		t.Fatalf("refreshCredential failed: %v", err)
	}

	if refreshed.AccessToken != "new_access_token" {
		t.Fatalf("AccessToken = %q, want %q", refreshed.AccessToken, "new_access_token")
	}
	if refreshed.RefreshToken != "new_refresh_token" {
		t.Fatalf("RefreshToken = %q, want %q", refreshed.RefreshToken, "new_refresh_token")
	}
	if refreshed.IDToken != testValidIDToken {
		t.Fatalf("IDToken was not updated from refresh response")
	}
	if refreshed.IDTokenExpiresAt != 9999999999*1000 {
		t.Fatalf("IDTokenExpiresAt = %d, want parsed test token exp", refreshed.IDTokenExpiresAt)
	}
	if refreshed.RefreshExpiresAt <= now+30*time.Minute.Milliseconds() {
		t.Fatalf("RefreshExpiresAt = %d, want value extended from refresh_expires_in", refreshed.RefreshExpiresAt)
	}
	if refreshed.UserInfo["email"] != "old@example.com" {
		t.Fatalf("UserInfo was not preserved: %#v", refreshed.UserInfo)
	}
}

func TestRefreshOIDCTokenResolvesBrandToOpenDomain(t *testing.T) {
	var seenHosts []string
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seenHosts = append(seenHosts, req.URL.Host)
		switch req.URL.Path {
		case "/open-apis/auth/v3/app_access_token/internal":
			return jsonResponse(`{"code":0,"msg":"success","app_access_token":"app_token","app_access_token_expire":7200}`), nil
		case "/open-apis/authen/v1/oidc/refresh_access_token":
			return jsonResponse(`{"code":0,"msg":"success","data":{"access_token":"new_access_token","refresh_token":"new_refresh_token","expires_in":7200,"refresh_expires_in":3600,"token_type":"Bearer"}}`), nil
		default:
			t.Fatalf("unexpected request path: %s", req.URL.Path)
			return nil, nil
		}
	})}

	stored := &StoredUAToken{
		AppId:        "cli_test",
		UserOpenId:   "ou_test",
		AuthProtocol: "legacy_oidc",
		RefreshToken: "refresh",
	}
	store := NewCredentialStore(&memoryCredentials{values: map[string]string{}}, t.TempDir())
	_, err := refreshCredential(context.Background(), client, UATCallOptions{
		UserOpenId: "ou_test",
		AppId:      "cli_test",
		AppSecret:  "secret",
		Domain:     core.BrandFeishu,
	}, store, stored)
	if err != nil {
		t.Fatalf("refreshCredential failed: %v", err)
	}
	for _, host := range seenHosts {
		if host != "open.feishu.cn" {
			t.Fatalf("refresh request host = %q, want open.feishu.cn; all hosts: %#v", host, seenHosts)
		}
	}
}

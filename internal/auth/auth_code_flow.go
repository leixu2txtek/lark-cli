// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AuthCodeFlowOptions contains options for authorization code flow.
type AuthCodeFlowOptions struct {
	AppID       string
	AppSecret   string
	Domain      string
	RedirectURI string
	Scope       string // OAuth scope (empty by default)
	Timeout     time.Duration
}

// AuthCodeFlowResult contains the result of authorization code flow.
type AuthCodeFlowResult struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	Scope            string
	OpenID           string
	UserName         string
}

// callbackData stores the OAuth callback parameters.
type callbackData struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
}

// StartAuthCodeFlow initiates the authorization code flow.
func StartAuthCodeFlow(ctx context.Context, opts *AuthCodeFlowOptions, httpClient *http.Client, errOut io.Writer) (*AuthCodeFlowResult, error) {
	if errOut == nil {
		errOut = io.Discard
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	// Parse redirect URI
	redirectURL, err := url.Parse(opts.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URI: %v", err)
	}

	host := redirectURL.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return nil, fmt.Errorf("redirect URI host must be localhost or 127.0.0.1")
	}

	port := redirectURL.Port()
	if port == "" {
		port = "3000"
	}

	// Generate state for CSRF protection
	state := fmt.Sprintf("oauth_%d", time.Now().Unix())

	// Start callback server
	callbackChan := make(chan *callbackData, 1)
	server := &http.Server{
		Addr:    host + ":" + port,
		Handler: createCallbackHandler(redirectURL.Path, callbackChan),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(errOut, "[WARN] callback server error: %v\n", err)
		}
	}()
	defer server.Shutdown(context.Background())

	// Build authorization URL
	params := url.Values{}
	params.Set("app_id", opts.AppID)
	params.Set("redirect_uri", opts.RedirectURI)
	params.Set("state", state)
	if opts.Scope != "" {
		params.Set("scope", opts.Scope)
	}
	authURL := fmt.Sprintf("%s/open-apis/authen/v1/user_auth_page_beta?%s",
		opts.Domain,
		params.Encode(),
	)

	fmt.Fprintf(errOut, "Authorization URL: %s\n", authURL)

	// Wait for callback
	fmt.Fprintf(errOut, "Waiting for authorization callback...\n")
	var callback *callbackData
	select {
	case callback = <-callbackChan:
	case <-time.After(opts.Timeout):
		return nil, fmt.Errorf("timeout waiting for callback after %v", opts.Timeout)
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled")
	}

	// Validate callback
	if callback.Error != "" {
		return nil, fmt.Errorf("OAuth error: %s - %s", callback.Error, callback.ErrorDescription)
	}
	if callback.Code == "" {
		return nil, fmt.Errorf("empty authorization code")
	}
	if callback.State != state {
		return nil, fmt.Errorf("state mismatch")
	}

	// Exchange code for token
	tokenResp, err := exchangeCodeForToken(httpClient, opts.Domain, opts.AppID, opts.AppSecret, callback.Code, opts.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %v", err)
	}

	// Get user info
	userInfo, err := getUserInfoWithToken(httpClient, opts.Domain, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	return &AuthCodeFlowResult{
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		ExpiresIn:        tokenResp.ExpiresIn,
		RefreshExpiresIn: tokenResp.RefreshExpiresIn,
		Scope:            tokenResp.Scope,
		OpenID:           userInfo.OpenID,
		UserName:         userInfo.Name,
	}, nil
}

// createCallbackHandler creates HTTP handler for OAuth callback.
func createCallbackHandler(path string, callbackChan chan<- *callbackData) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}

		query := r.URL.Query()
		data := &callbackData{
			Code:             query.Get("code"),
			State:            query.Get("state"),
			Error:            query.Get("error"),
			ErrorDescription: query.Get("error_description"),
		}

		callbackChan <- data

		if data.Code != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>授权成功</title></head><body><h1>授权成功</h1><p>可以关闭此页面并返回终端。</p></body></html>`)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>授权失败</title></head><body><h1>授权失败</h1><p>请返回终端查看错误信息。</p></body></html>`)
		}
	})
}

// tokenResponse represents the OAuth token response.
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
}

// exchangeCodeForToken exchanges authorization code for access token.
func exchangeCodeForToken(client *http.Client, domain, appID, appSecret, code, redirectURI string) (*tokenResponse, error) {
	payload := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  redirectURI,
		"client_id":     appID,
		"client_secret": appSecret,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", domain+"/open-apis/authen/v2/oauth/token", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Try to parse as wrapped response first (with "data" field)
	var wrappedResult struct {
		Code  int            `json:"code"`
		Msg   string         `json:"msg"`
		Error string         `json:"error"`
		Data  *tokenResponse `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &wrappedResult); err == nil {
		if wrappedResult.Error != "" {
			return nil, fmt.Errorf("API error: %s", wrappedResult.Error)
		}
		if wrappedResult.Code != 0 && wrappedResult.Msg != "" {
			return nil, fmt.Errorf("API error [%d]: %s", wrappedResult.Code, wrappedResult.Msg)
		}
		if wrappedResult.Data != nil && wrappedResult.Data.AccessToken != "" {
			return wrappedResult.Data, nil
		}
	}

	// Try to parse as direct response (token fields at root level)
	var directResult tokenResponse
	if err := json.Unmarshal(bodyBytes, &directResult); err == nil {
		if directResult.AccessToken != "" {
			return &directResult, nil
		}
	}

	return nil, fmt.Errorf("missing token data in response: %s", string(bodyBytes))
}

// userInfoData represents user information.
type userInfoData struct {
	OpenID string `json:"open_id"`
	Name   string `json:"name"`
}

// getUserInfoWithToken retrieves user information using access token.
func getUserInfoWithToken(client *http.Client, domain, accessToken string) (*userInfoData, error) {
	req, err := http.NewRequest("GET", domain+"/open-apis/authen/v1/user_info", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *userInfoData `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error [%d]: %s", result.Code, result.Msg)
	}
	if result.Data == nil || result.Data.OpenID == "" {
		return nil, fmt.Errorf("missing user info")
	}

	return result.Data, nil
}

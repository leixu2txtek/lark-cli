// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OidcAccessTokenClient OIDC Access Token API 客户端
type OidcAccessTokenClient struct {
	httpClient  *http.Client
	domain      string // API domain (Open domain for token endpoints)
	appID       string
	appSecret   string
	appAccessToken string // 缓存的 App Access Token
	tokenExpiresAt int64  // App Access Token 过期时间
}

// NewOidcAccessTokenClient 创建新的 OIDC Access Token 客户端
func NewOidcAccessTokenClient(httpClient *http.Client, domain string) *OidcAccessTokenClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OidcAccessTokenClient{
		httpClient: httpClient,
		domain:     domain,
	}
}

// SetAppCredentials 设置应用凭证
func (c *OidcAccessTokenClient) SetAppCredentials(appID, appSecret string) {
	c.appID = appID
	c.appSecret = appSecret
	c.appAccessToken = ""
	c.tokenExpiresAt = 0
}

// getAppAccessToken 获取 App Access Token（内部使用）
func (c *OidcAccessTokenClient) getAppAccessToken(ctx context.Context) (string, error) {
	now := time.Now().Unix()

	// 检查缓存的 Token 是否有效
	if c.appAccessToken != "" && c.tokenExpiresAt > now+60 {
		return c.appAccessToken, nil
	}

	// 获取新的 App Access Token
	// 处理 domain 可能包含协议前缀的情况
	scheme := "https"
	baseURL := strings.TrimSpace(c.domain)
	if idx := strings.Index(baseURL, "://"); idx != -1 {
		scheme = baseURL[:idx]
		baseURL = baseURL[idx+3:]
	}
	baseURL = strings.TrimLeft(baseURL, "/")
	url := fmt.Sprintf("%s://%s/open-apis/auth/v3/app_access_token/internal", scheme, baseURL)
	payload := map[string]string{
		"app_id":     c.appID,
		"app_secret": c.appSecret,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get app_access_token: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read app_access_token response: %w", err)
	}

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		AppAccessToken    string `json:"app_access_token"`
		AppAccessTokenExpire int `json:"app_access_token_expire"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse app_access_token response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("failed to get app_access_token [%d]: %s", result.Code, result.Msg)
	}

	c.appAccessToken = result.AppAccessToken
	c.tokenExpiresAt = now + int64(result.AppAccessTokenExpire) - 60 // 提前 60 秒过期

	return c.appAccessToken, nil
}

// CreateAccessTokenParams 创建 Access Token 的请求参数
type CreateAccessTokenParams struct {
	AppID       string `json:"app_id"`
	AppSecret   string `json:"app_secret"`
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

// CreateAccessTokenResponse API 响应
type CreateAccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	OpenID           string `json:"open_id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
}

// UserInfoResponse 用户信息 API 响应
type UserInfoResponse struct {
	OpenID  string `json:"open_id"`
	UnionID string `json:"union_id"`
	Name    string `json:"name"`
	EnName  string `json:"en_name"`
	Email   string `json:"email"`
}

// CreateAccessToken 使用授权码获取 OIDC access_token
// POST https://{domain}/open-apis/authen/v1/oidc/access_token
func (c *OidcAccessTokenClient) CreateAccessToken(
	ctx context.Context,
	params CreateAccessTokenParams,
) (*CreateAccessTokenResponse, error) {
	// 设置应用凭证（用于获取 App Access Token）
	if c.appID == "" || c.appSecret == "" {
		c.appID = params.AppID
		c.appSecret = params.AppSecret
	}

	// 获取 App Access Token
	appAccessToken, err := c.getAppAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// 构建请求体（不包含 app_secret，使用 App Access Token 认证）
	requestBody := map[string]string{
		"grant_type":   "authorization_code",
		"code":         params.Code,
		"redirect_uri": params.RedirectURI,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 支持 HTTP 和 HTTPS（测试时使用 HTTP）
	// domain 可能是 "https://open.feishu.cn" 或 "open.feishu.cn"
	scheme := "https"
	baseURL := strings.TrimSpace(c.domain)
	// 如果 domain 已包含协议前缀，则提取它
	if idx := strings.Index(baseURL, "://"); idx != -1 {
		scheme = baseURL[:idx]
		// 跳过 "://" 三个字符
		baseURL = baseURL[idx+3:]
	}
	// 移除可能存在的前导斜杠
	baseURL = strings.TrimLeft(baseURL, "/")
	url := fmt.Sprintf("%s://%s/open-apis/authen/v1/oidc/access_token", scheme, baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+appAccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OIDC API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应 - 飞书 API 响应格式：{ "code": 0, "msg": "success", "data": {...} }
	var result struct {
		Code  int                        `json:"code"`
		Msg   string                     `json:"msg"`
		Error string                     `json:"error"`
		Data  *CreateAccessTokenResponse `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查错误 - CreateAccessToken
	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s (response: %s)", result.Error, string(bodyBytes))
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("API error [%d]: %s (response: %s)", result.Code, result.Msg, string(bodyBytes))
	}
	if result.Data == nil {
		return nil, fmt.Errorf("missing data in response: %s", string(bodyBytes))
	}
	if result.Data.AccessToken == "" {
		return nil, fmt.Errorf("missing access_token in response")
	}

	return result.Data, nil
}

// GetUserInfo 使用 access token 获取用户信息
// GET https://{domain}/open-apis/authen/v1/user_info
func (c *OidcAccessTokenClient) GetUserInfo(
	ctx context.Context,
	accessToken string,
) (*UserInfoResponse, error) {
	// 处理 domain 可能包含协议前缀的情况
	scheme := "https"
	baseURL := strings.TrimSpace(c.domain)
	if idx := strings.Index(baseURL, "://"); idx != -1 {
		scheme = baseURL[:idx]
		baseURL = baseURL[idx+3:]
	}
	baseURL = strings.TrimLeft(baseURL, "/")
	url := fmt.Sprintf("%s://%s/open-apis/authen/v1/user_info", scheme, baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call user info API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var result struct {
		Code  int             `json:"code"`
		Msg   string          `json:"msg"`
		Error string          `json:"error"`
		Data  *UserInfoResponse `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查错误
	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s (response: %s)", result.Error, string(bodyBytes))
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("API error [%d]: %s (response: %s)", result.Code, result.Msg, string(bodyBytes))
	}
	if result.Data == nil {
		return nil, fmt.Errorf("missing data in response: %s", string(bodyBytes))
	}

	return result.Data, nil
}

// RefreshAccessTokenParams 刷新 Access Token 的请求参数
type RefreshAccessTokenParams struct {
	AppID        string `json:"app_id"`
	AppSecret    string `json:"app_secret"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshAccessToken 使用 refresh_token 刷新 access_token
// POST https://{domain}/open-apis/authen/v1/oidc/refresh_access_token
func (c *OidcAccessTokenClient) RefreshAccessToken(
	ctx context.Context,
	params RefreshAccessTokenParams,
) (*CreateAccessTokenResponse, error) {
	// 设置应用凭证（用于获取 App Access Token）
	if c.appID == "" || c.appSecret == "" {
		c.appID = params.AppID
		c.appSecret = params.AppSecret
	}

	// 获取 App Access Token
	appAccessToken, err := c.getAppAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// 构建请求体（不包含 app_secret，使用 App Access Token 认证）
	requestBody := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": params.RefreshToken,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 支持 HTTP 和 HTTPS（测试时使用 HTTP）
	// domain 可能是 "https://open.feishu.cn" 或 "open.feishu.cn"
	scheme := "https"
	baseURL := strings.TrimSpace(c.domain)
	// 如果 domain 已包含协议前缀，则提取它
	if idx := strings.Index(baseURL, "://"); idx != -1 {
		scheme = baseURL[:idx]
		// 跳过 "://" 三个字符
		baseURL = baseURL[idx+3:]
	}
	// 移除可能存在的前导斜杠
	baseURL = strings.TrimLeft(baseURL, "/")
	url := fmt.Sprintf("%s://%s/open-apis/authen/v1/oidc/refresh_access_token", scheme, baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+appAccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OIDC Refresh API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var result struct {
		Code  int                        `json:"code"`
		Msg   string                     `json:"msg"`
		Error string                     `json:"error"`
		Data  *CreateAccessTokenResponse `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查错误 - CreateAccessToken
	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s (response: %s)", result.Error, string(bodyBytes))
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("API error [%d]: %s (response: %s)", result.Code, result.Msg, string(bodyBytes))
	}
	if result.Data == nil {
		return nil, fmt.Errorf("missing data in response: %s", string(bodyBytes))
	}
	if result.Data.AccessToken == "" {
		return nil, fmt.Errorf("missing access_token in response")
	}

	return result.Data, nil
}

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/larksuite/cli/internal/keychain"
)

// TokenRefresher 负责自动刷新即将过期的 Token
type TokenRefresher struct {
	httpClient  *http.Client
	domain      string
	appID       string
	appSecret   string
	ticker      *time.Ticker
	stopCh      chan struct{}
	ctx         context.Context
	errOut      io.Writer
}

// NewTokenRefresher 创建新的 Token 刷新器
func NewTokenRefresher(ctx context.Context, httpClient *http.Client, domain, appID, appSecret string, errOut io.Writer) *TokenRefresher {
	return &TokenRefresher{
		httpClient:  httpClient,
		domain:      domain,
		appID:       appID,
		appSecret:   appSecret,
		ticker:      time.NewTicker(5 * time.Minute), // 每 5 分钟检查一次
		stopCh:      make(chan struct{}),
		ctx:         ctx,
		errOut:      errOut,
	}
}

// Start 开始后台刷新服务
func (tr *TokenRefresher) Start() {
	go func() {
		tr.checkAndRefreshTokens() // 立即执行一次检查
		for {
			select {
			case <-tr.ticker.C:
				tr.checkAndRefreshTokens()
			case <-tr.stopCh:
				tr.ticker.Stop()
				return
			case <-tr.ctx.Done():
				tr.ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止后台刷新服务
func (tr *TokenRefresher) Stop() {
	close(tr.stopCh)
}

// RefreshToken 刷新指定的 Token
func (tr *TokenRefresher) RefreshToken(storedToken *StoredUAToken) (*StoredUAToken, error) {
	if !storedToken.IsRefreshable() {
		return nil, fmt.Errorf("token is not refreshable")
	}

	// 使用新的 OIDC Refresh API 客户端
	client := NewOidcAccessTokenClient(tr.httpClient, tr.domain)
	client.SetAppCredentials(tr.appID, tr.appSecret)
	resp, err := client.RefreshAccessToken(tr.ctx, RefreshAccessTokenParams{
		AppID:        tr.appID,
		AppSecret:    tr.appSecret,
		RefreshToken: storedToken.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token via OIDC API: %w", err)
	}

	// 更新 Token 信息
	newStoredToken := *storedToken
	newStoredToken.AccessToken = resp.AccessToken
	// 如果 API 返回了新的刷新 Token，则使用它；否则保留原有的
	if resp.RefreshToken != "" {
		newStoredToken.RefreshToken = resp.RefreshToken
	}
	// 如果 API 返回了 ID Token，则更新它
	if resp.IDToken != "" {
		newStoredToken.IDToken = resp.IDToken
	}
	now := time.Now().UnixMilli()
	newStoredToken.ExpiresAt = now + int64(resp.ExpiresIn)*1000
	newStoredToken.RefreshExpiresAt = now + int64(resp.RefreshExpiresIn)*1000

	// 更新 ID Token 过期时间
	if resp.IDToken != "" {
		claims, err := GetClaims(resp.IDToken)
		if err == nil {
			if exp, exists := claims["exp"]; exists {
				if expFloat, ok := exp.(float64); ok {
					newStoredToken.IDTokenExpiresAt = int64(expFloat) * 1000
				}
			}
		}
	}

	return &newStoredToken, nil
}

// checkAndRefreshTokens 检查并刷新所有即将过期的 Token
func (tr *TokenRefresher) checkAndRefreshTokens() {
	fmt.Fprintln(tr.errOut, "Checking tokens for refresh...")

	// 获取所有存储的 keys
	keys, err := keychain.ListKeys(keychain.LarkCliService)
	if err != nil {
		fmt.Fprintf(tr.errOut, "Failed to list stored tokens: %v\n", err)
		return
	}

	// 遍历所有 keys 并检查是否为有效的 token
	for _, key := range keys {
		// 尝试获取存储的 token
		data := keychain.Get(keychain.LarkCliService, key)
		if data == "" {
			continue // Key 不存在，跳过
		}

		// 尝试解析为 StoredUAToken
		var storedToken StoredUAToken
		if err := json.Unmarshal([]byte(data), &storedToken); err != nil {
			// 如果不是有效的 StoredUAToken 格式，跳过
			continue
		}

		// 检查 token 状态
		status := TokenStatus(&storedToken)
		if status == "needs_refresh" {
			fmt.Fprintf(tr.errOut, "Refreshing token for app %s...\n", storedToken.AppId)

			refreshedToken, err := tr.RefreshToken(&storedToken)
			if err != nil {
				fmt.Fprintf(tr.errOut, "Failed to refresh token for app %s: %v\n", storedToken.AppId, err)
				continue
			}

			err = SetStoredToken(refreshedToken)
			if err != nil {
				fmt.Fprintf(tr.errOut, "Failed to store refreshed token for app %s: %v\n", storedToken.AppId, err)
				continue
			}

			fmt.Fprintf(tr.errOut, "Successfully refreshed token for app %s\n", storedToken.AppId)
		} else if status == "expired" {
			fmt.Fprintf(tr.errOut, "Token for app %s has expired and needs re-authentication\n", storedToken.AppId)
		} else {
			fmt.Fprintf(tr.errOut, "Token for app %s is still valid\n", storedToken.AppId)
		}
	}

	fmt.Fprintln(tr.errOut, "Token refresh check completed.")
}

// IsRefreshable 检查 Token 是否可以刷新
func (t *StoredUAToken) IsRefreshable() bool {
	now := time.Now().UnixMilli()
	return t.RefreshToken != "" && t.RefreshExpiresAt > now
}

// IsValid 检查 Token 是否仍然有效
func (t *StoredUAToken) IsValid() bool {
	return TokenStatus(t) == "valid"
}

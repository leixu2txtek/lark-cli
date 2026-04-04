// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/larksuite/cli/internal/keychain"
)

// StoredUAToken represents a stored user access token.
type StoredUAToken struct {
	UserOpenId       string                 `json:"userOpenId"`
	AppId            string                 `json:"appId"`
	AccessToken      string                 `json:"accessToken"`
	RefreshToken     string                 `json:"refreshToken"`
	IDToken          string                 `json:"idToken"`          // OIDC 特有
	ExpiresAt        int64                  `json:"expiresAt"`        // Unix ms - 访问 Token 过期时间
	RefreshExpiresAt int64                  `json:"refreshExpiresAt"` // Unix ms - 刷新 Token 过期时间
	IDTokenExpiresAt int64                  `json:"idTokenExpiresAt"` // Unix ms - ID Token 过期时间
	Scope            string                 `json:"scope"`
	GrantedAt        int64                  `json:"grantedAt"` // Unix ms
	UserInfo         map[string]interface{} `json:"userInfo"`  // 用户信息
}

const refreshAheadMs = 5 * 60 * 1000 // 5 minutes

func accountKey(appId, userOpenId string) string {
	return fmt.Sprintf("%s:%s", appId, userOpenId)
}

// MaskToken masks a token for safe logging.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return "****" + token[len(token)-4:]
}

// GetStoredToken reads the stored UAT for a given (appId, userOpenId) pair.
func GetStoredToken(appId, userOpenId string) *StoredUAToken {
	jsonStr := keychain.Get(keychain.LarkCliService, accountKey(appId, userOpenId))
	if jsonStr == "" {
		return nil
	}
	var token StoredUAToken
	if err := json.Unmarshal([]byte(jsonStr), &token); err != nil {
		return nil
	}
	return &token
}

// SetStoredToken persists a UAT.
func SetStoredToken(token *StoredUAToken) error {
	key := accountKey(token.AppId, token.UserOpenId)
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return keychain.Set(keychain.LarkCliService, key, string(data))
}

// RemoveStoredToken removes a stored UAT.
func RemoveStoredToken(appId, userOpenId string) error {
	return keychain.Remove(keychain.LarkCliService, accountKey(appId, userOpenId))
}

// TokenStatus determines the freshness of a stored token.
// Returns: "valid", "needs_refresh", "expired", or "id_token_expired"
func TokenStatus(token *StoredUAToken) string {
	now := time.Now().UnixMilli()

	// 首先检查 ID Token 状态（如果是 OIDC 令牌）
	if token.IDToken != "" && token.IDTokenExpiresAt > 0 {
		if now >= token.IDTokenExpiresAt {
			return "id_token_expired"
		}
	}

	// 检查访问 Token 是否即将过期（提前 5 分钟刷新）
	if now+refreshAheadMs >= token.ExpiresAt {
		if token.RefreshToken != "" && token.RefreshExpiresAt > now {
			return "needs_refresh"
		}
		return "expired"
	}

	return "valid"
}

// ShouldRefreshIDToken 判断是否需要重新获取 ID Token
// 当 ID Token 不存在或即将过期时返回 true
func (t *StoredUAToken) ShouldRefreshIDToken() bool {
	if t.IDToken == "" {
		return true
	}
	if t.IDTokenExpiresAt <= 0 {
		return true
	}
	now := time.Now().UnixMilli()
	// 如果 ID Token 已过期或即将过期（提前 5 分钟），需要刷新
	return now+refreshAheadMs >= t.IDTokenExpiresAt
}

// GetUserInfo 获取存储的用户信息
func (t *StoredUAToken) GetUserInfo() map[string]interface{} {
	if t.UserInfo == nil {
		return make(map[string]interface{})
	}
	// 返回 UserInfo 的副本，避免外部修改
	result := make(map[string]interface{}, len(t.UserInfo))
	for k, v := range t.UserInfo {
		result[k] = v
	}
	return result
}

// GetUserInfoString 获取用户信息的字符串表示
func (t *StoredUAToken) GetUserInfoString(key string) string {
	if t.UserInfo == nil {
		return ""
	}
	if v, ok := t.UserInfo[key].(string); ok {
		return v
	}
	return ""
}

// UpdateFromOIDCResult 使用 OIDCFlowResult 更新 Token 信息
func (t *StoredUAToken) UpdateFromOIDCResult(result *OIDCFlowResult) {
	t.AccessToken = result.AccessToken
	t.RefreshToken = result.RefreshToken
	t.IDToken = result.IDToken
	t.Scope = result.Scope
	t.UserOpenId = result.OpenID

	// 设置过期时间
	now := time.Now().UnixMilli()
	t.GrantedAt = now
	t.ExpiresAt = now + int64(result.ExpiresIn)*1000

	// 刷新 Token 过期时间 (假设为长期有效)
	t.RefreshExpiresAt = now + int64(result.RefreshExpiresIn)*1000

	// ID Token 过期时间
	if result.Claims != nil {
		if exp, ok := result.Claims["exp"].(int64); ok {
			t.IDTokenExpiresAt = exp * 1000 // 将秒转换为毫秒
		} else if exp, ok := result.Claims["exp"].(float64); ok {
			t.IDTokenExpiresAt = int64(exp) * 1000 // 将秒转换为毫秒
		}
	}

	// 更新用户信息
	t.UserInfo = map[string]interface{}{
		"email": result.Email,
		"name":  result.UserName,
		"sub":   result.OpenID,
	}
}

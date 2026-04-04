// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// IDTokenVerifier ID Token 验证器
type IDTokenVerifier struct {
	clientID string
	issuer   string
	jwksURL  string
}

// VerificationResult 验证结果
type VerificationResult struct {
	Valid  bool
	Claims map[string]interface{}
	Error  error
}

// NewIDTokenVerifier 创建新的 ID Token 验证器
func NewIDTokenVerifier(clientID, issuer, jwksURL string) *IDTokenVerifier {
	return &IDTokenVerifier{
		clientID: clientID,
		issuer:   issuer,
		jwksURL:  jwksURL,
	}
}

// Verify 验证 ID Token 的完整性和声明
func (v *IDTokenVerifier) Verify(ctx context.Context, idToken string) (*VerificationResult, error) {
	result := &VerificationResult{
		Valid: false,
	}

	if idToken == "" {
		result.Error = fmt.Errorf("ID token is empty")
		return result, result.Error
	}

	// 解析并验证 JWT
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// 从 JWKS 获取公钥
		if v.jwksURL == "" {
			return nil, fmt.Errorf("JWKS URL not configured")
		}

		// TODO: 实现从 JWKS 端点获取公钥的逻辑
		// 目前返回一个错误，表示签名验证未实现
		return nil, fmt.Errorf("public key verification not implemented")
	})

	if err != nil {
		result.Error = fmt.Errorf("failed to parse ID token: %w", err)
		return result, result.Error
	}

	if !token.Valid {
		result.Error = fmt.Errorf("invalid ID token")
		return result, result.Error
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		result.Error = fmt.Errorf("invalid claims format in ID token")
		return result, result.Error
	}

	// 验证 iss (issuer)
	if err := v.verifyIssuer(claims); err != nil {
		result.Error = err
		return result, result.Error
	}

	// 验证 aud (audience)
	if err := v.verifyAudience(claims); err != nil {
		result.Error = err
		return result, result.Error
	}

	// 验证 exp (expiration)
	if err := v.verifyExpiration(claims); err != nil {
		result.Error = err
		return result, result.Error
	}

	// 验证 iat (issued at)
	if err := v.verifyIssuedAt(claims); err != nil {
		result.Error = err
		return result, result.Error
	}

	// 所有验证通过
	result.Valid = true
	result.Claims = convertClaims(claims)

	return result, nil
}

// verifyIssuer 验证 issuer 声明
func (v *IDTokenVerifier) verifyIssuer(claims jwt.MapClaims) error {
	iss, exists := claims["iss"]
	if !exists {
		return fmt.Errorf("missing iss claim in ID token")
	}

	issStr, ok := iss.(string)
	if !ok {
		return fmt.Errorf("invalid iss claim type in ID token")
	}

	if issStr != v.issuer {
		return fmt.Errorf("invalid issuer in ID token: %s, expected: %s", issStr, v.issuer)
	}

	return nil
}

// verifyAudience 验证 audience 声明
func (v *IDTokenVerifier) verifyAudience(claims jwt.MapClaims) error {
	aud, exists := claims["aud"]
	if !exists {
		return fmt.Errorf("missing aud claim in ID token")
	}

	// aud 可能是字符串或字符串数组
	var audStr string
	switch v := aud.(type) {
	case string:
		audStr = v
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("empty aud claim in ID token")
		}
		str, ok := v[0].(string)
		if !ok {
			return fmt.Errorf("invalid aud claim type in ID token")
		}
		audStr = str
	default:
		return fmt.Errorf("invalid aud claim type in ID token")
	}

	if audStr != v.clientID {
		return fmt.Errorf("invalid audience in ID token: %s, expected: %s", audStr, v.clientID)
	}

	return nil
}

// verifyExpiration 验证过期时间
func (v *IDTokenVerifier) verifyExpiration(claims jwt.MapClaims) error {
	exp, exists := claims["exp"]
	if !exists {
		return fmt.Errorf("missing exp claim in ID token")
	}

	expFloat, ok := exp.(float64)
	if !ok {
		return fmt.Errorf("invalid exp claim type in ID token")
	}

	if int64(expFloat) < time.Now().Unix() {
		return fmt.Errorf("ID token has expired")
	}

	return nil
}

// verifyIssuedAt 验证签发时间
func (v *IDTokenVerifier) verifyIssuedAt(claims jwt.MapClaims) error {
	iat, exists := claims["iat"]
	if !exists {
		return fmt.Errorf("missing iat claim in ID token")
	}

	iatFloat, ok := iat.(float64)
	if !ok {
		return fmt.Errorf("invalid iat claim type in ID token")
	}

	if int64(iatFloat) > time.Now().Unix() {
		return fmt.Errorf("invalid iat claim in ID token (future timestamp)")
	}

	return nil
}

// GetClaims 解析 ID Token 的声明部分（不验证签名）
// 这是一个便捷函数，用于快速获取 ID Token 中的用户信息
func GetClaims(idToken string) (map[string]interface{}, error) {
	return parseIDTokenClaims(idToken)
}

// parseIDTokenClaims 解析 ID Token 的声明部分
func parseIDTokenClaims(idToken string) (map[string]interface{}, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID token format: expected 3 parts, got %d", len(parts))
	}

	// 解码 payload 部分（第二部分）
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode ID token payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ID token claims: %w", err)
	}

	return claims, nil
}

// convertClaims 将 jwt.MapClaims 转换为普通的 map[string]interface{}
func convertClaims(claims jwt.MapClaims) map[string]interface{} {
	result := make(map[string]interface{}, len(claims))
	for k, v := range claims {
		result[k] = v
	}
	return result
}

// GetUserInfoFromClaims 从 claims 中提取用户基本信息
func GetUserInfoFromClaims(claims map[string]interface{}) (openID, name, email string) {
	if sub, ok := claims["sub"].(string); ok {
		openID = sub
	}
	if n, ok := claims["name"].(string); ok {
		name = n
	}
	if e, ok := claims["email"].(string); ok {
		email = e
	}
	return
}

// IsIDTokenExpired 检查 ID Token 是否已过期
func IsIDTokenExpired(idToken string) bool {
	claims, err := GetClaims(idToken)
	if err != nil {
		return true // 解析失败视为过期
	}

	exp, exists := claims["exp"]
	if !exists {
		return true // 没有 exp 声明视为过期
	}

	expFloat, ok := exp.(float64)
	if !ok {
		return true
	}

	return int64(expFloat) < time.Now().Unix()
}

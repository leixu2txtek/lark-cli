// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"testing"
	"time"
)

// 测试用的 JWT token（未签名，仅用于测试解析）
// Header: {"alg":"HS256","typ":"JWT"}
// Payload: {"sub":"ou_test123","iss":"https://open.feishu.cn","aud":"cli_test","exp":9999999999,"iat":1000000000,"email":"test@example.com","name":"Test User"}
// 这是一个虚构的 token，仅用于测试解析功能
const testValidIDToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJvdV90ZXN0MTIzIiwiaXNzIjoiaHR0cHM6Ly9vcGVuLmZlaXNodS5jbiIsImF1ZCI6ImNsaV90ZXN0IiwiZXhwIjo5OTk5OTk5OTk5LCJpYXQiOjEwMDAwMDAwMDAsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsIm5hbWUiOiJUZXN0IFVzZXIifQ.test_signature"

// 过期的 token (exp: 1000000000 = 2001-09-09)
const testExpiredIDToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJvdV90ZXN0MTIzIiwiaXNzIjoiaHR0cHM6Ly9vcGVuLmZlaXNodS5jbiIsImF1ZCI6ImNsaV90ZXN0IiwiZXhwIjoxMDAwMDAwMDAwLCJpYXQiOjkwMDAwMDAwMCwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIiwibmFtZSI6IlRlc3QgVXNlciJ9.expired_signature"

// 缺少 exp  claim 的 token
const testNoExpIDToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJvdV90ZXN0MTIzIiwiaXNzIjoiaHR0cHM6Ly9vcGVuLmZlaXNodS5jbiIsImF1ZCI6ImNsaV90ZXN0IiwiaWF0IjoxMDAwMDAwMDAwLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJuYW1lIjoiVGVzdCBVc2VyIn99.no_exp_signature"

// 格式无效的 token
const testInvalidFormatToken = "not.a.valid.jwt.token"

func TestGetClaims_Success(t *testing.T) {
	claims, err := GetClaims(testValidIDToken)
	if err != nil {
		t.Fatalf("GetClaims failed: %v", err)
	}

	if claims["sub"] != "ou_test123" {
		t.Errorf("expected sub 'ou_test123', got '%v'", claims["sub"])
	}
	if claims["iss"] != "https://open.feishu.cn" {
		t.Errorf("expected iss 'https://open.feishu.cn', got '%v'", claims["iss"])
	}
	if claims["aud"] != "cli_test" {
		t.Errorf("expected aud 'cli_test', got '%v'", claims["aud"])
	}
	if claims["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%v'", claims["email"])
	}
	if claims["name"] != "Test User" {
		t.Errorf("expected name 'Test User', got '%v'", claims["name"])
	}
}

func TestGetClaims_InvalidFormat(t *testing.T) {
	_, err := GetClaims(testInvalidFormatToken)
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}

	// 测试空 token
	_, err = GetClaims("")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}

	// 测试只有一部分
	_, err = GetClaims("onlyonepart")
	if err == nil {
		t.Fatal("expected error for single part token, got nil")
	}
}

func TestGetClaims_ExpiredToken(t *testing.T) {
	// GetClaims 不验证过期时间，只解析 payload
	claims, err := GetClaims(testExpiredIDToken)
	if err != nil {
		t.Fatalf("GetClaims should parse expired token: %v", err)
	}
	if claims["sub"] != "ou_test123" {
		t.Errorf("expected sub 'ou_test123', got '%v'", claims["sub"])
	}
}

func TestVerifyIDToken_ValidToken(t *testing.T) {
	// 注意：VerifyIDToken 需要有效的签名验证，目前会返回错误
	// 这里测试验证逻辑的结构
	verifier := NewIDTokenVerifier("cli_test", "https://open.feishu.cn", "")

	_, err := verifier.Verify(context.Background(), testValidIDToken)
	// 由于签名验证未实现，应该返回错误
	if err == nil {
		t.Error("expected error for unimplemented signature verification, got nil")
	}
}

func TestVerifyIDToken_EmptyToken(t *testing.T) {
	verifier := NewIDTokenVerifier("cli_test", "https://open.feishu.cn", "")

	result, err := verifier.Verify(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
	if result.Valid {
		t.Error("expected Valid=false for empty token")
	}
}

func TestIsIDTokenExpired_Valid(t *testing.T) {
	expired := IsIDTokenExpired(testValidIDToken)
	// testValidIDToken 的 exp 是 9999999999，远未过期
	if expired {
		t.Error("expected valid token to not be expired")
	}
}

func TestIsIDTokenExpired_Expired(t *testing.T) {
	expired := IsIDTokenExpired(testExpiredIDToken)
	// testExpiredIDToken 的 exp 是 1000000000，已过期
	if !expired {
		t.Error("expected expired token to be expired")
	}
}

func TestIsIDTokenExpired_InvalidFormat(t *testing.T) {
	expired := IsIDTokenExpired(testInvalidFormatToken)
	// 格式无效的 token 应视为过期
	if !expired {
		t.Error("expected invalid format token to be treated as expired")
	}
}

func TestIsIDTokenExpired_NoExp(t *testing.T) {
	expired := IsIDTokenExpired(testNoExpIDToken)
	// 缺少 exp claim 的 token 应视为过期
	if !expired {
		t.Error("expected token without exp to be treated as expired")
	}
}

func TestGetUserInfoFromClaims(t *testing.T) {
	claims := map[string]interface{}{
		"sub":   "ou_test123",
		"email": "test@example.com",
		"name":  "Test User",
		"iss":   "https://open.feishu.cn",
	}

	openID, name, email := GetUserInfoFromClaims(claims)

	if openID != "ou_test123" {
		t.Errorf("expected openID 'ou_test123', got '%s'", openID)
	}
	if name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", name)
	}
	if email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", email)
	}
}

func TestGetUserInfoFromClaims_MissingFields(t *testing.T) {
	claims := map[string]interface{}{
		"iss": "https://open.feishu.cn",
	}

	openID, name, email := GetUserInfoFromClaims(claims)

	if openID != "" {
		t.Errorf("expected empty openID, got '%s'", openID)
	}
	if name != "" {
		t.Errorf("expected empty name, got '%s'", name)
	}
	if email != "" {
		t.Errorf("expected empty email, got '%s'", email)
	}
}

func TestIDTokenVerifier_VerifyIssuer(t *testing.T) {
	verifier := &IDTokenVerifier{
		clientID: "cli_test",
		issuer:   "https://open.feishu.cn",
	}

	claims := map[string]interface{}{
		"iss": "https://open.feishu.cn",
	}

	err := verifier.verifyIssuer(claims)
	if err != nil {
		t.Errorf("expected no error for matching issuer, got %v", err)
	}

	// 测试不匹配的 issuer
	claims["iss"] = "https://wrong.issuer.com"
	err = verifier.verifyIssuer(claims)
	if err == nil {
		t.Error("expected error for mismatched issuer")
	}

	// 测试缺失 issuer
	delete(claims, "iss")
	err = verifier.verifyIssuer(claims)
	if err == nil {
		t.Error("expected error for missing issuer")
	}
}

func TestIDTokenVerifier_VerifyAudience(t *testing.T) {
	verifier := &IDTokenVerifier{
		clientID: "cli_test",
	}

	// 测试字符串 aud
	claims := map[string]interface{}{
		"aud": "cli_test",
	}
	err := verifier.verifyAudience(claims)
	if err != nil {
		t.Errorf("expected no error for matching audience, got %v", err)
	}

	// 测试数组 aud
	claims["aud"] = []interface{}{"cli_test", "other_aud"}
	err = verifier.verifyAudience(claims)
	if err != nil {
		t.Errorf("expected no error for matching audience in array, got %v", err)
	}

	// 测试不匹配的 aud
	claims["aud"] = "wrong_aud"
	err = verifier.verifyAudience(claims)
	if err == nil {
		t.Error("expected error for mismatched audience")
	}

	// 测试缺失 aud
	delete(claims, "aud")
	err = verifier.verifyAudience(claims)
	if err == nil {
		t.Error("expected error for missing audience")
	}
}

func TestIDTokenVerifier_VerifyExpiration(t *testing.T) {
	verifier := &IDTokenVerifier{}

	// 测试未来过期时间（未过期）
	claims := map[string]interface{}{
		"exp": float64(time.Now().Unix() + 3600),
	}
	err := verifier.verifyExpiration(claims)
	if err != nil {
		t.Errorf("expected no error for valid expiration, got %v", err)
	}

	// 测试过去过期时间（已过期）
	claims["exp"] = float64(time.Now().Unix() - 3600)
	err = verifier.verifyExpiration(claims)
	if err == nil {
		t.Error("expected error for expired token")
	}

	// 测试缺失 exp
	delete(claims, "exp")
	err = verifier.verifyExpiration(claims)
	if err == nil {
		t.Error("expected error for missing exp")
	}
}

func TestIDTokenVerifier_VerifyIssuedAt(t *testing.T) {
	verifier := &IDTokenVerifier{}

	// 测试过去签发时间（有效）
	claims := map[string]interface{}{
		"iat": float64(time.Now().Unix() - 3600),
	}
	err := verifier.verifyIssuedAt(claims)
	if err != nil {
		t.Errorf("expected no error for valid iat, got %v", err)
	}

	// 测试未来签发时间（无效）
	claims["iat"] = float64(time.Now().Unix() + 3600)
	err = verifier.verifyIssuedAt(claims)
	if err == nil {
		t.Error("expected error for future iat")
	}

	// 测试缺失 iat
	delete(claims, "iat")
	err = verifier.verifyIssuedAt(claims)
	if err == nil {
		t.Error("expected error for missing iat")
	}
}

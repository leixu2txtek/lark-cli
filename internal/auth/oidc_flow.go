package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

// OIDCFlowOptions 包含 OIDC 认证选项
type OIDCFlowOptions struct {
	AppID       string
	AppSecret   string
	Domain      string // Accounts domain (for authorization)
	OpenDomain  string // Open domain (for token exchange)
	RedirectURI string
	Timeout     time.Duration
	Scope       string // 可选的 scope 参数，用于请求 API 权限
}

// OIDCFlowResult 包含 OIDC 认证结果
type OIDCFlowResult struct {
	AccessToken      string
	RefreshToken     string
	IDToken          string // OIDC 特有
	ExpiresIn        int
	RefreshExpiresIn int
	Scope            string
	OpenID           string
	UserName         string
	Email            string                 // 用户邮箱
	Claims           map[string]interface{} // JWT 声明
}

// StartOIDCFlow 启动 OIDC 认证流程
func StartOIDCFlow(ctx context.Context, opts *OIDCFlowOptions, httpClient *http.Client, errOut io.Writer) (*OIDCFlowResult, error) {
	// 生成 CSRF 保护的 state 参数
	state := generateState()

	// 从 redirect_uri 解析监听地址
	redirectURL, err := url.Parse(opts.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URI: %w", err)
	}

	// 启动本地 HTTP 服务器来接收回调
	listener, err := net.Listen("tcp", redirectURL.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener on %s: %w", redirectURL.Host, err)
	}
	defer listener.Close()

	// 用于接收授权码的通道
	codeChan := make(chan string, 1)
	errorChan := make(chan string, 1)

	// 创建回调处理器 - 监听所有路径，因为飞书可能回调到 / 或 /callback
	callbackHandler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// 检查是否有错误
		if err := query.Get("error"); err != "" {
			errDesc := query.Get("error_description")
			fmt.Fprintf(errOut, "OIDC flow error: %s - %s\n", err, errDesc)
			errorChan <- fmt.Sprintf("%s: %s", err, errDesc)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Authentication failed: %s - %s", err, errDesc)))
			return
		}

		// 验证 state 参数以防止 CSRF 攻击
		stateParam := query.Get("state")
		if stateParam != state {
			errMsg := "Invalid state parameter - possible CSRF attack"
			fmt.Fprintf(errOut, "%s (received: %s, expected: %s)\n", errMsg, stateParam, state)
			errorChan <- errMsg
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(errMsg))
			return
		}

		// 获取授权码
		code := query.Get("code")
		if code == "" {
			errMsg := "No authorization code received"
			fmt.Fprintf(errOut, "%s\n", errMsg)
			errorChan <- errMsg
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errMsg))
			return
		}

		// 验证成功，发送授权码
		codeChan <- code

		// 发送成功响应给浏览器
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><title>授权成功</title></head><body><h1>授权成功</h1><p>可以关闭此页面并返回终端。</p></body></html>`))
	}

	// 在 goroutine 中启动 HTTP 服务器
	server := &http.Server{
		Handler: http.HandlerFunc(callbackHandler),
	}
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			fmt.Fprintf(errOut, "[WARN] callback server error: %v\n", serveErr)
		}
	}()

	// 构建授权请求 URL
	authURL := fmt.Sprintf("%s/open-apis/authen/v1/user_auth_page_beta", opts.Domain)
	params := url.Values{}
	params.Add("app_id", opts.AppID)
	params.Add("redirect_uri", opts.RedirectURI)
	params.Add("state", state)
	// 如果指定了 scope，则添加到授权请求中（OIDC 模式支持 scope 参数）
	if opts.Scope != "" {
		params.Add("scope", opts.Scope)
	}

	authRequestURL := fmt.Sprintf("%s?%s", authURL, params.Encode())

	fmt.Fprintf(errOut, "Opening browser for authentication...\n")
	fmt.Fprintf(errOut, "If the browser does not open, visit the following URL:\n%s\n", authRequestURL)

	// 尝试打开浏览器
	openBrowser(authRequestURL)

	// 等待授权码或超时
	var authCode string
	select {
	case authCode = <-codeChan:
		// 成功收到授权码
		fmt.Fprintf(errOut, "Received authorization code.\n")
	case errorStr := <-errorChan:
		return nil, fmt.Errorf("authentication error: %s", errorStr)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(opts.Timeout):
		return nil, fmt.Errorf("authentication timed out after %v", opts.Timeout)
	}

	// 停止 HTTP 服务器
	server.Shutdown(context.Background())

	// 使用新的 OIDC API 客户端交换 Token
	client := NewOidcAccessTokenClient(httpClient, opts.OpenDomain)
	client.SetAppCredentials(opts.AppID, opts.AppSecret)
	tokenResp, err := client.CreateAccessToken(ctx, CreateAccessTokenParams{
		AppID:       opts.AppID,
		AppSecret:   opts.AppSecret,
		Code:        authCode,
		RedirectURI: opts.RedirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// 飞书 OIDC API 不直接返回用户信息，需要额外调用用户信息 API
	var openID, userName, email string
	var claims map[string]interface{}

	// 首先尝试从 ID Token 获取用户信息
	if tokenResp.IDToken != "" {
		claims, err = parseIDTokenClaims(tokenResp.IDToken)
		if err != nil {
			fmt.Fprintf(errOut, "Warning: Failed to parse ID token claims: %v\n", err)
		} else {
			// 从 claims 中提取用户信息
			if sub, ok := claims["sub"].(string); ok {
				openID = sub
			}
			if name, ok := claims["name"].(string); ok {
				userName = name
			}
			if emailVal, ok := claims["email"].(string); ok {
				email = emailVal
			}
		}
	}

	// 如果 ID Token 不可用或没有用户信息，调用用户信息 API
	if openID == "" && tokenResp.AccessToken != "" {
		fmt.Fprintf(errOut, "Fetching user info from API...\n")
		userInfo, err := client.GetUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			fmt.Fprintf(errOut, "Warning: Failed to get user info: %v\n", err)
		} else {
			openID = userInfo.OpenID
			userName = userInfo.Name
			email = userInfo.Email
			fmt.Fprintf(errOut, "User info retrieved: %s (%s)\n", userName, openID)
		}
	}

	return &OIDCFlowResult{
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		IDToken:          tokenResp.IDToken,
		ExpiresIn:        tokenResp.ExpiresIn,
		RefreshExpiresIn: tokenResp.RefreshExpiresIn,
		OpenID:           openID,
		UserName:         userName,
		Email:            email,
		Claims:           claims,
	}, nil
}

// openBrowser tries to open the URL in the default browser
func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return err
	}
	return nil
}

// generateState 生成 CSRF 保护的 state 参数
func generateState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if random fails
		return fmt.Sprintf("state_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

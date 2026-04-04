// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	larkauth "github.com/larksuite/cli/internal/auth"
	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/output"
)

// OIDCLoginOptions OIDC 登录选项
type OIDCLoginOptions struct {
	Factory     *cmdutil.Factory
	Ctx         context.Context
	JSON        bool
	Scope       string
	RedirectURI string
	Timeout     int
	Prompt      string // 登录提示参数
	AppID       string
	AppSecret   string
	Domain      string
}

// NewCmdAuthLoginOIDC 创建 OIDC 登录子命令
func NewCmdAuthLoginOIDC(f *cmdutil.Factory, runF func(*OIDCLoginOptions) error) *cobra.Command {
	opts := &OIDCLoginOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "login-oidc",
		Short: "Authenticate using OIDC (OpenID Connect)",
		Long:  "Authenticate with the server using OpenID Connect (OIDC) protocol for enhanced user identity verification.",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()

			if runF != nil {
				return runF(opts)
			}
			return authLoginWithOIDCCommand(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Print session data in JSON format to stdout")
	cmd.Flags().StringVar(&opts.Scope, "scope", "openid email profile", "Scopes to request during login")
	cmd.Flags().StringVar(&opts.RedirectURI, "redirect-uri", "http://localhost:3000/callback", "Callback URL for the OIDC flow")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", 120, "Authentication flow timeout in seconds")
	cmd.Flags().StringVar(&opts.Prompt, "prompt", "", "Specify the prompt options sent to the OIDC provider")
	cmd.Flags().StringVar(&opts.AppID, "app-id", "", "App ID for the OIDC provider")
	cmd.Flags().StringVar(&opts.AppSecret, "app-secret", "", "App secret for the OIDC provider")
	cmd.Flags().StringVar(&opts.Domain, "domain", "", "Domain for the OIDC provider")

	// 标记为必需的参数
	_ = cmd.MarkFlagRequired("app-id")
	_ = cmd.MarkFlagRequired("app-secret")
	_ = cmd.MarkFlagRequired("domain")

	return cmd
}

// authLoginWithOIDCCommand executes the OIDC login flow
func authLoginWithOIDCCommand(opts *OIDCLoginOptions) error {
	f := opts.Factory

	config, err := f.Config()
	if err != nil {
		return err
	}

	// Use command-line provided credentials if given
	appID := opts.AppID
	appSecret := opts.AppSecret

	if appID == "" {
		appID = config.AppID
	}
	if appSecret == "" {
		appSecret = config.AppSecret
	}

	httpClient, err := f.HttpClient()
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	timeout := time.Duration(opts.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(opts.Ctx, timeout)
	defer cancel()

	// OIDC mode does not use scope parameter

	endpoints := core.ResolveEndpoints(config.Brand)

	oidcOpts := &larkauth.OIDCFlowOptions{
		AppID:       appID,
		AppSecret:   appSecret,
		Domain:      endpoints.Accounts, // Authorization endpoint is on Accounts domain
		OpenDomain:  endpoints.Open,     // Token endpoint is on Open domain
		RedirectURI: opts.RedirectURI,
		Timeout:     timeout,
	}

	result, err := larkauth.StartOIDCFlow(ctx, oidcOpts, httpClient, f.IOStreams.ErrOut)
	if err != nil {
		return output.ErrAuth("OIDC authentication failed: %v", err)
	}

	// Store token using the UpdateFromOIDCResult helper
	storedToken := &larkauth.StoredUAToken{
		AppId:      appID,
		UserOpenId: result.OpenID,
	}
	storedToken.UpdateFromOIDCResult(result)

	err = larkauth.SetStoredToken(storedToken)
	if err != nil {
		return output.Errorf(output.ExitInternal, "internal", "failed to store token: %v", err)
	}

	// Update config
	multi, _ := core.LoadMultiAppConfig()
	if multi != nil && len(multi.Apps) > 0 {
		app := &multi.Apps[0]
		for _, oldUser := range app.Users {
			if oldUser.UserOpenId != result.OpenID {
				larkauth.RemoveStoredToken(appID, oldUser.UserOpenId)
			}
		}
		app.Users = []core.AppUser{{UserOpenId: result.OpenID, UserName: result.UserName}}
		if err := core.SaveMultiAppConfig(multi); err != nil {
			return output.Errorf(output.ExitInternal, "internal", "failed to save config: %v", err)
		}
	}

	if opts.JSON {
		// 输出 JSON 格式的结果，包含 ID Token 相关信息
		outputData := map[string]interface{}{
			"status":     "success",
			"app_id":     appID,
			"user_id":    result.OpenID,
			"user_name":  result.UserName,
			"email":      result.Email,
			"scope":      result.Scope,
			"expires_in": result.ExpiresIn,
			"token_type": "Bearer",
		}

		// 添加 ID Token 信息（如果存在）
		if result.IDToken != "" {
			// 只输出解析后的 claims，不输出完整的 token（安全考虑）
			outputData["id_token_claims"] = result.Claims
			outputData["id_token_expires_at"] = storedToken.IDTokenExpiresAt
		}

		output.PrintJson(f.IOStreams.Out, outputData)
	} else {
		output.PrintSuccess(f.IOStreams.ErrOut, fmt.Sprintf("Successfully authenticated using OIDC for app %s", appID))

		// 输出用户详细信息
		fmt.Fprintf(f.IOStreams.Out, "\nUser Information:\n")
		fmt.Fprintf(f.IOStreams.Out, "  User ID:   %s\n", result.OpenID)
		fmt.Fprintf(f.IOStreams.Out, "  Name:      %s\n", result.UserName)
		fmt.Fprintf(f.IOStreams.Out, "  Email:     %s\n", result.Email)

		// 输出 Token 状态
		fmt.Fprintf(f.IOStreams.Out, "\nToken Information:\n")
		fmt.Fprintf(f.IOStreams.Out, "  Access Token expires in:  %d seconds (%.1f minutes)\n",
			result.ExpiresIn, float64(result.ExpiresIn)/60.0)
		fmt.Fprintf(f.IOStreams.Out, "  Refresh Token expires in: %d seconds (%.1f days)\n",
			result.RefreshExpiresIn, float64(result.RefreshExpiresIn)/3600.0/24.0)

		// 输出 ID Token 信息（如果存在）
		if result.IDToken != "" && result.Claims != nil {
			fmt.Fprintf(f.IOStreams.Out, "\nID Token Information:\n")
			if exp, ok := result.Claims["exp"].(float64); ok {
				expTime := time.Unix(int64(exp), 0)
				fmt.Fprintf(f.IOStreams.Out, "  ID Token expires at: %s\n", expTime.Format("2006-01-02 15:04:05"))
			}
			// 输出其他有用的 claims
			if aud, ok := result.Claims["aud"].(string); ok {
				fmt.Fprintf(f.IOStreams.Out, "  Audience: %s\n", aud)
			}
			if iss, ok := result.Claims["iss"].(string); ok {
				fmt.Fprintf(f.IOStreams.Out, "  Issuer: %s\n", iss)
			}
		}
	}

	return nil
}

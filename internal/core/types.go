// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package core

// LarkBrand represents the Lark platform brand.
// "feishu" targets China-mainland, "lark" targets international.
// Any other string is treated as a custom base URL.
type LarkBrand string

const (
	BrandFeishu LarkBrand = "feishu"
	BrandLark   LarkBrand = "lark"
)

// Endpoints holds resolved endpoint URLs for different Lark services.
type Endpoints struct {
	Open     string // e.g. "https://open.feishu.cn"
	Accounts string // e.g. "https://accounts.feishu.cn"
	MCP      string // e.g. "https://mcp.feishu.cn"
}

// ResolveEndpoints resolves endpoint URLs based on brand.
// If brand is a URL (starts with http:// or https://), use it as custom domain.
func ResolveEndpoints(brand LarkBrand) Endpoints {
	// Check if brand is a custom URL
	if len(brand) > 8 && (brand[:7] == "http://" || brand[:8] == "https://") {
		// Custom domain: extract base URL and derive other endpoints
		baseURL := string(brand)
		// Remove trailing slash if present
		if baseURL[len(baseURL)-1] == '/' {
			baseURL = baseURL[:len(baseURL)-1]
		}
		
		// For custom deployments, typically:
		// - Open API: <domain>/open-apis
		// - Accounts: <domain>/accounts (or same domain)
		// - MCP: <domain>/mcp
		return Endpoints{
			Open:     baseURL,
			Accounts: baseURL,  // Often same as Open for private deployments
			MCP:      baseURL,  // Often same as Open for private deployments
		}
	}
	
	// Standard brands
	switch brand {
	case BrandLark:
		return Endpoints{
			Open:     "https://open.larksuite.com",
			Accounts: "https://accounts.larksuite.com",
			MCP:      "https://mcp.larksuite.com",
		}
	default:
		return Endpoints{
			Open:     "https://open.feishu.cn",
			Accounts: "https://accounts.feishu.cn",
			MCP:      "https://mcp.feishu.cn",
		}
	}
}

// ResolveOpenBaseURL returns the Open API base URL for the given brand.
func ResolveOpenBaseURL(brand LarkBrand) string {
	return ResolveEndpoints(brand).Open
}

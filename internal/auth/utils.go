// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"github.com/larksuite/cli/internal/keychain"
)

// Key constructs a standardized key for storing OIDC tokens in keychain
func Key(appID string) string {
	return keychain.LarkCliService + ":" + appID
}

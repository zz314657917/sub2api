package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

// PlatformFromAPIKey returns the billing/quota platform implied by the API key.
func PlatformFromAPIKey(apiKey *APIKey) string {
	if apiKey == nil || apiKey.Group == nil {
		return ""
	}
	return strings.TrimSpace(apiKey.Group.Platform)
}

// QuotaPlatform captures the request-time platform used for user quota billing.
func QuotaPlatform(ctx context.Context, apiKey *APIKey) string {
	if ctx != nil {
		if forcePlatform, _ := ctx.Value(ctxkey.ForcePlatform).(string); strings.TrimSpace(forcePlatform) != "" {
			return strings.TrimSpace(forcePlatform)
		}
	}
	return PlatformFromAPIKey(apiKey)
}

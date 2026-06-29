package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestOpenAIQuotaPlatformUsesForcePlatform(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxkey.ForcePlatform, PlatformAntigravity)
	apiKey := &APIKey{Group: &Group{Platform: PlatformOpenAI}}

	require.Equal(t, PlatformAntigravity, QuotaPlatform(ctx, apiKey))
}

func TestOpenAIQuotaPlatformFallsBackToAPIKeyGroup(t *testing.T) {
	apiKey := &APIKey{Group: &Group{Platform: PlatformOpenAI}}

	require.Equal(t, PlatformOpenAI, QuotaPlatform(context.Background(), apiKey))
}

func TestOpenAIQuotaPlatformEmptyWhenUnavailable(t *testing.T) {
	require.Empty(t, QuotaPlatform(context.Background(), nil))
	require.Empty(t, PlatformFromAPIKey(&APIKey{}))
}

//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyServiceResolveForRequestSkipsMarkedRouteCooldown(t *testing.T) {
	ctx := context.Background()
	defaultGroupID := int64(1)
	fallbackGroupID := int64(2)
	key := &APIKey{
		ID:      44,
		UserID:  7,
		GroupID: &defaultGroupID,
		User: &User{
			ID:            7,
			AllowedGroups: []int64{defaultGroupID, fallbackGroupID},
		},
		Group: &Group{
			ID:       defaultGroupID,
			Platform: PlatformOpenAI,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: defaultGroupID, Priority: 1, Weight: 1, CooldownSeconds: 60, Enabled: true},
			{GroupID: fallbackGroupID, Priority: 2, Weight: 1, CooldownSeconds: 60, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: fallbackGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true},
		},
	}
	svc := &APIKeyService{}

	resolved := svc.ResolveForRequest(ctx, key, "/unknown", "")
	require.NotNil(t, resolved)
	require.Equal(t, defaultGroupID, *resolved.GroupID)

	svc.MarkRouteGroupCooldown(ctx, key, defaultGroupID, 60)
	resolved = svc.ResolveForRequest(ctx, key, "/unknown", "")
	require.NotNil(t, resolved)
	require.Equal(t, fallbackGroupID, *resolved.GroupID)

	svc.ClearRouteGroupCooldown(ctx, key, defaultGroupID)
	resolved = svc.ResolveForRequest(ctx, key, "/unknown", "")
	require.NotNil(t, resolved)
	require.Equal(t, defaultGroupID, *resolved.GroupID)
}

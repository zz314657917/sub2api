//go:build unit

package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyResolveForRequestSelectsPlatformRoute(t *testing.T) {
	defaultGroupID := int64(1)
	openAIGroupID := int64(2)
	geminiGroupID := int64(3)
	key := &APIKey{
		UserID:  42,
		GroupID: &defaultGroupID,
		User: &User{
			ID:                   42,
			UserGroupRPMOverride: apiKeyRoutingIntPtr(25),
			AllowedGroups:        []int64{defaultGroupID, openAIGroupID, geminiGroupID},
		},
		Group: &Group{
			ID:       defaultGroupID,
			Platform: PlatformAnthropic,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: defaultGroupID, Priority: 100, Weight: 1, CooldownSeconds: 30, Enabled: true},
			{GroupID: openAIGroupID, Priority: 100, Weight: 1, CooldownSeconds: 30, Enabled: true},
			{GroupID: geminiGroupID, Priority: 100, Weight: 1, CooldownSeconds: 30, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: openAIGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true},
			{ID: geminiGroupID, Platform: PlatformGemini, Status: StatusActive, Hydrated: true},
		},
	}

	resolved := key.ResolveForRequest("/v1beta/models/gemini-2.5-pro:generateContent", "")

	require.NotNil(t, resolved)
	require.NotSame(t, key, resolved)
	require.Equal(t, geminiGroupID, *resolved.GroupID)
	require.Equal(t, PlatformGemini, resolved.Group.Platform)
	require.Nil(t, resolved.User.UserGroupRPMOverride)
	require.Equal(t, apiKeyRoutingIntPtr(25), key.User.UserGroupRPMOverride)
}

func TestAPIKeyResolveForRequestFallsBackToPriorityWhenPathHasNoPlatform(t *testing.T) {
	defaultGroupID := int64(1)
	openAIGroupID := int64(2)
	key := &APIKey{
		GroupID: &defaultGroupID,
		Group: &Group{
			ID:       defaultGroupID,
			Platform: PlatformAnthropic,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: defaultGroupID, Priority: 200, Weight: 1, Enabled: true},
			{GroupID: openAIGroupID, Priority: 50, Weight: 1, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: openAIGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true},
		},
	}

	resolved := key.ResolveForRequest("/unknown", "")

	require.NotNil(t, resolved)
	require.Equal(t, openAIGroupID, *resolved.GroupID)
	require.Equal(t, PlatformOpenAI, resolved.Group.Platform)
}

func TestAPIKeyResolveForRequestSkipsCoolingRoute(t *testing.T) {
	defaultGroupID := int64(1)
	fallbackGroupID := int64(2)
	key := &APIKey{
		GroupID: &defaultGroupID,
		Group: &Group{
			ID:       defaultGroupID,
			Platform: PlatformOpenAI,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: defaultGroupID, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true},
			{GroupID: fallbackGroupID, Priority: 2, Weight: 1, CooldownSeconds: 30, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: fallbackGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true},
		},
	}

	resolved := key.ResolveForRequestWithGroupSkipper("/unknown", "", func(groupID int64) bool {
		return groupID == defaultGroupID
	})

	require.NotNil(t, resolved)
	require.Equal(t, fallbackGroupID, *resolved.GroupID)
	require.Equal(t, PlatformOpenAI, resolved.Group.Platform)
}

func TestAPIKeyRouteCooldownSeconds(t *testing.T) {
	configuredGroupID := int64(1)
	defaultGroupID := int64(2)
	disabledGroupID := int64(3)
	key := &APIKey{
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: configuredGroupID, CooldownSeconds: 45, Enabled: true},
			{GroupID: defaultGroupID, CooldownSeconds: 0, Enabled: true},
			{GroupID: disabledGroupID, CooldownSeconds: 60, Enabled: false},
		},
	}

	cooldown, ok := key.RouteCooldownSeconds(configuredGroupID)
	require.True(t, ok)
	require.Equal(t, 45, cooldown)

	cooldown, ok = key.RouteCooldownSeconds(defaultGroupID)
	require.True(t, ok)
	require.Equal(t, apiKeyRouteDefaultCooldown, cooldown)

	cooldown, ok = key.RouteCooldownSeconds(disabledGroupID)
	require.False(t, ok)
	require.Zero(t, cooldown)

	cooldown, ok = key.RouteCooldownSeconds(99)
	require.False(t, ok)
	require.Zero(t, cooldown)
}

func apiKeyRoutingIntPtr(v int) *int {
	return &v
}

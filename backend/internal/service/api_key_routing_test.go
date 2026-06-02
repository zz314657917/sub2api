//go:build unit

package service

import (
	"encoding/json"
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
			{GroupID: configuredGroupID, CooldownSeconds: 20, Enabled: true, TextOnly: true},
			{GroupID: configuredGroupID, CooldownSeconds: 90, Enabled: true, ImageOnly: true},
			{GroupID: disabledGroupID, CooldownSeconds: 120, Enabled: false},
			{GroupID: disabledGroupID, CooldownSeconds: 60, Enabled: false},
		},
	}

	cooldown, ok := key.RouteCooldownSeconds(configuredGroupID)
	require.True(t, ok)
	require.Equal(t, 90, cooldown)

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

func TestAPIKeyResolveForModelRequestMatchesModelPattern(t *testing.T) {
	textGroupID := int64(1)
	imageGroupID := int64(2)
	key := &APIKey{
		GroupID: &textGroupID,
		Group: &Group{
			ID:       textGroupID,
			Platform: PlatformOpenAI,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: textGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true},
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"gpt-image-*"}, ImageOnly: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: imageGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, AllowImageGeneration: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-image-2", true)
	require.NotNil(t, resolved)
	require.Equal(t, imageGroupID, *resolved.GroupID)

	resolved = key.ResolveForModelRequest("/v1/responses", "", "gpt-5.5", false)
	require.NotNil(t, resolved)
	require.Equal(t, textGroupID, *resolved.GroupID)
}

func TestAPIKeyResolveForModelRequestImageIntentRequiresImageEnabledOpenAIGroup(t *testing.T) {
	textGroupID := int64(1)
	disabledImageGroupID := int64(2)
	enabledImageGroupID := int64(3)
	key := &APIKey{
		GroupID: &textGroupID,
		Group: &Group{
			ID:       textGroupID,
			Platform: PlatformOpenAI,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: textGroupID, Priority: 1, Weight: 1, Enabled: true},
			{GroupID: disabledImageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
			{GroupID: enabledImageGroupID, Priority: 2, Weight: 1, Enabled: true, ImageOnly: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: disabledImageGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, AllowImageGeneration: false},
			{ID: enabledImageGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, AllowImageGeneration: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/images/generations", "", "gpt-image-2", true)

	require.NotNil(t, resolved)
	require.Equal(t, enabledImageGroupID, *resolved.GroupID)
}

func TestAPIKeyResolveForModelRequestImageOnlyRulePreferredForImageIntent(t *testing.T) {
	genericImageGroupID := int64(1)
	imageOnlyGroupID := int64(2)
	key := &APIKey{
		GroupID: &genericImageGroupID,
		Group: &Group{
			ID:                   genericImageGroupID,
			Platform:             PlatformOpenAI,
			Status:               StatusActive,
			Hydrated:             true,
			AllowImageGeneration: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: genericImageGroupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"gpt-*"}},
			{GroupID: imageOnlyGroupID, Priority: 50, Weight: 1, Enabled: true, ImageOnly: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: imageOnlyGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, AllowImageGeneration: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-image-2", true)

	require.NotNil(t, resolved)
	require.Equal(t, imageOnlyGroupID, *resolved.GroupID)
}

func TestAPIKeyResolveForModelRequestTextOnlyExcludesImageIntent(t *testing.T) {
	textGroupID := int64(1)
	imageGroupID := int64(2)
	key := &APIKey{
		GroupID: &textGroupID,
		Group: &Group{
			ID:       textGroupID,
			Platform: PlatformOpenAI,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: textGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true},
			{GroupID: imageGroupID, Priority: 2, Weight: 1, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: imageGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, AllowImageGeneration: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.5", true)

	require.NotNil(t, resolved)
	require.Equal(t, imageGroupID, *resolved.GroupID)
}

func TestAPIKeyResolveForModelRequestTextOnlyExcludesVideoModels(t *testing.T) {
	textGroupID := int64(10)
	videoGroupID := int64(20)
	key := &APIKey{
		GroupID: &textGroupID,
		Group: &Group{
			ID:       textGroupID,
			Platform: PlatformOpenAI,
			Status:   StatusActive,
			Hydrated: true,
		},
		MultiGroupRouteGroups: []*Group{
			{
				ID:       videoGroupID,
				Platform: PlatformOpenAI,
				Status:   StatusActive,
				Hydrated: true,
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: textGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true},
			{GroupID: videoGroupID, Priority: 2, Weight: 1, Enabled: true, ModelPatterns: []string{"doubao-seedance-*"}},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/videos/generations", "", "doubao-seedance-2.0", false)

	require.NotNil(t, resolved)
	require.NotNil(t, resolved.GroupID)
	require.Equal(t, videoGroupID, *resolved.GroupID)
}

func TestAPIKeyResolveForModelRequestFallsBackWhenNoModelRuleMatches(t *testing.T) {
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
			{GroupID: defaultGroupID, Priority: 20, Weight: 1, Enabled: true, ModelPatterns: []string{"claude-*"}},
			{GroupID: fallbackGroupID, Priority: 10, Weight: 1, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{
			{ID: fallbackGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.5", false)

	require.NotNil(t, resolved)
	require.Equal(t, fallbackGroupID, *resolved.GroupID)
}

func TestAPIKeyMultiGroupRouteOldJSONDefaultsRemainCompatible(t *testing.T) {
	var route domain.APIKeyMultiGroupRoute
	err := json.Unmarshal([]byte(`{"group_id":7,"priority":1,"weight":2,"cooldown_seconds":30,"enabled":true}`), &route)

	require.NoError(t, err)
	require.Equal(t, int64(7), route.GroupID)
	require.Empty(t, route.ModelPatterns)
	require.False(t, route.ImageOnly)
	require.False(t, route.TextOnly)
}

func apiKeyRoutingIntPtr(v int) *int {
	return &v
}

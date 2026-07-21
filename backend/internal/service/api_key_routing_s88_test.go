package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestS88ModelAwareRouteRejectsImageDefaultForTextRequest(t *testing.T) {
	groupID := int64(1)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:                   groupID,
			Platform:             PlatformOpenAI,
			Status:               StatusActive,
			Hydrated:             true,
			RoutingScope:         GroupRoutingScopeImage,
			AllowImageGeneration: true,
			ModelMatchPatterns:   []string{"*"},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: groupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Nil(t, resolved)
}

func TestS88ModelAwareRouteRejectsExplicitImageOnlyDefaultForTextRequest(t *testing.T) {
	groupID := int64(1)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:                 groupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"*"},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: groupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Nil(t, resolved)
}

func TestS88ModelAwareRouteRejectsMismatchedDefaultModelPattern(t *testing.T) {
	groupID := int64(1)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:                 groupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"claude-*"},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: groupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"claude-*"}},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Nil(t, resolved)
}

func TestS88ModelAwareRouteKeepsCompatibleTextDefaultFallback(t *testing.T) {
	defaultGroupID := int64(1)
	imageGroupID := int64(2)
	key := &APIKey{
		GroupID: &defaultGroupID,
		Group: &Group{
			ID:                 defaultGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"*"},
		},
		MultiGroupRouteGroups: []*Group{
			{
				ID:                   imageGroupID,
				Platform:             PlatformOpenAI,
				Status:               StatusActive,
				Hydrated:             true,
				RoutingScope:         GroupRoutingScopeImage,
				AllowImageGeneration: true,
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Same(t, key, resolved)
	require.Equal(t, defaultGroupID, *resolved.GroupID)
}

func TestS88ModelAwareRouteKeepsMatchingConfiguredRoute(t *testing.T) {
	defaultGroupID := int64(1)
	textGroupID := int64(2)
	key := &APIKey{
		GroupID: &defaultGroupID,
		Group: &Group{
			ID:                   defaultGroupID,
			Platform:             PlatformOpenAI,
			Status:               StatusActive,
			Hydrated:             true,
			RoutingScope:         GroupRoutingScopeImage,
			AllowImageGeneration: true,
		},
		MultiGroupRouteGroups: []*Group{
			{
				ID:                 textGroupID,
				Platform:           PlatformOpenAI,
				Status:             StatusActive,
				Hydrated:           true,
				RoutingScope:       GroupRoutingScopeInference,
				ModelMatchPatterns: []string{"gpt-*"},
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: defaultGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
			{GroupID: textGroupID, Priority: 2, Weight: 1, Enabled: true, TextOnly: true, ModelPatterns: []string{"claude-*"}},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.NotNil(t, resolved)
	require.NotSame(t, key, resolved)
	require.Equal(t, textGroupID, *resolved.GroupID)
	require.Same(t, key.MultiGroupRouteGroups[0], resolved.Group)
}

func TestS88ModelAwareRouteRejectsWrongPlatformDefault(t *testing.T) {
	defaultGroupID := int64(1)
	imageGroupID := int64(2)
	key := &APIKey{
		GroupID: &defaultGroupID,
		Group: &Group{
			ID:           defaultGroupID,
			Platform:     PlatformAnthropic,
			Status:       StatusActive,
			Hydrated:     true,
			RoutingScope: GroupRoutingScopeInference,
		},
		MultiGroupRouteGroups: []*Group{
			{
				ID:                   imageGroupID,
				Platform:             PlatformOpenAI,
				Status:               StatusActive,
				Hydrated:             true,
				RoutingScope:         GroupRoutingScopeImage,
				AllowImageGeneration: true,
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Nil(t, resolved)
}

func TestS88ModelAwareRouteLeavesSingleGroupKeyUnchanged(t *testing.T) {
	groupID := int64(1)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:                 groupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"*"},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Same(t, key, resolved)
}

func TestS88PreBodyRouteKeepsDefaultFallback(t *testing.T) {
	groupID := int64(1)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:                   groupID,
			Platform:             PlatformOpenAI,
			Status:               StatusActive,
			Hydrated:             true,
			RoutingScope:         GroupRoutingScopeImage,
			AllowImageGeneration: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: groupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
	}

	resolved := key.ResolveForRequest("/v1/responses", "")

	require.Same(t, key, resolved)
}

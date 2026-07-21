package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyRoutingS93FallsBackToCompatibleBaseGroupAfterDedicatedRoutesMiss(t *testing.T) {
	fallbackGroupID := int64(1)
	dedicatedGroupID := int64(2)
	key := &APIKey{
		GroupID: &fallbackGroupID,
		Group: &Group{
			ID:                 fallbackGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"gpt-*"},
		},
		MultiGroupRouteGroups: []*Group{{
			ID:                 dedicatedGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"claude-*"},
		}},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:  dedicatedGroupID,
			Priority: 1,
			Weight:   1,
			Enabled:  true,
		}},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Same(t, key, resolved)
	require.Equal(t, fallbackGroupID, *resolved.GroupID)
}

func TestAPIKeyRoutingS93RejectsFallbackThatFailsGroupModelRule(t *testing.T) {
	fallbackGroupID := int64(1)
	dedicatedGroupID := int64(2)
	key := &APIKey{
		GroupID: &fallbackGroupID,
		Group: &Group{
			ID:                 fallbackGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"claude-*"},
		},
		MultiGroupRouteGroups: []*Group{{
			ID:                 dedicatedGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"gemini-*"},
		}},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:  dedicatedGroupID,
			Priority: 1,
			Weight:   1,
			Enabled:  true,
		}},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false)

	require.Nil(t, resolved)
}

func TestAPIKeyRoutingS93RejectsInactiveFallback(t *testing.T) {
	fallbackGroupID := int64(1)
	key := &APIKey{
		GroupID: &fallbackGroupID,
		Group: &Group{
			ID:                 fallbackGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusDisabled,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"*"},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:  2,
			Priority: 1,
			Weight:   1,
			Enabled:  true,
		}},
	}

	require.Nil(t, key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false))
}

func TestAPIKeyRoutingS93ChecksModelRuleWhenFallbackIsTheOnlyGroup(t *testing.T) {
	fallbackGroupID := int64(1)
	key := &APIKey{
		GroupID: &fallbackGroupID,
		Group: &Group{
			ID:                 fallbackGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"claude-*"},
		},
	}

	require.Nil(t, key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", false))
}

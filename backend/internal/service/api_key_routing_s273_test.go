package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
)

func TestS273LegacySingleGroupEmptyRulesRemainRoutable(t *testing.T) {
	groupID := int64(273)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:           groupID,
			Platform:     PlatformOpenAI,
			Status:       StatusActive,
			Hydrated:     true,
			RoutingScope: GroupRoutingScopeInference,
			// Empty rules represent rows created before S91 introduced the
			// administrator-owned model matching field.
			ModelMatchPatterns: nil,
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.6", false)
	if resolved != key {
		t.Fatalf("legacy single-group key should remain routable, got %#v", resolved)
	}
}

func TestS273ConfiguredSingleGroupEmptyRuleCompatibilityDoesNotRelaxMultiGroup(t *testing.T) {
	defaultGroupID := int64(274)
	key := &APIKey{
		GroupID: &defaultGroupID,
		Group: &Group{
			ID:           defaultGroupID,
			Platform:     PlatformOpenAI,
			Status:       StatusActive,
			Hydrated:     true,
			RoutingScope: GroupRoutingScopeInference,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:  defaultGroupID,
			Priority: 1,
			Weight:   1,
			Enabled:  true,
		}},
		MultiGroupRouteGroups: []*Group{{
			ID:           defaultGroupID,
			Platform:     PlatformOpenAI,
			Status:       StatusActive,
			Hydrated:     true,
			RoutingScope: GroupRoutingScopeInference,
		}},
	}

	if resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.6", false); resolved != nil {
		t.Fatalf("multi-group empty rules must remain fail-closed, got %#v", resolved)
	}
}

func TestS273PinnedSingleGroupEmptyRulesRemainFailClosed(t *testing.T) {
	groupID := int64(275)
	key := &APIKey{
		GroupID:         &groupID,
		PinnedAccountID: 991,
		Group: &Group{
			ID:           groupID,
			Platform:     PlatformOpenAI,
			Status:       StatusActive,
			Hydrated:     true,
			RoutingScope: GroupRoutingScopeInference,
		},
	}

	if resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.6", false); resolved != nil {
		t.Fatalf("pinned empty rules must remain fail-closed, got %#v", resolved)
	}
}

func TestS273LegacySingleGroupDisabledImageDefersPermissionToHandler(t *testing.T) {
	groupID := int64(276)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:           groupID,
			Platform:     PlatformOpenAI,
			Status:       StatusActive,
			Hydrated:     true,
			RoutingScope: GroupRoutingScopeInference,
			// The gateway handler owns the image permission response.
			AllowImageGeneration: false,
		},
	}

	if resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", true); resolved != key {
		t.Fatalf("disabled legacy image request should reach handler permission gate, got %#v", resolved)
	}
}

func TestS273ConfiguredLegacyRuleStillRejectsImageModelMismatch(t *testing.T) {
	groupID := int64(278)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:                   groupID,
			Platform:             PlatformOpenAI,
			Status:               StatusActive,
			Hydrated:             true,
			RoutingScope:         GroupRoutingScopeInference,
			ModelMatchPatterns:   []string{"gpt-image-*"},
			AllowImageGeneration: false,
		},
	}

	if resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5.4", true); resolved != nil {
		t.Fatalf("configured legacy rule must reject an image model mismatch, got %#v", resolved)
	}
}

func TestS273LegacySingleGroupUsesConfiguredModelWithoutEndpointPlatformFilter(t *testing.T) {
	groupID := int64(277)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:                 groupID,
			Platform:           PlatformGrok,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"grok-*"},
		},
	}

	if resolved := key.ResolveForModelRequest("/v1/responses", "", "grok-4.5", false); resolved != key {
		t.Fatalf("legacy single-group Grok key should use its configured model rule, got %#v", resolved)
	}
}

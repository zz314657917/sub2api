package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestS91NormalizeGroupModelMatchPatterns(t *testing.T) {
	patterns := NormalizeGroupModelMatchPatterns([]string{" GPT-* ", "gpt-*", "Claude-*", "", "  "})
	require.Equal(t, []string{"claude-*", "gpt-*"}, patterns)

	all, err := ValidateGroupModelMatchPatterns([]string{" * ", "*"})
	require.NoError(t, err)
	require.Equal(t, []string{"*"}, all)
	_, err = ValidateGroupModelMatchPatterns([]string{" ", ""})
	require.ErrorIs(t, err, ErrGroupModelMatchPatternsRequired)
}

func TestS91GroupMatchesModelExactAndWildcard(t *testing.T) {
	group := &Group{ModelMatchPatterns: []string{"gpt-4o", "claude-*", "*image*"}}
	require.True(t, group.MatchesModel("GPT-4O"))
	require.True(t, group.MatchesModel("claude-sonnet-4-6"))
	require.True(t, group.MatchesModel("provider-image-model"))
	require.False(t, group.MatchesModel("gpt-5"))
	allGroup := &Group{ModelMatchPatterns: []string{"*"}}
	require.True(t, allGroup.MatchesModel("any-model"))
}

func TestS91ModelRoutingFallsThroughHigherPriorityMismatch(t *testing.T) {
	defaultGroupID := int64(1)
	fallbackGroupID := int64(2)
	key := &APIKey{
		GroupID: &defaultGroupID,
		Group: &Group{
			ID:                 defaultGroupID,
			Platform:           PlatformOpenAI,
			Status:             StatusActive,
			Hydrated:           true,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"claude-*"},
		},
		MultiGroupRouteGroups: []*Group{
			{
				ID:                 fallbackGroupID,
				Platform:           PlatformOpenAI,
				Status:             StatusActive,
				Hydrated:           true,
				RoutingScope:       GroupRoutingScopeInference,
				ModelMatchPatterns: []string{"gpt-*"},
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			// The legacy route patterns are intentionally contradictory. S91
			// routing must use the administrator-owned group patterns instead.
			{GroupID: defaultGroupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"gpt-*"}},
			{GroupID: fallbackGroupID, Priority: 2, Weight: 1, Enabled: true, ModelPatterns: []string{"claude-*"}},
		},
	}

	resolved := key.ResolveForModelRequest("/v1/responses", "", "gpt-5", false)
	require.NotNil(t, resolved)
	require.Equal(t, fallbackGroupID, *resolved.GroupID)
}

func TestS91ModelRoutingReturnsNilWhenNoGroupMatches(t *testing.T) {
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
			{GroupID: groupID, Priority: 1, Weight: 1, Enabled: true},
		},
	}

	require.Nil(t, key.ResolveForModelRequest("/v1/responses", "", "gpt-5", false))
}

func TestS91SamePriorityRoutesUseWeights(t *testing.T) {
	groupA := &Group{ID: 1, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, RoutingScope: GroupRoutingScopeInference, ModelMatchPatterns: []string{"*"}}
	groupB := &Group{ID: 2, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, RoutingScope: GroupRoutingScopeInference, ModelMatchPatterns: []string{"*"}}
	candidates := []apiKeyRouteCandidate{
		{group: groupA, priority: 1, weight: 9},
		{group: groupB, priority: 1, weight: 1},
	}

	counts := map[int64]int{}
	for i := 0; i < 2000; i++ {
		counts[selectBestRouteGroup(candidates).ID]++
	}
	require.Greater(t, counts[groupA.ID], counts[groupB.ID])
	require.Positive(t, counts[groupB.ID])
}

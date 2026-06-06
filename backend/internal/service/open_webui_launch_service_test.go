package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenWebUIAPIKeyOptionFromAPIKeyUsesRouteImageGroup(t *testing.T) {
	textGroupID := int64(10)
	key := &APIKey{
		ID:      7,
		Key:     "sk-route-image",
		Name:    "smart route",
		Status:  StatusAPIKeyActive,
		GroupID: &textGroupID,
		Group: &Group{
			ID:                   textGroupID,
			Name:                 "text",
			Platform:             PlatformAnthropic,
			Status:               StatusActive,
			AllowImageGeneration: false,
		},
		MultiGroupRouteGroups: []*Group{
			{
				ID:                   11,
				Name:                 "image",
				Platform:             PlatformOpenAI,
				Status:               StatusActive,
				AllowImageGeneration: true,
			},
		},
	}

	option, ok := openWebUIAPIKeyOptionFromAPIKey(key)

	require.True(t, ok)
	require.Equal(t, int64(11), option.GroupID)
	require.Equal(t, "image", option.GroupName)
	require.Equal(t, PlatformOpenAI, option.GroupPlatform)
	require.True(t, option.SupportsImageGeneration)
}

func TestOpenWebUIAPIKeyOptionFromAPIKeyRejectsKeyWithoutImageRoute(t *testing.T) {
	textGroupID := int64(10)
	key := &APIKey{
		ID:      7,
		Key:     "sk-text-only",
		Name:    "text only",
		Status:  StatusAPIKeyActive,
		GroupID: &textGroupID,
		Group: &Group{
			ID:                   textGroupID,
			Name:                 "text",
			Platform:             PlatformAnthropic,
			Status:               StatusActive,
			AllowImageGeneration: false,
		},
	}

	_, ok := openWebUIAPIKeyOptionFromAPIKey(key)

	require.False(t, ok)
}

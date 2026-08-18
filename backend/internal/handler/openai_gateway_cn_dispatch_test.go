package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAllowOpenAICompatibleMessagesDispatch_CNProvidersExempt(t *testing.T) {
	require.True(t, allowOpenAICompatibleMessagesDispatch(nil))

	for _, platform := range []string{service.PlatformKimi, service.PlatformZhipu, service.PlatformDeepseek, service.PlatformGrok} {
		t.Run(platform, func(t *testing.T) {
			apiKey := &service.APIKey{Group: &service.Group{Platform: platform, AllowMessagesDispatch: false}}
			require.True(t, allowOpenAICompatibleMessagesDispatch(apiKey))
		})
	}

	t.Run("openai remains controlled", func(t *testing.T) {
		openaiOff := &service.APIKey{Group: &service.Group{Platform: service.PlatformOpenAI, AllowMessagesDispatch: false}}
		require.False(t, allowOpenAICompatibleMessagesDispatch(openaiOff))
		openaiOn := &service.APIKey{Group: &service.Group{Platform: service.PlatformOpenAI, AllowMessagesDispatch: true}}
		require.True(t, allowOpenAICompatibleMessagesDispatch(openaiOn))
	})
}

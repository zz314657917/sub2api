package repository

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUpstreamFiveSchedulerCapabilityProjection(t *testing.T) {
	for _, tc := range []struct {
		name         string
		capabilities any
		chat, embed  bool
	}{
		{"missing", nil, true, true},
		{"strings", []string{"embeddings"}, false, true},
		{"json_array", []any{"chat_completions"}, true, false},
		{"explicit_deny", map[string]any{"chat_completions": true, "embeddings": false}, true, false},
		{"deny_all", map[string]bool{"chat_completions": false, "embeddings": false}, false, false},
		{"explicit_allow", map[string]any{"chat_completions": true, "embeddings": true}, true, true},
		{"empty_legacy", []string{}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			credentials := map[string]any{
				"model_mapping": map[string]any{"client": "upstream"}, "api_key": "kept-existing-key",
				"project_id": "project", "oauth_type": "type",
				"access_token": "excluded-access", "refresh_token": "excluded-refresh", "unrelated": "excluded",
			}
			if tc.capabilities != nil {
				credentials["openai_capabilities"] = tc.capabilities
			}
			account := service.Account{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: credentials}
			// Exercise the production metadata projection and its actual JSON representation.
			payload, err := json.Marshal(buildSchedulerMetadataAccount(account))
			require.NoError(t, err)
			var metadata service.Account
			require.NoError(t, json.Unmarshal(payload, &metadata))
			for _, key := range []string{"model_mapping", "api_key", "project_id", "oauth_type"} {
				require.Equal(t, credentials[key], metadata.Credentials[key])
			}
			for _, key := range []string{"access_token", "refresh_token", "unrelated"} {
				require.NotContains(t, metadata.Credentials, key)
			}
			for capability, want := range map[service.OpenAIEndpointCapability]bool{
				service.OpenAIEndpointCapabilityChatCompletions: tc.chat,
				service.OpenAIEndpointCapabilityEmbeddings:      tc.embed,
			} {
				require.Equal(t, want, account.SupportsOpenAIEndpointCapability(capability))
				require.Equal(t, want, metadata.SupportsOpenAIEndpointCapability(capability))
			}
			if tc.capabilities != nil {
				require.Contains(t, metadata.Credentials, "openai_capabilities")
			} else {
				require.NotContains(t, metadata.Credentials, "openai_capabilities")
			}
		})
	}
}

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIChannelRestriction_NativeCompactionUsesForwardModelWithoutCompactMapping(t *testing.T) {
	account := &Account{Credentials: map[string]any{
		"model_mapping":         map[string]any{"gpt-5.6": "channel-forward"},
		"compact_model_mapping": map[string]any{"channel-forward": "legacy-compact"},
	}}
	ctx := WithOpenAIForwardModel(context.Background(), "channel-forward", false)
	forwardModel, ok := openAIForwardModelFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "channel-forward", forwardModel.model)
	require.False(t, forwardModel.useCompactModelMapping)
	require.Equal(t, "channel-forward", resolveOpenAIAccountUpstreamModelForRequest(account, forwardModel.model, forwardModel.useCompactModelMapping))
}

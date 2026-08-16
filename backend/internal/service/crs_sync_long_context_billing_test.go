package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeCRSOpenAILongContextBillingExtraDefaultsPreservesAndRejectsMalformedValue(t *testing.T) {
	created, err := mergeCRSOpenAILongContextBillingExtra(nil, nil)
	require.NoError(t, err)
	require.Equal(t, false, created[openAILongContextBillingEnabledKey])

	existing := &Account{Platform: PlatformOpenAI, Extra: map[string]any{
		openAILongContextBillingEnabledKey: true,
	}}
	preserved, err := mergeCRSOpenAILongContextBillingExtra(existing, map[string]any{"custom": "value"})
	require.NoError(t, err)
	require.Equal(t, true, preserved[openAILongContextBillingEnabledKey])
	require.Equal(t, "value", preserved["custom"])

	updated, err := mergeCRSOpenAILongContextBillingExtra(existing, map[string]any{
		openAILongContextBillingEnabledKey: false,
	})
	require.NoError(t, err)
	require.Equal(t, false, updated[openAILongContextBillingEnabledKey])

	_, err = mergeCRSOpenAILongContextBillingExtra(existing, map[string]any{
		openAILongContextBillingEnabledKey: "false",
	})
	require.Error(t, err)
}

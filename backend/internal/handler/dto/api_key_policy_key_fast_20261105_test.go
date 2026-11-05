package dto

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestKeyFast20261105DTOJSON(t *testing.T) {
	require.Nil(t, APIKeyFromService(nil))
	for _, policy := range []string{"", "passthrough", "force", "filter"} {
		wire, err := json.Marshal(APIKeyFromService(&service.APIKey{ID: 7, OpenAIFastPolicy: policy}))
		require.NoError(t, err)
		require.True(t, gjson.GetBytes(wire, "openai_fast_policy").Exists())
		require.Equal(t, service.NormalizeAPIKeyOpenAIFastPolicy(policy), gjson.GetBytes(wire, "openai_fast_policy").String())
	}
}

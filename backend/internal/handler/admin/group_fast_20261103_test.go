package admin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupFast20261103RequestBinding(t *testing.T) {
	var create CreateGroupRequest
	require.NoError(t, json.Unmarshal([]byte(`{"force_openai_fast":true}`), &create))
	require.True(t, create.ForceOpenAIFast)
	var omitted UpdateGroupRequest
	require.NoError(t, json.Unmarshal([]byte(`{}`), &omitted))
	require.Nil(t, omitted.ForceOpenAIFast)
	for _, value := range []bool{false, true} {
		raw, err := json.Marshal(map[string]bool{"force_openai_fast": value})
		require.NoError(t, err)
		var update UpdateGroupRequest
		require.NoError(t, json.Unmarshal(raw, &update))
		require.NotNil(t, update.ForceOpenAIFast)
		require.Equal(t, value, *update.ForceOpenAIFast)
	}
}

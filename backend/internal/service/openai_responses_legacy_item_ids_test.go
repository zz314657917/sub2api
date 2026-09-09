package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestStripLegacyResponsesFunctionItemIDs(t *testing.T) {
	body := []byte(`{"metadata":{"n":9007199254740993},"input":[{"type":"function_call","id":"item_7526c3dc5ad07ada6171bc29","call_id":"call_1","name":"exec","arguments":"{}"},{"type":"function_call_output","call_id":"call_1","output":"ok"},{"type":"function_call","id":"fc_valid","call_id":"call_2"},{"type":"message","id":"item_message"},{"type":"item_reference","id":"item_reference"},{"type":"function_call","id":"item_second","call_id":"call_3"}]}`)
	out, changed, err := stripLegacyResponsesFunctionItemIDs(body)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(out, "input.0.id").Exists())
	require.False(t, gjson.GetBytes(out, "input.5.id").Exists())
	require.Equal(t, "call_1", gjson.GetBytes(out, "input.0.call_id").String())
	require.Equal(t, "call_1", gjson.GetBytes(out, "input.1.call_id").String())
	require.Equal(t, "{}", gjson.GetBytes(out, "input.0.arguments").String())
	require.Equal(t, "fc_valid", gjson.GetBytes(out, "input.2.id").String())
	require.Equal(t, "item_message", gjson.GetBytes(out, "input.3.id").String())
	require.Equal(t, "item_reference", gjson.GetBytes(out, "input.4.id").String())
	require.Contains(t, string(out), "9007199254740993")
	again, changed, err := stripLegacyResponsesFunctionItemIDs(out)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, out, again)
	for _, raw := range []string{`{"input":"hello"}`, `{"messages":[]}`, `{"input":[{"type":"function_call","id":"fc_valid"}]}`} {
		out, changed, err := stripLegacyResponsesFunctionItemIDs([]byte(raw))
		require.NoError(t, err)
		require.False(t, changed)
		require.Equal(t, raw, string(out))
	}
}

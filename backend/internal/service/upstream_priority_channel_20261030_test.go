package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamPriorityChannelSearchMerge(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{"omitted block", `null`, `{"web_search_emulation":{"anthropic":true,"openai":true,"future":{"mode":"keep"}},"old_only":true}`},
		{"omitted key", `{"other":{"replacement":true}}`, `{"web_search_emulation":{"anthropic":true,"openai":true,"future":{"mode":"keep"}},"other":{"replacement":true}}`},
		{"empty block", `{}`, `{"web_search_emulation":{"anthropic":true,"openai":true,"future":{"mode":"keep"}}}`},
		{"incoming wins", `{"web_search_emulation":{"anthropic":false,"new_platform":true},"other":{"replacement":true}}`, `{"web_search_emulation":{"anthropic":false,"openai":true,"future":{"mode":"keep"},"new_platform":true},"other":{"replacement":true}}`},
		{"clear empty map", `{"web_search_emulation":{}}`, `{"web_search_emulation":{}}`},
		{"clear null", `{"web_search_emulation":null}`, `{"web_search_emulation":null}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var existing, incoming, want map[string]any
			require.NoError(t, json.Unmarshal([]byte(`{"web_search_emulation":{"anthropic":true,"openai":true,"future":{"mode":"keep"}},"old_only":true}`), &existing))
			require.NoError(t, json.Unmarshal([]byte(tc.input), &incoming))
			require.NoError(t, json.Unmarshal([]byte(tc.want), &want))
			channel := &Channel{FeaturesConfig: existing}
			require.NoError(t, (&ChannelService{}).applyUpdateInput(context.Background(), channel, &UpdateChannelInput{FeaturesConfig: incoming}))
			require.Equal(t, want, channel.FeaturesConfig)
			incomingAfter, err := json.Marshal(incoming)
			require.NoError(t, err)
			require.JSONEq(t, tc.input, string(incomingAfter), "update must not mutate the caller's config")
		})
	}
}

func TestUpstreamPriorityChannelSearchMapClones(t *testing.T) {
	for _, supplied := range []bool{false, true} {
		t.Run(map[bool]string{false: "inherited", true: "merged"}[supplied], func(t *testing.T) {
			existingSearch := map[string]any{"openai": true, "anthropic": true}
			existing := map[string]any{"web_search_emulation": existingSearch, "old_only": true}
			incomingSearch := map[string]any{"anthropic": false}
			incoming := map[string]any{"other": map[string]any{"new": true}}
			if supplied {
				incoming["web_search_emulation"] = incomingSearch
			}
			channel := &Channel{FeaturesConfig: existing}
			require.NoError(t, (&ChannelService{}).applyUpdateInput(context.Background(), channel, &UpdateChannelInput{FeaturesConfig: incoming}))
			search, ok := channel.FeaturesConfig["web_search_emulation"].(map[string]any)
			require.True(t, ok)
			search["openai"] = false
			search["anthropic"] = "changed-result"
			channel.FeaturesConfig["result_only"] = true
			require.Equal(t, true, existingSearch["openai"])
			require.Equal(t, true, existingSearch["anthropic"])
			require.Equal(t, false, incomingSearch["anthropic"])
			require.NotContains(t, incoming, "result_only")
			require.NotContains(t, existing, "result_only")
			require.NotContains(t, channel.FeaturesConfig, "old_only", "unrelated config retains whole-block replacement semantics")
		})
	}
}

func TestUpstreamPriorityChannelSearchWithoutExistingConfig(t *testing.T) {
	for _, existing := range []map[string]any{nil, {}, {"other": true}} {
		input := map[string]any{"web_search_emulation": map[string]any{"anthropic": false}, "new": true}
		channel := &Channel{FeaturesConfig: existing}
		require.NoError(t, (&ChannelService{}).applyUpdateInput(context.Background(), channel, &UpdateChannelInput{FeaturesConfig: input}))
		require.Equal(t, input, channel.FeaturesConfig)
	}
}

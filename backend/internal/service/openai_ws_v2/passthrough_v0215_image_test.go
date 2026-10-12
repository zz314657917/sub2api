package openai_ws_v2

import (
	"context"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

func TestV0215WSImageUsageForms(t *testing.T) {
	for _, tc := range []struct {
		name, body    string
		input, output int
	}{
		{"responses", `{"response":{"usage":{"input_tokens":120,"output_tokens":70,"input_tokens_details":{"image_tokens":40},"output_tokens_details":{"image_tokens":55}}}}`, 40, 55},
		{"chat-details", `{"response":{"usage":{"prompt_tokens":120,"completion_tokens":70,"prompt_tokens_details":{"image_tokens":41},"completion_tokens_details":{"image_tokens":56}}}}`, 41, 56},
		{"managed-fallback", `{"response":{"usage":{"input_tokens":120,"output_tokens":70},"tool_usage":{"image_gen":{"input_tokens_details":{"image_tokens":42},"output_tokens_details":{"image_tokens":57}}}}}`, 42, 57},
		{"usage-precedence", `{"response":{"usage":{"input_tokens":120,"output_tokens":70,"input_tokens_details":{"image_tokens":43},"output_tokens_details":{"image_tokens":58}},"tool_usage":{"image_gen":{"input_tokens_details":{"image_tokens":999},"output_tokens_details":{"image_tokens":999}}}}}`, 43, 58},
		{"top-level", `{"usage":{"prompt_tokens":120,"completion_tokens":70},"tool_usage":{"image_gen":{"input_tokens_details":{"image_tokens":44},"output_tokens_details":{"image_tokens":59}}}}`, 44, 59},
		{"negative", `{"response":{"usage":{"input_tokens":120,"output_tokens":70,"input_tokens_details":{"image_tokens":-8},"output_tokens_details":{"image_tokens":-9}}}}`, 0, 0},
		{"invalid-totals", `{"response":{"usage":{"input_tokens":"oops","output_tokens":70,"input_tokens_details":{"image_tokens":40}}}}`, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &relayState{}
			parsed := parseUsageAndAccumulate(state, []byte(tc.body), "response.completed", nil)
			require.Equal(t, tc.input, parsed.ImageInputTokens)
			require.Equal(t, tc.output, parsed.ImageOutputTokens)
			require.Equal(t, parsed, state.usage)
		})
	}
}

func TestV0215WSImageMultiTurnDuplicateTerminal(t *testing.T) {
	first := []byte(`{"type":"response.completed","response":{"id":"r1","usage":{"input_tokens":120,"output_tokens":70,"input_tokens_details":{"image_tokens":40,"cached_tokens":5,"cache_write_tokens":3}},"tool_usage":{"image_gen":{"output_tokens_details":{"image_tokens":55}}}}}`)
	duplicate := []byte(`{"type":"response.done","response":{"id":"r1","usage":{"input_tokens":120,"output_tokens":70,"input_tokens_details":{"image_tokens":40,"cached_tokens":5,"cache_write_tokens":3}}}}`)
	second := []byte(`{"type":"response.completed","response":{"id":"r2","usage":{"prompt_tokens":100,"completion_tokens":60,"prompt_tokens_details":{"image_tokens":20,"cached_tokens":7},"completion_tokens_details":{"image_tokens":30}}}}`)
	client := newPassthroughTestFrameConn(nil, false)
	upstream := newPassthroughTestFrameConn([]passthroughTestFrame{
		{coderws.MessageText, first}, {coderws.MessageText, duplicate}, {coderws.MessageText, second}, {coderws.MessageText, duplicate},
	}, true)
	var turns []RelayTurnResult
	result, exit := Relay(context.Background(), client, upstream, []byte(`{"type":"response.create","model":"fake"}`), RelayOptions{
		UpstreamDrainTimeout: time.Millisecond * 10,
		OnTurnComplete:       func(turn RelayTurnResult) { turns = append(turns, turn) },
	})
	require.Nil(t, exit)
	require.Len(t, turns, 2)
	require.Equal(t, 40, turns[0].Usage.ImageInputTokens)
	require.Equal(t, 20, turns[1].Usage.ImageInputTokens)
	require.Equal(t, 60, result.Usage.ImageInputTokens)
	require.Equal(t, 85, result.Usage.ImageOutputTokens)
	require.Equal(t, 220, result.Usage.InputTokens)
	require.Equal(t, 12, result.Usage.CacheReadInputTokens)
	require.Equal(t, 3, result.Usage.CacheCreationInputTokens)
	require.Len(t, client.Writes(), 4, "duplicate wire frames are still forwarded, including late retransmissions")
}

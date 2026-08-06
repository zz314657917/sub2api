package apicompat

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testToolImageDataURL = "data:image/png;base64,AQID"
	testToolImageRemote  = "https://example.com/tool-output.png"
)

func TestResponsesToolOutputMedia_ExtractsSupportedShapes(t *testing.T) {
	tests := []struct {
		name       string
		call       string
		outputType string
		output     string
		imageURL   string
		toolText   string
	}{
		{
			name:       "image-only array",
			call:       `{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"}`,
			outputType: "function_call_output",
			output:     `[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]`,
			imageURL:   testToolImageDataURL,
		},
		{
			name:       "text and nested image URL",
			call:       `{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"}`,
			outputType: "function_call_output",
			output:     `[{"type":"input_text","text":"render complete"},{"type":"image_url","image_url":{"url":"https://example.com/tool-output.png"}}]`,
			imageURL:   testToolImageRemote,
			toolText:   "render complete",
		},
		{
			name:       "top-level image object",
			call:       `{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"}`,
			outputType: "function_call_output",
			output:     `{"type":"input_image","image_url":"data:image/png;base64,AQID"}`,
			imageURL:   testToolImageDataURL,
		},
		{
			name:       "content wrapper keeps large number",
			call:       `{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"}`,
			outputType: "function_call_output",
			output:     `{"status":"ok","content":[{"type":"input_image","image_url":"data:image/png;base64,AQID"}],"unknown":{"large":9007199254740993}}`,
			imageURL:   testToolImageDataURL,
			toolText:   `"large":9007199254740993`,
		},
		{
			name:       "JSON string output",
			call:       `{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"}`,
			outputType: "function_call_output",
			output:     `"[{\"type\":\"input_image\",\"image_url\":\"data:image/png;base64,AQID\"}]"`,
			imageURL:   testToolImageDataURL,
		},
		{
			name:       "bare data URL",
			call:       `{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"}`,
			outputType: "function_call_output",
			output:     `"data:image/png;base64,AQID"`,
			imageURL:   testToolImageDataURL,
		},
		{
			name:       "custom tool output",
			call:       `{"type":"custom_tool_call","call_id":"call_image","name":"view_image","input":"{}"}`,
			outputType: "custom_tool_call_output",
			output:     `[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]`,
			imageURL:   testToolImageDataURL,
		},
		{
			name:       "tool search output",
			call:       `{"type":"tool_search_call","call_id":"call_image","arguments":{"query":"image"}}`,
			outputType: "tool_search_output",
			output:     `[{"type":"image_url","image_url":{"url":"https://example.com/tool-output.png"}}]`,
			imageURL:   testToolImageRemote,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`[%s,{"type":%q,"call_id":"call_image","output":%s}]`, tt.call, tt.outputType, tt.output)
			messages := convertToolOutputMedia(t, input)

			require.Equal(t, []string{"assistant", "tool", "user"}, chatMessageRoles(messages))
			require.Equal(t, "call_image", messages[1].ToolCallID)

			toolText := chatToolContentString(t, messages[1])
			require.Contains(t, toolText, toolOutputMediaMarker)
			require.NotContains(t, toolText, tt.imageURL)
			if tt.toolText != "" {
				require.Contains(t, toolText, tt.toolText)
			}

			parts := chatContentParts(t, messages[2])
			require.Len(t, parts, 2)
			require.Equal(t, fmt.Sprintf(toolOutputMediaAttribution, "call_image"), parts[0].Text)
			require.Equal(t, tt.imageURL, parts[1].ImageURL.URL)
		})
	}
}

func TestResponsesToolOutputMedia_ParallelBatchUsesCallOrder(t *testing.T) {
	messages := convertToolOutputMedia(t, `[
		{"type":"function_call","call_id":"call_A","name":"view_image","arguments":"{}"},
		{"type":"function_call","call_id":"call_B","name":"view_image","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_B","output":[{"type":"input_image","image_url":{"url":"https://example.com/b.png"}}]},
		{"type":"function_call_output","call_id":"call_A","output":[{"type":"input_image","image_url":{"url":"https://example.com/a.png"}}]}
	]`)

	require.Equal(t, []string{"assistant", "tool", "tool", "user"}, chatMessageRoles(messages))
	require.Equal(t, []string{"call_A", "call_B"}, []string{messages[1].ToolCallID, messages[2].ToolCallID})

	parts := chatContentParts(t, messages[3])
	require.Len(t, parts, 4)
	require.Equal(t, fmt.Sprintf(toolOutputMediaAttribution, "call_A"), parts[0].Text)
	require.Equal(t, "https://example.com/a.png", parts[1].ImageURL.URL)
	require.Equal(t, fmt.Sprintf(toolOutputMediaAttribution, "call_B"), parts[2].Text)
	require.Equal(t, "https://example.com/b.png", parts[3].ImageURL.URL)
}

func TestResponsesToolOutputMedia_InterleavedMessagesFollowCompleteMediaBatch(t *testing.T) {
	messages := convertToolOutputMedia(t, `[
		{"type":"function_call","call_id":"call_A","name":"view_image","arguments":"{}"},
		{"type":"message","role":"developer","content":[{"type":"input_text","text":"approval saved"}]},
		{"type":"message","role":"user","content":[{"type":"input_text","text":"continue"}]},
		{"type":"function_call_output","call_id":"call_A","output":[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]}
	]`)

	require.Equal(t, []string{"assistant", "tool", "user", "system", "user"}, chatMessageRoles(messages))
	require.Equal(t, fmt.Sprintf(toolOutputMediaAttribution, "call_A"), chatContentParts(t, messages[2])[0].Text)
	require.JSONEq(t, `"approval saved"`, string(messages[3].Content))
	require.JSONEq(t, `"continue"`, string(messages[4].Content))
}

func TestResponsesToolOutputMedia_DropsOrphanAndUnansweredCallMedia(t *testing.T) {
	t.Run("orphan", func(t *testing.T) {
		messages := convertToolOutputMedia(t, `[
			{"type":"function_call_output","call_id":"call_ghost","output":[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]}
		]`)
		require.Empty(t, messages)
	})

	t.Run("unanswered parallel call", func(t *testing.T) {
		messages := convertToolOutputMedia(t, `[
			{"type":"function_call","call_id":"call_A","name":"view_image","arguments":"{}"},
			{"type":"function_call","call_id":"call_B","name":"view_image","arguments":"{}"},
			{"type":"function_call_output","call_id":"call_A","output":[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]}
		]`)
		require.Len(t, messages[0].ToolCalls, 1)
		require.Equal(t, "call_A", messages[0].ToolCalls[0].ID)
		require.NotContains(t, string(messages[2].Content), "call_B")
	})
}

func TestResponsesToolOutputMedia_PreservesMediaFreeOutputAndDuplicateLastWins(t *testing.T) {
	t.Run("media free output", func(t *testing.T) {
		output := `[ { "type": "input_text", "text": "ok" }, {"unknown":true} ]`
		messages := convertToolOutputMedia(t, fmt.Sprintf(`[
			{"type":"function_call","call_id":"call_text","name":"exec","arguments":"{}"},
			{"type":"function_call_output","call_id":"call_text","output":%s}
		]`, output))
		require.Equal(t, []string{"assistant", "tool"}, chatMessageRoles(messages))
		require.JSONEq(t, fmt.Sprintf(`%q`, output), string(messages[1].Content))
	})

	t.Run("later media replaces earlier media", func(t *testing.T) {
		messages := convertToolOutputMedia(t, `[
			{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"},
			{"type":"function_call_output","call_id":"call_image","output":[{"type":"input_image","image_url":{"url":"https://example.com/first.png"}}]},
			{"type":"function_call_output","call_id":"call_image","output":[{"type":"input_image","image_url":{"url":"https://example.com/last.png"}}]}
		]`)
		require.Equal(t, []string{"assistant", "tool", "user"}, chatMessageRoles(messages))
		require.NotContains(t, string(messages[1].Content), "first.png")
		require.NotContains(t, string(messages[2].Content), "first.png")
		require.Contains(t, string(messages[2].Content), "last.png")
	})

	t.Run("later text clears earlier media", func(t *testing.T) {
		messages := convertToolOutputMedia(t, `[
			{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"},
			{"type":"function_call_output","call_id":"call_image","output":[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]},
			{"type":"function_call_output","call_id":"call_image","output":"latest text"}
		]`)
		require.Equal(t, []string{"assistant", "tool"}, chatMessageRoles(messages))
		require.Equal(t, "latest text", chatToolContentString(t, messages[1]))
	})
}

func TestResponsesToolOutputMedia_EmptyCallIDDoesNotAttachMedia(t *testing.T) {
	messages := convertToolOutputMedia(t, `[
		{"type":"function_call","name":"view_image","arguments":"{}"},
		{"type":"function_call_output","output":[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]}
	]`)

	require.Equal(t, []string{"tool"}, chatMessageRoles(messages))
	require.NotContains(t, string(messages[0].Content), "data:image/")
	require.Contains(t, string(messages[0].Content), toolOutputMediaMarker)
}

func TestResponsesToChatCompletionsRequest_ToolContentNeverContainsExtractedMedia(t *testing.T) {
	req := &ResponsesRequest{
		Model: "vision-model",
		Input: json.RawMessage(`[
			{"type":"function_call","call_id":"call_function","name":"view_image","arguments":"{}"},
			{"type":"custom_tool_call","call_id":"call_custom","name":"custom_image","input":"{}"},
			{"type":"tool_search_call","call_id":"call_search","arguments":{"query":"image"}},
			{"type":"function_call_output","call_id":"call_function","output":[{"type":"input_image","image_url":"data:image/png;base64,AQID"}]},
			{"type":"custom_tool_call_output","call_id":"call_custom","output":{"content":[{"type":"image_url","image_url":{"url":"https://example.com/custom.png"}}]}},
			{"type":"tool_search_output","call_id":"call_search","output":"data:image/jpeg;base64,BAUG"}
		]`),
	}

	out, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	assertToolCallReplyAdjacency(t, out.Messages)

	var toolCount int
	for _, message := range out.Messages {
		if message.Role != "tool" {
			continue
		}
		toolCount++
		content := string(message.Content)
		require.NotContains(t, content, "data:image/")
		require.NotContains(t, content, "https://example.com/custom.png")
	}
	require.Equal(t, 3, toolCount)
	require.Equal(t, []string{"assistant", "tool", "tool", "tool", "user"}, chatMessageRoles(out.Messages))
}

func convertToolOutputMedia(t *testing.T, input string) []ChatMessage {
	t.Helper()
	messages, err := responsesInputToChatMessages("", json.RawMessage(input))
	require.NoError(t, err)
	assertToolCallReplyAdjacency(t, messages)
	return messages
}

func assertToolCallReplyAdjacency(t *testing.T, messages []ChatMessage) {
	t.Helper()
	for index, message := range messages {
		if len(message.ToolCalls) == 0 {
			continue
		}
		for offset, toolCall := range message.ToolCalls {
			require.Less(t, index+offset+1, len(messages))
			require.Equal(t, "tool", messages[index+offset+1].Role)
			require.Equal(t, toolCall.ID, messages[index+offset+1].ToolCallID)
		}
	}
}

func chatToolContentString(t *testing.T, message ChatMessage) string {
	t.Helper()
	require.Equal(t, "tool", message.Role)
	var content string
	require.NoError(t, json.Unmarshal(message.Content, &content))
	return content
}

func chatContentParts(t *testing.T, message ChatMessage) []ChatContentPart {
	t.Helper()
	var parts []ChatContentPart
	require.NoError(t, json.Unmarshal(message.Content, &parts))
	for _, part := range parts {
		if part.Type == "image_url" {
			require.NotNil(t, part.ImageURL)
			require.NotEmpty(t, strings.TrimSpace(part.ImageURL.URL))
		}
	}
	return parts
}

package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeOpenAIResponsesRejectedFieldRetryBody(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		responseBody string
		absent       []string
		present      []string
	}{
		{
			name:         "callable namespace",
			body:         `{"input":[{"type":"function_call","namespace":"remove","name":"tool"}]}`,
			responseBody: `{"error":{"code":"unknown_parameter","message":"Unknown parameter: 'input[0].namespace'.","param":"input[0].namespace"}}`,
			absent:       []string{"input.0.namespace"},
			present:      []string{"input.0.name"},
		},
		{
			name:         "top level max output tokens",
			body:         `{"max_output_tokens":256,"input":"keep"}`,
			responseBody: `{"error":{"code":"unsupported_parameter","message":"Unsupported parameter: max_output_tokens","param":"max_output_tokens"}}`,
			absent:       []string{"max_output_tokens"},
			present:      []string{"input"},
		},
		{
			name:         "top level truncation",
			body:         `{"truncation":"auto","input":"keep"}`,
			responseBody: `{"error":{"code":"unknown_parameter","message":"Unknown parameter: truncation","param":"truncation"}}`,
			absent:       []string{"truncation"},
			present:      []string{"input"},
		},
		{
			name:         "indexed prompt cache breakpoint",
			body:         `{"input":[{"type":"message","prompt_cache_breakpoint":{"type":"message_start"}},{"type":"message","prompt_cache_breakpoint":{"type":"message_end"}}]}`,
			responseBody: `{"error":{"code":"invalid_parameter","message":"input[1].prompt_cache_breakpoint is not supported on this model","param":"input[1].prompt_cache_breakpoint"}}`,
			absent:       []string{"input.1.prompt_cache_breakpoint"},
			present:      []string{"input.0.prompt_cache_breakpoint"},
		},
		{
			name:         "top level prompt cache breakpoint",
			body:         `{"prompt_cache_breakpoint":{"type":"message_start"},"input":"keep"}`,
			responseBody: `{"error":{"code":"invalid_parameter","message":"prompt_cache_breakpoint is not supported on this model","param":"prompt_cache_breakpoint"}}`,
			absent:       []string{"prompt_cache_breakpoint"},
			present:      []string{"input"},
		},
		{
			name:         "reasoning null content",
			body:         `{"input":[{"type":"reasoning","content":null,"id":"keep"}]}`,
			responseBody: `{"error":{"code":"invalid_type","message":"Invalid type for 'input[0].content': expected one of a string or a list of input items, but got null instead.","param":"input[0].content"}}`,
			absent:       []string{"input.0.content"},
			present:      []string{"input.0.id"},
		},
		{
			name:         "reasoning maximum length zero",
			body:         `{"input":[{"type":"reasoning","content":[{"type":"reasoning_text","text":"keep"}],"id":"keep"}]}`,
			responseBody: `{"error":{"code":"array_above_max_length","message":"Invalid 'input[0].content': Array too long. Expected maximum length 0, but got 1.","param":"input[0].content"}}`,
			absent:       []string{"input.0.content"},
			present:      []string{"input.0.id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryBody, _, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(http.StatusBadRequest, []byte(tt.body), []byte(tt.responseBody))
			require.NoError(t, err)
			require.True(t, changed)
			for _, path := range tt.absent {
				require.False(t, gjson.GetBytes(retryBody, path).Exists(), path)
			}
			for _, path := range tt.present {
				require.True(t, gjson.GetBytes(retryBody, path).Exists(), path)
			}
		})
	}
}

func TestNormalizeOpenAIResponsesRejectedFieldRetryBodyStatusBatchAndFallback(t *testing.T) {
	body := []byte(`{"input":[{"type":"tool_search_output","status":"completed","call_id":"first"},{"type":"message","status":"completed","content":"keep"},{"type":"tool_search_output","status":"completed","call_id":"second"}]}`)
	retryBody, _, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(http.StatusBadRequest, body, []byte(`{"error":{"code":"unknown_parameter","message":"Unknown parameter: 'input[2].status'.","param":"input[2].status"}}`))
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(retryBody, "input.0.status").Exists())
	require.Equal(t, "completed", gjson.GetBytes(retryBody, "input.1.status").String())
	require.False(t, gjson.GetBytes(retryBody, "input.2.status").Exists())
	require.Equal(t, "first", gjson.GetBytes(retryBody, "input.0.call_id").String())

	fallbackBody := []byte(`{"input":[{"status":"keep"},{"status":"remove"}]}`)
	fallbackRetryBody, _, fallbackChanged, err := normalizeOpenAIResponsesRejectedFieldRetryBody(http.StatusBadRequest, fallbackBody, []byte(`{"error":{"code":"unknown_parameter","message":"Unknown parameter: 'input[1].status'.","param":"input[1].status"}}`))
	require.NoError(t, err)
	require.True(t, fallbackChanged)
	require.Equal(t, "keep", gjson.GetBytes(fallbackRetryBody, "input.0.status").String())
	require.False(t, gjson.GetBytes(fallbackRetryBody, "input.1.status").Exists())
}

func TestNormalizeOpenAIResponsesRejectedFieldRetryBodyStrictNoRetryGuards(t *testing.T) {
	body := []byte(`{"truncation":"auto","input":[{"type":"message","status":"keep"}]}`)
	tests := []struct {
		name         string
		status       int
		responseBody string
	}{
		{name: "non 400", status: http.StatusInternalServerError, responseBody: `{"error":{"code":"unknown_parameter","message":"Unknown parameter: truncation","param":"truncation"}}`},
		{name: "malformed JSON", status: http.StatusBadRequest, responseBody: `{`},
		{name: "param message mismatch", status: http.StatusBadRequest, responseBody: `{"error":{"code":"unknown_parameter","message":"Unknown parameter: truncation","param":"input[0].status"}}`},
		{name: "not explicit", status: http.StatusBadRequest, responseBody: `{"error":{"code":"invalid_request_error","message":"truncation must be auto","param":"truncation"}}`},
		{name: "unsupported item type", status: http.StatusBadRequest, responseBody: `{"error":{"code":"unknown_parameter","message":"Unknown parameter: 'input[0].namespace'.","param":"input[0].namespace"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryBody, _, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(tt.status, body, []byte(tt.responseBody))
			require.NoError(t, err)
			require.False(t, changed)
			require.Nil(t, retryBody)
		})
	}

	retryBody, _, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(http.StatusBadRequest, []byte(`{"truncation":"auto"}`), []byte(`{"error":{"code":"unknown_parameter","message":"Unknown parameter: truncation"}}`))
	require.NoError(t, err)
	require.True(t, changed, "a missing param may use the explicit message field")
	require.False(t, gjson.GetBytes(retryBody, "truncation").Exists())
}

func TestOpenAIResponsesRejectedFieldRetryState(t *testing.T) {
	initial := []byte(`{"model":"gpt-5"}`)
	state := newOpenAIResponsesRejectedFieldRetryState(initial)
	require.False(t, state.Allow(initial))
	require.False(t, state.Allow(nil))
	for attempt := 0; attempt < maxOpenAIResponsesRejectedFieldRetries; attempt++ {
		next := []byte(fmt.Sprintf(`{"model":"gpt-5","variant":%d}`, attempt))
		require.True(t, state.Allow(next))
		require.False(t, state.Allow(next))
	}
	require.False(t, state.Allow([]byte(`{"model":"gpt-5","variant":"overflow"}`)))
	rewrittenByInvalidEncryptedContent := []byte(`{"model":"gpt-5","input":[]}`)
	state.remember(rewrittenByInvalidEncryptedContent)
	require.False(t, state.Allow(rewrittenByInvalidEncryptedContent))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	accountA := openAIResponsesRejectedFieldRetryStateForRequest(c, initial)
	require.True(t, accountA.Allow([]byte(`{"retry":"a"}`)))
	require.False(t, accountA.Allow([]byte(`{"retry":"a"}`)))
	accountB := openAIResponsesRejectedFieldRetryStateForRequest(c, initial)
	require.True(t, accountB.Allow([]byte(`{"retry":"a"}`)), "a failover account may apply the same rewrite")
	for attempt := 0; attempt < maxOpenAIResponsesRejectedFieldRetries-2; attempt++ {
		state := openAIResponsesRejectedFieldRetryStateForRequest(c, []byte(fmt.Sprintf(`{"account":%d}`, attempt)))
		require.True(t, state.Allow([]byte(fmt.Sprintf(`{"retry":%d}`, attempt))))
	}
	overflow := openAIResponsesRejectedFieldRetryStateForRequest(c, initial)
	require.False(t, overflow.Allow([]byte(`{"retry":"overflow"}`)))
}

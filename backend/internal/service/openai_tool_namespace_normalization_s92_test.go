package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestS92NormalizeOpenAIResponsesCustomToolNamespaces(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		changed   bool
		assertion func(t *testing.T, output []byte)
	}{
		{
			name:    "root custom tool item",
			input:   `{"type":"custom_tool_call","name":"exec","namespace":"exec","call_id":"call_1","input":"pwd"}`,
			changed: true,
			assertion: func(t *testing.T, output []byte) {
				require.False(t, gjson.GetBytes(output, "namespace").Exists())
			},
		},
		{
			name:    "stream item",
			input:   `{"type":"response.output_item.done","item":{"type":"custom_tool_call","name":"exec","namespace":"exec","call_id":"call_1","input":"pwd"}}`,
			changed: true,
			assertion: func(t *testing.T, output []byte) {
				require.False(t, gjson.GetBytes(output, "item.namespace").Exists())
			},
		},
		{
			name: "terminal response output preserves distinct and non-custom namespaces",
			input: `{"type":"response.completed","response":{"output":[` +
				`{"type":"custom_tool_call","name":"exec","namespace":"exec"},` +
				`{"type":"custom_tool_call","name":"exec","namespace":"mcp__shell"},` +
				`{"type":"function_call","name":"exec","namespace":"exec"}]}}`,
			changed: true,
			assertion: func(t *testing.T, output []byte) {
				require.False(t, gjson.GetBytes(output, "response.output.0.namespace").Exists())
				require.Equal(t, "mcp__shell", gjson.GetBytes(output, "response.output.1.namespace").String())
				require.Equal(t, "exec", gjson.GetBytes(output, "response.output.2.namespace").String())
			},
		},
		{
			name:    "plain response output",
			input:   `{"id":"resp_1","output":[{"type":"custom_tool_call","name":"exec","namespace":"exec"}]}`,
			changed: true,
			assertion: func(t *testing.T, output []byte) {
				require.False(t, gjson.GetBytes(output, "output.0.namespace").Exists())
			},
		},
		{
			name:    "different case remains distinct",
			input:   `{"type":"custom_tool_call","name":"exec","namespace":"Exec"}`,
			changed: false,
		},
		{
			name:    "invalid json remains unchanged",
			input:   `{"type":"custom_tool_call","name":"exec","namespace":"exec"`,
			changed: false,
		},
		{
			name:    "unrelated payload remains unchanged",
			input:   `{"type":"response.output_text.delta","delta":"namespace custom_tool_call"}`,
			changed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []byte(tt.input)
			output, changed := normalizeOpenAIResponsesCustomToolNamespaces(input)
			require.Equal(t, tt.changed, changed)
			if !tt.changed {
				require.Equal(t, input, output)
			}
			if tt.assertion != nil {
				tt.assertion(t, output)
			}
		})
	}
}

func TestS92CorrectToolCallsInResponseBodyNormalizesNamespace(t *testing.T) {
	service := &OpenAIGatewayService{toolCorrector: NewCodexToolCorrector()}
	input := []byte(`{"id":"resp_1","output":[{"type":"custom_tool_call","name":"exec","namespace":"exec"}]}`)

	output := service.correctToolCallsInResponseBody(input)

	require.False(t, gjson.GetBytes(output, "output.0.namespace").Exists())
}

func TestS92NonStreamingPassthroughNormalizesNamespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	service := &OpenAIGatewayService{cfg: &config.Config{}}
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"id":"resp_1","model":"gpt-5.6-sol","output":[{"type":"custom_tool_call","name":"exec","namespace":"exec"}],"usage":{"input_tokens":1,"output_tokens":1}}`,
		)),
	}

	_, err := service.handleNonStreamingResponsePassthrough(context.Background(), response, c, "gpt-5.6-sol", "gpt-5.6-sol")

	require.NoError(t, err)
	require.False(t, gjson.GetBytes(recorder.Body.Bytes(), "output.0.namespace").Exists())
}

func TestS92StreamingPassthroughNormalizesNamespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	service := &OpenAIGatewayService{cfg: &config.Config{}}
	stream := strings.Join([]string{
		`data: {"type":"response.output_item.done","item":{"type":"custom_tool_call","name":"exec","namespace":"exec","call_id":"call_1","input":"pwd"}}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_1","output":[{"type":"custom_tool_call","name":"exec","namespace":"exec","call_id":"call_1","input":"pwd"}],"usage":{"input_tokens":1,"output_tokens":1}}}`,
		"",
	}, "\n")
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(stream)),
	}

	_, err := service.handleStreamingResponsePassthrough(
		context.Background(), response, c, &Account{ID: 92}, time.Now(), "gpt-5.6-sol", "gpt-5.6-sol",
	)

	require.NoError(t, err)
	require.NotContains(t, recorder.Body.String(), `"namespace":"exec"`)
	require.Contains(t, recorder.Body.String(), `"type":"custom_tool_call"`)
}

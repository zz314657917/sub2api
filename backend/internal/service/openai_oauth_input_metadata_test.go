package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const oauthInputMetadataField = "internal_chat_message_metadata_passthrough"

func TestOAuthInputInternalMetadata(t *testing.T) {
	const body = `{"model":"gpt-5.5","input":[{"role":"user","content":[{"type":"input_text","text":"hello","internal_chat_message_metadata_passthrough":{"keep":true}}],"internal_chat_message_metadata_passthrough":{"content_item_kinds":["text"]}},{"type":"function_call","call_id":"fc_123","arguments":"{\"internal_chat_message_metadata_passthrough\":true}","internal_chat_message_metadata_passthrough":null},null,"text"],"internal_chat_message_metadata_passthrough":{"keep":true}}`
	tests := []struct {
		name      string
		normalize func([]byte) ([]byte, bool, error)
		strip     bool
	}{
		{"OAuth raw HTTP", normalizeOpenAIOAuthResponsesCompatibilityBody, true},
		{"OAuth websocket", func(b []byte) ([]byte, bool, error) {
			return normalizeOpenAIResponsesWebSocketCompatibilityBody(b, &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth})
		}, true},
		{"API-key websocket", func(b []byte) ([]byte, bool, error) {
			return normalizeOpenAIResponsesWebSocketCompatibilityBody(b, &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey})
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, changed, err := tt.normalize([]byte(body))
			require.NoError(t, err)
			require.Equal(t, tt.strip, changed)
			require.Equal(t, !tt.strip, gjson.GetBytes(out, "input.0."+oauthInputMetadataField).Exists())
			require.True(t, gjson.GetBytes(out, "input.0.content.0."+oauthInputMetadataField+".keep").Bool())
			require.Equal(t, `{"internal_chat_message_metadata_passthrough":true}`, gjson.GetBytes(out, "input.1.arguments").String())
			require.True(t, gjson.GetBytes(out, oauthInputMetadataField+".keep").Bool())
		})
	}
}

func TestOAuthInputInternalMetadataHTTPForwardAndAPIKeyBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5.5","stream":false,"input":[{"type":"message","role":"user","content":"hello","internal_chat_message_metadata_passthrough":{"drop":true},"nested":{"internal_chat_message_metadata_passthrough":{"keep":true}}}]}`)
	tests := []struct {
		name        string
		account     *Account
		shouldStrip bool
	}{
		{
			name:        "OAuth forwarding strips only input item field",
			account:     &Account{ID: 99101, Name: "oauth", Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1, Credentials: map[string]any{"access_token": "oauth-token", "chatgpt_account_id": "chatgpt-account"}, Extra: map[string]any{"openai_passthrough": true}, Status: StatusActive, Schedulable: true, RateMultiplier: f64p(1)},
			shouldStrip: true,
		},
		{
			name:        "API key forwarding preserves field",
			account:     &Account{ID: 99102, Name: "api-key", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://api.openai.com"}, Extra: map[string]any{"openai_passthrough": true}, Status: StatusActive, Schedulable: true, RateMultiplier: f64p(1)},
			shouldStrip: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
			contentType := "application/json"
			responseBody := []byte(`{"output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)
			if tt.shouldStrip {
				contentType = "text/event-stream"
				responseBody = []byte("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_meta\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\ndata: [DONE]\n\n")
			}
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(bytes.NewReader(responseBody))}}
			svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{ForceCodexCLI: false}}, httpUpstream: upstream}

			_, err := svc.Forward(context.Background(), c, tt.account, body)
			require.NoError(t, err)
			require.NotNil(t, upstream.lastReq)
			require.Equal(t, !tt.shouldStrip, gjson.GetBytes(upstream.lastBody, "input.0."+oauthInputMetadataField).Exists())
			require.True(t, gjson.GetBytes(upstream.lastBody, "input.0.nested."+oauthInputMetadataField+".keep").Bool())
		})
	}
}

func TestOAuthInputInternalMetadataMapAndInputOnlyCompatibility(t *testing.T) {
	request := map[string]any{"input": []any{map[string]any{oauthInputMetadataField: true, "content": map[string]any{oauthInputMetadataField: "keep"}}, "text", nil}}
	require.True(t, normalizeOpenAIOAuthResponsesCompatibilityFields(request))
	input := request["input"].([]any)
	item := input[0].(map[string]any)
	require.NotContains(t, item, oauthInputMetadataField)
	require.Equal(t, "keep", item["content"].(map[string]any)[oauthInputMetadataField])

	oauthRequest := map[string]any{"model": "gpt-5.5", "input": []any{map[string]any{"role": "user", "content": "hello", oauthInputMetadataField: true}}}
	transform := applyCodexOAuthTransform(oauthRequest, false, false)
	require.NoError(t, transform.Error)
	transformedItem := oauthRequest["input"].([]any)[0].(map[string]any)
	require.NotContains(t, transformedItem, oauthInputMetadataField)

	for _, body := range []string{
		`{"input":"hello"}`,
		`{"input":{"internal_chat_message_metadata_passthrough":true}}`,
		`{"input":[]}`,
		`{"input":null}`,
		`{"model":"gpt-5.5"}`,
		`{"input":[null,"hello",{"content":{"internal_chat_message_metadata_passthrough":true}}]}`,
	} {
		out, changed, err := normalizeOpenAIOAuthResponsesCompatibilityBody([]byte(body))
		require.NoError(t, err)
		require.False(t, changed)
		require.JSONEq(t, body, string(out))
	}

	legacy := []byte(`{"prompt":[{"internal_chat_message_metadata_passthrough":{}}],"previous_response_id":"resp_123"}`)
	out, changed, err := normalizeOpenAIOAuthResponsesCompatibilityBody(legacy)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(out, "input.0."+oauthInputMetadataField).Exists())
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(out, &decoded))
	require.Equal(t, "resp_123", decoded["previous_response_id"])
}

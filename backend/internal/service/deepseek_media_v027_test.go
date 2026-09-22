package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestV027DeepSeekForwardsLiftedMediaToLocalhost(t *testing.T) {
	var forwarded []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/responses", r.URL.Path)
		var err error
		forwarded, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	gin.SetMode(gin.TestMode)
	account := cnAccount(1, PlatformDeepseek, AccountModePayG, upstream.URL)
	account.Credentials["api_protocol"] = APIProtocolResponses
	body := []byte(`{
		"model":"deepseek-reasoner","store":true,"previous_response_id":"resp_1","trace_id":9007199254740993,
		"input":[
			{"type":"function_call","call_id":"call_A","name":"view_image","arguments":"{}"},
			{"type":"function_call","call_id":"call_B","name":"view_image","arguments":"{}"},
			{"type":"function_call_output","call_id":"call_A","output":[{"type":"input_image","image_url":"data:image/png;base64,QQ=="},{"type":"input_text","text":"first"}]},
			{"type":"message","role":"developer","content":"notice A"},
			{"type":"function_call_output","call_id":"call_B","output":[{"type":"input_image","image_url":"data:image/png;base64,Qg=="},{"type":"input_text","text":"second"}]},
			{"type":"message","role":"system","content":"notice B"}
		]
	}`)

	svc := newCNNativeGatewayTestService(nil)
	svc.cfg.Security.URLAllowlist.Enabled = false
	svc.cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	req, err := svc.buildUpstreamRequest(context.Background(), c, account, body, "", false, "", false)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	require.Equal(t, "false", gjson.GetBytes(forwarded, "store").Raw)
	require.False(t, gjson.GetBytes(forwarded, "previous_response_id").Exists())
	require.Equal(t, "9007199254740993", gjson.GetBytes(forwarded, "trace_id").Raw)
	input := gjson.GetBytes(forwarded, "input")
	require.Equal(t, "function_call_output", input.Get("2.type").String())
	require.Equal(t, "function_call_output", input.Get("3.type").String())
	require.NotContains(t, input.Get("2.output").String(), "data:image")
	require.NotContains(t, input.Get("3.output").String(), "data:image")
	require.Equal(t, "user", input.Get("4.role").String())
	require.Equal(t, "data:image/png;base64,QQ==", input.Get("4.content.1.image_url").String())
	require.Equal(t, "data:image/png;base64,Qg==", input.Get("4.content.3.image_url").String())
	require.Equal(t, "developer", input.Get("5.role").String())
	require.Equal(t, "system", input.Get("6.role").String())
}

func TestV027DeepSeekNormalizationLeavesNonTargetsAndPlainOutputUntouched(t *testing.T) {
	plain := []byte(`{"input":[{"type":"function_call_output","output":"plain"}],"trace_id":9007199254740993}`)
	deepseek := cnAccount(1, PlatformDeepseek, AccountModePayG, "https://example.invalid")
	deepseek.Credentials["api_protocol"] = APIProtocolResponses
	normalized := normalizeDeepSeekResponsesRequestBody(deepseek, plain)
	var got map[string]any
	require.NoError(t, json.Unmarshal(normalized, &got))
	require.Equal(t, "plain", gjson.GetBytes(normalized, "input.0.output").String())
	require.Equal(t, "9007199254740993", gjson.GetBytes(normalized, "trace_id").Raw)

	nonTarget := &Account{Platform: PlatformOpenAI}
	require.Equal(t, plain, normalizeDeepSeekResponsesRequestBody(nonTarget, plain))
}

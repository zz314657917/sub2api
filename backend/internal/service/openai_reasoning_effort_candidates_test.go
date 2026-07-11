package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestExtractOpenAIReasoningEffortFromBodyModelCandidates(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		candidates []string
		want       string
	}{
		{name: "suffix fallback", body: []byte(`{"input":"hello"}`), candidates: []string{"gpt-5.4", "gpt-5.4-xhigh"}, want: "xhigh"},
		{name: "gpt56 suffix max", body: []byte(`{"input":"hello"}`), candidates: []string{"gpt-5.6-sol", "gpt-5.6-sol-max"}, want: "max"},
		{name: "explicit mapped gpt56 max", body: []byte(`{"reasoning":{"effort":"max"}}`), candidates: []string{"gpt-5.6-sol", "sol"}, want: "max"},
		{name: "explicit non gpt56 max", body: []byte(`{"reasoning":{"effort":"max"}}`), candidates: []string{"gpt-5.4", "sol"}, want: "xhigh"},
		{name: "no effort", body: []byte(`{"input":"hello"}`), candidates: []string{"gpt-5.4"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOpenAIReasoningEffortFromBody(tt.body, tt.candidates...)
			if tt.want == "" {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, tt.want, *got)
		})
	}
}

func TestExtractOpenAIReasoningEffortModelCandidates(t *testing.T) {
	got := extractOpenAIReasoningEffort(map[string]any{"model": "gpt-5.3-codex-high"}, "gpt-5.3-codex", "gpt-5.3-codex-high")
	require.NotNil(t, got)
	require.Equal(t, "high", *got)
}

func TestOpenAIGatewayServiceForwardOAuthReasoningEffortCandidateSuffix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"usage":{"input_tokens":1,"output_tokens":2}}`)),
	}}
	cfg := &config.Config{}
	cfg.Security.URLAllowlist.Enabled = false
	svc := &OpenAIGatewayService{cfg: cfg, httpUpstream: upstream}
	account := &Account{
		ID: 11, Name: "openai-oauth-suffix", Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1,
		Credentials: map[string]any{"access_token": "oauth-token", "chatgpt_account_id": "chatgpt-acc"},
		Status:      StatusActive, Schedulable: true,
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

	result, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-5.3-codex-xhigh","instructions":"suffix-test","input":"hello","stream":false}`))

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "gpt-5.3-codex", gjson.GetBytes(upstream.lastBody, "model").String())
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "xhigh", *result.ReasoningEffort)
}

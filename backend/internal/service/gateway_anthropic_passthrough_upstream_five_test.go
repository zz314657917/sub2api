package service

import (
	"context"
	"fmt"
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

func TestUpstreamFiveAnthropicPassthroughMappedThinkingOutbound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, original, mapped string
		strict                 bool
	}{
		{"strict", "claude-sonnet-4", "claude-sonnet-4", true},
		{"alias_to_strict", "client-alias", "claude-sonnet-4", true},
		{"passback_to_strict", "deepseek-v4-pro", "claude-sonnet-4", true},
		{"strict_to_deepseek", "claude-sonnet-4", "deepseek-v4-pro", false},
		{"strict_to_kimi", "claude-sonnet-4", "kimi-k2", false},
		{"strict_to_glm", "claude-sonnet-4", "glm-4.7", false},
		{"strict_to_unknown", "claude-sonnet-4", "custom-model", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			body := []byte(fmt.Sprintf(`{"model":%q,"max_tokens":64,"thinking":{"type":"enabled","budget_tokens":1024},"messages":[{"role":"user","content":[{"type":"text","text":"hi"}]},{"role":"assistant","content":[{"type":"thinking","thinking":"missing"},{"type":"thinking","thinking":"empty","signature":""},{"type":"thinking","thinking":"valid","signature":"valid-signature"},{"type":"text","text":"kept"}]},{"role":"user","content":[{"type":"text","text":"continue"}]}]}`, tc.original))
			account := newAnthropicAPIKeyAccountForTest()
			account.Credentials["model_mapping"] = map[string]any{tc.original: tc.mapped}
			upstream := &anthropicHTTPUpstreamRecorder{resp: &http.Response{
				StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{"id":"msg_1","type":"message","usage":{"input_tokens":5,"output_tokens":3}}`)),
			}}
			svc := &GatewayService{cfg: &config.Config{}, httpUpstream: upstream, rateLimitService: &RateLimitService{}}
			result, err := svc.Forward(context.Background(), c, account, &ParsedRequest{Body: body, Model: tc.original})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, upstream.lastReq)
			require.Equal(t, tc.mapped, gjson.GetBytes(upstream.lastBody, "model").String())
			blocks := gjson.GetBytes(upstream.lastBody, "messages.1.content").Array()
			want := 4
			if tc.strict {
				want = 2
			}
			require.Len(t, blocks, want)
			require.Equal(t, "valid-signature", blocks[len(blocks)-2].Get("signature").String())
			require.Equal(t, "kept", blocks[len(blocks)-1].Get("text").String())
			require.Equal(t, "continue", gjson.GetBytes(upstream.lastBody, "messages.2.content.0.text").String())
			if !tc.strict {
				require.Equal(t, "missing", blocks[0].Get("thinking").String())
				require.Equal(t, "", blocks[1].Get("signature").String())
			}
		})
	}
}

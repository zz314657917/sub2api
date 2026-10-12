package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// All requests, including hard-coded OAuth endpoint URLs, are physically sent
// only to this test's loopback server. The production request body is untouched.
type v0215LocalHTTPUpstream struct {
	target *url.URL
	client *http.Client
}

func (u *v0215LocalHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	local := req.Clone(req.Context())
	target := *req.URL
	target.Scheme, target.Host = u.target.Scheme, u.target.Host
	local.URL, local.Host = &target, u.target.Host
	return u.client.Do(local)
}

func (u *v0215LocalHTTPUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

func v0215LocalHTTPService(server *httptest.Server) *OpenAIGatewayService {
	target, _ := url.Parse(server.URL)
	cfg := &config.Config{}
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	return &OpenAIGatewayService{cfg: cfg, httpUpstream: &v0215LocalHTTPUpstream{target: target, client: server.Client()}}
}

func v0215HTTPAccount(typ string) *Account {
	return &Account{
		ID: 453, Name: "fake-local-v0215", Platform: PlatformOpenAI, Type: typ, Concurrency: 1,
		Credentials: map[string]any{"api_key": "fake-key", "access_token": "fake-token", "base_url": "https://unused.invalid"},
		Extra:       map[string]any{"use_responses_api": true},
	}
}

func TestV0215WSHTTPThinkingSignatureRecovery(t *testing.T) {
	const exact = "The encrypted content gAAA could not be verified. It could not be decrypted or parsed."
	for _, tc := range []struct {
		name, code, message  string
		status               int
		encrypted, failAgain bool
		attempts             int
	}{
		{"exact-recover", "thinking_signature_invalid", exact, 400, true, false, 2},
		{"bounded-once", "thinking_signature_invalid", exact, 400, true, true, 2},
		{"no-encrypted", "thinking_signature_invalid", exact, 400, false, false, 1},
		{"missing-verify", "thinking_signature_invalid", "The encrypted content could not be decrypted or parsed", 400, true, false, 1},
		{"missing-decrypt", "thinking_signature_invalid", "The encrypted content could not be verified", 400, true, false, 1},
		{"other-code", "thinking_invalid", exact, 400, true, false, 1},
		{"wrong-status", "thinking_signature_invalid", exact, 422, true, false, 1},
		{"legacy-code", "invalid_encrypted_content", "legacy invalid content", 400, true, false, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var bodies [][]byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				bodies = append(bodies, body)
				w.Header().Set("Content-Type", "application/json")
				if len(bodies) == 1 || tc.failAgain {
					w.WriteHeader(tc.status)
					_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": tc.code, "type": "invalid_request_error", "message": tc.message}})
					return
				}
				_, _ = io.WriteString(w, `{"id":"resp_local","usage":{"input_tokens":10,"output_tokens":2}}`)
			}))
			defer server.Close()
			svc := v0215LocalHTTPService(server)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
			SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)
			reasoning := `{"type":"reasoning","summary":[{"type":"summary_text","text":"keep"}]}`
			if tc.encrypted {
				reasoning = `{"type":"reasoning","encrypted_content":"fake-encrypted","summary":[{"type":"summary_text","text":"keep"}]}`
			}
			body := []byte(fmt.Sprintf(`{"model":"gpt-5.4","stream":false,"previous_response_id":"resp_stale","input":[%s,{"role":"user","content":"hello"}]}`, reasoning))
			result, err := svc.Forward(context.Background(), c, v0215HTTPAccount(AccountTypeAPIKey), body)
			require.Len(t, bodies, tc.attempts)
			if tc.attempts == 2 && !tc.failAgain {
				require.NoError(t, err)
				require.NotNil(t, result)
			} else {
				require.Error(t, err)
			}
			if tc.attempts == 2 {
				require.True(t, gjson.GetBytes(bodies[0], "input.0.encrypted_content").Exists())
				require.False(t, gjson.GetBytes(bodies[1], "input.0.encrypted_content").Exists())
				require.Equal(t, "keep", gjson.GetBytes(bodies[1], "input.0.summary.0.text").String())
				require.False(t, gjson.GetBytes(bodies[1], "previous_response_id").Exists())
			}
		})
	}
}

func TestV0215OAuthHistoryFakeLocalHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, typ                  string
		lite, compact, passthrough bool
	}{
		{"oauth-normal", AccountTypeOAuth, false, false, false},
		{"oauth-lite", AccountTypeOAuth, true, false, false},
		{"oauth-passthrough", AccountTypeOAuth, false, false, true},
		{"oauth-passthrough-lite", AccountTypeOAuth, true, false, true},
		{"oauth-compact", AccountTypeOAuth, false, true, false},
		{"oauth-passthrough-compact", AccountTypeOAuth, true, true, true},
		{"apikey-normal", AccountTypeAPIKey, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var captured []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured, _ = io.ReadAll(r.Body)
				if strings.Contains(r.URL.Path, "compact") {
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"id":"resp_compact","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)
				} else {
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = io.WriteString(w, "data: "+`{"type":"response.completed","response":{"id":"resp_local","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1}}}`+"\n\n")
				}
			}))
			defer server.Close()
			svc := v0215LocalHTTPService(server)
			account := v0215HTTPAccount(tc.typ)
			if tc.passthrough {
				account.Extra["openai_passthrough"] = true
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			path := "/openai/v1/responses"
			if tc.compact {
				path += "/compact"
			}
			c.Request = httptest.NewRequest(http.MethodPost, path, nil)
			if tc.lite {
				c.Request.Header.Set(responsesLiteHeader, "true")
			}
			SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)
			body := `{"model":"gpt-5.4","stream":true,"instructions":"local","input":[{"type":"web_search_call","id":"ws_local","status":"completed"}],"tools":[]}`
			if tc.compact {
				body = strings.Replace(body, `"stream":true`, `"stream":false`, 1)
			}
			_, err := svc.Forward(context.Background(), c, account, []byte(body))
			require.NoError(t, err)
			require.NotEmpty(t, captured)
			expect := tc.typ == AccountTypeOAuth && !tc.compact
			if tc.lite {
				require.Equal(t, expect, gjson.GetBytes(captured, "input.1.tools.0.type").String() == "web_search")
				require.Empty(t, gjson.GetBytes(captured, "tools").Array())
			} else {
				require.Equal(t, expect, gjson.GetBytes(captured, "tools.0.type").String() == "web_search")
			}
			if expect {
				require.Equal(t, "none", gjson.GetBytes(captured, "tool_choice").String())
			}
		})
	}
}

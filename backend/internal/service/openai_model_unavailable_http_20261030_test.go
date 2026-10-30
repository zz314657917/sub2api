package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type modelUnavailableLoopback struct {
	HTTPUpstream
	client *http.Client
	url    string
	paths  []string
	models []string
}

func (u *modelUnavailableLoopback) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	u.paths = append(u.paths, req.URL.Path)
	u.models = append(u.models, gjson.GetBytes(body, "model").String())
	local, err := http.NewRequestWithContext(req.Context(), req.Method, u.url+req.URL.Path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	local.Header = req.Header.Clone()
	return u.client.Do(local)
}

func TestModelUnavailableRealHTTPOwners(t *testing.T) {
	for _, owner := range []string{"responses", "automatic-passthrough", "explicit-passthrough", "chat-to-responses", "chat-native", "responses-to-chat"} {
		for _, code := range []int{400, 404} {
			t.Run(owner+"/"+http.StatusText(code), func(t *testing.T) {
				for _, typ := range []string{AccountTypeAPIKey, AccountTypeOAuth} {
					t.Run(typ, func(t *testing.T) {
						payload := `{"error":{"code":"model_not_found","message":"The model does not exist"}}`
						if code == 404 {
							payload = modelUnavailable404
						} else if typ == AccountTypeOAuth {
							payload = modelUnavailablePlan
						}
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("Content-Type", "application/json")
							w.Header().Set("X-Request-Id", "rejected-model")
							w.WriteHeader(code)
							_, _ = io.WriteString(w, payload)
						}))
						defer server.Close()
						up := &modelUnavailableLoopback{client: server.Client(), url: server.URL}
						cfg := &config.Config{}
						cfg.Security.URLAllowlist.AllowInsecureHTTP = true
						a := modelUnavailableAccount(1)
						a.Type = typ
						a.Credentials["base_url"] = server.URL
						a.Credentials["model_mapping"] = map[string]any{"alias": "gpt-6-luna", "gpt-6-luna": "another"}
						if owner == "automatic-passthrough" || owner == "explicit-passthrough" {
							a.Extra["openai_passthrough"] = true
						}
						if owner == "chat-to-responses" {
							delete(a.Extra, "openai_responses_supported") // unknown => would previously probe Chat
						}
						if owner == "responses-to-chat" || owner == "chat-native" {
							a.Extra["openai_responses_supported"] = false
							a.Type = AccountTypeAPIKey
							payload = `{"error":{"code":"model_not_found"}}`
						}
						repo := &modelUnavailableRepo{schedulerTestOpenAIAccountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{a}}}
						s := &OpenAIGatewayService{cfg: cfg, httpUpstream: up, accountRepo: repo,
							rateLimitService: &RateLimitService{accountRepo: repo}}
						c, rec := modelUnavailableContext()
						body := []byte(`{"model":"alias","input":"hi","stream":false}`)
						var result *OpenAIForwardResult
						var err error
						switch owner {
						case "chat-to-responses", "chat-native":
							result, err = s.ForwardAsChatCompletions(context.Background(), c, &a, []byte(`{"model":"alias","messages":[{"role":"user","content":"hi"}]}`), "", "")
						case "explicit-passthrough":
							result, err = s.forwardOpenAIPassthrough(context.Background(), c, &a, body, "alias", nil, false, time.Now())
						default:
							result, err = s.Forward(context.Background(), c, &a, body)
						}
						var f *UpstreamFailoverError
						require.ErrorAs(t, err, &f)
						require.Nil(t, result)
						require.False(t, f.RetryableOnSameAccount)
						require.False(t, c.Writer.Written())
						require.Empty(t, rec.Body.String())
						require.Len(t, up.paths, 1, "must not probe another endpoint on rejected account")
						require.NotContains(t, repo.keys, "another", "must not double-map the actual upstream model")
						require.Contains(t, repo.keys, up.models[0], "effective sent model has cooldown")
					})
				}
			})
		}
	}
}

func TestModelUnavailableOrdinaryRoute404And405StillProbeChat(t *testing.T) {
	for _, status := range []int{404, 405} {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.Header().Set("Content-Type", "application/json")
			if strings.HasSuffix(r.URL.Path, "/responses") {
				w.WriteHeader(status)
				_, _ = io.WriteString(w, `{"error":{"message":"route not found"}}`)
				return
			}
			_, _ = io.WriteString(w, `{"id":"chat_ok","object":"chat.completion","model":"gpt-6-luna","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
		}))
		a := modelUnavailableAccount(1)
		delete(a.Extra, "openai_responses_supported")
		a.Credentials["base_url"] = server.URL
		cfg := &config.Config{}
		cfg.Security.URLAllowlist.AllowInsecureHTTP = true
		up := &modelUnavailableLoopback{client: server.Client(), url: server.URL}
		s := &OpenAIGatewayService{cfg: cfg, httpUpstream: up}
		c, _ := modelUnavailableContext()
		result, err := s.ForwardAsChatCompletions(context.Background(), c, &a, []byte(`{"model":"gpt-6-luna","messages":[{"role":"user","content":"hi"}]}`), "", "")
		server.Close()
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 2, calls)
		require.Equal(t, []string{"/v1/responses", "/v1/chat/completions"}, up.paths)
	}
}

func TestModelUnavailableLoopbackSwitchAndExhaustion(t *testing.T) {
	for _, exhausted := range []bool{false, true} {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.Header().Set("Content-Type", "application/json")
			if r.Header.Get("Authorization") == "Bearer a" || exhausted {
				w.WriteHeader(404)
				_, _ = io.WriteString(w, modelUnavailable404)
				return
			}
			_, _ = io.WriteString(w, `{"id":"resp_ok","object":"response","status":"completed","model":"gpt-6-luna","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}))
		a, b := modelUnavailableAccount(1), modelUnavailableAccount(2)
		a.Credentials["api_key"], b.Credentials["api_key"] = "a", "b"
		a.Credentials["base_url"], b.Credentials["base_url"] = server.URL, server.URL
		repo := &modelUnavailableRepo{schedulerTestOpenAIAccountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{a, b}}}
		cfg := &config.Config{}
		cfg.Security.URLAllowlist.AllowInsecureHTTP = true
		s := &OpenAIGatewayService{cfg: cfg, accountRepo: repo, httpUpstream: &modelUnavailableLoopback{client: server.Client(), url: server.URL},
			rateLimitService: &RateLimitService{accountRepo: repo}}
		c, rec := modelUnavailableContext()
		excluded := map[int64]struct{}{}
		for {
			selected, err := s.selectAccountForModelWithExclusions(context.Background(), nil, PlatformOpenAI, "", "gpt-6-luna", excluded, false, 0, "", AccountCapabilityChat)
			if err != nil {
				require.True(t, exhausted)
				break
			}
			result, err := s.Forward(context.Background(), c, selected, []byte(`{"model":"gpt-6-luna","input":"hi","stream":false}`))
			if err == nil {
				require.False(t, exhausted)
				require.NotNil(t, result)
				break
			}
			var f *UpstreamFailoverError
			require.True(t, errors.As(err, &f))
			require.Nil(t, result)
			require.False(t, c.Writer.Written())
			require.Empty(t, rec.Body.String())
			require.False(t, f.RetryableOnSameAccount)
			excluded[selected.ID] = struct{}{}
			require.LessOrEqual(t, len(excluded), 2)
		}
		server.Close()
		require.Equal(t, 2, calls)
	}
}

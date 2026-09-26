package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
)

func TestPelicanPromptIsUsedByClaudeStylePayload(t *testing.T) {
	payload, err := createTestPayloadWithPrompt("claude-sonnet-4-5", "生成完整 HTML 鹈鹕作品")
	if err != nil {
		t.Fatalf("create payload: %v", err)
	}
	messages, ok := payload["messages"].([]map[string]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("messages=%#v", payload["messages"])
	}
	content, ok := messages[0]["content"].([]map[string]any)
	if !ok || len(content) != 1 || content[0]["text"] != "生成完整 HTML 鹈鹕作品" {
		t.Fatalf("content=%#v", messages[0]["content"])
	}
}

func TestClaudePelicanPayloadUsesLargerOutputBudget(t *testing.T) {
	ordinary, err := createTestPayloadWithPrompt("claude-sonnet-4-5", "hi")
	if err != nil {
		t.Fatalf("ordinary payload: %v", err)
	}
	if got := int(ordinary["max_tokens"].(int)); got != 1024 {
		t.Fatalf("ordinary max_tokens=%d, want 1024", got)
	}
	pelican, err := createTestPayloadWithMaxTokens("claude-sonnet-4-5", PelicanPrompt, pelicanClaudeMaxOutputTokens)
	if err != nil {
		t.Fatalf("pelican payload: %v", err)
	}
	if got := int(pelican["max_tokens"].(int)); got != pelicanClaudeMaxOutputTokens {
		t.Fatalf("pelican max_tokens=%d, want %d", got, pelicanClaudeMaxOutputTokens)
	}
}

type pelicanAdapterAccountRepo struct {
	AccountRepository
	account *Account
}

func (r pelicanAdapterAccountRepo) GetByID(context.Context, int64) (*Account, error) {
	return r.account, nil
}

type pelicanAdapterUpstream struct {
	HTTPUpstream
	mu       sync.Mutex
	requests []*http.Request
	body     [][]byte
	response func(*http.Request) (*http.Response, error)
}

type pelicanCloseCountingBody struct {
	io.Reader
	mu     sync.Mutex
	closes int
}

func (b *pelicanCloseCountingBody) Close() error {
	b.mu.Lock()
	b.closes++
	b.mu.Unlock()
	return nil
}

func (b *pelicanCloseCountingBody) CloseCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.closes
}

func (u *pelicanAdapterUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return u.record(req)
}

func (u *pelicanAdapterUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.record(req)
}

func (u *pelicanAdapterUpstream) record(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
	req.Body = io.NopCloser(strings.NewReader(string(body)))
	u.mu.Lock()
	u.requests = append(u.requests, req)
	u.body = append(u.body, append([]byte(nil), body...))
	u.mu.Unlock()
	return u.response(req)
}

func pelicanSSE(body string) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
}

func TestRunPelicanTestPlatformAdaptersUseExactPromptAndCompleteHTML(t *testing.T) {
	const prompt = "  生成鹈鹕自行车 HTML\n保留空格  "
	const html = "<!doctype html><html><body>pelican</body></html>"
	tests := []struct {
		name       string
		account    *Account
		response   string
		model      string
		assertBody func(*testing.T, []byte)
	}{
		{
			name:     "anthropic api key",
			model:    "test-model",
			account:  &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://anthropic.invalid"}},
			response: "data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"" + html + "\"}}\n\ndata: {\"type\":\"message_stop\"}\n\n",
			assertBody: func(t *testing.T, body []byte) {
				t.Helper()
				var payload map[string]any
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				messages := payload["messages"].([]any)
				content := messages[0].(map[string]any)["content"].([]any)
				if content[0].(map[string]any)["text"] != prompt {
					t.Fatalf("prompt=%q", content[0].(map[string]any)["text"])
				}
			},
		},
		{
			name:     "gemini api key",
			model:    "test-model",
			account:  &Account{ID: 2, Platform: PlatformGemini, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://gemini.invalid"}},
			response: "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"" + html + "\"}]},\"finishReason\":\"STOP\"}]}\n",
			assertBody: func(t *testing.T, body []byte) {
				t.Helper()
				var payload map[string]any
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				contents := payload["contents"].([]any)
				parts := contents[0].(map[string]any)["parts"].([]any)
				if parts[0].(map[string]any)["text"] != prompt {
					t.Fatalf("prompt=%q", parts[0].(map[string]any)["text"])
				}
			},
		},
		{
			name:     "grok oauth",
			model:    "grok-4.3",
			account:  &Account{ID: 3, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "fixture-token", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "base_url": "https://api.x.ai/v1"}},
			response: "data: {\"type\":\"response.output_text.delta\",\"delta\":\"" + html + "\"}\n\ndata: {\"type\":\"response.completed\"}\n\n",
			assertBody: func(t *testing.T, body []byte) {
				t.Helper()
				var payload map[string]any
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				if payload["input"] != prompt {
					t.Fatalf("prompt=%q", payload["input"])
				}
			},
		},
		{
			name:     "kimi chat completions",
			model:    "test-model",
			account:  &Account{ID: 4, Platform: PlatformKimi, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://kimi.invalid", "api_protocol": APIProtocolChatCompletions}},
			response: "data: {\"choices\":[{\"delta\":{\"content\":\"" + html + "\"}}]}\n\ndata: [DONE]\n\n",
			assertBody: func(t *testing.T, body []byte) {
				t.Helper()
				var payload map[string]any
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				messages := payload["messages"].([]any)
				if messages[1].(map[string]any)["content"] != prompt {
					t.Fatalf("prompt=%q", messages[1].(map[string]any)["content"])
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) { return pelicanSSE(tc.response) }}
			svc := &AccountTestService{accountRepo: pelicanAdapterAccountRepo{account: tc.account}, httpUpstream: upstream, cfg: &config.Config{}, grokTokenProvider: NewGrokTokenProvider(pelicanAdapterAccountRepo{account: tc.account}, nil)}
			result, err := svc.RunPelicanTest(context.Background(), tc.account.ID, tc.model, prompt, "")
			if err != nil {
				t.Fatalf("RunPelicanTest: %v", err)
			}
			if result.HTML != html {
				t.Fatalf("html=%q", result.HTML)
			}
			upstream.mu.Lock()
			calls, body := len(upstream.requests), append([]byte(nil), upstream.body[0]...)
			upstream.mu.Unlock()
			if calls != 1 {
				t.Fatalf("calls=%d", calls)
			}
			tc.assertBody(t, body)
		})
	}
}

func TestRunPelicanTestEveryPlatformProtocolSuccessClosesBodyOnce(t *testing.T) {
	const prompt = "  八平台精确提示词\n保留两端空格  "
	const html = "<html><body>eight-platform</body></html>"
	for _, tc := range []struct {
		name       string
		account    func() *Account
		model      string
		terminal   string
		assertBody func(*testing.T, []byte)
	}{
		{
			"openai responses", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformOpenAI) }, "gpt-test", "response.completed",
			func(t *testing.T, body []byte) { pelicanAssertResponsesPrompt(t, body, prompt) },
		},
		{
			"anthropic", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformAnthropic) }, "claude-test", "message_stop",
			func(t *testing.T, body []byte) { pelicanAssertAnthropicPrompt(t, body, prompt) },
		},
		{
			"gemini", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformGemini) }, "gemini-test", "STOP",
			func(t *testing.T, body []byte) { pelicanAssertGeminiPrompt(t, body, prompt) },
		},
		{
			"grok", pelicanAdapterGrokAccount, "grok-4.3", "response.completed",
			func(t *testing.T, body []byte) {
				var p map[string]any
				if err := json.Unmarshal(body, &p); err != nil {
					t.Fatal(err)
				}
				if p["input"] != prompt {
					t.Fatalf("prompt=%q", p["input"])
				}
			},
		},
		{
			"antigravity", pelicanAdapterAntigravityAccount, "gemini-3-pro", "STOP",
			func(t *testing.T, body []byte) {
				var p map[string]any
				if err := json.Unmarshal(body, &p); err != nil {
					t.Fatal(err)
				}
				parts := p["request"].(map[string]any)["contents"].([]any)[0].(map[string]any)["parts"].([]any)
				if parts[0].(map[string]any)["text"] != prompt {
					t.Fatalf("prompt=%q", parts[0].(map[string]any)["text"])
				}
			},
		},
		{
			"kimi chat", func() *Account {
				a := pelicanAdapterAPIKeyAccount(PlatformKimi)
				a.Credentials["api_protocol"] = APIProtocolChatCompletions
				return a
			}, "kimi-test", "[DONE]",
			func(t *testing.T, body []byte) { pelicanAssertChatPrompt(t, body, prompt) },
		},
		{
			"zhipu anthropic", func() *Account {
				a := pelicanAdapterAPIKeyAccount(PlatformZhipu)
				a.Credentials["api_protocol"] = APIProtocolAnthropic
				return a
			}, "glm-test", "message_stop",
			func(t *testing.T, body []byte) { pelicanAssertAnthropicPrompt(t, body, prompt) },
		},
		{
			"deepseek responses", func() *Account {
				a := pelicanAdapterAPIKeyAccount(PlatformDeepseek)
				a.Credentials["api_protocol"] = APIProtocolResponses
				return a
			}, "deepseek-test", "response.completed",
			func(t *testing.T, body []byte) { pelicanAssertResponsesPrompt(t, body, prompt) },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &pelicanCloseCountingBody{Reader: strings.NewReader(pelicanProtocolSuccessSSE(tc.name, html))}
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: body}, nil
			}}
			account := tc.account()
			result, err := pelicanAdapterService(account, upstream).RunPelicanTest(context.Background(), account.ID, tc.model, prompt, "")
			if err != nil || result.HTML != html {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			upstream.mu.Lock()
			calls, requestBody := len(upstream.requests), append([]byte(nil), upstream.body[0]...)
			upstream.mu.Unlock()
			if calls != 1 {
				t.Fatalf("calls=%d", calls)
			}
			if body.CloseCount() != 1 {
				t.Fatalf("body closes=%d", body.CloseCount())
			}
			if !strings.Contains(pelicanProtocolSuccessSSE(tc.name, html), tc.terminal) {
				t.Fatalf("fixture missing terminal %q", tc.terminal)
			}
			tc.assertBody(t, requestBody)
		})
	}
}

func TestRunPelicanTestEveryPlatformCloses503BodyOnce(t *testing.T) {
	for _, tc := range pelicanAdapterPlatforms() {
		t.Run(tc.name, func(t *testing.T) {
			body := &pelicanCloseCountingBody{Reader: strings.NewReader("provider unavailable")}
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header), Body: body}, nil
			}}
			account := tc.account()
			_, err := pelicanAdapterService(account, upstream).RunPelicanTest(context.Background(), account.ID, tc.model, PelicanPrompt, "")
			if !errors.Is(err, ErrPelicanUpstreamRequest) {
				t.Fatalf("err=%v", err)
			}
			if body.CloseCount() != 1 {
				t.Fatalf("body closes=%d", body.CloseCount())
			}
		})
	}
}

func TestPelicanAdapterAccountTypeSupportedAllowsOpenAIAndAnthropicSetupToken(t *testing.T) {
	for _, platform := range []string{PlatformOpenAI, PlatformAnthropic} {
		t.Run(platform, func(t *testing.T) {
			if !pelicanAdapterAccountTypeSupported(&Account{Platform: platform, Type: AccountTypeSetupToken}) {
				t.Fatalf("%s setup token rejected", platform)
			}
		})
	}
}

func TestRunPelicanTestRejectsGeminiSetupTokenBeforeAnyProviderCall(t *testing.T) {
	account := &Account{
		ID:          1,
		Platform:    PlatformGemini,
		Type:        AccountTypeSetupToken,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{"setup_token": "fixture-token"},
	}
	upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
		t.Fatal("gemini setup token reached upstream")
		return nil, nil
	}}

	_, err := pelicanAdapterService(account, upstream).RunPelicanTest(context.Background(), account.ID, "gemini-test", PelicanPrompt, "")
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("err=%v", err)
	}
	upstream.mu.Lock()
	calls := len(upstream.requests)
	upstream.mu.Unlock()
	if calls != 0 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestRunPelicanTestAnthropicSpecialAccountRoutesPreservePrompt(t *testing.T) {
	const prompt = "  Vertex 和 Bedrock 必须保留的 Pelican HTML 提示词  "
	const html = "<html><body>pelican</body></html>"
	for _, tc := range []struct {
		name     string
		account  *Account
		prepare  func(*AccountTestService)
		response string
		assert   func(*testing.T, *http.Request, []byte)
	}{
		{
			name:    "vertex service account",
			account: pelicanAdapterVertexAccount(t, PlatformAnthropic),
			prepare: func(svc *AccountTestService) {
				svc.claudeTokenProvider = NewClaudeTokenProvider(svc.accountRepo, pelicanStaticTokenCache{token: "vertex-token"}, nil)
			},
			response: "data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"" + html + "\"}}\n\ndata: {\"type\":\"message_stop\"}\n",
			assert: func(t *testing.T, req *http.Request, body []byte) {
				if req.Header.Get("Authorization") != "Bearer vertex-token" || !strings.Contains(req.URL.String(), "/projects/vertex-project/locations/us-central1/") {
					t.Fatalf("url=%s authorization=%q", req.URL, req.Header.Get("Authorization"))
				}
				pelicanAssertAnthropicPrompt(t, body, prompt)
			},
		},
		{
			name: "bedrock api key",
			account: &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeBedrock, Status: StatusActive, Schedulable: true, Credentials: map[string]any{
				"auth_mode": "apikey", "api_key": "bedrock-token", "aws_region": "us-east-1",
			}},
			response: `{"content":[{"text":"` + html + `"}],"stop_reason":"end_turn"}`,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				if req.Header.Get("Authorization") != "Bearer bedrock-token" {
					t.Fatalf("authorization=%q", req.Header.Get("Authorization"))
				}
				var payload map[string]any
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatal(err)
				}
				content := payload["messages"].([]any)[0].(map[string]any)["content"].([]any)
				if content[0].(map[string]any)["text"] != prompt || int(payload["max_tokens"].(float64)) != pelicanAntigravityMaxOutputTokens {
					t.Fatalf("payload=%#v", payload)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &pelicanAdapterUpstream{response: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.response))}, nil
			}}
			svc := pelicanAdapterService(tc.account, upstream)
			if tc.prepare != nil {
				tc.prepare(svc)
			}
			result, err := svc.RunPelicanTest(context.Background(), tc.account.ID, "claude-sonnet-4-5", prompt, "")
			if err != nil || result.HTML != html {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			upstream.mu.Lock()
			req, body := upstream.requests[0], append([]byte(nil), upstream.body[0]...)
			upstream.mu.Unlock()
			tc.assert(t, req, body)
		})
	}
}

func TestRunPelicanTestRejectsAnthropicAndBedrockIncompleteTerminals(t *testing.T) {
	const html = "<html><body>closed-but-truncated</body></html>"
	for _, tc := range []struct {
		name, response string
		account        *Account
	}{
		{"anthropic eof", "data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"" + html + "\"}}\n", pelicanAdapterAPIKeyAccount(PlatformAnthropic)},
		{"anthropic done without message stop", "data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"" + html + "\"}}\n\ndata: [DONE]\n\n", pelicanAdapterAPIKeyAccount(PlatformAnthropic)},
		{"anthropic max tokens", "data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"" + html + "\"}}\n\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"max_tokens\"}}\n\ndata: {\"type\":\"message_stop\"}\n", pelicanAdapterAPIKeyAccount(PlatformAnthropic)},
		{"bedrock max tokens", `{"content":[{"text":"` + html + `"}],"stop_reason":"max_tokens"}`, &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeBedrock, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"auth_mode": "apikey", "api_key": "fixture-key"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.response))}, nil
			}}
			_, err := pelicanAdapterService(tc.account, upstream).RunPelicanTest(context.Background(), tc.account.ID, "claude-sonnet-4-5", PelicanPrompt, "")
			if !errors.Is(err, ErrPelicanIncompleteHTML) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestClaudeOrdinaryProbeStillAcceptsDone(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	svc := &AccountTestService{}
	if err := svc.processClaudeStream(c, strings.NewReader("data: [DONE]\n\n")); err != nil {
		t.Fatalf("ordinary probe rejected [DONE]: %v", err)
	}
	if !strings.Contains(recorder.Body.String(), `"success":true`) {
		t.Fatalf("ordinary probe did not complete: %s", recorder.Body.String())
	}
}

func TestRunPelicanTestRejectsExplicitOpenAICompatibleTruncation(t *testing.T) {
	const html = "<html><body>closed-but-truncated</body></html>"
	for _, tc := range []struct {
		name, model, response string
		account               *Account
	}{
		{"responses incomplete", "gpt-test", "data: {\"type\":\"response.output_text.delta\",\"delta\":\"" + html + "\"}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"status\":\"incomplete\"}}\n", pelicanAdapterAPIKeyAccount(PlatformOpenAI)},
		{"chat length", "kimi-test", "data: {\"choices\":[{\"delta\":{\"content\":\"" + html + "\"},\"finish_reason\":\"length\"}]}\n", &Account{ID: 1, Platform: PlatformKimi, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://kimi.invalid", "api_protocol": APIProtocolChatCompletions}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) { return pelicanSSE(tc.response) }}
			_, err := pelicanAdapterService(tc.account, upstream).RunPelicanTest(context.Background(), tc.account.ID, tc.model, PelicanPrompt, "")
			if !errors.Is(err, ErrPelicanIncompleteHTML) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestRunPelicanTestGeminiOAuthAndServiceAccountRoutes(t *testing.T) {
	const prompt = " Gemini OAuth 和 SA 精确提示词 "
	const html = "<html><body>gemini-auth</body></html>"
	for _, tc := range []struct {
		name    string
		account *Account
		prepare func(*AccountTestService)
		assert  func(*testing.T, *http.Request, []byte)
	}{
		{"oauth ai studio", &Account{ID: 1, Platform: PlatformGemini, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "oauth-token", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "base_url": "https://gemini.invalid"}}, nil, func(t *testing.T, req *http.Request, body []byte) {
			if req.Header.Get("Authorization") != "Bearer oauth-token" {
				t.Fatalf("authorization=%q", req.Header.Get("Authorization"))
			}
			pelicanAssertGeminiPrompt(t, body, prompt)
		}},
		{"oauth code assist", &Account{ID: 1, Platform: PlatformGemini, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "code-token", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "project_id": "code-project"}}, nil, func(t *testing.T, req *http.Request, body []byte) {
			if req.Header.Get("Authorization") != "Bearer code-token" || !strings.Contains(req.URL.Path, "v1internal:streamGenerateContent") {
				t.Fatalf("url=%s authorization=%q", req.URL, req.Header.Get("Authorization"))
			}
			var wrapped map[string]any
			if err := json.Unmarshal(body, &wrapped); err != nil {
				t.Fatal(err)
			}
			pelicanAssertGeminiPrompt(t, pelicanMustJSON(t, wrapped["request"]), prompt)
		}},
		{"service account", pelicanAdapterVertexAccount(t, PlatformGemini), func(svc *AccountTestService) {
			svc.geminiTokenProvider = NewGeminiTokenProvider(svc.accountRepo, pelicanStaticTokenCache{token: "sa-token"}, nil)
		}, func(t *testing.T, req *http.Request, body []byte) {
			if req.Header.Get("Authorization") != "Bearer sa-token" || !strings.Contains(req.URL.String(), "/projects/vertex-project/locations/us-central1/") {
				t.Fatalf("url=%s authorization=%q", req.URL, req.Header.Get("Authorization"))
			}
			pelicanAssertGeminiPrompt(t, body, prompt)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
				return pelicanSSE("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"" + html + "\"}]},\"finishReason\":\"STOP\"}]}\n")
			}}
			svc := pelicanAdapterService(tc.account, upstream)
			if tc.prepare != nil {
				tc.prepare(svc)
			}
			result, err := svc.RunPelicanTest(context.Background(), tc.account.ID, "gemini-test", prompt, "")
			if err != nil || result.HTML != html {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			upstream.mu.Lock()
			req, body := upstream.requests[0], append([]byte(nil), upstream.body[0]...)
			upstream.mu.Unlock()
			tc.assert(t, req, body)
		})
	}
}

type pelicanStaticTokenCache struct{ token string }

func (c pelicanStaticTokenCache) GetAccessToken(context.Context, string) (string, error) {
	return c.token, nil
}
func (pelicanStaticTokenCache) SetAccessToken(context.Context, string, string, time.Duration) error {
	return nil
}
func (pelicanStaticTokenCache) DeleteAccessToken(context.Context, string) error { return nil }
func (pelicanStaticTokenCache) AcquireRefreshLock(context.Context, string, time.Duration) (bool, error) {
	return false, nil
}
func (pelicanStaticTokenCache) ReleaseRefreshLock(context.Context, string) error { return nil }

func pelicanAdapterVertexAccount(t *testing.T, platform string) *Account {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	raw, err := json.Marshal(map[string]string{"type": "service_account", "project_id": "vertex-project", "private_key_id": "fixture-kid", "private_key": string(pemKey), "client_email": "fixture@vertex-project.iam.gserviceaccount.com"})
	if err != nil {
		t.Fatal(err)
	}
	return &Account{ID: 1, Platform: platform, Type: AccountTypeServiceAccount, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"service_account_json": string(raw), "location": "us-central1"}}
}

func pelicanMustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func pelicanProtocolSuccessSSE(name, html string) string {
	switch name {
	case "anthropic", "zhipu anthropic":
		return "data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"" + html + "\"}}\n\ndata: {\"type\":\"message_stop\"}\n"
	case "gemini":
		return "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"" + html + "\"}]},\"finishReason\":\"STOP\"}]}\n"
	case "antigravity":
		return "data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"" + html + "\"}]},\"finishReason\":\"STOP\"}]}}\n"
	case "kimi chat":
		return "data: {\"choices\":[{\"delta\":{\"content\":\"" + html + "\"}}]}\n\ndata: [DONE]\n"
	default:
		return "data: {\"type\":\"response.output_text.delta\",\"delta\":\"" + html + "\"}\n\ndata: {\"type\":\"response.completed\"}\n"
	}
}

func pelicanAssertResponsesPrompt(t *testing.T, body []byte, want string) {
	t.Helper()
	var p map[string]any
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatal(err)
	}
	text := p["input"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"]
	if text != want {
		t.Fatalf("prompt=%q", text)
	}
}
func pelicanAssertAnthropicPrompt(t *testing.T, body []byte, want string) {
	t.Helper()
	var p map[string]any
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatal(err)
	}
	text := p["messages"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"]
	if text != want {
		t.Fatalf("prompt=%q", text)
	}
}
func pelicanAssertGeminiPrompt(t *testing.T, body []byte, want string) {
	t.Helper()
	var p map[string]any
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatal(err)
	}
	text := p["contents"].([]any)[0].(map[string]any)["parts"].([]any)[0].(map[string]any)["text"]
	if text != want {
		t.Fatalf("prompt=%q", text)
	}
}
func pelicanAssertChatPrompt(t *testing.T, body []byte, want string) {
	t.Helper()
	var p map[string]any
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatal(err)
	}
	text := p["messages"].([]any)[1].(map[string]any)["content"]
	if text != want {
		t.Fatalf("prompt=%q", text)
	}
}

func TestRunPelicanTestGrokDoClassifiesFailureAndHonorsCancellation(t *testing.T) {
	account := &Account{ID: 3, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "fixture-token", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "base_url": "https://api.x.ai/v1"}}
	for _, tc := range []struct {
		name            string
		err             error
		ctx             func() (context.Context, context.CancelFunc)
		cancelBeforeRun bool
		want            error
	}{
		{name: "request error", err: errors.New("provider unavailable"), ctx: func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }, want: ErrPelicanUpstreamRequest},
		{name: "cancelled", err: context.Canceled, ctx: func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }, cancelBeforeRun: true, want: ErrPelicanUpstreamRequest},
		{name: "deadline", err: context.DeadlineExceeded, ctx: func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }, want: ErrPelicanUpstreamTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) { return nil, tc.err }}
			svc := &AccountTestService{accountRepo: pelicanAdapterAccountRepo{account: account}, httpUpstream: upstream, cfg: &config.Config{}, grokTokenProvider: NewGrokTokenProvider(pelicanAdapterAccountRepo{account: account}, nil)}
			ctx, cancel := tc.ctx()
			defer cancel()
			if tc.cancelBeforeRun {
				cancel()
			}
			_, err := svc.RunPelicanTest(ctx, account.ID, "grok-4.3", PelicanPrompt, "")
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
		})
	}
}

func TestRunPelicanTestRejectsIneligibleModelsBeforeAnyProviderCall(t *testing.T) {
	platforms := []string{PlatformOpenAI, PlatformAnthropic, PlatformGemini, PlatformGrok, PlatformAntigravity, PlatformKimi, PlatformZhipu, PlatformDeepseek}
	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			account := &Account{
				ID: 1, Platform: platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true,
				Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://provider.invalid", "model_mapping": map[string]any{"allowed-model": "allowed-model"}},
			}
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
				t.Fatal("ineligible model reached upstream")
				return nil, nil
			}}
			svc := &AccountTestService{accountRepo: pelicanAdapterAccountRepo{account: account}, httpUpstream: upstream, cfg: &config.Config{}}
			_, err := svc.RunPelicanTest(context.Background(), account.ID, "rejected-model", PelicanPrompt, "")
			if err == nil || !strings.Contains(err.Error(), "unsupported") {
				t.Fatalf("err=%v", err)
			}
			upstream.mu.Lock()
			calls := len(upstream.requests)
			upstream.mu.Unlock()
			if calls != 0 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}

func TestRunPelicanTestRejectsUnsupportedAccountTypesBeforeAnyProviderCall(t *testing.T) {
	platforms := []string{PlatformOpenAI, PlatformAnthropic, PlatformGemini, PlatformGrok, PlatformAntigravity, PlatformKimi, PlatformZhipu, PlatformDeepseek}
	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			account := &Account{ID: 1, Platform: platform, Type: "unsupported_fixture_type", Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-key"}}
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
				t.Fatal("unsupported account type reached upstream")
				return nil, nil
			}}
			svc := &AccountTestService{accountRepo: pelicanAdapterAccountRepo{account: account}, httpUpstream: upstream, cfg: &config.Config{}}
			_, err := svc.RunPelicanTest(context.Background(), account.ID, "test-model", PelicanPrompt, "")
			if err == nil || !strings.Contains(err.Error(), "unsupported") {
				t.Fatalf("err=%v", err)
			}
			upstream.mu.Lock()
			calls := len(upstream.requests)
			upstream.mu.Unlock()
			if calls != 0 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}

func TestRunPelicanTestCNProtocolAdaptersPreservePrompt(t *testing.T) {
	const prompt = "  CN 完整 HTML 提示词\n不裁剪  "
	const html = "<html><body>cn</body></html>"
	for _, tc := range []struct {
		name     string
		platform string
		protocol string
		response string
	}{
		{"zhipu anthropic", PlatformZhipu, APIProtocolAnthropic, "data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"" + html + "\"}}\n\ndata: {\"type\":\"message_stop\"}\n\n"},
		{"deepseek responses", PlatformDeepseek, APIProtocolResponses, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"" + html + "\"}\n\ndata: {\"type\":\"response.completed\"}\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := &Account{ID: 1, Platform: tc.platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://cn.invalid", "api_protocol": tc.protocol}}
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) { return pelicanSSE(tc.response) }}
			svc := &AccountTestService{accountRepo: pelicanAdapterAccountRepo{account: account}, httpUpstream: upstream, cfg: &config.Config{}}
			result, err := svc.RunPelicanTest(context.Background(), account.ID, "test-model", prompt, "")
			if err != nil || result.HTML != html {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			upstream.mu.Lock()
			body := append([]byte(nil), upstream.body[0]...)
			upstream.mu.Unlock()
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatal(err)
			}
			if tc.protocol == APIProtocolAnthropic {
				content := payload["messages"].([]any)[0].(map[string]any)["content"].([]any)
				if content[0].(map[string]any)["text"] != prompt {
					t.Fatalf("prompt=%q", content[0].(map[string]any)["text"])
				}
			} else {
				input := payload["input"].([]any)[0].(map[string]any)["content"].([]any)
				if input[0].(map[string]any)["text"] != prompt {
					t.Fatalf("prompt=%q", input[0].(map[string]any)["text"])
				}
			}
		})
	}
}

func TestRunPelicanTestAntigravityOAuthUsesSingleCappedGatewayRequest(t *testing.T) {
	const prompt = "生成完整 Antigravity 鹈鹕 HTML"
	const html = "<html><body>antigravity</body></html>"
	account := &Account{ID: 1, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{
		"access_token": "fixture-token", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "project_id": "fixture-project",
		"model_mapping": map[string]any{"gemini-3-pro": "gemini-3-pro"},
	}}
	repo := pelicanAdapterAccountRepo{account: account}
	upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
		return pelicanSSE("data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"" + html + "\"}]},\"finishReason\":\"STOP\"}]}}\n")
	}}
	gateway := &AntigravityGatewayService{tokenProvider: NewAntigravityTokenProvider(repo, nil, nil)}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: &config.Config{}, antigravityGatewayService: gateway}
	result, err := svc.RunPelicanTest(context.Background(), account.ID, "gemini-3-pro", prompt, "")
	if err != nil || result.HTML != html {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	upstream.mu.Lock()
	calls, body := len(upstream.requests), append([]byte(nil), upstream.body[0]...)
	upstream.mu.Unlock()
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
	var wrapped map[string]any
	if err := json.Unmarshal(body, &wrapped); err != nil {
		t.Fatal(err)
	}
	request := wrapped["request"].(map[string]any)
	parts := request["contents"].([]any)[0].(map[string]any)["parts"].([]any)
	if parts[0].(map[string]any)["text"] != prompt {
		t.Fatalf("prompt=%q", parts[0].(map[string]any)["text"])
	}
	if request["generationConfig"].(map[string]any)["maxOutputTokens"] != float64(pelicanAntigravityMaxOutputTokens) {
		t.Fatalf("generationConfig=%#v", request["generationConfig"])
	}
}

func TestRunPelicanTestEveryPlatformClassifiesProviderErrorTimeoutAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		platform string
		model    string
		account  func() *Account
		header   string
		value    string
	}{
		{"openai", PlatformOpenAI, "gpt-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformOpenAI) }, "Authorization", "Bearer fixture-key"},
		{"anthropic", PlatformAnthropic, "claude-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformAnthropic) }, "x-api-key", "fixture-key"},
		{"gemini", PlatformGemini, "gemini-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformGemini) }, "x-goog-api-key", "fixture-key"},
		{"grok", PlatformGrok, "grok-4.3", func() *Account { return pelicanAdapterGrokAccount() }, "Authorization", "Bearer fixture-token"},
		{"antigravity", PlatformAntigravity, "gemini-3-pro", func() *Account { return pelicanAdapterAntigravityAccount() }, "Authorization", "Bearer fixture-token"},
		{"kimi", PlatformKimi, "kimi-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformKimi) }, "Authorization", "Bearer fixture-key"},
		{"zhipu", PlatformZhipu, "glm-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformZhipu) }, "Authorization", "Bearer fixture-key"},
		{"deepseek", PlatformDeepseek, "deepseek-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformDeepseek) }, "Authorization", "Bearer fixture-key"},
	} {
		for _, outcome := range []struct {
			name string
			err  error
			want error
		}{
			{"provider error", errors.New("fixture provider error"), ErrPelicanUpstreamRequest},
			{"cancelled", context.Canceled, ErrPelicanUpstreamRequest},
			{"timeout", context.DeadlineExceeded, ErrPelicanUpstreamTimeout},
		} {
			t.Run(tc.name+"/"+outcome.name, func(t *testing.T) {
				account := tc.account()
				upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) { return nil, outcome.err }}
				svc := pelicanAdapterService(account, upstream)
				_, err := svc.RunPelicanTest(context.Background(), account.ID, tc.model, PelicanPrompt, "")
				if !errors.Is(err, outcome.want) {
					t.Fatalf("err=%v want=%v", err, outcome.want)
				}
				upstream.mu.Lock()
				calls := len(upstream.requests)
				header := ""
				headers := http.Header(nil)
				if calls == 1 {
					headers = upstream.requests[0].Header.Clone()
					for key, values := range headers {
						if strings.EqualFold(key, tc.header) && len(values) > 0 {
							header = values[0]
							break
						}
					}
				}
				upstream.mu.Unlock()
				if calls != 1 {
					t.Fatalf("calls=%d", calls)
				}
				if header != tc.value {
					t.Fatalf("%s=%q want=%q headers=%v", tc.header, header, tc.value, headers)
				}
			})
		}
	}
}

func TestRunPelicanTestRejectsGeminiTokenLimitTruncation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		account *Account
		model   string
		gateway bool
	}{
		{"gemini", pelicanAdapterAPIKeyAccount(PlatformGemini), "gemini-test", false},
		{"antigravity envelope", pelicanAdapterAntigravityAccount(), "gemini-3-pro", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"<html>partial\"}]},\"finishReason\":\"MAX_TOKENS\"}]}\n"
			if tc.gateway {
				payload = "data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"<html>partial\"}]},\"finishReason\":\"MAX_TOKENS\"}]}}\n"
			}
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) { return pelicanSSE(payload) }}
			svc := pelicanAdapterService(tc.account, upstream)
			_, err := svc.RunPelicanTest(context.Background(), tc.account.ID, tc.model, PelicanPrompt, "")
			if !errors.Is(err, ErrPelicanIncompleteHTML) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestRunPelicanTestGeminiUsesWildcardMappedModel(t *testing.T) {
	account := pelicanAdapterAPIKeyAccount(PlatformGemini)
	account.Credentials["model_mapping"] = map[string]any{"gemini-*": "gemini-3-pro-upstream"}
	upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
		return pelicanSSE("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"<html>wildcard</html>\"}]},\"finishReason\":\"STOP\"}]}\n")
	}}
	svc := pelicanAdapterService(account, upstream)
	result, err := svc.RunPelicanTest(context.Background(), account.ID, "gemini-2.5-pro", PelicanPrompt, "")
	if err != nil || result.HTML != "<html>wildcard</html>" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	upstream.mu.Lock()
	url := upstream.requests[0].URL.String()
	upstream.mu.Unlock()
	if !strings.Contains(url, "/models/gemini-3-pro-upstream:streamGenerateContent") {
		t.Fatalf("url=%q", url)
	}
}

func TestRunPelicanTestEveryPlatformHonorsBlockedRequestContext(t *testing.T) {
	for _, tc := range pelicanAdapterPlatforms() {
		for _, outcome := range []struct {
			name string
			ctx  func() (context.Context, context.CancelFunc)
			want error
		}{
			{"cancel", func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }, ErrPelicanUpstreamRequest},
			{"deadline", func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 20*time.Millisecond)
			}, ErrPelicanUpstreamTimeout},
		} {
			t.Run(tc.name+"/"+outcome.name, func(t *testing.T) {
				started := make(chan struct{})
				upstream := &pelicanAdapterUpstream{response: func(req *http.Request) (*http.Response, error) {
					close(started)
					<-req.Context().Done()
					return nil, req.Context().Err()
				}}
				svc := pelicanAdapterService(tc.account(), upstream)
				ctx, cancel := outcome.ctx()
				defer cancel()
				done := make(chan error, 1)
				go func() { _, err := svc.RunPelicanTest(ctx, 1, tc.model, PelicanPrompt, ""); done <- err }()
				select {
				case <-started:
					if outcome.name == "cancel" {
						cancel()
					}
				case <-time.After(time.Second):
					t.Fatal("provider request did not start")
				}
				select {
				case err := <-done:
					if !errors.Is(err, outcome.want) {
						t.Fatalf("err=%v want=%v", err, outcome.want)
					}
				case <-time.After(time.Second):
					t.Fatal("RunPelicanTest ignored request context")
				}
			})
		}
	}
}

func TestRunPelicanTestEveryPlatformRejectsMissingHTMLTerminalAnd503(t *testing.T) {
	for _, tc := range pelicanAdapterPlatforms() {
		for _, outcome := range []struct {
			name   string
			body   func(bool) string
			status int
		}{
			{"missing html", pelicanAdapterSSE, http.StatusOK},
			{"missing terminal", func(envelope bool) string { return pelicanAdapterSSEWithoutTerminal(envelope) }, http.StatusOK},
			{"http 503", func(bool) string { return "provider unavailable" }, http.StatusServiceUnavailable},
		} {
			t.Run(tc.name+"/"+outcome.name, func(t *testing.T) {
				upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: outcome.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(outcome.body(tc.antigravity)))}, nil
				}}
				svc := pelicanAdapterService(tc.account(), upstream)
				_, err := svc.RunPelicanTest(context.Background(), 1, tc.model, PelicanPrompt, "")
				if err == nil {
					t.Fatal("invalid provider response succeeded")
				}
			})
		}
	}
}

func TestRunPelicanTestGeminiRejectsEveryNonSTOPTerminal(t *testing.T) {
	for _, reason := range []string{"MAX_TOKENS", "SAFETY", "RECITATION", "OTHER"} {
		t.Run(reason, func(t *testing.T) {
			body := "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"<html>partial</html>\"}]},\"finishReason\":\"" + reason + "\"}]}\n"
			upstream := &pelicanAdapterUpstream{response: func(*http.Request) (*http.Response, error) { return pelicanSSE(body) }}
			svc := pelicanAdapterService(pelicanAdapterAPIKeyAccount(PlatformGemini), upstream)
			_, err := svc.RunPelicanTest(context.Background(), 1, "gemini-test", PelicanPrompt, "")
			if !errors.Is(err, ErrPelicanIncompleteHTML) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

type pelicanAdapterPlatform struct {
	name        string
	model       string
	account     func() *Account
	antigravity bool
}

func pelicanAdapterPlatforms() []pelicanAdapterPlatform {
	return []pelicanAdapterPlatform{
		{"openai", "gpt-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformOpenAI) }, false},
		{"anthropic", "claude-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformAnthropic) }, false},
		{"gemini", "gemini-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformGemini) }, false},
		{"grok", "grok-4.3", pelicanAdapterGrokAccount, false},
		{"antigravity", "gemini-3-pro", pelicanAdapterAntigravityAccount, true},
		{"kimi", "kimi-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformKimi) }, false},
		{"zhipu", "glm-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformZhipu) }, false},
		{"deepseek", "deepseek-test", func() *Account { return pelicanAdapterAPIKeyAccount(PlatformDeepseek) }, false},
	}
}

func pelicanAdapterSSE(antigravity bool) string {
	if antigravity {
		return "data: {\"response\":{\"candidates\":[{\"finishReason\":\"STOP\"}]}}\n"
	}
	return "data: {\"candidates\":[{\"finishReason\":\"STOP\"}]}\n"
}

func pelicanAdapterSSEWithoutTerminal(antigravity bool) string {
	if antigravity {
		return "data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"<html>partial\"}]}}]}}\n"
	}
	return "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"<html>partial\"}]}}]}\n"
}

func pelicanAdapterAPIKeyAccount(platform string) *Account {
	return &Account{ID: 1, Platform: platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{
		"api_key": "fixture-key", "base_url": "https://provider.invalid",
	}}
}

func pelicanAdapterGrokAccount() *Account {
	return &Account{ID: 1, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{
		"access_token": "fixture-token", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "base_url": "https://api.x.ai/v1",
	}}
}

func pelicanAdapterAntigravityAccount() *Account {
	return &Account{ID: 1, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{
		"access_token": "fixture-token", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "project_id": "fixture-project",
		"model_mapping": map[string]any{"gemini-3-pro": "gemini-3-pro"},
	}}
}

func pelicanAdapterService(account *Account, upstream HTTPUpstream) *AccountTestService {
	repo := pelicanAdapterAccountRepo{account: account}
	svc := &AccountTestService{
		accountRepo:         repo,
		httpUpstream:        upstream,
		cfg:                 &config.Config{},
		grokTokenProvider:   NewGrokTokenProvider(repo, nil),
		geminiTokenProvider: NewGeminiTokenProvider(repo, nil, nil),
	}
	if account.Platform == PlatformAntigravity {
		svc.antigravityGatewayService = &AntigravityGatewayService{tokenProvider: NewAntigravityTokenProvider(repo, nil, nil)}
	}
	return svc
}

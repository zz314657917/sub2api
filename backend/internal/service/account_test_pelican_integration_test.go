package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
)

func TestPelicanProvider503AndBrokenStream(t *testing.T) {
	for _, tc := range []struct {
		status int
		broken bool
		code   string
	}{
		{503, false, "upstream_request_failed"},
		{200, true, "upstream_stream_closed"},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			if tc.broken {
				w.Header().Set("Content-Length", "99999")
			}
			w.WriteHeader(tc.status)
			fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"<html>\"}\n\n")
		}))
		account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-only", "base_url": srv.URL}}
		cfg := &config.Config{}
		cfg.Security.URLAllowlist.AllowInsecureHTTP = true
		svc := &AccountTestService{accountRepo: pelicanProviderAccounts{account: account}, cfg: cfg, httpUpstream: pelicanLocalUpstream{host: strings.TrimPrefix(srv.URL, "http://")}}
		_, err := svc.RunPelicanTest(context.Background(), 1, "gpt-test", PelicanPrompt, "")
		srv.Close()
		code, safe := classifyPelicanFailure(err, nil)
		if code != tc.code {
			t.Fatalf("code=%s err=%v", code, err)
		}
		if tc.status == 503 && (!errors.Is(err, ErrPelicanUpstreamRequest) || !strings.Contains(safe, "503")) {
			t.Fatalf("missing status: %s", safe)
		}
	}
}

type pelicanProviderAccounts struct {
	AccountRepository
	account *Account
}

func (a pelicanProviderAccounts) GetByID(context.Context, int64) (*Account, error) {
	return a.account, nil
}

type pelicanLocalUpstream struct {
	HTTPUpstream
	host string
}

type pelicanCaptureUpstream struct {
	HTTPUpstream
	mu      sync.Mutex
	payload []byte
}

func (u *pelicanCaptureUpstream) DoWithTLS(r *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	payload, _ := io.ReadAll(r.Body)
	u.mu.Lock()
	u.payload = append([]byte(nil), payload...)
	u.mu.Unlock()
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("data: {\"type\":\"response.output_text.delta\",\"delta\":\"<html>" + strings.Repeat("x", 100) + "</html>\"}\n\ndata: {\"type\":\"response.completed\"}\n\n"))}, nil
}

func TestPelicanProviderOAuthPreservesExactPrompt(t *testing.T) {
	prompt := "  OAuth 第一行\n第二行  "
	upstream := &pelicanCaptureUpstream{}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "fixture-only"}}
	s := &AccountTestService{accountRepo: pelicanProviderAccounts{account: account}, cfg: &config.Config{}, httpUpstream: upstream}
	result, err := s.RunPelicanTest(context.Background(), 1, "gpt-test", prompt, "high")
	if err != nil || result == nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	upstream.mu.Lock()
	payload := append([]byte(nil), upstream.payload...)
	upstream.mu.Unlock()
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		t.Fatal(err)
	}
	input := data["input"].([]any)[0].(map[string]any)
	if data["instructions"] != pelicanTestInstructions {
		t.Fatalf("unexpected instructions: %v", data["instructions"])
	}
	content := input["content"].([]any)[0].(map[string]any)
	if content["text"] != prompt {
		t.Fatalf("OAuth prompt=%q want=%q", content["text"], prompt)
	}
	if reasoning, ok := data["reasoning"].(map[string]any); !ok || reasoning["effort"] != "high" {
		t.Fatalf("OAuth reasoning=%#v", data["reasoning"])
	}
}

func TestPelicanProviderReasoningEffortPayloadAndDefaultOmission(t *testing.T) {
	tests := []struct {
		name       string
		extra      map[string]any
		effort     string
		path       string
		assertBody func(*testing.T, map[string]any)
	}{
		{
			name:   "responses effort",
			effort: "high",
			path:   "/v1/responses",
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if reasoning, ok := body["reasoning"].(map[string]any); !ok || reasoning["effort"] != "high" {
					t.Fatalf("Responses reasoning=%#v", body["reasoning"])
				}
			},
		},
		{
			name: "responses default omits effort",
			path: "/v1/responses",
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if _, exists := body["reasoning"]; exists {
					t.Fatalf("default Responses payload unexpectedly included reasoning: %#v", body["reasoning"])
				}
			},
		},
		{
			name:   "chat fallback effort",
			extra:  map[string]any{"openai_responses_supported": false},
			effort: "medium",
			path:   "/v1/chat/completions",
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if body["reasoning_effort"] != "medium" {
					t.Fatalf("Chat reasoning_effort=%#v", body["reasoning_effort"])
				}
				if _, exists := body["reasoning"]; exists {
					t.Fatalf("Chat payload unexpectedly used Responses reasoning: %#v", body["reasoning"])
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var payload map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Errorf("path=%q want=%q", r.URL.Path, tc.path)
				}
				r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
				defer r.Body.Close()
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("decode payload: %v", err)
				}
				w.Header().Set("Content-Type", "text/event-stream")
				if tc.path == "/v1/chat/completions" {
					fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"<html>"+strings.Repeat("x", 100)+"</html>\"}}]}\n\ndata: [DONE]\n\n")
					return
				}
				fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"<html>"+strings.Repeat("x", 100)+"</html>\"}\n\ndata: {\"type\":\"response.completed\"}\n\n")
			}))
			defer srv.Close()
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Extra: tc.extra, Credentials: map[string]any{"api_key": "fixture-only", "base_url": srv.URL}}
			cfg := &config.Config{}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			s := &AccountTestService{accountRepo: pelicanProviderAccounts{account: account}, cfg: cfg, httpUpstream: pelicanLocalUpstream{host: strings.TrimPrefix(srv.URL, "http://")}}
			result, err := s.RunPelicanTest(context.Background(), 1, "gpt-test", PelicanPrompt, tc.effort)
			if err != nil || result == nil {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			tc.assertBody(t, payload)
			if tc.path == "/v1/chat/completions" {
				messages := payload["messages"].([]any)
				if messages[0].(map[string]any)["content"] != pelicanTestInstructions {
					t.Fatal("missing test system instruction")
				}
			} else if payload["instructions"] != pelicanTestInstructions {
				t.Fatal("missing test instructions")
			}
		})
	}
}

func (u pelicanLocalUpstream) DoWithTLS(r *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	if r.URL.Host != u.host {
		return nil, fmt.Errorf("test refuses nonlocal host")
	}
	return http.DefaultClient.Do(r)
}

func TestPelicanProviderLocalSSE(t *testing.T) {
	for _, tc := range []struct {
		name     string
		html     string
		complete bool
		fail     bool
		wantOK   bool
	}{
		{"unicode success", "<!doctype html><html><body>鹈鹕骑车</body></html>", true, false, true},
		{"incomplete stream", "<html><body>鹈鹕</body></html>", false, false, false},
		{"incomplete document", "<html><body>鹈鹕", true, false, false},
		{"provider failure", "<html><body>鹈鹕</body></html>", false, true, false},
		{"html byte cap", "<html>" + strings.Repeat("鸟", 90000) + "</html>", true, false, false},
		// JSON escaping expands the SSE frame well beyond 256 KiB. The HTML
		// itself is below the cap and must still be accepted.
		{"escaped SSE overhead", "<html>" + strings.Repeat("<svg></svg>", 20000) + "</html>", true, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var promptOK bool
			prompt := "  " + PelicanPrompt + "\n保留换行  "
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				payload, _ := io.ReadAll(r.Body)
				var data map[string]any
				_ = json.Unmarshal(payload, &data)
				input := data["input"].([]any)[0].(map[string]any)
				content := input["content"].([]any)[0].(map[string]any)
				promptOK = content["text"] == prompt && data["model"] == "gpt-test" && r.URL.Path == "/v1/responses"
				w.Header().Set("Content-Type", "text/event-stream")
				event, _ := json.Marshal(map[string]any{"type": "response.output_text.delta", "delta": tc.html})
				fmt.Fprintf(w, "data: %s\n\n", event)
				if tc.fail {
					fmt.Fprint(w, "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"message\":\"secret-provider-key\"}}}\n\n")
				}
				if tc.complete {
					fmt.Fprint(w, "data: {\"type\":\"response.completed\"}\n\n")
				}
			}))
			defer srv.Close()
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: "apikey", Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-only", "base_url": srv.URL}}
			cfg := &config.Config{}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			s := &AccountTestService{accountRepo: pelicanProviderAccounts{account: account}, cfg: cfg, httpUpstream: pelicanLocalUpstream{host: strings.TrimPrefix(srv.URL, "http://")}}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result, err := s.RunPelicanTest(ctx, 1, "gpt-test", prompt, "")
			if !promptOK {
				t.Fatal("provider request did not contain the real prompt/model/path")
			}
			if (err == nil) != tc.wantOK {
				t.Fatalf("result=%v err=%v", result, err)
			}
			if err != nil && strings.Contains(err.Error(), "secret-provider-key") {
				t.Fatal("private provider error leaked")
			}
			if tc.wantOK && (result.HTML != tc.html || result.CharCount != utf8.RuneCountInString(tc.html)) {
				t.Fatal("HTML or Unicode count changed")
			}
		})
	}
}

func TestPelicanProviderCancellationClosesLocalRequest(t *testing.T) {
	started, closed := make(chan struct{}), make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		close(started)
		<-r.Context().Done()
		close(closed)
	}))
	defer srv.Close()
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: "apikey", Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-only", "base_url": srv.URL}}
	cfg := &config.Config{}
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	s := &AccountTestService{accountRepo: pelicanProviderAccounts{account: account}, cfg: cfg, httpUpstream: pelicanLocalUpstream{host: strings.TrimPrefix(srv.URL, "http://")}}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := s.RunPelicanTest(ctx, 1, "gpt-test", PelicanPrompt, ""); done <- err }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("local request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled provider succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("provider ignored cancellation")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("local request was left open")
	}
}

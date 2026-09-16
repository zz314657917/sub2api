package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
)

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
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				payload, _ := io.ReadAll(r.Body)
				var data map[string]any
				_ = json.Unmarshal(payload, &data)
				promptOK = strings.Contains(string(payload), "CSS keyframes") && data["model"] == "gpt-test" && r.URL.Path == "/v1/responses"
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
			result, err := s.RunPelicanTest(ctx, 1, "gpt-test")
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
	go func() { _, err := s.RunPelicanTest(ctx, 1, "gpt-test"); done <- err }()
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

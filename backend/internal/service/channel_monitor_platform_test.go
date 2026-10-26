//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChannelMonitorOpenAICompatibleProviders(t *testing.T) {
	providers := []struct {
		name string
		want string
	}{
		{name: MonitorProviderOpenAI, want: "/v1/chat/completions"},
		{name: MonitorProviderGrok, want: "/v1/chat/completions"},
		{name: MonitorProviderKimi, want: "/v1/chat/completions"},
		// httptest is an OpenAI relay host; official Zhipu roots and Coding
		// Plan bases are covered by TestMonitorRequestPath_Zhipu.
		{name: MonitorProviderZhipu, want: "/v1/chat/completions"},
		{name: MonitorProviderDeepseek, want: "/v1/chat/completions"},
	}

	for _, tc := range providers {
		t.Run(tc.name, func(t *testing.T) {
			swapMonitorHTTPClient(t)
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				if body["model"] != "test-model" {
					t.Fatalf("model = %v, want test-model", body["model"])
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"the answer is 2"}}]}`))
			}))
			t.Cleanup(server.Close)

			result := runCheckForModel(
				context.Background(),
				tc.name,
				server.URL,
				"test-key",
				"test-model",
				&CheckOptions{
					BodyOverrideMode: MonitorBodyOverrideModeReplace,
					BodyOverride: map[string]any{
						"model":    "test-model",
						"messages": []any{map[string]any{"role": "user", "content": "ping"}},
					},
				},
			)
			if result.Status != MonitorStatusOperational {
				t.Fatalf("status = %s, message = %q", result.Status, result.Message)
			}
			if gotPath != tc.want {
				t.Fatalf("request path = %q, want %q", gotPath, tc.want)
			}
		})
	}
}

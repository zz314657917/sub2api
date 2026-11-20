package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestVertexPublisherURLSafety(t *testing.T) {
	for _, location := range []string{"global", "us-central1", "europe-west4"} {
		raw, err := buildVertexPublisherModelsURL(location, `opaque&x=y?/#%`+"\r\n")
		require.NoError(t, err)
		u, err := url.Parse(raw)
		require.NoError(t, err)
		require.Equal(t, "https", u.Scheme)
		require.True(t, strings.HasSuffix(u.Host, ".googleapis.com"))
		require.Equal(t, "/v1beta1/publishers/google/models", u.Path)
		require.Equal(t, `opaque&x=y?/#%`+"\r\n", u.Query().Get("pageToken"))
		require.Equal(t, "300", u.Query().Get("pageSize"))
		require.Len(t, u.Query(), 2)
		require.Empty(t, u.Fragment)
	}
	for _, location := range []string{"", "../global", "global.evil.test", "global/", "global?x", "global#x", "global:443", "GLOBAL", "-global", "global-", strings.Repeat("a", 64), " global", "global\n"} {
		_, err := buildVertexPublisherModelsURL(location, "")
		require.Error(t, err, location)
	}
	_, err := buildVertexPublisherModelsURL("global", strings.Repeat("t", vertexPublisherPageTokenMaxBytes+1))
	require.Error(t, err)
}

func TestVertexPublisherFiltersModelPathsAndDeduplicates(t *testing.T) {
	names := []string{
		"publishers/google/models/gemini-z", "publishers/google/models/gemini-a",
		"publishers/google/models/gemini-a", "publishers/google/models/gemma-3",
		"publishers/google/models/veo-3", "publishers/other/models/gemini-a",
		"models/gemini-a", "publishers/google/models/gemini-",
		"publishers/google/models/gemini-../x", "publishers/google/models/gemini-?key=x",
		"publishers/google/models/gemini-x%2f..", "publishers/google/models/gemini-x#frag",
		"publishers/google/models/gemini-" + strings.Repeat("a", 129),
	}
	entries := make([]map[string]string, 0, len(names))
	for _, name := range names {
		entries = append(entries, map[string]string{"name": name})
	}
	body, err := json.Marshal(map[string]any{"publisherModels": entries})
	require.NoError(t, err)
	got, err := listVertexGeminiPublisherModels("global", func(int, string) ([]byte, error) { return body, nil })
	require.NoError(t, err)
	require.Equal(t, []string{"gemini-a", "gemini-z"}, got)
}

func TestVertexPublisherRejectsMalformedEmptyAndIncompleteCatalog(t *testing.T) {
	for _, body := range []string{"", "{", "null", "[]", "{}", `{"publisherModels":null}`, `{"publisherModels":{}}`, `{"publisherModels":[]}`, `{"publisherModels":[{"name":"publishers/google/models/veo-3"}]}`, `{"publisherModels":[],"nextPageToken":12}`} {
		t.Run(fmt.Sprintf("body-%x", []byte(body)), func(t *testing.T) {
			got, err := listVertexGeminiPublisherModels("global", func(int, string) ([]byte, error) { return []byte(body), nil })
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
	for _, kind := range []string{"repeated", "endless", "midpage-error", "oversized-token"} {
		t.Run(kind, func(t *testing.T) {
			calls := 0
			got, err := listVertexGeminiPublisherModels("global", func(page int, _ string) ([]byte, error) {
				calls++
				token := "repeat"
				switch kind {
				case "endless":
					token = fmt.Sprint(page)
				case "midpage-error":
					if page > 0 {
						return nil, errors.New("fixture page failed")
					}
				case "oversized-token":
					token = strings.Repeat("x", vertexPublisherPageTokenMaxBytes+1)
				}
				return json.Marshal(map[string]any{
					"publisherModels": []map[string]string{{"name": "publishers/google/models/gemini-a"}},
					"nextPageToken":   token,
				})
			})
			require.Error(t, err)
			require.Nil(t, got)
			if kind == "endless" {
				require.Equal(t, vertexPublisherModelsMaxPages, calls)
			} else {
				require.LessOrEqual(t, calls, 2)
			}
		})
	}
}

func TestVertexPublisherAdminSyncSharesPaginationAndTypedErrors(t *testing.T) {
	for _, kind := range []string{"success", "invalid-location", "invalid-json-key", "api-key-only", "midpage-404", "read-error", "read-limit", "malformed"} {
		t.Run(kind, func(t *testing.T) {
			account := vertexModelsAccount()
			cache := &vertexModelsTokenCache{}
			calls := 0
			svc := &AccountTestService{cfg: &config.Config{}, geminiTokenProvider: NewGeminiTokenProvider(nil, cache, nil)}
			switch kind {
			case "invalid-location":
				account.Credentials["location"] = "global/evil"
			case "invalid-json-key":
				account.Credentials["service_account_json"] = "{"
			case "api-key-only":
				account.Credentials = map[string]any{"api_key": "fixture"}
			case "read-limit":
				svc.cfg.Gateway.ModelsListReadMaxBytes = 10
			}
			svc.httpUpstream = vertexModelsUpstream{do: func(r *http.Request, _ string, id int64, concurrency int) (*http.Response, error) {
				calls++
				require.Equal(t, account.ID, id)
				require.Equal(t, account.Concurrency, concurrency)
				require.Equal(t, "aiplatform.googleapis.com", r.URL.Host)
				require.Equal(t, "Bearer fixture-cached-token", r.Header.Get("Authorization"))
				resp := vertexModelsResponse(200, `{"publisherModels":[{"name":"publishers/google/models/gemini-a"}],"nextPageToken":"next"}`)
				if calls == 2 {
					resp = vertexModelsResponse(200, `{"publisherModels":[{"name":"publishers/google/models/gemini-b"},{"name":"publishers/google/models/gemini-a"}]}`)
				}
				switch kind {
				case "midpage-404":
					if calls > 1 {
						resp = vertexModelsResponse(404, `{"error":"fixture missing"}`)
					}
				case "read-error":
					resp.Body = io.NopCloser(vertexModelsReadFailure{})
				case "malformed":
					resp = vertexModelsResponse(200, `{"publisherModels":`)
				}
				return resp, nil
			}}
			models, err := svc.FetchUpstreamSupportedModels(context.Background(), account)
			if kind == "success" {
				require.NoError(t, err)
				require.Equal(t, []string{"gemini-a", "gemini-b"}, models)
				require.Equal(t, 2, calls)
				return
			}
			require.Nil(t, models)
			var syncErr *UpstreamModelSyncError
			require.ErrorAs(t, err, &syncErr)
			if strings.HasPrefix(kind, "invalid-") || kind == "api-key-only" {
				require.Equal(t, UpstreamModelSyncErrorConfiguration, syncErr.Kind)
				require.Zero(t, calls)
				require.Zero(t, cache.calls)
			} else {
				require.Equal(t, UpstreamModelSyncErrorUpstream, syncErr.Kind)
				if kind == "midpage-404" {
					require.Equal(t, 2, calls)
					require.Contains(t, err.Error(), "HTTP 404")
				}
			}
		})
	}
}

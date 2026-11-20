package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type vertexModelsTokenCache struct {
	GeminiTokenCache
	calls int
}

func (c *vertexModelsTokenCache) GetAccessToken(context.Context, string) (string, error) {
	c.calls++
	return "fixture-cached-token", nil
}

type vertexModelsUpstream struct {
	HTTPUpstream
	do func(*http.Request, string, int64, int) (*http.Response, error)
}

func (u vertexModelsUpstream) Do(r *http.Request, proxy string, id int64, concurrency int) (*http.Response, error) {
	return u.do(r, proxy, id, concurrency)
}

func (u vertexModelsUpstream) DoWithTLS(r *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.do(r, proxy, id, concurrency)
}

func vertexModelsAccount() *Account {
	return &Account{
		ID: 410, Platform: PlatformGemini, Type: AccountTypeServiceAccount,
		Status: StatusActive, Schedulable: true, Concurrency: 7,
		Credentials: map[string]any{
			"location":             "global",
			"service_account_json": `{"project_id":"fixture","client_email":"fixture@example.test","private_key":"cached-fixture-only"}`,
		},
	}
}

func vertexModelsResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"fixture-id"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func vertexModelsService(upstream HTTPUpstream) *GeminiMessagesCompatService {
	return NewGeminiMessagesCompatService(nil, nil, nil, nil,
		NewGeminiTokenProvider(nil, &vertexModelsTokenCache{}, nil), nil, upstream, nil, &config.Config{})
}

func TestVertexMultiRegionNativeAndCompatRouting(t *testing.T) {
	const geminiResp = `{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":1}}`
	for _, tc := range []struct {
		name, entry, mapped, region, pin, wantHost, wantRegion string
		factor                                                 float64
	}{
		{"native us", "native", "gemini-3.5-flash", "us", "", "aiplatform.us.rep.googleapis.com", "us", 1.1},
		{"messages eu", "messages", "gemini-3.5-flash", "eu", "", "aiplatform.eu.rep.googleapis.com", "eu", 1.1},
		{"chat eu", "chat", "gemini-3.5-flash", "eu", "", "aiplatform.eu.rep.googleapis.com", "eu", 1.1},
		{"native global fallback", "native", "gemini-4-carbon", "us", "", "aiplatform.googleapis.com", "global", 1},
		{"native explicit region", "native", "gemini-4-carbon", "global", "explicit", "aiplatform.eu.rep.googleapis.com", "eu", 1},
		{"native cached pin", "native", "gemini-3.5-flash", "us", "global", "aiplatform.googleapis.com", "global", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := vertexModelsAccount()
			account.Credentials["location"] = tc.region
			account.Credentials["model_mapping"] = map[string]any{"alias": tc.mapped}
			if tc.pin == "explicit" {
				account.Credentials["vertex_model_locations"] = map[string]any{tc.mapped: "eu"}
			}
			calls := 0
			svc := vertexModelsService(vertexModelsUpstream{do: func(r *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
				calls++
				require.Equal(t, tc.wantHost, r.URL.Host)
				require.Equal(t, "/v1/projects/fixture/locations/"+tc.wantRegion+"/publishers/google/models/"+tc.mapped+":generateContent", r.URL.Path)
				return vertexModelsResponse(200, geminiResp), nil
			}})
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			ctx := context.Background()
			claude := []byte(`{"model":"alias","max_tokens":20,"messages":[{"role":"user","content":"hello"}]}`)
			var result *ForwardResult
			var err error
			switch tc.entry {
			case "native":
				body := []byte(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`)
				if tc.pin == "global" {
					body = append([]byte(`{"cachedContent":"projects/client-project/locations/global/cachedContents/42",`), body[1:]...)
				}
				result, err = svc.ForwardNative(ctx, c, account, "alias", "generateContent", false, body)
			case "messages":
				result, err = svc.Forward(ctx, c, account, claude)
			case "chat":
				result, err = svc.forwardClaudeBodyAsChatCompletions(ctx, c, account, claude, "alias", false, false, time.Now(), nil)
			}
			require.NoError(t, err)
			require.Equal(t, 1, calls)
			require.Equal(t, tc.factor, result.UpstreamPriceMultiplier)
		})
	}
}

func TestForwardListModelsVertexPaginationAndTransport(t *testing.T) {
	account := vertexModelsAccount()
	account.Credentials["location"] = "us"
	proxyID := int64(9)
	account.ProxyID = &proxyID
	account.Proxy = &Proxy{Protocol: "http", Host: "proxy.example.test", Port: 8080}
	// Neither an account base URL nor a supplied token URI can select the publisher host.
	account.Credentials["base_url"] = "https://untrusted.example.test"
	calls := 0
	svc := vertexModelsService(vertexModelsUpstream{do: func(r *http.Request, proxy string, id int64, concurrency int) (*http.Response, error) {
		calls++
		require.Equal(t, "https", r.URL.Scheme)
		require.Equal(t, "aiplatform.googleapis.com", r.URL.Host)
		require.Equal(t, "/v1beta1/publishers/google/models", r.URL.Path)
		require.Equal(t, "300", r.URL.Query().Get("pageSize"))
		require.Equal(t, "Bearer fixture-cached-token", r.Header.Get("Authorization"))
		require.Empty(t, r.Header.Get("x-goog-api-key"))
		require.Equal(t, "http://proxy.example.test:8080", proxy)
		require.Equal(t, account.ID, id)
		require.Equal(t, account.Concurrency, concurrency)
		if calls == 1 {
			return vertexModelsResponse(200, `{"publisherModels":[{"name":"publishers/google/models/gemini-2.5-pro"}],"nextPageToken":"next?&host=evil#x"}`), nil
		}
		require.Equal(t, "next?&host=evil#x", r.URL.Query().Get("pageToken"))
		require.Empty(t, r.URL.Fragment)
		return vertexModelsResponse(200, `{"publisherModels":[{"name":"publishers/google/models/gemini-2.5-pro"},{"name":"publishers/google/models/gemini-2.5-flash"},{"name":"publishers/google/models/veo-3"}]}`), nil
	}})
	result, err := svc.ForwardListModels(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, 2, calls)
	require.Equal(t, http.StatusOK, result.StatusCode)
	var models gemini.ModelsListResponse
	require.NoError(t, json.Unmarshal(result.Body, &models))
	require.Equal(t, []gemini.Model{gemini.FallbackModel("gemini-2.5-flash"), gemini.FallbackModel("gemini-2.5-pro")}, models.Models)
}

func TestForwardGetModelVertexRegionAndMissing(t *testing.T) {
	for _, model := range []string{"gemini-2.5-pro", "gemini-2.5-flash"} {
		t.Run(model, func(t *testing.T) {
			account := vertexModelsAccount()
			account.Credentials["vertex_model_locations"] = map[string]any{model: "europe-west4"}
			svc := vertexModelsService(vertexModelsUpstream{do: func(r *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
				require.Equal(t, "europe-west4-aiplatform.googleapis.com", r.URL.Host)
				return vertexModelsResponse(200, `{"publisherModels":[{"name":"publishers/google/models/gemini-2.5-pro"}]}`), nil
			}})
			result, err := svc.ForwardGetModel(context.Background(), account, model)
			require.NoError(t, err)
			if model == "gemini-2.5-pro" {
				require.Equal(t, 200, result.StatusCode)
				require.JSONEq(t, `{"name":"models/gemini-2.5-pro","supportedGenerationMethods":["generateContent","streamGenerateContent"]}`, string(result.Body))
			} else {
				require.Equal(t, 404, result.StatusCode)
				require.JSONEq(t, `{"error":{"code":404,"status":"NOT_FOUND","message":"models/gemini-2.5-flash is not found"}}`, string(result.Body))
			}
		})
	}
}

func TestGeminiModelsVertexRejectsConfigurationBeforeTokenOrHTTP(t *testing.T) {
	for _, kind := range []string{"location", "credentials", "api-key-only", "model", "platform"} {
		t.Run(kind, func(t *testing.T) {
			account := vertexModelsAccount()
			tokenCache := &vertexModelsTokenCache{}
			svc := vertexModelsService(vertexModelsUpstream{do: func(*http.Request, string, int64, int) (*http.Response, error) {
				t.Fatal("invalid configuration reached HTTP")
				return nil, nil
			}})
			svc.tokenProvider = NewGeminiTokenProvider(nil, tokenCache, nil)
			model := "gemini-2.5-pro"
			switch kind {
			case "location":
				account.Credentials["vertex_model_locations"] = map[string]any{model: "global/../../evil"}
			case "credentials":
				account.Credentials["service_account_json"] = "{"
			case "api-key-only":
				account.Credentials = map[string]any{"api_key": "fixture-key", "location": "global"}
			case "model":
				model = "gemini-pro%2f.."
			case "platform":
				account.Platform = PlatformAnthropic
			}
			result, err := svc.ForwardGetModel(context.Background(), account, model)
			require.Error(t, err)
			require.Nil(t, result)
			require.Zero(t, tokenCache.calls)
		})
	}
}

func TestGeminiModelsVertexNon2xxDiscardsPreviousPages(t *testing.T) {
	for _, status := range []int{401, 403, 404, 405, 429, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			svc := vertexModelsService(vertexModelsUpstream{do: func(*http.Request, string, int64, int) (*http.Response, error) {
				calls++
				if calls == 1 {
					return vertexModelsResponse(200, `{"publisherModels":[{"name":"publishers/google/models/gemini-2.5-pro"}],"nextPageToken":"next"}`), nil
				}
				resp := vertexModelsResponse(status, `{"error":{"message":"ACCESS_TOKEN_SCOPE_INSUFFICIENT"}}`)
				resp.Header["Www-Authenticate"] = []string{`Bearer error="insufficient_scope"`, `Bearer realm="fixture"`}
				resp.Header.Set("Set-Cookie", "private=fixture")
				return resp, nil
			}})
			result, err := svc.ForwardListModels(context.Background(), vertexModelsAccount())
			require.NoError(t, err)
			require.Equal(t, status, result.StatusCode)
			require.Equal(t, `{"error":{"message":"ACCESS_TOKEN_SCOPE_INSUFFICIENT"}}`, string(result.Body))
			require.Len(t, result.Headers.Values("Www-Authenticate"), 2)
			require.Equal(t, "fixture-id", result.Headers.Get("X-Request-Id"))
			require.Empty(t, result.Headers.Get("Set-Cookie"))
			require.Equal(t, 2, calls)
		})
	}
}

type vertexModelsReadFailure struct{}

func (vertexModelsReadFailure) Read([]byte) (int, error) {
	return 0, errors.New("fixture read failure")
}

type vertexModelsTrackedBody struct {
	io.Reader
	closed bool
}

func (b *vertexModelsTrackedBody) Close() error { b.closed = true; return nil }

func TestGeminiModelsReadLimitAndReadErrors(t *testing.T) {
	for _, vertex := range []bool{false, true} {
		for _, kind := range []string{"exact-limit", "over-limit", "read-error", "large-config"} {
			t.Run(kind+map[bool]string{true: "/vertex", false: "/aistudio"}[vertex], func(t *testing.T) {
				account := vertexModelsAccount()
				if !vertex {
					account.Type = AccountTypeAPIKey
					account.Credentials = map[string]any{"api_key": "fixture-key"}
				}
				payload := `{"publisherModels":[{"name":"publishers/google/models/gemini-2.5-pro"}]}`
				limit := int64(len(payload))
				var reader io.Reader = strings.NewReader(payload)
				switch kind {
				case "over-limit":
					limit--
				case "read-error":
					reader = io.MultiReader(strings.NewReader(payload), vertexModelsReadFailure{})
					limit += 10
				case "large-config":
					payload += strings.Repeat(" ", 8<<20)
					reader = strings.NewReader(payload)
					limit = int64(len(payload))
				}
				body := &vertexModelsTrackedBody{Reader: reader}
				svc := vertexModelsService(vertexModelsUpstream{do: func(*http.Request, string, int64, int) (*http.Response, error) {
					resp := vertexModelsResponse(200, "")
					resp.Body = body
					return resp, nil
				}})
				svc.cfg.Gateway.ModelsListReadMaxBytes = limit
				result, err := svc.ForwardListModels(context.Background(), account)
				require.True(t, body.closed)
				if kind == "read-error" || kind == "over-limit" {
					require.Error(t, err)
					require.Nil(t, result)
				} else {
					require.NoError(t, err)
					require.Equal(t, 200, result.StatusCode)
					if !vertex {
						require.Equal(t, payload, string(result.Body))
					}
				}
			})
		}
	}
}

func TestGeminiModelsAIStudioMetadataAndPathControl(t *testing.T) {
	account := &Account{Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture-key"}}
	calls := 0
	svc := vertexModelsService(vertexModelsUpstream{do: func(r *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		calls++
		require.Equal(t, "generativelanguage.googleapis.com", r.URL.Host)
		require.Equal(t, "fixture-key", r.Header.Get("x-goog-api-key"))
		return vertexModelsResponse(200, `{"models":[{"name":"models/gemini-native","inputTokenLimit":123}],"nextPageToken":"next"}`), nil
	}})
	svc.cfg.Security.URLAllowlist.Enabled = false
	result, err := svc.ForwardAIStudioGET(context.Background(), account, "/v1beta/models")
	require.NoError(t, err)
	require.Contains(t, string(result.Body), `"inputTokenLimit":123`)
	require.Contains(t, string(result.Body), `"nextPageToken":"next"`)
	for _, model := range []string{"../x", "gemini-pro?key=x", "models/gemini-pro", "gemini-pro%2f..", "gemini-pro#x", ""} {
		_, err := svc.ForwardGetModel(context.Background(), account, model)
		require.Error(t, err)
	}
	for _, path := range []string{"/v1beta/models/../x", "//evil.test/x", "/v1beta/models?x=1"} {
		_, err := svc.ForwardAIStudioGET(context.Background(), account, path)
		require.Error(t, err)
	}
	require.Equal(t, 1, calls)
}

type vertexModelsAccountRepo struct {
	AccountRepository
	accounts []Account
	byID     map[int64]*Account
}

func (r vertexModelsAccountRepo) ListSchedulableByGroupIDAndPlatforms(_ context.Context, _ int64, platforms []string) ([]Account, error) {
	var result []Account
	for _, a := range r.accounts {
		for _, platform := range platforms {
			if a.Platform == platform {
				result = append(result, a)
			}
		}
	}
	return result, nil
}

func (r vertexModelsAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	if a := r.byID[id]; a != nil {
		return a, nil
	}
	return nil, errors.New("fixture account missing")
}

func TestGeminiModelsAccountAIStudioPriorityAndTies(t *testing.T) {
	for _, kind := range []string{"apikey", "oauth", "code-assist", "priority", "last-used", "never-used", "vertex-only"} {
		t.Run(kind, func(t *testing.T) {
			vertex := *vertexModelsAccount()
			vertex.Priority = -100
			other := vertex
			other.ID++
			other.Type = AccountTypeOAuth
			other.Priority = 100
			other.Credentials = map[string]any{"access_token": "fixture"}
			want := other.ID
			switch kind {
			case "apikey":
				other.Type = AccountTypeAPIKey
				other.Credentials = map[string]any{"api_key": "fixture"}
			case "code-assist":
				other.Credentials["project_id"] = "fixture"
			case "priority":
				other = vertex
				other.ID++
				other.Priority = -101
			case "last-used", "never-used":
				other = vertex
				other.ID++
				now, earlier := time.Now(), time.Now().Add(-time.Hour)
				vertex.LastUsedAt = &now
				if kind == "last-used" {
					other.LastUsedAt = &earlier
				}
			case "vertex-only":
				want = vertex.ID
			}
			accounts := []Account{vertex, other}
			if kind == "vertex-only" {
				accounts = accounts[:1]
			}
			svc := &GeminiMessagesCompatService{accountRepo: vertexModelsAccountRepo{accounts: accounts}}
			groupID := int64(1)
			got, err := svc.SelectAccountForModelsEndpoints(context.Background(), &groupID)
			require.NoError(t, err)
			require.Equal(t, want, got.ID)
			old, err := svc.SelectAccountForAIStudioEndpoints(context.Background(), &groupID)
			require.NoError(t, err)
			require.Equal(t, want, old.ID)
		})
	}
}

func TestGeminiModelsPinnedAccountCannotFallback(t *testing.T) {
	groupID := int64(19)
	pinned := vertexModelsAccount()
	pinned.GroupIDs = []int64{groupID}
	other := *pinned
	other.ID++
	other.Type = AccountTypeAPIKey
	other.Credentials = map[string]any{"api_key": "fixture"}
	svc := &GeminiMessagesCompatService{accountRepo: vertexModelsAccountRepo{
		accounts: []Account{other}, byID: map[int64]*Account{pinned.ID: pinned},
	}}
	ctx := context.WithValue(context.Background(), ctxkey.Group, &Group{ID: groupID, Platform: PlatformGemini, Status: StatusActive, Hydrated: true})
	ctx = context.WithValue(ctx, ctxkey.APIKeyPinnedAccountID, pinned.ID)
	got, err := svc.SelectAccountForModelsEndpoints(ctx, &groupID)
	require.NoError(t, err)
	require.Same(t, pinned, got)
	pinned.Schedulable = false
	got, err = svc.SelectAccountForModelsEndpoints(ctx, &groupID)
	require.ErrorIs(t, err, ErrCafeAccountUnavailable)
	require.Nil(t, got)
}

type vertexModelsSnapshotCache struct {
	SchedulerCache
	lite     *Account
	hydrated *Account
}

func (c vertexModelsSnapshotCache) GetSnapshot(context.Context, SchedulerBucket) ([]*Account, bool, error) {
	return []*Account{c.lite}, true, nil
}

func (c vertexModelsSnapshotCache) GetAccount(context.Context, int64) (*Account, error) {
	return c.hydrated, nil
}

func TestGeminiModelsSelectionHydratesCredentialsAndProxy(t *testing.T) {
	full := vertexModelsAccount()
	proxyID := int64(1)
	full.ProxyID, full.Proxy = &proxyID, &Proxy{Host: "proxy.example.test", Protocol: "http", Port: 8080}
	lite := *full
	lite.Credentials, lite.Proxy = nil, nil
	snapshot := NewSchedulerSnapshotService(vertexModelsSnapshotCache{lite: &lite, hydrated: full}, nil, nil, nil, nil)
	svc := &GeminiMessagesCompatService{schedulerSnapshot: snapshot}
	groupID := int64(19)
	got, err := svc.SelectAccountForModelsEndpoints(context.Background(), &groupID)
	require.NoError(t, err)
	require.Same(t, full, got)
	require.NotEmpty(t, got.GetCredential("service_account_json"))
	require.Same(t, full.Proxy, got.Proxy)
}

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type vertexHandlerTokenCache struct{ service.GeminiTokenCache }

func (vertexHandlerTokenCache) GetAccessToken(context.Context, string) (string, error) {
	return "fixture-cached-token", nil
}

type vertexHandlerUpstream struct {
	service.HTTPUpstream
	do func(*http.Request) (*http.Response, error)
}

func (u vertexHandlerUpstream) Do(r *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return u.do(r)
}

type vertexHandlerRepo struct {
	service.AccountRepository
	accounts []service.Account
}

func (r vertexHandlerRepo) ListSchedulableByGroupIDAndPlatforms(_ context.Context, _ int64, platforms []string) ([]service.Account, error) {
	var result []service.Account
	for _, a := range r.accounts {
		for _, platform := range platforms {
			if a.Platform == platform {
				result = append(result, a)
			}
		}
	}
	return result, nil
}

func (r vertexHandlerRepo) GetByID(_ context.Context, id int64) (*service.Account, error) {
	for i := range r.accounts {
		if r.accounts[i].ID == id {
			return &r.accounts[i], nil
		}
	}
	return nil, errors.New("fixture account missing")
}

func vertexHandlerAccount() service.Account {
	return service.Account{
		ID: 1, Platform: service.PlatformGemini, Type: service.AccountTypeServiceAccount,
		Status: service.StatusActive, Schedulable: true, GroupIDs: []int64{77},
		Credentials: map[string]any{
			"location":             "global",
			"service_account_json": `{"project_id":"fixture","client_email":"fixture@example.test","private_key":"cached-fixture-only"}`,
		},
	}
}

func vertexHandlerContext(t *testing.T, detail bool) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1beta/models", nil)
	group := &service.Group{ID: 77, Platform: service.PlatformGemini, Status: service.StatusActive, Hydrated: true}
	c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{GroupID: &group.ID, Group: group})
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.Group, group))
	if detail {
		c.Params = gin.Params{{Key: "model", Value: "gemini-2.5-pro"}}
	}
	return c, rec
}

func runVertexModelsHandler(h *GatewayHandler, c *gin.Context, detail bool) {
	if detail {
		h.GeminiV1BetaGetModel(c)
	} else {
		h.GeminiV1BetaListModels(c)
	}
}

// These invoke the real handlers and concrete service. Scope errors and known
// model 404s must remain errors for Vertex, while AI Studio keeps its fallback.
func TestGeminiVertexModelsHandlerStatusAndAIStudioFallback(t *testing.T) {
	for _, vertex := range []bool{true, false} {
		for _, detail := range []bool{false, true} {
			for _, status := range []int{200, 401, 403, 404, 429, 500} {
				name := map[bool]string{true: "vertex", false: "aistudio"}[vertex] + "/" +
					map[bool]string{true: "detail/", false: "list/"}[detail] + http.StatusText(status)
				t.Run(name, func(t *testing.T) {
					account := vertexHandlerAccount()
					if !vertex {
						account.Type = service.AccountTypeAPIKey
						account.Credentials = map[string]any{"api_key": "fixture"}
					}
					body := `{"error":{"message":"fixture error"}}`
					if status == 200 {
						if vertex {
							body = `{"publisherModels":[{"name":"publishers/google/models/gemini-2.5-pro"}]}`
						} else {
							body = `{"models":[{"name":"models/gemini-native","inputTokenLimit":123}],"nextPageToken":"next"}`
						}
					}
					calls := 0
					upstream := vertexHandlerUpstream{do: func(r *http.Request) (*http.Response, error) {
						calls++
						if vertex {
							require.Equal(t, "aiplatform.googleapis.com", r.URL.Host)
							require.Equal(t, "/v1beta1/publishers/google/models", r.URL.Path)
							require.Equal(t, "Bearer fixture-cached-token", r.Header.Get("Authorization"))
						} else {
							require.Equal(t, "generativelanguage.googleapis.com", r.URL.Host)
						}
						headers := http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"fixture-id"}, "Set-Cookie": {"private=fixture"}}
						if status == 401 || status == 403 {
							headers.Set("Www-Authenticate", `Bearer error="insufficient_scope"`)
						}
						return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}, nil
					}}
					svc := service.NewGeminiMessagesCompatService(vertexHandlerRepo{accounts: []service.Account{account}},
						nil, nil, nil, service.NewGeminiTokenProvider(nil, vertexHandlerTokenCache{}, nil),
						nil, upstream, nil, &config.Config{})
					h := &GatewayHandler{geminiCompatService: svc}
					c, rec := vertexHandlerContext(t, detail)
					runVertexModelsHandler(h, c, detail)
					require.Equal(t, 1, calls)
					fallback := !vertex && (status == 401 || status == 403 || (status == 404 && detail))
					wantStatus := status
					if fallback {
						wantStatus = 200
					}
					require.Equal(t, wantStatus, rec.Code, rec.Body.String())
					if fallback {
						require.Contains(t, rec.Body.String(), "models/gemini-2.5-pro")
					} else if status != 200 || !vertex {
						require.Equal(t, body, rec.Body.String())
						require.Equal(t, "fixture-id", rec.Header().Get("X-Request-Id"))
						require.Empty(t, rec.Header().Get("Set-Cookie"))
						if status == 401 || status == 403 {
							require.Equal(t, `Bearer error="insufficient_scope"`, rec.Header().Get("Www-Authenticate"))
						}
					} else {
						require.Contains(t, rec.Body.String(), "models/gemini-2.5-pro")
					}
				})
			}
		}
	}
}

func TestGeminiVertexModelHandlerMissingKnownModelStays404(t *testing.T) {
	upstream := vertexHandlerUpstream{do: func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(
			`{"publisherModels":[{"name":"publishers/google/models/gemini-other"}]}`))}, nil
	}}
	svc := service.NewGeminiMessagesCompatService(vertexHandlerRepo{accounts: []service.Account{vertexHandlerAccount()}},
		nil, nil, nil, service.NewGeminiTokenProvider(nil, vertexHandlerTokenCache{}, nil), nil, upstream, nil, &config.Config{})
	h := &GatewayHandler{geminiCompatService: svc}
	c, rec := vertexHandlerContext(t, true)
	runVertexModelsHandler(h, c, true)
	require.Equal(t, 404, rec.Code)
	require.JSONEq(t, `{"error":{"code":404,"status":"NOT_FOUND","message":"models/gemini-2.5-pro is not found"}}`, rec.Body.String())
}

func TestGeminiVertexModelsHandlerMixedPreferenceAndPinnedProtection(t *testing.T) {
	for _, kind := range []string{"mixed", "pinned-vertex", "pinned-unavailable-list", "pinned-unavailable-detail"} {
		t.Run(kind, func(t *testing.T) {
			vertex := vertexHandlerAccount()
			native := vertex
			native.ID, native.Type = 2, service.AccountTypeAPIKey
			native.Credentials = map[string]any{"api_key": "fixture"}
			ag := vertex
			ag.ID, ag.Platform = 3, service.PlatformAntigravity
			ag.Extra = map[string]any{"mixed_scheduling": true}
			ag.Credentials = map[string]any{"model_mapping": map[string]any{"gemini-ag": "upstream"}}
			if strings.Contains(kind, "unavailable") {
				vertex.Schedulable = false
			}
			calls := 0
			upstream := vertexHandlerUpstream{do: func(r *http.Request) (*http.Response, error) {
				calls++
				body := `{"models":[{"name":"models/gemini-native"}]}`
				if kind == "pinned-vertex" {
					require.Equal(t, "aiplatform.googleapis.com", r.URL.Host)
					body = `{"publisherModels":[{"name":"publishers/google/models/gemini-2.5-pro"}]}`
				} else {
					require.Equal(t, "generativelanguage.googleapis.com", r.URL.Host)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			}}
			svc := service.NewGeminiMessagesCompatService(vertexHandlerRepo{accounts: []service.Account{vertex, native, ag}},
				nil, nil, nil, service.NewGeminiTokenProvider(nil, vertexHandlerTokenCache{}, nil), nil, upstream, nil, &config.Config{})
			h := &GatewayHandler{geminiCompatService: svc}
			detail := kind == "pinned-unavailable-detail"
			c, rec := vertexHandlerContext(t, detail)
			if kind != "mixed" {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.APIKeyPinnedAccountID, vertex.ID))
			}
			runVertexModelsHandler(h, c, detail)
			if strings.Contains(kind, "unavailable") {
				require.Equal(t, http.StatusServiceUnavailable, rec.Code, rec.Body.String())
				require.Zero(t, calls)
			} else {
				require.Equal(t, 200, rec.Code, rec.Body.String())
				require.Equal(t, 1, calls)
				if kind == "mixed" {
					require.Contains(t, rec.Body.String(), "gemini-native")
					require.Contains(t, rec.Body.String(), "gemini-ag")
				}
			}
		})
	}
}

func TestGeminiVertexModelsHandlerCustomAndForcedControls(t *testing.T) {
	for _, forced := range []bool{false, true} {
		t.Run(map[bool]string{false: "custom", true: "forced-antigravity"}[forced], func(t *testing.T) {
			c, rec := vertexHandlerContext(t, false)
			key, _ := middleware.GetAPIKeyFromContext(c)
			key.Group.ModelsListConfig = service.GroupModelsListConfig{Enabled: true, Models: []string{"gemini-custom"}}
			if forced {
				c.Set(string(middleware.ContextKeyForcePlatform), service.PlatformAntigravity)
			}
			(&GatewayHandler{}).GeminiV1BetaListModels(c)
			require.Equal(t, 200, rec.Code)
			var result gemini.ModelsListResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
			require.NotEmpty(t, result.Models)
			if !forced {
				require.Equal(t, []gemini.Model{gemini.FallbackModel("gemini-custom")}, result.Models)
			} else {
				require.NotContains(t, rec.Body.String(), "gemini-custom")
			}
		})
	}
}

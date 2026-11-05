package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type keyFast20261105HandlerRepo struct {
	service.APIKeyRepository
	key    *service.APIKey
	writes int
}

func (r *keyFast20261105HandlerRepo) Create(_ context.Context, key *service.APIKey) error {
	key.ID = 7
	clone := *key
	r.key = &clone
	r.writes++
	return nil
}
func (r *keyFast20261105HandlerRepo) GetByID(context.Context, int64) (*service.APIKey, error) {
	clone := *r.key
	return &clone, nil
}
func (r *keyFast20261105HandlerRepo) Update(_ context.Context, key *service.APIKey, _ service.APIKeyUpdateFields) error {
	clone := *key
	r.key = &clone
	r.writes++
	return nil
}
func (r *keyFast20261105HandlerRepo) ListByUserID(context.Context, int64, pagination.PaginationParams, service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error) {
	return []service.APIKey{*r.key}, &pagination.PaginationResult{Total: 1}, nil
}

type keyFast20261105HandlerUserRepo struct{ service.UserRepository }

func (*keyFast20261105HandlerUserRepo) GetByID(_ context.Context, id int64) (*service.User, error) {
	return &service.User{ID: id, Status: service.StatusActive}, nil
}

func keyFast20261105Router(h *APIKeyHandler, userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
	})
	router.POST("/keys", h.Create)
	router.PUT("/keys/:id", h.Update)
	router.GET("/keys/:id", h.GetByID)
	router.GET("/keys", h.List)
	return router
}

func keyFast20261105Request(router http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestKeyFast20261105HandlerRejectsInvalidJSONPolicy(t *testing.T) {
	// A nil service proves invalid values stop at the HTTP validation boundary.
	router := keyFast20261105Router(NewAPIKeyHandler(nil), 42)
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, raw := range []string{`null`, `1`, `true`, `[]`, `{}`, `""`, `"unknown"`, `"Force"`, `" filter "`} {
			t.Run(method+"/"+raw, func(t *testing.T) {
				target := "/keys"
				if method == http.MethodPut {
					target += "/7"
				}
				rec := keyFast20261105Request(router, method, target, `{"name":"test","openai_fast_policy":`+raw+`}`)
				require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), "openai_fast_policy")
			})
		}
	}
}

func TestKeyFast20261105HandlerCreateUpdateGetList(t *testing.T) {
	for _, policy := range []string{"", "passthrough", "force", "filter"} {
		t.Run(policy, func(t *testing.T) {
			repo := &keyFast20261105HandlerRepo{}
			svc := service.NewAPIKeyService(repo, &keyFast20261105HandlerUserRepo{}, nil, nil, nil, nil, &config.Config{})
			router := keyFast20261105Router(NewAPIKeyHandler(svc), 42)
			body := `{"name":"test"}`
			if policy != "" {
				body = `{"name":"test","openai_fast_policy":"` + policy + `"}`
			}
			rec := keyFast20261105Request(router, http.MethodPost, "/keys", body)
			require.Equal(t, 200, rec.Code, rec.Body.String())
			want := service.NormalizeAPIKeyOpenAIFastPolicy(policy)
			require.Equal(t, want, gjson.Get(rec.Body.String(), "data.openai_fast_policy").String())
			rec = keyFast20261105Request(router, http.MethodPut, "/keys/7", `{"name":"renamed"}`)
			require.Equal(t, 200, rec.Code, rec.Body.String())
			require.Equal(t, want, gjson.Get(rec.Body.String(), "data.openai_fast_policy").String())
			for _, next := range []string{"force", "filter", "passthrough"} {
				rec = keyFast20261105Request(router, http.MethodPut, "/keys/7", `{"openai_fast_policy":"`+next+`"}`)
				require.Equal(t, 200, rec.Code, rec.Body.String())
				require.Equal(t, next, repo.key.OpenAIFastPolicy)
				require.Equal(t, next, gjson.Get(rec.Body.String(), "data.openai_fast_policy").String())
				get := keyFast20261105Request(router, http.MethodGet, "/keys/7", "")
				require.Equal(t, 200, get.Code, get.Body.String())
				require.Equal(t, next, gjson.Get(get.Body.String(), "data.openai_fast_policy").String())
				list := keyFast20261105Request(router, http.MethodGet, "/keys", "")
				require.Equal(t, 200, list.Code, list.Body.String())
				require.Equal(t, next, gjson.Get(list.Body.String(), "data.items.0.openai_fast_policy").String())
			}
			writes := repo.writes
			foreign := keyFast20261105Router(NewAPIKeyHandler(svc), 99)
			rec = keyFast20261105Request(foreign, http.MethodPut, "/keys/7", `{"openai_fast_policy":"force"}`)
			require.Equal(t, http.StatusForbidden, rec.Code)
			require.Equal(t, writes, repo.writes)
			rec = keyFast20261105Request(foreign, http.MethodGet, "/keys/7", "")
			require.Equal(t, http.StatusNotFound, rec.Code)
			for _, source := range []string{service.APIKeyManagedSourceCafeRoomSeat, service.APIKeyManagedSourceCafeRoomMembership} {
				repo.key.ManagedSourceType = source
				sourceID := int64(77)
				repo.key.ManagedSourceID = &sourceID
				rec = keyFast20261105Request(router, http.MethodPut, "/keys/7", `{"openai_fast_policy":"force"}`)
				require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
				require.Equal(t, writes, repo.writes)
			}
		})
	}
}

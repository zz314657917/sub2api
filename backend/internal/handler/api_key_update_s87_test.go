package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type s87HandlerAPIKeyRepo struct {
	service.APIKeyRepository
	key     *service.APIKey
	updated *service.APIKey
}

func (r *s87HandlerAPIKeyRepo) GetByID(context.Context, int64) (*service.APIKey, error) {
	clone := *r.key
	clone.IPWhitelist = append([]string(nil), r.key.IPWhitelist...)
	clone.IPBlacklist = append([]string(nil), r.key.IPBlacklist...)
	return &clone, nil
}

func (r *s87HandlerAPIKeyRepo) Update(_ context.Context, key *service.APIKey) error {
	clone := *key
	clone.IPWhitelist = append([]string(nil), key.IPWhitelist...)
	clone.IPBlacklist = append([]string(nil), key.IPBlacklist...)
	r.updated = &clone
	return nil
}

func (r *s87HandlerAPIKeyRepo) ListByUserID(context.Context, int64, pagination.PaginationParams, service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error) {
	return []service.APIKey{*r.key}, &pagination.PaginationResult{Total: 1}, nil
}

func TestS87APIKeyUpdateJSONPresence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oldWhitelist := []string{"10.0.0.0/8"}
	oldBlacklist := []string{"192.168.1.0/24"}

	for _, tt := range []struct {
		name          string
		body          string
		wantWhitelist []string
		wantBlacklist []string
	}{
		{name: "omitted preserves", body: `{}`, wantWhitelist: oldWhitelist, wantBlacklist: oldBlacklist},
		{name: "null preserves", body: `{"ip_whitelist":null,"ip_blacklist":null}`, wantWhitelist: oldWhitelist, wantBlacklist: oldBlacklist},
		{name: "empty clears whitelist", body: `{"ip_whitelist":[]}`, wantWhitelist: []string{}, wantBlacklist: oldBlacklist},
		{name: "empty clears blacklist", body: `{"ip_blacklist":[]}`, wantWhitelist: oldWhitelist, wantBlacklist: []string{}},
		{name: "non-empty replaces both", body: `{"ip_whitelist":["203.0.113.7"],"ip_blacklist":["198.51.100.0/24"]}`, wantWhitelist: []string{"203.0.113.7"}, wantBlacklist: []string{"198.51.100.0/24"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &s87HandlerAPIKeyRepo{key: &service.APIKey{
				ID: 7, UserID: 42, Key: "s87-key", Status: service.StatusActive,
				IPWhitelist: oldWhitelist, IPBlacklist: oldBlacklist,
			}}
			svc := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, &config.Config{})
			h := NewAPIKeyHandler(svc)
			router := gin.New()
			router.PUT("/api/v1/api-keys/:id", func(c *gin.Context) {
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
				h.Update(c)
			})
			request := httptest.NewRequest(http.MethodPut, "/api/v1/api-keys/7", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if repo.updated == nil {
				t.Fatal("repository did not receive update")
			}
			if !equalS87Strings(repo.updated.IPWhitelist, tt.wantWhitelist) || !equalS87Strings(repo.updated.IPBlacklist, tt.wantBlacklist) {
				t.Fatalf("updated IP lists = %#v / %#v, want %#v / %#v", repo.updated.IPWhitelist, repo.updated.IPBlacklist, tt.wantWhitelist, tt.wantBlacklist)
			}
		})
	}

	var req UpdateAPIKeyRequest
	if err := json.Unmarshal([]byte(`{"ip_whitelist":null}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.IPWhitelist != nil {
		t.Fatal("JSON null must preserve nil update semantics")
	}
}

func equalS87Strings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

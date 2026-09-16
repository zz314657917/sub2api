package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func firstResponseControllerFixture(t *testing.T, seconds int) (*gin.Context, *httptest.ResponseRecorder, *service.APIKey, *service.APIKeyService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	key := &service.APIKey{ID: 940}
	for i := int64(1); i <= 3; i++ {
		group := &service.Group{ID: 940 + i, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
		key.MultiGroupRouteGroups = append(key.MultiGroupRouteGroups, group)
		key.MultiGroupRoutes = append(key.MultiGroupRoutes, domain.APIKeyMultiGroupRoute{GroupID: group.ID, Priority: int(i), Enabled: true, FirstResponseTimeoutSeconds: seconds})
		if i == 1 {
			key.Group, key.GroupID = group, &group.ID
		}
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5","stream":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)
	return c, recorder, key, &service.APIKeyService{}
}

func TestKeyRouteFirstResponseExpiryGateBeforeCancellation(t *testing.T) {
	for _, total := range []bool{false, true} {
		name := "group"
		if total {
			name = "total"
		}
		t.Run(name, func(t *testing.T) {
			c, recorder, _, svc := firstResponseControllerFixture(t, 30)
			calls := 0
			withKeyRouteFirstResponse(c, svc, func(attempt *gin.Context) {
				calls++
				attempt.JSON(http.StatusOK, gin.H{"attempt": calls})
				if calls == 1 {
					// Pause the timer's conceptual execution after winning the
					// gate but before it cancels the context. A ready response must
					// not become an empty HTTP 200 in this scheduling window.
					writer := attempt.Writer.(*keyRouteFirstResponseWriter)
					if total {
						writer.gate.expire()
					} else {
						writer.Expire()
					}
				}
			})
			if total {
				if calls != 1 || recorder.Code != http.StatusGatewayTimeout || recorder.Body.Len() == 0 {
					t.Fatalf("total expiration: calls=%d status=%d body=%q", calls, recorder.Code, recorder.Body.String())
				}
			} else if calls != 2 || recorder.Code != http.StatusOK || recorder.Body.String() != `{"attempt":2}` {
				t.Fatalf("group expiration: calls=%d status=%d body=%q", calls, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestKeyRouteFirstResponseTotalBudgetStopsBeforeThirdGroup(t *testing.T) {
	oldSecond := keyRouteFirstResponseSecond
	keyRouteFirstResponseSecond = time.Millisecond
	defer func() { keyRouteFirstResponseSecond = oldSecond }()
	c, recorder, key, svc := firstResponseControllerFixture(t, 80)
	calls := 0
	started := time.Now()
	withKeyRouteFirstResponse(c, svc, func(attempt *gin.Context) {
		calls++
		<-attempt.Request.Context().Done()
	})
	if calls != 2 || recorder.Code != http.StatusGatewayTimeout || time.Since(started) > time.Second {
		t.Fatalf("total budget must clip B and skip C: calls=%d status=%d elapsed=%s", calls, recorder.Code, time.Since(started))
	}
	resolved := svc.ResolveForModelRequest(c.Request.Context(), key, "/v1/responses", "", "gpt-5", false)
	if resolved == nil || resolved.GroupID == nil || *resolved.GroupID != 943 {
		t.Fatalf("unattempted C must remain available, got %#v", resolved)
	}
}

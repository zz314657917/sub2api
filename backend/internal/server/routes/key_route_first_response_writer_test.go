package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func TestKeyRouteFirstResponseWriterStagesHeadersAndSplitSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	writer := newKeyRouteFirstResponseWriter(context.Writer, true, &keyRouteFirstResponseGate{})
	writer.Header().Set("X-Attempt", "A")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(": ping\n\ndata: {\"type\":\"response.created\"}\n\n"))
	writer.Flush()
	if recorder.Code != http.StatusOK || recorder.Header().Get("X-Attempt") != "" || recorder.Body.Len() != 0 {
		t.Fatalf("prelude leaked: status=%d header=%q body=%q", recorder.Code, recorder.Header().Get("X-Attempt"), recorder.Body.String())
	}
	_, _ = writer.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"de"))
	if writer.Committed() {
		t.Fatal("partial SSE event committed")
	}
	_, _ = writer.Write([]byte("lta\":\"ok\"}\n\n"))
	if !writer.Committed() {
		t.Fatal("valid split text delta did not commit")
	}
	if recorder.Header().Get("X-Attempt") != "A" || recorder.Body.Len() == 0 {
		t.Fatalf("committed result missing: header=%q body=%q", recorder.Header().Get("X-Attempt"), recorder.Body.String())
	}
}

func TestKeyRouteFirstResponseCancelsABeforeBStarts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oldSecond := keyRouteFirstResponseSecond
	keyRouteFirstResponseSecond = 15 * time.Millisecond
	defer func() { keyRouteFirstResponseSecond = oldSecond }()
	aID, bID := int64(31), int64(32)
	groupA := &service.Group{ID: aID, Status: service.StatusActive, Platform: service.PlatformOpenAI, Hydrated: true, ModelMatchPatterns: []string{"*"}}
	groupB := &service.Group{ID: bID, Status: service.StatusActive, Platform: service.PlatformOpenAI, Hydrated: true, ModelMatchPatterns: []string{"*"}}
	key := &service.APIKey{ID: 101, GroupID: &aID, Group: groupA, MultiGroupRouteGroups: []*service.Group{groupA, groupB}, MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: aID, Enabled: true, Priority: 1, FirstResponseTimeoutSeconds: 1}, {GroupID: bID, Enabled: true, Priority: 2, FirstResponseTimeoutSeconds: 1}}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-5","stream":false}`))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)
	var aStopped atomic.Bool
	withKeyRouteFirstResponse(context, &service.APIKeyService{}, func(attempt *gin.Context) {
		resolved, _ := middleware.GetAPIKeyFromContext(attempt)
		if resolved.GroupID != nil && *resolved.GroupID == aID {
			<-attempt.Request.Context().Done()
			aStopped.Store(true)
			return
		}
		if !aStopped.Load() {
			t.Error("B began before A returned after cancellation")
		}
		attempt.JSON(http.StatusOK, gin.H{"ok": true})
	})
	if recorder.Code != http.StatusOK || !aStopped.Load() || !bytes.Contains(recorder.Body.Bytes(), []byte(`"ok":true`)) {
		t.Fatalf("want A cancellation then B success, status=%d body=%q stopped=%v", recorder.Code, recorder.Body.String(), aStopped.Load())
	}
}

func TestKeyRouteFirstResponseLocalFourXXIgnoresStaleUpstreamFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	aID, bID := int64(41), int64(42)
	groupA := &service.Group{ID: aID, Status: service.StatusActive, Platform: service.PlatformOpenAI, Hydrated: true, ModelMatchPatterns: []string{"*"}}
	groupB := &service.Group{ID: bID, Status: service.StatusActive, Platform: service.PlatformOpenAI, Hydrated: true, ModelMatchPatterns: []string{"*"}}
	key := &service.APIKey{ID: 102, GroupID: &aID, Group: groupA, MultiGroupRouteGroups: []*service.Group{groupA, groupB}, MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: aID, Enabled: true, Priority: 1, FirstResponseTimeoutSeconds: 1}, {GroupID: bID, Enabled: true, Priority: 2, FirstResponseTimeoutSeconds: 1}}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5"}`))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)
	var calls atomic.Int32
	withKeyRouteFirstResponse(context, &service.APIKeyService{}, func(attempt *gin.Context) {
		calls.Add(1)
		attempt.Set(service.OpsUpstreamStatusCodeKey, http.StatusServiceUnavailable)
		attempt.JSON(http.StatusForbidden, gin.H{"error": "policy"})
	})
	if calls.Load() != 1 || recorder.Code != http.StatusForbidden {
		t.Fatalf("local 403 replayed through stale upstream status: calls=%d status=%d", calls.Load(), recorder.Code)
	}
}

func TestKeyRouteFirstResponseWriterTimeoutGateWins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	gate := &keyRouteFirstResponseGate{}
	writer := newKeyRouteFirstResponseWriter(context.Writer, true, gate)
	if !writer.Expire() {
		t.Fatal("expire did not win")
	}
	if _, err := writer.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"late\"}\n\n")); err == nil {
		t.Fatal("late business response committed after timeout")
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("late body leaked: %q", recorder.Body.String())
	}
}

func TestKeyRouteFirstResponseWriterUsageCommitsBeforeLifecycleTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	writer := newKeyRouteFirstResponseWriter(context.Writer, true, &keyRouteFirstResponseGate{})
	_, _ = writer.Write([]byte("data: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":12}}}\n\n"))
	if !writer.Committed() {
		t.Fatal("usage-bearing lifecycle event must commit and disable replay")
	}
}

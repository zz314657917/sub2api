package routes

import (
	"bytes"
	"context"
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

func TestQAKeyRouteFirstResponseDefaultFallbackTimeoutCoolsDefaultGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oldSecond := keyRouteFirstResponseSecond
	keyRouteFirstResponseSecond = 5 * time.Millisecond
	t.Cleanup(func() { keyRouteFirstResponseSecond = oldSecond })

	aID, defaultID := int64(831), int64(832)
	groupA := &service.Group{ID: aID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	defaultGroup := &service.Group{ID: defaultID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	key := &service.APIKey{
		ID: 831, GroupID: &defaultID, Group: defaultGroup,
		MultiGroupRouteGroups: []*service.Group{groupA, defaultGroup},
		MultiGroupRoutes:      []domain.APIKeyMultiGroupRoute{{GroupID: aID, Priority: 1, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1}},
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5","stream":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)
	apiKeyService := &service.APIKeyService{}

	withKeyRouteFirstResponse(c, apiKeyService, func(attempt *gin.Context) {
		resolved, _ := middleware.GetAPIKeyFromContext(attempt)
		if resolved.GroupID != nil && *resolved.GroupID == aID {
			attempt.Writer.WriteHeader(http.StatusBadGateway)
			_, _ = attempt.Writer.Write([]byte(`{"error":"upstream"}`))
			return
		}
		<-attempt.Request.Context().Done()
	})
	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("want exhausted fallback timeout, got %d body=%q", recorder.Code, recorder.Body.String())
	}
	next := apiKeyService.ResolveForModelRequest(context.Background(), key, "/v1/responses", "", "gpt-5", false)
	if next != nil && next.GroupID != nil && *next.GroupID == defaultID {
		t.Fatal("timed-out default fallback C remains immediately selectable; dispatcher did not cool it")
	}
}

func TestQAKeyRouteFirstResponseLocalProviderCancellationThenDefaultFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oldSecond := keyRouteFirstResponseSecond
	keyRouteFirstResponseSecond = 10 * time.Millisecond
	t.Cleanup(func() { keyRouteFirstResponseSecond = oldSecond })

	aCancelled := make(chan struct{})
	providerA := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(aCancelled)
	}))
	t.Cleanup(providerA.Close)

	aID, defaultID := int64(811), int64(812)
	groupA := &service.Group{ID: aID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	defaultGroup := &service.Group{ID: defaultID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	key := &service.APIKey{
		ID: 811, GroupID: &defaultID, Group: defaultGroup,
		MultiGroupRouteGroups: []*service.Group{groupA, defaultGroup},
		MultiGroupRoutes:      []domain.APIKeyMultiGroupRoute{{GroupID: aID, Priority: 1, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1}},
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-5","stream":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)

	var bStarted atomic.Bool
	withKeyRouteFirstResponse(c, &service.APIKeyService{}, func(attempt *gin.Context) {
		resolved, _ := middleware.GetAPIKeyFromContext(attempt)
		if resolved.GroupID != nil && *resolved.GroupID == aID {
			req, err := http.NewRequestWithContext(attempt.Request.Context(), http.MethodGet, providerA.URL, nil)
			if err != nil {
				t.Error(err)
				return
			}
			_, _ = http.DefaultClient.Do(req)
			return
		}
		select {
		case <-aCancelled:
			bStarted.Store(true)
			attempt.JSON(http.StatusOK, gin.H{"group": "default"})
		case <-time.After(time.Second):
			t.Error("default fallback began before the cancelled provider request returned")
		}
	})
	if !bStarted.Load() || recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte(`"group":"default"`)) {
		t.Fatalf("want cancelled A then successful default fallback, started=%v status=%d body=%q", bStarted.Load(), recorder.Code, recorder.Body.String())
	}
}

func TestQAKeyRouteFirstResponseCommittedStreamFailureCoolsRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	firstID, fallbackID := int64(801), int64(802)
	firstGroup := &service.Group{ID: firstID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	fallbackGroup := &service.Group{ID: fallbackID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	key := &service.APIKey{
		ID:                    801,
		GroupID:               &firstID,
		Group:                 firstGroup,
		MultiGroupRouteGroups: []*service.Group{firstGroup, fallbackGroup},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: firstID, Priority: 1, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1},
			{GroupID: fallbackID, Priority: 2, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1},
		},
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-5","stream":true}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)
	apiKeyService := &service.APIKeyService{}

	withKeyRouteFirstResponse(c, apiKeyService, func(attempt *gin.Context) {
		_, _ = attempt.Writer.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"ok\"}\n\n"))
		middleware.MarkAPIKeyRouteCooldown(attempt, http.StatusTooManyRequests)
	})

	if !bytes.Contains(recorder.Body.Bytes(), []byte("response.output_text.delta")) {
		t.Fatalf("committed stream was not delivered: %q", recorder.Body.String())
	}
	resolved := apiKeyService.ResolveForRequest(c.Request.Context(), key, "/v1/chat/completions", "")
	if resolved == nil || resolved.GroupID == nil || *resolved.GroupID != fallbackID {
		t.Fatalf("committed 429 stream must retain A cooldown and choose B next time, got %#v", resolved)
	}
}

func TestQAKeyRouteFirstResponseLocalFourXXDoesNotReplay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	aID, bID := int64(821), int64(822)
	groupA := &service.Group{ID: aID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	groupB := &service.Group{ID: bID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	key := &service.APIKey{
		ID: 821, GroupID: &aID, Group: groupA, MultiGroupRouteGroups: []*service.Group{groupA, groupB},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: aID, Priority: 1, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1},
			{GroupID: bID, Priority: 2, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1},
		},
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5","stream":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)

	var calls atomic.Int32
	withKeyRouteFirstResponse(c, &service.APIKeyService{}, func(attempt *gin.Context) {
		calls.Add(1)
		attempt.JSON(http.StatusForbidden, gin.H{"error": "local policy"})
	})
	if calls.Load() != 1 || recorder.Code != http.StatusForbidden {
		t.Fatalf("local 4xx must stay terminal, calls=%d status=%d body=%q", calls.Load(), recorder.Code, recorder.Body.String())
	}
}

func TestQAKeyRouteFirstResponseCopiesOnlyTerminalAttemptContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	aID, bID := int64(851), int64(852)
	groupA := &service.Group{ID: aID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	groupB := &service.Group{ID: bID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	key := &service.APIKey{
		ID: 851, GroupID: &aID, Group: groupA, MultiGroupRouteGroups: []*service.Group{groupA, groupB},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: aID, Priority: 1, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1},
			{GroupID: bID, Priority: 2, CooldownSeconds: 60, Enabled: true, FirstResponseTimeoutSeconds: 1},
		},
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5","stream":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), key)

	withKeyRouteFirstResponse(c, &service.APIKeyService{}, func(attempt *gin.Context) {
		resolved, _ := middleware.GetAPIKeyFromContext(attempt)
		if resolved.GroupID != nil && *resolved.GroupID == aID {
			attempt.Set("qa_failed_attempt_only", true)
			attempt.JSON(http.StatusBadGateway, gin.H{"error": "upstream"})
			return
		}
		attempt.Set("qa_terminal_attempt", "B")
		attempt.JSON(http.StatusOK, gin.H{"group": "B"})
	})
	if _, leaked := c.Get("qa_failed_attempt_only"); leaked {
		t.Fatal("failed A attempt context leaked into the terminal request context")
	}
	if winner, ok := c.Get("qa_terminal_attempt"); !ok || winner != "B" {
		t.Fatalf("terminal B context was not copied, got %#v", winner)
	}
}

func TestQAKeyRouteFirstResponseEligibilityExclusions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testCases := []struct {
		name   string
		path   string
		body   string
		forced bool
	}{
		{name: "responses subpath", path: "/v1/responses/compact", body: `{"model":"gpt-5"}`},
		{name: "stateful previous response", path: "/v1/responses", body: `{"model":"gpt-5","previous_response_id":"resp_1"}`},
		{name: "stateful conversation", path: "/v1/responses", body: `{"model":"gpt-5","conversation":{}}`},
		{name: "invalid stream type", path: "/v1/responses", body: `{"model":"gpt-5","stream":"true"}`},
		{name: "image tool", path: "/v1/responses", body: `{"model":"gpt-5","tools":[{"type":"image_generation"}]}`},
		{name: "forced platform", path: "/v1/responses", body: `{"model":"gpt-5"}`, forced: true},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, testCase.path, bytes.NewBufferString(testCase.body))
			c.Request.Header.Set("Content-Type", "application/json; charset=utf-8")
			if testCase.forced {
				c.Set(string(middleware.ContextKeyForcePlatform), service.PlatformAntigravity)
			}
			if _, ok := keyRouteFirstResponseEligible(c); ok {
				t.Fatal("excluded request was admitted to first-response replay")
			}
		})
	}
}

func TestQAKeyRouteFirstResponseDisabledAndPinnedStayOnLegacyDispatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(841)
	group := &service.Group{ID: groupID, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI, ModelMatchPatterns: []string{"*"}}
	for _, testCase := range []struct {
		name string
		key  *service.APIKey
	}{
		{
			name: "all route timeouts disabled",
			key:  &service.APIKey{ID: 841, GroupID: &groupID, Group: group, MultiGroupRouteGroups: []*service.Group{group}, MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: groupID, Enabled: true, FirstResponseTimeoutSeconds: 0}}},
		},
		{
			name: "pinned account",
			key:  &service.APIKey{ID: 842, GroupID: &groupID, Group: group, PinnedAccountID: 99, MultiGroupRouteGroups: []*service.Group{group}, MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: groupID, Enabled: true, FirstResponseTimeoutSeconds: 1}}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-5","stream":false}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set(string(middleware.ContextKeyAPIKeyFirstResponsePristine), testCase.key)
			var calls atomic.Int32
			withKeyRouteFirstResponse(c, &service.APIKeyService{}, func(attempt *gin.Context) {
				calls.Add(1)
				if attempt != c {
					t.Error("legacy request was replaced with an attempt context")
				}
				attempt.JSON(http.StatusOK, gin.H{"legacy": true})
			})
			if calls.Load() != 1 || recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte(`"legacy":true`)) {
				t.Fatalf("want one legacy dispatch, calls=%d status=%d body=%q", calls.Load(), recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestQAKeyRouteFirstResponseWriterOnlyCommitsReasoningOrToolNotPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, testCase := range []struct {
		name      string
		payload   string
		committed bool
	}{
		{name: "ping", payload: "data: {\"type\":\"ping\"}\n\n", committed: false},
		{name: "reasoning delta", payload: "data: {\"type\":\"response.reasoning.delta\",\"delta\":\"trace\"}\n\n", committed: true},
		{name: "tool delta", payload: "data: {\"type\":\"response.function_call_arguments.delta\",\"delta\":\"{}\"}\n\n", committed: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			writer := newKeyRouteFirstResponseWriter(c.Writer, true, &keyRouteFirstResponseGate{})
			if _, err := writer.Write([]byte(testCase.payload)); err != nil {
				t.Fatal(err)
			}
			if writer.Committed() != testCase.committed {
				t.Fatalf("committed=%v, want %v", writer.Committed(), testCase.committed)
			}
		})
	}
}

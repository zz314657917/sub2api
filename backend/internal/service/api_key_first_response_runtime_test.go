package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

func TestAPIKeyFirstResponseDetachedUpstreamCancellation(t *testing.T) {
	cancelled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(cancelled)
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
	upstreamCtx, release := detachUpstreamContext(ctx)
	defer release()
	req, err := http.NewRequestWithContext(upstreamCtx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { _, _ = http.DefaultClient.Do(req); close(done) }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("real upstream HTTP request was not cancelled")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("client did not return after cancellation")
	}
}

func TestAPIKeyFirstResponseAttemptNeverRepeatsVisitedGroup(t *testing.T) {
	aID, bID := int64(11), int64(12)
	groupA := &Group{ID: aID, Status: StatusActive, Platform: PlatformOpenAI, Hydrated: true, ModelMatchPatterns: []string{"*"}}
	groupB := &Group{ID: bID, Status: StatusActive, Platform: PlatformOpenAI, Hydrated: true, ModelMatchPatterns: []string{"*"}}
	apiKey := &APIKey{ID: 99, GroupID: &aID, Group: groupA, MultiGroupRouteGroups: []*Group{groupA, groupB}, MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: aID, Enabled: true, Priority: 1}, {GroupID: bID, Enabled: true, Priority: 2}}}
	service := &APIKeyService{}
	resolved := service.ResolveForFirstResponseAttempt(context.Background(), apiKey, "/v1/chat/completions", "", "gpt-5", map[int64]struct{}{aID: {}})
	if resolved == nil || resolved.GroupID == nil || *resolved.GroupID != bID {
		t.Fatalf("visited A selected again: %#v", resolved)
	}
}

func TestAPIKeyFirstResponseDetachedStreamCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
	defer cancel()
	detached, release := detachStreamUpstreamContext(ctx, true)
	defer release()
	cancel()
	select {
	case <-detached.Done():
	case <-time.After(time.Second):
		t.Fatal("stream detach removed marked cancellation")
	}
}

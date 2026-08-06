package oauth

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBuildAuthorizationURLUsesCaiEndpoint(t *testing.T) {
	authorizationURL := BuildAuthorizationURL("state", "challenge", ScopeOAuth)
	if !strings.HasPrefix(authorizationURL, "https://claude.com/cai/oauth/authorize?") {
		t.Fatalf("BuildAuthorizationURL() = %q, want claude.com/cai endpoint", authorizationURL)
	}
}

func TestSessionStore_Stop_Idempotent(t *testing.T) {
	store := NewSessionStore()

	store.Stop()
	store.Stop()

	select {
	case <-store.stopCh:
		// ok
	case <-time.After(time.Second):
		t.Fatal("stopCh 未关闭")
	}
}

func TestSessionStore_Stop_Concurrent(t *testing.T) {
	store := NewSessionStore()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Stop()
		}()
	}

	wg.Wait()

	select {
	case <-store.stopCh:
		// ok
	case <-time.After(time.Second):
		t.Fatal("stopCh 未关闭")
	}
}

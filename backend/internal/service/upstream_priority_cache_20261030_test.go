package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"
)

// Use the real localhost transport, including request cancellation and ETags.
type TestUpstreamPriorityManifestHTTPClient struct{}

func (TestUpstreamPriorityManifestHTTPClient) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func (c TestUpstreamPriorityManifestHTTPClient) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return c.Do(req, proxy, id, concurrency)
}

func (TestUpstreamPriorityManifestHTTPClient) awaitResult(t *testing.T, ch <-chan singleflight.Result) singleflight.Result {
	t.Helper()
	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("shared manifest refresh did not finish")
		return singleflight.Result{}
	}
}

func TestUpstreamPriorityCacheLateRefreshRechecksFresh(t *testing.T) {
	for _, initial := range []string{"miss", "stale"} {
		t.Run(initial, func(t *testing.T) {
			var calls atomic.Int32
			started, release := make(chan struct{}), make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					close(started)
					<-release
				}
				w.Header().Set("ETag", `"new"`)
				_, _ = w.Write([]byte(`{"models":[{"slug":"test-model"}]}`))
			}))
			defer server.Close()
			svc := &OpenAIGatewayService{httpUpstream: TestUpstreamPriorityManifestHTTPClient{}}
			req := codexModelsManifestRequest{url: server.URL, headers: make(http.Header), useAPIKeyUpstream: true}
			key := buildCodexModelsManifestCacheKey(req)
			if initial == "stale" {
				svc.codexModelsManifestCache.set(key, &CodexModelsManifest{Body: []byte(`{"models":[]}`), upstreamETag: `"old"`}, time.Now().Add(-time.Minute))
			}
			// Pause the late caller at the real get -> refresh call boundary.
			// A previous refresh must finish before it enters singleflight.
			observed, enter := make(chan struct{}), make(chan struct{})
			late := make(chan singleflight.Result, 1)
			go func() {
				_, state := svc.codexModelsManifestCache.get(key, time.Now())
				if state == codexModelsManifestCacheFresh {
					late <- singleflight.Result{Err: context.Canceled}
					close(observed)
					return
				}
				close(observed)
				<-enter
				late <- <-svc.refreshCachedAPIKeyCodexModelsManifest(key, req)
			}()
			<-observed
			first := svc.refreshCachedAPIKeyCodexModelsManifest(key, req)
			<-started
			close(release)
			result := (TestUpstreamPriorityManifestHTTPClient{}).awaitResult(t, first)
			require.NoError(t, result.Err)
			close(enter)
			lateResult := (TestUpstreamPriorityManifestHTTPClient{}).awaitResult(t, late)
			require.NoError(t, lateResult.Err)
			require.EqualValues(t, 1, calls.Load(), "late entry after an older miss/stale read must reuse the now-fresh manifest")
			require.Same(t, result.Val, lateResult.Val)
		})
	}
}

func TestUpstreamPriorityCacheFreshAndStale304(t *testing.T) {
	var calls atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		close(started)
		<-release
		if r.Header.Get("If-None-Match") != `"source"` {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()
	svc := &OpenAIGatewayService{httpUpstream: TestUpstreamPriorityManifestHTTPClient{}}
	req := codexModelsManifestRequest{url: server.URL, headers: make(http.Header), useAPIKeyUpstream: true}
	key := buildCodexModelsManifestCacheKey(req)
	cached := &CodexModelsManifest{Body: []byte(`{"models":[]}`), ETag: `"client"`, upstreamETag: `"source"`}
	svc.codexModelsManifestCache.set(key, cached, time.Now())
	fresh, err := svc.fetchCachedAPIKeyCodexModelsManifest(context.Background(), req, `"client"`)
	require.NoError(t, err)
	require.True(t, fresh.NotModified)
	require.Zero(t, calls.Load())

	svc.codexModelsManifestCache.set(key, cached, time.Now().Add(-time.Minute))
	// This call must return stale without waiting for the blocked HTTP response.
	returned := make(chan *CodexModelsManifest, 1)
	go func() {
		manifest, _ := svc.fetchCachedAPIKeyCodexModelsManifest(context.Background(), req, "")
		returned <- manifest
	}()
	select {
	case manifest := <-returned:
		require.Same(t, cached, manifest)
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("stale response waited for the upstream")
	}
	<-started
	joined := svc.refreshCachedAPIKeyCodexModelsManifest(key, req)
	close(release)
	result := (TestUpstreamPriorityManifestHTTPClient{}).awaitResult(t, joined)
	require.NoError(t, result.Err)
	require.Same(t, cached, result.Val)
	manifest, state := svc.codexModelsManifestCache.get(key, time.Now())
	require.Equal(t, codexModelsManifestCacheFresh, state)
	require.Same(t, cached, manifest)
	require.EqualValues(t, 1, calls.Load())
}

func TestUpstreamPriorityCacheMissSharingSurvivesCancellation(t *testing.T) {
	var calls atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		close(started)
		<-release
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()
	svc := &OpenAIGatewayService{httpUpstream: TestUpstreamPriorityManifestHTTPClient{}}
	req := codexModelsManifestRequest{url: server.URL, headers: make(http.Header), useAPIKeyUpstream: true}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := svc.fetchCachedAPIKeyCodexModelsManifest(ctx, req, "")
		done <- err
	}()
	<-started
	key := buildCodexModelsManifestCacheKey(req)
	joined := svc.refreshCachedAPIKeyCodexModelsManifest(key, req)
	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("cancelled caller did not return")
	}
	close(release)
	result := (TestUpstreamPriorityManifestHTTPClient{}).awaitResult(t, joined)
	require.NoError(t, result.Err)
	require.True(t, result.Shared)
	require.EqualValues(t, 1, calls.Load())
	manifest, err := svc.fetchCachedAPIKeyCodexModelsManifest(context.Background(), req, "")
	require.NoError(t, err)
	require.NotNil(t, manifest)
	require.EqualValues(t, 1, calls.Load())
}

func TestUpstreamPriorityCacheUpstreamErrorsPreserved(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusTooManyRequests, http.StatusBadGateway} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
				_, _ = w.Write([]byte("upstream unavailable"))
			}))
			defer server.Close()
			svc := &OpenAIGatewayService{httpUpstream: TestUpstreamPriorityManifestHTTPClient{}}
			req := codexModelsManifestRequest{url: server.URL, headers: make(http.Header), useAPIKeyUpstream: true}
			key := buildCodexModelsManifestCacheKey(req)
			manifest, err := svc.fetchCachedAPIKeyCodexModelsManifest(context.Background(), req, "")
			require.Nil(t, manifest)
			var upstreamErr *codexModelsManifestUpstreamError
			require.ErrorAs(t, err, &upstreamErr)
			require.Equal(t, status, upstreamErr.statusCode)
			require.Equal(t, status != http.StatusUnauthorized, IsRetryableCodexModelsManifestError(err))
			_, state := svc.codexModelsManifestCache.get(key, time.Now())
			require.Equal(t, codexModelsManifestCacheMiss, state)

			stale := &CodexModelsManifest{Body: []byte(`{"models":[]}`)}
			svc.codexModelsManifestCache.set(key, stale, time.Now().Add(-time.Minute))
			result := (TestUpstreamPriorityManifestHTTPClient{}).awaitResult(t, svc.refreshCachedAPIKeyCodexModelsManifest(key, req))
			require.Error(t, result.Err)
			retained, state := svc.codexModelsManifestCache.get(key, time.Now())
			require.Equal(t, codexModelsManifestCacheStale, state)
			require.Same(t, stale, retained)
			require.EqualValues(t, 2, calls.Load())
		})
	}
}

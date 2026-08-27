package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type moderationProxyRepo struct {
	ProxyRepository
	proxy   *Proxy
	err     error
	getCall atomic.Int64
}

func (r *moderationProxyRepo) GetByID(_ context.Context, _ int64) (*Proxy, error) {
	r.getCall.Add(1)
	return r.proxy, r.err
}

func TestContentModerationValidateConfigRejectsEmptyProxyRecord(t *testing.T) {
	proxyID := int64(7)
	svc := &ContentModerationService{proxyRepo: &moderationProxyRepo{}}
	cfg := defaultContentModerationConfig()
	cfg.ProxyID = &proxyID

	err := svc.validateConfig(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "代理服务器不存在")
}

func TestContentModerationProxyURLResolutionCachesLookup(t *testing.T) {
	repo := &moderationProxyRepo{proxy: &Proxy{ID: 7, Name: "audit", Protocol: "http", Host: "127.0.0.1", Port: 8080, Status: StatusActive}}
	svc := &ContentModerationService{proxyRepo: repo}

	first, err := svc.resolveModerationProxyURL(context.Background(), 7)
	require.NoError(t, err)
	second, err := svc.resolveModerationProxyURL(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Equal(t, int64(1), repo.getCall.Load())
}

func TestContentModerationProxyResolutionFailureIsObservable(t *testing.T) {
	repo := &moderationProxyRepo{err: errors.New("deleted")}
	svc := &ContentModerationService{proxyRepo: repo}

	_, err := svc.resolveModerationProxyURL(context.Background(), 9)
	require.Error(t, err)
	require.Contains(t, err.Error(), "resolve moderation proxy")
	require.Equal(t, int64(1), repo.getCall.Load())

	// A failed lookup must not populate the cache and therefore remains observable.
	repo.err = nil
	repo.proxy = &Proxy{ID: 9, Protocol: "http", Host: "127.0.0.1", Port: 8081, Status: StatusActive}
	_, err = svc.resolveModerationProxyURL(context.Background(), 9)
	require.NoError(t, err)
	require.Equal(t, int64(2), repo.getCall.Load())
}

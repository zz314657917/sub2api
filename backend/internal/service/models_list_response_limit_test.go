package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestResolveModelsListReadLimit(t *testing.T) {
	require.Equal(t, config.DefaultModelsListReadMaxBytes, resolveModelsListReadLimit(nil))
	require.Equal(t, config.DefaultModelsListReadMaxBytes, resolveModelsListReadLimit(&config.Config{}))
	require.Equal(t, int64(16<<20), resolveModelsListReadLimit(&config.Config{Gateway: config.GatewayConfig{ModelsListReadMaxBytes: 16 << 20}}))
}

func TestAntigravityQuotaFetcherModelsListLimit(t *testing.T) {
	cfg := &config.Config{Gateway: config.GatewayConfig{ModelsListReadMaxBytes: 16 << 20}}
	fetcher := NewAntigravityQuotaFetcher(nil, cfg)
	require.Equal(t, int64(16<<20), resolveModelsListReadLimit(fetcher.cfg))
	require.Equal(t, config.DefaultModelsListReadMaxBytes, resolveModelsListReadLimit(NewAntigravityQuotaFetcher(nil, nil).cfg))
}

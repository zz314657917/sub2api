package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCNValidateProbeURL_AllowlistPolicy(t *testing.T) {
	cfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: true, UpstreamHosts: []string{"api.kimi.com"}}}}
	_, err := cnValidateProbeURL(cfg, "https://blocked.example/v1/usages")
	require.Error(t, err)
	url, err := cnValidateProbeURL(cfg, "https://api.kimi.com/v1/usages")
	require.NoError(t, err)
	require.Equal(t, "https://api.kimi.com/v1/usages", url)
}

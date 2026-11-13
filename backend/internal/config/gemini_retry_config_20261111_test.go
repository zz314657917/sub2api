package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGeminiRetryConfig20261111(t *testing.T) {
	resetViperWithJWTSecret(t)
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, 4, cfg.Gateway.MaxSameAccountRetriesGemini)
	for _, n := range []int{0, 10, -1, 11} {
		cfg := &Config{}
		cfg.Gateway.MaxSameAccountRetriesGemini = n
		err := cfg.Validate()
		if n < 0 || n > 10 {
			require.ErrorContains(t, err, "gateway.max_same_account_retries_gemini")
		} else if err != nil {
			require.False(t, strings.Contains(err.Error(), "gateway.max_same_account_retries_gemini"))
		}
	}
	viper.Set("gateway.max_same_account_retries_gemini", 0)
	cfg, err = Load()
	require.NoError(t, err)
	require.Zero(t, cfg.Gateway.MaxSameAccountRetriesGemini)
}

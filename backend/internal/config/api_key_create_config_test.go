package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyCreateConfigDefaultsAndZero(t *testing.T) {
	resetViperWithJWTSecret(t)
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, 200, cfg.APIKeyCreate.MaxActivePerUser)
	require.Equal(t, 60, cfg.APIKeyCreate.MaxPerUserPerHour)
	viper.Set("api_key_create.max_active_per_user", 0)
	cfg, err = Load()
	require.NoError(t, err)
	require.Zero(t, cfg.APIKeyCreate.MaxActivePerUser)
	require.Equal(t, 60, cfg.APIKeyCreate.MaxPerUserPerHour)
	viper.Set("api_key_create.max_active_per_user", 200)
	viper.Set("api_key_create.max_per_user_per_hour", 0)
	cfg, err = Load()
	require.NoError(t, err)
	require.Equal(t, 200, cfg.APIKeyCreate.MaxActivePerUser)
	require.Zero(t, cfg.APIKeyCreate.MaxPerUserPerHour)
}

func TestAPIKeyCreateConfigRejectsNegative(t *testing.T) {
	for _, key := range []string{"max_active_per_user", "max_per_user_per_hour"} {
		t.Run(key, func(t *testing.T) {
			resetViperWithJWTSecret(t)
			viper.Set("api_key_create."+key, -1)
			_, err := Load()
			require.ErrorContains(t, err, "api_key_create."+key)
		})
	}
}

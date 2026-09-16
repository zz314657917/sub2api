package repository

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyFirstResponseTimeoutRouteJSONRoundTrip(t *testing.T) {
	routes := []domain.APIKeyMultiGroupRoute{{
		GroupID:                     11,
		Priority:                    2,
		Weight:                      3,
		CooldownSeconds:             30,
		FirstResponseTimeoutSeconds: 45,
		Enabled:                     true,
	}}

	raw, err := marshalAPIKeyMultiGroupRoutes(routes)
	require.NoError(t, err)
	require.Contains(t, raw, `"first_response_timeout_seconds":45`)

	var decoded []domain.APIKeyMultiGroupRoute
	require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
	require.Equal(t, routes, decoded)

	var legacy []domain.APIKeyMultiGroupRoute
	require.NoError(t, json.Unmarshal([]byte(`[{"group_id":11,"enabled":true}]`), &legacy))
	require.Zero(t, legacy[0].FirstResponseTimeoutSeconds)
}

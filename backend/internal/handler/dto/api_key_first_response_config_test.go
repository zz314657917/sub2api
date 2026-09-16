package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyFromServicePreservesFirstResponseTimeout(t *testing.T) {
	out := APIKeyFromService(&service.APIKey{MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
		GroupID:                     11,
		FirstResponseTimeoutSeconds: 30,
	}}})

	require.Len(t, out.MultiGroupRoutes, 1)
	require.Equal(t, 30, out.MultiGroupRoutes[0].FirstResponseTimeoutSeconds)
}

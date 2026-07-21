package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestS91APIKeyFromServiceDoesNotExposeLegacyModelPatterns(t *testing.T) {
	out := APIKeyFromService(&service.APIKey{
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:       11,
			ModelPatterns: []string{"gpt-*"},
		}},
	})

	require.Len(t, out.MultiGroupRoutes, 1)
	require.Nil(t, out.MultiGroupRoutes[0].ModelPatterns)
}

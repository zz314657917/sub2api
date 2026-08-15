package server_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminOpenAIQuotaRefreshRouteContract(t *testing.T) {
	source, err := os.ReadFile("routes/admin.go")
	require.NoError(t, err)
	routes := string(source)
	require.Contains(t, routes, "openai.GET(\"/accounts/:id/quota\", h.Admin.OpenAIOAuth.QueryQuota)")
	require.Contains(t, routes, "openai.POST(\"/accounts/:id/quota/refresh\", h.Admin.OpenAIOAuth.RefreshQuota)")
}

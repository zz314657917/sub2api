package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGatewayRoutesCodexModelsManifestPathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()
	registered := make(map[string]string)
	for _, route := range router.Routes() {
		if route.Method == http.MethodGet {
			registered[route.Path] = route.Handler
		}
	}
	require.NotEmpty(t, registered["/backend-api/codex/models"])
	require.NotEmpty(t, registered["/v1/models"])
	require.NotEmpty(t, registered["/models"])
}

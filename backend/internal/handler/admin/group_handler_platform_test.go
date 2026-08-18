//go:build unit

package admin

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func bindGroupPlatformJSON(t *testing.T, target any, body string) error {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c.ShouldBindJSON(target)
}

func TestGroupPlatformBinding_AllowedPlatforms(t *testing.T) {
	allowed := []string{
		"anthropic", "openai", "gemini", "antigravity", "grok",
		"kimi", "zhipu", "deepseek",
	}
	for _, platform := range allowed {
		t.Run("create_"+platform, func(t *testing.T) {
			var req CreateGroupRequest
			body := fmt.Sprintf(`{"name":"g","platform":%q}`, platform)
			require.NoError(t, bindGroupPlatformJSON(t, &req, body))
			require.Equal(t, platform, req.Platform)
		})
		t.Run("update_"+platform, func(t *testing.T) {
			var req UpdateGroupRequest
			body := fmt.Sprintf(`{"platform":%q}`, platform)
			require.NoError(t, bindGroupPlatformJSON(t, &req, body))
			require.Equal(t, platform, req.Platform)
		})
	}
}

func TestGroupPlatformBinding_RejectsInvalidPlatforms(t *testing.T) {
	invalid := []string{"moonshot", "Kimi", "openai ", "glm", "bogus"}
	for _, platform := range invalid {
		t.Run("create_"+platform, func(t *testing.T) {
			var req CreateGroupRequest
			body := fmt.Sprintf(`{"name":"g","platform":%q}`, platform)
			require.Error(t, bindGroupPlatformJSON(t, &req, body))
		})
		t.Run("update_"+platform, func(t *testing.T) {
			var req UpdateGroupRequest
			body := fmt.Sprintf(`{"platform":%q}`, platform)
			require.Error(t, bindGroupPlatformJSON(t, &req, body))
		})
	}
}

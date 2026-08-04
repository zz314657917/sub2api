package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type s76AvailabilityDiagnoser struct {
	platform string
}

func (d *s76AvailabilityDiagnoser) DiagnoseModelAvailabilityForPlatform(
	_ context.Context,
	_ *int64,
	_ string,
	platform string,
) service.ModelAvailabilityDiagnosis {
	d.platform = platform
	return service.ModelAvailabilityDiagnosis{HasAccountsInPool: true, HasModelSupport: false}
}

func TestClassifyOpenAICompatibleNoAccountError_GrokUsesGrokPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	diag := &s76AvailabilityDiagnoser{}
	groupID := int64(43)
	apiKey := &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformGrok,
		},
	}
	platform := openAICompatibleRequestPlatform(apiKey)

	cls := classifyNoAccountErrorFromGin(c, diag, apiKey, "grok-4.5", "grok-4.5", platform)

	require.Equal(t, http.StatusNotFound, cls.Status)
	require.Equal(t, "model_not_found", cls.ErrType)
	require.True(t, cls.ModelNotFound)
	require.Equal(t, service.PlatformGrok, diag.platform)

	logErr := openAICompatibleSelectionErrorForLog(
		fmt.Errorf("no available OpenAI accounts supporting model: grok-4.5"),
		platform,
	)
	require.EqualError(t, logErr, "no available Grok accounts supporting model: grok-4.5")
}

func TestOpenAICompatibleSelectionErrorForLog_PreservesOtherPlatforms(t *testing.T) {
	err := fmt.Errorf("no available OpenAI accounts supporting model: gpt-5")
	require.Same(t, err, openAICompatibleSelectionErrorForLog(err, service.PlatformOpenAI))
	require.NoError(t, openAICompatibleSelectionErrorForLog(nil, service.PlatformGrok))
}

func TestClassifyNoAccountError_CafeManagedKeyKeepsStableUnavailableCode(t *testing.T) {
	managedSourceID := int64(701)
	apiKey := &service.APIKey{
		ManagedSourceType: service.APIKeyManagedSourceCafeRoomSeat,
		ManagedSourceID:   &managedSourceID,
	}

	cls := classifyNoAccountError(context.Background(), nil, apiKey, "", "", service.PlatformOpenAI)

	require.Equal(t, http.StatusForbidden, cls.Status)
	require.Equal(t, "CAFE_ACCOUNT_UNAVAILABLE", cls.ErrType)
	require.Equal(t, "the cafe account is temporarily unavailable", cls.Message)
	require.False(t, cls.ModelNotFound)
}

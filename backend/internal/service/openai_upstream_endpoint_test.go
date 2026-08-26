package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestActualOpenAIUpstreamEndpointIsRequestScoped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	require.Empty(t, GetActualOpenAIUpstreamEndpoint(c))
	SetActualOpenAIUpstreamEndpoint(c, " /v1/chat/completions ")
	require.Equal(t, "/v1/chat/completions", GetActualOpenAIUpstreamEndpoint(c))

	ClearActualOpenAIUpstreamEndpoint(c)
	require.Empty(t, GetActualOpenAIUpstreamEndpoint(c))
}

func TestForwardRecordsRawChatEndpointBeforeResponsesFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceChatCompletions),
		},
	}

	_, err := (&OpenAIGatewayService{}).Forward(context.Background(), c, account, []byte(`{`))
	require.Error(t, err)
	// The malformed payload fails before an OpenAIForwardResult exists. The
	// handler can still log the actual raw endpoint selected for this account.
	require.Equal(t, "/v1/chat/completions", GetActualOpenAIUpstreamEndpoint(c))
}

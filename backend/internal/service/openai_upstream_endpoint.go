package service

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// openAIUpstreamEndpointContextKey stores the protocol endpoint selected by
// the latest OpenAI forwarding attempt. It is intentionally request-scoped:
// failover reuses the same Gin context and must overwrite the value on every
// new upstream attempt.
const openAIUpstreamEndpointContextKey = "openai_actual_upstream_endpoint"

// SetActualOpenAIUpstreamEndpoint records a normalized upstream endpoint for
// the current request. Empty values clear the override.
func SetActualOpenAIUpstreamEndpoint(c *gin.Context, endpoint string) {
	if c == nil {
		return
	}
	c.Set(openAIUpstreamEndpointContextKey, strings.TrimSpace(endpoint))
}

// ClearActualOpenAIUpstreamEndpoint clears the endpoint from a previous
// failover attempt before a new forwarding path starts.
func ClearActualOpenAIUpstreamEndpoint(c *gin.Context) {
	SetActualOpenAIUpstreamEndpoint(c, "")
}

// GetActualOpenAIUpstreamEndpoint returns the endpoint recorded by the latest
// OpenAI forwarding attempt, or an empty string when no attempt was built.
func GetActualOpenAIUpstreamEndpoint(c *gin.Context) string {
	if c == nil {
		return ""
	}
	value, ok := c.Get(openAIUpstreamEndpointContextKey)
	if !ok {
		return ""
	}
	endpoint, _ := value.(string)
	return strings.TrimSpace(endpoint)
}

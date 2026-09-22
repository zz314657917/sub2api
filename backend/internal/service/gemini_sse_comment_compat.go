package service

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func geminiClientRejectsSSEComments(clientHint string) bool {
	hint := strings.ToLower(strings.TrimSpace(clientHint))
	return strings.Contains(hint, "google-genai-sdk/") &&
		(strings.Contains(hint, "gl-go/") || strings.Contains(hint, "gl-python/"))
}

func downstreamRejectsSSEComments(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	return geminiClientRejectsSSEComments(c.GetHeader("User-Agent")) ||
		geminiClientRejectsSSEComments(c.GetHeader("X-Goog-Api-Client"))
}

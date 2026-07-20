package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestS87APIKeyQuotaErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	abortWithAPIKeyQuotaError(c)

	var body struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Param   any    `json:"param"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusTooManyRequests || body.Error.Type != "insufficient_quota" || body.Error.Code != "insufficient_quota" || body.Error.Param != nil {
		t.Fatalf("unexpected OpenAI quota response: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestS87APIKeyQuotaErrorPathMatrix(t *testing.T) {
	for _, tt := range []struct {
		path string
		want bool
	}{
		{path: "/v1/responses", want: true},
		{path: "/v1/responses/compact/detail", want: true},
		{path: "/responses", want: true},
		{path: "/backend-api/codex/responses/compact", want: true},
		{path: "/v1/responsesx", want: false},
		{path: "/responses-old", want: false},
		{path: "/backend-api/codex/responsesx", want: false},
		{path: "/v1/messages", want: false},
		{path: "/v1/chat/completions", want: false},
		{path: "/v1/images/generations", want: false},
		{path: "/v1/usage", want: false},
		{path: "/v1beta/models/gemini:generateContent", want: false},
	} {
		t.Run(tt.path, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)
			if got := isOpenAIResponsesAPIKeyRequest(c); got != tt.want {
				t.Fatalf("isOpenAIResponsesAPIKeyRequest(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

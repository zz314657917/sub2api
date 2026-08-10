package handler

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func TestS209ValidateAPIKeyCreateRequest(t *testing.T) {
	zero, large := 0.0, 1e100
	positiveDays := 1
	if err := validateAPIKeyCreateRequest(CreateAPIKeyRequest{}); err != nil {
		t.Fatalf("empty create request rejected: %v", err)
	}
	if err := validateAPIKeyCreateRequest(CreateAPIKeyRequest{
		Quota: &zero, RateLimit5h: &large, RateLimit1d: &zero, RateLimit7d: &large, ExpiresInDays: &positiveDays,
	}); err != nil {
		t.Fatalf("valid create request rejected: %v", err)
	}

	negative, nan, positiveInf, negativeInf := -1.0, math.NaN(), math.Inf(1), math.Inf(-1)
	zeroDays, negativeDays := 0, -1
	for _, tt := range []struct {
		name string
		req  CreateAPIKeyRequest
	}{
		{name: "negative quota", req: CreateAPIKeyRequest{Quota: &negative}},
		{name: "nan quota", req: CreateAPIKeyRequest{Quota: &nan}},
		{name: "positive infinite 5h", req: CreateAPIKeyRequest{RateLimit5h: &positiveInf}},
		{name: "negative infinite 1d", req: CreateAPIKeyRequest{RateLimit1d: &negativeInf}},
		{name: "negative 7d", req: CreateAPIKeyRequest{RateLimit7d: &negative}},
		{name: "zero expiry", req: CreateAPIKeyRequest{ExpiresInDays: &zeroDays}},
		{name: "negative expiry", req: CreateAPIKeyRequest{ExpiresInDays: &negativeDays}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateAPIKeyCreateRequest(tt.req); err == nil {
				t.Fatal("invalid create request was accepted")
			}
		})
	}
}

func TestS209ValidateAPIKeyUpdateRequest(t *testing.T) {
	zero, large := 0.0, 1e100
	if err := validateAPIKeyUpdateRequest(UpdateAPIKeyRequest{Quota: &zero, RateLimit7d: &large}); err != nil {
		t.Fatalf("valid update request rejected: %v", err)
	}

	negative, nan, positiveInf, negativeInf := -1.0, math.NaN(), math.Inf(1), math.Inf(-1)
	for _, tt := range []struct {
		name string
		req  UpdateAPIKeyRequest
	}{
		{name: "negative quota", req: UpdateAPIKeyRequest{Quota: &negative}},
		{name: "nan 5h", req: UpdateAPIKeyRequest{RateLimit5h: &nan}},
		{name: "positive infinite 1d", req: UpdateAPIKeyRequest{RateLimit1d: &positiveInf}},
		{name: "negative infinite 7d", req: UpdateAPIKeyRequest{RateLimit7d: &negativeInf}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateAPIKeyUpdateRequest(tt.req); err == nil {
				t.Fatal("invalid update request was accepted")
			}
		})
	}
}

func TestS209APIKeyHandlerRejectsInvalidInputBeforeExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAPIKeyHandler(nil)
	router := gin.New()
	router.POST("/api/v1/api-keys", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		h.Create(c)
	})
	router.PUT("/api/v1/api-keys/:id", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		h.Update(c)
	})

	for _, tt := range []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create negative quota", method: http.MethodPost, path: "/api/v1/api-keys", body: `{"name":"invalid","quota":-1}`},
		{name: "create zero expiry", method: http.MethodPost, path: "/api/v1/api-keys", body: `{"name":"invalid","expires_in_days":0}`},
		{name: "update negative rate", method: http.MethodPut, path: "/api/v1/api-keys/7", body: `{"rate_limit_1d":-1}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "s209-invalid-input")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), "numeric limits must be finite and non-negative") {
				t.Fatalf("unexpected response body: %s", response.Body.String())
			}
		})
	}
}

package service

import (
	"context"
	"math"
	"net/http"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func TestS209ValidateCreateAPIKeyRequest(t *testing.T) {
	positiveDays := 1
	if err := validateCreateAPIKeyRequest(CreateAPIKeyRequest{}); err != nil {
		t.Fatalf("empty create request rejected: %v", err)
	}
	if err := validateCreateAPIKeyRequest(CreateAPIKeyRequest{
		Quota: 1e100, RateLimit5h: 0, RateLimit1d: 1e100, RateLimit7d: 0, ExpiresInDays: &positiveDays,
	}); err != nil {
		t.Fatalf("valid create request rejected: %v", err)
	}

	zeroDays, negativeDays := 0, -1
	for _, tt := range []struct {
		name       string
		req        CreateAPIKeyRequest
		wantReason string
	}{
		{name: "negative quota", req: CreateAPIKeyRequest{Quota: -1}, wantReason: "API_KEY_LIMIT_INVALID"},
		{name: "nan quota", req: CreateAPIKeyRequest{Quota: math.NaN()}, wantReason: "API_KEY_LIMIT_INVALID"},
		{name: "positive infinite 5h", req: CreateAPIKeyRequest{RateLimit5h: math.Inf(1)}, wantReason: "API_KEY_LIMIT_INVALID"},
		{name: "nan 1d", req: CreateAPIKeyRequest{RateLimit1d: math.NaN()}, wantReason: "API_KEY_LIMIT_INVALID"},
		{name: "negative infinite 7d", req: CreateAPIKeyRequest{RateLimit7d: math.Inf(-1)}, wantReason: "API_KEY_LIMIT_INVALID"},
		{name: "zero expiry", req: CreateAPIKeyRequest{ExpiresInDays: &zeroDays}, wantReason: "API_KEY_EXPIRY_INVALID"},
		{name: "negative expiry", req: CreateAPIKeyRequest{ExpiresInDays: &negativeDays}, wantReason: "API_KEY_EXPIRY_INVALID"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertS209APIKeyValidationError(t, validateCreateAPIKeyRequest(tt.req), tt.wantReason)
		})
	}
}

func TestS209ValidateUpdateAPIKeyRequest(t *testing.T) {
	zero, large := 0.0, 1e100
	if err := validateUpdateAPIKeyRequest(UpdateAPIKeyRequest{Quota: &zero, RateLimit7d: &large}); err != nil {
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
			assertS209APIKeyValidationError(t, validateUpdateAPIKeyRequest(tt.req), "API_KEY_LIMIT_INVALID")
		})
	}
}

func TestS209APIKeyServiceRejectsInvalidInputBeforeRepositories(t *testing.T) {
	svc := &APIKeyService{}
	_, createErr := svc.Create(context.Background(), 42, CreateAPIKeyRequest{Quota: -1})
	assertS209APIKeyValidationError(t, createErr, "API_KEY_LIMIT_INVALID")

	invalidRate := math.Inf(1)
	_, updateErr := svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{RateLimit5h: &invalidRate})
	assertS209APIKeyValidationError(t, updateErr, "API_KEY_LIMIT_INVALID")
}

func assertS209APIKeyValidationError(t *testing.T, err error, wantReason string) {
	t.Helper()
	if err == nil {
		t.Fatal("invalid API key request was accepted")
	}
	if got := infraerrors.Code(err); got != http.StatusBadRequest {
		t.Fatalf("error code = %d, want %d: %v", got, http.StatusBadRequest, err)
	}
	if got := infraerrors.Reason(err); got != wantReason {
		t.Fatalf("error reason = %q, want %q: %v", got, wantReason, err)
	}
}

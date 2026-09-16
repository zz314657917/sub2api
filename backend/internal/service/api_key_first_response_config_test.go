package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func TestAPIKeyFirstResponseTimeoutValidation(t *testing.T) {
	for _, seconds := range []int{0, 1, 30, 600} {
		err := validateCreateAPIKeyRequest(CreateAPIKeyRequest{MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:                     1,
			FirstResponseTimeoutSeconds: seconds,
		}}})
		if err != nil {
			t.Fatalf("seconds %d rejected: %v", seconds, err)
		}
	}

	for _, seconds := range []int{-1, 601} {
		err := validateUpdateAPIKeyRequest(UpdateAPIKeyRequest{MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:                     1,
			FirstResponseTimeoutSeconds: seconds,
		}}})
		if infraerrors.Reason(err) != "API_KEY_ROUTE_INVALID" {
			t.Fatalf("seconds %d error = %v, want API_KEY_ROUTE_INVALID", seconds, err)
		}
	}
}

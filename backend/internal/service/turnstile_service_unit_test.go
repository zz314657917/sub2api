//go:build unit

package service

import (
	"context"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestTurnstileService_VerifyToken_FailureIncludesErrorCodesMetadata(t *testing.T) {
	verifier := &turnstileVerifierSpy{
		result: &TurnstileVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"timeout-or-duplicate", "invalid-input-response"},
		},
	}
	service := NewTurnstileService(
		&SettingService{settingRepo: &settingRepoStub{values: map[string]string{
			SettingKeyTurnstileEnabled:   "true",
			SettingKeyTurnstileSecretKey: "secret",
		}}},
		verifier,
	)

	err := service.VerifyToken(context.Background(), "token", "1.1.1.1")

	require.ErrorIs(t, err, ErrTurnstileVerificationFailed)
	appErr := infraErrorFromError(t, err)
	require.Equal(t, "timeout-or-duplicate,invalid-input-response", appErr.Metadata["turnstile_error_codes"])
}

func infraErrorFromError(t *testing.T, err error) *infraerrors.ApplicationError {
	t.Helper()
	appErr := infraerrors.FromError(err)
	require.NotNil(t, appErr)
	return appErr
}

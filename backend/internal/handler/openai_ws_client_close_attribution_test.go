package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

func TestOpenAIWSIngressEndedByClient_BareNormalClosure(t *testing.T) {
	err := coderws.CloseError{Code: coderws.StatusNormalClosure, Reason: "client done"}
	var wrapped *service.OpenAIWSClientCloseError
	require.False(t, errors.As(err, &wrapped))
	require.Equal(t, coderws.StatusNormalClosure, coderws.CloseStatus(err))
	require.True(t, openAIWSIngressEndedByClient(err))
}

func TestOpenAIWSIngressEndedByClient_WrappedBareNormalClosure(t *testing.T) {
	err := fmt.Errorf("ingress turn: %w", coderws.CloseError{Code: coderws.StatusNormalClosure})
	require.True(t, openAIWSIngressEndedByClient(err))
}

func TestOpenAIWSIngressEndedByClient_ClientCancellation(t *testing.T) {
	err := service.NewOpenAIWSClientCloseError(coderws.StatusGoingAway, "websocket request canceled", context.Canceled)
	require.True(t, openAIWSIngressEndedByClient(err))
}

func TestOpenAIWSIngressEndedByClient_GatewayNormalClosure(t *testing.T) {
	err := service.NewOpenAIWSClientCloseError(coderws.StatusNormalClosure, "websocket idle timeout", context.DeadlineExceeded)
	require.True(t, openAIWSIngressEndedByClient(err))
}

func TestOpenAIWSIngressEndedByClient_GoingAwayWithoutCancellation(t *testing.T) {
	err := service.NewOpenAIWSClientCloseError(coderws.StatusGoingAway, "upstream going away", errors.New("upstream closed session"))
	require.False(t, openAIWSIngressEndedByClient(err))
}

func TestOpenAIWSIngressEndedByClient_AbnormalClosures(t *testing.T) {
	cases := []error{
		service.NewOpenAIWSClientCloseError(coderws.StatusPolicyViolation, "upstream rejected credentials", errors.New("rejected")),
		service.NewOpenAIWSClientCloseError(coderws.StatusInternalError, "upstream websocket proxy failed", nil),
		coderws.CloseError{Code: coderws.StatusAbnormalClosure, Reason: "connection reset"},
		errors.New("upstream websocket read failed"),
		fmt.Errorf("upstream stalled: %w", context.DeadlineExceeded),
	}
	for _, err := range cases {
		require.False(t, openAIWSIngressEndedByClient(err), "%v", err)
	}
}

func TestOpenAIWSIngressEndedByClient_MatchesCloseCodeLog(t *testing.T) {
	errs := []error{
		coderws.CloseError{Code: coderws.StatusNormalClosure},
		fmt.Errorf("wrapped: %w", coderws.CloseError{Code: coderws.StatusNormalClosure}),
		service.NewOpenAIWSClientCloseError(coderws.StatusNormalClosure, "idle", context.DeadlineExceeded),
		service.NewOpenAIWSClientCloseError(coderws.StatusGoingAway, "canceled", context.Canceled),
		coderws.CloseError{Code: coderws.StatusAbnormalClosure},
		errors.New("upstream read failed"),
	}
	for _, err := range errs {
		status, _ := summarizeWSCloseErrorForLog(err)
		if status == "1000(StatusNormalClosure)" {
			require.True(t, openAIWSIngressEndedByClient(err))
		}
	}
}

func TestShouldReportOpenAIWSProxyAccountFailure(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "normal client closure is not attributed",
			err:  coderws.CloseError{Code: coderws.StatusNormalClosure},
			want: false,
		},
		{
			name: "client cancellation is not attributed",
			err:  service.NewOpenAIWSClientCloseError(coderws.StatusGoingAway, "client canceled", context.Canceled),
			want: false,
		},
		{
			name: "model switch is filtered",
			err:  newOpenAIWSUnsupportedModelSwitchError("gpt-5"),
			want: false,
		},
		{
			name: "real upstream failure is attributed",
			err:  errors.New("upstream websocket read failed"),
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, shouldReportOpenAIWSProxyAccountFailure(tc.err))
		})
	}
}

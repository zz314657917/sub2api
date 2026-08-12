package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestGatewayAccountAvailabilityUsesRequestStartedAt(t *testing.T) {
	setTestTimezone(t, "UTC")
	account := newAccountWithAvailability(true, "18:00", "22:00")
	svc := &GatewayService{}

	inside := context.WithValue(context.Background(), ctxkey.RequestStartedAt, at(21, 59))
	require.True(t, svc.isAccountSchedulableForSelection(inside, account))
	require.False(t, shouldClearStickySessionWithContext(inside, account, ""))

	outside := context.WithValue(context.Background(), ctxkey.RequestStartedAt, at(22, 0))
	require.False(t, svc.isAccountSchedulableForSelection(outside, account))
	require.True(t, shouldClearStickySessionWithContext(outside, account, ""))
}

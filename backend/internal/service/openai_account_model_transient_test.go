package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIModelTransient_StreakSurvivesSparseTraffic(t *testing.T) {
	state := newOpenAIAccountModelTransientState(128)
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	gap := 5 * time.Minute
	require.Greater(t, gap, openAIModelTransientLongCooldown)

	first := state.recordFailure(35, "gpt-5.5", now)
	second := state.recordFailure(35, "gpt-5.5", now.Add(gap))
	third := state.recordFailure(35, "gpt-5.5", now.Add(2*gap))

	assert.Equal(t, 1, first.FailureStreak)
	assert.Zero(t, first.Cooldown)
	assert.Equal(t, 2, second.FailureStreak)
	assert.Equal(t, openAIModelTransientShortCooldown, second.Cooldown)
	assert.Equal(t, 3, third.FailureStreak)
	assert.Equal(t, openAIModelTransientLongCooldown, third.Cooldown)
	assert.True(t, state.isBlocked(35, "gpt-5.5", now.Add(2*gap+time.Second)))
}

func TestOpenAIModelTransient_SuccessResetsAndStaleStateExpires(t *testing.T) {
	state := newOpenAIAccountModelTransientState(128)
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)

	state.recordFailure(35, "gpt-5.5", now)
	state.recordSuccess(35, "gpt-5.5")
	decision := state.recordFailure(35, "gpt-5.5", now.Add(5*time.Minute))
	assert.Equal(t, 1, decision.FailureStreak)

	decision = state.recordFailure(35, "gpt-5.5", now.Add(5*time.Minute+openAIModelTransientStreakTTL+time.Second))
	assert.Equal(t, 1, decision.FailureStreak)
	assert.Zero(t, decision.Cooldown)
}

func TestOpenAIModelTransient_ClockRollbackClearsState(t *testing.T) {
	state := newOpenAIAccountModelTransientState(128)
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)

	state.recordFailure(35, "gpt-5.5", now)
	state.recordFailure(35, "gpt-5.5", now.Add(time.Second))
	assert.True(t, state.isBlocked(35, "gpt-5.5", now.Add(2*time.Second)))

	assert.False(t, state.isBlocked(35, "gpt-5.5", now.Add(-time.Second)))
	decision := state.recordFailure(35, "gpt-5.5", now)
	assert.Equal(t, 1, decision.FailureStreak)
	assert.Zero(t, decision.Cooldown)
}

func TestOpenAIModelTransient_OnlyBlocksAffectedAccountAndModel(t *testing.T) {
	svc := &OpenAIGatewayService{openaiModelTransient: newOpenAIAccountModelTransientState(128)}
	account := &Account{ID: 35, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	now := time.Now()
	svc.recordOpenAIAccountModelTransientFailure(account, "gpt-5.5", now.Add(-time.Second))
	svc.recordOpenAIAccountModelTransientFailure(account, "gpt-5.5", now)

	assert.True(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "GPT-5.5"))
	assert.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5.4"))
	assert.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(&Account{ID: 36}, "gpt-5.5"))

	svc.ReportOpenAIAccountScheduleResult(account.ID, true, nil, "gpt-5.5")
	assert.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5.5"))
}

func TestOpenAIModelTransient_OnlyRecordsEligibleTransientFailures(t *testing.T) {
	svc := &OpenAIGatewayService{openaiModelTransient: newOpenAIAccountModelTransientState(128)}
	account := &Account{ID: 35, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

	svc.handleOpenAIAccountUpstreamError(context.Background(), account, http.StatusInternalServerError, nil, nil, "gpt-5.5")
	svc.handleOpenAIAccountUpstreamError(context.Background(), account, http.StatusInternalServerError, nil, nil, "gpt-5.5")
	assert.True(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5.5"))

	other := &OpenAIGatewayService{openaiModelTransient: newOpenAIAccountModelTransientState(128)}
	other.handleOpenAIAccountUpstreamError(context.Background(), account, http.StatusTooManyRequests, nil, nil, "gpt-5.5")
	other.handleOpenAIAccountUpstreamError(context.Background(), account, 529, nil, nil, "gpt-5.5")
	other.handleOpenAIAccountUpstreamError(context.Background(), account, http.StatusBadRequest, nil, []byte(`{"error":{"message":"missing required parameter"}}`), "gpt-5.5")
	assert.False(t, other.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5.5"))
}

func TestOpenAIModelTransient_WSFailedTerminalDoesNotClearState(t *testing.T) {
	svc := &OpenAIGatewayService{openaiModelTransient: newOpenAIAccountModelTransientState(128)}
	account := &Account{ID: 35, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	svc.recordOpenAIAccountModelTransientFailure(account, "gpt-5.5", time.Now())
	svc.recordOpenAIAccountModelTransientFailure(account, "gpt-5.5", time.Now())

	failed := &OpenAIForwardResult{OpenAIWSMode: true, UpstreamTerminalEvent: "response.failed"}
	svc.ReportOpenAIAccountScheduleResult(account.ID, failed.SucceededForScheduling(), nil, "gpt-5.5")
	assert.True(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5.5"))

	completed := &OpenAIForwardResult{OpenAIWSMode: true, UpstreamTerminalEvent: "response.completed"}
	svc.ReportOpenAIAccountScheduleResult(account.ID, completed.SucceededForScheduling(), nil, "gpt-5.5")
	assert.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5.5"))
}

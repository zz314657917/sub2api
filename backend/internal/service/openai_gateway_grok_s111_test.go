package service

import (
	"context"
	"net/http"
	"testing"
	"time"
)

type grokS111AccountRepo struct {
	AccountRepository
	tempUnschedCalls int
	accountID        int64
	until            time.Time
	reason           string
}

func (r *grokS111AccountRepo) SetTempUnschedulable(_ context.Context, id int64, until time.Time, reason string) error {
	r.tempUnschedCalls++
	r.accountID = id
	r.until = until
	r.reason = reason
	return nil
}

func TestHandleGrokAccountUpstreamErrorPaymentRequiredPausesAccount(t *testing.T) {
	account := &Account{
		ID:          111,
		Platform:    PlatformGrok,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
	}
	repo := &grokS111AccountRepo{}
	svc := &OpenAIGatewayService{accountRepo: repo}
	before := time.Now()

	svc.handleGrokAccountUpstreamError(
		context.Background(),
		account,
		http.StatusPaymentRequired,
		http.Header{},
		nil,
	)

	if account.TempUnschedulableUntil == nil {
		t.Fatal("payment-required Grok account was not temporarily unscheduled")
	}
	if account.TempUnschedulableReason != "grok payment required" {
		t.Fatalf("temporary unschedule reason = %q, want %q", account.TempUnschedulableReason, "grok payment required")
	}
	if account.TempUnschedulableUntil.Before(before.Add(30*time.Minute-time.Second)) ||
		account.TempUnschedulableUntil.After(before.Add(30*time.Minute+time.Second)) {
		t.Fatalf("temporary unschedule deadline = %s, want about 30 minutes after %s", account.TempUnschedulableUntil, before)
	}
	if account.IsSchedulable() {
		t.Fatal("payment-required Grok account remained schedulable during cooldown")
	}
	if repo.tempUnschedCalls != 1 || repo.accountID != account.ID {
		t.Fatalf("temporary unschedule persistence calls = %d for account %d, want 1 for account %d", repo.tempUnschedCalls, repo.accountID, account.ID)
	}
	if repo.reason != "grok payment required" {
		t.Fatalf("persisted temporary unschedule reason = %q, want %q", repo.reason, "grok payment required")
	}
	if repo.until.Before(before.Add(30*time.Minute-time.Second)) ||
		repo.until.After(before.Add(30*time.Minute+time.Second)) {
		t.Fatalf("persisted temporary unschedule deadline = %s, want about 30 minutes after %s", repo.until, before)
	}

	expired := time.Now().Add(-time.Second)
	account.TempUnschedulableUntil = &expired
	if !account.IsSchedulable() {
		t.Fatal("Grok account did not become schedulable after cooldown expiry")
	}
}

//go:build unit

package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

type modelNotFoundRateLimitCall struct {
	id      int64
	scope   string
	resetAt time.Time
}

type modelNotFoundTempUnschedCall struct {
	id     int64
	until  time.Time
	reason string
}

type modelNotFoundAccountRepoStub struct {
	mockAccountRepoForGemini
	modelRateLimitCalls []modelNotFoundRateLimitCall
	tempUnschedCalls    []modelNotFoundTempUnschedCall
	modelRateLimitErr   error
}

func (m *modelNotFoundAccountRepoStub) SetModelRateLimit(ctx context.Context, id int64, scope string, resetAt time.Time) error {
	m.modelRateLimitCalls = append(m.modelRateLimitCalls, modelNotFoundRateLimitCall{
		id:      id,
		scope:   scope,
		resetAt: resetAt,
	})
	return m.modelRateLimitErr
}

func (m *modelNotFoundAccountRepoStub) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	m.tempUnschedCalls = append(m.tempUnschedCalls, modelNotFoundTempUnschedCall{
		id:     id,
		until:  until,
		reason: reason,
	})
	return nil
}

func TestRateLimitService_HandleUpstreamError_ModelNotFoundUsesModelCooldown(t *testing.T) {
	repo := &modelNotFoundAccountRepoStub{}
	service := NewRateLimitService(repo, nil, nil, nil, nil)
	account := modelNotFoundTestAccount()
	body := []byte(`{"error":{"message":"The model 'gpt-5.4' does not exist","code":"model_not_found"}}`)

	shouldDisable := service.HandleUpstreamError(context.Background(), account, http.StatusNotFound, http.Header{}, body, "gpt-5.4")

	if !shouldDisable {
		t.Fatal("expected model 404 to trigger failover")
	}
	if len(repo.modelRateLimitCalls) != 1 {
		t.Fatalf("expected one model rate limit call, got %d", len(repo.modelRateLimitCalls))
	}
	call := repo.modelRateLimitCalls[0]
	if call.id != account.ID {
		t.Fatalf("model rate limit account id = %d, want %d", call.id, account.ID)
	}
	if call.scope != "gpt-5.4" {
		t.Fatalf("model rate limit scope = %q, want gpt-5.4", call.scope)
	}
	if until := time.Until(call.resetAt); until < 29*time.Minute || until > 31*time.Minute {
		t.Fatalf("model rate limit cooldown = %s, want about 30m", until)
	}
	if len(repo.tempUnschedCalls) != 0 {
		t.Fatalf("model 404 should not set temp unschedulable, got %d calls", len(repo.tempUnschedCalls))
	}
}

func TestRateLimitService_HandleUpstreamError_ModelNotFoundWriteFailureStillFailsOver(t *testing.T) {
	repo := &modelNotFoundAccountRepoStub{modelRateLimitErr: errors.New("db unavailable")}
	service := NewRateLimitService(repo, nil, nil, nil, nil)
	account := modelNotFoundTestAccount()
	body := []byte(`{"error":{"message":"unknown model: gpt-missing"}}`)

	shouldDisable := service.HandleUpstreamError(context.Background(), account, http.StatusNotFound, http.Header{}, body, "gpt-missing")

	if !shouldDisable {
		t.Fatal("expected write failure to still trigger failover")
	}
	if len(repo.modelRateLimitCalls) != 1 {
		t.Fatalf("expected one model rate limit call, got %d", len(repo.modelRateLimitCalls))
	}
	if len(repo.tempUnschedCalls) != 0 {
		t.Fatalf("model 404 write failure should not fall through to temp unschedulable, got %d calls", len(repo.tempUnschedCalls))
	}
}

func TestRateLimitService_HandleUpstreamError_NonModel404KeepsTempUnschedulable(t *testing.T) {
	repo := &modelNotFoundAccountRepoStub{}
	service := NewRateLimitService(repo, nil, nil, nil, nil)
	account := modelNotFoundTestAccount()
	body := []byte(`{"error":{"message":"route not found"}}`)

	shouldDisable := service.HandleUpstreamError(context.Background(), account, http.StatusNotFound, http.Header{}, body, "gpt-5.4")

	if !shouldDisable {
		t.Fatal("expected configured non-model 404 to trigger temp unschedulable")
	}
	if len(repo.modelRateLimitCalls) != 0 {
		t.Fatalf("non-model 404 should not set model rate limit, got %d calls", len(repo.modelRateLimitCalls))
	}
	if len(repo.tempUnschedCalls) != 1 {
		t.Fatalf("expected one temp unschedulable call, got %d", len(repo.tempUnschedCalls))
	}
}

func modelNotFoundTestAccount() *Account {
	return &Account{
		ID:       42,
		Name:     "openai-test",
		Type:     AccountTypeAPIKey,
		Platform: PlatformOpenAI,
		Status:   StatusActive,
		Credentials: map[string]any{
			"temp_unschedulable_enabled": true,
			"temp_unschedulable_rules": []any{
				map[string]any{
					"error_code":       float64(http.StatusNotFound),
					"duration_minutes": float64(10),
					"keywords":         []any{"not found"},
				},
			},
		},
	}
}

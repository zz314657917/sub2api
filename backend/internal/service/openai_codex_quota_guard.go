package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const openAICodex7dQuotaTempBlockMessage = "OpenAI Codex 7d usage reached 100%"

type openAICodex7dTempBlock struct {
	until  time.Time
	reason string
	state  *TempUnschedState
}

func resolveOpenAICodex7dResetAt(updates map[string]any, now time.Time) (time.Time, bool) {
	if len(updates) == 0 {
		return time.Time{}, false
	}
	if parseExtraFloat64(updates["codex_7d_used_percent"]) < 100 {
		return time.Time{}, false
	}
	resetRaw, ok := updates["codex_7d_reset_at"]
	if !ok || resetRaw == nil {
		return time.Time{}, false
	}
	resetAt, err := parseTime(fmt.Sprint(resetRaw))
	if err != nil {
		return time.Time{}, false
	}
	if !now.Before(resetAt) {
		return time.Time{}, false
	}
	return resetAt.UTC(), true
}

func resolveOpenAICodex7dTempBlock(account *Account, updates map[string]any, now time.Time) *openAICodex7dTempBlock {
	if account == nil || !account.IsOpenAIOAuth() {
		return nil
	}
	until, ok := resolveOpenAICodex7dResetAt(updates, now)
	if !ok {
		return nil
	}

	state := &TempUnschedState{
		UntilUnix:       until.Unix(),
		TriggeredAtUnix: now.Unix(),
		StatusCode:      http.StatusTooManyRequests,
		MatchedKeyword:  "codex_7d_used_percent",
		ErrorMessage:    openAICodex7dQuotaTempBlockMessage,
	}

	reason := openAICodex7dQuotaTempBlockMessage
	if raw, err := json.Marshal(state); err == nil {
		reason = string(raw)
	}
	return &openAICodex7dTempBlock{until: until, reason: reason, state: state}
}

func applyOpenAICodex7dTempBlock(ctx context.Context, repo AccountRepository, account *Account, accountID int64, updates map[string]any, now time.Time) bool {
	if repo == nil || len(updates) == 0 {
		return false
	}
	if _, ok := resolveOpenAICodex7dResetAt(updates, now); !ok {
		return false
	}
	if account == nil {
		if accountID <= 0 {
			return false
		}
		loaded, err := repo.GetByID(ctx, accountID)
		if err != nil || loaded == nil {
			if err != nil {
				slog.Warn("openai_codex_7d_temp_block_account_load_failed", "account_id", accountID, "error", err)
			}
			return false
		}
		account = loaded
	}

	block := resolveOpenAICodex7dTempBlock(account, updates, now)
	if block == nil {
		return false
	}
	if err := repo.SetTempUnschedulable(ctx, account.ID, block.until, block.reason); err != nil {
		slog.Warn("openai_codex_7d_temp_block_set_failed", "account_id", account.ID, "error", err)
		return true
	}
	slog.Info("openai_codex_7d_temp_block_set", "account_id", account.ID, "until", block.until)
	return true
}

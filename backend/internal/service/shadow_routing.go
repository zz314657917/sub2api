package service

import (
	"strconv"
	"strings"
	"time"
)

// parentHealthyForShadow deliberately checks only credential health. A parent's
// operator scheduling switch and global quota/overload state belong to its
// global quota dimension and must not suppress an independently metered shadow.
func parentHealthyForShadow(account *Account, lookup func(int64) *Account, now time.Time) bool {
	if account == nil || !account.IsShadow() || account.ParentAccountID == nil {
		return account != nil && !account.IsShadow()
	}
	parent := lookup(*account.ParentAccountID)
	if parent == nil || !parent.IsOpenAIOAuth() || !parent.IsActive() {
		return false
	}
	if parent.TempUnschedulableUntil != nil && now.Before(*parent.TempUnschedulableUntil) {
		return false
	}
	if strings.TrimSpace(parent.GetOpenAIAccessToken()) == "" {
		return false
	}
	if expiresAt, ok := shadowCredentialExpiry(parent); ok && !now.Before(expiresAt) {
		return false
	}
	return true
}

func shadowCredentialExpiry(account *Account) (time.Time, bool) {
	if account == nil {
		return time.Time{}, false
	}
	raw := strings.TrimSpace(account.GetCredential("expires_at"))
	if raw == "" {
		return time.Time{}, false
	}
	if unix, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if unix > 1e11 {
			unix /= 1000
		}
		return time.Unix(unix, 0), true
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	return parsed, err == nil
}

func defaultSparkShadowModelMapping() map[string]any {
	return map[string]any{"gpt-5.3-codex-spark": "gpt-5.3-codex-spark"}
}

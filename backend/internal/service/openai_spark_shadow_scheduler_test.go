package service

import (
	"testing"
	"time"
)

func TestSparkShadowParentHealthFailsClosedAndKeepsGlobalQuotaIndependent(t *testing.T) {
	now := time.Now()
	parentID := int64(1)
	shadow := &Account{ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark}
	parent := &Account{ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: false,
		RateLimitResetAt: sparkShadowTimePtr(now.Add(time.Hour)), OverloadUntil: sparkShadowTimePtr(now.Add(time.Hour)), Credentials: map[string]any{"access_token": "parent-token", "expires_at": now.Add(time.Hour).Format(time.RFC3339)}}
	if !parentHealthyForShadow(shadow, func(int64) *Account { return parent }, now) {
		t.Fatal("parent global quota or manual schedulable state must not block spark shadow")
	}
	parent.TempUnschedulableUntil = sparkShadowTimePtr(now.Add(time.Minute))
	if parentHealthyForShadow(shadow, func(int64) *Account { return parent }, now) {
		t.Fatal("temporary parent credential failure must block shadow")
	}
	if mapping := defaultSparkShadowModelMapping(); len(mapping) != 1 || mapping["gpt-5.3-codex-spark"] != "gpt-5.3-codex-spark" {
		t.Fatalf("unexpected default spark mapping: %#v", mapping)
	}
}

func sparkShadowTimePtr(v time.Time) *time.Time { return &v }

func TestSparkShadowTokenRefresherRejectsChild(t *testing.T) {
	parentID := int64(1)
	child := &Account{ParentAccountID: &parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	if (&OpenAITokenRefresher{}).CanRefresh(child) {
		t.Fatal("shadow must not be refreshed directly")
	}
}

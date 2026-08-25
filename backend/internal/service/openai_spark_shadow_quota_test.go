package service

import (
	"testing"
	"time"
)

func TestSparkShadowQuotaUsesOnlyBengalfoxWindows(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	updates := buildCodexSparkWindowExtraUpdates(&OpenAIQuotaUsage{AdditionalRateLimits: []OpenAIAdditionalRateLimit{{MeteredFeature: "codex_bengalfox", RateLimit: &OpenAIRateLimit{PrimaryWindow: &OpenAIRateLimitWindow{UsedPercent: 12, LimitWindowSeconds: 18000, ResetAfterSeconds: 60}}}}}, now)
	if updates["codex_5h_used_percent"] != 12.0 {
		t.Fatalf("unexpected spark update: %#v", updates)
	}
	if _, ok := updates["codex_spark_5h_used_percent"]; ok {
		t.Fatal("spark snapshot must use canonical codex prefix")
	}
}

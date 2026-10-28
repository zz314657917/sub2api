package service

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type upstreamACNFrequencyRepo struct {
	v027CNQuotaRepo
	extra map[string]any
}

func (r *upstreamACNFrequencyRepo) UpdateExtra(_ context.Context, _ int64, extra map[string]any) error {
	r.extra = extra
	return nil
}

func TestUpstreamAFixesCNFrequency429Scheduling(t *testing.T) {
	for _, platform := range []string{PlatformZhipu, PlatformKimi, PlatformDeepseek} {
		for _, tc := range []struct {
			name     string
			settings string
			seconds  int
		}{
			{"default", "", 5},
			{"configured", `{"enabled":true,"cooldown_seconds":17}`, 17},
			{"disabled", `{"enabled":false,"cooldown_seconds":17}`, 0},
		} {
			t.Run(platform+"/"+tc.name, func(t *testing.T) {
				account := newV027CNQuotaAccount(time.Now().Add(2*time.Hour), time.Now().Add(7*24*time.Hour))
				account.Platform = platform
				account.Extra = map[string]any{
					cnExtraKey(platform, cnExtraSuffix5hReset):    time.Now().Add(2 * time.Hour),
					cnExtraKey(platform, cnExtraSuffix5hUsed):     12.5,
					cnExtraKey(platform, cnExtraSuffixWeeklyUsed): 20,
				}
				repo := &upstreamACNFrequencyRepo{}
				blocker := &v027CNQuotaRuntimeBlocker{}
				svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
				svc.SetAccountRuntimeBlocker(blocker)
				if tc.settings != "" {
					svc.SetSettingService(NewSettingService(&gatewayTTLSettingRepo{data: map[string]string{
						SettingKeyRateLimit429CooldownSettings: tc.settings,
					}}, &config.Config{}))
				}
				before := time.Now()
				require.False(t, svc.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests,
					http.Header{}, []byte(`{"error":{"message":"您的账户已达到速率限制，请您控制请求频率"}}`)))
				after := time.Now()
				require.Zero(t, repo.errorCalls)
				require.Zero(t, repo.tempCalls)
				require.Empty(t, repo.extra)
				if tc.seconds == 0 {
					require.Zero(t, repo.rateCalls)
					require.Empty(t, blocker.accounts)
					return
				}
				require.Equal(t, 1, repo.rateCalls)
				require.False(t, repo.rateUntil.Before(before.Add(time.Duration(tc.seconds)*time.Second)))
				require.False(t, repo.rateUntil.After(after.Add(time.Duration(tc.seconds)*time.Second)))
				require.Equal(t, []*Account{account}, blocker.accounts)
				require.Equal(t, []time.Time{repo.rateUntil}, blocker.untils)
				require.Equal(t, []string{"429_fallback"}, blocker.reasons)
			})
		}
	}
}

func TestUpstreamAFixesCNFrequency429SnapshotGuards(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		short bool
	}{
		{"zero", 0, true}, {"integer", int64(12), true}, {"float32", float32(12.5), true},
		{"decimal", 89.99, true}, {"string", " 12.5 ", true},
		{"json_number", json.Number("12.5"), true},
		{"threshold", 90, false}, {"full", 100, false}, {"over", 100.1, false},
		{"negative", -1, false}, {"nan", math.NaN(), false}, {"inf", math.Inf(1), false},
		{"negative_inf", math.Inf(-1), false}, {"nan_string", "NaN", false},
		{"inf_string", "Inf", false}, {"malformed", "12oops", false},
		{"empty", "", false}, {"nil", nil, false}, {"bool", false, false},
		{"map", map[string]any{}, false}, {"invalid_number", json.Number("oops"), false},
		{"overflow_number", json.Number("1e1000"), false},
		{"hex_string", "0x1p+4", false}, {"underscore_string", "1_0", false},
		{"hex_number", json.Number("0x1p+4"), false},
	} {
		for _, suffix := range []string{cnExtraSuffix5hUsed, cnExtraSuffixWeeklyUsed} {
			t.Run(tc.name+"/"+suffix, func(t *testing.T) {
				until := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
				account := newV027CNQuotaAccount(until, time.Time{})
				// A valid companion window must never hide another invalid or exhausted window.
				account.Extra[cnExtraKey(account.Platform, cnExtraSuffix5hUsed)] = 10
				account.Extra[cnExtraKey(account.Platform, cnExtraSuffixWeeklyUsed)] = 15
				account.Extra[cnExtraKey(account.Platform, suffix)] = tc.value
				repo := &upstreamACNFrequencyRepo{}
				svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
				before := time.Now()
				svc.HandleUpstreamError(context.Background(), account, 429, http.Header{},
					[]byte(`{"error":{"message":"Too many requests"}}`))
				require.Equal(t, 1, repo.rateCalls)
				if tc.short {
					require.WithinDuration(t, before.Add(5*time.Second), repo.rateUntil, time.Second)
				} else {
					require.Equal(t, until, repo.rateUntil)
				}
				require.Zero(t, repo.errorCalls)
				require.Zero(t, repo.tempCalls)
			})
		}
	}
	for _, suffix := range []string{"missing", cnExtraSuffix5hUsed, cnExtraSuffixWeeklyUsed} {
		t.Run("single_window/"+suffix, func(t *testing.T) {
			until := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
			account := newV027CNQuotaAccount(until, time.Time{})
			if suffix != "missing" {
				account.Extra[cnExtraKey(account.Platform, suffix)] = 10
			}
			repo := &upstreamACNFrequencyRepo{}
			svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			before := time.Now()
			svc.HandleUpstreamError(context.Background(), account, 429, http.Header{}, []byte("rate limit exceeded"))
			require.Equal(t, 1, repo.rateCalls)
			if suffix == "missing" {
				require.Equal(t, until, repo.rateUntil)
			} else {
				require.WithinDuration(t, before.Add(5*time.Second), repo.rateUntil, time.Second)
			}
		})
	}
}

func TestUpstreamAFixesCNFrequency429WordingAndBalance(t *testing.T) {
	for _, message := range []string{
		"rate_limit_exceeded quota", "请求频率 额度", "访问频率 已用尽",
		"请求太频繁 用完", "too many requests usage limit", "rate limit exceeded insufficient",
		"too many requests balance", "请求过于频繁 余额", "rate limit exceeded credit",
		"请求频率 配额耗尽", "请求频率 已达限额", "请求频率 用量已耗尽",
		"额度不足", "unknown error",
	} {
		t.Run(message, func(t *testing.T) {
			until := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
			account := newV027CNQuotaAccount(until, time.Time{})
			account.Extra[cnExtraKey(account.Platform, cnExtraSuffix5hUsed)] = 10
			repo := &upstreamACNFrequencyRepo{}
			svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			svc.HandleUpstreamError(context.Background(), account, 429, http.Header{}, []byte(message))
			require.Equal(t, 1, repo.rateCalls)
			require.Equal(t, until, repo.rateUntil)
		})
	}
	for _, mode := range []string{AccountModeCoding, AccountModePayG} {
		t.Run("insufficient_balance/"+mode, func(t *testing.T) {
			account := newV027CNQuotaAccount(time.Now().Add(time.Hour), time.Time{})
			account.Credentials["account_mode"] = mode
			repo := &upstreamACNFrequencyRepo{}
			svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			before := time.Now()
			svc.HandleUpstreamError(context.Background(), account, 429, http.Header{},
				[]byte(`{"error":{"message":"请求频率 余额不足"}}`))
			require.Zero(t, repo.rateCalls)
			require.Zero(t, repo.errorCalls)
			require.Equal(t, 1, repo.tempCalls)
			require.Equal(t, true, repo.extra[cnExtraKey(account.Platform, cnBalanceExtraSuffixLow)])
			require.WithinDuration(t, before.Add(20*time.Minute), repo.tempUntil, time.Second)
		})
	}
}

func TestUpstreamAFixesCNFrequency429UnrelatedAccounts(t *testing.T) {
	for _, tc := range []struct{ platform, mode string }{
		{PlatformZhipu, AccountModePayG}, {PlatformOpenAI, AccountModeCoding},
	} {
		t.Run(tc.platform+"/"+tc.mode, func(t *testing.T) {
			account := newV027CNQuotaAccount(time.Now().Add(time.Hour), time.Time{})
			account.Platform = tc.platform
			account.Credentials["account_mode"] = tc.mode
			repo := &upstreamACNFrequencyRepo{}
			svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			before := time.Now()
			svc.HandleUpstreamError(context.Background(), account, 429, http.Header{}, []byte("too many requests"))
			require.Equal(t, 1, repo.rateCalls)
			require.WithinDuration(t, before.Add(5*time.Second), repo.rateUntil, time.Second)
			require.Zero(t, repo.errorCalls)
			require.Zero(t, repo.tempCalls)
		})
	}
}

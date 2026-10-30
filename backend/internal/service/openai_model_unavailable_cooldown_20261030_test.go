package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type modelUnavailableRepo struct {
	schedulerTestOpenAIAccountRepo
	keys     []string
	writeErr error
}

func (r *modelUnavailableRepo) SetModelRateLimit(_ context.Context, id int64, key string, until time.Time) error {
	r.keys = append(r.keys, key)
	if r.writeErr != nil {
		return r.writeErr
	}
	for i := range r.accounts {
		a := &r.accounts[i]
		if a.ID != id {
			continue
		}
		if a.Extra == nil {
			a.Extra = map[string]any{}
		}
		limits, _ := a.Extra[modelRateLimitsKey].(map[string]any)
		if limits == nil {
			limits = map[string]any{}
			a.Extra[modelRateLimitsKey] = limits
		}
		limits[key] = map[string]any{"rate_limit_reset_at": until.Format(time.RFC3339)}
	}
	return nil
}

func modelUnavailableAccount(id int64) Account {
	return Account{
		ID: id, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive,
		Schedulable: true, Concurrency: 1, Priority: int(id),
		Credentials: map[string]any{"api_key": "test", "access_token": "test", "base_url": "http://127.0.0.1:1"},
		Extra:       map[string]any{"openai_responses_supported": true},
	}
}

func TestModelUnavailableCooldownMappingAndSticky(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name, requested, upstream, typ string
		mapping                        map[string]any
		keys                           []string
	}{
		{"identity", "gpt-6-luna", "gpt-6-luna", AccountTypeAPIKey, nil, []string{"gpt-6-luna"}},
		{"alias", "alias", "upstream", AccountTypeAPIKey, map[string]any{"alias": "upstream", "*": "*"}, []string{"upstream"}},
		{"non-idempotent", "alias", "upstream", AccountTypeAPIKey, map[string]any{"alias": "upstream", "upstream": "another", "other": "other"}, []string{"upstream"}},
		{"oauth-normalized", "gpt-5.4-high", "gpt-5.4", AccountTypeOAuth, nil, []string{"gpt-5.4", "gpt-5.4-high"}},
		{"spark", "gpt-5.3-codex-spark-high", "gpt-5.3-codex-spark", AccountTypeOAuth, nil, []string{"gpt-5.3-codex-spark"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := modelUnavailableAccount(1), modelUnavailableAccount(2)
			a.Type = tc.typ
			if tc.mapping != nil {
				a.Credentials["model_mapping"] = tc.mapping
			}
			repo := &modelUnavailableRepo{schedulerTestOpenAIAccountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{a, b}}}
			cache := &schedulerTestGatewayCache{sessionBindings: map[string]int64{"openai:sticky": 1}}
			s := &OpenAIGatewayService{cfg: &config.Config{}, accountRepo: repo, cache: cache,
				rateLimitService: &RateLimitService{accountRepo: repo}}
			c, _ := modelUnavailableContext()
			rememberOpenAIModelUnavailableRequest(c, &a, []byte(`{"model":"`+tc.requested+`"}`))
			s.persistOpenAIModelUnavailable(ctx, c, &a, 404, tc.upstream)
			require.Equal(t, tc.keys, repo.keys)
			persisted, err := repo.GetByID(ctx, 1)
			require.NoError(t, err)
			require.False(t, persisted.IsSchedulableForModelWithContext(ctx, tc.requested))
			require.True(t, persisted.IsSchedulableForModelWithContext(ctx, "other"))
			require.True(t, b.IsSchedulableForModelWithContext(ctx, tc.requested))
			require.InDelta(t, 1800, persisted.GetModelRateLimitRemainingTime(tc.requested).Seconds(), 3)
			if tc.name == "non-idempotent" {
				require.Nil(t, persisted.modelRateLimitResetAt("another"))
				require.True(t, persisted.IsSchedulableForModelWithContext(ctx, "upstream"), "request upstream maps to another")
			}
			selected, err := s.selectAccountForModelWithExclusions(ctx, nil, PlatformOpenAI, "sticky",
				tc.requested, nil, false, 1, "", AccountCapabilityChat)
			require.NoError(t, err)
			require.EqualValues(t, 2, selected.ID)
			require.Equal(t, 1, cache.deletedSessions["openai:sticky"])
			selected, err = s.selectAccountForModelWithExclusions(ctx, nil, PlatformOpenAI, "",
				tc.requested, nil, false, 0, "", AccountCapabilityChat)
			require.NoError(t, err)
			require.EqualValues(t, 2, selected.ID)
		})
	}
}

func TestModelUnavailableWriteFailureStillSwitchesAndImageGuard(t *testing.T) {
	a := modelUnavailableAccount(1)
	repo := &modelUnavailableRepo{writeErr: errors.New("write failed")}
	s := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
	c, _ := modelUnavailableContext()
	require.True(t, s.handleOpenAIHTTPAccountUpstreamError(context.Background(), c, &a, 404, nil, []byte(modelUnavailable404), "gpt-6-luna"))
	require.Equal(t, []string{"gpt-6-luna"}, repo.keys)
	repo.keys = nil
	s.persistOpenAIModelUnavailable(context.Background(), c, &a, 404, "gpt-image-2.5-flare")
	require.Empty(t, repo.keys, "wrong text endpoint must not poison image availability")
	a.Type = AccountTypeOAuth
	rememberOpenAIModelUnavailableRequest(c, &a, []byte(`{"model":"gpt-image-2.5-flare"}`))
	s.persistOpenAIModelUnavailable(context.Background(), c, &a, 400, "gpt-5.4-mini")
	require.Equal(t, []string{"gpt-5.4-mini"}, repo.keys, "driver and image model are independent")
}

func TestOpenAIModelUnavailableSchedulerTopKCooldown(t *testing.T) {
	for _, tc := range []struct {
		name, requested, upstream string
		count, topK               int
		oauth                     bool
		mapping                   map[string]any
	}{
		{name: "topk-one", requested: "gpt-6-luna", upstream: "gpt-6-luna", count: 2, topK: 1},
		{name: "default-seven-eight-accounts", requested: "gpt-6-luna", upstream: "gpt-6-luna", count: 8},
		{name: "non-idempotent", requested: "alias", upstream: "gpt-6-luna", count: 2, topK: 1,
			mapping: map[string]any{"alias": "gpt-6-luna", "gpt-6-luna": "another", "other": "other"}},
		{name: "oauth-normalized", requested: "gpt-5.4-high", upstream: "gpt-5.4", count: 2, topK: 1, oauth: true},
		{name: "spark-normalized", requested: "gpt-5.3-codex-spark-high", upstream: "gpt-5.3-codex-spark", count: 2, topK: 1, oauth: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			repo := &modelUnavailableRepo{}
			for id := 1; id <= tc.count; id++ {
				account := modelUnavailableAccount(int64(id))
				if tc.oauth {
					account.Type = AccountTypeOAuth
				}
				if tc.mapping != nil {
					account.Credentials["model_mapping"] = tc.mapping
				}
				repo.accounts = append(repo.accounts, account)
			}
			cfg := &config.Config{}
			cfg.Gateway.OpenAIWS.LBTopK = tc.topK
			svc := &OpenAIGatewayService{cfg: cfg, accountRepo: repo,
				rateLimitService: &RateLimitService{accountRepo: repo}}
			if tc.topK == 0 {
				require.Equal(t, 7, svc.openAIWSLBTopK(), "retain production default")
			}
			// Persist through the unchanged HTTP write owner, then let the
			// advanced scheduler apply its existing request-model lookup.
			for i := 0; i < tc.count-1; i++ {
				account := &repo.accounts[i]
				c, _ := modelUnavailableContext()
				rememberOpenAIModelUnavailableRequest(c, account, []byte(`{"model":"`+tc.requested+`"}`))
				svc.persistOpenAIModelUnavailable(ctx, c, account, 404, tc.upstream)
				require.True(t, account.IsSchedulableWithContext(ctx))
				require.True(t, account.isModelRateLimitedWithContext(ctx, tc.requested))
			}
			scheduler := newDefaultOpenAIAccountScheduler(svc, nil)
			req := OpenAIAccountScheduleRequest{Platform: PlatformOpenAI, RequestedModel: tc.requested,
				RequiredTransport: OpenAIUpstreamTransportHTTPSSE, RequiredAccountCapability: AccountCapabilityChat}
			selected, decision, err := scheduler.Select(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, selected)
			require.EqualValues(t, tc.count, selected.Account.ID)
			require.Equal(t, 1, decision.CandidateCount, "cooled accounts must be filtered before ranking")
			require.Equal(t, 1, decision.TopK)
			require.Equal(t, tc.topK, cfg.Gateway.OpenAIWS.LBTopK, "never enlarge configured top-K")
			if selected.ReleaseFunc != nil {
				selected.ReleaseFunc()
			}
			if tc.mapping != nil {
				require.NotContains(t, repo.keys, "another")
			}
			req.RequestedModel = "other"
			if tc.oauth {
				req.RequestedModel = "gpt-5.1"
			}
			selected, decision, err = scheduler.Select(ctx, req)
			require.NoError(t, err)
			require.Equal(t, tc.count, decision.CandidateCount, "other models remain eligible on every account")
			require.Less(t, selected.Account.ID, int64(tc.count))
			if selected.ReleaseFunc != nil {
				selected.ReleaseFunc()
			}
		})
	}
}

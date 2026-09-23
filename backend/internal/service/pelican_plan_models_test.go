package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type pelicanModelsGroups struct {
	GroupRepository
	group *Group
	err   error
}

func (r pelicanModelsGroups) GetByIDLite(context.Context, int64) (*Group, error) {
	return r.group, r.err
}

type pelicanModelsAccounts struct {
	AccountRepository
	accounts []Account
	err      error
}

func (r pelicanModelsAccounts) ListByGroup(context.Context, int64) ([]Account, error) {
	return r.accounts, r.err
}

func pelicanCatalogAccount(mapping map[string]any) Account {
	return Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Credentials: map[string]any{"model_mapping": mapping}}
}

func TestPelicanPlanModelsUsesLocalConfiguredCatalog(t *testing.T) {
	group := &Group{ID: 9, Platform: PlatformOpenAI, Status: StatusActive}
	cases := []struct {
		name     string
		group    *Group
		accounts []Account
		want     []string // nil means assert contains instead of exact directory contents.
		contains []string
	}{
		{
			name:  "concrete mappings are sorted and image targets excluded",
			group: group,
			accounts: []Account{pelicanCatalogAccount(map[string]any{
				"gpt-5.2": "gpt-5.2-upstream", "gpt-image-2": "gpt-image-2", "poster": "gpt-image-2",
			})},
			want: []string{"gpt-5.2"},
		},
		{
			name:  "explicit group list is a whitelist",
			group: &Group{ID: 9, Platform: PlatformOpenAI, Status: StatusActive, ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-5.2"}}},
			accounts: []Account{pelicanCatalogAccount(map[string]any{
				"gpt-5.2": "gpt-5.2", "gpt-4.1": "gpt-4.1",
			})},
			want: []string{"gpt-5.2"},
		},
		{
			name:  "explicit group model is supported through wildcard mapping",
			group: &Group{ID: 9, Platform: PlatformOpenAI, Status: StatusActive, ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-5.9-preview"}}},
			accounts: []Account{pelicanCatalogAccount(map[string]any{
				"gpt-*": "gpt-upstream",
			})},
			want: []string{"gpt-5.9-preview"},
		},
		{
			name:     "unrestricted account contributes local defaults",
			group:    group,
			accounts: []Account{pelicanCatalogAccount(nil)},
			contains: []string{"gpt-6-astra", "gpt-5.6-sol"},
		},
		{
			name:     "non eligible account contributes no choices",
			group:    group,
			accounts: []Account{{Platform: PlatformOpenAI, Type: "agent_identity", Status: StatusActive, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-5.2": "gpt-5.2"}}}},
			want:     []string{},
		},
		{
			name:  "wildcard mapped to image contributes no html choice",
			group: group,
			accounts: []Account{pelicanCatalogAccount(map[string]any{
				"gpt-*": "gpt-image-2",
			})},
			want: []string{},
		},
		{
			name:     "enabled empty custom list keeps existing fallback semantics",
			group:    &Group{ID: 9, Platform: PlatformOpenAI, Status: StatusActive, ModelsListConfig: GroupModelsListConfig{Enabled: true}},
			accounts: []Account{pelicanCatalogAccount(nil)},
			contains: []string{"gpt-6-astra", "gpt-5.6-sol"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewPelicanTestService(nil, pelicanModelsAccounts{accounts: tc.accounts}, pelicanModelsGroups{group: tc.group}, nil)
			got, err := svc.Models(context.Background(), 9)
			if err != nil {
				t.Fatalf("models=%#v err=%v", got, err)
			}
			if tc.want != nil && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("models=%#v want=%#v", got, tc.want)
			}
			for _, expected := range tc.contains {
				if !containsPelicanModel(got, expected) {
					t.Fatalf("models=%#v missing %q", got, expected)
				}
			}
		})
	}
}

func TestPelicanPlanModelsPropagatesRepositoryErrorsAndValidatesSubmittedModel(t *testing.T) {
	group := &Group{ID: 9, Platform: PlatformOpenAI, Status: StatusActive}
	account := pelicanCatalogAccount(map[string]any{"gpt-5.2": "gpt-5.2"})
	svc := NewPelicanTestService(nil, pelicanModelsAccounts{err: errors.New("accounts unavailable")}, pelicanModelsGroups{group: group}, nil)
	if _, err := svc.Models(context.Background(), 9); err == nil {
		t.Fatal("expected account repository error")
	}
	svc = NewPelicanTestService(nil, pelicanModelsAccounts{accounts: []Account{account}}, pelicanModelsGroups{group: group}, nil)
	if _, err := svc.validPlanModel(context.Background(), 9, "gpt-5.2"); err != nil {
		t.Fatalf("configured model rejected: %v", err)
	}
	if _, err := svc.validPlanModel(context.Background(), 9, "gpt-4.1"); !errors.Is(err, ErrPelicanInvalid) {
		t.Fatalf("unconfigured model err=%v", err)
	}
}

func TestPelicanPlanModelsSupportsNonOpenAIGroupCatalogs(t *testing.T) {
	group := &Group{ID: 11, Platform: PlatformGemini, Status: StatusActive}
	account := Account{ID: 11, Platform: PlatformGemini, Type: AccountTypeAPIKey, Status: StatusActive, Credentials: map[string]any{
		"model_mapping": map[string]any{"gemini-2.0-flash": "gemini-2.0-flash"},
	}}
	svc := NewPelicanTestService(nil, pelicanModelsAccounts{accounts: []Account{account}}, pelicanModelsGroups{group: group}, nil)
	models, err := svc.Models(context.Background(), group.ID)
	if err != nil {
		t.Fatalf("models err: %v", err)
	}
	if !containsPelicanModel(models, "gemini-2.0-flash") {
		t.Fatalf("models=%#v missing Gemini model", models)
	}
}

type pelicanLeaseCache struct {
	ConcurrencyCache
	acquire    bool
	acquireErr error
	acquires   int
	releases   int
}

type pelicanLeaseAccounts struct {
	pelicanModelsAccounts
	account *Account
}

func (r pelicanLeaseAccounts) GetByID(context.Context, int64) (*Account, error) {
	return r.account, nil
}

type pelicanLeaseRepo struct {
	PelicanTestRepository
	reserved int
	saved    int
}

func (*pelicanLeaseRepo) Finalize(context.Context, int64, int64) error { return nil }
func (*pelicanLeaseRepo) CanRunAccount(context.Context, int64, int64, int64) (bool, error) {
	return true, nil
}
func (r *pelicanLeaseRepo) ReserveAttempt(context.Context, int64, int64) error {
	r.reserved++
	return nil
}
func (r *pelicanLeaseRepo) SaveResult(context.Context, *PelicanResult, int, int64) error {
	r.saved++
	return nil
}

type pelicanLeaseExecutor func(context.Context, int64, string, string, string) (*PelicanResult, error)

func (f pelicanLeaseExecutor) RunPelicanTest(ctx context.Context, accountID int64, model, prompt, effort string) (*PelicanResult, error) {
	return f(ctx, accountID, model, prompt, effort)
}

func (c *pelicanLeaseCache) AcquireAccountSlot(context.Context, int64, int, string) (bool, error) {
	c.acquires++
	return c.acquire, c.acquireErr
}

func (c *pelicanLeaseCache) ReleaseAccountSlot(context.Context, int64, string) error {
	c.releases++
	return nil
}

func TestPelicanMixedSelectorNonOpenAIAccountLease(t *testing.T) {
	group := &Group{ID: 23, Platform: PlatformGemini, Status: StatusActive}
	account := Account{ID: 24, Platform: PlatformGemini, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, Credentials: map[string]any{
		"model_mapping": map[string]any{"gemini-2.0-flash": "gemini-2.0-flash"},
	}}

	t.Run("acquires and releases shared slot", func(t *testing.T) {
		cache := &pelicanLeaseCache{acquire: true}
		selector := pelicanMixedSelector{
			accounts:    pelicanModelsAccounts{accounts: []Account{account}},
			groups:      pelicanModelsGroups{group: group},
			concurrency: NewConcurrencyService(cache),
		}
		selection, err := selector.SelectPelicanAccount(context.Background(), group.ID, "gemini-2.0-flash")
		if err != nil || selection == nil || !selection.Acquired || selection.ReleaseFunc == nil {
			t.Fatalf("selection=%#v err=%v", selection, err)
		}
		selection.ReleaseFunc()
		if cache.acquires != 1 || cache.releases != 1 {
			t.Fatalf("acquires=%d releases=%d", cache.acquires, cache.releases)
		}
	})

	t.Run("busy slot is not acquired", func(t *testing.T) {
		cache := &pelicanLeaseCache{}
		selector := pelicanMixedSelector{
			accounts:    pelicanModelsAccounts{accounts: []Account{account}},
			groups:      pelicanModelsGroups{group: group},
			concurrency: NewConcurrencyService(cache),
		}
		selection, err := selector.SelectPelicanAccount(context.Background(), group.ID, "gemini-2.0-flash")
		if err != nil || selection == nil || selection.Acquired || selection.ReleaseFunc != nil {
			t.Fatalf("selection=%#v err=%v", selection, err)
		}
		if cache.acquires != 1 || cache.releases != 0 {
			t.Fatalf("acquires=%d releases=%d", cache.acquires, cache.releases)
		}
	})

	t.Run("cache error fails closed", func(t *testing.T) {
		cache := &pelicanLeaseCache{acquireErr: errors.New("cache unavailable")}
		selector := pelicanMixedSelector{
			accounts:    pelicanModelsAccounts{accounts: []Account{account}},
			groups:      pelicanModelsGroups{group: group},
			concurrency: NewConcurrencyService(cache),
		}
		selection, err := selector.SelectPelicanAccount(context.Background(), group.ID, "gemini-2.0-flash")
		if selection != nil || err == nil || err.Error() != "cache unavailable" {
			t.Fatalf("selection=%#v err=%v", selection, err)
		}
		if cache.acquires != 1 || cache.releases != 0 {
			t.Fatalf("acquires=%d releases=%d", cache.acquires, cache.releases)
		}
	})

	t.Run("ineligible account does not reserve a slot", func(t *testing.T) {
		cache := &pelicanLeaseCache{acquire: true}
		ineligible := account
		ineligible.Schedulable = false
		selector := pelicanMixedSelector{
			accounts:    pelicanModelsAccounts{accounts: []Account{ineligible}},
			groups:      pelicanModelsGroups{group: group},
			concurrency: NewConcurrencyService(cache),
		}
		selection, err := selector.SelectPelicanAccount(context.Background(), group.ID, "gemini-2.0-flash")
		if selection != nil || err != nil {
			t.Fatalf("selection=%#v err=%v", selection, err)
		}
		if cache.acquires != 0 || cache.releases != 0 {
			t.Fatalf("acquires=%d releases=%d", cache.acquires, cache.releases)
		}
	})
}

func TestPelicanMixedSelectorNonOpenAILeaseReleasedExactlyOnce(t *testing.T) {
	group := &Group{ID: 25, Platform: PlatformGemini, Status: StatusActive}
	account := &Account{ID: 26, Platform: PlatformGemini, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []int64{group.ID}, Credentials: map[string]any{
		"model_mapping": map[string]any{"gemini-2.0-flash": "gemini-2.0-flash"},
	}}
	newService := func(cache *pelicanLeaseCache, executor pelicanLeaseExecutor) (*PelicanTestService, *pelicanLeaseRepo) {
		accounts := pelicanLeaseAccounts{pelicanModelsAccounts: pelicanModelsAccounts{accounts: []Account{*account}}, account: account}
		repo := &pelicanLeaseRepo{}
		svc := NewPelicanTestService(repo, accounts, pelicanModelsGroups{group: group}, executor)
		svc.SetAccountSelector(pelicanMixedSelector{accounts: accounts, groups: pelicanModelsGroups{group: group}, concurrency: NewConcurrencyService(cache)})
		return svc, repo
	}
	run := func(svc *PelicanTestService, ctx context.Context) <-chan struct{} {
		done := make(chan struct{})
		svc.planGate <- struct{}{}
		go func() {
			defer close(done)
			svc.run(ctx, &PelicanPlan{ID: 1, GroupID: group.ID, ModelID: "gemini-2.0-flash", RunGeneration: 1, TimeoutSeconds: 1}, pelicanPromptSnapshot{prompt: "test"})
		}()
		return done
	}
	assertLifecycle := func(t *testing.T, cache *pelicanLeaseCache, repo *pelicanLeaseRepo, providerCalls int) {
		t.Helper()
		if cache.acquires != 1 || cache.releases != 1 || repo.reserved != 1 || providerCalls != 1 {
			t.Fatalf("acquires=%d releases=%d reserved=%d provider_calls=%d", cache.acquires, cache.releases, repo.reserved, providerCalls)
		}
	}

	for _, tc := range []struct {
		name string
		run  func(*testing.T, *PelicanTestService) int
	}{
		{
			name: "success",
			run: func(t *testing.T, svc *PelicanTestService) int {
				calls := 0
				svc.tester = pelicanLeaseExecutor(func(context.Context, int64, string, string, string) (*PelicanResult, error) {
					calls++
					return &PelicanResult{HTML: "<html>ok</html>"}, nil
				})
				<-run(svc, context.Background())
				return calls
			},
		},
		{
			name: "provider failure",
			run: func(t *testing.T, svc *PelicanTestService) int {
				calls := 0
				svc.tester = pelicanLeaseExecutor(func(context.Context, int64, string, string, string) (*PelicanResult, error) {
					calls++
					return nil, errors.New("upstream failed")
				})
				<-run(svc, context.Background())
				return calls
			},
		},
		{
			name: "provider in-flight cancellation",
			run: func(t *testing.T, svc *PelicanTestService) int {
				calls := 0
				started := make(chan struct{})
				svc.tester = pelicanLeaseExecutor(func(ctx context.Context, _ int64, _, _, _ string) (*PelicanResult, error) {
					calls++
					close(started)
					<-ctx.Done()
					return nil, ctx.Err()
				})
				ctx, cancel := context.WithCancel(context.Background())
				done := run(svc, ctx)
				select {
				case <-started:
					cancel()
				case <-time.After(time.Second):
					t.Fatal("provider did not start")
				}
				select {
				case <-done:
				case <-time.After(time.Second):
					t.Fatal("cancelled provider did not finish")
				}
				return calls
			},
		},
		{
			name: "provider in-flight timeout",
			run: func(t *testing.T, svc *PelicanTestService) int {
				calls := 0
				started := make(chan struct{})
				svc.tester = pelicanLeaseExecutor(func(ctx context.Context, _ int64, _, _, _ string) (*PelicanResult, error) {
					calls++
					close(started)
					<-ctx.Done()
					return nil, ctx.Err()
				})
				done := run(svc, context.Background())
				select {
				case <-started:
				case <-time.After(time.Second):
					t.Fatal("provider did not start")
				}
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("timed out provider did not finish")
				}
				return calls
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := &pelicanLeaseCache{acquire: true}
			svc, repo := newService(cache, nil)
			assertLifecycle(t, cache, repo, tc.run(t, svc))
		})
	}

	for _, tc := range []struct {
		name       string
		cache      *pelicanLeaseCache
		candidate  *Account
		acquires   int
	}{
		{name: "busy slot", cache: &pelicanLeaseCache{}, candidate: account, acquires: 1},
		{name: "ineligible account", cache: &pelicanLeaseCache{acquire: true}, candidate: func() *Account { copy := *account; copy.Schedulable = false; return &copy }(), acquires: 0},
	} {
		t.Run(tc.name+" does not reserve or call provider", func(t *testing.T) {
			calls := 0
			svc, repo := newService(tc.cache, pelicanLeaseExecutor(func(context.Context, int64, string, string, string) (*PelicanResult, error) {
				calls++
				return nil, nil
			}))
			accounts := svc.accounts.(pelicanLeaseAccounts)
			accounts.account = tc.candidate
			accounts.accounts = []Account{*tc.candidate}
			svc.accounts = accounts
			svc.selector = pelicanMixedSelector{accounts: accounts, groups: pelicanModelsGroups{group: group}, concurrency: NewConcurrencyService(tc.cache)}
			<-run(svc, context.Background())
			if tc.cache.acquires != tc.acquires || tc.cache.releases != 0 || repo.reserved != 0 || calls != 0 {
				t.Fatalf("acquires=%d releases=%d reserved=%d provider_calls=%d", tc.cache.acquires, tc.cache.releases, repo.reserved, calls)
			}
		})
	}
}

func TestPelicanPlatformAccountTypeSupportMatrix(t *testing.T) {
	for _, tc := range []struct {
		platform string
		account  string
		want     bool
	}{
		{PlatformOpenAI, AccountTypeAPIKey, true},
		{PlatformOpenAI, AccountTypeOAuth, true},
		{PlatformOpenAI, AccountTypeSetupToken, true},
		{PlatformOpenAI, AccountTypeServiceAccount, false},
		{PlatformAnthropic, AccountTypeAPIKey, true},
		{PlatformAnthropic, AccountTypeOAuth, true},
		{PlatformAnthropic, AccountTypeSetupToken, true},
		{PlatformAnthropic, AccountTypeBedrock, true},
		{PlatformAnthropic, AccountTypeServiceAccount, true},
		{PlatformGemini, AccountTypeAPIKey, true},
		{PlatformGemini, AccountTypeOAuth, true},
		{PlatformGemini, AccountTypeServiceAccount, true},
		{PlatformGemini, AccountTypeSetupToken, false},
		{PlatformGrok, AccountTypeOAuth, true},
		{PlatformGrok, AccountTypeAPIKey, false},
		{PlatformAntigravity, AccountTypeAPIKey, true},
		{PlatformAntigravity, AccountTypeOAuth, true},
		{PlatformAntigravity, AccountTypeServiceAccount, false},
		{PlatformKimi, AccountTypeAPIKey, true},
		{PlatformKimi, AccountTypeOAuth, false},
		{PlatformZhipu, AccountTypeAPIKey, true},
		{PlatformZhipu, AccountTypeOAuth, false},
		{PlatformDeepseek, AccountTypeAPIKey, true},
		{PlatformDeepseek, AccountTypeOAuth, false},
	} {
		t.Run(tc.platform+"/"+tc.account, func(t *testing.T) {
			if got := pelicanPlatformAccountTypeSupported(tc.platform, tc.account); got != tc.want {
				t.Fatalf("supported=%t want=%t", got, tc.want)
			}
		})
	}
}

func TestPelicanRunnerRejectsUnsupportedPlatformTypesBeforeLease(t *testing.T) {
	for _, tc := range []struct {
		platform string
		account  string
	}{
		{PlatformOpenAI, AccountTypeServiceAccount},
		{PlatformAnthropic, "unsupported"},
		{PlatformGemini, AccountTypeSetupToken},
		{PlatformGrok, AccountTypeAPIKey},
		{PlatformAntigravity, AccountTypeServiceAccount},
		{PlatformKimi, AccountTypeOAuth},
		{PlatformZhipu, AccountTypeOAuth},
		{PlatformDeepseek, AccountTypeOAuth},
	} {
		t.Run(tc.platform+"/"+tc.account, func(t *testing.T) {
			group := &Group{ID: 31, Platform: tc.platform, Status: StatusActive}
			account := &Account{ID: 32, Platform: tc.platform, Type: tc.account, Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []int64{group.ID}, Credentials: map[string]any{
				"model_mapping": map[string]any{"pelican-test": "pelican-test"},
			}}
			cache := &pelicanLeaseCache{acquire: true}
			accounts := pelicanLeaseAccounts{pelicanModelsAccounts: pelicanModelsAccounts{accounts: []Account{*account}}, account: account}
			repo := &pelicanLeaseRepo{}
			providerCalls := 0
			svc := NewPelicanTestService(repo, accounts, pelicanModelsGroups{group: group}, pelicanLeaseExecutor(func(context.Context, int64, string, string, string) (*PelicanResult, error) {
				providerCalls++
				return nil, nil
			}))
			svc.SetAccountSelector(pelicanMixedSelector{accounts: accounts, groups: pelicanModelsGroups{group: group}, concurrency: NewConcurrencyService(cache)})
			svc.planGate <- struct{}{}
			svc.run(context.Background(), &PelicanPlan{ID: 1, GroupID: group.ID, ModelID: "pelican-test", RunGeneration: 1, TimeoutSeconds: 1}, pelicanPromptSnapshot{prompt: "test"})
			if cache.acquires != 0 || cache.releases != 0 || repo.reserved != 0 || providerCalls != 0 {
				t.Fatalf("acquires=%d releases=%d reserved=%d provider_calls=%d", cache.acquires, cache.releases, repo.reserved, providerCalls)
			}
		})
	}
}

func TestPelicanPlanModelsProvidesCNProviderDefaultsWithoutMapping(t *testing.T) {
	for _, tc := range []struct {
		platform string
		model    string
	}{
		{PlatformKimi, "kimi-k2.6"},
		{PlatformZhipu, "glm-5.3"},
		{PlatformDeepseek, "deepseek-chat"},
	} {
		t.Run(tc.platform, func(t *testing.T) {
			group := &Group{ID: 12, Platform: tc.platform, Status: StatusActive}
			account := Account{ID: 12, Platform: tc.platform, Type: AccountTypeAPIKey, Status: StatusActive}
			svc := NewPelicanTestService(nil, pelicanModelsAccounts{accounts: []Account{account}}, pelicanModelsGroups{group: group}, nil)
			models, err := svc.Models(context.Background(), group.ID)
			if err != nil {
				t.Fatalf("models err: %v", err)
			}
			if !containsPelicanModel(models, tc.model) {
				t.Fatalf("models=%#v missing default %q", models, tc.model)
			}
		})
	}
}

func TestPelicanReasoningEffortValidation(t *testing.T) {
	base := PelicanPlanInput{GroupID: 9, ModelID: "gpt-5.2", IntervalMinutes: 15, MaxResults: 1, MinChars: 100}
	for _, effort := range []string{"", "none", "minimal", "low", "medium", "high", "xhigh"} {
		input := base
		input.ReasoningEffort = effort
		if err := validPelicanInput(input); err != nil {
			t.Fatalf("effort %q rejected: %v", effort, err)
		}
	}
	for _, effort := range []string{" default", "ultra", "HIGH", "medium\n"} {
		input := base
		input.ReasoningEffort = effort
		if !errors.Is(validPelicanInput(input), ErrPelicanInvalid) {
			t.Fatalf("invalid effort %q accepted", effort)
		}
	}
}

func containsPelicanModel(models []string, expected string) bool {
	for _, model := range models {
		if model == expected {
			return true
		}
	}
	return false
}

func TestPelicanSkipReasonClassifiesCurrentState(t *testing.T) {
	group := &Group{ID: 7, Platform: PlatformOpenAI, Status: StatusActive}
	now := time.Now()
	for _, tc := range []struct {
		name, want string
		account    *Account
	}{
		{"available", "", &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}},
		{"disabled", "account_scheduling_disabled", &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive}},
		{"inactive", "account_inactive", &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusDisabled, Schedulable: true}},
		{"expired", "account_expired", &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, ExpiresAt: ptrPelicanTime(now.Add(-time.Minute))}},
		{"rate limited", "account_rate_limited", &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, RateLimitResetAt: ptrPelicanTime(now.Add(time.Minute))}},
		{"cooldown", "account_cooldown", &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, OverloadUntil: ptrPelicanTime(now.Add(time.Minute))}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := pelicanSkipReason(tc.account, nil, group, nil, true); got != tc.want {
				t.Fatalf("reason=%q want=%q", got, tc.want)
			}
		})
	}
	if got := pelicanSkipReason(nil, errors.New("db detail"), group, nil, true); got != "account_lookup_failed" {
		t.Fatalf("reason=%q", got)
	}
	if got := pelicanSkipReason(nil, nil, nil, nil, true); got != "account_lookup_failed" {
		t.Fatalf("reason=%q", got)
	}
	if got := pelicanSkipReason(&Account{}, nil, nil, nil, true); got != "group_unavailable" {
		t.Fatalf("reason=%q", got)
	}
}

func ptrPelicanTime(value time.Time) *time.Time { return &value }

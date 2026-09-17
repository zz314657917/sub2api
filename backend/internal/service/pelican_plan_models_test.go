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

package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type pelicanPreferredSelector struct {
	pelicanSelectorFunc
	selectPreferred func(context.Context, int64, string, []int64, map[int64]struct{}) (*AccountSelectionResult, bool, error)
}

func (s pelicanPreferredSelector) SelectPreferredPelicanAccount(ctx context.Context, group int64, model string, ids []int64, excluded map[int64]struct{}) (*AccountSelectionResult, bool, error) {
	return s.selectPreferred(ctx, group, model, ids, excluded)
}

func preferredPelicanAccount(id int64) Account {
	return Account{ID: id, Name: "test", GroupIDs: []int64{7}, Platform: PlatformOpenAI,
		Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1}
}

type pelicanPreferredCache struct {
	ConcurrencyCache
	busy     map[int64]bool
	acquired []int64
	released []int64
	err      error
}

func (c *pelicanPreferredCache) AcquireAccountSlot(_ context.Context, id int64, _ int, _ string) (bool, error) {
	c.acquired = append(c.acquired, id)
	if c.err != nil {
		return false, c.err
	}
	return !c.busy[id], nil
}
func (c *pelicanPreferredCache) ReleaseAccountSlot(_ context.Context, id int64, _ string) error {
	c.released = append(c.released, id)
	return nil
}

type pelicanPreferredLookup struct {
	pelicanReviewAccounts
	err     error
	lookups []int64
	lists   int
}

func (r *pelicanPreferredLookup) GetByID(ctx context.Context, id int64) (*Account, error) {
	r.lookups = append(r.lookups, id)
	if id == 9 && r.err != nil {
		return nil, r.err
	}
	return r.pelicanReviewAccounts.GetByID(ctx, id)
}
func (r *pelicanPreferredLookup) ListByGroup(ctx context.Context, id int64) ([]Account, error) {
	r.lists++
	return r.pelicanReviewAccounts.ListByGroup(ctx, id)
}

func TestPelicanPreferredSelectorInfrastructureErrorsTerminate(t *testing.T) {
	for _, scenario := range []string{"lookup_error", "slot_error", "not_found"} {
		t.Run(scenario, func(t *testing.T) {
			a, b := preferredPelicanAccount(9), preferredPelicanAccount(2)
			a.Platform, b.Platform = PlatformGemini, PlatformGemini
			sentinel := errors.New("infrastructure unavailable")
			accounts := &pelicanPreferredLookup{pelicanReviewAccounts: pelicanReviewAccounts{
				accounts: []Account{a, b}, fresh: map[int64]*Account{9: &a, 2: &b},
			}}
			cache := &pelicanPreferredCache{}
			switch scenario {
			case "lookup_error":
				accounts.err = sentinel
			case "slot_error":
				cache.err = sentinel
			case "not_found":
				accounts.err = fmt.Errorf("lookup: %w", ErrAccountNotFound)
			}
			selector := pelicanMixedSelector{accounts: accounts, groups: pelicanModelsGroups{group: &Group{ID: 7, Platform: PlatformGemini, Status: StatusActive}}, concurrency: NewConcurrencyService(cache)}
			selected, auto, err := selector.SelectPreferredPelicanAccount(context.Background(), 7, "gemini-2.0-flash", []int64{9, 2}, map[int64]struct{}{})
			if scenario == "not_found" {
				if err != nil || auto || selected == nil || selected.Account.ID != 2 || !reflect.DeepEqual(accounts.lookups, []int64{9, 2}) {
					t.Fatalf("selected=%v auto=%v err=%v lookups=%v", selected, auto, err, accounts.lookups)
				}
				releasePelicanSelection(selected)
			} else {
				if !errors.Is(err, sentinel) || selected != nil || auto || !reflect.DeepEqual(accounts.lookups, []int64{9}) {
					t.Fatalf("selected=%v auto=%v err=%v lookups=%v", selected, auto, err, accounts.lookups)
				}
				if len(cache.released) != 0 {
					t.Fatalf("released an unacquired slot: %v", cache.released)
				}
			}
			if accounts.lists != 0 {
				t.Fatal("infrastructure error silently fell back to automatic selection")
			}
		})
	}
}

func TestPelicanPreferredSelectorOrderAndFallback(t *testing.T) {
	for _, scenario := range []string{"order", "busy", "unavailable", "fallback", "excluded"} {
		t.Run(scenario, func(t *testing.T) {
			a, b, fallback := preferredPelicanAccount(9), preferredPelicanAccount(2), preferredPelicanAccount(8)
			group := &Group{ID: 7, Platform: PlatformGemini, Status: StatusActive}
			a.Platform, b.Platform, fallback.Platform = PlatformGemini, PlatformGemini, PlatformGemini
			cache := &pelicanPreferredCache{busy: map[int64]bool{}}
			excluded := map[int64]struct{}{}
			want := int64(9)
			auto := false
			switch scenario {
			case "busy":
				cache.busy[9] = true
				want = 2
			case "unavailable":
				a.Schedulable = false
				want = 2
			case "fallback":
				a.Schedulable, b.Schedulable = false, false
				want, auto = 8, true
			case "excluded":
				excluded[9] = struct{}{}
				want = 2
			}
			accounts := pelicanReviewAccounts{accounts: []Account{a, b, fallback}, fresh: map[int64]*Account{9: &a, 2: &b, 8: &fallback}}
			selector := pelicanMixedSelector{accounts: accounts, groups: pelicanModelsGroups{group: group}, concurrency: NewConcurrencyService(cache)}
			selected, automatic, err := selector.SelectPreferredPelicanAccount(context.Background(), 7, "gemini-2.0-flash", []int64{9, 2}, excluded)
			if err != nil || selected == nil || selected.Account.ID != want || !selected.Acquired || automatic != auto {
				t.Fatalf("selected=%+v auto=%v err=%v", selected, automatic, err)
			}
			releasePelicanSelection(selected)
			releasePelicanSelection(selected)
			if !reflect.DeepEqual(cache.released, []int64{want}) {
				t.Fatalf("releases=%v", cache.released)
			}
			if scenario == "busy" {
				if !reflect.DeepEqual(cache.acquired, []int64{9, 2}) {
					t.Fatalf("acquires=%v", cache.acquired)
				}
				if _, ok := excluded[9]; !ok {
					t.Fatal("busy candidate not excluded")
				}
			}
		})
	}
}

type pelicanPreferredRepo struct {
	pelicanReviewRepo
	reserveErr error
}

func (r *pelicanPreferredRepo) ReserveAttempt(ctx context.Context, plan, generation int64) error {
	r.reserved++
	return r.reserveErr
}

func TestPelicanPreferredRunnerLeaseAndTerminalBoundaries(t *testing.T) {
	for _, scenario := range []string{"success", "revalidation", "auto_invalid", "auto_excluded", "selection_error", "lookup_error", "generation", "quota", "reservation", "provider", "cancel", "group"} {
		t.Run(scenario, func(t *testing.T) {
			a, b := preferredPelicanAccount(9), preferredPelicanAccount(2)
			repo := &pelicanPreferredRepo{}
			accounts := pelicanReviewAccounts{fresh: map[int64]*Account{9: &a, 2: &b}}
			group := &Group{ID: 7, Platform: PlatformOpenAI, Status: StatusActive}
			switch scenario {
			case "revalidation", "auto_invalid":
				a.Schedulable = false
			case "generation":
				repo.canRun = map[int64]bool{9: false}
			case "quota":
				repo.reserveErr = ErrPelicanQuota
			case "reservation":
				repo.reserveErr = errors.New("reservation failed")
			case "group":
				group.Status = "inactive"
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			selections, released, calls := 0, 0, 0
			selector := pelicanPreferredSelector{selectPreferred: func(_ context.Context, _ int64, _ string, ids []int64, excluded map[int64]struct{}) (*AccountSelectionResult, bool, error) {
				selections++
				if !reflect.DeepEqual(ids, []int64{9, 2}) {
					t.Fatalf("order=%v", ids)
				}
				account := &a
				if selections == 2 {
					if released != 1 {
						t.Fatal("rejected lease not released before reselection")
					}
					if _, ok := excluded[9]; !ok {
						t.Fatal("rejected candidate not excluded")
					}
					account = &b
				}
				if scenario == "auto_excluded" {
					excluded[9] = struct{}{}
				}
				if scenario == "cancel" {
					cancel()
				}
				result := &AccountSelectionResult{Account: account, Acquired: true, ReleaseFunc: func() { released++ }}
				if scenario == "selection_error" {
					return result, false, errors.New("failed after acquire")
				}
				return result, scenario == "auto_invalid" || scenario == "auto_excluded", nil
			}}
			lookup := &pelicanPreferredLookup{pelicanReviewAccounts: accounts}
			if scenario == "lookup_error" {
				lookup.err = errors.New("repository unavailable")
			}
			svc := NewPelicanTestService(repo, lookup, pelicanModelsGroups{group: group},
				pelicanLeaseExecutor(func(_ context.Context, id int64, _, _, _ string) (*PelicanResult, error) {
					calls++
					if scenario == "revalidation" && id != 2 {
						t.Fatalf("provider account=%d", id)
					}
					if scenario == "provider" {
						return nil, ErrPelicanUpstreamRequest
					}
					return &PelicanResult{HTML: "<!DOCTYPE html><html></html>"}, nil
				}))
			svc.selector = selector
			keepRunning := svc.runModel(ctx, &PelicanPlan{ID: 1, GroupID: 7, ModelID: "gpt-test", PreferredAccountIDs: []int64{9, 2}}, pelicanPromptSnapshot{})
			wantSelect, wantCalls := 1, 0
			if scenario == "revalidation" {
				wantSelect, wantCalls = 2, 1
			}
			if scenario == "success" || scenario == "provider" {
				wantCalls = 1
			}
			if selections != wantSelect || released != wantSelect || calls != wantCalls || len(repo.saved) != 1 {
				t.Fatalf("select=%d releases=%d calls=%d results=%d", selections, released, calls, len(repo.saved))
			}
			if keepRunning != (scenario != "quota" && scenario != "reservation") {
				t.Fatalf("keepRunning=%v", keepRunning)
			}
		})
	}
}

func TestPelicanPreferredLegacyCompatibility(t *testing.T) {
	for _, preferred := range [][]int64{nil, {9}} {
		a := preferredPelicanAccount(9)
		repo := &pelicanPreferredRepo{}
		legacy, calls, releases := 0, 0, 0
		svc := NewPelicanTestService(repo, pelicanReviewAccounts{fresh: map[int64]*Account{9: &a}}, pelicanReviewGroups{},
			pelicanLeaseExecutor(func(context.Context, int64, string, string, string) (*PelicanResult, error) {
				calls++
				return &PelicanResult{}, nil
			}))
		svc.selector = pelicanSelectorFunc(func(context.Context, int64, string) (*AccountSelectionResult, error) {
			legacy++
			return &AccountSelectionResult{Account: &a, Acquired: true, ReleaseFunc: func() { releases++ }}, nil
		})
		svc.runModel(context.Background(), &PelicanPlan{GroupID: 7, ModelID: "gpt-test", PreferredAccountIDs: preferred}, pelicanPromptSnapshot{})
		want := 1
		if len(preferred) > 0 {
			want = 0
		}
		if legacy != want || calls != want || releases != want {
			t.Fatalf("preferred=%v legacy=%d calls=%d releases=%d", preferred, legacy, calls, releases)
		}
	}
}

func TestPelicanPreferredReselectionBound(t *testing.T) {
	ids := make([]int64, 50)
	fresh := make(map[int64]*Account)
	for index := range ids {
		ids[index] = int64(index + 1)
		account := preferredPelicanAccount(ids[index])
		account.Schedulable = false
		fresh[account.ID] = &account
	}
	fallback := preferredPelicanAccount(99)
	fallback.Schedulable = false
	fresh[99] = &fallback
	repo := &pelicanPreferredRepo{}
	selections, releases := 0, 0
	svc := NewPelicanTestService(repo, pelicanReviewAccounts{fresh: fresh}, pelicanReviewGroups{}, nil)
	svc.selector = pelicanPreferredSelector{selectPreferred: func(_ context.Context, _ int64, _ string, _ []int64, excluded map[int64]struct{}) (*AccountSelectionResult, bool, error) {
		if releases != selections || len(excluded) != selections {
			t.Fatalf("leases accumulated: selected=%d released=%d excluded=%d", selections, releases, len(excluded))
		}
		selections++
		id := int64(selections)
		if selections == 51 {
			id = 99
		}
		return &AccountSelectionResult{Account: fresh[id], Acquired: true, ReleaseFunc: func() { releases++ }}, selections == 51, nil
	}}
	svc.runModel(context.Background(), &PelicanPlan{GroupID: 7, ModelID: "gpt-test", PreferredAccountIDs: ids}, pelicanPromptSnapshot{})
	if selections != 51 || releases != 51 || len(repo.saved) != 1 || repo.reserved != 0 {
		t.Fatalf("selections=%d releases=%d results=%d reserved=%d", selections, releases, len(repo.saved), repo.reserved)
	}
}

type pelicanPreferredPlanWriter struct {
	PelicanTestRepository
	inputs []PelicanPlanInput
}

func (r *pelicanPreferredPlanWriter) CreatePlan(_ context.Context, input PelicanPlanInput, _ string) (*PelicanPlan, error) {
	r.inputs = append(r.inputs, input)
	return &PelicanPlan{PreferredAccountIDs: input.PreferredAccountIDs}, nil
}
func (r *pelicanPreferredPlanWriter) UpdatePlan(ctx context.Context, _ int64, input PelicanPlanInput, group string) (*PelicanPlan, error) {
	return r.CreatePlan(ctx, input, group)
}

func TestPelicanPreferredPlanWritesValidateAndKeepOrder(t *testing.T) {
	a, b := preferredPelicanAccount(9), preferredPelicanAccount(2)
	foreign := preferredPelicanAccount(8)
	foreign.GroupIDs = []int64{99}
	repo := &pelicanPreferredPlanWriter{}
	svc := NewPelicanTestService(repo, pelicanModelsAccounts{accounts: []Account{a, b, foreign}}, pelicanReviewGroups{}, nil)
	for _, ids := range [][]int64{{9, 2}, {8}, {0}, {9, 9}, nil} {
		input := PelicanPlanInput{GroupID: 7, ModelID: "gpt-5.2", IntervalMinutes: 15, MaxResults: 20, MinChars: 100, PreferredAccountIDs: ids}
		for _, update := range []bool{false, true} {
			before := len(repo.inputs)
			var plan *PelicanPlan
			var err error
			if update {
				plan, err = svc.UpdatePlan(context.Background(), 1, input)
			} else {
				plan, err = svc.CreatePlan(context.Background(), input)
			}
			valid := len(ids) == 0 || reflect.DeepEqual(ids, []int64{9, 2})
			if !valid {
				if !errors.Is(err, ErrPelicanInvalid) || len(repo.inputs) != before {
					t.Fatalf("invalid write: ids=%v err=%v", ids, err)
				}
				continue
			}
			want := ids
			if want == nil {
				want = []int64{}
			}
			if err != nil || plan == nil || !reflect.DeepEqual(plan.PreferredAccountIDs, want) || len(repo.inputs) != before+1 {
				t.Fatalf("ids=%v plan=%v err=%v", ids, plan, err)
			}
		}
	}
}

func TestPelicanPreferredValidationAndCatalogue(t *testing.T) {
	good, foreign, inactive, unsupported, shadow := preferredPelicanAccount(9), preferredPelicanAccount(2), preferredPelicanAccount(3), preferredPelicanAccount(4), preferredPelicanAccount(5)
	foreign.GroupIDs = []int64{8}
	inactive.Status = "inactive"
	unsupported.Type = "unsupported"
	parent := int64(1)
	shadow.ParentAccountID = &parent
	svc := NewPelicanTestService(nil, pelicanModelsAccounts{accounts: []Account{good, foreign, inactive, unsupported, shadow}}, pelicanReviewGroups{}, nil)
	items, err := svc.Accounts(context.Background(), 7)
	if err != nil || !reflect.DeepEqual(items, []PelicanSelectableAccount{{ID: 9, Name: "test"}}) {
		t.Fatalf("catalogue=%v err=%v", items, err)
	}
	for _, ids := range [][]int64{{9}, {}, {2}, {3}, {4}, {5}, {99}} {
		err := svc.validatePreferredAccounts(context.Background(), PelicanPlanInput{GroupID: 7, PreferredAccountIDs: ids})
		wantValid := len(ids) == 0 || ids[0] == 9
		if (err == nil) != wantValid {
			t.Fatalf("ids=%v err=%v", ids, err)
		}
	}
	for _, ids := range [][]int64{{0}, {-1}, {9, 9}, make([]int64, 51)} {
		input := PelicanPlanInput{ModelID: "gpt-test", PreferredAccountIDs: ids}
		if !errors.Is(normalizePelicanInput(&input), ErrPelicanInvalid) {
			t.Fatalf("accepted invalid IDs=%v", ids)
		}
	}
}

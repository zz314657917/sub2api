package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type pelicanReviewRepo struct {
	PelicanTestRepository
	mu       sync.Mutex
	manuals  []bool
	saved    []PelicanResult
	released chan struct{}
	canRun   map[int64]bool
}

func (r *pelicanReviewRepo) Claim(_ context.Context, _ int64, manual bool, _ time.Time) (*PelicanPlan, error) {
	r.mu.Lock()
	r.manuals = append(r.manuals, manual)
	r.mu.Unlock()
	return &PelicanPlan{ID: 1, GroupID: 7, ModelID: "gpt-test", MinChars: 100, MaxResults: 10, RunGeneration: 1}, nil
}
func (r *pelicanReviewRepo) Release(context.Context, int64, int64) error {
	r.released <- struct{}{}
	return nil
}
func (r *pelicanReviewRepo) CanRunAccount(_ context.Context, _ int64, _ int64, accountID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	allowed, configured := r.canRun[accountID]
	if !configured {
		return true, nil
	}
	return allowed, nil
}
func (r *pelicanReviewRepo) SaveResult(_ context.Context, result *PelicanResult, _ int, _ int64) error {
	r.mu.Lock()
	r.saved = append(r.saved, *result)
	r.mu.Unlock()
	return nil
}

type pelicanReviewAccounts struct {
	AccountRepository
	accounts []Account
	fresh    map[int64]*Account
}

func (a pelicanReviewAccounts) ListByGroup(context.Context, int64) ([]Account, error) {
	return a.accounts, nil
}
func (a pelicanReviewAccounts) GetByID(_ context.Context, id int64) (*Account, error) {
	if account := a.fresh[id]; account != nil {
		return account, nil
	}
	return nil, errors.New("missing account")
}

type pelicanReviewGroups struct{ GroupRepository }

func (pelicanReviewGroups) GetByIDLite(context.Context, int64) (*Group, error) {
	return &Group{ID: 7, Platform: PlatformOpenAI, Status: StatusActive}, nil
}

type pelicanReviewExecutor struct {
	mu        sync.Mutex
	active    int
	maxActive int
	started   chan int64
	release   <-chan struct{}
	cancelled chan struct{}
	calls     []int64
}

func (e *pelicanReviewExecutor) RunPelicanTest(ctx context.Context, accountID int64, _ string) (*PelicanResult, error) {
	e.mu.Lock()
	e.active++
	if e.active > e.maxActive {
		e.maxActive = e.active
	}
	e.calls = append(e.calls, accountID)
	e.mu.Unlock()
	e.started <- accountID
	defer func() {
		e.mu.Lock()
		e.active--
		e.mu.Unlock()
	}()
	select {
	case <-e.release:
		return &PelicanResult{HTML: "<html>" + strings.Repeat("x", 100) + "</html>"}, nil
	case <-ctx.Done():
		select {
		case e.cancelled <- struct{}{}:
		default:
		}
		return nil, ctx.Err()
	}
}

func reviewAccount(id int64, member bool) *Account {
	groups := []int64{}
	if member {
		groups = []int64{7}
	}
	return &Account{ID: id, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, GroupIDs: groups}
}

func newPelicanReviewService(accounts []Account, fresh map[int64]*Account, executor *pelicanReviewExecutor) (*PelicanTestService, *pelicanReviewRepo) {
	repo := &pelicanReviewRepo{released: make(chan struct{}, 4)}
	service := NewPelicanTestService(repo, pelicanReviewAccounts{accounts: accounts, fresh: fresh}, pelicanReviewGroups{}, executor)
	service.Start(context.Background())
	return service, repo
}

func TestPelicanReviewRunnerBoundsConcurrencyAndSkipsStaleMember(t *testing.T) {
	accounts := []Account{*reviewAccount(1, true), *reviewAccount(2, true), *reviewAccount(3, true), *reviewAccount(4, true)}
	fresh := map[int64]*Account{1: reviewAccount(1, true), 2: reviewAccount(2, true), 3: reviewAccount(3, true), 4: reviewAccount(4, false)}
	release := make(chan struct{})
	executor := &pelicanReviewExecutor{started: make(chan int64, 4), release: release, cancelled: make(chan struct{}, 4)}
	service, repo := newPelicanReviewService(accounts, fresh, executor)
	defer service.Stop(context.Background())
	if err := service.RunNow(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		select {
		case <-executor.started:
		case <-time.After(time.Second):
			t.Fatal("three workers did not start")
		}
	}
	executor.mu.Lock()
	if executor.maxActive != 3 {
		t.Fatalf("max concurrent provider calls = %d, want 3", executor.maxActive)
	}
	executor.mu.Unlock()
	close(release)
	select {
	case <-repo.released:
	case <-time.After(time.Second):
		t.Fatal("plan did not release")
	}
	repo.mu.Lock()
	defer repo.mu.Unlock()
	foundSkipped := false
	for _, result := range repo.saved {
		if result.AccountID == 4 && result.Status == "skipped" {
			foundSkipped = true
		}
	}
	if !foundSkipped {
		t.Fatal("stale group member reached no skipped result")
	}
}

func TestPelicanReviewRunnerSkipsAccountThatFailsLiveRepositoryGate(t *testing.T) {
	executor := &pelicanReviewExecutor{started: make(chan int64, 1), release: make(chan struct{}), cancelled: make(chan struct{}, 1)}
	service, repo := newPelicanReviewService(
		[]Account{*reviewAccount(1, true)},
		map[int64]*Account{1: reviewAccount(1, true)},
		executor,
	)
	repo.canRun = map[int64]bool{1: false}
	defer service.Stop(context.Background())

	if err := service.RunNow(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	select {
	case <-repo.released:
	case <-time.After(time.Second):
		t.Fatal("plan did not release after live repository gate rejected account")
	}

	executor.mu.Lock()
	calls := append([]int64(nil), executor.calls...)
	executor.mu.Unlock()
	if len(calls) != 0 {
		t.Fatalf("provider calls = %#v, want none", calls)
	}
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.saved) != 1 || repo.saved[0].AccountID != 1 || repo.saved[0].Status != "skipped" || repo.saved[0].ErrorMessage != "plan or membership changed" {
		t.Fatalf("saved results = %#v", repo.saved)
	}
}

func TestPelicanReviewStopCancelsProviderAndReturns(t *testing.T) {
	release := make(chan struct{})
	executor := &pelicanReviewExecutor{started: make(chan int64, 1), release: release, cancelled: make(chan struct{}, 1)}
	service, _ := newPelicanReviewService([]Account{*reviewAccount(1, true)}, map[int64]*Account{1: reviewAccount(1, true)}, executor)
	if err := service.RunNow(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	select {
	case <-executor.started:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	select {
	case <-executor.cancelled:
	default:
		t.Fatal("provider did not observe cancellation")
	}
}

func TestPelicanReviewAcceptedManualOutlivesHTTPContextAndSchedulerUsesAutomaticClaim(t *testing.T) {
	release := make(chan struct{})
	executor := &pelicanReviewExecutor{started: make(chan int64, 2), release: release, cancelled: make(chan struct{}, 2)}
	service, repo := newPelicanReviewService([]Account{*reviewAccount(1, true)}, map[int64]*Account{1: reviewAccount(1, true)}, executor)
	defer service.Stop(context.Background())
	httpCtx, cancel := context.WithCancel(context.Background())
	if err := service.RunNow(httpCtx, 1); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case <-executor.started:
	case <-time.After(time.Second):
		t.Fatal("accepted manual run was cancelled with HTTP context")
	}
	close(release)
	select {
	case <-repo.released:
	case <-time.After(time.Second):
		t.Fatal("manual run did not release")
	}
	if err := service.claimAndRun(context.Background(), 1, false); err != nil {
		t.Fatal(err)
	}
	repo.mu.Lock()
	if len(repo.manuals) != 2 || repo.manuals[0] != true || repo.manuals[1] != false {
		t.Fatalf("claim modes = %#v", repo.manuals)
	}
	repo.mu.Unlock()
}

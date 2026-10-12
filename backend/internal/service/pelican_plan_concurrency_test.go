package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// This fake owns an atomic per-plan Claim, independent of the local planGate.
// It also records quota and due-time changes so a capacity rejection cannot
// accidentally be counted as a successful repository claim.
type pelicanConcurrentRepo struct {
	PelicanTestRepository
	mu          sync.Mutex
	active      map[int64]int64
	generations map[int64]int64
	due         map[int64]time.Time
	claims      []pelicanConcurrentClaim
	reserved    int
	claimErr    error
	claimHook   func(context.Context) error
	models      []string
	finalized   chan int64
}

type pelicanConcurrentClaim struct {
	id     int64
	manual bool
}

func newPelicanConcurrentRepo() *pelicanConcurrentRepo {
	return &pelicanConcurrentRepo{
		active: make(map[int64]int64), generations: make(map[int64]int64),
		due: make(map[int64]time.Time), finalized: make(chan int64, 32),
	}
}

func (r *pelicanConcurrentRepo) Claim(ctx context.Context, id int64, manual bool, now time.Time) (*PelicanPlan, error) {
	r.mu.Lock()
	r.claims = append(r.claims, pelicanConcurrentClaim{id, manual})
	hook, claimErr := r.claimHook, r.claimErr
	r.mu.Unlock()
	if hook != nil {
		if err := hook(ctx); err != nil {
			return nil, err
		}
	}
	if claimErr != nil {
		return nil, claimErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active[id] != 0 {
		return nil, ErrPelicanConflict
	}
	if !manual && r.due[id].After(now) {
		return nil, ErrPelicanConflict
	}
	r.generations[id]++
	r.active[id] = r.generations[id]
	r.due[id] = now.Add(time.Hour)
	return &PelicanPlan{
		ID: id, GroupID: id, ModelID: "gpt-test", ModelIDs: append([]string(nil), r.models...),
		MinChars: 100, MaxResults: 1, TimeoutSeconds: 30, RunGeneration: r.generations[id],
	}, nil
}

func (r *pelicanConcurrentRepo) Finalize(_ context.Context, id, generation int64) error {
	r.mu.Lock()
	if r.active[id] == generation {
		delete(r.active, id)
	}
	r.mu.Unlock()
	r.finalized <- id
	return nil
}

func (r *pelicanConcurrentRepo) CanRunAccount(_ context.Context, id, generation, _ int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active[id] == generation, nil
}

func (r *pelicanConcurrentRepo) ReserveAttempt(_ context.Context, id, generation int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active[id] != generation {
		return ErrPelicanConflict
	}
	r.reserved++
	return nil
}

func (r *pelicanConcurrentRepo) CleanupDue(context.Context) error { return nil }
func (r *pelicanConcurrentRepo) SaveResult(context.Context, *PelicanResult, int, int64) error {
	return nil
}

func (r *pelicanConcurrentRepo) snapshot() (claims []pelicanConcurrentClaim, reserved, active int, due map[int64]time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	due = make(map[int64]time.Time, len(r.due))
	for id, at := range r.due {
		due[id] = at
	}
	return append([]pelicanConcurrentClaim(nil), r.claims...), r.reserved, len(r.active), due
}

type pelicanConcurrentAccounts struct{ AccountRepository }

func (pelicanConcurrentAccounts) GetByID(_ context.Context, id int64) (*Account, error) {
	account := reviewAccount(id, true)
	account.GroupIDs = []int64{id}
	return account, nil
}

type pelicanConcurrentGroups struct{ GroupRepository }

func (pelicanConcurrentGroups) GetByIDLite(_ context.Context, id int64) (*Group, error) {
	return &Group{ID: id, Platform: PlatformOpenAI, Status: StatusActive}, nil
}

type pelicanConcurrentCall struct {
	id    int64
	model string
}

type pelicanConcurrentExecutor struct {
	mu         sync.Mutex
	active     int
	maxActive  int
	byPlan     map[int64]int
	overlap    bool
	leases     map[int64]int
	leaseError bool
	releases   map[int64]chan struct{}
	started    chan pelicanConcurrentCall
	cancelled  chan int64
	cancelExit <-chan struct{}
}

func newPelicanConcurrentExecutor() *pelicanConcurrentExecutor {
	ex := &pelicanConcurrentExecutor{
		byPlan: make(map[int64]int), leases: make(map[int64]int),
		releases: make(map[int64]chan struct{}),
		started:  make(chan pelicanConcurrentCall, 32), cancelled: make(chan int64, 32),
	}
	for id := int64(1); id <= 10; id++ {
		ex.releases[id] = make(chan struct{}, 8)
	}
	return ex
}

func (e *pelicanConcurrentExecutor) selectAccount(ctx context.Context, id int64, _ string) (*AccountSelectionResult, error) {
	account, _ := (pelicanConcurrentAccounts{}).GetByID(ctx, id)
	e.mu.Lock()
	e.leases[id]++
	if e.leases[id] != 1 {
		e.leaseError = true
	}
	e.mu.Unlock()
	return &AccountSelectionResult{Account: account, Acquired: true, ReleaseFunc: func() {
		e.mu.Lock()
		e.leases[id]--
		e.mu.Unlock()
	}}, nil
}

func (e *pelicanConcurrentExecutor) RunPelicanTest(ctx context.Context, id int64, model, _, _ string) (*PelicanResult, error) {
	e.mu.Lock()
	e.active++
	e.byPlan[id]++
	if e.byPlan[id] != 1 {
		e.overlap = true
	}
	if e.active > e.maxActive {
		e.maxActive = e.active
	}
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		e.active--
		e.byPlan[id]--
		e.mu.Unlock()
	}()
	e.started <- pelicanConcurrentCall{id, model}
	select {
	case <-e.releases[id]:
		return &PelicanResult{HTML: "<html>" + strings.Repeat("x", 100) + "</html>"}, nil
	case <-ctx.Done():
		e.cancelled <- id
		if e.cancelExit != nil {
			<-e.cancelExit
		}
		return nil, ctx.Err()
	}
}

func newPelicanConcurrentService(t *testing.T, r *pelicanConcurrentRepo, ex *pelicanConcurrentExecutor, parent context.Context) *PelicanTestService {
	t.Helper()
	s := NewPelicanTestService(r, pelicanConcurrentAccounts{}, pelicanConcurrentGroups{}, ex)
	s.SetSettingsRepository(pelicanDefaultSettingsRepo{})
	s.SetAccountSelector(pelicanSelectorFunc(ex.selectAccount))
	s.Start(parent)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.Stop(ctx); err != nil {
			t.Errorf("cleanup Stop: %v", err)
		}
	})
	return s
}

func pelicanConcurrentReceive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("channel barrier timed out")
		var zero T
		return zero
	}
}

// Finalize's notification precedes the deferred gate release. Wait for that
// observable release instead of assuming goroutine scheduling after the signal.
func pelicanConcurrentWaitGate(t *testing.T, s *PelicanTestService, want int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for len(s.planGate) != want {
		select {
		case <-tick.C:
		case <-ctx.Done():
			t.Fatalf("occupied slots=%d want=%d", len(s.planGate), want)
		}
	}
}

func TestPelicanPlanConcurrencyCapacityAndScheduledReuse(t *testing.T) {
	r, ex := newPelicanConcurrentRepo(), newPelicanConcurrentExecutor()
	due := time.Now().Add(-time.Hour)
	r.due[5] = due
	s := newPelicanConcurrentService(t, r, ex, context.Background())
	for id := int64(1); id <= 4; id++ {
		var err error
		if id%2 == 0 {
			err = s.claimAndRun(context.Background(), id, false)
		} else {
			err = s.RunNow(context.Background(), id)
		}
		if err != nil {
			t.Fatal(err)
		}
		call := pelicanConcurrentReceive(t, ex.started)
		if call.id != id {
			t.Fatalf("started=%+v want plan=%d", call, id)
		}
	}
	ex.mu.Lock()
	active, maxActive := ex.active, ex.maxActive
	ex.mu.Unlock()
	if active != 4 || maxActive != 4 {
		t.Fatalf("simultaneous executions=%d maximum=%d", active, maxActive)
	}
	for _, manual := range []bool{true, false} {
		if err := s.claimAndRun(context.Background(), 5, manual); !errors.Is(err, ErrPelicanConflict) {
			t.Fatalf("full gate manual=%v err=%v", manual, err)
		}
	}
	claims, reserved, _, times := r.snapshot()
	if len(claims) != 4 || reserved != 4 || !times[5].Equal(due) {
		t.Fatalf("capacity rejection side effects: claims=%v reserved=%d due=%v", claims, reserved, times[5])
	}
	for i, claim := range claims {
		if claim.manual != (i%2 == 0) {
			t.Fatalf("manual/scheduled flag lost: %v", claims)
		}
	}
	ex.releases[1] <- struct{}{}
	if id := pelicanConcurrentReceive(t, r.finalized); id != 1 {
		t.Fatalf("finalized=%d", id)
	}
	pelicanConcurrentWaitGate(t, s, 3)
	if err := s.claimAndRun(context.Background(), 5, false); err != nil {
		t.Fatalf("due scheduled plan cannot reuse completed slot: %v", err)
	}
	if call := pelicanConcurrentReceive(t, ex.started); call.id != 5 {
		t.Fatalf("retry started=%+v", call)
	}
	claims, reserved, _, times = r.snapshot()
	if len(claims) != 5 || claims[4].manual || reserved != 5 || !times[5].After(due) {
		t.Fatalf("scheduled retry=%v reserved=%d due=%v", claims, reserved, times[5])
	}
}

func TestPelicanPlanConcurrencySamePlanClaimConflict(t *testing.T) {
	r, ex := newPelicanConcurrentRepo(), newPelicanConcurrentExecutor()
	s := newPelicanConcurrentService(t, r, ex, context.Background())
	if err := s.RunNow(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	pelicanConcurrentReceive(t, ex.started)
	if err := s.RunNow(context.Background(), 1); !errors.Is(err, ErrPelicanConflict) {
		t.Fatalf("same-plan Claim err=%v", err)
	}
	claims, reserved, active, _ := r.snapshot()
	if len(claims) != 2 || reserved != 1 || active != 1 || len(s.planGate) != 1 {
		t.Fatalf("repository conflict or gate release missing: claims=%v reserved=%d active=%d gate=%d", claims, reserved, active, len(s.planGate))
	}
	for id := int64(2); id <= 4; id++ {
		if err := s.RunNow(context.Background(), id); err != nil {
			t.Fatalf("conflict slot cannot be reused: %v", err)
		}
		pelicanConcurrentReceive(t, ex.started)
	}
	if err := s.RunNow(context.Background(), 5); !errors.Is(err, ErrPelicanConflict) {
		t.Fatalf("capacity after conflict err=%v", err)
	}
}

func TestPelicanPlanConcurrencyFailureSlotReuse(t *testing.T) {
	for _, scenario := range []string{"settings", "disabled", "claim"} {
		t.Run(scenario, func(t *testing.T) {
			r, ex := newPelicanConcurrentRepo(), newPelicanConcurrentExecutor()
			s := newPelicanConcurrentService(t, r, ex, context.Background())
			failure := errors.New("injected " + scenario + " failure")
			wantErr := failure
			switch scenario {
			case "settings":
				s.SetSettingsRepository(&pelicanSettingsRepoStub{err: failure})
			case "disabled":
				s.SetSettingsRepository(&pelicanSettingsRepoStub{value: `{"enabled":false}`})
				wantErr = ErrPelicanDisabled
			case "claim":
				r.claimErr = failure
			}
			// More attempts than capacity ensure failures do not silently leak slots.
			for i := 0; i < 6; i++ {
				if err := s.RunNow(context.Background(), 1); !errors.Is(err, wantErr) {
					t.Fatalf("failure %d err=%v want=%v", i, err, wantErr)
				}
				if len(s.planGate) != 0 {
					t.Fatalf("failure leaked %d slots", len(s.planGate))
				}
			}
			claims, reserved, active, _ := r.snapshot()
			wantClaims := 0
			if scenario == "claim" {
				wantClaims = 6
			}
			if len(claims) != wantClaims || reserved != 0 || active != 0 {
				t.Fatalf("failure side effects: claims=%v reserved=%d active=%d", claims, reserved, active)
			}
			s.SetSettingsRepository(pelicanDefaultSettingsRepo{})
			r.mu.Lock()
			r.claimErr = nil
			r.mu.Unlock()
			for id := int64(1); id <= 4; id++ {
				if err := s.RunNow(context.Background(), id); err != nil {
					t.Fatalf("failure slot cannot be reused: %v", err)
				}
				pelicanConcurrentReceive(t, ex.started)
			}
		})
	}
}

func TestPelicanPlanConcurrencyStopWaitsForExecutions(t *testing.T) {
	r, ex := newPelicanConcurrentRepo(), newPelicanConcurrentExecutor()
	exit := make(chan struct{})
	ex.cancelExit = exit
	// Unblock cleanup even when an assertion fails.
	var exitOnce sync.Once
	unblock := func() { exitOnce.Do(func() { close(exit) }) }
	defer unblock()
	s := newPelicanConcurrentService(t, r, ex, context.Background())
	for id := int64(1); id <= 4; id++ {
		if err := s.RunNow(context.Background(), id); err != nil {
			t.Fatal(err)
		}
		pelicanConcurrentReceive(t, ex.started)
	}
	stopCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopResult := make(chan error, 1)
	go func() { stopResult <- s.Stop(stopCtx) }()
	seen := make(map[int64]bool)
	for i := 0; i < 4; i++ {
		seen[pelicanConcurrentReceive(t, ex.cancelled)] = true
	}
	if len(seen) != 4 {
		t.Fatalf("cancelled plan identities=%v", seen)
	}
	select {
	case err := <-stopResult:
		t.Fatalf("Stop returned before workers exited: %v", err)
	default:
	}
	cancel()
	if err := pelicanConcurrentReceive(t, stopResult); !errors.Is(err, context.Canceled) {
		t.Fatalf("bounded Stop err=%v", err)
	}
	expiredCtx, expiredCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer expiredCancel()
	if err := s.Stop(expiredCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timed-out Stop err=%v", err)
	}
	if len(s.planGate) != 4 {
		t.Fatalf("workers released slots before executor exit: %d", len(s.planGate))
	}
	for _, manual := range []bool{true, false} {
		if err := s.claimAndRun(context.Background(), 5, manual); !errors.Is(err, ErrPelicanConflict) {
			t.Fatalf("start after Stop manual=%v err=%v", manual, err)
		}
	}
	// Let all cancellation acknowledgements finish, then an unbounded Stop
	// must wait for Finalize, account releases and planGate releases.
	unblock()
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	claims, reserved, active, _ := r.snapshot()
	if len(claims) != 4 || reserved != 4 || active != 0 || len(s.planGate) != 0 {
		t.Fatalf("Stop cleanup: claims=%v reserved=%d active=%d gate=%d", claims, reserved, active, len(s.planGate))
	}
	ex.mu.Lock()
	defer ex.mu.Unlock()
	if ex.active != 0 {
		t.Fatalf("Stop retained %d executions", ex.active)
	}
	for id, leases := range ex.leases {
		if leases != 0 {
			t.Fatalf("Stop leaked plan %d account leases=%d", id, leases)
		}
	}
}

func TestPelicanPlanConcurrencyStopCancelsClaim(t *testing.T) {
	r, ex := newPelicanConcurrentRepo(), newPelicanConcurrentExecutor()
	entered, cancelled, exit := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var exitOnce sync.Once
	unblock := func() { exitOnce.Do(func() { close(exit) }) }
	defer unblock()
	r.claimHook = func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		close(cancelled)
		<-exit
		return ctx.Err()
	}
	s := newPelicanConcurrentService(t, r, ex, context.Background())
	startResult := make(chan error, 1)
	go func() { startResult <- s.RunNow(context.Background(), 1) }()
	pelicanConcurrentReceive(t, entered)
	stopResult := make(chan error, 1)
	go func() { stopResult <- s.Stop(context.Background()) }()
	pelicanConcurrentReceive(t, cancelled)
	select {
	case err := <-stopResult:
		t.Fatalf("Stop returned while Claim was still registered: %v", err)
	default:
	}
	if len(s.planGate) != 1 {
		t.Fatalf("blocked claim gate=%d", len(s.planGate))
	}
	if err := s.RunNow(context.Background(), 2); !errors.Is(err, ErrPelicanConflict) {
		t.Fatalf("start during Stop err=%v", err)
	}
	unblock()
	if err := pelicanConcurrentReceive(t, startResult); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Claim err=%v", err)
	}
	if err := pelicanConcurrentReceive(t, stopResult); err != nil {
		t.Fatalf("Stop after Claim exit err=%v", err)
	}
	claims, reserved, active, _ := r.snapshot()
	if len(claims) != 1 || reserved != 0 || active != 0 || len(s.planGate) != 0 {
		t.Fatalf("Claim cancellation cleanup: claims=%v reserved=%d active=%d gate=%d", claims, reserved, active, len(s.planGate))
	}
}

func TestPelicanPlanConcurrencyParentCancellationReleasesSlots(t *testing.T) {
	r, ex := newPelicanConcurrentRepo(), newPelicanConcurrentExecutor()
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := newPelicanConcurrentService(t, r, ex, parent)
	for id := int64(1); id <= 4; id++ {
		if err := s.RunNow(context.Background(), id); err != nil {
			t.Fatal(err)
		}
		pelicanConcurrentReceive(t, ex.started)
	}
	cancel()
	for i := 0; i < 4; i++ {
		pelicanConcurrentReceive(t, ex.cancelled)
		pelicanConcurrentReceive(t, r.finalized)
	}
	pelicanConcurrentWaitGate(t, s, 0)
	if err := s.RunNow(context.Background(), 5); !errors.Is(err, ErrPelicanConflict) {
		t.Fatalf("cancelled parent accepted new round: %v", err)
	}
	claims, reserved, active, _ := r.snapshot()
	if len(claims) != 4 || reserved != 4 || active != 0 {
		t.Fatalf("parent cancellation cleanup: claims=%v reserved=%d active=%d", claims, reserved, active)
	}
}

func TestPelicanPlanConcurrencyModelsRemainSerial(t *testing.T) {
	r, ex := newPelicanConcurrentRepo(), newPelicanConcurrentExecutor()
	r.models = []string{"gpt-first", "gpt-second"}
	s := newPelicanConcurrentService(t, r, ex, context.Background())
	for id := int64(1); id <= 2; id++ {
		if err := s.RunNow(context.Background(), id); err != nil {
			t.Fatal(err)
		}
		if call := pelicanConcurrentReceive(t, ex.started); call != (pelicanConcurrentCall{id, "gpt-first"}) {
			t.Fatalf("first model=%+v", call)
		}
	}
	select {
	case call := <-ex.started:
		t.Fatalf("second model started before first released: %+v", call)
	default:
	}
	for id := int64(1); id <= 2; id++ {
		ex.releases[id] <- struct{}{}
		if call := pelicanConcurrentReceive(t, ex.started); call != (pelicanConcurrentCall{id, "gpt-second"}) {
			t.Fatalf("second model=%+v", call)
		}
	}
	for id := int64(1); id <= 2; id++ {
		ex.releases[id] <- struct{}{}
		pelicanConcurrentReceive(t, r.finalized)
	}
	pelicanConcurrentWaitGate(t, s, 0)
	ex.mu.Lock()
	defer ex.mu.Unlock()
	if ex.overlap || ex.leaseError || ex.active != 0 || ex.maxActive != 2 {
		t.Fatalf("model/lease sequencing: overlap=%v leaseError=%v active=%d max=%d", ex.overlap, ex.leaseError, ex.active, ex.maxActive)
	}
	for id, leases := range ex.leases {
		if leases != 0 {
			t.Fatalf("plan %d leaked %d account leases", id, leases)
		}
	}
}

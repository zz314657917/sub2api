package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const PelicanPromptVersion = "pelican-v1"
const PelicanPrompt = "请生成一个精致、完整、独立的中文海边鹈鹕骑自行车动画 HTML：鹈鹕双脚连续踩踏，车轮持续转动，海浪和云层轻轻移动，循环流畅。使用内联 CSS keyframes 和 SVG，支持不同屏幕宽度。不得包含 JavaScript、外部资源、链接或网络请求。仅输出完整 HTML 文档，不加 Markdown 代码围栏或解释。"

var (
	ErrPelicanConflict = errors.New("pelican test is already running")
	ErrPelicanNotFound = errors.New("pelican test not found")
	ErrPelicanInvalid  = errors.New("invalid pelican test request")
)

type PelicanAuthorization struct {
	GroupIDs []int64
	Admin    bool
}
type PelicanPlan struct {
	ID              int64      `json:"id"`
	GroupID         int64      `json:"group_id"`
	GroupName       string     `json:"group_name"`
	ModelID         string     `json:"model_id"`
	IntervalMinutes int        `json:"interval_minutes"`
	Enabled         bool       `json:"enabled"`
	MaxResults      int        `json:"max_results"`
	MinChars        int        `json:"min_chars"`
	LastRunAt       *time.Time `json:"last_run_at,omitempty"`
	NextRunAt       *time.Time `json:"next_run_at,omitempty"`
	RunningUntil    *time.Time `json:"running_until,omitempty"`
	RunGeneration   int64      `json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
type PelicanPlanInput struct {
	GroupID         int64  `json:"group_id"`
	ModelID         string `json:"model_id"`
	IntervalMinutes int    `json:"interval_minutes"`
	Enabled         bool   `json:"enabled"`
	MaxResults      int    `json:"max_results"`
	MinChars        int    `json:"min_chars"`
}
type PelicanResult struct {
	ID            int64      `json:"id"`
	PlanID        int64      `json:"plan_id"`
	GroupID       int64      `json:"group_id"`
	AccountID     int64      `json:"account_id"`
	ModelID       string     `json:"model_id"`
	PromptVersion string     `json:"prompt_version"`
	Status        string     `json:"status"`
	ErrorMessage  string     `json:"error_message"`
	LatencyMS     int64      `json:"latency_ms"`
	CharCount     int        `json:"char_count"`
	MinChars      int        `json:"min_chars"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	HTML          string     `json:"html,omitempty"`
}
type PelicanEntry struct {
	PlanID          int64      `json:"plan_id"`
	GroupID         int64      `json:"group_id"`
	GroupName       string     `json:"group_name"`
	AccountID       int64      `json:"account_id"`
	ModelID         string     `json:"model_id"`
	Status          string     `json:"status"`
	LatencyMS       int64      `json:"latency_ms"`
	CharCount       int        `json:"char_count"`
	MinChars        int        `json:"min_chars"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	HistoryCount    int        `json:"history_count"`
	ResultID        int64      `json:"result_id"`
	ArtworkResultID *int64     `json:"artwork_result_id"`
	ErrorMessage    string     `json:"error_message"`
}
type PelicanGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type PelicanListResponse struct {
	Items    []PelicanEntry `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Groups   []PelicanGroup `json:"groups"`
}
type PelicanTestRepository interface {
	CreatePlan(context.Context, PelicanPlanInput, string) (*PelicanPlan, error)
	UpdatePlan(context.Context, int64, PelicanPlanInput, string) (*PelicanPlan, error)
	DeletePlan(context.Context, int64) error
	ListPlans(context.Context) ([]PelicanPlan, error)
	Claim(context.Context, int64, bool, time.Time) (*PelicanPlan, error)
	Release(context.Context, int64, int64) error
	CanRunAccount(context.Context, int64, int64, int64) (bool, error)
	ListGallery(context.Context, PelicanAuthorization, int64, int64, int, int) (*PelicanListResponse, error)
	ListHistory(context.Context, PelicanAuthorization, int64, int64) ([]PelicanResult, error)
	GetResult(context.Context, PelicanAuthorization, int64) (*PelicanResult, error)
	SaveResult(context.Context, *PelicanResult, int, int64) error
}
type PelicanTestExecutor interface {
	RunPelicanTest(context.Context, int64, string) (*PelicanResult, error)
}
type PelicanTestService struct {
	repo     PelicanTestRepository
	accounts AccountRepository
	groups   GroupRepository
	tester   PelicanTestExecutor
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
	rootCtx  context.Context
	planGate chan struct{}
	started  bool
	stopped  bool
	stopOnce sync.Once
	done     chan struct{}
}

func NewPelicanTestService(r PelicanTestRepository, a AccountRepository, g GroupRepository, t PelicanTestExecutor) *PelicanTestService {
	return &PelicanTestService{repo: r, accounts: a, groups: g, tester: t, planGate: make(chan struct{}, 1), done: make(chan struct{})}
}
func (s *PelicanTestService) ListGallery(c context.Context, a PelicanAuthorization, g, ac int64, p, ps int) (*PelicanListResponse, error) {
	return s.repo.ListGallery(c, a, g, ac, p, ps)
}
func (s *PelicanTestService) ListHistory(c context.Context, a PelicanAuthorization, p, ac int64) ([]PelicanResult, error) {
	return s.repo.ListHistory(c, a, p, ac)
}
func (s *PelicanTestService) GetResult(c context.Context, a PelicanAuthorization, id int64) (*PelicanResult, error) {
	return s.repo.GetResult(c, a, id)
}
func (s *PelicanTestService) ListPlans(c context.Context) ([]PelicanPlan, error) {
	return s.repo.ListPlans(c)
}
func validPelicanInput(i PelicanPlanInput) error {
	if i.GroupID <= 0 || strings.TrimSpace(i.ModelID) == "" || len(i.ModelID) > 200 || isOpenAIImageModel(i.ModelID) || strings.ContainsAny(i.ModelID, "\r\n\x00") || i.IntervalMinutes < 15 || i.IntervalMinutes > 1440 || i.MaxResults < 1 || i.MaxResults > 50 || i.MinChars < 100 || i.MinChars > 50000 {
		return ErrPelicanInvalid
	}
	return nil
}
func (s *PelicanTestService) CreatePlan(c context.Context, i PelicanPlanInput) (*PelicanPlan, error) {
	if e := validPelicanInput(i); e != nil {
		return nil, e
	}
	g, e := s.groups.GetByIDLite(c, i.GroupID)
	if e != nil {
		return nil, e
	}
	if g == nil || !g.IsActive() || !strings.EqualFold(g.Platform, PlatformOpenAI) {
		return nil, ErrPelicanInvalid
	}
	i.Enabled = false
	return s.repo.CreatePlan(c, i, g.Name)
}
func (s *PelicanTestService) UpdatePlan(c context.Context, id int64, i PelicanPlanInput) (*PelicanPlan, error) {
	if e := validPelicanInput(i); e != nil {
		return nil, e
	}
	g, e := s.groups.GetByIDLite(c, i.GroupID)
	if e != nil {
		return nil, e
	}
	if g == nil || !g.IsActive() || !strings.EqualFold(g.Platform, PlatformOpenAI) {
		return nil, ErrPelicanInvalid
	}
	return s.repo.UpdatePlan(c, id, i, g.Name)
}
func (s *PelicanTestService) DeletePlan(c context.Context, id int64) error {
	return s.repo.DeletePlan(c, id)
}
func (s *PelicanTestService) RunNow(c context.Context, id int64) error {
	return s.claimAndRun(c, id, true)
}
func (s *PelicanTestService) claimAndRun(c context.Context, id int64, manual bool) error {
	s.mu.Lock()
	if s.stopped || s.rootCtx == nil || s.rootCtx.Err() != nil {
		s.mu.Unlock()
		return ErrPelicanConflict
	}
	runCtx := s.rootCtx
	select {
	case s.planGate <- struct{}{}:
	default:
		s.mu.Unlock()
		return ErrPelicanConflict
	}
	// Register work under the same lock Stop uses before it begins waiting.
	s.wg.Add(1)
	s.mu.Unlock()
	claimCtx, cancel := context.WithTimeout(c, 10*time.Second)
	stopCancel := context.AfterFunc(runCtx, cancel)
	defer func() { stopCancel(); cancel() }()
	p, e := s.repo.Claim(claimCtx, id, manual, time.Now())
	if e != nil {
		<-s.planGate
		s.wg.Done()
		return e
	}
	go func() { defer s.wg.Done(); s.run(runCtx, p) }()
	return nil
}
func (s *PelicanTestService) run(c context.Context, p *PelicanPlan) {
	defer func() { <-s.planGate }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.Release(ctx, p.ID, p.RunGeneration); err != nil {
			log.Printf("pelican plan %d: release failed", p.ID)
		}
	}()
	planCtx, cancelPlan := context.WithTimeout(c, 20*time.Minute)
	defer cancelPlan()
	as, e := s.accounts.ListByGroup(planCtx, p.GroupID)
	if e != nil {
		log.Printf("pelican plan %d: account lookup failed", p.ID)
		return
	}
	sem := make(chan struct{}, 3)
	var workers sync.WaitGroup
	defer workers.Wait()
	for index, a := range as {
		if planCtx.Err() != nil {
			s.saveRemaining(p, as[index:])
			return
		}
		select {
		case sem <- struct{}{}:
		case <-planCtx.Done():
			s.saveRemaining(p, as[index:])
			return
		}
		workers.Add(1)
		go func(a Account) {
			defer workers.Done()
			defer func() { <-sem }()
			accountCtx, cancel := context.WithTimeout(planCtx, 180*time.Second)
			defer cancel()
			started := time.Now()
			// Re-read immediately before a paid call: membership and availability
			// can change while an earlier account is still being tested.
			fresh, err := s.accounts.GetByID(accountCtx, a.ID)
			group, groupErr := s.groups.GetByIDLite(accountCtx, p.GroupID)
			member := false
			if fresh != nil {
				for _, id := range fresh.GroupIDs {
					if id == p.GroupID {
						member = true
						break
					}
				}
			}
			if err != nil || groupErr != nil || group == nil || !group.IsActive() || group.Platform != PlatformOpenAI || !member || !pelicanAccountAvailable(fresh) || accountCtx.Err() != nil {
				s.save(accountCtx, p, a.ID, nil, "skipped", "account unavailable", started)
				return
			}
			allowed, checkErr := s.repo.CanRunAccount(accountCtx, p.ID, p.RunGeneration, a.ID)
			if checkErr != nil || !allowed || accountCtx.Err() != nil {
				s.save(accountCtx, p, a.ID, nil, "skipped", "plan or membership changed", started)
				return
			}
			x, err := s.tester.RunPelicanTest(accountCtx, a.ID, p.ModelID)
			if err != nil || accountCtx.Err() != nil || x == nil {
				s.save(accountCtx, p, a.ID, nil, "failed", "generation failed or timed out", started)
				return
			}
			s.save(accountCtx, p, a.ID, x, "success", "", started)
		}(a)
	}
}
func pelicanAccountAvailable(a *Account) bool {
	return a != nil && a.IsOpenAI() && (a.Type == "apikey" || a.IsOAuth()) && !a.IsOpenAIAgentIdentity() && a.ParentAccountID == nil && a.IsSchedulable() && !a.IsRateLimited() && (a.ExpiresAt == nil || a.ExpiresAt.After(time.Now()))
}
func (s *PelicanTestService) saveRemaining(p *PelicanPlan, accounts []Account) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, a := range accounts {
		if ctx.Err() != nil {
			return
		}
		s.save(ctx, p, a.ID, nil, "skipped", "plan cancelled or time limit reached", time.Now())
	}
}
func (s *PelicanTestService) save(c context.Context, p *PelicanPlan, id int64, x *PelicanResult, status, msg string, started time.Time) {
	r := PelicanResult{PlanID: p.ID, GroupID: p.GroupID, AccountID: id, ModelID: p.ModelID, PromptVersion: PelicanPromptVersion, Status: status, ErrorMessage: msg, MinChars: p.MinChars, StartedAt: started, LatencyMS: time.Since(started).Milliseconds()}
	finished := time.Now()
	r.FinishedAt = &finished
	if x != nil {
		r.HTML = x.HTML
		r.CharCount = utf8.RuneCountInString(x.HTML)
		if r.CharCount < p.MinChars || len(r.HTML) > 256*1024 {
			r.Status = "failed"
			r.ErrorMessage = "generated content is too short"
			r.HTML = ""
		}
	}
	// Persist timeout/failure metadata even when the provider context expired.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c), 3*time.Second)
	defer cancel()
	if err := s.repo.SaveResult(ctx, &r, p.MaxResults, p.RunGeneration); err != nil {
		log.Printf("pelican plan %d account %d: result save failed", p.ID, id)
	}
}
func (s *PelicanTestService) Start(c context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started || s.stopped {
		return
	}
	ctx, cancel := context.WithCancel(c)
	s.cancel = cancel
	s.started = true
	s.rootCtx = ctx
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		tick := time.NewTicker(30 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				plans, err := s.repo.ListPlans(ctx)
				if err != nil {
					log.Print("pelican scheduler: plan lookup failed")
					continue
				}
				for _, p := range plans {
					if p.Enabled && p.NextRunAt != nil && !p.NextRunAt.After(time.Now()) {
						_ = s.claimAndRun(ctx, p.ID, false)
					}
				}
			}
		}
	}()
}
func (s *PelicanTestService) Stop(c context.Context) error {
	s.mu.Lock()
	s.stopped = true
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
		s.rootCtx = nil
	}
	s.mu.Unlock()
	s.stopOnce.Do(func() { go func() { s.wg.Wait(); close(s.done) }() })
	select {
	case <-s.done:
		return nil
	case <-c.Done():
		return fmt.Errorf("stop pelican tests: %w", c.Err())
	}
}

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

const PelicanPromptVersion = "pelican-v1"
const PelicanPrompt = "请生成一个精致、完整、独立的中文海边鹈鹕骑自行车动画 HTML：鹈鹕双脚连续踩踏，车轮持续转动，海浪和云层轻轻移动，循环流畅。使用内联 CSS keyframes 和 SVG，支持不同屏幕宽度。不得包含 JavaScript、外部资源、链接或网络请求。仅输出完整 HTML 文档，不加 Markdown 代码围栏或解释。"
const PelicanTestSettingsKey = "pelican_test_settings"

type PelicanTestSettings struct {
	DisplayName string `json:"display_name"`
	Prompt      string `json:"prompt"`
	Enabled     bool   `json:"enabled"`
}

// PelicanTestSettingsUpdate keeps enabled presence-aware: an older client that
// omits it must not accidentally turn the plaza off.
type PelicanTestSettingsUpdate struct {
	DisplayName string `json:"display_name"`
	Prompt      string `json:"prompt"`
	Enabled     *bool  `json:"enabled"`
}

type PelicanTestMetadata struct {
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}

type pelicanPromptSnapshot struct {
	prompt  string
	version string
}

var (
	ErrPelicanConflict = errors.New("pelican test is already running")
	ErrPelicanNotFound = errors.New("pelican test not found")
	ErrPelicanInvalid  = errors.New("invalid pelican test request")
	ErrPelicanPaused   = errors.New("pelican plan is paused after consecutive failures")
	ErrPelicanQuota    = errors.New("pelican plan daily call limit reached")
	ErrPelicanDisabled = errors.New("pelican tests are disabled")
)

type PelicanAuthorization struct {
	GroupIDs []int64
	Admin    bool
}
type PelicanPlan struct {
	TimeoutSeconds        int        `json:"timeout_seconds"`
	ID                    int64      `json:"id"`
	GroupID               int64      `json:"group_id"`
	GroupName             string     `json:"group_name"`
	ModelID               string     `json:"model_id"`
	IntervalMinutes       int        `json:"interval_minutes"`
	Enabled               bool       `json:"enabled"`
	MaxResults            int        `json:"max_results"`
	MinChars              int        `json:"min_chars"`
	LastRunAt             *time.Time `json:"last_run_at,omitempty"`
	NextRunAt             *time.Time `json:"next_run_at,omitempty"`
	RunningUntil          *time.Time `json:"running_until,omitempty"`
	RunGeneration         int64      `json:"-"`
	DailyCallLimit        int        `json:"daily_call_limit"`
	FailurePauseThreshold int        `json:"failure_pause_threshold"`
	RetentionDays         int        `json:"retention_days"`
	DailyCallsUsed        int        `json:"daily_calls_used"`
	UsageDay              string     `json:"usage_day"`
	ConsecutiveFailedRuns int        `json:"consecutive_failed_runs"`
	PauseReason           string     `json:"pause_reason"`
	EstimatedCallsPerRun  int        `json:"estimated_calls_per_run"`
	LastRunCalls          int        `json:"last_run_calls"`
	ReasoningEffort       string     `json:"reasoning_effort"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
type PelicanPlanInput struct {
	TimeoutSeconds        int    `json:"timeout_seconds"`
	GroupID               int64  `json:"group_id"`
	ModelID               string `json:"model_id"`
	IntervalMinutes       int    `json:"interval_minutes"`
	Enabled               bool   `json:"enabled"`
	MaxResults            int    `json:"max_results"`
	MinChars              int    `json:"min_chars"`
	DailyCallLimit        int    `json:"daily_call_limit"`
	FailurePauseThreshold int    `json:"failure_pause_threshold"`
	RetentionDays         int    `json:"retention_days"`
	ReasoningEffort       string `json:"reasoning_effort"`
}
type PelicanResult struct {
	ReasoningEffort *string    `json:"reasoning_effort"`
	ID              int64      `json:"id"`
	PlanID          int64      `json:"plan_id"`
	GroupID         int64      `json:"group_id"`
	AccountID       int64      `json:"account_id"`
	ModelID         string     `json:"model_id"`
	PromptVersion   string     `json:"prompt_version"`
	Status          string     `json:"status"`
	ErrorMessage    string     `json:"error_message"`
	LatencyMS       int64      `json:"latency_ms"`
	CharCount       int        `json:"char_count"`
	MinChars        int        `json:"min_chars"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	HTML            string     `json:"html,omitempty"`
}
type PelicanEntry struct {
	ReasoningEffort *string    `json:"reasoning_effort"`
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
	Finalize(context.Context, int64, int64) error
	// Release is retained for existing repository consumers; new runners use
	// Finalize so round counters are settled atomically.
	Release(context.Context, int64, int64) error
	CanRunAccount(context.Context, int64, int64, int64) (bool, error)
	ReserveAttempt(context.Context, int64, int64) error
	Resume(context.Context, int64) (*PelicanPlan, error)
	Cleanup(context.Context, int64, string) (int64, error)
	CleanupDue(context.Context) error
	ListGallery(context.Context, PelicanAuthorization, int64, int64, int, int) (*PelicanListResponse, error)
	ListHistory(context.Context, PelicanAuthorization, int64, int64) ([]PelicanResult, error)
	GetResult(context.Context, PelicanAuthorization, int64) (*PelicanResult, error)
	SaveResult(context.Context, *PelicanResult, int, int64) error
}
type PelicanTestExecutor interface {
	RunPelicanTest(context.Context, int64, string, string, string) (*PelicanResult, error)
}
type PelicanAccountSelector interface {
	SelectPelicanAccount(context.Context, int64, string) (*AccountSelectionResult, error)
}

type pelicanGatewaySelector struct{ gateway *OpenAIGatewayService }

func (s pelicanGatewaySelector) SelectPelicanAccount(ctx context.Context, groupID int64, model string) (*AccountSelectionResult, error) {
	selection, _, err := s.gateway.SelectAccountWithScheduler(ctx, &groupID, "", "", model, nil, OpenAIUpstreamTransportHTTPSSE, false)
	return selection, err
}

type PelicanTestService struct {
	repo     PelicanTestRepository
	accounts AccountRepository
	groups   GroupRepository
	tester   PelicanTestExecutor
	selector PelicanAccountSelector
	settings SettingRepository
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
func (s *PelicanTestService) SetAccountSelector(selector PelicanAccountSelector) {
	s.selector = selector
}
func (s *PelicanTestService) SetSettingsRepository(repo SettingRepository) { s.settings = repo }

func defaultPelicanTestSettings() PelicanTestSettings {
	return PelicanTestSettings{DisplayName: "鹈鹕测试", Prompt: PelicanPrompt, Enabled: true}
}

func validatePelicanTestSettings(settings PelicanTestSettings) error {
	if strings.TrimSpace(settings.DisplayName) == "" || utf8.RuneCountInString(settings.DisplayName) > 40 || strings.TrimSpace(settings.Prompt) == "" || utf8.RuneCountInString(settings.Prompt) > 20000 {
		return ErrPelicanInvalid
	}
	return nil
}

func (s *PelicanTestService) GetSettings(ctx context.Context) (PelicanTestSettings, error) {
	defaults := defaultPelicanTestSettings()
	if s.settings == nil {
		return PelicanTestSettings{}, errors.New("pelican settings repository unavailable")
	}
	raw, err := s.settings.GetValue(ctx, PelicanTestSettingsKey)
	if errors.Is(err, ErrSettingNotFound) {
		return defaults, nil
	}
	if err != nil {
		return PelicanTestSettings{}, fmt.Errorf("get pelican test settings: %w", err)
	}
	// Seed defaults before decoding so legacy JSON with no enabled property is
	// compatible and resolves to enabled=true.
	settings := defaults
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return PelicanTestSettings{}, fmt.Errorf("parse pelican test settings: %w", err)
	}
	if err := validatePelicanTestSettings(settings); err != nil {
		return PelicanTestSettings{}, fmt.Errorf("invalid persisted pelican test settings")
	}
	return settings, nil
}

func (s *PelicanTestService) UpdateSettings(ctx context.Context, input PelicanTestSettingsUpdate) (PelicanTestSettings, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return PelicanTestSettings{}, err
	}
	settings.DisplayName = strings.TrimSpace(input.DisplayName)
	settings.Prompt = input.Prompt
	if input.Enabled != nil {
		settings.Enabled = *input.Enabled
	}
	if err := validatePelicanTestSettings(settings); err != nil {
		return PelicanTestSettings{}, err
	}
	if s.settings == nil {
		return PelicanTestSettings{}, errors.New("pelican settings repository unavailable")
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return PelicanTestSettings{}, fmt.Errorf("encode pelican test settings: %w", err)
	}
	if err := s.settings.Set(ctx, PelicanTestSettingsKey, string(raw)); err != nil {
		return PelicanTestSettings{}, fmt.Errorf("save pelican test settings: %w", err)
	}
	return settings, nil
}

func (s *PelicanTestService) Metadata(ctx context.Context) (PelicanTestMetadata, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return PelicanTestMetadata{}, err
	}
	return PelicanTestMetadata{DisplayName: settings.DisplayName, Enabled: settings.Enabled}, nil
}

func (s *PelicanTestService) promptSnapshot(ctx context.Context) (pelicanPromptSnapshot, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return pelicanPromptSnapshot{}, err
	}
	if settings.Prompt == PelicanPrompt {
		return pelicanPromptSnapshot{prompt: settings.Prompt, version: PelicanPromptVersion}, nil
	}
	digest := sha256.Sum256([]byte(settings.Prompt))
	return pelicanPromptSnapshot{prompt: settings.Prompt, version: "pelican-" + hex.EncodeToString(digest[:])}, nil
}
func (s *PelicanTestService) allowUserAccess(ctx context.Context, auth PelicanAuthorization) error {
	if auth.Admin {
		return nil
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return ErrPelicanDisabled
	}
	return nil
}
func (s *PelicanTestService) ListGallery(c context.Context, a PelicanAuthorization, g, ac int64, p, ps int) (*PelicanListResponse, error) {
	if err := s.allowUserAccess(c, a); err != nil {
		return nil, err
	}
	return s.repo.ListGallery(c, a, g, ac, p, ps)
}
func (s *PelicanTestService) ListHistory(c context.Context, a PelicanAuthorization, p, ac int64) ([]PelicanResult, error) {
	if err := s.allowUserAccess(c, a); err != nil {
		return nil, err
	}
	return s.repo.ListHistory(c, a, p, ac)
}
func (s *PelicanTestService) GetResult(c context.Context, a PelicanAuthorization, id int64) (*PelicanResult, error) {
	if err := s.allowUserAccess(c, a); err != nil {
		return nil, err
	}
	return s.repo.GetResult(c, a, id)
}
func (s *PelicanTestService) ListPlans(c context.Context) ([]PelicanPlan, error) {
	plans, err := s.repo.ListPlans(c)
	if err != nil {
		return nil, err
	}
	for i := range plans {
		accounts, listErr := s.accounts.ListByGroup(c, plans[i].GroupID)
		if listErr != nil {
			return nil, listErr
		}
		group, groupErr := s.groups.GetByIDLite(c, plans[i].GroupID)
		if groupErr != nil && !errors.Is(groupErr, ErrGroupNotFound) {
			return nil, groupErr
		}
		if group == nil {
			continue
		}
		for j := range accounts {
			a := &accounts[j]
			member := false
			for _, groupID := range a.GroupIDs {
				if groupID == plans[i].GroupID {
					member = true
					break
				}
			}
			if pelicanSkipReason(a, nil, group, nil, member) == "" && a.IsModelSupported(plans[i].ModelID) && pelicanHTMLModel(a.GetMappedModel(plans[i].ModelID)) {
				plans[i].EstimatedCallsPerRun = 1
				break
			}
		}
	}
	return plans, nil
}

// Models returns locally configured, HTML-capable IDs for an active OpenAI
// group. It reads repositories only, so transient scheduling state cannot erase
// configured choices and no provider is contacted.
func (s *PelicanTestService) Models(c context.Context, groupID int64) ([]string, error) {
	models, _, err := s.modelsForGroup(c, groupID)
	return models, err
}

func (s *PelicanTestService) modelsForGroup(c context.Context, groupID int64) ([]string, *Group, error) {
	if groupID <= 0 || s.groups == nil || s.accounts == nil {
		return nil, nil, ErrPelicanInvalid
	}
	group, err := s.groups.GetByIDLite(c, groupID)
	if err != nil {
		return nil, nil, err
	}
	if group == nil || !group.IsActive() || !strings.EqualFold(group.Platform, PlatformOpenAI) {
		return nil, nil, ErrPelicanInvalid
	}
	accounts, err := s.accounts.ListByGroup(c, groupID)
	if err != nil {
		return nil, nil, err
	}
	candidates := make(map[string]struct{})
	for index := range accounts {
		account := &accounts[index]
		if !pelicanModelAccount(account) {
			continue
		}
		mapping := account.GetModelMapping()
		allowDefaults := len(mapping) == 0 || account.IsOpenAIPassthroughEnabled()
		for requested, mapped := range mapping {
			requested = strings.TrimSpace(requested)
			if strings.Contains(requested, "*") {
				allowDefaults = true
				continue
			}
			if pelicanHTMLModel(requested) && pelicanHTMLModel(mapped) {
				candidates[requested] = struct{}{}
			}
		}
		if allowDefaults {
			for _, model := range openai.DefaultModelIDs() {
				if pelicanHTMLModel(model) {
					candidates[model] = struct{}{}
				}
			}
		}
		if group.CustomModelsListEnabled() {
			for _, model := range group.ModelsListConfig.Models {
				if model = strings.TrimSpace(model); pelicanHTMLModel(model) {
					candidates[model] = struct{}{}
				}
			}
		}
	}
	if group.CustomModelsListEnabled() {
		allowed := make(map[string]struct{}, len(group.ModelsListConfig.Models))
		for _, model := range group.ModelsListConfig.Models {
			if model = strings.TrimSpace(model); pelicanHTMLModel(model) {
				allowed[model] = struct{}{}
			}
		}
		for model := range candidates {
			if _, ok := allowed[model]; !ok {
				delete(candidates, model)
			}
		}
	}
	models := make([]string, 0, len(candidates))
	for model := range candidates {
		for index := range accounts {
			account := &accounts[index]
			if pelicanModelAccount(account) && account.IsModelSupported(model) && pelicanHTMLModel(account.GetMappedModel(model)) {
				models = append(models, model)
				break
			}
		}
	}
	sort.Strings(models)
	return models, group, nil
}

func pelicanModelAccount(account *Account) bool {
	return account != nil && account.IsOpenAI() && (account.Type == AccountTypeAPIKey || account.IsOAuth()) && !account.IsOpenAIAgentIdentity() && !account.IsShadow()
}

func pelicanHTMLModel(model string) bool {
	model = strings.TrimSpace(model)
	if model == "" || strings.Contains(model, "*") || isOpenAIImageModel(model) {
		return false
	}
	lower := strings.ToLower(model)
	for _, prefix := range []string{"gpt-image-", "dall-e", "sora", "text-embedding-", "embedding-", "tts-", "whisper", "omni-moderation", "moderation", "realtime", "gpt-realtime", "gpt-audio", "gpt-4o-mini-tts", "gpt-4o-transcribe", "gpt-4o-mini-transcribe"} {
		if strings.HasPrefix(lower, prefix) {
			return false
		}
	}
	return true
}

func (s *PelicanTestService) validPlanModel(c context.Context, groupID int64, model string) (*Group, error) {
	if !pelicanHTMLModel(model) {
		return nil, ErrPelicanInvalid
	}
	models, group, err := s.modelsForGroup(c, groupID)
	if err != nil {
		return nil, err
	}
	for _, allowed := range models {
		if allowed == model {
			return group, nil
		}
	}
	return nil, ErrPelicanInvalid
}
func pelicanTestTimeout(seconds int) time.Duration {
	if seconds == 0 {
		seconds = 180
	}
	return time.Duration(seconds) * time.Second
}

func validPelicanInput(i PelicanPlanInput) error {
	if (i.TimeoutSeconds != 0 && (i.TimeoutSeconds < 30 || i.TimeoutSeconds > 3600)) || i.GroupID <= 0 || strings.TrimSpace(i.ModelID) == "" || len(i.ModelID) > 200 || isOpenAIImageModel(i.ModelID) || strings.ContainsAny(i.ModelID, "\r\n\x00") || i.IntervalMinutes < 15 || i.IntervalMinutes > 1440 || i.MaxResults < 1 || i.MaxResults > 50 || i.MinChars < 100 || i.MinChars > 50000 || i.DailyCallLimit < 0 || i.DailyCallLimit > 100000 || i.FailurePauseThreshold < 0 || i.FailurePauseThreshold > 100 || i.RetentionDays < 0 || i.RetentionDays > 3650 || !validPelicanReasoningEffort(i.ReasoningEffort) {
		return ErrPelicanInvalid
	}
	return nil
}

func validPelicanReasoningEffort(effort string) bool {
	switch effort {
	case "", "none", "minimal", "low", "medium", "high", "xhigh":
		return true
	default:
		return false
	}
}
func (s *PelicanTestService) CreatePlan(c context.Context, i PelicanPlanInput) (*PelicanPlan, error) {
	if e := validPelicanInput(i); e != nil {
		return nil, e
	}
	g, e := s.validPlanModel(c, i.GroupID, i.ModelID)
	if e != nil {
		return nil, e
	}
	i.Enabled = false
	return s.repo.CreatePlan(c, i, g.Name)
}
func (s *PelicanTestService) UpdatePlan(c context.Context, id int64, i PelicanPlanInput) (*PelicanPlan, error) {
	if e := validPelicanInput(i); e != nil {
		return nil, e
	}
	g, e := s.validPlanModel(c, i.GroupID, i.ModelID)
	if e != nil {
		return nil, e
	}
	return s.repo.UpdatePlan(c, id, i, g.Name)
}
func (s *PelicanTestService) DeletePlan(c context.Context, id int64) error {
	return s.repo.DeletePlan(c, id)
}
func (s *PelicanTestService) Resume(c context.Context, id int64) (*PelicanPlan, error) {
	return s.repo.Resume(c, id)
}
func (s *PelicanTestService) Cleanup(c context.Context, id int64, scope string) (int64, error) {
	if scope != "failed" && scope != "expired" {
		return 0, ErrPelicanInvalid
	}
	return s.repo.Cleanup(c, id, scope)
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
	settings, e := s.GetSettings(claimCtx)
	if e != nil {
		<-s.planGate
		s.wg.Done()
		return e
	}
	if !settings.Enabled {
		<-s.planGate
		s.wg.Done()
		return ErrPelicanDisabled
	}
	snapshot := pelicanPromptSnapshot{prompt: settings.Prompt, version: PelicanPromptVersion}
	if settings.Prompt != PelicanPrompt {
		digest := sha256.Sum256([]byte(settings.Prompt))
		snapshot.version = "pelican-" + hex.EncodeToString(digest[:])
	}
	p, e := s.repo.Claim(claimCtx, id, manual, time.Now())
	if e != nil {
		<-s.planGate
		s.wg.Done()
		return e
	}
	go func() { defer s.wg.Done(); s.run(runCtx, p, snapshot) }()
	return nil
}
func (s *PelicanTestService) run(c context.Context, p *PelicanPlan, snapshot pelicanPromptSnapshot) {
	defer func() { <-s.planGate }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.Finalize(ctx, p.ID, p.RunGeneration); err != nil {
			log.Printf("pelican plan %d: finalize failed", p.ID)
		}
	}()
	timeout := pelicanTestTimeout(p.TimeoutSeconds)
	planCtx, cancelPlan := context.WithTimeout(c, timeout+time.Minute)
	defer cancelPlan()
	started := time.Now()
	if s.selector == nil {
		s.save(planCtx, p, 0, nil, "skipped", "group_scheduler_unavailable", started, snapshot)
		return
	}
	selection, err := s.selector.SelectPelicanAccount(planCtx, p.GroupID, p.ModelID)
	if selection != nil && selection.ReleaseFunc != nil {
		defer selection.ReleaseFunc()
	}
	if err != nil || selection == nil || selection.Account == nil {
		s.save(planCtx, p, 0, nil, "skipped", "group_scheduler_unavailable", started, snapshot)
		return
	}
	if !selection.Acquired {
		s.save(planCtx, p, 0, nil, "skipped", "group_busy", started, snapshot)
		return
	}
	a := selection.Account
	accountCtx, cancel := context.WithTimeout(planCtx, timeout)
	defer cancel()
	fresh, err := s.accounts.GetByID(accountCtx, a.ID)
	group, groupErr := s.groups.GetByIDLite(accountCtx, p.GroupID)
	member := fresh != nil && containsInt64(fresh.GroupIDs, p.GroupID)
	if reason := pelicanSkipReason(fresh, err, group, groupErr, member); reason != "" {
		s.save(accountCtx, p, a.ID, nil, "skipped", reason, started, snapshot)
		return
	}
	if !fresh.IsModelSupported(p.ModelID) {
		s.save(accountCtx, p, a.ID, nil, "skipped", "account_model_not_allowed", started, snapshot)
		return
	}
	if !pelicanHTMLModel(fresh.GetMappedModel(p.ModelID)) {
		s.save(accountCtx, p, a.ID, nil, "skipped", "model_not_html_capable", started, snapshot)
		return
	}
	allowed, checkErr := s.repo.CanRunAccount(accountCtx, p.ID, p.RunGeneration, a.ID)
	if checkErr != nil || !allowed || accountCtx.Err() != nil {
		s.save(accountCtx, p, a.ID, nil, "skipped", "plan_changed", started, snapshot)
		return
	}
	if err := s.repo.ReserveAttempt(accountCtx, p.ID, p.RunGeneration); err != nil {
		reason := "daily_call_limit_reached"
		if !errors.Is(err, ErrPelicanQuota) {
			reason = "quota_reservation_failed"
		}
		s.save(accountCtx, p, a.ID, nil, "skipped", reason, started, snapshot)
		return
	}
	x, err := s.tester.RunPelicanTest(accountCtx, a.ID, p.ModelID, snapshot.prompt, p.ReasoningEffort)
	if err != nil || accountCtx.Err() != nil || x == nil {
		s.save(accountCtx, p, a.ID, nil, "failed", "generation failed or timed out", started, snapshot)
		return
	}
	s.save(accountCtx, p, a.ID, x, "success", "", started, snapshot)
}
func pelicanSkipReason(account *Account, accountErr error, group *Group, groupErr error, member bool) string {
	if accountErr != nil || account == nil {
		return "account_lookup_failed"
	}
	if groupErr != nil || group == nil || !group.IsActive() || !strings.EqualFold(group.Platform, PlatformOpenAI) {
		return "group_unavailable"
	}
	if !member {
		return "account_not_in_group"
	}
	if !pelicanModelAccount(account) {
		return "account_unsupported"
	}
	if !account.IsActive() {
		return "account_inactive"
	}
	if !account.Schedulable {
		return "account_scheduling_disabled"
	}
	now := time.Now()
	if account.ExpiresAt != nil && !account.ExpiresAt.After(now) {
		return "account_expired"
	}
	if account.IsRateLimited() {
		return "account_rate_limited"
	}
	if account.IsOverloaded() || (account.TempUnschedulableUntil != nil && now.Before(*account.TempUnschedulableUntil)) {
		return "account_cooldown"
	}
	if !account.IsAvailableAt(now) {
		return "account_outside_schedule"
	}
	if account.IsAPIKeyOrBedrock() && account.IsQuotaExceeded() {
		return "account_quota_exceeded"
	}
	if !account.IsSchedulable() {
		return "account_unavailable"
	}
	return ""
}

// pelicanAccountAvailable is also used by the provider adapter before it
// constructs an upstream request. Keep that paid-call guard aligned with the
// runner's classified skip reasons.
func pelicanAccountAvailable(account *Account) bool {
	return pelicanSkipReason(account, nil, &Group{Platform: PlatformOpenAI, Status: StatusActive}, nil, true) == ""
}
func (s *PelicanTestService) save(c context.Context, p *PelicanPlan, id int64, x *PelicanResult, status, msg string, started time.Time, snapshot pelicanPromptSnapshot) {
	r := PelicanResult{PlanID: p.ID, GroupID: p.GroupID, AccountID: id, ModelID: p.ModelID, PromptVersion: snapshot.version, Status: status, ErrorMessage: msg, MinChars: p.MinChars, StartedAt: started, LatencyMS: time.Since(started).Milliseconds()}
	finished := time.Now()
	r.ReasoningEffort = &p.ReasoningEffort
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
		// Run once on startup so disabled plans are retained promptly; the
		// separate bounded context also keeps shutdown responsive.
		cleanupCtx, cleanupCancel := context.WithTimeout(ctx, 30*time.Second)
		if err := s.repo.CleanupDue(cleanupCtx); err != nil {
			log.Print("pelican cleanup: failed")
		}
		cleanupCancel()
		tick := time.NewTicker(30 * time.Second)
		defer tick.Stop()
		cleanupTick := time.NewTicker(5 * time.Minute)
		defer cleanupTick.Stop()
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
			case <-cleanupTick.C:
				cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				if err := s.repo.CleanupDue(cleanupCtx); err != nil {
					log.Print("pelican cleanup: failed")
				}
				cancel()
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

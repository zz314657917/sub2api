package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type UserOwnedAccountRepository interface {
	AccountRepository
	ListUserOwned(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error)
	CountUserOwned(ctx context.Context, userID int64) (int64, error)
}

type UserAccountShareLedgerRepository interface {
	ListShareSummary(ctx context.Context, ownerUserID int64) (*UserAccountShareSummary, error)
	GetUsageSummary(ctx context.Context, ownerUserID int64, startTime, endTime time.Time) (*UserAccountUsageSummary, error)
	TransferAvailableShareToBalance(ctx context.Context, ownerUserID int64) (float64, float64, error)
}

var (
	ErrUserAccountNotOwned      = infraerrors.Forbidden("USER_ACCOUNT_NOT_OWNED", "account does not belong to current user")
	ErrUserAccountShareInvalid  = infraerrors.BadRequest("USER_ACCOUNT_SHARE_INVALID", "invalid account share state")
	ErrUserAccountShareDisabled = infraerrors.Forbidden("USER_ACCOUNT_SHARE_DISABLED", "account sharing is disabled")
	ErrUserAccountLimitReached  = infraerrors.Conflict("USER_ACCOUNT_LIMIT_REACHED", "account limit reached")
	ErrUserAccountAPIKeyBlocked = infraerrors.Forbidden("USER_ACCOUNT_APIKEY_BLOCKED", "API Key and custom upstream account uploads are disabled for normal users")
)

const userAccountDefaultConcurrency = 3

type UserAccountShareSummary struct {
	OwnerUserID       int64   `json:"owner_user_id"`
	FrozenAmount      float64 `json:"frozen_amount"`
	AvailableAmount   float64 `json:"available_amount"`
	TransferredAmount float64 `json:"transferred_amount"`
	TotalAmount       float64 `json:"total_amount"`
	CountFrozen       int64   `json:"count_frozen"`
	CountAvailable    int64   `json:"count_available"`
	CountTransferred  int64   `json:"count_transferred"`
}

type UserAccountUsageSummary struct {
	OwnerUserID             int64   `json:"owner_user_id"`
	StartDate               string  `json:"start_date"`
	EndDate                 string  `json:"end_date"`
	TotalAccounts           int64   `json:"total_accounts"`
	PrivateAccounts         int64   `json:"private_accounts"`
	PublicPendingAccounts   int64   `json:"public_pending_accounts"`
	PublicActiveAccounts    int64   `json:"public_active_accounts"`
	PublicSuspendedAccounts int64   `json:"public_suspended_accounts"`
	OwnUsageCost            float64 `json:"own_usage_cost"`
	OwnUsageRequests        int64   `json:"own_usage_requests"`
	SharedUsageCost         float64 `json:"shared_usage_cost"`
	SharedUsageRequests     int64   `json:"shared_usage_requests"`
	ShareIncome             float64 `json:"share_income"`
	PlatformAmount          float64 `json:"platform_amount"`
	AccountCost             float64 `json:"account_cost"`
	BalanceDeduction        float64 `json:"balance_deduction"`
	BalanceNetChange        float64 `json:"balance_net_change"`
}

type UserAccountCapacityPools struct {
	Mine     UserAccountCapacityPool  `json:"mine"`
	Shared   UserAccountCapacityPool  `json:"shared"`
	External *UserAccountCapacityPool `json:"external,omitempty"`
}

type UserAccountCapacityPool struct {
	Key                    string                           `json:"key"`
	Title                  string                           `json:"title"`
	TotalAccounts          int                              `json:"total_accounts"`
	ActiveAccounts         int                              `json:"active_accounts"`
	SchedulableAccounts    int                              `json:"schedulable_accounts"`
	OwnContributedAccounts int                              `json:"own_contributed_accounts,omitempty"`
	RateLimitedAccounts    int                              `json:"rate_limited_accounts"`
	ErrorAccounts          int                              `json:"error_accounts"`
	DisabledAccounts       int                              `json:"disabled_accounts"`
	AbnormalAccounts       int                              `json:"abnormal_accounts"`
	ConfiguredQuota        float64                          `json:"configured_quota"`
	RemainingQuota         float64                          `json:"remaining_quota"`
	PercentOnlyQuota       bool                             `json:"percent_only_quota,omitempty"`
	UnavailableReasons     map[string]int                   `json:"unavailable_reasons,omitempty"`
	Sections               []UserAccountCapacityPoolSection `json:"sections"`
	Groups                 []UserAccountCapacityPoolGroup   `json:"groups,omitempty"`
}

type UserAccountCapacityPoolSection struct {
	Platform               string                                       `json:"platform"`
	Type                   string                                       `json:"type"`
	TotalAccounts          int                                          `json:"total_accounts"`
	SchedulableAccounts    int                                          `json:"schedulable_accounts"`
	OwnContributedAccounts int                                          `json:"own_contributed_accounts,omitempty"`
	ConfiguredQuota        float64                                      `json:"configured_quota"`
	RemainingQuota         float64                                      `json:"remaining_quota"`
	PercentOnlyQuota       bool                                         `json:"percent_only_quota,omitempty"`
	UnavailableReasons     map[string]int                               `json:"unavailable_reasons,omitempty"`
	Windows                map[string]UserAccountCapacityWindowSnapshot `json:"windows,omitempty"`
}

type UserAccountCapacityWindowSnapshot struct {
	UsedPercent       float64 `json:"used_percent"`
	ResetAfterSeconds int     `json:"reset_after_seconds,omitempty"`
	ResetAt           string  `json:"reset_at,omitempty"`
	WindowMinutes     int     `json:"window_minutes,omitempty"`
	UsedAmount        float64 `json:"-"`
	LimitAmount       float64 `json:"-"`
}

type UserAccountCapacityPoolGroup struct {
	Key                    string                                      `json:"key"`
	GroupID                *int64                                      `json:"group_id,omitempty"`
	GroupName              string                                      `json:"group_name"`
	Platform               string                                      `json:"platform,omitempty"`
	SortOrder              int                                         `json:"sort_order,omitempty"`
	TotalAccounts          int                                         `json:"total_accounts"`
	ActiveAccounts         int                                         `json:"active_accounts"`
	SchedulableAccounts    int                                         `json:"schedulable_accounts"`
	OwnContributedAccounts int                                         `json:"own_contributed_accounts,omitempty"`
	RateLimitedAccounts    int                                         `json:"rate_limited_accounts"`
	ErrorAccounts          int                                         `json:"error_accounts"`
	DisabledAccounts       int                                         `json:"disabled_accounts"`
	AbnormalAccounts       int                                         `json:"abnormal_accounts"`
	ConfiguredQuota        float64                                     `json:"configured_quota"`
	RemainingQuota         float64                                     `json:"remaining_quota"`
	PercentOnlyQuota       bool                                        `json:"percent_only_quota,omitempty"`
	UnavailableReasons     map[string]int                              `json:"unavailable_reasons,omitempty"`
	Status                 string                                      `json:"status"`
	Windows                map[string]UserAccountCapacityWindowSummary `json:"windows,omitempty"`
}

type UserAccountCapacityWindowSummary struct {
	UsedPercent                 float64 `json:"used_percent"`
	ResetAfterSeconds           int     `json:"reset_after_seconds,omitempty"`
	ResetAt                     string  `json:"reset_at,omitempty"`
	WindowMinutes               int     `json:"window_minutes,omitempty"`
	SnapshotAccounts            int     `json:"snapshot_accounts"`
	SchedulableSnapshotAccounts int     `json:"schedulable_snapshot_accounts"`
	RemainingUnits              float64 `json:"remaining_units"`
	UsedAmount                  float64 `json:"-"`
	LimitAmount                 float64 `json:"-"`
	AmountPercentWeight         float64 `json:"-"`
	AmountPercentUsedTotal      float64 `json:"-"`
	PercentWeight               float64 `json:"-"`
	PercentUsedTotal            float64 `json:"-"`
}

type UserAccountService struct {
	accountRepo             AccountRepository
	proxyRepo               ProxyRepository
	settings                UserAccountShareSettings
	externalCapacityCacheMu sync.Mutex
	externalCapacityCache   externalCapacityPoolCache
}

type externalCapacityPoolCache struct {
	pool      *UserAccountCapacityPool
	expiresAt time.Time
}

type accountCapacityDisplayWindows struct {
	fiveHour   UserAccountCapacityWindowSnapshot
	has5h      bool
	sevenDay   UserAccountCapacityWindowSnapshot
	has7d      bool
	dailyQuota UserAccountCapacityWindowSnapshot
	hasDaily   bool
	weekQuota  UserAccountCapacityWindowSnapshot
	hasWeek    bool
	monthQuota UserAccountCapacityWindowSnapshot
	hasMonth   bool
}

type accountCapacityDisplayUsageDelta struct {
	fiveHourCost float64
	sevenDayCost float64
}

type accountCapacityDisplayUsageWindow struct {
	accountID int64
	suffix    string
	startTime time.Time
}

type accountCapacityDisplayCounts struct {
	total          int
	active         int
	schedulable    int
	ownContributed int
	rateLimited    int
	error          int
	disabled       int
	abnormal       int
}

type UserAccountShareSettings interface {
	IsAccountShareEnabled(ctx context.Context) bool
	IsExternalCapacityReferenceEnabled(ctx context.Context) bool
	IsAccountShareAutoReview(ctx context.Context) bool
	GetAccountShareUserAccountLimit(ctx context.Context) int
}

type accountUsageCostWindowRepository interface {
	GetAccountUsageCostsSince(ctx context.Context, accountIDs []int64, startTime time.Time) (map[int64]float64, error)
}

type accountUsageCostWindowBatchRepository interface {
	GetAccountUsageCostsSinceByWindow(ctx context.Context, windows []AccountUsageCostWindowRequest) (map[AccountUsageCostWindowRequestKey]float64, error)
}

type AccountUsageCostWindowRequest struct {
	AccountID int64
	Suffix    string
	StartTime time.Time
}

type AccountUsageCostWindowRequestKey struct {
	AccountID int64
	Suffix    string
}

func NewUserAccountService(accountRepo AccountRepository, settings UserAccountShareSettings) *UserAccountService {
	return &UserAccountService{
		accountRepo: accountRepo,
		settings:    settings,
	}
}

func (s *UserAccountService) SetProxyRepository(proxyRepo ProxyRepository) {
	if s != nil {
		s.proxyRepo = proxyRepo
	}
}

func (s *UserAccountService) ValidateProxyForUser(ctx context.Context, userID int64, proxyID *int64) (*int64, error) {
	return s.normalizeUserProxyID(ctx, userID, proxyID)
}

func isUserUploadedAPIKeyAccount(accountType string) bool {
	switch strings.ToLower(strings.TrimSpace(accountType)) {
	case AccountTypeAPIKey, AccountTypeUpstream:
		return true
	default:
		return false
	}
}

func containsUserUploadedAPIKeyCredential(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		for key, nested := range v {
			if isUserUploadedAPIKeyCredentialKey(key) || containsUserUploadedAPIKeyCredential(nested) {
				return true
			}
		}
	case map[string]string:
		for key := range v {
			if isUserUploadedAPIKeyCredentialKey(key) {
				return true
			}
		}
	case []any:
		for _, nested := range v {
			if containsUserUploadedAPIKeyCredential(nested) {
				return true
			}
		}
	}
	return false
}

func isUserUploadedAPIKeyCredentialKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "api_key", "apikey", "x_api_key", "xapikey":
		return true
	default:
		return false
	}
}

func (s *UserAccountService) List(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, nil, err
	}
	repo, err := s.ownedAccountRepo()
	if err != nil {
		return nil, nil, err
	}
	return repo.ListUserOwned(ctx, userID, params)
}

func (s *UserAccountService) Count(ctx context.Context, userID int64) (int64, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return 0, err
	}
	repo, err := s.ownedAccountRepo()
	if err != nil {
		return 0, err
	}
	return repo.CountUserOwned(ctx, userID)
}

func (s *UserAccountService) Create(ctx context.Context, userID int64, req CreateAccountRequest) (*Account, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	if isUserUploadedAPIKeyAccount(req.Type) || containsUserUploadedAPIKeyCredential(req.Credentials) {
		return nil, ErrUserAccountAPIKeyBlocked
	}
	if err := ValidateAccountAvailabilityConfig(req.Extra); err != nil {
		return nil, err
	}
	repo, err := s.ownedAccountRepo()
	if err != nil {
		return nil, err
	}
	if s.settings != nil {
		limit := s.settings.GetAccountShareUserAccountLimit(ctx)
		if limit > 0 {
			count, err := repo.CountUserOwned(ctx, userID)
			if err != nil {
				return nil, err
			}
			if count >= int64(limit) {
				return nil, ErrUserAccountLimitReached
			}
		}
	}
	account := &Account{
		Name:        req.Name,
		Notes:       normalizeAccountNotes(req.Notes),
		Platform:    req.Platform,
		Type:        req.Type,
		Credentials: req.Credentials,
		Extra:       req.Extra,
		ProxyID:     nil,
		Concurrency: normalizeUserAccountConcurrency(req.Concurrency),
		Priority:    0,
		Status:      StatusActive,
		Schedulable: true,
		OwnerUserID: &userID,
		ShareMode:   AccountShareModePrivate,
		ShareStatus: AccountShareStatusNotShared,
	}
	if req.AutoPauseOnExpired != nil {
		account.AutoPauseOnExpired = *req.AutoPauseOnExpired
	} else {
		account.AutoPauseOnExpired = true
	}
	if req.ExpiresAt != nil {
		account.ExpiresAt = req.ExpiresAt
	}
	if req.ProxyID != nil {
		proxyID, err := s.normalizeUserProxyID(ctx, userID, req.ProxyID)
		if err != nil {
			return nil, err
		}
		account.ProxyID = proxyID
	}
	if err := repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *UserAccountService) GetByID(ctx context.Context, userID, accountID int64) (*Account, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	account, err := s.getOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *UserAccountService) Update(ctx context.Context, userID, accountID int64, req UpdateAccountRequest) (*Account, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	if req.Extra != nil {
		if err := ValidateAccountAvailabilityConfig(*req.Extra); err != nil {
			return nil, err
		}
	}
	account, err := s.getOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Notes != nil {
		account.Notes = normalizeAccountNotes(req.Notes)
	}
	if req.Credentials != nil {
		if isUserUploadedAPIKeyAccount(account.Type) || containsUserUploadedAPIKeyCredential(*req.Credentials) {
			return nil, ErrUserAccountAPIKeyBlocked
		}
		account.Credentials = *req.Credentials
	}
	if req.Extra != nil {
		account.Extra = *req.Extra
	}
	if req.ClearExpiresAt {
		account.ExpiresAt = nil
	} else if req.ExpiresAt != nil {
		account.ExpiresAt = req.ExpiresAt
	}
	if req.AutoPauseOnExpired != nil {
		account.AutoPauseOnExpired = *req.AutoPauseOnExpired
	}
	if req.ProxyID != nil {
		proxyID, err := s.normalizeUserProxyID(ctx, userID, req.ProxyID)
		if err != nil {
			return nil, err
		}
		account.ProxyID = proxyID
	}
	if req.Concurrency != nil {
		account.Concurrency = normalizeUserAccountConcurrency(*req.Concurrency)
	}
	if req.GroupIDs != nil || req.Priority != nil || req.Status != nil {
		return nil, ErrUserAccountShareInvalid
	}

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *UserAccountService) Delete(ctx context.Context, userID, accountID int64) error {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return err
	}
	_, err := s.getOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return err
	}
	repo, err := s.accountRepoForMutation()
	if err != nil {
		return err
	}
	return repo.Delete(ctx, accountID)
}

func (s *UserAccountService) TestCredentials(ctx context.Context, userID, accountID int64) error {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return err
	}
	account, err := s.getOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return err
	}
	switch account.Platform {
	case PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity:
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", account.Platform)
	}
}

func (s *UserAccountService) UpdateShareMode(ctx context.Context, userID, accountID int64, shareMode string) (*Account, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	account, err := s.getOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(strings.TrimSpace(shareMode)) {
	case AccountShareModePrivate:
		account.ShareMode = AccountShareModePrivate
		account.ShareStatus = AccountShareStatusNotShared
	case AccountShareModePublic:
		if s.settings != nil && !s.settings.IsAccountShareEnabled(ctx) {
			return nil, ErrUserAccountShareDisabled
		}
		account.ShareMode = AccountShareModePublic
		autoReview := s.settings != nil && s.settings.IsAccountShareAutoReview(ctx)
		if autoReview {
			account.ShareStatus = AccountShareStatusActive
		} else if account.ShareStatus == "" || account.ShareStatus == AccountShareStatusNotShared {
			account.ShareStatus = AccountShareStatusPendingReview
		}
	default:
		return nil, ErrUserAccountShareInvalid
	}
	repo, err := s.accountRepoForMutation()
	if err != nil {
		return nil, err
	}
	if err := repo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *UserAccountService) UpdateShareStatus(ctx context.Context, userID, accountID int64, shareStatus string) (*Account, error) {
	account, err := s.getOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(strings.TrimSpace(shareStatus)) {
	case AccountShareStatusNotShared, AccountShareStatusPendingReview, AccountShareStatusActive, AccountShareStatusRejected, AccountShareStatusSuspended:
		account.ShareStatus = shareStatus
	default:
		return nil, ErrUserAccountShareInvalid
	}
	repo, err := s.accountRepoForMutation()
	if err != nil {
		return nil, err
	}
	if err := repo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *UserAccountService) GetShareSummary(ctx context.Context, userID int64) (*UserAccountShareSummary, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	repo, err := s.shareLedgerRepo()
	if err != nil {
		return nil, err
	}
	return repo.ListShareSummary(ctx, userID)
}

func (s *UserAccountService) GetUsageSummary(ctx context.Context, userID int64, startTime, endTime time.Time) (*UserAccountUsageSummary, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	repo, err := s.shareLedgerRepo()
	if err != nil {
		return nil, err
	}
	summary, err := repo.GetUsageSummary(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		summary = &UserAccountUsageSummary{OwnerUserID: userID}
	}
	summary.OwnerUserID = userID
	summary.StartDate = startTime.Format("2006-01-02")
	if endTime.After(startTime) {
		summary.EndDate = endTime.AddDate(0, 0, -1).Format("2006-01-02")
	}
	summary.BalanceNetChange = summary.ShareIncome - summary.BalanceDeduction
	return summary, nil
}

func (s *UserAccountService) TransferAvailableShareToBalance(ctx context.Context, userID int64) (float64, float64, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return 0, 0, err
	}
	repo, err := s.shareLedgerRepo()
	if err != nil {
		return 0, 0, err
	}
	return repo.TransferAvailableShareToBalance(ctx, userID)
}

func (s *UserAccountService) GetCapacityPools(ctx context.Context, userID int64) (*UserAccountCapacityPools, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	ownedRepo, err := s.ownedAccountRepo()
	if err != nil {
		return nil, err
	}
	allRepo, err := s.accountRepoForMutation()
	if err != nil {
		return nil, err
	}

	params := userAccountCapacityPaginationParams(1)
	owned, err := listAllUserOwnedAccounts(ctx, ownedRepo, userID, params)
	if err != nil {
		return nil, err
	}
	allAccounts, err := listAllCapacityAccounts(ctx, allRepo, params)
	if err != nil {
		return nil, err
	}

	shared := make([]Account, 0, len(allAccounts))
	for _, account := range allAccounts {
		if isSharedCapacityPoolAccount(&account, userID) {
			shared = append(shared, account)
		}
	}

	displayUsageDeltas, err := s.accountCapacityDisplayUsageDeltas(ctx, append(append([]Account(nil), owned...), shared...), time.Now())
	if err != nil {
		return nil, err
	}

	pools := &UserAccountCapacityPools{
		Mine:   buildUserAccountCapacityPool("mine", "我的账号容量池", owned, userID, displayUsageDeltas),
		Shared: buildUserAccountCapacityPool("shared", "平台共享容量池", shared, userID, displayUsageDeltas),
	}
	if ExternalCapacityReferenceFeatureEnabled && s.settings != nil && s.settings.IsExternalCapacityReferenceEnabled(ctx) {
		pools.External = s.getExternalCapacityPool(ctx, time.Now())
	}
	return pools, nil
}

const (
	externalAIPixelCapacityDefaultURL       = "https://ai-pixel.online/api/v1/accounts/quota-dashboard"
	externalAIPixelCapacityPoolKey          = "public_shared_capacity_reference"
	externalAIPixelCapacityDefaultTokenFile = "/app/data/ai_pixel_capacity_token.txt"
	externalAIPixelCapacityDefaultTimeout   = 8 * time.Second
	externalAIPixelCapacityDefaultTTL       = 60 * time.Second
)

type aiPixelCapacityDashboardEnvelope struct {
	Code    int                          `json:"code"`
	Message string                       `json:"message"`
	Data    aiPixelCapacityDashboardData `json:"data"`
}

type aiPixelCapacityDashboardData struct {
	Platform aiPixelCapacityPlatform `json:"platform"`
}

type aiPixelCapacityPlatform struct {
	GeneratedAt     string                        `json:"generated_at"`
	Totals          aiPixelCapacitySummary        `json:"totals"`
	GroupSummaries  []aiPixelCapacityGroupSummary `json:"group_summaries"`
	Summaries       []aiPixelCapacitySummary      `json:"summaries"`
	UsageWindowRows []aiPixelCapacityUsageWindow  `json:"usage_windows"`
}

type aiPixelCapacitySummary struct {
	Platform                string                       `json:"platform"`
	Type                    string                       `json:"type"`
	AccountCount            int                          `json:"account_count"`
	ActiveAccountCount      int                          `json:"active_account_count"`
	SchedulableAccountCount int                          `json:"schedulable_account_count"`
	RateLimitedAccountCount int                          `json:"rate_limited_account_count"`
	ErrorAccountCount       int                          `json:"error_account_count"`
	DisabledAccountCount    int                          `json:"disabled_account_count"`
	Total                   aiPixelQuotaUsage            `json:"total"`
	Daily                   aiPixelQuotaUsage            `json:"daily"`
	Weekly                  aiPixelQuotaUsage            `json:"weekly"`
	UsageWindows            []aiPixelCapacityUsageWindow `json:"usage_windows"`
}

type aiPixelCapacityGroupSummary struct {
	GroupID                 json.RawMessage              `json:"group_id"`
	GroupName               string                       `json:"group_name"`
	GroupStatus             string                       `json:"group_status"`
	Platform                string                       `json:"platform"`
	AccountCount            int                          `json:"account_count"`
	ActiveAccountCount      int                          `json:"active_account_count"`
	SchedulableAccountCount int                          `json:"schedulable_account_count"`
	RateLimitedAccountCount int                          `json:"rate_limited_account_count"`
	ErrorAccountCount       int                          `json:"error_account_count"`
	DisabledAccountCount    int                          `json:"disabled_account_count"`
	Total                   aiPixelQuotaUsage            `json:"total"`
	Daily                   aiPixelQuotaUsage            `json:"daily"`
	Weekly                  aiPixelQuotaUsage            `json:"weekly"`
	UsageWindows            []aiPixelCapacityUsageWindow `json:"usage_windows"`
}

type aiPixelQuotaUsage struct {
	EnabledAccountCount   int     `json:"enabled_account_count"`
	ExhaustedAccountCount int     `json:"exhausted_account_count"`
	Limit                 float64 `json:"limit"`
	Used                  float64 `json:"used"`
	Remaining             float64 `json:"remaining"`
	Utilization           float64 `json:"utilization"`
}

type aiPixelCapacityUsageWindow struct {
	Window                   string  `json:"window"`
	AccountCount             int     `json:"account_count"`
	KnownAccountCount        int     `json:"known_account_count"`
	AverageUtilization       float64 `json:"average_utilization"`
	RemainingCapacityPercent float64 `json:"remaining_capacity_percent"`
	EstimatedSupportHours    float64 `json:"estimated_support_hours"`
	MinRemainingSeconds      int     `json:"min_remaining_seconds"`
	NextResetAt              string  `json:"next_reset_at"`
}

func (s *UserAccountService) getExternalCapacityPool(ctx context.Context, now time.Time) *UserAccountCapacityPool {
	authorization, ok := externalAIPixelCapacityAuthorization()
	if !ok {
		return nil
	}

	s.externalCapacityCacheMu.Lock()
	if s.externalCapacityCache.pool != nil && now.Before(s.externalCapacityCache.expiresAt) {
		pool := s.externalCapacityCache.pool
		s.externalCapacityCacheMu.Unlock()
		return pool
	}
	s.externalCapacityCacheMu.Unlock()

	pool, err := fetchExternalAIPixelCapacityPool(ctx, authorization)
	if err != nil {
		return nil
	}
	s.externalCapacityCacheMu.Lock()
	s.externalCapacityCache = externalCapacityPoolCache{
		pool:      pool,
		expiresAt: now.Add(externalAIPixelCapacityCacheTTL()),
	}
	s.externalCapacityCacheMu.Unlock()
	return pool
}

func fetchExternalAIPixelCapacityPool(ctx context.Context, authorization string) (*UserAccountCapacityPool, error) {
	endpoint := strings.TrimSpace(os.Getenv("AI_PIXEL_CAPACITY_URL"))
	if endpoint == "" {
		endpoint = externalAIPixelCapacityDefaultURL
	}
	requestCtx, cancel := context.WithTimeout(ctx, externalAIPixelCapacityTimeout())
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", authorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("external AI Pixel capacity request failed: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	var envelope aiPixelCapacityDashboardEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != 0 {
		return nil, fmt.Errorf("external AI Pixel capacity request failed: %s", envelope.Message)
	}
	return buildExternalAIPixelCapacityPool(envelope.Data.Platform), nil
}

func buildExternalAIPixelCapacityPool(platform aiPixelCapacityPlatform) *UserAccountCapacityPool {
	pool := &UserAccountCapacityPool{
		Key:      externalAIPixelCapacityPoolKey,
		Title:    "公开共享容量参考",
		Sections: []UserAccountCapacityPoolSection{},
		Groups:   make([]UserAccountCapacityPoolGroup, 0, len(platform.GroupSummaries)),
	}

	for index, summary := range platform.GroupSummaries {
		if !externalAIPixelCapacityGroupVisible(summary.GroupName) {
			continue
		}
		group := buildExternalAIPixelCapacityGroup(summary, index)
		pool.Groups = append(pool.Groups, group)
		pool.TotalAccounts += group.TotalAccounts
		pool.ActiveAccounts += group.ActiveAccounts
		pool.SchedulableAccounts += group.SchedulableAccounts
		pool.RateLimitedAccounts += group.RateLimitedAccounts
		pool.ErrorAccounts += group.ErrorAccounts
		pool.DisabledAccounts += group.DisabledAccounts
		pool.AbnormalAccounts += group.AbnormalAccounts
		pool.ConfiguredQuota += group.ConfiguredQuota
		pool.RemainingQuota += group.RemainingQuota
	}
	sort.SliceStable(pool.Groups, func(i, j int) bool {
		left := externalAIPixelCapacityGroupRank(pool.Groups[i].GroupName)
		right := externalAIPixelCapacityGroupRank(pool.Groups[j].GroupName)
		if left != right {
			return left < right
		}
		return pool.Groups[i].SortOrder < pool.Groups[j].SortOrder
	})
	for index := range pool.Groups {
		pool.Groups[index].SortOrder = index
	}
	return pool
}

func externalAIPixelCapacityGroupVisible(groupName string) bool {
	return externalAIPixelCapacityGroupRank(groupName) < 2
}

func externalAIPixelCapacityGroupRank(groupName string) int {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(groupName), " ", ""))
	switch {
	case strings.HasPrefix(normalized, "FREE") && strings.Contains(normalized, "共享号池"):
		return 0
	case strings.HasPrefix(normalized, "PLUS") && strings.Contains(normalized, "共享号池"):
		return 1
	default:
		return 99
	}
}

func buildExternalAIPixelCapacityGroup(summary aiPixelCapacityGroupSummary, index int) UserAccountCapacityPoolGroup {
	groupName := strings.TrimSpace(summary.GroupName)
	if groupName == "" {
		groupName = fmt.Sprintf("公开共享分组 %d", index+1)
	}
	windows := externalAIPixelCapacityWindows(summary.UsageWindows)
	group := UserAccountCapacityPoolGroup{
		Key:                 fmt.Sprintf("%s:%s:%s", externalAIPixelCapacityPoolKey, strings.TrimSpace(summary.Platform), strings.ToLower(groupName)),
		GroupName:           groupName,
		Platform:            strings.TrimSpace(summary.Platform),
		SortOrder:           index,
		TotalAccounts:       positiveCapacityCount(summary.AccountCount),
		ActiveAccounts:      positiveCapacityCount(summary.ActiveAccountCount),
		SchedulableAccounts: positiveCapacityCount(summary.SchedulableAccountCount),
		RateLimitedAccounts: positiveCapacityCount(summary.RateLimitedAccountCount),
		ErrorAccounts:       positiveCapacityCount(summary.ErrorAccountCount),
		DisabledAccounts:    positiveCapacityCount(summary.DisabledAccountCount),
		AbnormalAccounts:    positiveCapacityCount(summary.ErrorAccountCount + summary.DisabledAccountCount),
		ConfiguredQuota:     summary.Total.Limit + summary.Daily.Limit + summary.Weekly.Limit,
		RemainingQuota:      summary.Total.Remaining + summary.Daily.Remaining + summary.Weekly.Remaining,
		Windows:             windows,
	}
	if id := externalAIPixelGroupID(summary.GroupID); id != nil {
		group.GroupID = id
	}
	if len(group.Windows) == 0 {
		group.Windows = nil
	}
	group.Status = externalAIPixelCapacityGroupStatus(summary, group.Windows)
	return group
}

func externalAIPixelCapacityWindows(rows []aiPixelCapacityUsageWindow) map[string]UserAccountCapacityWindowSummary {
	if len(rows) == 0 {
		return nil
	}
	windows := make(map[string]UserAccountCapacityWindowSummary, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Window)
		if key == "" {
			continue
		}
		snapshotAccounts := positiveCapacityCount(row.KnownAccountCount)
		if snapshotAccounts == 0 {
			snapshotAccounts = positiveCapacityCount(row.AccountCount)
		}
		windows[key] = UserAccountCapacityWindowSummary{
			UsedPercent:                 clampCapacityPercent(row.AverageUtilization),
			ResetAfterSeconds:           positiveCapacityCount(row.MinRemainingSeconds),
			ResetAt:                     normalizeCapacityWindowTime(row.NextResetAt),
			WindowMinutes:               externalAIPixelWindowMinutes(key),
			SnapshotAccounts:            snapshotAccounts,
			SchedulableSnapshotAccounts: positiveCapacityCount(row.AccountCount),
			RemainingUnits:              maxFloat64(row.RemainingCapacityPercent/100, 0),
		}
	}
	if len(windows) == 0 {
		return nil
	}
	return windows
}

func externalAIPixelCapacityGroupStatus(summary aiPixelCapacityGroupSummary, windows map[string]UserAccountCapacityWindowSummary) string {
	if summary.AccountCount <= 0 || summary.SchedulableAccountCount <= 0 {
		return "unavailable"
	}
	switch strings.ToLower(strings.TrimSpace(summary.GroupStatus)) {
	case "", "active", "normal", "healthy":
	default:
		return "unavailable"
	}
	for _, window := range windows {
		if window.UsedPercent >= 80 {
			return "degraded"
		}
	}
	return "healthy"
}

func externalAIPixelGroupID(raw json.RawMessage) *int64 {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" {
		return nil
	}
	text = strings.Trim(text, `"`)
	var id int64
	if _, err := fmt.Sscan(text, &id); err != nil {
		return nil
	}
	return &id
}

func externalAIPixelWindowMinutes(key string) int {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "5h":
		return int((5 * time.Hour).Minutes())
	case "7d":
		return int((7 * 24 * time.Hour).Minutes())
	default:
		return 0
	}
}

func externalAIPixelCapacityAuthorization() (string, bool) {
	if tokenFile := strings.TrimSpace(os.Getenv("AI_PIXEL_CAPACITY_TOKEN_FILE")); tokenFile != "" {
		content, err := os.ReadFile(tokenFile)
		if err == nil {
			return normalizeExternalAIPixelAuthorization(string(content))
		}
	}
	if content, err := os.ReadFile(externalAIPixelCapacityDefaultTokenFile); err == nil {
		return normalizeExternalAIPixelAuthorization(string(content))
	}
	return normalizeExternalAIPixelAuthorization(os.Getenv("AI_PIXEL_CAPACITY_TOKEN"))
}

func normalizeExternalAIPixelAuthorization(raw string) (string, bool) {
	selected := strings.TrimSpace(raw)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "authorization:") {
			selected = strings.TrimSpace(line[len("authorization:"):])
			break
		}
		selected = line
		if strings.HasPrefix(lower, "bearer ") {
			break
		}
	}
	selected = strings.Trim(strings.TrimSpace(selected), `"'`)
	lower := strings.ToLower(selected)
	if strings.HasPrefix(lower, "authorization:") {
		selected = strings.TrimSpace(selected[len("authorization:"):])
		lower = strings.ToLower(selected)
	}
	if strings.HasPrefix(lower, "bearer ") {
		selected = strings.TrimSpace(selected[len("bearer "):])
	}
	selected = strings.Trim(strings.TrimSpace(selected), `"'`)
	if selected == "" {
		return "", false
	}
	return "Bearer " + selected, true
}

func externalAIPixelCapacityTimeout() time.Duration {
	return externalAIPixelCapacityDefaultTimeout
}

func externalAIPixelCapacityCacheTTL() time.Duration {
	return externalAIPixelCapacityDefaultTTL
}

func positiveCapacityCount(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func clampCapacityPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func (s *UserAccountService) accountCapacityDisplayUsageDeltas(ctx context.Context, accounts []Account, now time.Time) (map[int64]accountCapacityDisplayUsageDelta, error) {
	reader, ok := s.accountRepo.(accountUsageCostWindowRepository)
	if !ok {
		return nil, nil
	}
	windows := shareDisplayUsageWindows(accounts, now)
	if len(windows) == 0 {
		return nil, nil
	}
	deltas := make(map[int64]accountCapacityDisplayUsageDelta, len(windows))
	if reader, ok := s.accountRepo.(accountUsageCostWindowBatchRepository); ok {
		requests := make([]AccountUsageCostWindowRequest, 0, len(windows))
		for _, window := range windows {
			requests = append(requests, AccountUsageCostWindowRequest{
				AccountID: window.accountID,
				Suffix:    window.suffix,
				StartTime: window.startTime,
			})
		}
		costs, err := reader.GetAccountUsageCostsSinceByWindow(ctx, requests)
		if err != nil {
			return nil, fmt.Errorf("load account display usage: %w", err)
		}
		for _, window := range windows {
			delta := deltas[window.accountID]
			switch window.suffix {
			case "5h":
				delta.fiveHourCost = maxFloat64(costs[AccountUsageCostWindowRequestKey{AccountID: window.accountID, Suffix: window.suffix}], 0)
			case "7d":
				delta.sevenDayCost = maxFloat64(costs[AccountUsageCostWindowRequestKey{AccountID: window.accountID, Suffix: window.suffix}], 0)
			}
			deltas[window.accountID] = delta
		}
		return deltas, nil
	}
	for _, window := range windows {
		costs, err := reader.GetAccountUsageCostsSince(ctx, []int64{window.accountID}, window.startTime)
		if err != nil {
			return nil, fmt.Errorf("load %s account display usage: %w", window.suffix, err)
		}
		delta := deltas[window.accountID]
		switch window.suffix {
		case "5h":
			delta.fiveHourCost = maxFloat64(costs[window.accountID], 0)
		case "7d":
			delta.sevenDayCost = maxFloat64(costs[window.accountID], 0)
		}
		deltas[window.accountID] = delta
	}
	return deltas, nil
}

func shareDisplayUsageWindows(accounts []Account, now time.Time) []accountCapacityDisplayUsageWindow {
	seen := make(map[string]struct{})
	windows := make([]accountCapacityDisplayUsageWindow, 0)
	for i := range accounts {
		account := &accounts[i]
		if account.ID <= 0 || !accountUsesShareDisplayWindowMask(account) {
			continue
		}
		appendWindow := func(suffix string, fallbackStart time.Time) {
			if !accountHasShareDisplayWindowLimit(account, suffix) {
				return
			}
			key := fmt.Sprintf("%d:%s", account.ID, suffix)
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}
			start := account.getExtraTime("share_display_" + suffix + "_start")
			if start.IsZero() {
				start = fallbackStart
			}
			windows = append(windows, accountCapacityDisplayUsageWindow{
				accountID: account.ID,
				suffix:    suffix,
				startTime: start,
			})
		}
		appendWindow("5h", now.Add(-5*time.Hour))
		appendWindow("7d", now.Add(-7*24*time.Hour))
	}
	return windows
}

func accountHasShareDisplayWindowLimit(account *Account, suffix string) bool {
	if account == nil || account.Extra == nil {
		return false
	}
	return parseExtraFloat64(account.Extra["share_display_"+suffix+"_limit"]) > 0
}

func isSharedCapacityPoolAccount(account *Account, _ int64) bool {
	if account == nil {
		return false
	}
	if isSystemShareDisplayCapacityAccount(account) {
		return true
	}
	return account.ShareMode == AccountShareModePublic && account.ShareStatus == AccountShareStatusActive
}

func isSystemShareDisplayCapacityAccount(account *Account) bool {
	return account != nil &&
		account.OwnerUserID == nil &&
		account.Platform == PlatformOpenAI &&
		(account.Type == AccountTypeAPIKey || account.Type == AccountTypeOAuth) &&
		accountShareDisplayConfigured(account)
}

func accountShareDisplayConfigured(account *Account) bool {
	if account == nil {
		return false
	}
	return strings.TrimSpace(account.GetShareDisplayName()) != "" ||
		openAIPlanCapacityGroupName(account.GetShareDisplayTier()) != "" ||
		(account.Platform == PlatformOpenAI && openAIPlanCapacityGroupName(account.GetCredential("plan_type")) != "")
}

func userAccountCapacityPaginationParams(page int) pagination.PaginationParams {
	return pagination.PaginationParams{
		Page:      page,
		PageSize:  1000,
		SortBy:    "created_at",
		SortOrder: pagination.SortOrderDesc,
	}
}

func listAllUserOwnedAccounts(ctx context.Context, repo userOwnedAccountRepositoryWithShare, userID int64, firstPage pagination.PaginationParams) ([]Account, error) {
	accounts, result, err := repo.ListUserOwned(ctx, userID, firstPage)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Pages <= firstPage.Page {
		return accounts, nil
	}
	for page := firstPage.Page + 1; page <= result.Pages; page++ {
		next, _, err := repo.ListUserOwned(ctx, userID, userAccountCapacityPaginationParams(page))
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, next...)
	}
	return accounts, nil
}

func listAllCapacityAccounts(ctx context.Context, repo AccountRepository, firstPage pagination.PaginationParams) ([]Account, error) {
	accounts, result, err := repo.List(ctx, firstPage)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Pages <= firstPage.Page {
		return accounts, nil
	}
	for page := firstPage.Page + 1; page <= result.Pages; page++ {
		next, _, err := repo.List(ctx, userAccountCapacityPaginationParams(page))
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, next...)
	}
	return accounts, nil
}

func (s *UserAccountService) UpdateWithShareTransition(ctx context.Context, userID, accountID int64, req UpdateAccountRequest) (*Account, error) {
	if err := s.ensureFeatureEnabled(ctx); err != nil {
		return nil, err
	}
	account, err := s.Update(ctx, userID, accountID, req)
	if err != nil {
		return nil, err
	}
	if account.ShareMode == AccountShareModePublic && account.ShareStatus == AccountShareStatusNotShared {
		account.ShareStatus = AccountShareStatusPendingReview
		repo, err := s.accountRepoForMutation()
		if err != nil {
			return nil, err
		}
		if err := repo.Update(ctx, account); err != nil {
			return nil, err
		}
	}
	return account, nil
}

func (s *UserAccountService) IsEnabled(ctx context.Context) bool {
	return s == nil || s.settings == nil || s.settings.IsAccountShareEnabled(ctx)
}

func (s *UserAccountService) ensureFeatureEnabled(ctx context.Context) error {
	if !s.IsEnabled(ctx) {
		return ErrUserAccountShareDisabled
	}
	return nil
}

func (s *UserAccountService) getOwnedAccount(ctx context.Context, userID, accountID int64) (*Account, error) {
	repo, err := s.ownedAccountRepo()
	if err != nil {
		return nil, err
	}
	account, err := repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.OwnerUserID == nil || *account.OwnerUserID != userID {
		return nil, ErrUserAccountNotOwned
	}
	return account, nil
}

func (s *UserAccountService) normalizeUserProxyID(ctx context.Context, userID int64, proxyID *int64) (*int64, error) {
	if proxyID == nil || *proxyID == 0 {
		return nil, nil
	}
	if s == nil || s.proxyRepo == nil {
		return nil, fmt.Errorf("proxy repository not configured")
	}
	proxy, err := s.proxyRepo.GetUserOwnedByID(ctx, userID, *proxyID)
	if err != nil {
		return nil, err
	}
	if proxy == nil || !proxy.IsActive() {
		return nil, ErrProxyNotFound
	}
	id := proxy.ID
	return &id, nil
}

type userOwnedAccountRepositoryWithShare interface {
	AccountRepository
	ListUserOwned(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error)
	CountUserOwned(ctx context.Context, userID int64) (int64, error)
}

type userAccountShareLedgerRepository interface {
	ListShareSummary(ctx context.Context, ownerUserID int64) (*UserAccountShareSummary, error)
	GetUsageSummary(ctx context.Context, ownerUserID int64, startTime, endTime time.Time) (*UserAccountUsageSummary, error)
	TransferAvailableShareToBalance(ctx context.Context, ownerUserID int64) (float64, float64, error)
}

func (s *UserAccountService) ownedAccountRepo() (userOwnedAccountRepositoryWithShare, error) {
	if s == nil || s.accountRepo == nil {
		return nil, fmt.Errorf("account repository not configured")
	}
	repo, ok := s.accountRepo.(userOwnedAccountRepositoryWithShare)
	if !ok {
		return nil, fmt.Errorf("account repository does not support user-owned account operations")
	}
	return repo, nil
}

func (s *UserAccountService) shareLedgerRepo() (userAccountShareLedgerRepository, error) {
	if s == nil || s.accountRepo == nil {
		return nil, fmt.Errorf("account repository not configured")
	}
	repo, ok := s.accountRepo.(userAccountShareLedgerRepository)
	if !ok {
		return nil, fmt.Errorf("account repository does not support share ledger operations")
	}
	return repo, nil
}

func (s *UserAccountService) accountRepoForMutation() (AccountRepository, error) {
	if s == nil || s.accountRepo == nil {
		return nil, fmt.Errorf("account repository not configured")
	}
	return s.accountRepo, nil
}

func buildUserAccountCapacityPool(key, title string, accounts []Account, currentUserID int64, displayUsageDeltas map[int64]accountCapacityDisplayUsageDelta) UserAccountCapacityPool {
	pool := UserAccountCapacityPool{
		Key:      key,
		Title:    title,
		Sections: []UserAccountCapacityPoolSection{},
	}
	sections := make(map[string]*UserAccountCapacityPoolSection)
	groups := make(map[string]*UserAccountCapacityPoolGroup)
	groupWindowAvailabilityOnly := make(map[string]map[string]accountCapacityDisplayCounts)
	for i := range accounts {
		account := &accounts[i]
		active := account.IsActive()
		schedulable := account.IsSchedulable()
		ownContributed := accountIsOwnedByUser(account, currentUserID) &&
			account.ShareMode == AccountShareModePublic &&
			account.ShareStatus == AccountShareStatusActive
		rateLimited := account.IsRateLimited() || account.IsOverloaded() || isAccountTempUnschedulable(account)
		errorState := account.Status == StatusError || strings.TrimSpace(account.ErrorMessage) != ""
		disabled := account.Status == StatusDisabled || account.Status == StatusExpired || account.Status == StatusUnused
		abnormal := rateLimited || errorState || disabled || !schedulable
		unavailableReason := accountCapacityUnavailableReason(account)
		counts := accountCapacityDisplayCountsFor(account, active, schedulable, ownContributed, rateLimited, errorState, disabled, abnormal)

		pool.TotalAccounts += counts.total
		pool.ActiveAccounts += counts.active
		pool.SchedulableAccounts += counts.schedulable
		pool.OwnContributedAccounts += counts.ownContributed
		pool.RateLimitedAccounts += counts.rateLimited
		pool.ErrorAccounts += counts.error
		pool.DisabledAccounts += counts.disabled
		pool.AbnormalAccounts += counts.abnormal
		pool.UnavailableReasons = addCapacityUnavailableReason(pool.UnavailableReasons, unavailableReason, counts.total)
		configuredQuota, remainingQuota := accountQuotaTotals(account)
		account.PopulateQuotaWindowSnapshots()
		percentOnlyQuota := account.IsShareDisplayPercentOnly()
		pool.ConfiguredQuota += configuredQuota
		pool.RemainingQuota += remainingQuota
		if percentOnlyQuota {
			pool.PercentOnlyQuota = true
		}

		sectionKey := account.Platform + "/" + account.Type
		section := sections[sectionKey]
		if section == nil {
			section = &UserAccountCapacityPoolSection{
				Platform: account.Platform,
				Type:     account.Type,
				Windows:  map[string]UserAccountCapacityWindowSnapshot{},
			}
			sections[sectionKey] = section
		}
		section.TotalAccounts += counts.total
		section.SchedulableAccounts += counts.schedulable
		section.OwnContributedAccounts += counts.ownContributed
		section.ConfiguredQuota += configuredQuota
		section.RemainingQuota += remainingQuota
		if percentOnlyQuota {
			section.PercentOnlyQuota = true
		}
		section.UnavailableReasons = addCapacityUnavailableReason(section.UnavailableReasons, unavailableReason, counts.total)
		windows := accountCapacityDisplayWindowSnapshots(account, displayUsageDeltas[account.ID])
		mergeCapacityWindowSnapshot(section.Windows, "5h", windows.fiveHour, windows.has5h)
		mergeCapacityWindowSnapshot(section.Windows, "7d", windows.sevenDay, windows.has7d)
		mergeCapacityWindowSnapshot(section.Windows, "1d", windows.dailyQuota, windows.hasDaily)
		mergeCapacityWindowSnapshot(section.Windows, "7d_quota", windows.weekQuota, windows.hasWeek)
		mergeCapacityWindowSnapshot(section.Windows, "30d", windows.monthQuota, windows.hasMonth)

		for _, groupRef := range accountCapacityGroupRefs(key, account) {
			group := groups[groupRef.key]
			if group == nil {
				group = &UserAccountCapacityPoolGroup{
					Key:       groupRef.key,
					GroupName: groupRef.name,
					Platform:  groupRef.platform,
					SortOrder: groupRef.sortOrder,
					Windows:   map[string]UserAccountCapacityWindowSummary{},
				}
				if groupRef.id != nil {
					id := *groupRef.id
					group.GroupID = &id
				}
				groups[groupRef.key] = group
			}
			group.TotalAccounts += counts.total
			group.ActiveAccounts += counts.active
			group.SchedulableAccounts += counts.schedulable
			group.OwnContributedAccounts += counts.ownContributed
			group.RateLimitedAccounts += counts.rateLimited
			group.ErrorAccounts += counts.error
			group.DisabledAccounts += counts.disabled
			group.AbnormalAccounts += counts.abnormal
			group.UnavailableReasons = addCapacityUnavailableReason(group.UnavailableReasons, unavailableReason, counts.total)
			group.ConfiguredQuota += configuredQuota
			group.RemainingQuota += remainingQuota
			if percentOnlyQuota {
				group.PercentOnlyQuota = true
			}
			mergeCapacityWindowSummary(group.Windows, "5h", windows.fiveHour, windows.has5h, schedulable, counts.total)
			mergeCapacityWindowSummary(group.Windows, "7d", windows.sevenDay, windows.has7d, schedulable, counts.total)
			mergeCapacityWindowSummary(group.Windows, "1d", windows.dailyQuota, windows.hasDaily, schedulable, counts.total)
			mergeCapacityWindowSummary(group.Windows, "7d_quota", windows.weekQuota, windows.hasWeek, schedulable, counts.total)
			mergeCapacityWindowSummary(group.Windows, "30d", windows.monthQuota, windows.hasMonth, schedulable, counts.total)
			if accountCapacityGroupRefIsOpenAIPlanDisplay(groupRef) {
				if !windows.has5h {
					addCapacityGroupWindowAvailabilityOnly(groupWindowAvailabilityOnly, groupRef.key, "5h", counts)
				}
				if !windows.has7d {
					addCapacityGroupWindowAvailabilityOnly(groupWindowAvailabilityOnly, groupRef.key, "7d", counts)
				}
			}
		}
	}

	pool.Sections = make([]UserAccountCapacityPoolSection, 0, len(sections))
	for _, section := range sections {
		if len(section.Windows) == 0 {
			section.Windows = nil
		}
		pool.Sections = append(pool.Sections, *section)
	}
	sort.Slice(pool.Sections, func(i, j int) bool {
		left := pool.Sections[i].Platform + "/" + pool.Sections[i].Type
		right := pool.Sections[j].Platform + "/" + pool.Sections[j].Type
		return left < right
	})
	pool.Groups = make([]UserAccountCapacityPoolGroup, 0, len(groups))
	for _, group := range groups {
		applyCapacityGroupWindowAvailabilityOnly(group, groupWindowAvailabilityOnly[group.Key])
		if len(group.Windows) == 0 {
			group.Windows = nil
		}
		group.Status = accountCapacityGroupStatus(group)
		pool.Groups = append(pool.Groups, *group)
	}
	sort.Slice(pool.Groups, func(i, j int) bool {
		left := pool.Groups[i]
		right := pool.Groups[j]
		if left.SortOrder != right.SortOrder {
			return left.SortOrder < right.SortOrder
		}
		if left.Platform != right.Platform {
			return left.Platform < right.Platform
		}
		if left.GroupName != right.GroupName {
			return left.GroupName < right.GroupName
		}
		return left.Key < right.Key
	})
	return pool
}

func accountCapacityGroupRefIsOpenAIPlanDisplay(ref accountCapacityGroupRef) bool {
	return ref.platform == PlatformOpenAI &&
		strings.HasPrefix(ref.key, "share-display:openai:") &&
		openAISharedCapacityGroupName(ref.name) != ""
}

func addCapacityGroupWindowAvailabilityOnly(target map[string]map[string]accountCapacityDisplayCounts, groupKey, windowKey string, counts accountCapacityDisplayCounts) {
	if counts.total <= 0 {
		return
	}
	byWindow := target[groupKey]
	if byWindow == nil {
		byWindow = make(map[string]accountCapacityDisplayCounts)
		target[groupKey] = byWindow
	}
	current := byWindow[windowKey]
	current.total += counts.total
	current.schedulable += counts.schedulable
	byWindow[windowKey] = current
}

func applyCapacityGroupWindowAvailabilityOnly(group *UserAccountCapacityPoolGroup, byWindow map[string]accountCapacityDisplayCounts) {
	if group == nil || len(byWindow) == 0 || len(group.Windows) == 0 {
		return
	}
	for key, counts := range byWindow {
		window, ok := group.Windows[key]
		if !ok {
			continue
		}
		window.SnapshotAccounts += counts.total
		window.SchedulableSnapshotAccounts += counts.schedulable
		group.Windows[key] = window
	}
}

func accountCapacityDisplayWindowSnapshots(account *Account, displayUsage accountCapacityDisplayUsageDelta) accountCapacityDisplayWindows {
	var windows accountCapacityDisplayWindows
	windows.fiveHour, windows.has5h = accountCapacityWindowSnapshot(account, "codex_5h")
	windows.sevenDay, windows.has7d = accountCapacityWindowSnapshot(account, "codex_7d")
	windows.dailyQuota, windows.hasDaily = accountCapacityWindowSnapshot(account, "quota_daily")
	windows.weekQuota, windows.hasWeek = accountCapacityWindowSnapshot(account, "quota_weekly")
	windows.monthQuota, windows.hasMonth = accountCapacityWindowSnapshot(account, "quota_monthly")

	if accountUsesShareDisplayWindowMask(account) {
		if display5h, ok := accountShareDisplayUsageWindowSnapshot(account, "5h", int((5 * time.Hour).Minutes()), displayUsage.fiveHourCost); ok {
			windows.fiveHour = display5h
			windows.has5h = true
		} else if !windows.has5h && windows.hasDaily {
			windows.fiveHour = windows.dailyQuota
			windows.fiveHour.WindowMinutes = int((5 * time.Hour).Minutes())
			windows.has5h = true
		}
		if display7d, ok := accountShareDisplayUsageWindowSnapshot(account, "7d", int((7 * 24 * time.Hour).Minutes()), displayUsage.sevenDayCost); ok {
			windows.sevenDay = display7d
			windows.has7d = true
		} else if !windows.has7d && windows.hasWeek {
			windows.sevenDay = windows.weekQuota
			windows.sevenDay.WindowMinutes = int((7 * 24 * time.Hour).Minutes())
			windows.has7d = true
		}
		windows.hasDaily = false
		windows.hasWeek = false
	}

	return windows
}

func accountShareDisplayUsageWindowSnapshot(account *Account, suffix string, defaultWindowMinutes int, actualUsed float64) (UserAccountCapacityWindowSnapshot, bool) {
	if account == nil || account.Extra == nil {
		return UserAccountCapacityWindowSnapshot{}, false
	}
	prefix := "share_display_" + suffix
	limit := parseExtraFloat64(account.Extra[prefix+"_limit"])
	if limit <= 0 {
		return UserAccountCapacityWindowSnapshot{}, false
	}
	usedRaw, ok := account.Extra[prefix+"_used"]
	if !ok && actualUsed <= 0 {
		return UserAccountCapacityWindowSnapshot{}, false
	}
	used := maxFloat64(actualUsed, 0)
	if ok {
		configuredUsed := maxFloat64(parseExtraFloat64(usedRaw), 0)
		if configuredUsed > 0 {
			used = configuredUsed
		}
	}
	if used > limit {
		used = limit
	}
	snapshot := UserAccountCapacityWindowSnapshot{
		UsedPercent:   used / limit * 100,
		UsedAmount:    used,
		LimitAmount:   limit,
		WindowMinutes: defaultWindowMinutes,
	}
	if raw, ok := account.Extra[prefix+"_reset_after_seconds"]; ok {
		snapshot.ResetAfterSeconds = parseExtraInt(raw)
	}
	if raw, ok := account.Extra[prefix+"_reset_at"]; ok {
		snapshot.ResetAt = normalizeCapacityWindowTime(raw)
	}
	if raw, ok := account.Extra[prefix+"_window_minutes"]; ok {
		if minutes := parseExtraInt(raw); minutes > 0 {
			snapshot.WindowMinutes = minutes
		}
	}
	return snapshot, true
}

func accountCapacityDisplayCountsFor(account *Account, active, schedulable, ownContributed, rateLimited, errorState, disabled, abnormal bool) accountCapacityDisplayCounts {
	weight := 1
	if accountUsesShareDisplayAccountCount(account) {
		weight = account.GetShareDisplayAccountCount()
	}
	counts := accountCapacityDisplayCounts{total: weight}
	if active {
		counts.active = weight
	}
	if schedulable {
		counts.schedulable = weight
	}
	if ownContributed {
		counts.ownContributed = weight
	}
	if rateLimited {
		counts.rateLimited = weight
	}
	if errorState {
		counts.error = weight
	}
	if disabled {
		counts.disabled = weight
	}
	if abnormal {
		counts.abnormal = weight
	}
	return counts
}

func accountUsesShareDisplayAccountCount(account *Account) bool {
	return isSystemShareDisplayCapacityAccount(account)
}

func accountUsesShareDisplayWindowMask(account *Account) bool {
	return account != nil &&
		account.Platform == PlatformOpenAI &&
		(account.Type == AccountTypeAPIKey || account.Type == AccountTypeOAuth) &&
		accountShareDisplayConfigured(account)
}

type accountCapacityGroupRef struct {
	key       string
	id        *int64
	name      string
	platform  string
	sortOrder int
}

func accountCapacityGroupRefs(poolKey string, account *Account) []accountCapacityGroupRef {
	if account == nil {
		return nil
	}
	if displayName := accountShareDisplayGroupName(poolKey, account); displayName != "" {
		return []accountCapacityGroupRef{{
			key:      "share-display:" + account.Platform + ":" + strings.ToLower(displayName),
			name:     displayName,
			platform: account.Platform,
		}}
	}
	if poolKey == "shared" {
		return nil
	}
	seen := make(map[string]struct{})
	refs := make([]accountCapacityGroupRef, 0, len(account.Groups)+len(account.AccountGroups))
	addGroup := func(group *Group) {
		if group == nil {
			return
		}
		id := group.ID
		key := fmt.Sprintf("group:%d", id)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		name := strings.TrimSpace(group.Name)
		if name == "" {
			name = fmt.Sprintf("Group %d", id)
		}
		refs = append(refs, accountCapacityGroupRef{
			key:       key,
			id:        &id,
			name:      name,
			platform:  group.Platform,
			sortOrder: group.SortOrder,
		})
	}
	for _, group := range account.Groups {
		addGroup(group)
	}
	for i := range account.AccountGroups {
		addGroup(account.AccountGroups[i].Group)
	}
	if len(refs) == 0 {
		key := "ungrouped:" + account.Platform
		refs = append(refs, accountCapacityGroupRef{
			key:      key,
			name:     "未分组共享池",
			platform: account.Platform,
		})
	}
	return refs
}

func accountShareDisplayGroupName(poolKey string, account *Account) string {
	if account == nil {
		return ""
	}
	if displayName := openAISharedCapacityGroupName(account.GetShareDisplayTier()); displayName != "" {
		return displayName
	}
	if account.Platform == PlatformOpenAI {
		if displayName := openAISharedCapacityGroupName(account.GetCredential("plan_type")); displayName != "" {
			return displayName
		}
	}
	if poolKey == "shared" {
		if account.Platform == PlatformOpenAI {
			if displayName := openAISharedCapacityGroupNameFromAccountGroups(account); displayName != "" {
				return displayName
			}
		}
		return ""
	}
	if displayName := strings.TrimSpace(account.GetShareDisplayName()); displayName != "" {
		return displayName
	}
	return ""
}

func openAIPlanCapacityGroupName(planType string) string {
	switch strings.ToLower(strings.TrimSpace(planType)) {
	case "plus":
		return "OpenAI Plus"
	case "pro":
		return "OpenAI Pro"
	case "chatgptpro":
		return "OpenAI Pro"
	case "team":
		return "OpenAI Team"
	case "free":
		return "OpenAI Free"
	default:
		return ""
	}
}

func openAISharedCapacityGroupName(planType string) string {
	switch strings.ToLower(strings.TrimSpace(planType)) {
	case "plus", "openai plus", "chatgpt plus":
		return "OpenAI Plus"
	case "pro", "chatgptpro", "openai pro", "chatgpt pro":
		return "OpenAI Pro"
	default:
		return ""
	}
}

func openAISharedCapacityGroupNameFromAccountGroups(account *Account) string {
	if account == nil {
		return ""
	}
	for _, group := range account.Groups {
		if group == nil {
			continue
		}
		if displayName := openAISharedCapacityGroupName(group.Name); displayName != "" {
			return displayName
		}
	}
	for i := range account.AccountGroups {
		group := account.AccountGroups[i].Group
		if group == nil {
			continue
		}
		if displayName := openAISharedCapacityGroupName(group.Name); displayName != "" {
			return displayName
		}
	}
	return ""
}

func accountIsOwnedByUser(account *Account, userID int64) bool {
	return account != nil && account.OwnerUserID != nil && userID > 0 && *account.OwnerUserID == userID
}

func isAccountTempUnschedulable(account *Account) bool {
	return account != nil && account.TempUnschedulableUntil != nil && time.Now().Before(*account.TempUnschedulableUntil)
}

func accountCapacityUnavailableReason(account *Account) string {
	if account == nil || account.IsSchedulable() {
		return ""
	}
	now := time.Now()
	switch {
	case account.Status == StatusError:
		return "error"
	case account.Status == StatusDisabled:
		return "disabled"
	case account.Status == StatusExpired:
		return "expired"
	case account.Status == StatusUnused:
		return "unused"
	case account.AutoPauseOnExpired && account.ExpiresAt != nil && !now.Before(*account.ExpiresAt):
		return "expired"
	case !account.IsActive():
		return "inactive"
	case !account.Schedulable:
		return "manual_unschedulable"
	case account.OverloadUntil != nil && now.Before(*account.OverloadUntil):
		return "overloaded"
	case account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt):
		return "rate_limited"
	case account.TempUnschedulableUntil != nil && now.Before(*account.TempUnschedulableUntil):
		return "temp_unschedulable"
	case account.IsAPIKeyOrBedrock() && account.IsQuotaExceeded():
		return accountCapacityQuotaExceededReason(account)
	default:
		return "unschedulable"
	}
}

func accountCapacityQuotaExceededReason(account *Account) string {
	if account == nil {
		return "quota_exceeded"
	}
	if account.GetQuotaLimit() > 0 && account.GetQuotaUsed() >= account.GetQuotaLimit() {
		return "quota_exceeded"
	}
	if account.GetQuotaDailyLimit() > 0 &&
		!account.IsDailyQuotaPeriodExpired() &&
		account.GetQuotaDailyUsed() >= account.GetQuotaDailyLimit() {
		return "daily_quota_exceeded"
	}
	if account.GetQuotaWeeklyLimit() > 0 &&
		!account.IsWeeklyQuotaPeriodExpired() &&
		account.GetQuotaWeeklyUsed() >= account.GetQuotaWeeklyLimit() {
		return "weekly_quota_exceeded"
	}
	if account.GetQuotaMonthlyLimit() > 0 &&
		!account.IsMonthlyQuotaPeriodExpired() &&
		account.GetQuotaMonthlyUsed() >= account.GetQuotaMonthlyLimit() {
		return "monthly_quota_exceeded"
	}
	return "quota_exceeded"
}

func addCapacityUnavailableReason(current map[string]int, reason string, count int) map[string]int {
	if reason == "" {
		return current
	}
	if count <= 0 {
		count = 1
	}
	if current == nil {
		current = make(map[string]int)
	}
	current[reason] += count
	return current
}

func mergeCapacityWindowSummary(windows map[string]UserAccountCapacityWindowSummary, key string, snapshot UserAccountCapacityWindowSnapshot, ok bool, schedulable bool, accountCount int) {
	if !ok {
		return
	}
	if accountCount <= 0 {
		accountCount = 1
	}
	current := windows[key]
	current.SnapshotAccounts += accountCount
	if schedulable {
		current.SchedulableSnapshotAccounts += accountCount
	}
	if snapshot.LimitAmount > 0 {
		if schedulable {
			current.UsedAmount += snapshot.UsedAmount
			current.LimitAmount += snapshot.LimitAmount
			current.AmountPercentUsedTotal += snapshot.UsedPercent * float64(accountCount)
			current.AmountPercentWeight += float64(accountCount)
			current.RemainingUnits += maxFloat64(snapshot.LimitAmount-snapshot.UsedAmount, 0)
		}
		current = refreshCapacityWindowSummaryPercent(current)
		if !currentHasReset(current) || snapshotResetEarlier(snapshot, current) {
			current.ResetAfterSeconds = snapshot.ResetAfterSeconds
			current.ResetAt = snapshot.ResetAt
			current.WindowMinutes = snapshot.WindowMinutes
		}
		windows[key] = current
		return
	}
	if schedulable {
		current.PercentUsedTotal += snapshot.UsedPercent * float64(accountCount)
		current.PercentWeight += float64(accountCount)
		remaining := 1 - snapshot.UsedPercent/100
		if remaining > 0 {
			current.RemainingUnits += remaining * float64(accountCount)
		}
	}
	current = refreshCapacityWindowSummaryPercent(current)
	if !currentHasReset(current) || snapshotResetEarlier(snapshot, current) {
		current.ResetAfterSeconds = snapshot.ResetAfterSeconds
		current.ResetAt = snapshot.ResetAt
		current.WindowMinutes = snapshot.WindowMinutes
	}
	windows[key] = current
}

func refreshCapacityWindowSummaryPercent(current UserAccountCapacityWindowSummary) UserAccountCapacityWindowSummary {
	if current.PercentWeight > 0 {
		totalWeight := current.PercentWeight + current.AmountPercentWeight
		if totalWeight <= 0 {
			current.UsedPercent = 0
			return current
		}
		current.UsedPercent = (current.PercentUsedTotal + current.AmountPercentUsedTotal) / totalWeight
		return current
	}
	if current.LimitAmount > 0 {
		current.UsedPercent = current.UsedAmount / current.LimitAmount * 100
		return current
	}
	current.UsedPercent = 0
	return current
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func currentHasReset(current UserAccountCapacityWindowSummary) bool {
	return current.ResetAfterSeconds > 0 || strings.TrimSpace(current.ResetAt) != ""
}

func snapshotResetEarlier(snapshot UserAccountCapacityWindowSnapshot, current UserAccountCapacityWindowSummary) bool {
	if snapshot.ResetAfterSeconds <= 0 {
		return false
	}
	if current.ResetAfterSeconds <= 0 {
		return true
	}
	return snapshot.ResetAfterSeconds < current.ResetAfterSeconds
}

func accountCapacityGroupStatus(group *UserAccountCapacityPoolGroup) string {
	if group == nil || group.TotalAccounts == 0 || group.SchedulableAccounts == 0 {
		return "unavailable"
	}
	if group.RateLimitedAccounts > 0 || group.ErrorAccounts > 0 || group.DisabledAccounts > 0 {
		return "degraded"
	}
	for _, window := range group.Windows {
		if window.UsedPercent >= 80 {
			return "degraded"
		}
	}
	return "healthy"
}

func accountQuotaTotals(account *Account) (configured, remaining float64) {
	if account == nil {
		return 0, 0
	}
	addQuota := func(limit, used float64) {
		if limit <= 0 {
			return
		}
		configured += limit
		left := limit - used
		if left > 0 {
			remaining += left
		}
	}
	addQuota(account.GetQuotaLimit(), account.GetQuotaUsed())
	dailyUsed := account.GetQuotaDailyUsed()
	if account.IsDailyQuotaPeriodExpired() {
		dailyUsed = 0
	}
	addQuota(account.GetQuotaDailyLimit(), dailyUsed)
	weeklyUsed := account.GetQuotaWeeklyUsed()
	if account.IsWeeklyQuotaPeriodExpired() {
		weeklyUsed = 0
	}
	addQuota(account.GetQuotaWeeklyLimit(), weeklyUsed)
	monthlyUsed := account.GetQuotaMonthlyUsed()
	if account.IsMonthlyQuotaPeriodExpired() {
		monthlyUsed = 0
	}
	addQuota(account.GetQuotaMonthlyLimit(), monthlyUsed)
	return configured, remaining
}

func accountCapacityWindowSnapshot(account *Account, prefix string) (UserAccountCapacityWindowSnapshot, bool) {
	if account == nil || account.Extra == nil {
		return UserAccountCapacityWindowSnapshot{}, false
	}
	usedRaw, ok := account.Extra[prefix+"_used_percent"]
	if !ok {
		return UserAccountCapacityWindowSnapshot{}, false
	}
	snapshot := UserAccountCapacityWindowSnapshot{
		UsedPercent: parseExtraFloat64(usedRaw),
	}
	if raw, ok := account.Extra[prefix+"_reset_after_seconds"]; ok {
		snapshot.ResetAfterSeconds = parseExtraInt(raw)
	}
	if raw, ok := account.Extra[prefix+"_reset_at"]; ok {
		snapshot.ResetAt = normalizeCapacityWindowTime(raw)
	}
	if raw, ok := account.Extra[prefix+"_window_minutes"]; ok {
		snapshot.WindowMinutes = parseExtraInt(raw)
	}
	populateQuotaWindowAmounts(account, prefix, &snapshot)
	return snapshot, true
}

func populateQuotaWindowAmounts(account *Account, prefix string, snapshot *UserAccountCapacityWindowSnapshot) {
	if account == nil || snapshot == nil {
		return
	}
	var limit, used float64
	switch prefix {
	case "quota_daily":
		limit = account.GetQuotaDailyLimit()
		used = account.GetQuotaDailyUsed()
		if account.IsDailyQuotaPeriodExpired() {
			used = 0
		}
	case "quota_weekly":
		limit = account.GetQuotaWeeklyLimit()
		used = account.GetQuotaWeeklyUsed()
		if account.IsWeeklyQuotaPeriodExpired() {
			used = 0
		}
	case "quota_monthly":
		limit = account.GetQuotaMonthlyLimit()
		used = account.GetQuotaMonthlyUsed()
		if account.IsMonthlyQuotaPeriodExpired() {
			used = 0
		}
	default:
		return
	}
	if limit <= 0 {
		return
	}
	snapshot.LimitAmount = limit
	snapshot.UsedAmount = maxFloat64(used, 0)
	if snapshot.UsedAmount > snapshot.LimitAmount {
		snapshot.UsedAmount = snapshot.LimitAmount
	}
	snapshot.UsedPercent = snapshot.UsedAmount / snapshot.LimitAmount * 100
}

func mergeCapacityWindowSnapshot(windows map[string]UserAccountCapacityWindowSnapshot, key string, snapshot UserAccountCapacityWindowSnapshot, ok bool) {
	if !ok {
		return
	}
	current, exists := windows[key]
	if !exists || snapshot.UsedPercent > current.UsedPercent {
		windows[key] = snapshot
	}
}

func normalizeCapacityWindowTime(raw any) string {
	switch v := raw.(type) {
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return ""
		}
		if parsed, err := time.Parse(time.RFC3339Nano, text); err == nil {
			return parsed.UTC().Format(time.RFC3339)
		}
		if parsed, err := time.Parse(time.RFC3339, text); err == nil {
			return parsed.UTC().Format(time.RFC3339)
		}
		return text
	default:
		return ""
	}
}

func normalizeUserAccountConcurrency(value int) int {
	if value >= 1 {
		return value
	}
	return userAccountDefaultConcurrency
}

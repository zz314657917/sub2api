package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
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
	TransferAvailableShareToBalance(ctx context.Context, ownerUserID int64) (float64, float64, error)
}

var (
	ErrUserAccountNotOwned      = infraerrors.Forbidden("USER_ACCOUNT_NOT_OWNED", "account does not belong to current user")
	ErrUserAccountShareInvalid  = infraerrors.BadRequest("USER_ACCOUNT_SHARE_INVALID", "invalid account share state")
	ErrUserAccountShareDisabled = infraerrors.Forbidden("USER_ACCOUNT_SHARE_DISABLED", "account sharing is disabled")
	ErrUserAccountLimitReached  = infraerrors.Conflict("USER_ACCOUNT_LIMIT_REACHED", "account limit reached")
)

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

type UserAccountCapacityPools struct {
	Mine   UserAccountCapacityPool `json:"mine"`
	Shared UserAccountCapacityPool `json:"shared"`
}

type UserAccountCapacityPool struct {
	Key                 string                           `json:"key"`
	Title               string                           `json:"title"`
	TotalAccounts       int                              `json:"total_accounts"`
	SchedulableAccounts int                              `json:"schedulable_accounts"`
	ConfiguredQuota     float64                          `json:"configured_quota"`
	RemainingQuota      float64                          `json:"remaining_quota"`
	Sections            []UserAccountCapacityPoolSection `json:"sections"`
}

type UserAccountCapacityPoolSection struct {
	Platform            string                                       `json:"platform"`
	Type                string                                       `json:"type"`
	TotalAccounts       int                                          `json:"total_accounts"`
	SchedulableAccounts int                                          `json:"schedulable_accounts"`
	ConfiguredQuota     float64                                      `json:"configured_quota"`
	RemainingQuota      float64                                      `json:"remaining_quota"`
	Windows             map[string]UserAccountCapacityWindowSnapshot `json:"windows,omitempty"`
}

type UserAccountCapacityWindowSnapshot struct {
	UsedPercent       float64 `json:"used_percent"`
	ResetAfterSeconds int     `json:"reset_after_seconds,omitempty"`
	ResetAt           string  `json:"reset_at,omitempty"`
	WindowMinutes     int     `json:"window_minutes,omitempty"`
}

type UserAccountService struct {
	accountRepo AccountRepository
	settings    UserAccountShareSettings
}

type UserAccountShareSettings interface {
	IsAccountShareEnabled(ctx context.Context) bool
	IsAccountShareAutoReview(ctx context.Context) bool
	GetAccountShareUserAccountLimit(ctx context.Context) int
}

func NewUserAccountService(accountRepo AccountRepository, settings UserAccountShareSettings) *UserAccountService {
	return &UserAccountService{
		accountRepo: accountRepo,
		settings:    settings,
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
		Concurrency: 0,
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
		account.Credentials = *req.Credentials
	}
	if req.Extra != nil {
		account.Extra = *req.Extra
	}
	if req.ExpiresAt != nil {
		account.ExpiresAt = req.ExpiresAt
	}
	if req.AutoPauseOnExpired != nil {
		account.AutoPauseOnExpired = *req.AutoPauseOnExpired
	}
	if req.GroupIDs != nil || req.ProxyID != nil || req.Concurrency != nil || req.Priority != nil || req.Status != nil {
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

	params := pagination.PaginationParams{
		Page:      1,
		PageSize:  1000,
		SortBy:    "created_at",
		SortOrder: pagination.SortOrderDesc,
	}
	owned, _, err := ownedRepo.ListUserOwned(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	allAccounts, _, err := allRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	shared := make([]Account, 0, len(allAccounts))
	for _, account := range allAccounts {
		if account.OwnerUserID != nil && userID > 0 && *account.OwnerUserID == userID {
			continue
		}
		if account.CanBeUsedByUser(userID) {
			shared = append(shared, account)
		}
	}

	return &UserAccountCapacityPools{
		Mine:   buildUserAccountCapacityPool("mine", "我的账号容量池", owned),
		Shared: buildUserAccountCapacityPool("shared", "平台共享容量池", shared),
	}, nil
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

type userOwnedAccountRepositoryWithShare interface {
	AccountRepository
	ListUserOwned(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error)
	CountUserOwned(ctx context.Context, userID int64) (int64, error)
}

type userAccountShareLedgerRepository interface {
	ListShareSummary(ctx context.Context, ownerUserID int64) (*UserAccountShareSummary, error)
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

func buildUserAccountCapacityPool(key, title string, accounts []Account) UserAccountCapacityPool {
	pool := UserAccountCapacityPool{
		Key:      key,
		Title:    title,
		Sections: []UserAccountCapacityPoolSection{},
	}
	sections := make(map[string]*UserAccountCapacityPoolSection)
	for i := range accounts {
		account := &accounts[i]
		pool.TotalAccounts++
		if account.IsSchedulable() {
			pool.SchedulableAccounts++
		}
		configuredQuota, remainingQuota := accountQuotaTotals(account)
		pool.ConfiguredQuota += configuredQuota
		pool.RemainingQuota += remainingQuota

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
		section.TotalAccounts++
		if account.IsSchedulable() {
			section.SchedulableAccounts++
		}
		section.ConfiguredQuota += configuredQuota
		section.RemainingQuota += remainingQuota
		snapshot5h, ok5h := accountCapacityWindowSnapshot(account, "codex_5h")
		mergeCapacityWindowSnapshot(section.Windows, "5h", snapshot5h, ok5h)
		snapshot7d, ok7d := accountCapacityWindowSnapshot(account, "codex_7d")
		mergeCapacityWindowSnapshot(section.Windows, "7d", snapshot7d, ok7d)
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
	return pool
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
	addQuota(account.GetQuotaDailyLimit(), account.GetQuotaDailyUsed())
	addQuota(account.GetQuotaWeeklyLimit(), account.GetQuotaWeeklyUsed())
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
	return snapshot, true
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

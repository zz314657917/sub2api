package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	apptimezone "github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"golang.org/x/sync/singleflight"
)

var (
	ErrUsageLogNotFound = infraerrors.NotFound("USAGE_LOG_NOT_FOUND", "usage log not found")
)

const leaderboardStatsCacheTTL = 5 * time.Minute

// CreateUsageLogRequest 创建使用日志请求
type CreateUsageLogRequest struct {
	UserID                int64   `json:"user_id"`
	APIKeyID              int64   `json:"api_key_id"`
	AccountID             int64   `json:"account_id"`
	RequestID             string  `json:"request_id"`
	Model                 string  `json:"model"`
	InputTokens           int     `json:"input_tokens"`
	OutputTokens          int     `json:"output_tokens"`
	CacheCreationTokens   int     `json:"cache_creation_tokens"`
	CacheReadTokens       int     `json:"cache_read_tokens"`
	CacheCreation5mTokens int     `json:"cache_creation_5m_tokens"`
	CacheCreation1hTokens int     `json:"cache_creation_1h_tokens"`
	InputCost             float64 `json:"input_cost"`
	OutputCost            float64 `json:"output_cost"`
	CacheCreationCost     float64 `json:"cache_creation_cost"`
	CacheReadCost         float64 `json:"cache_read_cost"`
	TotalCost             float64 `json:"total_cost"`
	ActualCost            float64 `json:"actual_cost"`
	RateMultiplier        float64 `json:"rate_multiplier"`
	Stream                bool    `json:"stream"`
	DurationMs            *int    `json:"duration_ms"`
}

// UsageStats 使用统计
type UsageStats struct {
	TotalRequests            int64   `json:"total_requests"`
	TotalInputTokens         int64   `json:"total_input_tokens"`
	TotalOutputTokens        int64   `json:"total_output_tokens"`
	TotalCacheTokens         int64   `json:"total_cache_tokens"`
	TotalCacheCreationTokens int64   `json:"total_cache_creation_tokens"`
	TotalCacheReadTokens     int64   `json:"total_cache_read_tokens"`
	TotalTokens              int64   `json:"total_tokens"`
	TotalCost                float64 `json:"total_cost"`
	TotalActualCost          float64 `json:"total_actual_cost"`
	AverageDurationMs        float64 `json:"average_duration_ms"`
}

// UsageService 使用统计服务
type UsageService struct {
	usageRepo            UsageLogRepository
	userRepo             UserRepository
	entClient            *dbent.Client
	authCacheInvalidator APIKeyAuthCacheInvalidator
	settingRepo          SettingRepository
	redeemRepo           RedeemCodeRepository
	badgeCacheMu         sync.Mutex
	badgeCache           map[string]leaderboardBadgeCacheEntry
	leaderboardCacheMu   sync.Mutex
	leaderboardCacheSF   singleflight.Group
	userLeaderboardCache map[string]leaderboardUserCacheEntry
	modelRankingCache    map[string]leaderboardModelRankingCacheEntry
	recentTrendCache     map[string]leaderboardRecentTrendCacheEntry
	dailyChampionsCache  map[string]leaderboardDailyChampionsCacheEntry
}

type userLeaderboardBadgeLeaderRepository interface {
	GetUserLeaderboardBadgeLeaders(ctx context.Context, weekStart, weekEnd, monthStart, monthEnd, costStart, costEnd time.Time, userTZ string) (*usagestats.UserLeaderboardBadgeLeaders, error)
}

type leaderboardModelRankingRepository interface {
	GetLeaderboardModelRanking(ctx context.Context, startTime, endTime time.Time, limit int) ([]usagestats.UserLeaderboardModelItem, int64, error)
}

type leaderboardDailyChampionsRepository interface {
	GetLeaderboardDailyChampions(ctx context.Context, startTime, endTime time.Time) ([]usagestats.UserLeaderboardDailyChampion, error)
}

type leaderboardBadgeCacheEntry struct {
	expiresAt time.Time
	leaders   *usagestats.UserLeaderboardBadgeLeaders
}

type leaderboardUserCacheEntry struct {
	expiresAt   time.Time
	leaderboard *usagestats.UserLeaderboardResponse
}

type leaderboardModelRankingCacheEntry struct {
	expiresAt   time.Time
	items       []usagestats.UserLeaderboardModelItem
	totalModels int64
}

type leaderboardRecentTrendCacheEntry struct {
	expiresAt time.Time
	points    []usagestats.UserLeaderboardTokenTrendPoint
}

type leaderboardDailyChampionsCacheEntry struct {
	expiresAt time.Time
	champions []usagestats.UserLeaderboardDailyChampion
}

// NewUsageService 创建使用统计服务实例
func NewUsageService(usageRepo UsageLogRepository, userRepo UserRepository, entClient *dbent.Client, authCacheInvalidator APIKeyAuthCacheInvalidator) *UsageService {
	return &UsageService{
		usageRepo:            usageRepo,
		userRepo:             userRepo,
		entClient:            entClient,
		authCacheInvalidator: authCacheInvalidator,
		badgeCache:           make(map[string]leaderboardBadgeCacheEntry),
		userLeaderboardCache: make(map[string]leaderboardUserCacheEntry),
		modelRankingCache:    make(map[string]leaderboardModelRankingCacheEntry),
		recentTrendCache:     make(map[string]leaderboardRecentTrendCacheEntry),
		dailyChampionsCache:  make(map[string]leaderboardDailyChampionsCacheEntry),
	}
}

// SetLeaderboardRewardDependencies wires optional repositories used by leaderboard rewards.
func (s *UsageService) SetLeaderboardRewardDependencies(settingRepo SettingRepository, redeemRepo RedeemCodeRepository) {
	s.settingRepo = settingRepo
	s.redeemRepo = redeemRepo
}

// ProvideUsageService creates UsageService with optional leaderboard reward dependencies.
func ProvideUsageService(usageRepo UsageLogRepository, userRepo UserRepository, settingRepo SettingRepository, redeemRepo RedeemCodeRepository, entClient *dbent.Client, authCacheInvalidator APIKeyAuthCacheInvalidator) *UsageService {
	svc := NewUsageService(usageRepo, userRepo, entClient, authCacheInvalidator)
	svc.SetLeaderboardRewardDependencies(settingRepo, redeemRepo)
	return svc
}

// Create 创建使用日志
func (s *UsageService) Create(ctx context.Context, req CreateUsageLogRequest) (*UsageLog, error) {
	// 使用数据库事务保证「使用日志插入」与「扣费」的原子性，避免重复扣费或漏扣风险。
	tx, err := s.entClient.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := ctx
	if err == nil {
		defer func() { _ = tx.Rollback() }()
		txCtx = dbent.NewTxContext(ctx, tx)
	}

	// 验证用户存在
	_, err = s.userRepo.GetByID(txCtx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 创建使用日志
	usageLog := &UsageLog{
		UserID:                req.UserID,
		APIKeyID:              req.APIKeyID,
		AccountID:             req.AccountID,
		RequestID:             req.RequestID,
		Model:                 req.Model,
		InputTokens:           req.InputTokens,
		OutputTokens:          req.OutputTokens,
		CacheCreationTokens:   req.CacheCreationTokens,
		CacheReadTokens:       req.CacheReadTokens,
		CacheCreation5mTokens: req.CacheCreation5mTokens,
		CacheCreation1hTokens: req.CacheCreation1hTokens,
		InputCost:             req.InputCost,
		OutputCost:            req.OutputCost,
		CacheCreationCost:     req.CacheCreationCost,
		CacheReadCost:         req.CacheReadCost,
		TotalCost:             req.TotalCost,
		ActualCost:            req.ActualCost,
		RateMultiplier:        req.RateMultiplier,
		Stream:                req.Stream,
		DurationMs:            req.DurationMs,
	}

	inserted, err := s.usageRepo.Create(txCtx, usageLog)
	if err != nil {
		return nil, fmt.Errorf("create usage log: %w", err)
	}

	// 扣除用户余额
	balanceUpdated := false
	if inserted && req.ActualCost > 0 {
		if err := s.userRepo.UpdateBalance(txCtx, req.UserID, -req.ActualCost); err != nil {
			return nil, fmt.Errorf("update user balance: %w", err)
		}
		balanceUpdated = true
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit transaction: %w", err)
		}
	}

	s.invalidateUsageCaches(ctx, req.UserID, balanceUpdated)

	return usageLog, nil
}

func (s *UsageService) invalidateUsageCaches(ctx context.Context, userID int64, balanceUpdated bool) {
	s.badgeCacheMu.Lock()
	s.badgeCache = make(map[string]leaderboardBadgeCacheEntry)
	s.badgeCacheMu.Unlock()

	if !balanceUpdated || s.authCacheInvalidator == nil {
		return
	}
	s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
}

// GetByID 根据ID获取使用日志
func (s *UsageService) GetByID(ctx context.Context, id int64) (*UsageLog, error) {
	log, err := s.usageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get usage log: %w", err)
	}
	return log, nil
}

// ListByUser 获取用户的使用日志列表
func (s *UsageService) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, pagination, err := s.usageRepo.ListByUser(ctx, userID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs: %w", err)
	}
	return logs, pagination, nil
}

// ListByAPIKey 获取API Key的使用日志列表
func (s *UsageService) ListByAPIKey(ctx context.Context, apiKeyID int64, params pagination.PaginationParams) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, pagination, err := s.usageRepo.ListByAPIKey(ctx, apiKeyID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs: %w", err)
	}
	return logs, pagination, nil
}

// ListByAccount 获取账号的使用日志列表
func (s *UsageService) ListByAccount(ctx context.Context, accountID int64, params pagination.PaginationParams) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, pagination, err := s.usageRepo.ListByAccount(ctx, accountID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs: %w", err)
	}
	return logs, pagination, nil
}

// GetStatsByUser 获取用户的使用统计
func (s *UsageService) GetStatsByUser(ctx context.Context, userID int64, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetUserStatsAggregated(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:            stats.TotalRequests,
		TotalInputTokens:         stats.TotalInputTokens,
		TotalOutputTokens:        stats.TotalOutputTokens,
		TotalCacheTokens:         stats.TotalCacheTokens,
		TotalCacheCreationTokens: stats.TotalCacheCreationTokens,
		TotalCacheReadTokens:     stats.TotalCacheReadTokens,
		TotalTokens:              stats.TotalTokens,
		TotalCost:                stats.TotalCost,
		TotalActualCost:          stats.TotalActualCost,
		AverageDurationMs:        stats.AverageDurationMs,
	}, nil
}

// GetStatsByAPIKey 获取API Key的使用统计
func (s *UsageService) GetStatsByAPIKey(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetAPIKeyStatsAggregated(ctx, apiKeyID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get api key stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:            stats.TotalRequests,
		TotalInputTokens:         stats.TotalInputTokens,
		TotalOutputTokens:        stats.TotalOutputTokens,
		TotalCacheTokens:         stats.TotalCacheTokens,
		TotalCacheCreationTokens: stats.TotalCacheCreationTokens,
		TotalCacheReadTokens:     stats.TotalCacheReadTokens,
		TotalTokens:              stats.TotalTokens,
		TotalCost:                stats.TotalCost,
		TotalActualCost:          stats.TotalActualCost,
		AverageDurationMs:        stats.AverageDurationMs,
	}, nil
}

// GetStatsByAccount 获取账号的使用统计
func (s *UsageService) GetStatsByAccount(ctx context.Context, accountID int64, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetAccountStatsAggregated(ctx, accountID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get account stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:            stats.TotalRequests,
		TotalInputTokens:         stats.TotalInputTokens,
		TotalOutputTokens:        stats.TotalOutputTokens,
		TotalCacheTokens:         stats.TotalCacheTokens,
		TotalCacheCreationTokens: stats.TotalCacheCreationTokens,
		TotalCacheReadTokens:     stats.TotalCacheReadTokens,
		TotalTokens:              stats.TotalTokens,
		TotalCost:                stats.TotalCost,
		TotalActualCost:          stats.TotalActualCost,
		AverageDurationMs:        stats.AverageDurationMs,
	}, nil
}

// GetStatsByModel 获取模型的使用统计
func (s *UsageService) GetStatsByModel(ctx context.Context, modelName string, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetModelStatsAggregated(ctx, modelName, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get model stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:            stats.TotalRequests,
		TotalInputTokens:         stats.TotalInputTokens,
		TotalOutputTokens:        stats.TotalOutputTokens,
		TotalCacheTokens:         stats.TotalCacheTokens,
		TotalCacheCreationTokens: stats.TotalCacheCreationTokens,
		TotalCacheReadTokens:     stats.TotalCacheReadTokens,
		TotalTokens:              stats.TotalTokens,
		TotalCost:                stats.TotalCost,
		TotalActualCost:          stats.TotalActualCost,
		AverageDurationMs:        stats.AverageDurationMs,
	}, nil
}

// GetDailyStats 获取每日使用统计（最近N天）
func (s *UsageService) GetDailyStats(ctx context.Context, userID int64, days int) ([]map[string]any, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	stats, err := s.usageRepo.GetDailyStatsAggregated(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get daily stats: %w", err)
	}

	return stats, nil
}

// Delete 删除使用日志（管理员功能，谨慎使用）
func (s *UsageService) Delete(ctx context.Context, id int64) error {
	if err := s.usageRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete usage log: %w", err)
	}
	return nil
}

// GetUserDashboardStats returns per-user dashboard summary stats.
func (s *UsageService) GetUserDashboardStats(ctx context.Context, userID int64) (*usagestats.UserDashboardStats, error) {
	stats, err := s.usageRepo.GetUserDashboardStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user dashboard stats: %w", err)
	}
	return stats, nil
}

// GetAPIKeyDashboardStats returns dashboard summary stats filtered by API Key.
func (s *UsageService) GetAPIKeyDashboardStats(ctx context.Context, apiKeyID int64) (*usagestats.UserDashboardStats, error) {
	stats, err := s.usageRepo.GetAPIKeyDashboardStats(ctx, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("get api key dashboard stats: %w", err)
	}
	return stats, nil
}

// GetUserLeaderboard returns the user dashboard leaderboard for a time range.
func (s *UsageService) GetUserLeaderboard(ctx context.Context, startTime, endTime time.Time, limit int, currentUserID int64) (*usagestats.UserLeaderboardResponse, error) {
	cacheKey := userLeaderboardStatsCacheKey(startTime, endTime, limit, currentUserID)
	if cached := s.getCachedUserLeaderboard(cacheKey, time.Now()); cached != nil {
		return cached, nil
	}

	value, err, _ := s.leaderboardCacheSF.Do("user:"+cacheKey, func() (any, error) {
		if cached := s.getCachedUserLeaderboard(cacheKey, time.Now()); cached != nil {
			return cached, nil
		}
		leaderboard, err := s.usageRepo.GetUserLeaderboard(ctx, startTime, endTime, limit, currentUserID)
		if err != nil {
			return nil, err
		}
		s.setCachedUserLeaderboard(cacheKey, leaderboard, time.Now())
		return cloneUserLeaderboardResponse(leaderboard), nil
	})
	if err != nil {
		return nil, fmt.Errorf("get user leaderboard: %w", err)
	}
	leaderboard, _ := value.(*usagestats.UserLeaderboardResponse)
	return cloneUserLeaderboardResponse(leaderboard), nil
}

// GetLeaderboardModelRanking returns global model usage ranking for the leaderboard period.
func (s *UsageService) GetLeaderboardModelRanking(ctx context.Context, startTime, endTime time.Time, limit int) ([]usagestats.UserLeaderboardModelItem, int64, error) {
	repo, ok := s.usageRepo.(leaderboardModelRankingRepository)
	if !ok {
		return []usagestats.UserLeaderboardModelItem{}, 0, nil
	}
	cacheKey := leaderboardStatsCacheKey(startTime, endTime, limit)
	if items, totalModels, ok := s.getCachedLeaderboardModelRanking(cacheKey, time.Now()); ok {
		return items, totalModels, nil
	}

	value, err, _ := s.leaderboardCacheSF.Do("model:"+cacheKey, func() (any, error) {
		if items, totalModels, ok := s.getCachedLeaderboardModelRanking(cacheKey, time.Now()); ok {
			return leaderboardModelRankingCacheEntry{items: items, totalModels: totalModels}, nil
		}
		items, totalModels, err := repo.GetLeaderboardModelRanking(ctx, startTime, endTime, limit)
		if err != nil {
			return nil, err
		}
		if items == nil {
			items = []usagestats.UserLeaderboardModelItem{}
		}
		s.setCachedLeaderboardModelRanking(cacheKey, items, totalModels, time.Now())
		return leaderboardModelRankingCacheEntry{items: cloneUserLeaderboardModelItems(items), totalModels: totalModels}, nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("get leaderboard model ranking: %w", err)
	}
	result, _ := value.(leaderboardModelRankingCacheEntry)
	if result.items == nil {
		result.items = []usagestats.UserLeaderboardModelItem{}
	}
	return cloneUserLeaderboardModelItems(result.items), result.totalModels, nil
}

// GetLeaderboardRecentTokenTrend returns global daily token totals for the leaderboard summary.
func (s *UsageService) GetLeaderboardRecentTokenTrend(ctx context.Context, startTime, endTime time.Time) ([]usagestats.UserLeaderboardTokenTrendPoint, error) {
	cacheKey := leaderboardStatsCacheKey(startTime, endTime, 0)
	if points, ok := s.getCachedLeaderboardRecentTrend(cacheKey, time.Now()); ok {
		return points, nil
	}

	value, err, _ := s.leaderboardCacheSF.Do("trend:"+cacheKey, func() (any, error) {
		if points, ok := s.getCachedLeaderboardRecentTrend(cacheKey, time.Now()); ok {
			return points, nil
		}
		trend, err := s.usageRepo.GetUsageTrendWithFilters(ctx, startTime, endTime, "day", 0, 0, 0, 0, "", nil, nil, nil)
		if err != nil {
			return nil, err
		}

		points := make([]usagestats.UserLeaderboardTokenTrendPoint, 0, len(trend))
		for _, item := range trend {
			points = append(points, usagestats.UserLeaderboardTokenTrendPoint{
				Date:        item.Date,
				TotalTokens: item.TotalTokens,
			})
		}
		s.setCachedLeaderboardRecentTrend(cacheKey, points, time.Now())
		return cloneLeaderboardRecentTrend(points), nil
	})
	if err != nil {
		return nil, fmt.Errorf("get leaderboard recent token trend: %w", err)
	}
	points, _ := value.([]usagestats.UserLeaderboardTokenTrendPoint)
	return cloneLeaderboardRecentTrend(points), nil
}

// GetLeaderboardDailyChampions returns daily top token users for the leaderboard calendar.
func (s *UsageService) GetLeaderboardDailyChampions(ctx context.Context, startTime, endTime time.Time) ([]usagestats.UserLeaderboardDailyChampion, error) {
	repo, ok := s.usageRepo.(leaderboardDailyChampionsRepository)
	if !ok {
		return []usagestats.UserLeaderboardDailyChampion{}, nil
	}
	cacheKey := leaderboardStatsCacheKey(startTime, endTime, 0)
	if champions, ok := s.getCachedLeaderboardDailyChampions(cacheKey, time.Now()); ok {
		return champions, nil
	}

	value, err, _ := s.leaderboardCacheSF.Do("daily-champions:"+cacheKey, func() (any, error) {
		if champions, ok := s.getCachedLeaderboardDailyChampions(cacheKey, time.Now()); ok {
			return champions, nil
		}
		champions, err := repo.GetLeaderboardDailyChampions(ctx, startTime, endTime)
		if err != nil {
			return nil, err
		}
		if champions == nil {
			champions = []usagestats.UserLeaderboardDailyChampion{}
		}
		s.setCachedLeaderboardDailyChampions(cacheKey, champions, time.Now())
		return cloneLeaderboardDailyChampions(champions), nil
	})
	if err != nil {
		return nil, fmt.Errorf("get leaderboard daily champions: %w", err)
	}
	champions, _ := value.([]usagestats.UserLeaderboardDailyChampion)
	return cloneLeaderboardDailyChampions(champions), nil
}

// GetUserLeaderboardBadgeLeaders returns special badge leader user IDs for the requested windows.
func (s *UsageService) GetUserLeaderboardBadgeLeaders(ctx context.Context, weekStart, weekEnd, monthStart, monthEnd, costStart, costEnd time.Time, userTZ string) (*usagestats.UserLeaderboardBadgeLeaders, error) {
	repo, ok := s.usageRepo.(userLeaderboardBadgeLeaderRepository)
	if !ok {
		return &usagestats.UserLeaderboardBadgeLeaders{}, nil
	}
	now := time.Now()
	cacheKey := leaderboardBadgeCacheKey(weekStart, weekEnd, monthStart, monthEnd, costStart, costEnd, userTZ)
	expiresAt := leaderboardBadgeCacheExpiry(userTZ, now)
	if cached := s.getCachedLeaderboardBadgeLeaders(cacheKey, now); cached != nil {
		return cached, nil
	}

	leaders, err := repo.GetUserLeaderboardBadgeLeaders(ctx, weekStart, weekEnd, monthStart, monthEnd, costStart, costEnd, userTZ)
	if err != nil {
		return nil, fmt.Errorf("get user leaderboard badge leaders: %w", err)
	}
	if leaders == nil {
		leaders = &usagestats.UserLeaderboardBadgeLeaders{}
	}
	if leaderboardBadgeLeadersHasAny(leaders) {
		s.setCachedLeaderboardBadgeLeaders(cacheKey, leaders, expiresAt, now)
	}
	return cloneLeaderboardBadgeLeaders(leaders), nil
}

func (s *UsageService) getCachedLeaderboardBadgeLeaders(key string, now time.Time) *usagestats.UserLeaderboardBadgeLeaders {
	s.badgeCacheMu.Lock()
	defer s.badgeCacheMu.Unlock()
	entry, ok := s.badgeCache[key]
	if !ok || entry.leaders == nil || !now.Before(entry.expiresAt) {
		return nil
	}
	return cloneLeaderboardBadgeLeaders(entry.leaders)
}

func (s *UsageService) setCachedLeaderboardBadgeLeaders(key string, leaders *usagestats.UserLeaderboardBadgeLeaders, expiresAt, now time.Time) {
	s.badgeCacheMu.Lock()
	defer s.badgeCacheMu.Unlock()
	if s.badgeCache == nil {
		s.badgeCache = make(map[string]leaderboardBadgeCacheEntry)
	}
	for existingKey, entry := range s.badgeCache {
		if !now.Before(entry.expiresAt) {
			delete(s.badgeCache, existingKey)
		}
	}
	s.badgeCache[key] = leaderboardBadgeCacheEntry{
		expiresAt: expiresAt,
		leaders:   cloneLeaderboardBadgeLeaders(leaders),
	}
}

func (s *UsageService) getCachedUserLeaderboard(key string, now time.Time) *usagestats.UserLeaderboardResponse {
	if s == nil || key == "" {
		return nil
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	entry, ok := s.userLeaderboardCache[key]
	if !ok || entry.leaderboard == nil || !now.Before(entry.expiresAt) {
		return nil
	}
	return cloneUserLeaderboardResponse(entry.leaderboard)
}

func (s *UsageService) setCachedUserLeaderboard(key string, leaderboard *usagestats.UserLeaderboardResponse, now time.Time) {
	if s == nil || key == "" || leaderboard == nil {
		return
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	if s.userLeaderboardCache == nil {
		s.userLeaderboardCache = make(map[string]leaderboardUserCacheEntry)
	}
	for existingKey, entry := range s.userLeaderboardCache {
		if !now.Before(entry.expiresAt) {
			delete(s.userLeaderboardCache, existingKey)
		}
	}
	s.userLeaderboardCache[key] = leaderboardUserCacheEntry{
		expiresAt:   now.Add(leaderboardStatsCacheTTL),
		leaderboard: cloneUserLeaderboardResponse(leaderboard),
	}
}

func (s *UsageService) getCachedLeaderboardModelRanking(key string, now time.Time) ([]usagestats.UserLeaderboardModelItem, int64, bool) {
	if s == nil || key == "" {
		return nil, 0, false
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	entry, ok := s.modelRankingCache[key]
	if !ok || !now.Before(entry.expiresAt) {
		return nil, 0, false
	}
	return cloneUserLeaderboardModelItems(entry.items), entry.totalModels, true
}

func (s *UsageService) setCachedLeaderboardModelRanking(key string, items []usagestats.UserLeaderboardModelItem, totalModels int64, now time.Time) {
	if s == nil || key == "" {
		return
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	if s.modelRankingCache == nil {
		s.modelRankingCache = make(map[string]leaderboardModelRankingCacheEntry)
	}
	for existingKey, entry := range s.modelRankingCache {
		if !now.Before(entry.expiresAt) {
			delete(s.modelRankingCache, existingKey)
		}
	}
	s.modelRankingCache[key] = leaderboardModelRankingCacheEntry{
		expiresAt:   now.Add(leaderboardStatsCacheTTL),
		items:       cloneUserLeaderboardModelItems(items),
		totalModels: totalModels,
	}
}

func (s *UsageService) getCachedLeaderboardRecentTrend(key string, now time.Time) ([]usagestats.UserLeaderboardTokenTrendPoint, bool) {
	if s == nil || key == "" {
		return nil, false
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	entry, ok := s.recentTrendCache[key]
	if !ok || !now.Before(entry.expiresAt) {
		return nil, false
	}
	return cloneLeaderboardRecentTrend(entry.points), true
}

func (s *UsageService) setCachedLeaderboardRecentTrend(key string, points []usagestats.UserLeaderboardTokenTrendPoint, now time.Time) {
	if s == nil || key == "" {
		return
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	if s.recentTrendCache == nil {
		s.recentTrendCache = make(map[string]leaderboardRecentTrendCacheEntry)
	}
	for existingKey, entry := range s.recentTrendCache {
		if !now.Before(entry.expiresAt) {
			delete(s.recentTrendCache, existingKey)
		}
	}
	s.recentTrendCache[key] = leaderboardRecentTrendCacheEntry{
		expiresAt: now.Add(leaderboardStatsCacheTTL),
		points:    cloneLeaderboardRecentTrend(points),
	}
}

func (s *UsageService) getCachedLeaderboardDailyChampions(key string, now time.Time) ([]usagestats.UserLeaderboardDailyChampion, bool) {
	if s == nil || key == "" {
		return nil, false
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	entry, ok := s.dailyChampionsCache[key]
	if !ok || !now.Before(entry.expiresAt) {
		return nil, false
	}
	return cloneLeaderboardDailyChampions(entry.champions), true
}

func (s *UsageService) setCachedLeaderboardDailyChampions(key string, champions []usagestats.UserLeaderboardDailyChampion, now time.Time) {
	if s == nil || key == "" {
		return
	}
	s.leaderboardCacheMu.Lock()
	defer s.leaderboardCacheMu.Unlock()
	if s.dailyChampionsCache == nil {
		s.dailyChampionsCache = make(map[string]leaderboardDailyChampionsCacheEntry)
	}
	for existingKey, entry := range s.dailyChampionsCache {
		if !now.Before(entry.expiresAt) {
			delete(s.dailyChampionsCache, existingKey)
		}
	}
	s.dailyChampionsCache[key] = leaderboardDailyChampionsCacheEntry{
		expiresAt: now.Add(leaderboardStatsCacheTTL),
		champions: cloneLeaderboardDailyChampions(champions),
	}
}

func userLeaderboardStatsCacheKey(startTime, endTime time.Time, limit int, currentUserID int64) string {
	return strings.Join([]string{
		leaderboardStatsCacheKey(startTime, endTime, limit),
		fmt.Sprintf("user:%d", currentUserID),
	}, "|")
}

func leaderboardStatsCacheKey(startTime, endTime time.Time, limit int) string {
	return strings.Join([]string{
		leaderboardStatsTimeKey(startTime),
		leaderboardStatsTimeKey(endTime),
		fmt.Sprintf("limit:%d", limit),
	}, "|")
}

func leaderboardStatsTimeKey(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func leaderboardBadgeCacheKey(weekStart, weekEnd, monthStart, monthEnd, costStart, costEnd time.Time, userTZ string) string {
	return strings.Join([]string{
		normalizeLeaderboardBadgeCacheTimezone(userTZ),
		leaderboardBadgeTimeKey(weekStart),
		leaderboardBadgeTimeKey(weekEnd),
		leaderboardBadgeTimeKey(monthStart),
		leaderboardBadgeTimeKey(monthEnd),
		leaderboardBadgeTimeKey(costStart),
		leaderboardBadgeTimeKey(costEnd),
	}, "|")
}

func leaderboardBadgeTimeKey(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func leaderboardBadgeCacheExpiry(userTZ string, now time.Time) time.Time {
	loc := leaderboardBadgeCacheLocation(userTZ)
	localNow := now.In(loc)
	return time.Date(localNow.Year(), localNow.Month(), localNow.Day()+1, 0, 0, 0, 0, loc)
}

func normalizeLeaderboardBadgeCacheTimezone(userTZ string) string {
	value := strings.TrimSpace(userTZ)
	if value == "" {
		return apptimezone.Name()
	}
	if _, err := time.LoadLocation(value); err != nil {
		return apptimezone.Name()
	}
	return value
}

func leaderboardBadgeCacheLocation(userTZ string) *time.Location {
	value := normalizeLeaderboardBadgeCacheTimezone(userTZ)
	if loc, err := time.LoadLocation(value); err == nil {
		return loc
	}
	return apptimezone.Location()
}

func cloneLeaderboardBadgeLeaders(leaders *usagestats.UserLeaderboardBadgeLeaders) *usagestats.UserLeaderboardBadgeLeaders {
	if leaders == nil {
		return &usagestats.UserLeaderboardBadgeLeaders{}
	}
	clone := *leaders
	return &clone
}

func cloneUserLeaderboardResponse(payload *usagestats.UserLeaderboardResponse) *usagestats.UserLeaderboardResponse {
	if payload == nil {
		return nil
	}
	clone := *payload
	clone.Ranking = cloneUserLeaderboardItems(payload.Ranking)
	if payload.CurrentUserEntry != nil {
		current := cloneUserLeaderboardItem(*payload.CurrentUserEntry)
		clone.CurrentUserEntry = &current
	}
	clone.ModelRanking = cloneUserLeaderboardModelItems(payload.ModelRanking)
	clone.RecentTokenTrend = cloneLeaderboardRecentTrend(payload.RecentTokenTrend)
	clone.DailyChampions = cloneLeaderboardDailyChampions(payload.DailyChampions)
	clone.DailyRewards = cloneLeaderboardDailyRewards(payload.DailyRewards)
	return &clone
}

func cloneUserLeaderboardItems(items []usagestats.UserLeaderboardItem) []usagestats.UserLeaderboardItem {
	if items == nil {
		return nil
	}
	clones := make([]usagestats.UserLeaderboardItem, len(items))
	for i, item := range items {
		clones[i] = cloneUserLeaderboardItem(item)
	}
	return clones
}

func cloneUserLeaderboardItem(item usagestats.UserLeaderboardItem) usagestats.UserLeaderboardItem {
	if item.AvatarURL != nil {
		avatarURL := *item.AvatarURL
		item.AvatarURL = &avatarURL
	}
	if item.RankChange != nil {
		rankChange := *item.RankChange
		item.RankChange = &rankChange
	}
	item.Badges = cloneStrings(item.Badges)
	return item
}

func cloneUserLeaderboardModelItems(items []usagestats.UserLeaderboardModelItem) []usagestats.UserLeaderboardModelItem {
	if items == nil {
		return nil
	}
	clones := make([]usagestats.UserLeaderboardModelItem, len(items))
	for i, item := range items {
		if item.GrowthPercent != nil {
			growthPercent := *item.GrowthPercent
			item.GrowthPercent = &growthPercent
		}
		if item.RankChange != nil {
			rankChange := *item.RankChange
			item.RankChange = &rankChange
		}
		clones[i] = item
	}
	return clones
}

func cloneLeaderboardRecentTrend(points []usagestats.UserLeaderboardTokenTrendPoint) []usagestats.UserLeaderboardTokenTrendPoint {
	if points == nil {
		return nil
	}
	clones := make([]usagestats.UserLeaderboardTokenTrendPoint, len(points))
	copy(clones, points)
	return clones
}

func cloneLeaderboardDailyChampions(champions []usagestats.UserLeaderboardDailyChampion) []usagestats.UserLeaderboardDailyChampion {
	if champions == nil {
		return nil
	}
	clones := make([]usagestats.UserLeaderboardDailyChampion, len(champions))
	for i, champion := range champions {
		if champion.AvatarURL != nil {
			avatarURL := *champion.AvatarURL
			champion.AvatarURL = &avatarURL
		}
		clones[i] = champion
	}
	return clones
}

func cloneLeaderboardDailyRewards(rewards *usagestats.LeaderboardDailyRewards) *usagestats.LeaderboardDailyRewards {
	if rewards == nil {
		return nil
	}
	clone := *rewards
	if rewards.Rewards != nil {
		clone.Rewards = make([]usagestats.LeaderboardDailyRewardTier, len(rewards.Rewards))
		copy(clone.Rewards, rewards.Rewards)
	}
	if rewards.TopUsers != nil {
		clone.TopUsers = make([]usagestats.LeaderboardDailyRewardTopUser, len(rewards.TopUsers))
		copy(clone.TopUsers, rewards.TopUsers)
		for i := range clone.TopUsers {
			if clone.TopUsers[i].AvatarURL != nil {
				avatarURL := *clone.TopUsers[i].AvatarURL
				clone.TopUsers[i].AvatarURL = &avatarURL
			}
		}
	}
	return &clone
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	clones := make([]string, len(values))
	copy(clones, values)
	return clones
}

func leaderboardBadgeLeadersHasAny(leaders *usagestats.UserLeaderboardBadgeLeaders) bool {
	if leaders == nil {
		return false
	}
	return leaders.WeeklyTokenKingUserID > 0 ||
		leaders.MonthlyTokenKingUserID > 0 ||
		leaders.TotalTokenKingUserID > 0 ||
		leaders.NightOwlUserID > 0 ||
		leaders.BurstTokenKingUserID > 0 ||
		leaders.CheckinKingUserID > 0 ||
		leaders.CostSaverUserID > 0 ||
		leaders.CostBurnerUserID > 0
}

// GetUserUsageTrendByUserID returns per-user usage trend.
func (s *UsageService) GetUserUsageTrendByUserID(ctx context.Context, userID int64, startTime, endTime time.Time, granularity string) ([]usagestats.TrendDataPoint, error) {
	trend, err := s.usageRepo.GetUserUsageTrendByUserID(ctx, userID, startTime, endTime, granularity)
	if err != nil {
		return nil, fmt.Errorf("get user usage trend: %w", err)
	}
	return trend, nil
}

// GetUserModelStats returns per-user model usage stats.
func (s *UsageService) GetUserModelStats(ctx context.Context, userID int64, startTime, endTime time.Time) ([]usagestats.ModelStat, error) {
	stats, err := s.usageRepo.GetUserModelStats(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get user model stats: %w", err)
	}
	return stats, nil
}

// GetAPIKeyModelStats returns per-model usage stats for a specific API Key.
func (s *UsageService) GetAPIKeyModelStats(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) ([]usagestats.ModelStat, error) {
	stats, err := s.usageRepo.GetModelStatsWithFilters(ctx, startTime, endTime, 0, apiKeyID, 0, 0, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get api key model stats: %w", err)
	}
	return stats, nil
}

// GetAPIKeyDailyUsage returns daily usage stats for a user's API key.
func (s *UsageService) GetAPIKeyDailyUsage(ctx context.Context, userID, apiKeyID int64, startTime, endTime time.Time) ([]usagestats.APIKeyDailyUsagePoint, error) {
	trend, err := s.usageRepo.GetUsageTrendWithFilters(ctx, startTime, endTime, "day", userID, apiKeyID, 0, 0, "", nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get api key daily usage: %w", err)
	}

	points := make([]usagestats.APIKeyDailyUsagePoint, 0, len(trend))
	for _, row := range trend {
		points = append(points, usagestats.APIKeyDailyUsagePoint{
			Date:             row.Date,
			Requests:         row.Requests,
			InputTokens:      row.InputTokens,
			OutputTokens:     row.OutputTokens,
			CacheReadTokens:  row.CacheReadTokens,
			CacheWriteTokens: row.CacheCreationTokens,
			TotalTokens:      row.TotalTokens,
			Cost:             row.Cost,
			ActualCost:       row.ActualCost,
		})
	}
	return points, nil
}

// GetBatchAPIKeyUsageStats returns today/total actual_cost for given api keys.
func (s *UsageService) GetBatchAPIKeyUsageStats(ctx context.Context, apiKeyIDs []int64, startTime, endTime time.Time) (map[int64]*usagestats.BatchAPIKeyUsageStats, error) {
	stats, err := s.usageRepo.GetBatchAPIKeyUsageStats(ctx, apiKeyIDs, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get batch api key usage stats: %w", err)
	}
	return stats, nil
}

// ListWithFilters lists usage logs with admin filters.
func (s *UsageService) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters usagestats.UsageLogFilters) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, result, err := s.usageRepo.ListWithFilters(ctx, params, filters)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs with filters: %w", err)
	}
	return logs, result, nil
}

// GetGlobalStats returns global usage stats for a time range.
func (s *UsageService) GetGlobalStats(ctx context.Context, startTime, endTime time.Time) (*usagestats.UsageStats, error) {
	stats, err := s.usageRepo.GetGlobalStats(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get global usage stats: %w", err)
	}
	return stats, nil
}

// GetStatsWithFilters returns usage stats with optional filters.
func (s *UsageService) GetStatsWithFilters(ctx context.Context, filters usagestats.UsageLogFilters) (*usagestats.UsageStats, error) {
	stats, err := s.usageRepo.GetStatsWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("get usage stats with filters: %w", err)
	}
	return stats, nil
}

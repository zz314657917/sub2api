package handler

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// UsageHandler handles usage-related requests
type UsageHandler struct {
	usageService  *service.UsageService
	apiKeyService *service.APIKeyService
}

// NewUsageHandler creates a new UsageHandler
func NewUsageHandler(usageService *service.UsageService, apiKeyService *service.APIKeyService) *UsageHandler {
	return &UsageHandler{
		usageService:  usageService,
		apiKeyService: apiKeyService,
	}
}

// List handles listing usage records with pagination
// GET /api/v1/usage
func (h *UsageHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	var apiKeyID int64
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		id, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}

		// [Security Fix] Verify API Key ownership to prevent horizontal privilege escalation
		apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if apiKey.UserID != subject.UserID {
			response.Forbidden(c, "Not authorized to access this API key's usage records")
			return
		}

		apiKeyID = id
	}

	// Parse additional filters
	model := c.Query("model")

	var requestType *int16
	var stream *bool
	if requestTypeStr := strings.TrimSpace(c.Query("request_type")); requestTypeStr != "" {
		parsed, err := service.ParseUsageRequestType(requestTypeStr)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		value := int16(parsed)
		requestType = &value
	} else if streamStr := c.Query("stream"); streamStr != "" {
		val, err := strconv.ParseBool(streamStr)
		if err != nil {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return
		}
		stream = &val
	}

	var billingType *int8
	if billingTypeStr := c.Query("billing_type"); billingTypeStr != "" {
		val, err := strconv.ParseInt(billingTypeStr, 10, 8)
		if err != nil {
			response.BadRequest(c, "Invalid billing_type")
			return
		}
		bt := int8(val)
		billingType = &bt
	}

	// Parse date range
	var startTime, endTime *time.Time
	userTZ := c.Query("timezone") // Get user's timezone from request
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		t, err := timezone.ParseInUserLocation("2006-01-02", startDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return
		}
		startTime = &t
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		t, err := timezone.ParseInUserLocation("2006-01-02", endDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return
		}
		// Use half-open range [start, end), move to next calendar day start (DST-safe).
		t = t.AddDate(0, 0, 1)
		endTime = &t
	}

	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	filters := usagestats.UsageLogFilters{
		UserID:      subject.UserID, // Always filter by current user for security
		APIKeyID:    apiKeyID,
		Model:       model,
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	records, result, err := h.usageService.ListWithFilters(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UsageLog, 0, len(records))
	for i := range records {
		out = append(out, *dto.UsageLogFromService(&records[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// GetByID handles getting a single usage record
// GET /api/v1/usage/:id
func (h *UsageHandler) GetByID(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	usageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid usage ID")
		return
	}

	record, err := h.usageService.GetByID(c.Request.Context(), usageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 验证所有权
	if record.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this record")
		return
	}

	response.Success(c, dto.UsageLogFromService(record))
}

// Stats handles getting usage statistics
// GET /api/v1/usage/stats
func (h *UsageHandler) Stats(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var apiKeyID int64
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		id, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}

		// [Security Fix] Verify API Key ownership to prevent horizontal privilege escalation
		apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
		if err != nil {
			response.NotFound(c, "API key not found")
			return
		}
		if apiKey.UserID != subject.UserID {
			response.Forbidden(c, "Not authorized to access this API key's statistics")
			return
		}

		apiKeyID = id
	}

	// 获取时间范围参数
	userTZ := c.Query("timezone") // Get user's timezone from request
	now := timezone.NowInUserLocation(userTZ)
	var startTime, endTime time.Time

	// 优先使用 start_date 和 end_date 参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" && endDateStr != "" {
		// 使用自定义日期范围
		var err error
		startTime, err = timezone.ParseInUserLocation("2006-01-02", startDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return
		}
		endTime, err = timezone.ParseInUserLocation("2006-01-02", endDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return
		}
		// 与 SQL 条件 created_at < end 对齐，使用次日 00:00 作为上边界（DST-safe）。
		endTime = endTime.AddDate(0, 0, 1)
	} else {
		// 使用 period 参数
		period := c.DefaultQuery("period", "today")
		switch period {
		case "today":
			startTime = timezone.StartOfDayInUserLocation(now, userTZ)
		case "week":
			startTime = now.AddDate(0, 0, -7)
		case "month":
			startTime = now.AddDate(0, -1, 0)
		default:
			startTime = timezone.StartOfDayInUserLocation(now, userTZ)
		}
		endTime = now
	}

	var stats *service.UsageStats
	var err error
	if apiKeyID > 0 {
		stats, err = h.usageService.GetStatsByAPIKey(c.Request.Context(), apiKeyID, startTime, endTime)
	} else {
		stats, err = h.usageService.GetStatsByUser(c.Request.Context(), subject.UserID, startTime, endTime)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

// parseUserTimeRange parses start_date, end_date query parameters for user dashboard
// Uses user's timezone if provided, otherwise falls back to server timezone
func parseUserTimeRange(c *gin.Context) (time.Time, time.Time) {
	userTZ := c.Query("timezone") // Get user's timezone from request
	now := timezone.NowInUserLocation(userTZ)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var startTime, endTime time.Time

	if startDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", startDate, userTZ); err == nil {
			startTime = t
		} else {
			startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
		}
	} else {
		startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
	}

	if endDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", endDate, userTZ); err == nil {
			endTime = t.Add(24 * time.Hour) // Include the end date
		} else {
			endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
		}
	} else {
		endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
	}

	return startTime, endTime
}

const (
	defaultLeaderboardLimit = 10
	maxLeaderboardLimit     = 10
)

var (
	leaderboardPhonePattern      = regexp.MustCompile(`(?:\+?86[ -]*)?1[3-9][0-9][ -]*[0-9]{4}[ -]*[0-9]{4}`)
	leaderboardExplicitQQPattern = regexp.MustCompile(`(?i)qq[ :：-]*[0-9]{5,12}`)
)

func parseDashboardLeaderboardLimit(raw string) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || limit <= 0 {
		return defaultLeaderboardLimit
	}
	if limit > maxLeaderboardLimit {
		return maxLeaderboardLimit
	}
	return limit
}

func userLocation(userTZ string) *time.Location {
	if strings.TrimSpace(userTZ) != "" {
		if loc, err := time.LoadLocation(userTZ); err == nil {
			return loc
		}
	}
	return timezone.Location()
}

func startOfDayInLocation(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func parseDashboardLeaderboardPeriod(rawPeriod, userTZ string, now time.Time) (string, time.Time, time.Time, string, string, error) {
	period := strings.TrimSpace(rawPeriod)
	if period == "" {
		period = "day"
	}
	loc := userLocation(userTZ)
	now = now.In(loc)

	switch period {
	case "day":
		start := startOfDayInLocation(now, loc)
		end := start.AddDate(0, 0, 1)
		date := start.Format("2006-01-02")
		return period, start, end, date, date, nil
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := startOfDayInLocation(now.AddDate(0, 0, -weekday+1), loc)
		end := start.AddDate(0, 0, 7)
		return period, start, end, start.Format("2006-01-02"), end.AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0)
		return period, start, end, start.Format("2006-01-02"), end.AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "all":
		return period, time.Time{}, time.Time{}, "", now.Format("2006-01-02"), nil
	default:
		return "", time.Time{}, time.Time{}, "", "", errors.New("invalid leaderboard period")
	}
}

func dashboardLeaderboardWeekWindow(userTZ string, now time.Time) (time.Time, time.Time) {
	loc := userLocation(userTZ)
	now = now.In(loc)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := startOfDayInLocation(now.AddDate(0, 0, -weekday+1), loc)
	return start, start.AddDate(0, 0, 7)
}

func dashboardLeaderboardMonthWindow(userTZ string, now time.Time) (time.Time, time.Time) {
	loc := userLocation(userTZ)
	now = now.In(loc)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	return start, start.AddDate(0, 1, 0)
}

func dashboardLeaderboardRecentTrendWindow(userTZ string, now time.Time) (time.Time, time.Time) {
	loc := userLocation(userTZ)
	todayStart := startOfDayInLocation(now, loc)
	return todayStart.AddDate(0, 0, -9), todayStart.AddDate(0, 0, 1)
}

func fillLeaderboardRecentTokenTrend(points []usagestats.UserLeaderboardTokenTrendPoint, startTime, endTime time.Time) []usagestats.UserLeaderboardTokenTrendPoint {
	byDate := make(map[string]int64, len(points))
	for _, point := range points {
		byDate[point.Date] = point.TotalTokens
	}

	result := make([]usagestats.UserLeaderboardTokenTrendPoint, 0, 10)
	for cursor := startOfDayInLocation(startTime, startTime.Location()); cursor.Before(endTime); cursor = cursor.AddDate(0, 0, 1) {
		date := cursor.Format("2006-01-02")
		result = append(result, usagestats.UserLeaderboardTokenTrendPoint{
			Date:        date,
			TotalTokens: byDate[date],
		})
	}
	return result
}

func applyUserLeaderboardBadges(payload *usagestats.UserLeaderboardResponse, leaders *usagestats.UserLeaderboardBadgeLeaders) {
	if payload == nil || leaders == nil {
		return
	}
	apply := func(item *usagestats.UserLeaderboardItem) {
		if item == nil {
			return
		}
		item.Badges = leaderboardBadgesForUser(item.UserID, leaders)
	}
	for i := range payload.Ranking {
		apply(&payload.Ranking[i])
	}
	apply(payload.CurrentUserEntry)
}

func leaderboardBadgesForUser(userID int64, leaders *usagestats.UserLeaderboardBadgeLeaders) []string {
	badges := make([]string, 0, 4)
	if userID > 0 && leaders.WeeklyTokenKingUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeWeeklyTokenKing)
	}
	if userID > 0 && leaders.MonthlyTokenKingUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeMonthlyTokenKing)
	}
	if userID > 0 && leaders.TotalTokenKingUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeTotalTokenKing)
	}
	if userID > 0 && leaders.NightOwlUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeNightOwl)
	}
	if userID > 0 && leaders.BurstTokenKingUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeBurstTokenKing)
	}
	if userID > 0 && leaders.CheckinKingUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeCheckinKing)
	}
	if userID > 0 && leaders.CostSaverUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeCostSaver)
	}
	if userID > 0 && leaders.CostBurnerUserID == userID {
		badges = append(badges, usagestats.LeaderboardBadgeCostBurner)
	}
	return badges
}

func finalizeUserLeaderboardItem(item *usagestats.UserLeaderboardItem) {
	if item == nil {
		return
	}
	email := strings.TrimSpace(item.Email)
	username := strings.TrimSpace(item.Username)
	if email == "" && isLikelyEmailAddress(username) {
		email = username
	}
	if email != "" {
		item.EmailMasked = service.MaskEmail(email)
	} else {
		item.EmailMasked = ""
	}
	switch {
	case username != "" && !isLikelyEmailAddress(username):
		item.DisplayName = maskSensitiveLeaderboardDisplayName(username)
	case item.EmailMasked != "":
		item.DisplayName = item.EmailMasked
	default:
		item.DisplayName = "User #" + strconv.FormatInt(item.UserID, 10)
	}
	item.Email = ""
	item.Username = ""
}

func maskSensitiveLeaderboardDisplayName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if isDigitsOnly(value) {
		if isLikelyMainlandMobile(value) {
			return maskMainlandMobile(value)
		}
		if isLikelyQQNumber(value) {
			return maskQQNumber(value)
		}
		return value
	}

	masked := leaderboardPhonePattern.ReplaceAllStringFunc(value, func(match string) string {
		digits := digitsOnly(match)
		if strings.HasPrefix(digits, "86") && len(digits) == 13 {
			digits = digits[2:]
		}
		if !isLikelyMainlandMobile(digits) {
			return match
		}
		return maskMainlandMobile(digits)
	})

	return leaderboardExplicitQQPattern.ReplaceAllStringFunc(masked, func(match string) string {
		digitStart := firstDigitIndex(match)
		if digitStart < 0 {
			return match
		}
		digits := digitsOnly(match[digitStart:])
		if !isLikelyQQNumber(digits) {
			return match
		}
		return match[:digitStart] + maskQQNumber(digits)
	})
}

func isDigitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func digitsOnly(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] >= '0' && value[i] <= '9' {
			b.WriteByte(value[i])
		}
	}
	return b.String()
}

func firstDigitIndex(value string) int {
	for i := 0; i < len(value); i++ {
		if value[i] >= '0' && value[i] <= '9' {
			return i
		}
	}
	return -1
}

func isLikelyMainlandMobile(value string) bool {
	return len(value) == 11 && isDigitsOnly(value) && value[0] == '1' && value[1] >= '3' && value[1] <= '9'
}

func maskMainlandMobile(value string) string {
	if !isLikelyMainlandMobile(value) {
		return value
	}
	return value[:3] + "****" + value[7:]
}

func isLikelyQQNumber(value string) bool {
	return len(value) >= 5 && len(value) <= 12 && isDigitsOnly(value) && value[0] != '0'
}

func maskQQNumber(value string) string {
	if !isLikelyQQNumber(value) || len(value) <= 4 {
		return value
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

func isLikelyEmailAddress(value string) bool {
	value = strings.TrimSpace(value)
	if strings.Count(value, "@") != 1 || strings.ContainsAny(value, " \t\r\n") {
		return false
	}
	parts := strings.Split(value, "@")
	return parts[0] != "" && strings.Contains(parts[1], ".")
}

func finalizeUserLeaderboardResponse(payload *usagestats.UserLeaderboardResponse) {
	if payload == nil {
		return
	}
	for i := range payload.Ranking {
		finalizeUserLeaderboardItem(&payload.Ranking[i])
	}
	finalizeUserLeaderboardItem(payload.CurrentUserEntry)
}

// DashboardStats handles getting user dashboard statistics
// GET /api/v1/usage/dashboard/stats
func (h *UsageHandler) DashboardStats(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	stats, err := h.usageService.GetUserDashboardStats(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

// DashboardLeaderboard handles getting user-visible spending leaderboard data.
// GET /api/v1/usage/dashboard/leaderboard
func (h *UsageHandler) DashboardLeaderboard(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userTZ := c.Query("timezone")
	period, startTime, endTime, startDate, endDate, err := parseDashboardLeaderboardPeriod(c.DefaultQuery("period", "day"), userTZ, timezone.NowInUserLocation(userTZ))
	if err != nil {
		response.BadRequest(c, "Invalid period, use day/week/month/all")
		return
	}
	limit := parseDashboardLeaderboardLimit(c.DefaultQuery("limit", strconv.Itoa(defaultLeaderboardLimit)))

	leaderboard, err := h.usageService.GetUserLeaderboard(c.Request.Context(), startTime, endTime, limit, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	now := timezone.NowInUserLocation(userTZ)
	weekStart, weekEnd := dashboardLeaderboardWeekWindow(userTZ, now)
	monthStart, monthEnd := dashboardLeaderboardMonthWindow(userTZ, now)
	leaders, err := h.usageService.GetUserLeaderboardBadgeLeaders(c.Request.Context(), weekStart, weekEnd, monthStart, monthEnd, startTime, endTime, userTZ)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if leaderboard == nil {
		leaderboard = &usagestats.UserLeaderboardResponse{}
	}
	recentTrendStart, recentTrendEnd := dashboardLeaderboardRecentTrendWindow(userTZ, now)
	recentTrend, err := h.usageService.GetLeaderboardRecentTokenTrend(c.Request.Context(), recentTrendStart, recentTrendEnd)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	recentTrend = fillLeaderboardRecentTokenTrend(recentTrend, recentTrendStart, recentTrendEnd)
	leaderboard.Period = period
	leaderboard.StartDate = startDate
	leaderboard.EndDate = endDate
	leaderboard.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	leaderboard.RecentTokenTrend = recentTrend
	if leaderboard.Ranking == nil {
		leaderboard.Ranking = []usagestats.UserLeaderboardItem{}
	}
	applyUserLeaderboardBadges(leaderboard, leaders)
	dailyRewards, err := h.usageService.GetLeaderboardDailyRewards(c.Request.Context(), subject.UserID, userTZ)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	leaderboard.DailyRewards = dailyRewards
	finalizeUserLeaderboardResponse(leaderboard)

	response.Success(c, leaderboard)
}

// ClaimDashboardLeaderboardDailyReward handles claiming yesterday's top-3 reward.
// POST /api/v1/usage/dashboard/leaderboard/daily-reward/claim
func (h *UsageHandler) ClaimDashboardLeaderboardDailyReward(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	result, err := h.usageService.ClaimLeaderboardDailyReward(c.Request.Context(), subject.UserID, "")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// DashboardTrend handles getting user usage trend data
// GET /api/v1/usage/dashboard/trend
func (h *UsageHandler) DashboardTrend(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	startTime, endTime := parseUserTimeRange(c)
	granularity := c.DefaultQuery("granularity", "day")

	trend, err := h.usageService.GetUserUsageTrendByUserID(c.Request.Context(), subject.UserID, startTime, endTime, granularity)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  startTime.Format("2006-01-02"),
		"end_date":    endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"granularity": granularity,
	})
}

// DashboardModels handles getting user model usage statistics
// GET /api/v1/usage/dashboard/models
func (h *UsageHandler) DashboardModels(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	startTime, endTime := parseUserTimeRange(c)

	stats, err := h.usageService.GetUserModelStats(c.Request.Context(), subject.UserID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"models":     stats,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
	})
}

// BatchAPIKeysUsageRequest represents the request for batch API keys usage
type BatchAPIKeysUsageRequest struct {
	APIKeyIDs []int64 `json:"api_key_ids" binding:"required"`
}

// DashboardAPIKeysUsage handles getting usage stats for user's own API keys
// POST /api/v1/usage/dashboard/api-keys-usage
func (h *UsageHandler) DashboardAPIKeysUsage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req BatchAPIKeysUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if len(req.APIKeyIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	// Limit the number of API key IDs to prevent SQL parameter overflow
	if len(req.APIKeyIDs) > 100 {
		response.BadRequest(c, "Too many API key IDs (maximum 100 allowed)")
		return
	}

	validAPIKeyIDs, err := h.apiKeyService.VerifyOwnership(c.Request.Context(), subject.UserID, req.APIKeyIDs)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if len(validAPIKeyIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	stats, err := h.usageService.GetBatchAPIKeyUsageStats(c.Request.Context(), validAPIKeyIDs, time.Time{}, time.Time{})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"stats": stats})
}

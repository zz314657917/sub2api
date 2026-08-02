package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
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

var leaderboardSampleModelRanking = []usagestats.UserLeaderboardModelItem{
	{Rank: 1, Model: "gpt-5.5", Requests: 52, InputTokens: 6000000000, OutputTokens: 130000000, Tokens: 6130000000, GrowthPercent: leaderboardFloat64Ptr(-77.7), RankChange: leaderboardInt64Ptr(1)},
	{Rank: 2, Model: "claude-opus-4-8", Requests: 31, InputTokens: 2700000000, OutputTokens: 180000000, Tokens: 2880000000, GrowthPercent: leaderboardFloat64Ptr(-87.3), RankChange: leaderboardInt64Ptr(-1)},
	{Rank: 3, Model: "gpt-5.4", Requests: 10826, InputTokens: 1000000000, OutputTokens: 230000000, Tokens: 1230000000, GrowthPercent: leaderboardFloat64Ptr(-74.7), RankChange: nil},
}

type userUsageFilters struct {
	Filters   usagestats.UsageLogFilters
	StartTime time.Time
	EndTime   time.Time
}

type userModelStat struct {
	Model               string  `json:"model"`
	Requests            int64   `json:"requests"`
	InputTokens         int64   `json:"input_tokens"`
	OutputTokens        int64   `json:"output_tokens"`
	CacheCreationTokens int64   `json:"cache_creation_tokens"`
	CacheReadTokens     int64   `json:"cache_read_tokens"`
	TotalTokens         int64   `json:"total_tokens"`
	Cost                float64 `json:"cost"`
	ActualCost          float64 `json:"actual_cost"`
}

type userGroupStat struct {
	GroupID     int64   `json:"group_id"`
	GroupName   string  `json:"group_name"`
	Requests    int64   `json:"requests"`
	TotalTokens int64   `json:"total_tokens"`
	Cost        float64 `json:"cost"`
	ActualCost  float64 `json:"actual_cost"`
}

// UsageHandler handles usage-related requests
type UsageHandler struct {
	usageService   *service.UsageService
	apiKeyService  *service.APIKeyService
	opsService     *service.OpsService
	settingService *service.SettingService
}

// NewUsageHandler creates a new UsageHandler
func NewUsageHandler(usageService *service.UsageService, apiKeyService *service.APIKeyService, opsService *service.OpsService, settingService *service.SettingService) *UsageHandler {
	return &UsageHandler{
		usageService:   usageService,
		apiKeyService:  apiKeyService,
		opsService:     opsService,
		settingService: settingService,
	}
}

func (h *UsageHandler) parseUserUsageFilters(c *gin.Context, requireRange bool) (*userUsageFilters, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return nil, false
	}

	var apiKeyID int64
	if raw := strings.TrimSpace(c.Query("api_key_id")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return nil, false
		}
		if h.apiKeyService == nil {
			response.InternalError(c, "API key service not available")
			return nil, false
		}
		apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorFrom(c, err)
			return nil, false
		}
		if apiKey.UserID != subject.UserID {
			response.Forbidden(c, "Not authorized to access this API key's usage records")
			return nil, false
		}
		apiKeyID = id
	}

	var groupID int64
	if raw := strings.TrimSpace(c.Query("group_id")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 0 {
			response.BadRequest(c, "Invalid group_id")
			return nil, false
		}
		groupID = id
	}

	var requestType *int16
	var stream *bool
	if raw := strings.TrimSpace(c.Query("request_type")); raw != "" {
		parsed, err := service.ParseUsageRequestType(raw)
		if err != nil {
			response.BadRequest(c, err.Error())
			return nil, false
		}
		value := int16(parsed)
		requestType = &value
	} else if raw := strings.TrimSpace(c.Query("stream")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return nil, false
		}
		stream = &value
	}

	var billingType *int8
	if raw := strings.TrimSpace(c.Query("billing_type")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			response.BadRequest(c, "Invalid billing_type")
			return nil, false
		}
		parsed := int8(value)
		billingType = &parsed
	}
	billingMode := strings.TrimSpace(c.Query("billing_mode"))
	if billingMode != "" && !service.BillingMode(billingMode).IsValid() {
		response.BadRequest(c, "Invalid billing_mode")
		return nil, false
	}

	userTZ := c.Query("timezone")
	now := timezone.NowInUserLocation(userTZ)
	var startTime, endTime time.Time
	var startPtr, endPtr *time.Time
	if raw := strings.TrimSpace(c.Query("start_date")); raw != "" {
		value, err := timezone.ParseInUserLocation("2006-01-02", raw, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return nil, false
		}
		startTime = value
		startPtr = &startTime
	}
	if raw := strings.TrimSpace(c.Query("end_date")); raw != "" {
		value, err := timezone.ParseInUserLocation("2006-01-02", raw, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return nil, false
		}
		endTime = value.AddDate(0, 0, 1)
		endPtr = &endTime
	}
	if requireRange {
		if startPtr == nil {
			switch c.DefaultQuery("period", "") {
			case "today":
				startTime = timezone.StartOfDayInUserLocation(now, userTZ)
			case "week":
				startTime = now.AddDate(0, 0, -7)
			case "month":
				startTime = now.AddDate(0, -1, 0)
			default:
				startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
			}
			startPtr = &startTime
		}
		if endPtr == nil {
			if strings.TrimSpace(c.Query("period")) != "" {
				endTime = now
			} else {
				endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
			}
			endPtr = &endTime
		}
	}

	return &userUsageFilters{
		Filters: usagestats.UsageLogFilters{
			UserID: subject.UserID, APIKeyID: apiKeyID, GroupID: groupID,
			Model: strings.TrimSpace(c.Query("model")), ModelFilterSource: usagestats.ModelSourceRequested,
			RequestType: requestType, Stream: stream, BillingType: billingType, BillingMode: billingMode,
			StartTime: startPtr, EndTime: endPtr,
		},
		StartTime: startTime,
		EndTime:   endTime,
	}, true
}

// List handles listing usage records with pagination
// GET /api/v1/usage
func (h *UsageHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	parsed, ok := h.parseUserUsageFilters(c, false)
	if !ok {
		return
	}

	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	records, result, err := h.usageService.ListWithFilters(c.Request.Context(), params, parsed.Filters)
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

// ListErrors handles listing the current user's failed requests.
func (h *UsageHandler) ListErrors(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h.settingService == nil || !h.settingService.IsUserErrorViewAllowed(c.Request.Context()) {
		response.Forbidden(c, "Error requests view is disabled")
		return
	}
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 100 {
		pageSize = 100
	}
	filter := &service.OpsErrorLogFilter{Page: page, PageSize: pageSize}
	userTZ := c.Query("timezone")
	if raw := strings.TrimSpace(c.Query("start_date")); raw != "" {
		value, err := timezone.ParseInUserLocation("2006-01-02", raw, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return
		}
		filter.StartTime = &value
	}
	if raw := strings.TrimSpace(c.Query("end_date")); raw != "" {
		value, err := timezone.ParseInUserLocation("2006-01-02", raw, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return
		}
		value = value.AddDate(0, 0, 1)
		filter.EndTime = &value
	}
	filter.Model = strings.TrimSpace(c.Query("model"))
	if raw := strings.TrimSpace(c.Query("api_key_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		filter.APIKeyID = &value
	}
	if raw := strings.TrimSpace(c.Query("status_code")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			response.BadRequest(c, "Invalid status_code")
			return
		}
		filter.StatusCodes = []int{value}
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		filter.ErrorPhasesAny, filter.ErrorTypesAny = service.CategoryToFilter(category)
	}
	filter.SetSort(c.Query("sort_by"), c.Query("sort_order"))

	result, err := h.opsService.ListUserErrorRequests(c.Request.Context(), subject.UserID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, int64(result.Total), result.Page, result.PageSize)
}

// GetErrorDetail handles fetching one redacted failed-request detail.
func (h *UsageHandler) GetErrorDetail(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h.settingService == nil || !h.settingService.IsUserErrorViewAllowed(c.Request.Context()) {
		response.Forbidden(c, "Error requests view is disabled")
		return
	}
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return
	}
	detail, err := h.opsService.GetUserErrorRequestDetail(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, detail)
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

	var groupID int64
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		id, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		groupID = id
	}

	model := c.Query("model")
	billingMode := strings.TrimSpace(c.Query("billing_mode"))
	if billingMode != "" && !service.BillingMode(billingMode).IsValid() {
		response.BadRequest(c, "Invalid billing_mode")
		return
	}

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

	filters := usagestats.UsageLogFilters{
		UserID:            subject.UserID,
		APIKeyID:          apiKeyID,
		GroupID:           groupID,
		Model:             model,
		ModelFilterSource: usagestats.ModelSourceRequested,
		RequestType:       requestType,
		Stream:            stream,
		BillingType:       billingType,
		BillingMode:       billingMode,
		StartTime:         &startTime,
		EndTime:           &endTime,
	}
	stats, err := h.usageService.GetStatsWithFilters(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	stats.TotalAccountCost = nil
	stats.UpstreamEndpoints = nil
	stats.EndpointPaths = nil

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
	defaultLeaderboardLimit     = 10
	maxLeaderboardLimit         = 10
	defaultAPIKeyDailyUsageDays = 30
	maxAPIKeyDailyUsageDays     = 90
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

func dashboardLeaderboardChampionCalendarWindow(userTZ string, now time.Time) (time.Time, time.Time) {
	loc := userLocation(userTZ)
	now = now.In(loc)
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	return currentMonthStart.AddDate(0, -1, 0), currentMonthStart.AddDate(0, 1, 0)
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

func leaderboardFloat64Ptr(value float64) *float64 {
	return &value
}

func leaderboardInt64Ptr(value int64) *int64 {
	return &value
}

func shouldUseLeaderboardSampleModelRanking() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SUB2API_LEADERBOARD_SAMPLE_MODELS")), "true") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("SERVER_MODE")), "debug") && gin.Mode() == gin.DebugMode
}

func cloneLeaderboardSampleModelRanking(limit int) []usagestats.UserLeaderboardModelItem {
	if limit <= 0 || limit > len(leaderboardSampleModelRanking) {
		limit = len(leaderboardSampleModelRanking)
	}
	items := make([]usagestats.UserLeaderboardModelItem, limit)
	copy(items, leaderboardSampleModelRanking[:limit])
	return items
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

func finalizeLeaderboardDailyRewardTopUser(item *usagestats.LeaderboardDailyRewardTopUser) {
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
	item.DisplayName = hideLeaderboardRewardTopUserName(item.DisplayName)
	item.Email = ""
	item.Username = ""
	item.UserID = 0
}

func finalizeLeaderboardDailyRewards(payload *usagestats.LeaderboardDailyRewards) {
	if payload == nil {
		return
	}
	if payload.TopUsers == nil {
		payload.TopUsers = []usagestats.LeaderboardDailyRewardTopUser{}
	}
	for i := range payload.TopUsers {
		finalizeLeaderboardDailyRewardTopUser(&payload.TopUsers[i])
		if payload.TopUsers[i].LotteryWinner {
			payload.LotteryWinnerDisplayName = &payload.TopUsers[i].DisplayName
			payload.LotteryWinnerEmailMasked = &payload.TopUsers[i].EmailMasked
			payload.LotteryWinnerUserID = nil
		}
	}
}

func finalizeLeaderboardDailyChampion(item *usagestats.UserLeaderboardDailyChampion) {
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

func hideLeaderboardRewardTopUserName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "*") {
		return value
	}
	if isLikelyEmailAddress(value) {
		return service.MaskEmail(value)
	}
	runes := []rune(value)
	switch len(runes) {
	case 0:
		return ""
	case 1:
		return "*"
	case 2:
		return string(runes[:1]) + "*"
	case 3:
		return string(runes[:1]) + "*" + string(runes[2:])
	default:
		return string(runes[:1]) + "***" + string(runes[len(runes)-1:])
	}
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
	if payload.DailyChampions == nil {
		payload.DailyChampions = []usagestats.UserLeaderboardDailyChampion{}
	}
	for i := range payload.DailyChampions {
		finalizeLeaderboardDailyChampion(&payload.DailyChampions[i])
	}
	finalizeLeaderboardDailyRewards(payload.DailyRewards)
}

func parseAPIKeyDailyUsageDays(raw string) (int, bool) {
	if strings.TrimSpace(raw) == "" {
		return defaultAPIKeyDailyUsageDays, true
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 || days > maxAPIKeyDailyUsageDays {
		return 0, false
	}
	return days, true
}

func apiKeyDailyUsageRange(days int, userTZ string) (time.Time, time.Time) {
	now := timezone.NowInUserLocation(userTZ)
	startTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -(days-1)), userTZ)
	endTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
	return startTime, endTime
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
	if err := h.usageService.EnsureLeaderboardAccess(c.Request.Context(), subject.UserID); err != nil {
		response.ErrorFrom(c, err)
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
	championCalendarStart, championCalendarEnd := dashboardLeaderboardChampionCalendarWindow(userTZ, now)
	dailyChampions, err := h.usageService.GetLeaderboardDailyChampions(c.Request.Context(), championCalendarStart, championCalendarEnd)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	leaderboard.Period = period
	leaderboard.StartDate = startDate
	leaderboard.EndDate = endDate
	leaderboard.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	leaderboard.RecentTokenTrend = recentTrend
	leaderboard.DailyChampions = dailyChampions
	if leaderboard.Ranking == nil {
		leaderboard.Ranking = []usagestats.UserLeaderboardItem{}
	}
	modelRanking, totalModels := h.getOptionalLeaderboardModelRanking(c.Request.Context(), startTime, endTime, limit)
	leaderboard.ModelRanking = modelRanking
	leaderboard.TotalModels = totalModels
	if leaderboard.ModelRanking == nil {
		leaderboard.ModelRanking = []usagestats.UserLeaderboardModelItem{}
	}
	if len(leaderboard.ModelRanking) == 0 && shouldUseLeaderboardSampleModelRanking() {
		leaderboard.ModelRanking = cloneLeaderboardSampleModelRanking(limit)
		leaderboard.TotalModels = int64(len(leaderboard.ModelRanking))
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

func (h *UsageHandler) getOptionalLeaderboardModelRanking(ctx context.Context, startTime, endTime time.Time, limit int) ([]usagestats.UserLeaderboardModelItem, int64) {
	modelCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	modelRanking, totalModels, err := h.usageService.GetLeaderboardModelRanking(modelCtx, startTime, endTime, limit)
	if err != nil {
		return []usagestats.UserLeaderboardModelItem{}, 0
	}
	return modelRanking, totalModels
}

// ClaimDashboardLeaderboardDailyReward handles claiming last week's top-10 reward.
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
	finalizeLeaderboardDailyRewards(result.DailyRewards)
	response.Success(c, result)
}

// DashboardTrend handles getting user usage trend data
// GET /api/v1/usage/dashboard/trend
func (h *UsageHandler) DashboardTrend(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}
	granularity := c.DefaultQuery("granularity", "day")

	trend, err := h.usageService.GetUsageTrendWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, granularity, parsed.Filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  parsed.StartTime.Format("2006-01-02"),
		"end_date":    parsed.EndTime.AddDate(0, 0, -1).Format("2006-01-02"),
		"granularity": granularity,
	})
}

// DashboardModels handles getting user model usage statistics
// GET /api/v1/usage/dashboard/models
func (h *UsageHandler) DashboardModels(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}
	modelSource := strings.TrimSpace(c.Query("model_source"))
	if modelSource != "" && modelSource != usagestats.ModelSourceRequested {
		response.BadRequest(c, "Invalid model_source, user usage only supports requested")
		return
	}
	stats, err := h.usageService.GetModelStatsWithFiltersBySource(c.Request.Context(), parsed.StartTime, parsed.EndTime, parsed.Filters, usagestats.ModelSourceRequested)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"models":     userModelStatsFromUsageStats(stats),
		"start_date": parsed.StartTime.Format("2006-01-02"),
		"end_date":   parsed.EndTime.AddDate(0, 0, -1).Format("2006-01-02"),
	})
}

// DashboardSnapshotV2 returns usage-page chart data scoped to the current user.
func (h *UsageHandler) DashboardSnapshotV2(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}
	granularity := strings.TrimSpace(c.DefaultQuery("granularity", "day"))
	if granularity != "hour" {
		granularity = "day"
	}
	includeTrend, ok := parseBoolQueryWithDefault(c, "include_trend", true)
	if !ok {
		return
	}
	includeModels, ok := parseBoolQueryWithDefault(c, "include_model_stats", true)
	if !ok {
		return
	}
	includeGroups, ok := parseBoolQueryWithDefault(c, "include_group_stats", false)
	if !ok {
		return
	}

	payload := gin.H{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"start_date":   parsed.StartTime.Format("2006-01-02"),
		"end_date":     parsed.EndTime.AddDate(0, 0, -1).Format("2006-01-02"),
		"granularity":  granularity,
	}
	if includeTrend {
		trend, err := h.usageService.GetUsageTrendWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, granularity, parsed.Filters)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		payload["trend"] = trend
	}
	if includeModels {
		models, err := h.usageService.GetModelStatsWithFiltersBySource(c.Request.Context(), parsed.StartTime, parsed.EndTime, parsed.Filters, usagestats.ModelSourceRequested)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		payload["models"] = userModelStatsFromUsageStats(models)
	}
	if includeGroups {
		groups, err := h.usageService.GetGroupStatsWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, parsed.Filters)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		payload["groups"] = userGroupStatsFromUsageStats(groups)
	}
	response.Success(c, payload)
}

func userModelStatsFromUsageStats(stats []usagestats.ModelStat) []userModelStat {
	out := make([]userModelStat, 0, len(stats))
	for _, stat := range stats {
		out = append(out, userModelStat{
			Model: stat.Model, Requests: stat.Requests,
			InputTokens: stat.InputTokens, OutputTokens: stat.OutputTokens,
			CacheCreationTokens: stat.CacheCreationTokens, CacheReadTokens: stat.CacheReadTokens,
			TotalTokens: stat.TotalTokens, Cost: stat.Cost, ActualCost: stat.ActualCost,
		})
	}
	return out
}

func userGroupStatsFromUsageStats(stats []usagestats.GroupStat) []userGroupStat {
	out := make([]userGroupStat, 0, len(stats))
	for _, stat := range stats {
		out = append(out, userGroupStat{
			GroupID: stat.GroupID, GroupName: stat.GroupName, Requests: stat.Requests,
			TotalTokens: stat.TotalTokens, Cost: stat.Cost, ActualCost: stat.ActualCost,
		})
	}
	return out
}

func parseBoolQueryWithDefault(c *gin.Context, key string, fallback bool) (bool, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback, true
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		response.BadRequest(c, "Invalid "+key+" value, use true or false")
		return false, false
	}
	return parsed, true
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

// GetMyAPIKeyDailyUsage handles getting daily usage details for the current user's API key.
// GET /api/v1/user/api-keys/:id/usage/daily?days=30
func (h *UsageHandler) GetMyAPIKeyDailyUsage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	apiKeyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid API key ID")
		return
	}

	days, ok := parseAPIKeyDailyUsageDays(c.DefaultQuery("days", ""))
	if !ok {
		response.BadRequest(c, "Invalid days, allowed range is 1-90")
		return
	}

	if h.apiKeyService == nil {
		response.InternalError(c, "API key service is not configured")
		return
	}

	apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), apiKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if apiKey.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this API key's usage")
		return
	}

	userTZ := c.Query("timezone")
	startTime, endTime := apiKeyDailyUsageRange(days, userTZ)
	items, err := h.usageService.GetAPIKeyDailyUsage(c.Request.Context(), subject.UserID, apiKeyID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":      items,
		"days":       days,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.AddDate(0, 0, -1).Format("2006-01-02"),
	})
}

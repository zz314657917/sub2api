package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	TicketStatusOpen         = "open"
	TicketStatusPendingAdmin = "pending_admin"
	TicketStatusPendingUser  = "pending_user"
	TicketStatusClosed       = "closed"

	TicketSenderUser   = "user"
	TicketSenderAdmin  = "admin"
	TicketSenderSystem = "system"

	TicketTypeSupport = "support"
	TicketTypeSystem  = "system"

	SystemTicketKeyDefault = "system"
	SystemTicketTitle      = "系统通知"
)

const (
	maxTicketTitleLength   = 200
	maxTicketContentLength = 8000
	defaultTicketPageSize  = 20
	maxTicketPageSize      = 100
)

const (
	SystemTicketEventGroupChanged             = "group_changed"
	SystemTicketEventPaymentCompleted         = "payment_completed"
	SystemTicketEventAffiliateFirstAPIReward  = "affiliate_first_api_reward"
	SystemTicketEventWelfareFirstAPIUnclaimed = "welfare_first_api_unclaimed"
)

type SupportTicket struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"user_id"`
	Title              string     `json:"title"`
	Status             string     `json:"status"`
	TicketType         string     `json:"ticket_type"`
	SystemKey          string     `json:"system_key,omitempty"`
	LastMessagePreview string     `json:"last_message_preview"`
	LastMessageAt      time.Time  `json:"last_message_at"`
	UserUnreadCount    int        `json:"user_unread_count"`
	AdminUnreadCount   int        `json:"admin_unread_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
}

type SupportTicketMessage struct {
	ID           int64           `json:"id"`
	TicketID     int64           `json:"ticket_id"`
	SenderType   string          `json:"sender_type"`
	SenderUserID *int64          `json:"sender_user_id,omitempty"`
	Content      string          `json:"content"`
	EventType    string          `json:"event_type,omitempty"`
	EventKey     string          `json:"event_key,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type TicketListFilter struct {
	Status     string
	TicketType string
	Search     string
	UserID     int64
	EventType  string
	EventKey   string
	DateFrom   time.Time
	DateTo     time.Time
	UnreadOnly bool
	UnreadFor  string
	SortBy     string
	SortOrder  string
	Page       int
	PageSize   int
}

type CreateTicketInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AddTicketMessageInput struct {
	Content string `json:"content"`
}

type TicketDetail struct {
	Ticket   *SupportTicket         `json:"ticket"`
	Messages []SupportTicketMessage `json:"messages"`
}

type TicketUnreadSummary struct {
	SupportUnread int `json:"support_unread"`
	SystemUnread  int `json:"system_unread"`
	TotalUnread   int `json:"total_unread"`
}

type SystemTicketNotification struct {
	EventType string
	EventKey  string
	Content   string
	Metadata  map[string]any
}

type TicketRepository interface {
	ListTickets(ctx context.Context, filter TicketListFilter) ([]SupportTicket, int64, error)
	GetUserUnreadSummary(ctx context.Context, userID int64) (TicketUnreadSummary, error)
	CreateTicketWithMessage(ctx context.Context, userID int64, title string, content string) (*SupportTicket, error)
	CreateTicketForUserByAdmin(ctx context.Context, userID int64, adminID int64, title string, content string) (*SupportTicket, error)
	GetTicket(ctx context.Context, ticketID int64) (*SupportTicket, error)
	ListTicketMessages(ctx context.Context, ticketID int64) ([]SupportTicketMessage, error)
	AddTicketMessage(ctx context.Context, ticketID int64, senderType string, senderUserID *int64, content string) (*SupportTicketMessage, *SupportTicket, error)
	EnsureSystemTicket(ctx context.Context, userID int64) (*SupportTicket, error)
	AddSystemTicketMessage(ctx context.Context, ticketID int64, eventType string, eventKey string, content string, metadata json.RawMessage) (*SupportTicketMessage, *SupportTicket, bool, error)
	MarkTicketRead(ctx context.Context, ticketID int64, readerType string) (*SupportTicket, error)
	CloseTicket(ctx context.Context, ticketID int64) (*SupportTicket, error)
	ReopenTicket(ctx context.Context, ticketID int64) (*SupportTicket, error)
}

type TicketService struct {
	repo TicketRepository
}

func NewTicketService(repo TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) ListUserTickets(ctx context.Context, userID int64, filter TicketListFilter) ([]SupportTicket, int64, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, 0, err
	}
	filter.UserID = userID
	filter = normalizeTicketListFilter(filter)
	if filter.UnreadOnly {
		filter.UnreadFor = TicketSenderUser
	}
	if filter.TicketType != TicketTypeSupport {
		if _, err := s.repo.EnsureSystemTicket(ctx, userID); err != nil {
			return nil, 0, err
		}
	}
	return s.repo.ListTickets(ctx, filter)
}

func (s *TicketService) ListAdminTickets(ctx context.Context, filter TicketListFilter) ([]SupportTicket, int64, error) {
	filter = normalizeTicketListFilter(filter)
	if filter.UserID < 0 {
		return nil, 0, infraerrors.BadRequest("INVALID_USER_ID", "invalid user id")
	}
	if filter.UnreadOnly {
		filter.UnreadFor = TicketSenderAdmin
	}
	return s.repo.ListTickets(ctx, filter)
}

func (s *TicketService) GetUserUnreadSummary(ctx context.Context, userID int64) (TicketUnreadSummary, error) {
	if err := validateTicketUser(userID); err != nil {
		return TicketUnreadSummary{}, err
	}
	return s.repo.GetUserUnreadSummary(ctx, userID)
}

func (s *TicketService) CreateUserTicket(ctx context.Context, userID int64, input CreateTicketInput) (*SupportTicket, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, err
	}
	title, content, err := normalizeCreateTicketInput(input)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateTicketWithMessage(ctx, userID, title, content)
}

func (s *TicketService) CreateAdminTicketForUser(ctx context.Context, userID int64, adminID int64, input CreateTicketInput) (*SupportTicket, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, err
	}
	if err := validateAdminUser(adminID); err != nil {
		return nil, err
	}
	title, content, err := normalizeCreateTicketInput(input)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateTicketForUserByAdmin(ctx, userID, adminID, title, content)
}

func (s *TicketService) GetUserTicket(ctx context.Context, userID int64, ticketID int64) (*TicketDetail, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, err
	}
	ticket, err := s.getTicketForUser(ctx, userID, ticketID)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, ticket)
}

func (s *TicketService) GetAdminTicket(ctx context.Context, ticketID int64) (*TicketDetail, error) {
	ticket, err := s.getTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, ticket)
}

func (s *TicketService) AddUserMessage(ctx context.Context, userID int64, ticketID int64, input AddTicketMessageInput) (*TicketDetail, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, err
	}
	ticket, err := s.getTicketForUser(ctx, userID, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.Status == TicketStatusClosed {
		return nil, infraerrors.Forbidden("TICKET_CLOSED", "closed ticket cannot be replied by user")
	}
	if ticket.TicketType == TicketTypeSystem {
		return nil, infraerrors.Forbidden("SYSTEM_TICKET_READ_ONLY", "system ticket cannot be replied")
	}
	content, err := normalizeTicketContent(input.Content)
	if err != nil {
		return nil, err
	}
	_, updated, err := s.repo.AddTicketMessage(ctx, ticketID, TicketSenderUser, &userID, content)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, updated)
}

func (s *TicketService) AddAdminMessage(ctx context.Context, adminID int64, ticketID int64, input AddTicketMessageInput) (*TicketDetail, error) {
	if err := validateAdminUser(adminID); err != nil {
		return nil, err
	}
	ticket, err := s.getTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.Status == TicketStatusClosed {
		return nil, infraerrors.Forbidden("TICKET_CLOSED", "closed ticket must be reopened before replying")
	}
	if ticket.TicketType == TicketTypeSystem {
		return nil, infraerrors.Forbidden("SYSTEM_TICKET_READ_ONLY", "system ticket cannot be replied")
	}
	content, err := normalizeTicketContent(input.Content)
	if err != nil {
		return nil, err
	}
	_, updated, err := s.repo.AddTicketMessage(ctx, ticketID, TicketSenderAdmin, &adminID, content)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, updated)
}

func (s *TicketService) MarkUserRead(ctx context.Context, userID int64, ticketID int64) (*SupportTicket, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, err
	}
	if _, err := s.getTicketForUser(ctx, userID, ticketID); err != nil {
		return nil, err
	}
	return s.repo.MarkTicketRead(ctx, ticketID, TicketSenderUser)
}

func (s *TicketService) MarkAdminRead(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	if _, err := s.getTicket(ctx, ticketID); err != nil {
		return nil, err
	}
	return s.repo.MarkTicketRead(ctx, ticketID, TicketSenderAdmin)
}

func (s *TicketService) CloseUserTicket(ctx context.Context, userID int64, ticketID int64) (*SupportTicket, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, err
	}
	ticket, err := s.getTicketForUser(ctx, userID, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.TicketType == TicketTypeSystem {
		return nil, infraerrors.Forbidden("SYSTEM_TICKET_READ_ONLY", "system ticket cannot be closed")
	}
	return s.repo.CloseTicket(ctx, ticketID)
}

func (s *TicketService) CloseAdminTicket(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	ticket, err := s.getTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.TicketType == TicketTypeSystem {
		return nil, infraerrors.Forbidden("SYSTEM_TICKET_READ_ONLY", "system ticket cannot be closed")
	}
	return s.repo.CloseTicket(ctx, ticketID)
}

func (s *TicketService) ReopenAdminTicket(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	ticket, err := s.getTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.TicketType == TicketTypeSystem {
		return nil, infraerrors.Forbidden("SYSTEM_TICKET_READ_ONLY", "system ticket cannot be reopened")
	}
	return s.repo.ReopenTicket(ctx, ticketID)
}

type SystemTicketService struct {
	repo TicketRepository
}

func NewSystemTicketService(repo TicketRepository) *SystemTicketService {
	return &SystemTicketService{repo: repo}
}

func (s *SystemTicketService) EnsureSystemTicket(ctx context.Context, userID int64) (*SupportTicket, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, err
	}
	return s.repo.EnsureSystemTicket(ctx, userID)
}

func (s *SystemTicketService) NotifyUser(ctx context.Context, userID int64, eventType string, eventKey string, content string, metadata any) (*TicketDetail, bool, error) {
	if err := validateTicketUser(userID); err != nil {
		return nil, false, err
	}
	eventType = strings.TrimSpace(eventType)
	eventKey = strings.TrimSpace(eventKey)
	if eventType == "" {
		return nil, false, infraerrors.BadRequest("SYSTEM_EVENT_TYPE_REQUIRED", "system event type is required")
	}
	if eventKey == "" {
		return nil, false, infraerrors.BadRequest("SYSTEM_EVENT_KEY_REQUIRED", "system event key is required")
	}
	content, err := normalizeTicketContent(content)
	if err != nil {
		return nil, false, err
	}
	rawMetadata, err := json.Marshal(normalizeTicketMetadata(metadata))
	if err != nil {
		return nil, false, err
	}
	ticket, err := s.repo.EnsureSystemTicket(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	_, updated, created, err := s.repo.AddSystemTicketMessage(ctx, ticket.ID, eventType, eventKey, content, rawMetadata)
	if err != nil {
		return nil, false, err
	}
	if !created {
		updated = ticket
	}
	detail, err := (&TicketService{repo: s.repo}).detail(ctx, updated)
	if err != nil {
		return nil, false, err
	}
	return detail, created, nil
}

func (s *SystemTicketService) NotifyEventBestEffort(ctx context.Context, module string, userID int64, event SystemTicketNotification) {
	if s == nil {
		return
	}
	_, _, err := s.NotifyUser(ctx, userID, event.EventType, event.EventKey, event.Content, event.Metadata)
	if err != nil {
		logSystemTicketNotificationFailure(module, userID, event, err)
	}
}

func NewGroupChangedSystemTicketNotification(userID int64, source string, metadata map[string]any) SystemTicketNotification {
	merged := cloneTicketMetadata(metadata)
	merged["action_type"] = SystemTicketEventGroupChanged
	merged["user_id"] = userID
	merged["source"] = strings.TrimSpace(source)
	return SystemTicketNotification{
		EventType: SystemTicketEventGroupChanged,
		EventKey:  fmt.Sprintf("%s:%d:%d", SystemTicketEventGroupChanged, userID, time.Now().UnixNano()),
		Content:   buildGroupChangedSystemTicketContent(merged),
		Metadata:  merged,
	}
}

func buildGroupChangedSystemTicketContent(metadata map[string]any) string {
	groupName := strings.TrimSpace(ticketString(metadata["group_name"]))
	groupID := ticketInt64(metadata["group_id"])
	groupLabel := "你的配置"
	if groupName != "" {
		groupLabel = fmt.Sprintf("分组「%s」", groupName)
	} else if groupID > 0 {
		groupLabel = fmt.Sprintf("分组 #%d", groupID)
	}

	changes := make([]string, 0, 6)
	appendFloatChange := func(label string, oldKey string, newKey string, fallbackNewKey string) {
		if oldValue, okOld := ticketFloat(metadata[oldKey]); okOld {
			if newValue, okNew := ticketFloat(metadata[newKey]); okNew {
				if oldValue != newValue {
					changes = append(changes, fmt.Sprintf("%s：%s -> %s", label, formatTicketMultiplier(oldValue), formatTicketMultiplier(newValue)))
				}
				return
			}
		}
		if newValue, ok := ticketFloat(metadata[newKey]); ok {
			changes = append(changes, fmt.Sprintf("%s已更新为 %s", label, formatTicketMultiplier(newValue)))
			return
		}
		if fallbackNewKey != "" {
			if newValue, ok := ticketFloat(metadata[fallbackNewKey]); ok {
				changes = append(changes, fmt.Sprintf("%s已更新为 %s", label, formatTicketMultiplier(newValue)))
			}
		}
	}

	appendIntChange := func(label string, oldKey string, newKey string, fallbackNewKey string) {
		if oldValue, okOld := ticketInt(metadata[oldKey]); okOld {
			if newValue, okNew := ticketInt(metadata[newKey]); okNew {
				if oldValue != newValue {
					changes = append(changes, fmt.Sprintf("%s：%s -> %s", label, formatTicketLimit(oldValue), formatTicketLimit(newValue)))
				}
				return
			}
		}
		if newValue, ok := ticketInt(metadata[newKey]); ok {
			changes = append(changes, fmt.Sprintf("%s已更新为 %s", label, formatTicketLimit(newValue)))
			return
		}
		if fallbackNewKey != "" {
			if newValue, ok := ticketInt(metadata[fallbackNewKey]); ok {
				changes = append(changes, fmt.Sprintf("%s已更新为 %s", label, formatTicketLimit(newValue)))
			}
		}
	}

	appendFloatChange("计费倍率", "old_rate_multiplier", "new_rate_multiplier", "rate_multiplier")
	appendFloatChange("图片倍率", "old_image_rate_multiplier", "new_image_rate_multiplier", "image_rate_multiplier")
	appendIntChange("RPM 限制", "old_rpm_limit", "new_rpm_limit", "rpm_limit")
	appendGroupRateChanges(&changes, metadata["group_rate_changes"])
	appendRPMOverrideChanges(&changes, metadata["rpm_override_changes"])

	if oldGroupID := ticketInt64(metadata["old_group_id"]); oldGroupID > 0 {
		if newGroupID := ticketInt64(metadata["new_group_id"]); newGroupID > 0 && newGroupID != oldGroupID {
			changes = append(changes, fmt.Sprintf("专属分组：#%d -> #%d", oldGroupID, newGroupID))
		}
	}

	if len(changes) == 0 {
		return "你的可用分组、专属倍率或调用限制已更新，后续使用将按新的配置计费或限流。"
	}
	return groupLabel + "已更新：" + strings.Join(changes, "；") + "。"
}

func appendGroupRateChanges(changes *[]string, raw any) {
	items := ticketChangeList(raw)
	sort.Slice(items, func(i, j int) bool {
		return ticketInt64(items[i]["group_id"]) < ticketInt64(items[j]["group_id"])
	})
	for _, change := range items {
		groupID := ticketInt64(change["group_id"])
		label := "专属倍率"
		if groupID > 0 {
			label = fmt.Sprintf("分组 #%d 专属倍率", groupID)
		}
		oldValue, hasOld := ticketFloat(change["old_rate_multiplier"])
		newValue, hasNew := ticketFloat(change["new_rate_multiplier"])
		cleared, _ := ticketBool(change["cleared"])
		switch {
		case hasOld && hasNew && oldValue != newValue:
			*changes = append(*changes, fmt.Sprintf("%s：%s -> %s", label, formatTicketMultiplier(oldValue), formatTicketMultiplier(newValue)))
		case hasOld && cleared:
			*changes = append(*changes, fmt.Sprintf("%s：%s -> 使用分组默认倍率", label, formatTicketMultiplier(oldValue)))
		case hasNew:
			*changes = append(*changes, fmt.Sprintf("%s已更新为 %s", label, formatTicketMultiplier(newValue)))
		}
	}
}

func appendRPMOverrideChanges(changes *[]string, raw any) {
	items := ticketChangeList(raw)
	sort.Slice(items, func(i, j int) bool {
		return ticketInt64(items[i]["group_id"]) < ticketInt64(items[j]["group_id"])
	})
	for _, change := range items {
		groupID := ticketInt64(change["group_id"])
		label := "专属 RPM 限制"
		if groupID > 0 {
			label = fmt.Sprintf("分组 #%d 专属 RPM 限制", groupID)
		}
		oldValue, hasOld := ticketInt(change["old_rpm_override"])
		newValue, hasNew := ticketInt(change["new_rpm_override"])
		cleared, _ := ticketBool(change["cleared"])
		switch {
		case hasOld && hasNew && oldValue != newValue:
			*changes = append(*changes, fmt.Sprintf("%s：%s -> %s", label, formatTicketLimit(oldValue), formatTicketLimit(newValue)))
		case hasOld && cleared:
			*changes = append(*changes, fmt.Sprintf("%s：%s -> 使用分组默认限制", label, formatTicketLimit(oldValue)))
		case hasNew:
			*changes = append(*changes, fmt.Sprintf("%s已更新为 %s", label, formatTicketLimit(newValue)))
		}
	}
}

func ticketChangeList(raw any) []map[string]any {
	switch v := raw.(type) {
	case []map[string]any:
		return v
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func formatTicketMultiplier(value float64) string {
	return fmt.Sprintf("%gx", value)
}

func formatTicketLimit(value int) string {
	if value <= 0 {
		return "不限制"
	}
	return fmt.Sprintf("%d", value)
}

func ticketString(raw any) string {
	if value, ok := raw.(string); ok {
		return value
	}
	return ""
}

func ticketBool(raw any) (bool, bool) {
	value, ok := raw.(bool)
	return value, ok
}

func ticketFloat(raw any) (float64, bool) {
	switch v := raw.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case *float64:
		if v == nil {
			return 0, false
		}
		return *v, true
	default:
		return 0, false
	}
}

func ticketInt(raw any) (int, bool) {
	switch v := raw.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case *int:
		if v == nil {
			return 0, false
		}
		return *v, true
	default:
		return 0, false
	}
}

func ticketInt64(raw any) int64 {
	switch v := raw.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func NewPaymentCompletedSystemTicketNotification(orderID int64, outTradeNo string, orderType string, amount float64, payAmount float64, completedAt time.Time) SystemTicketNotification {
	content := fmt.Sprintf("你的订单 #%d 已到账，到账额度 %.4f。", orderID, amount)
	if orderType == "subscription" {
		content = fmt.Sprintf("你的订阅订单 #%d 已到账，订阅已生效。", orderID)
	}
	return SystemTicketNotification{
		EventType: SystemTicketEventPaymentCompleted,
		EventKey:  fmt.Sprintf("%s:%d", SystemTicketEventPaymentCompleted, orderID),
		Content:   content,
		Metadata: map[string]any{
			"action_type":  SystemTicketEventPaymentCompleted,
			"order_id":     orderID,
			"out_trade_no": outTradeNo,
			"order_type":   orderType,
			"amount":       amount,
			"pay_amount":   payAmount,
			"completed_at": completedAt.UTC().Format(time.RFC3339),
		},
	}
}

func NewAffiliateFirstAPIRewardSystemTicketNotification(inviteeUserID int64, amount float64, claimable bool) SystemTicketNotification {
	content := fmt.Sprintf("你邀请的用户 #%d 已完成首次 API 调用，可领取的邀请奖励 %.4f 已处理。", inviteeUserID, amount)
	if claimable {
		content = fmt.Sprintf("你邀请的用户 #%d 已完成首次 API 调用，有 %.4f 邀请奖励可领取。", inviteeUserID, amount)
	}
	return SystemTicketNotification{
		EventType: SystemTicketEventAffiliateFirstAPIReward,
		EventKey:  fmt.Sprintf("%s:%d", SystemTicketEventAffiliateFirstAPIReward, inviteeUserID),
		Content:   content,
		Metadata: map[string]any{
			"action_type":     SystemTicketEventAffiliateFirstAPIReward,
			"invitee_user_id": inviteeUserID,
			"amount":          amount,
			"claimable":       claimable,
		},
	}
}

func NewWelfareFirstAPIUnclaimedSystemTicketNotification(userID int64, trialID int64, rewardAmount float64, model string, apiKeyID int64, firstSuccessAt time.Time) SystemTicketNotification {
	return SystemTicketNotification{
		EventType: SystemTicketEventWelfareFirstAPIUnclaimed,
		EventKey:  fmt.Sprintf("%s:%d", SystemTicketEventWelfareFirstAPIUnclaimed, userID),
		Content:   fmt.Sprintf("你已完成首次 API 调用，可领取首次调用福利额度 %.4f。", rewardAmount),
		Metadata: map[string]any{
			"action_type":      SystemTicketEventWelfareFirstAPIUnclaimed,
			"user_id":          userID,
			"trial_id":         trialID,
			"reward_amount":    rewardAmount,
			"model":            model,
			"api_key_id":       apiKeyID,
			"first_success_at": firstSuccessAt.UTC().Format(time.RFC3339),
		},
	}
}

func cloneTicketMetadata(metadata map[string]any) map[string]any {
	cloned := make(map[string]any, len(metadata)+2)
	for k, v := range metadata {
		cloned[k] = v
	}
	return cloned
}

func logSystemTicketNotificationFailure(module string, userID int64, event SystemTicketNotification, err error) {
	slog.Warn("system ticket notification failed",
		"user_id", userID,
		"event_type", event.EventType,
		"event_key", event.EventKey,
		"module", strings.TrimSpace(module),
		"error", err,
	)
}

func (s *TicketService) getTicketForUser(ctx context.Context, userID int64, ticketID int64) (*SupportTicket, error) {
	ticket, err := s.getTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.UserID != userID {
		return nil, infraerrors.NotFound("TICKET_NOT_FOUND", "ticket not found")
	}
	return ticket, nil
}

func (s *TicketService) getTicket(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	if ticketID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_TICKET_ID", "invalid ticket id")
	}
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, infraerrors.NotFound("TICKET_NOT_FOUND", "ticket not found")
	}
	return ticket, nil
}

func (s *TicketService) detail(ctx context.Context, ticket *SupportTicket) (*TicketDetail, error) {
	messages, err := s.repo.ListTicketMessages(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}
	if messages == nil {
		messages = []SupportTicketMessage{}
	}
	return &TicketDetail{Ticket: ticket, Messages: messages}, nil
}

func normalizeTicketListFilter(filter TicketListFilter) TicketListFilter {
	filter.Status = strings.TrimSpace(filter.Status)
	if !isValidTicketStatus(filter.Status) {
		filter.Status = ""
	}
	filter.TicketType = strings.TrimSpace(filter.TicketType)
	if !isValidTicketType(filter.TicketType) {
		filter.TicketType = ""
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if len([]rune(filter.Search)) > 100 {
		filter.Search = string([]rune(filter.Search)[:100])
	}
	filter.EventType = strings.TrimSpace(filter.EventType)
	if len([]rune(filter.EventType)) > 100 {
		filter.EventType = string([]rune(filter.EventType)[:100])
	}
	filter.EventKey = strings.TrimSpace(filter.EventKey)
	if len([]rune(filter.EventKey)) > 200 {
		filter.EventKey = string([]rune(filter.EventKey)[:200])
	}
	if filter.UnreadFor != TicketSenderUser && filter.UnreadFor != TicketSenderAdmin {
		filter.UnreadFor = ""
	}
	filter.SortBy = strings.TrimSpace(filter.SortBy)
	if filter.SortBy != "" && filter.SortBy != "last_message_at" && filter.SortBy != "unread_first" {
		filter.SortBy = ""
	}
	filter.SortOrder = strings.ToLower(strings.TrimSpace(filter.SortOrder))
	if filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		filter.SortOrder = "desc"
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = defaultTicketPageSize
	}
	if filter.PageSize > maxTicketPageSize {
		filter.PageSize = maxTicketPageSize
	}
	if !filter.DateFrom.IsZero() && !filter.DateTo.IsZero() && filter.DateFrom.After(filter.DateTo) {
		filter.DateFrom, filter.DateTo = filter.DateTo, filter.DateFrom
	}
	return filter
}

func normalizeTicketMetadata(metadata any) any {
	if metadata == nil {
		return map[string]any{}
	}
	return metadata
}

func normalizeCreateTicketInput(input CreateTicketInput) (string, string, error) {
	title := strings.TrimSpace(input.Title)
	content, err := normalizeTicketContent(input.Content)
	if err != nil {
		return "", "", err
	}
	if title == "" {
		return "", "", infraerrors.BadRequest("TICKET_TITLE_REQUIRED", "ticket title is required")
	}
	if len([]rune(title)) > maxTicketTitleLength {
		return "", "", infraerrors.BadRequest("TICKET_TITLE_TOO_LONG", "ticket title is too long")
	}
	return title, content, nil
}

func normalizeTicketContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", infraerrors.BadRequest("TICKET_CONTENT_REQUIRED", "ticket content is required")
	}
	if len([]rune(content)) > maxTicketContentLength {
		return "", infraerrors.BadRequest("TICKET_CONTENT_TOO_LONG", "ticket content is too long")
	}
	return content, nil
}

func validateTicketUser(userID int64) error {
	if userID <= 0 {
		return infraerrors.Unauthorized("UNAUTHORIZED", "authentication required")
	}
	return nil
}

func validateAdminUser(userID int64) error {
	if userID <= 0 {
		return infraerrors.Unauthorized("UNAUTHORIZED", "admin authentication required")
	}
	return nil
}

func isValidTicketStatus(status string) bool {
	switch status {
	case "", TicketStatusOpen, TicketStatusPendingAdmin, TicketStatusPendingUser, TicketStatusClosed:
		return true
	default:
		return false
	}
}

func isValidTicketType(ticketType string) bool {
	switch ticketType {
	case "", TicketTypeSupport, TicketTypeSystem:
		return true
	default:
		return false
	}
}

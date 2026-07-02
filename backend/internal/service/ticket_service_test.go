package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestTicketService_UserOwnershipAndClosedReply(t *testing.T) {
	repo := newFakeTicketRepo()
	svc := NewTicketService(repo)
	ctx := context.Background()

	ticket, err := svc.CreateUserTicket(ctx, 10, CreateTicketInput{Title: "Need help", Content: "first message"})
	require.NoError(t, err)

	_, err = svc.GetUserTicket(ctx, 11, ticket.ID)
	require.Error(t, err)
	require.True(t, infraerrors.IsNotFound(err))

	closed, err := svc.CloseUserTicket(ctx, 10, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, TicketStatusClosed, closed.Status)
	require.NotNil(t, closed.ClosedAt)

	_, err = svc.AddUserMessage(ctx, 10, ticket.ID, AddTicketMessageInput{Content: "cannot reply"})
	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err))
}

func TestTicketService_UnreadStatusAndReopenFlow(t *testing.T) {
	repo := newFakeTicketRepo()
	svc := NewTicketService(repo)
	ctx := context.Background()

	ticket, err := svc.CreateUserTicket(ctx, 20, CreateTicketInput{Title: "Question", Content: "hello admin"})
	require.NoError(t, err)
	require.Equal(t, TicketStatusPendingAdmin, ticket.Status)
	require.Equal(t, 1, ticket.AdminUnreadCount)
	require.Equal(t, 0, ticket.UserUnreadCount)

	readByAdmin, err := svc.MarkAdminRead(ctx, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, 0, readByAdmin.AdminUnreadCount)

	detail, err := svc.AddAdminMessage(ctx, 1, ticket.ID, AddTicketMessageInput{Content: "hello user"})
	require.NoError(t, err)
	require.Equal(t, TicketStatusPendingUser, detail.Ticket.Status)
	require.Equal(t, 1, detail.Ticket.UserUnreadCount)
	require.Len(t, detail.Messages, 2)

	readByUser, err := svc.MarkUserRead(ctx, 20, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, 0, readByUser.UserUnreadCount)

	_, err = svc.CloseAdminTicket(ctx, ticket.ID)
	require.NoError(t, err)
	_, err = svc.AddAdminMessage(ctx, 1, ticket.ID, AddTicketMessageInput{Content: "closed reply"})
	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err))

	reopened, err := svc.ReopenAdminTicket(ctx, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, TicketStatusOpen, reopened.Status)
	require.Nil(t, reopened.ClosedAt)

	detail, err = svc.AddUserMessage(ctx, 20, ticket.ID, AddTicketMessageInput{Content: "after reopen"})
	require.NoError(t, err)
	require.Equal(t, TicketStatusPendingAdmin, detail.Ticket.Status)
	require.Equal(t, 1, detail.Ticket.AdminUnreadCount)
}

func TestTicketService_AdminCreateForUser(t *testing.T) {
	repo := newFakeTicketRepo()
	svc := NewTicketService(repo)
	ctx := context.Background()

	ticket, err := svc.CreateAdminTicketForUser(ctx, 30, 1, CreateTicketInput{Title: "Admin notice", Content: "please check"})
	require.NoError(t, err)
	require.Equal(t, int64(30), ticket.UserID)
	require.Equal(t, TicketStatusPendingUser, ticket.Status)
	require.Equal(t, 1, ticket.UserUnreadCount)
	require.Equal(t, 0, ticket.AdminUnreadCount)

	detail, err := svc.GetUserTicket(ctx, 30, ticket.ID)
	require.NoError(t, err)
	require.Len(t, detail.Messages, 1)
	require.Equal(t, TicketSenderAdmin, detail.Messages[0].SenderType)
}

func TestTicketService_ListUserTicketsEnsuresSystemTicket(t *testing.T) {
	repo := newFakeTicketRepo()
	svc := NewTicketService(repo)
	ctx := context.Background()

	items, total, err := svc.ListUserTickets(ctx, 30, TicketListFilter{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, TicketTypeSystem, items[0].TicketType)
	require.Equal(t, SystemTicketTitle, items[0].Title)

	items, total, err = svc.ListUserTickets(ctx, 30, TicketListFilter{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)

	items, total, err = svc.ListUserTickets(ctx, 31, TicketListFilter{TicketType: TicketTypeSupport})
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, items)
	require.Empty(t, repo.systemByUser[31])
}

func TestTicketService_UserUnreadSummary(t *testing.T) {
	repo := newFakeTicketRepo()
	ticketSvc := NewTicketService(repo)
	systemSvc := NewSystemTicketService(repo)
	ctx := context.Background()

	supportByAdmin, err := ticketSvc.CreateAdminTicketForUser(ctx, 30, 1, CreateTicketInput{Title: "Admin notice", Content: "please check"})
	require.NoError(t, err)
	supportByUser, err := ticketSvc.CreateUserTicket(ctx, 30, CreateTicketInput{Title: "Question", Content: "hello admin"})
	require.NoError(t, err)
	_, err = ticketSvc.AddAdminMessage(ctx, 1, supportByUser.ID, AddTicketMessageInput{Content: "hello user"})
	require.NoError(t, err)
	_, _, err = systemSvc.NotifyUser(ctx, 30, "payment_completed", "payment_completed:order-9", "充值已到账", nil)
	require.NoError(t, err)
	_, _, err = systemSvc.NotifyUser(ctx, 31, "payment_completed", "payment_completed:order-10", "其他用户通知", nil)
	require.NoError(t, err)

	summary, err := ticketSvc.GetUserUnreadSummary(ctx, 30)
	require.NoError(t, err)
	require.Equal(t, 2, summary.SupportUnread)
	require.Equal(t, 1, summary.SystemUnread)
	require.Equal(t, 3, summary.TotalUnread)

	_, err = ticketSvc.MarkUserRead(ctx, 30, supportByAdmin.ID)
	require.NoError(t, err)
	summary, err = ticketSvc.GetUserUnreadSummary(ctx, 30)
	require.NoError(t, err)
	require.Equal(t, 1, summary.SupportUnread)
	require.Equal(t, 1, summary.SystemUnread)
	require.Equal(t, 2, summary.TotalUnread)
}

func TestSystemTicketService_LazyCreateDeduplicateAndRead(t *testing.T) {
	repo := newFakeTicketRepo()
	svc := NewSystemTicketService(repo)
	ctx := context.Background()

	first, err := svc.EnsureSystemTicket(ctx, 40)
	require.NoError(t, err)
	second, err := svc.EnsureSystemTicket(ctx, 40)
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, TicketTypeSystem, first.TicketType)
	require.Equal(t, SystemTicketKeyDefault, first.SystemKey)

	detail, created, err := svc.NotifyUser(ctx, 40, "payment_completed", "payment_completed:order-1", "充值已到账", map[string]any{"order_id": "order-1"})
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, 1, detail.Ticket.UserUnreadCount)
	require.Len(t, detail.Messages, 1)
	require.Equal(t, TicketSenderSystem, detail.Messages[0].SenderType)
	require.Equal(t, "payment_completed", detail.Messages[0].EventType)

	detail, created, err = svc.NotifyUser(ctx, 40, "payment_completed", "payment_completed:order-1", "充值已到账", nil)
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, 1, detail.Ticket.UserUnreadCount)
	require.Len(t, detail.Messages, 1)

	ticketSvc := NewTicketService(repo)
	read, err := ticketSvc.MarkUserRead(ctx, 40, first.ID)
	require.NoError(t, err)
	require.Equal(t, 0, read.UserUnreadCount)
}

func TestSystemTicketNotificationTemplates(t *testing.T) {
	completedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	groupEvent := NewGroupChangedSystemTicketNotification(42, "admin_update", map[string]any{
		"group_id":            int64(7),
		"group_name":          "PLUS共享号池",
		"old_rate_multiplier": 0.06,
		"new_rate_multiplier": 0.08,
		"old_rpm_limit":       60,
		"new_rpm_limit":       120,
	})
	require.Equal(t, SystemTicketEventGroupChanged, groupEvent.EventType)
	require.Contains(t, groupEvent.EventKey, "group_changed:42:")
	require.Equal(t, SystemTicketEventGroupChanged, groupEvent.Metadata["action_type"])
	require.Equal(t, int64(42), groupEvent.Metadata["user_id"])
	require.Equal(t, int64(7), groupEvent.Metadata["group_id"])
	require.Contains(t, groupEvent.Content, "PLUS共享号池")
	require.Contains(t, groupEvent.Content, "计费倍率：0.06x -> 0.08x")
	require.Contains(t, groupEvent.Content, "RPM 限制：60 -> 120")

	allowedGroupEvent := NewGroupChangedSystemTicketNotification(42, "user_update", map[string]any{
		"allowed_group_changes": map[string]any{
			"added":   []int64{8},
			"removed": []int64{7},
		},
	})
	require.Contains(t, allowedGroupEvent.Content, "可用分组：增加 #8，移除 #7")

	unchangedRateEvent := NewGroupChangedSystemTicketNotification(42, "user_update", map[string]any{
		"group_rate_changes": []map[string]any{
			{
				"group_id":            int64(7),
				"old_rate_multiplier": 0.08,
				"new_rate_multiplier": 0.08,
			},
		},
	})
	require.NotContains(t, unchangedRateEvent.Content, "专属倍率已更新为 0.08x")
	require.NotContains(t, unchangedRateEvent.Content, "专属倍率：0.08x -> 0.08x")

	paymentEvent := NewPaymentCompletedSystemTicketNotification(99, "sub2_abc", "balance", 12.5, 12.3, completedAt)
	require.Equal(t, "payment_completed:99", paymentEvent.EventKey)
	require.Equal(t, SystemTicketEventPaymentCompleted, paymentEvent.Metadata["action_type"])
	require.Equal(t, int64(99), paymentEvent.Metadata["order_id"])
	require.Equal(t, "sub2_abc", paymentEvent.Metadata["out_trade_no"])
	require.Equal(t, "你的订单已到账，到账额度 12.5000。", paymentEvent.Content)
	require.NotContains(t, paymentEvent.Content, "#99")

	subscriptionPaymentEvent := NewPaymentCompletedSystemTicketNotification(100, "sub2_def", "subscription", 0, 0, completedAt)
	require.Equal(t, "你的订阅订单已到账，订阅已生效。", subscriptionPaymentEvent.Content)
	require.NotContains(t, subscriptionPaymentEvent.Content, "#100")

	issuedAt := completedAt.Add(time.Hour)
	invoiceEvent := NewInvoiceIssuedSystemTicketNotification(InvoiceRequestView{
		ID:          7,
		UserID:      42,
		Amount:      1999,
		Currency:    "CNY",
		InvoiceType: InvoiceTypeVATGeneral,
		Title:       "ACME",
		InvoiceNo:   "INV-20260701",
		FileName:    "invoice-7.pdf",
		IssuedAt:    &issuedAt,
	})
	require.Equal(t, "invoice_issued:7", invoiceEvent.EventKey)
	require.Equal(t, SystemTicketEventInvoiceIssued, invoiceEvent.Metadata["action_type"])
	require.Equal(t, int64(7), invoiceEvent.Metadata["invoice_request_id"])
	require.Equal(t, "invoice-7.pdf", invoiceEvent.Metadata["file_name"])
	require.Equal(t, "INV-20260701", invoiceEvent.Metadata["invoice_no"])
	require.Contains(t, invoiceEvent.Content, "你的发票已开具")
	require.Contains(t, invoiceEvent.Content, "1999.00 CNY")
	require.Contains(t, invoiceEvent.Content, "本工单消息中下载")

	affiliateEvent := NewAffiliateFirstAPIRewardSystemTicketNotification(88, 0.5, true)
	require.Equal(t, "affiliate_first_api_reward:88", affiliateEvent.EventKey)
	require.Equal(t, SystemTicketEventAffiliateFirstAPIReward, affiliateEvent.Metadata["action_type"])
	require.Equal(t, int64(88), affiliateEvent.Metadata["invitee_user_id"])
	require.Equal(t, true, affiliateEvent.Metadata["claimable"])

	welfareEvent := NewWelfareFirstAPIUnclaimedSystemTicketNotification(77, 66, 0.2, "gpt-test", 55, completedAt)
	require.Equal(t, "welfare_first_api_unclaimed:77", welfareEvent.EventKey)
	require.Equal(t, SystemTicketEventWelfareFirstAPIUnclaimed, welfareEvent.Metadata["action_type"])
	require.Equal(t, int64(77), welfareEvent.Metadata["user_id"])
	require.Equal(t, int64(66), welfareEvent.Metadata["trial_id"])
	require.Equal(t, "gpt-test", welfareEvent.Metadata["model"])
}

func TestAdminService_GroupChangedSystemNotificationSmoke(t *testing.T) {
	repo := newFakeTicketRepo()
	adminSvc := &adminServiceImpl{}
	adminSvc.SetSystemTicketService(NewSystemTicketService(repo))

	adminSvc.notifyGroupChangedBestEffort(context.Background(), 42, "api_key_group_update", map[string]any{
		"group_id":            int64(7),
		"group_name":          "PLUS共享号池",
		"old_rate_multiplier": 0.06,
		"new_rate_multiplier": 0.08,
		"old_rpm_limit":       60,
		"new_rpm_limit":       120,
	})

	got := requireSystemTicketNotification(t, repo, 42, SystemTicketEventGroupChanged, "")
	require.Contains(t, got.Message.EventKey, "group_changed:42:")
	require.Contains(t, got.Message.Content, "PLUS共享号池")
	require.Contains(t, got.Message.Content, "计费倍率：0.06x -> 0.08x")
	require.Contains(t, got.Message.Content, "RPM 限制：60 -> 120")
	require.Equal(t, "api_key_group_update", got.Metadata["source"])
	require.Equal(t, 0.06, got.Metadata["old_rate_multiplier"])
	require.Equal(t, 0.08, got.Metadata["new_rate_multiplier"])
	require.Equal(t, float64(60), got.Metadata["old_rpm_limit"])
	require.Equal(t, float64(120), got.Metadata["new_rpm_limit"])
}

func TestSystemTicketNotificationFailureLogFields(t *testing.T) {
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(previous)

	logSystemTicketNotificationFailure("service.test", 9, SystemTicketNotification{
		EventType: SystemTicketEventPaymentCompleted,
		EventKey:  "payment_completed:1",
	}, errors.New("boom"))

	out := buf.String()
	require.Contains(t, out, `"module":"service.test"`)
	require.Contains(t, out, `"user_id":9`)
	require.Contains(t, out, `"event_type":"payment_completed"`)
	require.Contains(t, out, `"event_key":"payment_completed:1"`)
	require.Contains(t, out, `"error":"boom"`)
}

func TestTicketService_SystemTicketIsReadOnly(t *testing.T) {
	repo := newFakeTicketRepo()
	systemSvc := NewSystemTicketService(repo)
	ticketSvc := NewTicketService(repo)
	ctx := context.Background()

	ticket, err := systemSvc.EnsureSystemTicket(ctx, 50)
	require.NoError(t, err)

	_, err = ticketSvc.AddUserMessage(ctx, 50, ticket.ID, AddTicketMessageInput{Content: "reply"})
	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err))

	_, err = ticketSvc.AddAdminMessage(ctx, 1, ticket.ID, AddTicketMessageInput{Content: "reply"})
	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err))

	_, err = ticketSvc.CloseUserTicket(ctx, 50, ticket.ID)
	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err))

	_, err = ticketSvc.CloseAdminTicket(ctx, ticket.ID)
	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err))

	_, err = ticketSvc.ReopenAdminTicket(ctx, ticket.ID)
	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err))
}

func TestTicketService_Validation(t *testing.T) {
	svc := NewTicketService(newFakeTicketRepo())
	ctx := context.Background()

	_, err := svc.CreateUserTicket(ctx, 1, CreateTicketInput{Title: "", Content: "content"})
	require.Error(t, err)
	require.Equal(t, "TICKET_TITLE_REQUIRED", infraerrors.Reason(err))

	_, err = svc.CreateUserTicket(ctx, 1, CreateTicketInput{Title: "title", Content: "  "})
	require.Error(t, err)
	require.Equal(t, "TICKET_CONTENT_REQUIRED", infraerrors.Reason(err))

	_, err = svc.CreateUserTicket(ctx, 0, CreateTicketInput{Title: "title", Content: "content"})
	require.Error(t, err)
	require.True(t, infraerrors.IsUnauthorized(err))
}

type fakeTicketRepo struct {
	nextTicketID  int64
	nextMessageID int64
	tickets       map[int64]*SupportTicket
	messages      map[int64][]SupportTicketMessage
	systemByUser  map[int64]int64
	eventKeys     map[int64]map[string]bool
}

func newFakeTicketRepo() *fakeTicketRepo {
	return &fakeTicketRepo{
		nextTicketID:  1,
		nextMessageID: 1,
		tickets:       map[int64]*SupportTicket{},
		messages:      map[int64][]SupportTicketMessage{},
		systemByUser:  map[int64]int64{},
		eventKeys:     map[int64]map[string]bool{},
	}
}

type systemTicketNotificationCheck struct {
	Message  SupportTicketMessage
	Metadata map[string]any
}

func requireSystemTicketNotification(t *testing.T, repo *fakeTicketRepo, userID int64, eventType string, eventKey string) systemTicketNotificationCheck {
	t.Helper()

	ticketID := repo.systemByUser[userID]
	require.NotZero(t, ticketID)
	ticket := repo.tickets[ticketID]
	require.NotNil(t, ticket)
	require.Equal(t, TicketTypeSystem, ticket.TicketType)
	require.Equal(t, SystemTicketKeyDefault, ticket.SystemKey)
	require.Equal(t, SystemTicketTitle, ticket.Title)
	require.Equal(t, TicketStatusOpen, ticket.Status)
	require.Equal(t, 1, ticket.UserUnreadCount)

	messages := repo.messages[ticketID]
	require.Len(t, messages, 1)
	msg := messages[0]
	require.Equal(t, TicketSenderSystem, msg.SenderType)
	require.Equal(t, eventType, msg.EventType)
	if eventKey != "" {
		require.Equal(t, eventKey, msg.EventKey)
	}
	require.NotEmpty(t, msg.Content)
	require.NotEmpty(t, msg.Metadata)

	var metadata map[string]any
	require.NoError(t, json.Unmarshal(msg.Metadata, &metadata))
	require.Equal(t, eventType, metadata["action_type"])

	return systemTicketNotificationCheck{
		Message:  msg,
		Metadata: metadata,
	}
}

func (r *fakeTicketRepo) ListTickets(ctx context.Context, filter TicketListFilter) ([]SupportTicket, int64, error) {
	_ = ctx
	items := make([]SupportTicket, 0)
	for _, ticket := range r.tickets {
		if filter.UserID > 0 && ticket.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && ticket.Status != filter.Status {
			continue
		}
		if filter.TicketType != "" && ticket.TicketType != filter.TicketType {
			continue
		}
		items = append(items, *ticket)
	}
	return items, int64(len(items)), nil
}

func (r *fakeTicketRepo) GetUserUnreadSummary(ctx context.Context, userID int64) (TicketUnreadSummary, error) {
	_ = ctx
	var summary TicketUnreadSummary
	for _, ticket := range r.tickets {
		if ticket.UserID != userID {
			continue
		}
		switch ticket.TicketType {
		case TicketTypeSystem:
			summary.SystemUnread += ticket.UserUnreadCount
		case TicketTypeSupport:
			summary.SupportUnread += ticket.UserUnreadCount
		}
		summary.TotalUnread += ticket.UserUnreadCount
	}
	return summary, nil
}

func (r *fakeTicketRepo) CreateTicketWithMessage(ctx context.Context, userID int64, title string, content string) (*SupportTicket, error) {
	_ = ctx
	return r.createTicket(userID, title, TicketSenderUser, &userID, content)
}

func (r *fakeTicketRepo) CreateTicketForUserByAdmin(ctx context.Context, userID int64, adminID int64, title string, content string) (*SupportTicket, error) {
	_ = ctx
	return r.createTicket(userID, title, TicketSenderAdmin, &adminID, content)
}

func (r *fakeTicketRepo) GetTicket(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	_ = ctx
	ticket := r.tickets[ticketID]
	if ticket == nil {
		return nil, nil
	}
	copy := *ticket
	return &copy, nil
}

func (r *fakeTicketRepo) ListTicketMessages(ctx context.Context, ticketID int64) ([]SupportTicketMessage, error) {
	_ = ctx
	items := r.messages[ticketID]
	out := append([]SupportTicketMessage(nil), items...)
	return out, nil
}

func (r *fakeTicketRepo) AddTicketMessage(ctx context.Context, ticketID int64, senderType string, senderUserID *int64, content string) (*SupportTicketMessage, *SupportTicket, error) {
	_ = ctx
	msg := r.addMessage(ticketID, senderType, senderUserID, content)
	ticket := r.tickets[ticketID]
	ticket.LastMessagePreview = content
	ticket.LastMessageAt = msg.CreatedAt
	ticket.UpdatedAt = msg.CreatedAt
	switch senderType {
	case TicketSenderAdmin:
		ticket.Status = TicketStatusPendingUser
		ticket.UserUnreadCount++
	default:
		ticket.Status = TicketStatusPendingAdmin
		ticket.AdminUnreadCount++
	}
	copy := *ticket
	return msg, &copy, nil
}

func (r *fakeTicketRepo) EnsureSystemTicket(ctx context.Context, userID int64) (*SupportTicket, error) {
	_ = ctx
	if id := r.systemByUser[userID]; id > 0 {
		copy := *r.tickets[id]
		return &copy, nil
	}
	now := time.Now()
	id := r.nextTicketID
	r.nextTicketID++
	ticket := &SupportTicket{
		ID:            id,
		UserID:        userID,
		Title:         SystemTicketTitle,
		Status:        TicketStatusOpen,
		TicketType:    TicketTypeSystem,
		SystemKey:     SystemTicketKeyDefault,
		LastMessageAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	r.tickets[id] = ticket
	r.systemByUser[userID] = id
	copy := *ticket
	return &copy, nil
}

func (r *fakeTicketRepo) AddSystemTicketMessage(ctx context.Context, ticketID int64, eventType string, eventKey string, content string, metadata json.RawMessage) (*SupportTicketMessage, *SupportTicket, bool, error) {
	_ = ctx
	if r.eventKeys[ticketID] == nil {
		r.eventKeys[ticketID] = map[string]bool{}
	}
	if r.eventKeys[ticketID][eventKey] {
		copy := *r.tickets[ticketID]
		return nil, &copy, false, nil
	}
	r.eventKeys[ticketID][eventKey] = true
	msg := r.addMessage(ticketID, TicketSenderSystem, nil, content)
	msg.EventType = eventType
	msg.EventKey = eventKey
	msg.Metadata = metadata
	messages := r.messages[ticketID]
	messages[len(messages)-1] = *msg
	r.messages[ticketID] = messages
	ticket := r.tickets[ticketID]
	ticket.Title = SystemTicketTitle
	ticket.Status = TicketStatusOpen
	ticket.TicketType = TicketTypeSystem
	ticket.SystemKey = SystemTicketKeyDefault
	ticket.LastMessagePreview = content
	ticket.LastMessageAt = msg.CreatedAt
	ticket.UserUnreadCount++
	ticket.UpdatedAt = msg.CreatedAt
	copy := *ticket
	return msg, &copy, true, nil
}

func (r *fakeTicketRepo) MarkTicketRead(ctx context.Context, ticketID int64, readerType string) (*SupportTicket, error) {
	_ = ctx
	ticket := r.tickets[ticketID]
	if readerType == TicketSenderUser {
		ticket.UserUnreadCount = 0
	} else {
		ticket.AdminUnreadCount = 0
	}
	copy := *ticket
	return &copy, nil
}

func (r *fakeTicketRepo) CloseTicket(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	_ = ctx
	now := time.Now()
	ticket := r.tickets[ticketID]
	ticket.Status = TicketStatusClosed
	ticket.ClosedAt = &now
	ticket.UpdatedAt = now
	copy := *ticket
	return &copy, nil
}

func (r *fakeTicketRepo) ReopenTicket(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	_ = ctx
	ticket := r.tickets[ticketID]
	ticket.Status = TicketStatusOpen
	ticket.ClosedAt = nil
	ticket.UpdatedAt = time.Now()
	copy := *ticket
	return &copy, nil
}

func (r *fakeTicketRepo) createTicket(userID int64, title string, senderType string, senderUserID *int64, content string) (*SupportTicket, error) {
	now := time.Now()
	id := r.nextTicketID
	r.nextTicketID++
	status := TicketStatusPendingAdmin
	userUnread := 0
	adminUnread := 1
	if senderType == TicketSenderAdmin {
		status = TicketStatusPendingUser
		userUnread = 1
		adminUnread = 0
	}
	ticket := &SupportTicket{
		ID:                 id,
		UserID:             userID,
		Title:              title,
		Status:             status,
		TicketType:         TicketTypeSupport,
		LastMessagePreview: content,
		LastMessageAt:      now,
		UserUnreadCount:    userUnread,
		AdminUnreadCount:   adminUnread,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	r.tickets[id] = ticket
	r.addMessage(id, senderType, senderUserID, content)
	copy := *ticket
	return &copy, nil
}

func (r *fakeTicketRepo) addMessage(ticketID int64, senderType string, senderUserID *int64, content string) *SupportTicketMessage {
	id := r.nextMessageID
	r.nextMessageID++
	msg := SupportTicketMessage{
		ID:           id,
		TicketID:     ticketID,
		SenderType:   senderType,
		SenderUserID: senderUserID,
		Content:      content,
		CreatedAt:    time.Now(),
	}
	r.messages[ticketID] = append(r.messages[ticketID], msg)
	return &msg
}

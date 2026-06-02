package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ticketRepository struct {
	db  *sql.DB
	sql sqlExecutor
}

func NewTicketRepository(sqlDB *sql.DB) service.TicketRepository {
	return &ticketRepository{db: sqlDB, sql: sqlDB}
}

func (r *ticketRepository) ListTickets(ctx context.Context, filter service.TicketListFilter) ([]service.SupportTicket, int64, error) {
	where, args := buildTicketWhere(filter)
	countQuery := `SELECT COUNT(*) FROM support_tickets` + where
	var total int64
	if err := scanSingleRow(ctx, r.sql, countQuery, args, &total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	queryArgs := append(append([]any{}, args...), filter.PageSize, offset)
	query := ticketSelectSQL() + where + `
		ORDER BY ` + ticketOrderBySQL(filter) + `
		LIMIT $` + placeholder(len(queryArgs)-1) + ` OFFSET $` + placeholder(len(queryArgs)) + `
	`
	rows, err := r.sql.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.SupportTicket, 0)
	for rows.Next() {
		item, err := scanTicket(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func ticketOrderBySQL(filter service.TicketListFilter) string {
	if filter.SortBy == "unread_first" {
		unreadField := "admin_unread_count"
		if filter.UnreadFor == service.TicketSenderUser {
			unreadField = "user_unread_count"
		}
		return `CASE WHEN ticket_type = 'system' THEN 0 ELSE 1 END, ` + unreadField + ` DESC, last_message_at DESC, id DESC`
	}
	if filter.SortOrder == "asc" {
		return `CASE WHEN ticket_type = 'system' THEN 0 ELSE 1 END, last_message_at ASC, id ASC`
	}
	return `CASE WHEN ticket_type = 'system' THEN 0 ELSE 1 END, last_message_at DESC, id DESC`
}

func (r *ticketRepository) GetUserUnreadSummary(ctx context.Context, userID int64) (service.TicketUnreadSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN ticket_type = 'support' THEN user_unread_count ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN ticket_type = 'system' THEN user_unread_count ELSE 0 END), 0),
			COALESCE(SUM(user_unread_count), 0)
		FROM support_tickets
		WHERE user_id = $1
	`
	var summary service.TicketUnreadSummary
	if err := scanSingleRow(ctx, r.sql, query, []any{userID}, &summary.SupportUnread, &summary.SystemUnread, &summary.TotalUnread); err != nil {
		return service.TicketUnreadSummary{}, err
	}
	return summary, nil
}

func (r *ticketRepository) CreateTicketWithMessage(ctx context.Context, userID int64, title string, content string) (*service.SupportTicket, error) {
	return r.createTicketWithMessage(ctx, userID, title, service.TicketSenderUser, &userID, content)
}

func (r *ticketRepository) CreateTicketForUserByAdmin(ctx context.Context, userID int64, adminID int64, title string, content string) (*service.SupportTicket, error) {
	return r.createTicketWithMessage(ctx, userID, title, service.TicketSenderAdmin, &adminID, content)
}

func (r *ticketRepository) createTicketWithMessage(ctx context.Context, userID int64, title string, senderType string, senderUserID *int64, content string) (*service.SupportTicket, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	status := service.TicketStatusPendingAdmin
	userUnread := 0
	adminUnread := 1
	if senderType == service.TicketSenderAdmin {
		status = service.TicketStatusPendingUser
		userUnread = 1
		adminUnread = 0
	}
	query := `
		INSERT INTO support_tickets (
			user_id, title, status, ticket_type, last_message_preview, last_message_at,
			user_unread_count, admin_unread_count
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6, $7)
		RETURNING id, user_id, title, status, ticket_type, COALESCE(system_key, ''), last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
	`
	var ticket service.SupportTicket
	if err = scanSingleRow(ctx, tx, query, []any{userID, title, status, service.TicketTypeSupport, previewTicketMessage(content), userUnread, adminUnread}, ticketScanDest(&ticket)...); err != nil {
		return nil, err
	}
	if _, err = insertTicketMessage(ctx, tx, ticket.ID, senderType, senderUserID, content); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) GetTicket(ctx context.Context, ticketID int64) (*service.SupportTicket, error) {
	query := ticketSelectSQL() + ` WHERE id = $1`
	var ticket service.SupportTicket
	if err := scanSingleRow(ctx, r.sql, query, []any{ticketID}, ticketScanDest(&ticket)...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) ListTicketMessages(ctx context.Context, ticketID int64) ([]service.SupportTicketMessage, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, ticket_id, sender_type, sender_user_id, content,
			COALESCE(event_type, ''), COALESCE(event_key, ''), COALESCE(metadata, '{}'::jsonb), created_at
		FROM support_ticket_messages
		WHERE ticket_id = $1
		ORDER BY created_at ASC, id ASC
	`, ticketID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.SupportTicketMessage, 0)
	for rows.Next() {
		item, err := scanTicketMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ticketRepository) AddTicketMessage(ctx context.Context, ticketID int64, senderType string, senderUserID *int64, content string) (*service.SupportTicketMessage, *service.SupportTicket, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	msg, err := insertTicketMessage(ctx, tx, ticketID, senderType, senderUserID, content)
	if err != nil {
		return nil, nil, err
	}

	status := service.TicketStatusPendingAdmin
	userUnreadDelta := 0
	adminUnreadDelta := 1
	if senderType == service.TicketSenderAdmin {
		status = service.TicketStatusPendingUser
		userUnreadDelta = 1
		adminUnreadDelta = 0
	}
	query := `
		UPDATE support_tickets
		SET status = $2,
			last_message_preview = $3,
			last_message_at = $4,
			user_unread_count = user_unread_count + $5,
			admin_unread_count = admin_unread_count + $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, title, status, ticket_type, COALESCE(system_key, ''), last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
	`
	var ticket service.SupportTicket
	if err = scanSingleRow(ctx, tx, query, []any{ticketID, status, previewTicketMessage(content), msg.CreatedAt, userUnreadDelta, adminUnreadDelta}, ticketScanDest(&ticket)...); err != nil {
		return nil, nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}
	return msg, &ticket, nil
}

func (r *ticketRepository) EnsureSystemTicket(ctx context.Context, userID int64) (*service.SupportTicket, error) {
	query := `
		INSERT INTO support_tickets (
			user_id, title, status, ticket_type, system_key, last_message_preview, last_message_at,
			user_unread_count, admin_unread_count
		)
		VALUES ($1, $2, $3, $4, $5, '', NOW(), 0, 0)
		ON CONFLICT (user_id, system_key) WHERE ticket_type = 'system' AND system_key IS NOT NULL
		DO UPDATE SET updated_at = support_tickets.updated_at
		RETURNING id, user_id, title, status, ticket_type, COALESCE(system_key, ''), last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
	`
	var ticket service.SupportTicket
	if err := scanSingleRow(ctx, r.sql, query, []any{userID, service.SystemTicketTitle, service.TicketStatusOpen, service.TicketTypeSystem, service.SystemTicketKeyDefault}, ticketScanDest(&ticket)...); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) AddSystemTicketMessage(ctx context.Context, ticketID int64, eventType string, eventKey string, content string, metadata json.RawMessage) (*service.SupportTicketMessage, *service.SupportTicket, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	msg, created, err := insertSystemTicketMessage(ctx, tx, ticketID, eventType, eventKey, content, metadata)
	if err != nil {
		return nil, nil, false, err
	}
	if !created {
		var ticket service.SupportTicket
		query := ticketSelectSQL() + ` WHERE id = $1`
		if err = scanSingleRow(ctx, tx, query, []any{ticketID}, ticketScanDest(&ticket)...); err != nil {
			return nil, nil, false, err
		}
		if err = tx.Commit(); err != nil {
			return nil, nil, false, err
		}
		return nil, &ticket, false, nil
	}

	query := `
		UPDATE support_tickets
		SET status = $2,
			title = $3,
			ticket_type = $4,
			system_key = COALESCE(system_key, $5),
			last_message_preview = $6,
			last_message_at = $7,
			user_unread_count = user_unread_count + 1,
			updated_at = NOW(),
			closed_at = NULL
		WHERE id = $1
		RETURNING id, user_id, title, status, ticket_type, COALESCE(system_key, ''), last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
	`
	var ticket service.SupportTicket
	if err = scanSingleRow(ctx, tx, query, []any{ticketID, service.TicketStatusOpen, service.SystemTicketTitle, service.TicketTypeSystem, service.SystemTicketKeyDefault, previewTicketMessage(content), msg.CreatedAt}, ticketScanDest(&ticket)...); err != nil {
		return nil, nil, false, err
	}
	if err = tx.Commit(); err != nil {
		return nil, nil, false, err
	}
	return msg, &ticket, true, nil
}

func (r *ticketRepository) MarkTicketRead(ctx context.Context, ticketID int64, readerType string) (*service.SupportTicket, error) {
	field := "admin_unread_count"
	if readerType == service.TicketSenderUser {
		field = "user_unread_count"
	}
	query := `
		UPDATE support_tickets
		SET ` + field + ` = 0, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, title, status, ticket_type, COALESCE(system_key, ''), last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
	`
	var ticket service.SupportTicket
	if err := scanSingleRow(ctx, r.sql, query, []any{ticketID}, ticketScanDest(&ticket)...); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) CloseTicket(ctx context.Context, ticketID int64) (*service.SupportTicket, error) {
	query := `
		UPDATE support_tickets
		SET status = $2,
			closed_at = COALESCE(closed_at, NOW()),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, title, status, ticket_type, COALESCE(system_key, ''), last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
	`
	var ticket service.SupportTicket
	if err := scanSingleRow(ctx, r.sql, query, []any{ticketID, service.TicketStatusClosed}, ticketScanDest(&ticket)...); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) ReopenTicket(ctx context.Context, ticketID int64) (*service.SupportTicket, error) {
	query := `
		UPDATE support_tickets
		SET status = $2,
			closed_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, title, status, ticket_type, COALESCE(system_key, ''), last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
	`
	var ticket service.SupportTicket
	if err := scanSingleRow(ctx, r.sql, query, []any{ticketID, service.TicketStatusOpen}, ticketScanDest(&ticket)...); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func buildTicketWhere(filter service.TicketListFilter) (string, []any) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
		clauses = append(clauses, `user_id = $`+placeholder(len(args)))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, `status = $`+placeholder(len(args)))
	}
	if filter.TicketType != "" {
		args = append(args, filter.TicketType)
		clauses = append(clauses, `ticket_type = $`+placeholder(len(args)))
	}
	if filter.EventType != "" {
		args = append(args, filter.EventType)
		eventTypePlaceholder := placeholder(len(args))
		clauses = append(clauses, fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM support_ticket_messages stm_event_type
			WHERE stm_event_type.ticket_id = support_tickets.id
				AND stm_event_type.event_type = $%s
		)`, eventTypePlaceholder))
	}
	if filter.EventKey != "" {
		args = append(args, "%"+filter.EventKey+"%")
		eventKeyPlaceholder := placeholder(len(args))
		clauses = append(clauses, fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM support_ticket_messages stm_event_key
			WHERE stm_event_key.ticket_id = support_tickets.id
				AND stm_event_key.event_key ILIKE $%s
		)`, eventKeyPlaceholder))
	}
	if !filter.DateFrom.IsZero() {
		args = append(args, filter.DateFrom.UTC())
		clauses = append(clauses, `last_message_at >= $`+placeholder(len(args)))
	}
	if !filter.DateTo.IsZero() {
		args = append(args, endOfTicketFilterDay(filter.DateTo))
		clauses = append(clauses, `last_message_at < $`+placeholder(len(args)))
	}
	if filter.UnreadOnly {
		switch filter.UnreadFor {
		case service.TicketSenderUser:
			clauses = append(clauses, `user_unread_count > 0`)
		case service.TicketSenderAdmin:
			clauses = append(clauses, `admin_unread_count > 0`)
		}
	}
	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		searchPlaceholder := placeholder(len(args))
		clauses = append(clauses, fmt.Sprintf(`(
			title ILIKE $%[1]s
			OR last_message_preview ILIKE $%[1]s
			OR EXISTS (
				SELECT 1
				FROM support_ticket_messages stm
				WHERE stm.ticket_id = support_tickets.id AND stm.content ILIKE $%[1]s
			)
			OR EXISTS (
				SELECT 1
				FROM users u
				WHERE u.id = support_tickets.user_id
					AND (u.email ILIKE $%[1]s OR u.username ILIKE $%[1]s)
			)
		)`, searchPlaceholder))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func endOfTicketFilterDay(value time.Time) time.Time {
	utc := value.UTC()
	if utc.Hour() == 0 && utc.Minute() == 0 && utc.Second() == 0 && utc.Nanosecond() == 0 {
		return utc.AddDate(0, 0, 1)
	}
	return utc
}

func ticketSelectSQL() string {
	return `
		SELECT id, user_id, title, status, ticket_type, COALESCE(system_key, ''),
			last_message_preview, last_message_at,
			user_unread_count, admin_unread_count, created_at, updated_at, closed_at
		FROM support_tickets
	`
}

func scanTicket(row interface{ Scan(dest ...any) error }) (service.SupportTicket, error) {
	var ticket service.SupportTicket
	if err := row.Scan(ticketScanDest(&ticket)...); err != nil {
		return ticket, err
	}
	return ticket, nil
}

func ticketScanDest(ticket *service.SupportTicket) []any {
	return []any{
		&ticket.ID,
		&ticket.UserID,
		&ticket.Title,
		&ticket.Status,
		&ticket.TicketType,
		&ticket.SystemKey,
		&ticket.LastMessagePreview,
		&ticket.LastMessageAt,
		&ticket.UserUnreadCount,
		&ticket.AdminUnreadCount,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
		&ticket.ClosedAt,
	}
}

func insertTicketMessage(ctx context.Context, exec sqlExecutor, ticketID int64, senderType string, senderUserID *int64, content string) (*service.SupportTicketMessage, error) {
	query := `
		INSERT INTO support_ticket_messages (ticket_id, sender_type, sender_user_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, ticket_id, sender_type, sender_user_id, content,
			COALESCE(event_type, ''), COALESCE(event_key, ''), COALESCE(metadata, '{}'::jsonb), created_at
	`
	return scanSingleTicketMessage(ctx, exec, query, []any{ticketID, senderType, senderUserID, content})
}

func insertSystemTicketMessage(ctx context.Context, exec sqlExecutor, ticketID int64, eventType string, eventKey string, content string, metadata json.RawMessage) (*service.SupportTicketMessage, bool, error) {
	query := `
		INSERT INTO support_ticket_messages (ticket_id, sender_type, sender_user_id, content, event_type, event_key, metadata)
		VALUES ($1, $2, NULL, $3, $4, $5, $6::jsonb)
		ON CONFLICT (ticket_id, event_key) WHERE event_key IS NOT NULL AND event_key <> ''
		DO NOTHING
		RETURNING id, ticket_id, sender_type, sender_user_id, content,
			COALESCE(event_type, ''), COALESCE(event_key, ''), COALESCE(metadata, '{}'::jsonb), created_at
	`
	msg, err := scanSingleTicketMessage(ctx, exec, query, []any{ticketID, service.TicketSenderSystem, content, eventType, eventKey, string(metadata)})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return msg, true, nil
}

func scanSingleTicketMessage(ctx context.Context, q sqlQueryer, query string, args []any) (msg *service.SupportTicketMessage, err error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	msg, err = scanTicketMessage(rows)
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return msg, nil
}

func scanTicketMessage(row interface{ Scan(dest ...any) error }) (*service.SupportTicketMessage, error) {
	var msg service.SupportTicketMessage
	var metadata []byte
	if err := row.Scan(
		&msg.ID,
		&msg.TicketID,
		&msg.SenderType,
		&msg.SenderUserID,
		&msg.Content,
		&msg.EventType,
		&msg.EventKey,
		&metadata,
		&msg.CreatedAt,
	); err != nil {
		return nil, err
	}
	if len(metadata) == 0 {
		msg.Metadata = json.RawMessage("{}")
	} else {
		msg.Metadata = append(json.RawMessage(nil), metadata...)
	}
	return &msg, nil
}

func previewTicketMessage(content string) string {
	const maxPreviewRunes = 120
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if len(runes) <= maxPreviewRunes {
		return content
	}
	return string(runes[:maxPreviewRunes])
}

func placeholder(index int) string {
	return strconv.Itoa(index)
}

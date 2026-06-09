package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type studioBridgeRepository struct {
	db *sql.DB
}

func NewStudioBridgeRepository(db *sql.DB) service.StudioBridgeRepository {
	return &studioBridgeRepository{db: db}
}

func (r *studioBridgeRepository) GetUserSummary(ctx context.Context, userID int64, rechargeURL string, usageLimit int) (*service.StudioBridgeUserSummary, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("studio bridge repository db is nil")
	}
	var summary service.StudioBridgeUserSummary
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, username, balance
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&summary.UserID, &summary.Email, &summary.Username, &summary.Balance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	summary.RechargeURL = rechargeURL
	if usageLimit > 0 {
		usageRows, err := r.db.QueryContext(ctx, `
			SELECT request_id, COALESCE(NULLIF(requested_model, ''), model), actual_cost, created_at
			FROM usage_logs
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, userID, usageLimit)
		if err != nil {
			return nil, err
		}
		defer usageRows.Close()
		for usageRows.Next() {
			var item service.StudioBridgeUsageSummary
			if err := usageRows.Scan(&item.RequestID, &item.Model, &item.ActualCost, &item.CreatedAt); err != nil {
				return nil, err
			}
			summary.Usage = append(summary.Usage, item)
		}
		if err := usageRows.Err(); err != nil {
			return nil, err
		}
		orderRows, err := r.db.QueryContext(ctx, `
			SELECT id, amount, status, created_at, paid_at
			FROM payment_orders
			WHERE user_id = $1 AND order_type = 'balance'
			ORDER BY created_at DESC
			LIMIT 10
		`, userID)
		if err != nil {
			return nil, err
		}
		defer orderRows.Close()
		for orderRows.Next() {
			var item service.StudioBridgeRechargeOrder
			if err := orderRows.Scan(&item.ID, &item.Amount, &item.Status, &item.CreatedAt, &item.PaidAt); err != nil {
				return nil, err
			}
			summary.Orders = append(summary.Orders, item)
		}
		if err := orderRows.Err(); err != nil {
			return nil, err
		}
	}
	if summary.Usage == nil {
		summary.Usage = []service.StudioBridgeUsageSummary{}
	}
	if summary.Orders == nil {
		summary.Orders = []service.StudioBridgeRechargeOrder{}
	}
	return &summary, nil
}

func (r *studioBridgeRepository) ReserveCharge(ctx context.Context, userID int64, amount float64) (float64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("studio bridge repository db is nil")
	}
	var balance float64
	err := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET balance = balance - $1,
			updated_at = NOW()
		WHERE id = $2
			AND deleted_at IS NULL
			AND balance >= $1
		RETURNING balance
	`, amount, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		var exists bool
		if e := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`, userID).Scan(&exists); e != nil {
			return 0, e
		}
		if exists {
			return 0, service.ErrStudioBridgeInsufficient
		}
		return 0, service.ErrUserNotFound
	}
	return balance, err
}

func (r *studioBridgeRepository) RefundCharge(ctx context.Context, userID int64, amount float64) (float64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("studio bridge repository db is nil")
	}
	var balance float64
	err := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET balance = balance + $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING balance
	`, amount, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrUserNotFound
	}
	return balance, err
}

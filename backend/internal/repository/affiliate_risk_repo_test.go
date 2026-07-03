package repository

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAffiliateRiskRepositoryListClusters(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &affiliateRepository{db: db}
	start := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	end := start.Add(12 * time.Hour)
	rewardAt := start.Add(10 * time.Minute)

	rows := sqlmock.NewRows([]string{
		"inviter_id",
		"inviter_email",
		"inviter_username",
		"inviter_register_ip",
		"inviter_last_login_ip",
		"invitee_id",
		"invitee_email",
		"invitee_username",
		"invitee_register_ip",
		"invitee_last_login_ip",
		"created_at",
		"affiliate_revoked_at",
		"affiliate_revoked_reason",
		"first_usage_at",
		"first_usage_ip",
		"api_call_reward_at",
	}).AddRow(
		int64(100),
		"inviter@example.com",
		"inviter",
		"1.1.1.1",
		"2409:8962:e1:391d::1",
		int64(201),
		"invitee@example.com",
		"invitee",
		"2.2.2.2",
		"2409:8962:e1:391d::2",
		start,
		nil,
		"",
		rewardAt,
		"2409:8962:e1:391d::3",
		rewardAt,
	)

	mock.ExpectQuery("WITH candidates AS").
		WithArgs(start, end).
		WillReturnRows(rows)

	clusters, err := repo.ListAffiliateRiskClusters(context.Background(), start, end)
	require.NoError(t, err)
	require.Len(t, clusters, 1)
	require.Equal(t, int64(100), clusters[0].InviterID)
	require.Len(t, clusters[0].Invitees, 1)
	require.Equal(t, int64(201), clusters[0].Invitees[0].UserID)
	require.NotNil(t, clusters[0].Invitees[0].APICallRewardAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAffiliateRiskRepositoryUpsertFreeze(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &affiliateRepository{db: db}
	start := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	end := start.Add(12 * time.Hour)

	mock.ExpectExec("INSERT INTO affiliate_risk_freezes").
		WithArgs(int64(100), "fp", "P1", 95, "reason", start, end).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok, err := repo.UpsertAffiliateRiskFreeze(context.Background(), service.AffiliateRiskFreeze{
		InviterID:         100,
		Fingerprint:       "fp",
		Severity:          "P1",
		Score:             95,
		Reason:            "reason",
		SourceWindowStart: start,
		SourceWindowEnd:   end,
	})
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAffiliateRiskRepositoryHasActiveFreeze(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &affiliateRepository{db: db}

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	frozen, err := repo.HasActiveRiskFreeze(context.Background(), 100)
	require.NoError(t, err)
	require.True(t, frozen)
	require.NoError(t, mock.ExpectationsWereMet())
}

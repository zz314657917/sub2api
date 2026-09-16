package repository

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"strings"
	"testing"
	"time"
)

func TestPelicanClaimUsesDueFenceAndGeneration(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_, actual string) error {
		if !strings.Contains(actual, "next_run_at <= $2") || !strings.Contains(actual, "run_generation=run_generation+1") {
			return errors.New("missing fence")
		}
		return nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	cols := strings.Split(planCols, ",")
	mock.ExpectQuery(".*").WithArgs(int64(7), sqlmock.AnyArg(), true).WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(7), int64(3), "openai", "m", 15, true, 20, 100, nil, now, nil, int64(4), now, now))
	p, err := NewPelicanTestRepository(db).Claim(context.Background(), 7, true, now)
	if err != nil {
		t.Fatal(err)
	}
	if p.RunGeneration != 4 {
		t.Fatalf("generation=%d", p.RunGeneration)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPelicanEmptyAuthorizationDoesNotQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := NewPelicanTestRepository(db)
	got, err := r.ListHistory(context.Background(), service.PelicanAuthorization{}, 0, 0)
	if err != nil || len(got) != 0 {
		t.Fatalf("got=%v err=%v", got, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPelicanSaveResultFenceDoesNotPruneWhenStale(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pelican_test_plans").WithArgs(int64(1), int64(2), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	now := time.Now()
	x := &service.PelicanResult{PlanID: 1, GroupID: 2, AccountID: 3, Status: "failed", StartedAt: now, FinishedAt: &now}
	if err := NewPelicanTestRepository(db).SaveResult(context.Background(), x, 20, 9); !errors.Is(err, service.ErrPelicanConflict) {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

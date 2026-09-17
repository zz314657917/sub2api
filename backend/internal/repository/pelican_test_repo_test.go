package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"strings"
	"testing"
	"time"
)

func TestPelicanClaimUsesDueFenceAndGeneration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	cols := strings.Split(planCols, ",")
	row := []driver.Value{int64(7), int64(3), "openai", "m", 15, true, 20, 100, nil, now, nil, int64(3), 0, 0, 0, nil, 0, 0, "", 0, "", now, now}
	updatedRow := append([]driver.Value(nil), row...)
	updatedRow[11] = int64(4)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT " + planCols).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows(cols).AddRow(row...))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"exhausted"}).AddRow(false))
	mock.ExpectQuery("UPDATE pelican_test_plans SET running_until").WithArgs(int64(7), now).WillReturnRows(sqlmock.NewRows(cols).AddRow(updatedRow...))
	mock.ExpectCommit()
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

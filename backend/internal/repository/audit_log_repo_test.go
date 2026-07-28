package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestAuditLogRepositoryClearAllCommitsTraceAtomically(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM audit_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))
	mock.ExpectExec(`TRUNCATE TABLE audit_logs`).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnResult(sqlmock.NewResult(4, 1))
	mock.ExpectCommit()

	trace := &service.AuditLog{Action: service.AuditActionAuditLogClear}
	deleted, err := (&auditLogRepository{db: db}).ClearAll(context.Background(), trace)
	if err != nil {
		t.Fatalf("ClearAll() error = %v", err)
	}
	if deleted != 3 {
		t.Fatalf("ClearAll() deleted = %d, want 3", deleted)
	}
	if got := trace.Extra["deleted_rows"]; got != int64(3) {
		t.Fatalf("trace deleted_rows = %#v, want int64(3)", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestAuditLogRepositoryClearAllRollsBackWhenTraceInsertFails(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	traceErr := errors.New("trace insert failed")
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM audit_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))
	mock.ExpectExec(`TRUNCATE TABLE audit_logs`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnError(traceErr)
	mock.ExpectRollback()

	deleted, err := (&auditLogRepository{db: db}).ClearAll(context.Background(), &service.AuditLog{})
	if err == nil || !errors.Is(err, traceErr) {
		t.Fatalf("ClearAll() error = %v, want trace insert failure", err)
	}
	if deleted != 0 {
		t.Fatalf("ClearAll() deleted = %d, want 0 after rollback", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

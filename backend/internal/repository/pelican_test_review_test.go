package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestPelicanReviewSaveResultCommitsFenceInsertAndRetention(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pelican_test_plans").WithArgs(int64(1), int64(2), int64(3)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec("INSERT INTO pelican_test_results").WithArgs(
		int64(1), int64(2), int64(4), "gpt-test", "pelican-v1", "success", "", int64(8), 120, 100, int64(3), now, now, "<html>x</html>", "high",
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE pelican_test_plans SET round_successes").WithArgs(int64(1), int64(3)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM pelican_test_results").WithArgs(int64(1), int64(2), 20).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	result := &service.PelicanResult{PlanID: 1, GroupID: 2, AccountID: 4, ModelID: "gpt-test", PromptVersion: "pelican-v1", Status: "success", LatencyMS: 8, CharCount: 120, MinChars: 100, StartedAt: now, FinishedAt: &now, HTML: "<html>x</html>"}
	effort := "high"
	result.ReasoningEffort = &effort
	if err := NewPelicanTestRepository(db).SaveResult(context.Background(), result, 20, 3); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPelicanReviewResultDetailUsesAllowedGroups(t *testing.T) {
	matcher := sqlmock.QueryMatcherFunc(func(_ string, actual string) error {
		for _, required := range []string{"p.group_id=r.group_id", "g.deleted_at IS NULL", "g.status='active'", "r.group_id=ANY($2)"} {
			if !strings.Contains(actual, required) {
				return &pelicanSQLRequirementError{required: required}
			}
		}
		return nil
	})
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	columns := []string{"id", "plan_id", "group_id", "account_id", "model_id", "prompt_version", "status", "error_message", "latency_ms", "char_count", "min_chars", "started_at", "finished_at", "html", "reasoning_effort"}
	mock.ExpectQuery(".*").WithArgs(int64(88), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(88), int64(1), int64(7), int64(3), "gpt-test", "pelican-v1", "success", "", int64(1), 100, 100, now, now, "<html>x</html>", nil))
	result, err := NewPelicanTestRepository(db).GetResult(context.Background(), service.PelicanAuthorization{GroupIDs: []int64{7}}, 88)
	if err != nil || result == nil || result.ID != 88 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPelicanReviewCanRunAccountUsesLiveEligibilityAndFence(t *testing.T) {
	matcher := sqlmock.QueryMatcherFunc(func(_ string, actual string) error {
		for _, required := range []string{
			"pelican_test_plans", "JOIN groups g", "g.deleted_at IS NULL", "g.status='active'", "g.platform='openai'",
			"JOIN account_groups", "JOIN accounts a", "a.deleted_at IS NULL", "a.status='active'",
			"p.run_generation=$2", "p.running_until>NOW()",
		} {
			if !strings.Contains(actual, required) {
				return &pelicanSQLRequirementError{required: required}
			}
		}
		return nil
	})
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(".*").WithArgs(int64(11), int64(12), int64(13)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	allowed, err := NewPelicanTestRepository(db).CanRunAccount(context.Background(), 11, 12, 13)
	if err != nil || allowed {
		t.Fatalf("allowed=%v err=%v, want false and nil", allowed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type pelicanSQLRequirementError struct{ required string }

func (e *pelicanSQLRequirementError) Error() string { return "missing SQL requirement: " + e.required }

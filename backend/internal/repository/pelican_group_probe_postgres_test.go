package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestPelicanGroupProbePostgres(t *testing.T) {
	dsn := os.Getenv("PELICAN_COST_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	schema := fmt.Sprintf("pelican_group_%d", time.Now().UnixNano())
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("CREATE SCHEMA " + schema)
	defer db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE")
	exec("SET search_path TO " + schema)
	for _, name := range []string{"243_pelican_tests.sql", "244_pelican_cost_cleanup.sql", "245_pelican_reasoning_effort.sql", "247_pelican_result_reasoning_effort.sql", "248_pelican_plan_timeout.sql"} {
		body, err := os.ReadFile("../../migrations/" + name)
		if err != nil {
			t.Fatal(err)
		}
		exec(string(body))
	}
	exec("CREATE TABLE groups (id BIGINT PRIMARY KEY,name TEXT,status TEXT,deleted_at TIMESTAMPTZ)")
	exec("INSERT INTO groups VALUES (7,'group seven','active',NULL),(8,'private group','active',NULL)")
	r := NewPelicanTestRepository(db)
	p, err := r.CreatePlan(ctx, service.PelicanPlanInput{GroupID: 7, ModelID: "gpt-test", IntervalMinutes: 15, MaxResults: 2, MinChars: 100}, "group seven")
	if err != nil {
		t.Fatal(err)
	}
	p, err = r.Claim(ctx, p.ID, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for i, account := range []int64{11, 22, 33, 0} {
		status := "failed"
		if i == 0 {
			status = "success"
		}
		if account == 0 {
			status = "skipped"
		}
		when := time.Now().Add(time.Duration(i-10) * time.Minute)
		err = r.SaveResult(ctx, &service.PelicanResult{PlanID: p.ID, GroupID: 7, AccountID: account, ModelID: "gpt-test", PromptVersion: "v1", Status: status, StartedAt: when, FinishedAt: &when, MinChars: 100}, 2, p.RunGeneration)
		if err != nil {
			t.Fatal(err)
		}
	}
	auth := service.PelicanAuthorization{GroupIDs: []int64{7}}
	gallery, err := r.ListGallery(ctx, auth, 0, 0, 1, 24)
	if err != nil {
		t.Fatal(err)
	}
	if gallery.Total != 1 || len(gallery.Items) != 1 || gallery.Items[0].AccountID != 0 || gallery.Items[0].HistoryCount != 3 || gallery.Items[0].ArtworkResultID == nil {
		t.Fatalf("gallery=%+v", gallery)
	}
	history, err := r.ListHistory(ctx, auth, p.ID, 0)
	if err != nil || len(history) != 3 {
		t.Fatalf("history=%v err=%v", history, err)
	}
	for _, item := range history {
		if item.AccountID == 22 {
			t.Fatal("old cross-account failure was not pruned")
		}
	}
	if _, err = r.GetResult(ctx, auth, gallery.Items[0].ResultID); err != nil {
		t.Fatal(err)
	}
	denied := service.PelicanAuthorization{GroupIDs: []int64{8}}
	other, err := r.ListGallery(ctx, denied, 0, 0, 1, 24)
	if err != nil || other.Total != 0 {
		t.Fatalf("unauthorized gallery=%v err=%v", other, err)
	}
	if _, err = r.GetResult(ctx, denied, gallery.Items[0].ResultID); err != service.ErrPelicanNotFound {
		t.Fatalf("unauthorized detail=%v", err)
	}
	if err = r.Finalize(ctx, p.ID, p.RunGeneration); err != nil {
		t.Fatal(err)
	}
	exec("UPDATE pelican_test_plans SET max_results=1 WHERE id=$1", p.ID)
	if _, err = r.Cleanup(ctx, p.ID, "expired"); err != nil {
		t.Fatal(err)
	}
	history, err = r.ListHistory(ctx, auth, p.ID, 0)
	if err != nil || len(history) != 2 {
		t.Fatalf("cleaned history=%v err=%v", history, err)
	}
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	_ "github.com/lib/pq"
)

// TestPelicanEightPlatformPostgres is deliberately not optional once the
// task-owned DSN is supplied by the PostgreSQL QA runner.
func TestPelicanEightPlatformPostgres(t *testing.T) {
	dsn := os.Getenv("PELICAN_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PELICAN_TEST_POSTGRES_DSN not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	schema := fmt.Sprintf("pelican_platforms_%d", time.Now().UnixNano())
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("CREATE SCHEMA " + schema)
	defer db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE")
	exec("SET search_path TO " + schema)
	for _, name := range []string{"243_pelican_tests.sql", "244_pelican_cost_cleanup.sql", "245_pelican_reasoning_effort.sql", "247_pelican_result_reasoning_effort.sql", "248_pelican_plan_timeout.sql", "249_pelican_result_error_details.sql"} {
		body, err := os.ReadFile("../../migrations/" + name)
		if err != nil {
			t.Fatal(err)
		}
		exec(string(body))
	}
	exec("CREATE TABLE groups (id BIGINT PRIMARY KEY,name TEXT,platform TEXT NOT NULL,status TEXT,deleted_at TIMESTAMPTZ)")
	exec("CREATE TABLE accounts (id BIGINT PRIMARY KEY,status TEXT,deleted_at TIMESTAMPTZ)")
	exec("CREATE TABLE account_groups (group_id BIGINT,account_id BIGINT)")

	platforms := []string{"openai", "anthropic", "gemini", "grok", "antigravity", "kimi", "zhipu", "deepseek"}
	repo := NewPelicanTestRepository(db)
	type platformRun struct {
		platform  string
		groupID   int64
		accountID int64
		plan      *service.PelicanPlan
		resultID  int64
	}
	runs := make([]platformRun, 0, len(platforms))
	for index, platform := range platforms {
		groupID := int64(index + 1)
		accountID := int64(index + 101)
		exec("INSERT INTO groups VALUES ($1,$2,$3,'active',NULL)", groupID, platform+" group", platform)
		exec("INSERT INTO accounts VALUES ($1,'active',NULL)", accountID)
		exec("INSERT INTO account_groups VALUES ($1,$2)", groupID, accountID)
		plan, err := repo.CreatePlan(ctx, service.PelicanPlanInput{GroupID: groupID, ModelID: platform + "-model", IntervalMinutes: 15, Enabled: true, MaxResults: 2, MinChars: 100}, platform+" group")
		if err != nil {
			t.Fatalf("create %s plan: %v", platform, err)
		}
		plan, err = repo.Claim(ctx, plan.ID, true, time.Now())
		if err != nil {
			t.Fatalf("claim %s plan: %v", platform, err)
		}
		allowed, err := repo.CanRunAccount(ctx, plan.ID, plan.RunGeneration, accountID)
		if err != nil || !allowed {
			t.Fatalf("CanRunAccount %s allowed=%v err=%v", platform, allowed, err)
		}
		finished := time.Now()
		if err = repo.SaveResult(ctx, &service.PelicanResult{PlanID: plan.ID, GroupID: groupID, AccountID: accountID, ModelID: platform + "-model", PromptVersion: "v1", Status: "success", MinChars: 100, StartedAt: finished, FinishedAt: &finished, HTML: "<html><body>ok</body></html>"}, 2, plan.RunGeneration); err != nil {
			t.Fatalf("save %s result: %v", platform, err)
		}
		var resultID int64
		if err = db.QueryRowContext(ctx, "SELECT id FROM pelican_test_results WHERE plan_id=$1 AND group_id=$2 AND account_id=$3", plan.ID, groupID, accountID).Scan(&resultID); err != nil {
			t.Fatalf("lookup %s result: %v", platform, err)
		}
		if err = repo.Finalize(ctx, plan.ID, plan.RunGeneration); err != nil {
			t.Fatalf("finalize %s plan: %v", platform, err)
		}
		runs = append(runs, platformRun{platform: platform, groupID: groupID, accountID: accountID, plan: plan, resultID: resultID})
	}

	gallery, err := repo.ListGallery(ctx, service.PelicanAuthorization{Admin: true}, 0, 0, 1, 24)
	if err != nil {
		t.Fatalf("admin gallery: %v", err)
	}
	if gallery.Total != len(platforms) || len(gallery.Items) != len(platforms) || len(gallery.Groups) != len(platforms) {
		t.Fatalf("admin gallery total=%d items=%d groups=%d err=%v", gallery.Total, len(gallery.Items), len(gallery.Groups), err)
	}
	itemPlatforms := make(map[string]bool, len(platforms))
	for _, item := range gallery.Items {
		itemPlatforms[item.Platform] = true
	}
	groupPlatforms := make(map[string]bool, len(platforms))
	for _, group := range gallery.Groups {
		if group.Platform == "" {
			t.Fatal("gallery group missing platform")
		}
		groupPlatforms[group.Platform] = true
	}
	for _, platform := range platforms {
		if !itemPlatforms[platform] {
			t.Fatalf("gallery entry missing %s", platform)
		}
		if !groupPlatforms[platform] {
			t.Fatalf("gallery group missing %s", platform)
		}
	}

	ordinary, err := repo.ListGallery(ctx, service.PelicanAuthorization{GroupIDs: []int64{2}}, 0, 0, 1, 24)
	if err != nil || ordinary.Total != 1 || len(ordinary.Items) != 1 || ordinary.Items[0].Platform != "anthropic" || len(ordinary.Groups) != 1 || ordinary.Groups[0].Platform != "anthropic" {
		t.Fatalf("ordinary gallery=%+v err=%v", ordinary, err)
	}
	denied, err := repo.ListGallery(ctx, service.PelicanAuthorization{GroupIDs: []int64{999}}, 0, 0, 1, 24)
	if err != nil || denied.Total != 0 || len(denied.Items) != 0 {
		t.Fatalf("unauthorized gallery=%+v err=%v", denied, err)
	}
	for index, run := range runs {
		allowedAuth := service.PelicanAuthorization{GroupIDs: []int64{run.groupID}}
		history, err := repo.ListHistory(ctx, allowedAuth, run.plan.ID, run.accountID)
		if err != nil || len(history) != 1 || history[0].ID != run.resultID || history[0].GroupID != run.groupID {
			t.Fatalf("%s authorized history=%+v err=%v", run.platform, history, err)
		}
		result, err := repo.GetResult(ctx, allowedAuth, run.resultID)
		if err != nil || result.ID != run.resultID || result.GroupID != run.groupID {
			t.Fatalf("%s authorized result=%+v err=%v", run.platform, result, err)
		}
		other := runs[(index+1)%len(runs)]
		deniedAuth := service.PelicanAuthorization{GroupIDs: []int64{other.groupID}}
		deniedHistory, err := repo.ListHistory(ctx, deniedAuth, run.plan.ID, run.accountID)
		if err != nil || len(deniedHistory) != 0 {
			t.Fatalf("%s cross-group history=%+v err=%v", run.platform, deniedHistory, err)
		}
		if _, err = repo.GetResult(ctx, deniedAuth, run.resultID); !errors.Is(err, service.ErrPelicanNotFound) {
			t.Fatalf("%s cross-group result err=%v", run.platform, err)
		}
	}

	// Cleanup must remove expired failures while retaining the latest success.
	plan := runs[0].plan
	exec("UPDATE pelican_test_plans SET retention_days=1 WHERE id=$1", plan.ID)
	exec("INSERT INTO pelican_test_results(plan_id,group_id,account_id,model_id,prompt_version,status,min_chars,started_at,finished_at) VALUES ($1,1,101,'openai-model','v1','failed',100,NOW()-interval '3 days',NOW()-interval '3 days')", plan.ID)
	if _, err = repo.Cleanup(ctx, plan.ID, "expired"); err != nil {
		t.Fatal(err)
	}
	var successes, failures int
	if err = db.QueryRowContext(ctx, "SELECT count(*) FILTER (WHERE status='success'),count(*) FILTER (WHERE status='failed') FROM pelican_test_results WHERE plan_id=$1", plan.ID).Scan(&successes, &failures); err != nil || successes != 1 || failures != 0 {
		t.Fatalf("cleanup successes=%d failures=%d err=%v", successes, failures, err)
	}
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	_ "github.com/lib/pq"
)

// This is deliberately opt-in: the controller supplies a task-owned DSN and
// this test creates a private schema, never touching the shared public schema.
func TestPelicanCostCleanupPostgres(t *testing.T) {
	dsn := os.Getenv("PELICAN_COST_TEST_DSN")
	if dsn == "" {
		t.Skip("PELICAN_COST_TEST_DSN not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	schema := fmt.Sprintf("pelican_cost_%d", time.Now().UnixNano())
	defer func() { _, _ = db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	if _, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"../../migrations/243_pelican_tests.sql", "../../migrations/244_pelican_cost_cleanup.sql", "../../migrations/245_pelican_reasoning_effort.sql"} {
		body, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, err = db.ExecContext(ctx, string(body)); err != nil {
			t.Fatalf("apply %s: %v", file, err)
		}
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,daily_call_limit,failure_pause_threshold,retention_days,next_run_at) VALUES (1,1,'g','m',15,TRUE,1,100,2,2,1,NOW()-interval '1 minute')`); err != nil {
		t.Fatal(err)
	}
	repo := NewPelicanTestRepository(db)
	// Migration 245 must round-trip the administrator-selected value through
	// the normal repository create/update/list flow.
	if _, err = db.ExecContext(ctx, "SELECT setval(pg_get_serial_sequence('pelican_test_plans','id'), 100, false)"); err != nil {
		t.Fatal(err)
	}
	effortPlan, err := repo.CreatePlan(ctx, service.PelicanPlanInput{GroupID: 17, ModelID: "gpt-test", IntervalMinutes: 15, MaxResults: 1, MinChars: 100, ReasoningEffort: "high"}, "effort-group")
	if err != nil || effortPlan.ReasoningEffort != "high" {
		t.Fatalf("create effort plan=%+v err=%v", effortPlan, err)
	}
	effortPlan, err = repo.UpdatePlan(ctx, effortPlan.ID, service.PelicanPlanInput{GroupID: 17, ModelID: "gpt-test", IntervalMinutes: 15, Enabled: true, MaxResults: 1, MinChars: 100, ReasoningEffort: "none"}, "effort-group")
	if err != nil || effortPlan.ReasoningEffort != "none" {
		t.Fatalf("update effort plan=%+v err=%v", effortPlan, err)
	}
	plans, err := repo.ListPlans(ctx)
	foundEffortPlan := false
	for _, plan := range plans {
		if plan.ID == effortPlan.ID && plan.ReasoningEffort == "none" {
			foundEffortPlan = true
		}
	}
	if err != nil || !foundEffortPlan {
		t.Fatalf("list effort plans=%+v err=%v", plans, err)
	}
	p, err := repo.Claim(ctx, 1, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err = repo.ReserveAttempt(ctx, p.ID, p.RunGeneration); err != nil {
			t.Fatalf("reserve %d: %v", i, err)
		}
	}
	if err = repo.ReserveAttempt(ctx, p.ID, p.RunGeneration); err != service.ErrPelicanQuota {
		t.Fatalf("quota err=%v", err)
	}
	if err = repo.Finalize(ctx, p.ID, p.RunGeneration); err != nil {
		t.Fatal(err)
	}
	// Manual and scheduled claims both fail before work when today's allowance is spent.
	if _, err = repo.Claim(ctx, 1, true, time.Now()); err != service.ErrPelicanQuota {
		t.Fatalf("manual quota err=%v", err)
	}
	// Save an old successful row and an old failed row; expiry preserves the success.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_results(plan_id,group_id,account_id,model_id,prompt_version,status,min_chars,started_at) VALUES (1,1,1,'m','v','success',100,NOW()-interval '5 days'),(1,1,1,'m','v','failed',100,NOW()-interval '5 days')`); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Cleanup(ctx, 1, "expired"); err != nil {
		t.Fatal(err)
	}
	var success, failed int
	if err = db.QueryRowContext(ctx, "SELECT count(*) FILTER (WHERE status='success'),count(*) FILTER (WHERE status='failed') FROM pelican_test_results").Scan(&success, &failed); err != nil {
		t.Fatal(err)
	}
	if success != 1 || failed != 0 {
		t.Fatalf("retention success=%d failed=%d", success, failed)
	}
	var usedBefore, usedAfter int
	if err = db.QueryRowContext(ctx, "SELECT daily_calls_used FROM pelican_test_plans WHERE id=1").Scan(&usedBefore); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Cleanup(ctx, 1, "failed"); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, "SELECT daily_calls_used FROM pelican_test_plans WHERE id=1").Scan(&usedAfter); err != nil {
		t.Fatal(err)
	}
	if usedBefore != usedAfter {
		t.Fatalf("cleanup changed quota: %d -> %d", usedBefore, usedAfter)
	}
	// A failed attempted round pauses exactly once, then requires explicit resume.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,failure_pause_threshold,next_run_at) VALUES (3,3,'g3','m',15,TRUE,1,100,1,NOW())`); err != nil {
		t.Fatal(err)
	}
	pausePlan, err := repo.Claim(ctx, 3, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.ReserveAttempt(ctx, pausePlan.ID, pausePlan.RunGeneration); err != nil {
		t.Fatal(err)
	}
	if err = repo.Finalize(ctx, pausePlan.ID, pausePlan.RunGeneration); err != nil {
		t.Fatal(err)
	}
	if err = repo.Finalize(ctx, pausePlan.ID, pausePlan.RunGeneration); err != nil {
		t.Fatal(err)
	}
	var enabled bool
	var pauseReason string
	var streak int
	if err = db.QueryRowContext(ctx, "SELECT enabled,pause_reason,consecutive_failed_runs FROM pelican_test_plans WHERE id=3").Scan(&enabled, &pauseReason, &streak); err != nil {
		t.Fatal(err)
	}
	if enabled || pauseReason != "consecutive_failures" || streak != 1 {
		t.Fatalf("pause state enabled=%v reason=%q streak=%d", enabled, pauseReason, streak)
	}
	if _, err = repo.Claim(ctx, 3, true, time.Now()); err != service.ErrPelicanPaused {
		t.Fatalf("paused manual err=%v", err)
	}
	if _, err = db.ExecContext(ctx, "UPDATE pelican_test_plans SET usage_day=(NOW() AT TIME ZONE 'Asia/Shanghai')::date,daily_calls_used=7 WHERE id=3"); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Resume(ctx, 3); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, "SELECT daily_calls_used FROM pelican_test_plans WHERE id=3").Scan(&usedAfter); err != nil || usedAfter != 7 {
		t.Fatalf("resume changed usage=%d err=%v", usedAfter, err)
	}
	// A stale generation cannot finalize the new lease.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,next_run_at) VALUES (5,5,'g5','m',15,TRUE,1,100,NOW())`); err != nil {
		t.Fatal(err)
	}
	stalePlan, err := repo.Claim(ctx, 5, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Finalize(ctx, stalePlan.ID, stalePlan.RunGeneration+1); err != nil {
		t.Fatal(err)
	}
	var leaseLive bool
	if err = db.QueryRowContext(ctx, "SELECT running_until>NOW() FROM pelican_test_plans WHERE id=5").Scan(&leaseLive); err != nil || !leaseLive {
		t.Fatalf("stale finalize altered lease err=%v", err)
	}
	// A stored previous-day counter resets under database UTC+8 time when reserved.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,daily_call_limit,usage_day,daily_calls_used,next_run_at) VALUES (4,4,'g4','m',15,TRUE,1,100,1,(NOW() AT TIME ZONE 'Asia/Shanghai')::date-1,9,NOW())`); err != nil {
		t.Fatal(err)
	}
	dayPlan, err := repo.Claim(ctx, 4, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.ReserveAttempt(ctx, dayPlan.ID, dayPlan.RunGeneration); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, "SELECT daily_calls_used FROM pelican_test_plans WHERE id=4").Scan(&usedAfter); err != nil || usedAfter != 1 {
		t.Fatalf("midnight used=%d err=%v", usedAfter, err)
	}
	// A second pool proves the reservation predicate is enforced by PostgreSQL,
	// rather than by this process's single connection.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,daily_call_limit,next_run_at) VALUES (2,2,'g2','m',15,TRUE,1,100,1,NOW())`); err != nil {
		t.Fatal(err)
	}
	p2, err := repo.Claim(ctx, 2, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	parallelDB, err := sql.Open("postgres", dsn+"&search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	defer parallelDB.Close()
	parallelDB.SetMaxOpenConns(8)
	parallelRepo := NewPelicanTestRepository(parallelDB)
	var wg sync.WaitGroup
	successes := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if parallelRepo.ReserveAttempt(ctx, p2.ID, p2.RunGeneration) == nil {
				successes <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(successes)
	count := 0
	for range successes {
		count++
	}
	if count != 1 {
		t.Fatalf("concurrent reservations=%d, want 1", count)
	}
	// A real stale generation must not clear a newer claim's lease.
	if _, err = db.ExecContext(ctx, `UPDATE pelican_test_plans SET running_until=NOW()-interval '1 minute' WHERE id=5`); err != nil {
		t.Fatal(err)
	}
	newerPlan, err := repo.Claim(ctx, 5, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if newerPlan.RunGeneration <= stalePlan.RunGeneration {
		t.Fatalf("generation did not advance: stale=%d newer=%d", stalePlan.RunGeneration, newerPlan.RunGeneration)
	}
	if err = repo.Finalize(ctx, 5, stalePlan.RunGeneration); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, "SELECT running_until>NOW() AND run_generation=$1 FROM pelican_test_plans WHERE id=5", newerPlan.RunGeneration).Scan(&leaseLive); err != nil || !leaseLive {
		t.Fatalf("old generation finalized newer lease err=%v", err)
	}

	// Reclaiming an abandoned attempted failure must finalize it before accepting
	// another run, so a threshold pause cannot be bypassed by a process crash.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,failure_pause_threshold,next_run_at) VALUES (6,6,'g6','m',15,TRUE,1,100,1,NOW())`); err != nil {
		t.Fatal(err)
	}
	abandoned, err := repo.Claim(ctx, 6, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.ReserveAttempt(ctx, 6, abandoned.RunGeneration); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, "UPDATE pelican_test_plans SET running_until=NOW()-interval '1 minute' WHERE id=6"); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Claim(ctx, 6, true, time.Now()); err != service.ErrPelicanPaused {
		t.Fatalf("abandoned failed round claim err=%v, want paused", err)
	}

	// A success clears the streak, while an all-skipped round (zero attempts)
	// leaves it unchanged.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,next_run_at,consecutive_failed_runs,round_attempted,round_successes,running_until,run_generation) VALUES (7,7,'g7','m',15,TRUE,1,100,NOW(),4,1,1,NOW()+interval '10 minutes',1),(8,8,'g8','m',15,TRUE,1,100,NOW(),4,0,0,NOW()+interval '10 minutes',1)`); err != nil {
		t.Fatal(err)
	}
	if err = repo.Finalize(ctx, 7, 1); err != nil {
		t.Fatal(err)
	}
	if err = repo.Finalize(ctx, 8, 1); err != nil {
		t.Fatal(err)
	}
	var successStreak, skippedStreak int
	if err = db.QueryRowContext(ctx, "SELECT consecutive_failed_runs FROM pelican_test_plans WHERE id=7").Scan(&successStreak); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRowContext(ctx, "SELECT consecutive_failed_runs FROM pelican_test_plans WHERE id=8").Scan(&skippedStreak); err != nil {
		t.Fatal(err)
	}
	if successStreak != 0 || skippedStreak != 4 {
		t.Fatalf("round state success=%d skipped=%d", successStreak, skippedStreak)
	}

	// Due cleanup includes disabled plans and protects the latest success for
	// every plan/account/group partition, even when started_at ordering differs.
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_plans (id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,retention_days,next_run_at) VALUES (9,9,'g9','m',15,FALSE,1,100,1,NOW())`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO pelican_test_results(plan_id,group_id,account_id,model_id,prompt_version,status,min_chars,started_at) VALUES
	 (9,9,91,'m','v','success',100,NOW()-interval '9 days'),
	 (9,9,91,'m','v','failed',100,NOW()-interval '8 days'),
	 (9,10,91,'m','v','success',100,NOW()-interval '7 days'),
	 (9,10,91,'m','v','skipped',100,NOW()-interval '6 days')`); err != nil {
		t.Fatal(err)
	}
	if err = repo.CleanupDue(ctx); err != nil {
		t.Fatal(err)
	}
	var retainedSuccess, retainedOther int
	if err = db.QueryRowContext(ctx, "SELECT count(*) FILTER (WHERE status='success'),count(*) FILTER (WHERE status<>'success') FROM pelican_test_results WHERE plan_id=9").Scan(&retainedSuccess, &retainedOther); err != nil {
		t.Fatal(err)
	}
	if retainedSuccess != 2 || retainedOther != 0 {
		t.Fatalf("disabled cleanup retained success=%d other=%d", retainedSuccess, retainedOther)
	}
}

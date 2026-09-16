package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestPelicanConcurrentClaim(t *testing.T) {
	dsn := os.Getenv("PELICAN_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PELICAN_TEST_POSTGRES_DSN not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	repo := NewPelicanTestRepository(db)
	ctx := context.Background()

	// Create a plan
	input := service.PelicanPlanInput{
		GroupID:         500,
		ModelID:         "gpt-4",
		IntervalMinutes: 60,
		Enabled:         true,
		MaxResults:      20,
		MinChars:        9366,
	}

	plan, err := repo.CreatePlan(ctx, input, "concurrency-test")
	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}
	defer repo.DeletePlan(ctx, plan.ID)

	// Launch 10 concurrent claim attempts
	const workers = 10
	var wg sync.WaitGroup
	successChan := make(chan int, workers)
	conflictChan := make(chan int, workers)

	now := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			claimed, err := repo.Claim(ctx, plan.ID, true, now)
			if err == service.ErrPelicanConflict {
				conflictChan <- workerID
			} else if err != nil {
				t.Errorf("worker %d unexpected error: %v", workerID, err)
			} else if claimed != nil {
				successChan <- workerID
			}
		}(i)
	}

	wg.Wait()
	close(successChan)
	close(conflictChan)

	successCount := 0
	for range successChan {
		successCount++
	}

	conflictCount := 0
	for range conflictChan {
		conflictCount++
	}

	t.Logf("Success: %d, Conflicts: %d", successCount, conflictCount)

	// Exactly one should succeed
	if successCount != 1 {
		t.Errorf("expected exactly 1 success, got %d", successCount)
	}

	// The rest should conflict
	if conflictCount != workers-1 {
		t.Errorf("expected %d conflicts, got %d", workers-1, conflictCount)
	}
}

func TestPelicanRowLockBehavior(t *testing.T) {
	dsn := os.Getenv("PELICAN_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PELICAN_TEST_POSTGRES_DSN not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	repo := NewPelicanTestRepository(db)
	ctx := context.Background()

	// Create a plan
	input := service.PelicanPlanInput{
		GroupID:         600,
		ModelID:         "gpt-4",
		IntervalMinutes: 60,
		Enabled:         true,
		MaxResults:      20,
		MinChars:        9366,
	}

	plan, err := repo.CreatePlan(ctx, input, "lock-test")
	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}
	defer repo.DeletePlan(ctx, plan.ID)

	// Claim the plan
	now := time.Now()
	claimed, err := repo.Claim(ctx, plan.ID, true, now)
	if err != nil {
		t.Fatalf("Claim failed: %v", err)
	}

	// Verify running_until is set
	if claimed.RunningUntil == nil {
		t.Fatal("expected RunningUntil to be set")
	}

	// Try to update while running - should conflict
	updateInput := service.PelicanPlanInput{
		GroupID:         600,
		ModelID:         "gpt-4o",
		IntervalMinutes: 120,
		Enabled:         true,
		MaxResults:      25,
		MinChars:        10000,
	}

	_, err = repo.UpdatePlan(ctx, plan.ID, updateInput, "lock-test-updated")
	if err != service.ErrPelicanConflict {
		t.Errorf("expected ErrPelicanConflict during update, got %v", err)
	}

	// Try to delete while running - should conflict
	err = repo.DeletePlan(ctx, plan.ID)
	if err != service.ErrPelicanConflict {
		t.Errorf("expected ErrPelicanConflict during delete, got %v", err)
	}

	// Release the lock
	err = repo.Release(ctx, plan.ID, claimed.RunGeneration)
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	// Now update should succeed
	updated, err := repo.UpdatePlan(ctx, plan.ID, updateInput, "lock-test-updated")
	if err != nil {
		t.Errorf("UpdatePlan after release failed: %v", err)
	}
	if updated.ModelID != "gpt-4o" {
		t.Errorf("expected ModelID=gpt-4o, got %s", updated.ModelID)
	}
}

func TestPelicanAuthorizationQueries(t *testing.T) {
	dsn := os.Getenv("PELICAN_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PELICAN_TEST_POSTGRES_DSN not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	repo := NewPelicanTestRepository(db)
	ctx := context.Background()

	// Create test plans in different groups
	groups := []int64{700, 701, 702}
	planIDs := make([]int64, len(groups))

	for i, gid := range groups {
		input := service.PelicanPlanInput{
			GroupID:         gid,
			ModelID:         "gpt-4",
			IntervalMinutes: 60,
			Enabled:         false,
			MaxResults:      20,
			MinChars:        9366,
		}
		plan, err := repo.CreatePlan(ctx, input, fmt.Sprintf("auth-test-%d", gid))
		if err != nil {
			t.Fatalf("CreatePlan for group %d failed: %v", gid, err)
		}
		planIDs[i] = plan.ID
		defer repo.DeletePlan(ctx, plan.ID)
	}

	// Test admin authorization (sees all)
	_ = service.PelicanAuthorization{Admin: true} // Would be used with ListGallery
	adminPlans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans admin failed: %v", err)
	}
	adminCount := 0
	for _, p := range adminPlans {
		for _, gid := range groups {
			if p.GroupID == gid {
				adminCount++
				break
			}
		}
	}
	if adminCount != len(groups) {
		t.Errorf("admin expected to see %d plans, saw %d", len(groups), adminCount)
	}
	t.Logf("Admin authorization test passed (saw %d/%d plans)", adminCount, len(groups))

	// Note: Full authorization filtering is implemented in ListGallery and ListHistory
	// which require actual groups, accounts, and account_groups records.
	// This test verifies the basic plan CRUD which has no authorization layer.
	t.Log("Note: ListPlans has no auth filtering; full auth is in ListGallery/ListHistory")
}

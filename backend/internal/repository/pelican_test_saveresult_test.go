package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestPelicanSaveResultIntegration(t *testing.T) {
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

	t.Run("SaveResultBasic", func(t *testing.T) {
		// Create plan
		input := service.PelicanPlanInput{
			GroupID:         800,
			ModelID:         "gpt-4",
			IntervalMinutes: 60,
			Enabled:         true,
			MaxResults:      20,
			MinChars:        9366,
		}
		plan, err := repo.CreatePlan(ctx, input, "save-result-test")
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

		// Save a result
		finishedAt := time.Now()
		result := &service.PelicanResult{
			PlanID:         claimed.ID,
			GroupID:        claimed.GroupID,
			AccountID:      12345,
			ModelID:        "gpt-4",
			PromptVersion:  "v1",
			Status:         "success",
			ErrorMessage:   "",
			LatencyMS:      1500,
			CharCount:      10000,
			MinChars:       9366,
			StartedAt:      now,
			FinishedAt:     &finishedAt,
			HTML:           "<html><body>Test</body></html>",
		}

		err = repo.SaveResult(ctx, result, 20, claimed.RunGeneration)
		if err != nil {
			t.Fatalf("SaveResult failed: %v", err)
		}

		// Release
		err = repo.Release(ctx, claimed.ID, claimed.RunGeneration)
		if err != nil {
			t.Fatalf("Release failed: %v", err)
		}
	})

	t.Run("SaveResultGenerationConflict", func(t *testing.T) {
		// Create plan
		input := service.PelicanPlanInput{
			GroupID:         801,
			ModelID:         "gpt-4",
			IntervalMinutes: 60,
			Enabled:         true,
			MaxResults:      20,
			MinChars:        9366,
		}
		plan, err := repo.CreatePlan(ctx, input, "generation-conflict-test")
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

		// Try to save with wrong generation
		finishedAt := time.Now()
		result := &service.PelicanResult{
			PlanID:         claimed.ID,
			GroupID:        claimed.GroupID,
			AccountID:      12345,
			ModelID:        "gpt-4",
			PromptVersion:  "v1",
			Status:         "success",
			ErrorMessage:   "",
			LatencyMS:      1500,
			CharCount:      10000,
			MinChars:       9366,
			StartedAt:      now,
			FinishedAt:     &finishedAt,
			HTML:           "<html><body>Test</body></html>",
		}

		wrongGeneration := claimed.RunGeneration + 999
		err = repo.SaveResult(ctx, result, 20, wrongGeneration)
		if err != service.ErrPelicanConflict {
			t.Errorf("expected ErrPelicanConflict, got %v", err)
		}

		// Release
		err = repo.Release(ctx, claimed.ID, claimed.RunGeneration)
		if err != nil {
			t.Fatalf("Release failed: %v", err)
		}
	})

	t.Run("SaveResultPruning", func(t *testing.T) {
		// Create plan
		input := service.PelicanPlanInput{
			GroupID:         802,
			ModelID:         "gpt-4",
			IntervalMinutes: 60,
			Enabled:         true,
			MaxResults:      5, // Keep only 5 results
			MinChars:        9366,
		}
		plan, err := repo.CreatePlan(ctx, input, "pruning-test")
		if err != nil {
			t.Fatalf("CreatePlan failed: %v", err)
		}
		defer repo.DeletePlan(ctx, plan.ID)

		accountID := int64(99999)

		// Save 8 results
		for i := 0; i < 8; i++ {
			now := time.Now()
			claimed, err := repo.Claim(ctx, plan.ID, true, now)
			if err != nil {
				t.Fatalf("Claim %d failed: %v", i, err)
			}

			finishedAt := time.Now()
			result := &service.PelicanResult{
				PlanID:         claimed.ID,
				GroupID:        claimed.GroupID,
				AccountID:      accountID,
				ModelID:        "gpt-4",
				PromptVersion:  "v1",
				Status:         "success",
				ErrorMessage:   "",
				LatencyMS:      1500,
				CharCount:      10000,
				MinChars:       9366,
				StartedAt:      now,
				FinishedAt:     &finishedAt,
				HTML:           "<html><body>Test</body></html>",
			}

			err = repo.SaveResult(ctx, result, input.MaxResults, claimed.RunGeneration)
			if err != nil {
				t.Fatalf("SaveResult %d failed: %v", i, err)
			}

			err = repo.Release(ctx, claimed.ID, claimed.RunGeneration)
			if err != nil {
				t.Fatalf("Release %d failed: %v", i, err)
			}

			time.Sleep(10 * time.Millisecond) // Ensure distinct timestamps
		}

		// Count results for this plan+account
		var count int
		err = db.QueryRowContext(ctx, "SELECT count(*) FROM pelican_test_results WHERE plan_id=$1 AND account_id=$2", plan.ID, accountID).Scan(&count)
		if err != nil {
			t.Fatalf("count query failed: %v", err)
		}

		if count > input.MaxResults {
			t.Errorf("expected at most %d results after pruning, got %d", input.MaxResults, count)
		}
		t.Logf("Pruning test: %d results retained (max %d)", count, input.MaxResults)
	})
}

func TestPelicanCanRunAccountIntegration(t *testing.T) {
	dsn := os.Getenv("PELICAN_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PELICAN_TEST_POSTGRES_DSN not set")
	}

	t.Run("CanRunAccountSyntax", func(t *testing.T) {
		t.Skip("CanRunAccount requires groups/accounts/account_groups schema - tested in full integration environment")
	})
}

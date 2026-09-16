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

func TestPelicanTestRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("PELICAN_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PELICAN_TEST_POSTGRES_DSN not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping: %v", err)
	}

	repo := NewPelicanTestRepository(db)
	ctx := context.Background()

	t.Run("CreateAndList", func(t *testing.T) {
		input := service.PelicanPlanInput{
			GroupID:         100,
			ModelID:         "gpt-4",
			IntervalMinutes: 60,
			Enabled:         false,
			MaxResults:      20,
			MinChars:        9366,
		}

		plan, err := repo.CreatePlan(ctx, input, "test-group")
		if err != nil {
			t.Fatalf("CreatePlan failed: %v", err)
		}
		if plan.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if plan.GroupID != 100 {
			t.Errorf("expected GroupID=100, got %d", plan.GroupID)
		}
		if plan.GroupName != "test-group" {
			t.Errorf("expected GroupName=test-group, got %s", plan.GroupName)
		}

		plans, err := repo.ListPlans(ctx)
		if err != nil {
			t.Fatalf("ListPlans failed: %v", err)
		}
		if len(plans) == 0 {
			t.Error("expected at least one plan")
		}

		// Cleanup
		_ = repo.DeletePlan(ctx, plan.ID)
	})

	t.Run("UpdatePlan", func(t *testing.T) {
		input := service.PelicanPlanInput{
			GroupID:         200,
			ModelID:         "gpt-4-turbo",
			IntervalMinutes: 120,
			Enabled:         false,
			MaxResults:      15,
			MinChars:        5000,
		}

		plan, err := repo.CreatePlan(ctx, input, "update-test")
		if err != nil {
			t.Fatalf("CreatePlan failed: %v", err)
		}
		defer repo.DeletePlan(ctx, plan.ID)

		updateInput := service.PelicanPlanInput{
			GroupID:         200,
			ModelID:         "gpt-4o",
			IntervalMinutes: 180,
			Enabled:         true,
			MaxResults:      25,
			MinChars:        10000,
		}

		updated, err := repo.UpdatePlan(ctx, plan.ID, updateInput, "updated-group")
		if err != nil {
			t.Fatalf("UpdatePlan failed: %v", err)
		}
		if updated.ModelID != "gpt-4o" {
			t.Errorf("expected ModelID=gpt-4o, got %s", updated.ModelID)
		}
		if !updated.Enabled {
			t.Error("expected Enabled=true")
		}
		if updated.GroupName != "updated-group" {
			t.Errorf("expected GroupName=updated-group, got %s", updated.GroupName)
		}
	})

	t.Run("ClaimAndRelease", func(t *testing.T) {
		input := service.PelicanPlanInput{
			GroupID:         300,
			ModelID:         "gpt-4",
			IntervalMinutes: 60,
			Enabled:         true,
			MaxResults:      20,
			MinChars:        9366,
		}

		plan, err := repo.CreatePlan(ctx, input, "claim-test")
		if err != nil {
			t.Fatalf("CreatePlan failed: %v", err)
		}
		defer repo.DeletePlan(ctx, plan.ID)

		now := time.Now()
		claimed, err := repo.Claim(ctx, plan.ID, true, now)
		if err != nil {
			t.Fatalf("Claim failed: %v", err)
		}
		if claimed.RunGeneration != plan.RunGeneration+1 {
			t.Errorf("expected RunGeneration=%d, got %d", plan.RunGeneration+1, claimed.RunGeneration)
		}
		if claimed.RunningUntil == nil || claimed.RunningUntil.Before(now) {
			t.Error("expected RunningUntil to be set in the future")
		}

		// Second claim should conflict
		_, err = repo.Claim(ctx, plan.ID, true, now)
		if err != service.ErrPelicanConflict {
			t.Errorf("expected ErrPelicanConflict, got %v", err)
		}

		// Release
		err = repo.Release(ctx, plan.ID, claimed.RunGeneration)
		if err != nil {
			t.Fatalf("Release failed: %v", err)
		}

		// Can claim again after release
		_, err = repo.Claim(ctx, plan.ID, true, time.Now())
		if err != nil {
			t.Errorf("Claim after Release failed: %v", err)
		}
	})

	t.Run("DeleteConflict", func(t *testing.T) {
		input := service.PelicanPlanInput{
			GroupID:         400,
			ModelID:         "gpt-4",
			IntervalMinutes: 60,
			Enabled:         false,
			MaxResults:      20,
			MinChars:        9366,
		}

		plan, err := repo.CreatePlan(ctx, input, "delete-test")
		if err != nil {
			t.Fatalf("CreatePlan failed: %v", err)
		}

		// Claim it
		claimed, err := repo.Claim(ctx, plan.ID, true, time.Now())
		if err != nil {
			t.Fatalf("Claim failed: %v", err)
		}

		// Delete should conflict while running
		err = repo.DeletePlan(ctx, plan.ID)
		if err != service.ErrPelicanConflict {
			t.Errorf("expected ErrPelicanConflict, got %v", err)
		}

		// Release and delete should succeed
		_ = repo.Release(ctx, plan.ID, claimed.RunGeneration)
		err = repo.DeletePlan(ctx, plan.ID)
		if err != nil {
			t.Errorf("DeletePlan after Release failed: %v", err)
		}
	})
}

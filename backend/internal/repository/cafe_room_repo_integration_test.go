//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

const cafeRoomIntegrationAttempts = 8

func TestCafeRoomRepositoryIntegration(t *testing.T) {
	t.Run("serializes operational account assignment", func(t *testing.T) {
		ctx := context.Background()
		fixture := newCafeRoomRepositoryIntegrationFixture(t)
		repo := NewCafeRoomRepository(fixture.client)

		results := make(chan error, cafeRoomIntegrationAttempts)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for index := 0; index < cafeRoomIntegrationAttempts; index++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				<-start
				callCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
				defer cancel()
				accountID := fixture.accountID
				_, err := repo.Create(callCtx, &service.CafeRoom{
					Code:      fmt.Sprintf("S157-ACCOUNT-%d-%d", fixture.sequence, index),
					Name:      fmt.Sprintf("S157 account race %d", index),
					PlanID:    fixture.planID,
					AccountID: &accountID,
					Status:    service.CafeRoomStatusEnabled,
				})
				results <- err
			}(index)
		}
		close(start)
		waitForCafeRoomIntegrationWorkers(t, &wg)
		close(results)

		successes := 0
		for err := range results {
			if err == nil {
				successes++
				continue
			}
			require.ErrorIs(t, err, service.ErrCafeAccountAssigned)
		}
		require.Equal(t, 1, successes)

		count, err := fixture.client.CafeRoom.Query().Where(
			caferoom.AccountIDEQ(fixture.accountID),
			caferoom.StatusIn(service.CafeRoomStatusEnabled, service.CafeRoomStatusMaintenance),
		).Count(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("serializes one live round per room", func(t *testing.T) {
		ctx := context.Background()
		fixture := newCafeRoomRepositoryIntegrationFixture(t)
		repo := NewCafeRoomRepository(fixture.client)
		accountID := fixture.accountID
		room, err := repo.Create(ctx, &service.CafeRoom{
			Code:      fmt.Sprintf("S157-ROUND-%d", fixture.sequence),
			Name:      "S157 round race",
			PlanID:    fixture.planID,
			AccountID: &accountID,
			Status:    service.CafeRoomStatusEnabled,
		})
		require.NoError(t, err)

		results := make(chan error, cafeRoomIntegrationAttempts)
		start := make(chan struct{})
		var wg sync.WaitGroup
		now := time.Now().UTC()
		for index := 0; index < cafeRoomIntegrationAttempts; index++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				callCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
				defer cancel()
				_, err := repo.CreateOpenRound(callCtx, room.ID, now)
				results <- err
			}()
		}
		close(start)
		waitForCafeRoomIntegrationWorkers(t, &wg)
		close(results)

		successes := 0
		for err := range results {
			if err == nil {
				successes++
				continue
			}
			require.ErrorIs(t, err, service.ErrCafeRoundExists)
		}
		require.Equal(t, 1, successes)

		count, err := fixture.client.GroupBuyRound.Query().Where(
			groupbuyround.CafeRoomIDEQ(room.ID),
			groupbuyround.StatusIn(service.CafeRoundStatusOpen, "activating", "active"),
		).Count(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("persists anonymous lobby facts in namespaced redis", func(t *testing.T) {
		ctx := context.Background()
		rdb := testRedis(t)
		now := time.Now().UTC().Truncate(time.Second)
		lobby := service.NewCafeLobbyActivityService(rdb, &config.Config{
			Timezone: "UTC",
			JWT:      config.JWTConfig{Secret: "s157-test-secret"},
		})

		lobby.RecordPersistedUsage(101, now.Add(-20*time.Minute))
		lobby.RecordPersistedUsage(202, now.Add(-2*time.Minute))
		lobby.RecordPersistedUsage(101, now.Add(-5*time.Minute))

		var snapshot service.CafeLobbyActivity
		deadline := time.Now().Add(5 * time.Second)
		for {
			snapshot = lobby.Snapshot(ctx)
			if snapshot.Available && snapshot.UniqueUsers == 2 && snapshot.SuccessfulRequests == 3 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("anonymous lobby facts did not persist before deadline: available=%t users=%d requests=%d", snapshot.Available, snapshot.UniqueUsers, snapshot.SuccessfulRequests)
			}
			time.Sleep(20 * time.Millisecond)
		}

		require.Equal(t, now.Format("2006-01-02"), snapshot.Date)
		require.Equal(t, "UTC", snapshot.Timezone)
		require.Equal(t, "今日使用用户", snapshot.Label)
		require.Equal(t, 50, snapshot.DisplayMax)
		require.Len(t, snapshot.Avatars, 2)
		for _, avatar := range snapshot.Avatars {
			require.Len(t, avatar.AvatarSeed, 16)
			require.GreaterOrEqual(t, avatar.SeatIndex, 1)
			require.LessOrEqual(t, avatar.SeatIndex, snapshot.DisplayMax)
			require.Contains(t, []string{"recent", "today"}, avatar.Activity)
		}

		userTTL, err := rdb.TTL(ctx, "cafe:daily-users:"+snapshot.Date).Result()
		require.NoError(t, err)
		assertTTLWithin(t, userTTL, 71*time.Hour, 72*time.Hour)
		requestTTL, err := rdb.TTL(ctx, "cafe:daily-requests:"+snapshot.Date).Result()
		require.NoError(t, err)
		assertTTLWithin(t, requestTTL, 71*time.Hour, 72*time.Hour)
	})
}

type cafeRoomRepositoryIntegrationFixture struct {
	client    *dbent.Client
	planID    int64
	accountID int64
	sequence  int64
}

func newCafeRoomRepositoryIntegrationFixture(t *testing.T) cafeRoomRepositoryIntegrationFixture {
	t.Helper()
	ctx := context.Background()
	client := testEntClient(t)
	sequence := time.Now().UnixNano()

	group, err := client.Group.Create().
		SetName(fmt.Sprintf("S157 Cafe Group %d", sequence)).
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetAccessMode(service.CafeRoomGroupAccessMode).
		Save(ctx)
	require.NoError(t, err)

	plan, err := client.GroupBuyPlan.Create().
		SetTitle(fmt.Sprintf("S157 Cafe Plan %d", sequence)).
		SetTotalShares(2).
		SetSeatCount(2).
		SetPricePerShare(1).
		SetPricePerSeat(1).
		SetTargetGroupID(group.ID).
		SetFulfillmentMode(service.CafeRoomFulfillmentMode).
		SetStatus(service.GroupBuyPlanStatusActive).
		Save(ctx)
	require.NoError(t, err)

	account, err := client.Account.Create().
		SetName(fmt.Sprintf("S157 Cafe Account %d", sequence)).
		SetPlatform(service.PlatformOpenAI).
		SetType(service.AccountTypeAPIKey).
		SetCredentials(map[string]any{"test": true}).
		SetExtra(map[string]any{}).
		SetConcurrency(1).
		SetPriority(1).
		SetStatus(service.StatusActive).
		SetSchedulable(true).
		Save(ctx)
	require.NoError(t, err)

	return cafeRoomRepositoryIntegrationFixture{
		client:    client,
		planID:    plan.ID,
		accountID: account.ID,
		sequence:  sequence,
	}
}

func waitForCafeRoomIntegrationWorkers(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(25 * time.Second):
		t.Fatal("Cafe Room integration workers did not finish before deadline")
	}
}

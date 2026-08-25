//go:build integration

package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/stretchr/testify/require"
)

const (
	cafeRoomPostgresConcurrencyCalls    = 100
	cafeRoomPostgresConcurrencyInFlight = 24
	cafeRoomPostgresActivationSeats     = 4
)

type cafeRoomPostgresOrderResult struct {
	orderID int64
	err     error
}

func TestCafeRoomPostgresConcurrencyIntegration(t *testing.T) {
	client := newCafeRoomOrderPostgresIntegrationClient(t)

	t.Run("one final share winner across one hundred distinct requests", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		now := time.Date(2026, 8, 4, 1, 30, 0, 0, time.UTC)
		fixture := newCafeRoomOrderFixture(t, ctx, client, now, 1)
		users := make([]*User, 0, cafeRoomPostgresConcurrencyCalls)
		users = append(users, fixture.user)
		for index := 1; index < cafeRoomPostgresConcurrencyCalls; index++ {
			users = append(users, createGroupBuyTestUser(
				t,
				ctx,
				client,
				fmt.Sprintf("cafe-concurrency-seat-%03d@example.com", index),
			))
		}

		results := runCafeRoomPostgresOrderRace(t, fixture, users)
		successes := 0
		winnerOrderID := int64(0)
		for _, result := range results {
			if result.err == nil {
				successes++
				winnerOrderID = result.orderID
				continue
			}
			require.ErrorIs(t, result.err, ErrCafeShareUnavailable)
			require.Zero(t, result.orderID)
		}
		require.Len(t, results, cafeRoomPostgresConcurrencyCalls)
		require.Equal(t, 1, successes)
		require.NotZero(t, winnerOrderID)

		orders, err := client.PaymentOrder.Query().Where(paymentorder.PlanIDEQ(fixture.plan.ID)).All(ctx)
		require.NoError(t, err)
		require.Len(t, orders, 1)
		require.Equal(t, winnerOrderID, orders[0].ID)
		require.Equal(t, OrderStatusPending, orders[0].Status)

		seats, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).All(ctx)
		require.NoError(t, err)
		require.Len(t, seats, 1)
		require.Equal(t, GroupBuySeatStatusLocked, seats[0].Status)
		require.Nil(t, seats[0].SeatNo)
		require.NotNil(t, seats[0].MembershipID)
		require.Equal(t, 1, seats[0].ShareCount)
		require.NotNil(t, seats[0].OrderID)
		require.Equal(t, winnerOrderID, *seats[0].OrderID)

		round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
		require.NoError(t, err)
		require.Equal(t, CafeRoundStatusOpen, round.Status)
		require.Equal(t, 1, round.ReservedSeats)
		require.Equal(t, 1, round.ReservedShares)
		require.Zero(t, round.PaidSeats)
		require.Zero(t, round.PaidShares)

		lockEvents, err := client.GroupBuyEvent.Query().Where(
			groupbuyevent.RoundIDEQ(fixture.round.ID),
			groupbuyevent.EventTypeEQ(groupBuyEventSharesLocked),
		).Count(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, lockEvents)
	})

	t.Run("one coherent paid full activation across one hundred retries", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		now := time.Date(2026, 8, 4, 1, 35, 0, 0, time.UTC)
		fixture := newCafeActivationFixture(t, ctx, client, now, cafeRoomPostgresActivationSeats)
		var keySequence atomic.Int64
		fixture.service.generateKey = func() (string, error) {
			sequence := keySequence.Add(1)
			return fmt.Sprintf("sk-cafe-concurrency-%032d", sequence), nil
		}

		errors := runCafeRoomPostgresActivationRace(t, fixture)
		require.Len(t, errors, cafeRoomPostgresConcurrencyCalls)
		for _, err := range errors {
			require.NoError(t, err)
		}

		round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
		require.NoError(t, err)
		require.Equal(t, GroupBuyRoundStatusActive, round.Status)
		require.NotNil(t, round.ActivationToken)
		require.NotEmpty(t, *round.ActivationToken)
		require.NotNil(t, round.ActivatedAt)
		require.Equal(t, now, *round.ActivatedAt)
		require.NotNil(t, round.EntitlementExpiresAt)
		require.Equal(t, now.AddDate(0, 0, fixture.plan.ValidityDays), *round.EntitlementExpiresAt)

		seats, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).All(ctx)
		require.NoError(t, err)
		require.Len(t, seats, cafeRoomPostgresActivationSeats)
		seatsByID := make(map[int64]*dbent.GroupBuySeat, len(seats))
		keyIDsBySeat := make(map[int64]int64, len(seats))
		for _, seat := range seats {
			require.Equal(t, GroupBuySeatStatusActive, seat.Status)
			require.NotNil(t, seat.BoundAPIKeyID)
			require.NotNil(t, seat.ActivatedAt)
			require.Equal(t, *round.ActivatedAt, *seat.ActivatedAt)
			require.NotNil(t, seat.ExpiresAt)
			require.Equal(t, *round.EntitlementExpiresAt, *seat.ExpiresAt)
			seatsByID[seat.ID] = seat
			keyIDsBySeat[seat.ID] = *seat.BoundAPIKeyID
		}

		keys, err := client.APIKey.Query().Where(
			apikey.ManagedSourceTypeEQ(APIKeyManagedSourceCafeRoomSeat),
			apikey.DeletedAtIsNil(),
		).All(ctx)
		require.NoError(t, err)
		require.Len(t, keys, cafeRoomPostgresActivationSeats)
		seenKeyIDs := make(map[int64]struct{}, len(keys))
		for _, key := range keys {
			require.Equal(t, StatusAPIKeyActive, key.Status)
			require.NotNil(t, key.ManagedSourceID)
			seat, ok := seatsByID[*key.ManagedSourceID]
			require.True(t, ok)
			require.Equal(t, seat.UserID, key.UserID)
			require.NotNil(t, key.GroupID)
			require.Equal(t, fixture.groupID, *key.GroupID)
			require.Equal(t, keyIDsBySeat[seat.ID], key.ID)
			require.NotNil(t, key.ExpiresAt)
			require.Equal(t, *round.EntitlementExpiresAt, *key.ExpiresAt)
			_, duplicate := seenKeyIDs[key.ID]
			require.False(t, duplicate)
			seenKeyIDs[key.ID] = struct{}{}
		}

		bindings, err := client.APIKeyAccountBinding.Query().Where(
			apikeyaccountbinding.RoundIDEQ(fixture.round.ID),
			apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive),
		).All(ctx)
		require.NoError(t, err)
		require.Len(t, bindings, cafeRoomPostgresActivationSeats)
		seenBindingSeats := make(map[int64]struct{}, len(bindings))
		for _, binding := range bindings {
			require.NotNil(t, binding.SeatID)
			seatID := *binding.SeatID
			seat, ok := seatsByID[seatID]
			require.True(t, ok)
			require.Equal(t, keyIDsBySeat[seat.ID], binding.APIKeyID)
			require.Equal(t, seat.UserID, binding.UserID)
			require.Equal(t, fixture.groupID, binding.GroupID)
			require.Equal(t, fixture.accountID, binding.AccountID)
			require.Equal(t, fixture.room.ID, binding.CafeRoomID)
			require.Equal(t, fixture.round.ID, binding.RoundID)
			require.True(t, binding.StrictMode)
			require.Equal(t, *round.ActivatedAt, binding.StartsAt)
			require.Equal(t, *round.EntitlementExpiresAt, binding.ExpiresAt)
			_, duplicate := seenBindingSeats[seatID]
			require.False(t, duplicate)
			seenBindingSeats[seatID] = struct{}{}
		}

		activationEvents, err := client.GroupBuyEvent.Query().Where(
			groupbuyevent.RoundIDEQ(fixture.round.ID),
			groupbuyevent.EventTypeEQ(groupBuyEventRoundActivated),
		).Count(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, activationEvents)
		subscriptionCount, err := client.UserSubscription.Query().Count(ctx)
		require.NoError(t, err)
		require.Zero(t, subscriptionCount)
	})
}

func runCafeRoomPostgresOrderRace(t *testing.T, fixture cafeRoomOrderFixture, users []*User) []cafeRoomPostgresOrderResult {
	t.Helper()
	results := make(chan cafeRoomPostgresOrderResult, len(users))
	start := make(chan struct{})
	inFlight := make(chan struct{}, cafeRoomPostgresConcurrencyInFlight)
	var wg sync.WaitGroup
	for _, user := range users {
		userID := user.ID
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			inFlight <- struct{}{}
			defer func() { <-inFlight }()
			callCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			order, _, err := fixture.orderService.lockSeatAndCreateOrder(
				callCtx,
				CreateOrderRequest{UserID: userID, PaymentType: "integration"},
				fixture.room.ID,
				1,
				&PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30},
				0,
				fixture.plan.PricePerShare,
				nil,
			)
			if order == nil {
				results <- cafeRoomPostgresOrderResult{err: err}
				return
			}
			results <- cafeRoomPostgresOrderResult{orderID: order.ID, err: err}
		}()
	}
	close(start)
	waitForCafeRoomPostgresConcurrencyWorkers(t, &wg)
	close(results)

	collected := make([]cafeRoomPostgresOrderResult, 0, len(users))
	for result := range results {
		collected = append(collected, result)
	}
	return collected
}

func runCafeRoomPostgresActivationRace(t *testing.T, fixture cafeActivationFixture) []error {
	t.Helper()
	results := make(chan error, cafeRoomPostgresConcurrencyCalls)
	start := make(chan struct{})
	inFlight := make(chan struct{}, cafeRoomPostgresConcurrencyInFlight)
	var wg sync.WaitGroup
	for index := 0; index < cafeRoomPostgresConcurrencyCalls; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			inFlight <- struct{}{}
			defer func() { <-inFlight }()
			callCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			results <- fixture.service.ActivateRound(callCtx, fixture.round.ID)
		}()
	}
	close(start)
	waitForCafeRoomPostgresConcurrencyWorkers(t, &wg)
	close(results)

	collected := make([]error, 0, cafeRoomPostgresConcurrencyCalls)
	for err := range results {
		collected = append(collected, err)
	}
	return collected
}

func waitForCafeRoomPostgresConcurrencyWorkers(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(120 * time.Second):
		t.Fatal("Pixel Cafe PostgreSQL concurrency workers did not finish before deadline")
	}
}

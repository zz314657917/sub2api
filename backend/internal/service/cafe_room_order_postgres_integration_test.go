//go:build integration

package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const cafeRoomOrderPostgresContenders = 16

func TestCafeRoomOrderLastSharePostgresIntegration(t *testing.T) {
	ctx := context.Background()
	client := newCafeRoomOrderPostgresIntegrationClient(t)
	now := time.Date(2026, 8, 3, 20, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 10)
	cfg := &PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30}
	firstOrder, _, err := fixture.orderService.lockSeatAndCreateOrder(
		ctx,
		CreateOrderRequest{UserID: fixture.user.ID, PaymentType: "integration"},
		fixture.room.ID,
		9,
		cfg,
		0,
		fixture.plan.PricePerShare*9,
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, firstOrder)

	users := make([]*User, 0, cafeRoomOrderPostgresContenders)
	for index := 0; index < cafeRoomOrderPostgresContenders; index++ {
		users = append(users, createGroupBuyTestUser(t, ctx, client, fmt.Sprintf("cafe-order-postgres-%d@example.com", index)))
	}

	type result struct {
		orderID int64
		err     error
	}
	results := make(chan result, len(users))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, user := range users {
		userID := user.ID
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			callCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
			if order != nil {
				results <- result{orderID: order.ID, err: err}
				return
			}
			results <- result{err: err}
		}()
	}
	close(start)
	waitForCafeRoomOrderPostgresWorkers(t, &wg)
	close(results)

	var winnerOrderID int64
	successes := 0
	for result := range results {
		if result.err == nil {
			successes++
			winnerOrderID = result.orderID
			continue
		}
		require.ErrorIs(t, result.err, ErrCafeShareUnavailable)
	}
	require.Equal(t, 1, successes)
	require.NotZero(t, winnerOrderID)

	seats, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, seats, 2)
	sharesByOrder := make(map[int64]int, len(seats))
	var winnerUserID int64
	for _, seat := range seats {
		require.Equal(t, GroupBuySeatStatusLocked, seat.Status)
		require.Equal(t, fixture.round.ID, seat.RoundID)
		require.Equal(t, fixture.plan.ID, seat.PlanID)
		require.Nil(t, seat.SeatNo)
		require.NotNil(t, seat.MembershipID)
		require.NotNil(t, seat.OrderID)
		sharesByOrder[*seat.OrderID] = seat.ShareCount
		if *seat.OrderID == winnerOrderID {
			winnerUserID = seat.UserID
		}
	}
	require.Equal(t, 9, sharesByOrder[firstOrder.ID])
	require.Equal(t, 1, sharesByOrder[winnerOrderID])

	memberships, err := client.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, memberships, 2)
	reservedShares := 0
	for _, membership := range memberships {
		reservedShares += membership.ReservedShares
	}
	require.Equal(t, 10, reservedShares)

	orders, err := client.PaymentOrder.Query().Where(paymentorder.IDEQ(winnerOrderID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	order := orders[0]
	require.Equal(t, OrderStatusPending, order.Status)
	require.NotNil(t, order.PlanID)
	require.Equal(t, fixture.plan.ID, *order.PlanID)
	require.Equal(t, winnerUserID, order.UserID)

	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, CafeRoundStatusOpen, round.Status)
	require.Equal(t, 10, round.ReservedSeats)
	require.Equal(t, 10, round.ReservedShares)
	require.Zero(t, round.PaidSeats)
	require.Zero(t, round.PaidShares)
}

func newCafeRoomOrderPostgresIntegrationClient(t *testing.T) *dbent.Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:18.1-alpine3.23",
		tcpostgres.WithDatabase("sub2api_cafe_order_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "TimeZone=UTC")
	require.NoError(t, err)
	db, err := openCafeRoomOrderPostgres(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(ctx))
	return client
}

func openCafeRoomOrderPostgres(ctx context.Context, dsn string) (*sql.DB, error) {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err == nil {
			return db, nil
		}
		lastErr = err
		_ = db.Close()
		time.Sleep(250 * time.Millisecond)
	}
	return nil, fmt.Errorf("PostgreSQL did not become ready: %w", lastErr)
}

func waitForCafeRoomOrderPostgresWorkers(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(25 * time.Second):
		t.Fatal("Cafe Room PostgreSQL last-seat workers did not finish before deadline")
	}
}

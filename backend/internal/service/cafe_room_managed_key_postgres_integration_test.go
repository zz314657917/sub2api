//go:build integration

package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const cafeManagedKeyConcurrencyCalls = 20

type cafeManagedKeyPostgresFixture struct {
	client    *dbent.Client
	service   *service.APIKeyService
	userID    int64
	groupID   int64
	accountID int64
	roomID    int64
	roundID   int64
	seatID    int64
	keyID     int64
	bindingID int64
	expiresAt time.Time
}

func TestCafeManagedKeyPostgresIntegration(t *testing.T) {
	t.Run("active_inactive_and_reenable", func(t *testing.T) {
		fixture := newCafeManagedKeyPostgresFixture(t, service.StatusAPIKeyActive)
		inactive := "inactive"
		pausedName := "Paused Cafe key"
		updated, err := fixture.service.Update(context.Background(), fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{
			Name:        &pausedName,
			Status:      &inactive,
			IPWhitelist: &[]string{"127.0.0.1"},
		})
		require.NoError(t, err)
		require.Equal(t, inactive, updated.Status)
		assertCafeManagedKeyPostgresState(t, fixture, pausedName, inactive, []string{"127.0.0.1"}, nil)

		active := service.StatusAPIKeyActive
		updated, err = fixture.service.Update(context.Background(), fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Status: &active})
		require.NoError(t, err)
		require.Equal(t, active, updated.Status)
		assertCafeManagedKeyPostgresState(t, fixture, pausedName, active, []string{"127.0.0.1"}, nil)
	})

	t.Run("legacy_disabled_reenable_is_idempotent", func(t *testing.T) {
		fixture := newCafeManagedKeyPostgresFixture(t, service.StatusAPIKeyDisabled)
		active := service.StatusAPIKeyActive
		for range 2 {
			updated, err := fixture.service.Update(context.Background(), fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Status: &active})
			require.NoError(t, err)
			require.Equal(t, active, updated.Status)
		}
		assertCafeManagedKeyPostgresState(t, fixture, "Managed Cafe key", active, nil, nil)
		keyCount, err := fixture.client.APIKey.Query().Count(context.Background())
		require.NoError(t, err)
		require.Equal(t, 1, keyCount)
		bindingCount, err := fixture.client.APIKeyAccountBinding.Query().Count(context.Background())
		require.NoError(t, err)
		require.Equal(t, 1, bindingCount)
	})

	t.Run("ownership_protected_fields_and_terminal_state", func(t *testing.T) {
		fixture := newCafeManagedKeyPostgresFixture(t, "inactive")
		active := service.StatusAPIKeyActive
		_, err := fixture.service.Update(context.Background(), fixture.keyID, fixture.userID+1000, service.UpdateAPIKeyRequest{Status: &active})
		require.ErrorIs(t, err, service.ErrInsufficientPerms)

		invalid := service.StatusAPIKeyDisabled
		_, err = fixture.service.Update(context.Background(), fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Status: &invalid})
		require.ErrorIs(t, err, service.ErrCafeManagedKeyStatusInvalid)

		quota := 99.0
		_, err = fixture.service.Update(context.Background(), fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Quota: &quota})
		require.ErrorIs(t, err, service.ErrCafeManagedKeyProtected)

		_, err = fixture.client.APIKey.UpdateOneID(fixture.keyID).SetStatus(service.StatusAPIKeyExpired).Save(context.Background())
		require.NoError(t, err)
		inactive := "inactive"
		_, err = fixture.service.Update(context.Background(), fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Status: &inactive})
		require.ErrorIs(t, err, service.ErrCafeManagedKeyStateUnavailable)
		assertCafeManagedKeyPostgresState(t, fixture, "Managed Cafe key", service.StatusAPIKeyExpired, nil, nil)
	})
}

func TestCafeManagedKeyEnableValidationPostgresIntegration(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(context.Context, *testing.T, cafeManagedKeyPostgresFixture)
	}{
		{
			name: "round_inactive",
			mutate: func(ctx context.Context, t *testing.T, fixture cafeManagedKeyPostgresFixture) {
				_, err := fixture.client.GroupBuyRound.UpdateOneID(fixture.roundID).SetStatus("closed").Save(ctx)
				require.NoError(t, err)
			},
		},
		{
			name: "seat_inactive",
			mutate: func(ctx context.Context, t *testing.T, fixture cafeManagedKeyPostgresFixture) {
				_, err := fixture.client.GroupBuySeat.UpdateOneID(fixture.seatID).SetStatus("expired").Save(ctx)
				require.NoError(t, err)
			},
		},
		{
			name: "binding_inactive",
			mutate: func(ctx context.Context, t *testing.T, fixture cafeManagedKeyPostgresFixture) {
				_, err := fixture.client.APIKeyAccountBinding.UpdateOneID(fixture.bindingID).SetStatus("inactive").Save(ctx)
				require.NoError(t, err)
			},
		},
		{
			name: "entitlement_expired",
			mutate: func(ctx context.Context, t *testing.T, fixture cafeManagedKeyPostgresFixture) {
				_, err := fixture.client.APIKey.UpdateOneID(fixture.keyID).SetExpiresAt(time.Now().UTC().Add(-time.Hour)).Save(ctx)
				require.NoError(t, err)
			},
		},
		{
			name: "account_inactive",
			mutate: func(ctx context.Context, t *testing.T, fixture cafeManagedKeyPostgresFixture) {
				_, err := fixture.client.Account.UpdateOneID(fixture.accountID).SetStatus("disabled").Save(ctx)
				require.NoError(t, err)
			},
		},
		{
			name: "group_inactive",
			mutate: func(ctx context.Context, t *testing.T, fixture cafeManagedKeyPostgresFixture) {
				_, err := fixture.client.Group.UpdateOneID(fixture.groupID).SetStatus("inactive").Save(ctx)
				require.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newCafeManagedKeyPostgresFixture(t, "inactive")
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			test.mutate(ctx, t, fixture)

			active := service.StatusAPIKeyActive
			changedName := "Must not persist"
			_, err := fixture.service.Update(ctx, fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{
				Name:        &changedName,
				Status:      &active,
				IPWhitelist: &[]string{"10.0.0.0/8"},
				IPBlacklist: &[]string{"192.0.2.1"},
			})
			require.Error(t, err)
			require.True(t,
				errors.Is(err, service.ErrCafeManagedKeyEnableUnavailable) || errors.Is(err, service.ErrCafeManagedKeyStateUnavailable),
				"unexpected managed-Key validation error: %v", err,
			)
			assertCafeManagedKeyPostgresState(t, fixture, "Managed Cafe key", "inactive", nil, nil)
		})
	}
}

func TestCafeManagedKeyConcurrencyPostgresIntegration(t *testing.T) {
	fixture := newCafeManagedKeyPostgresFixture(t, "inactive")
	start := make(chan struct{})
	results := make(chan error, cafeManagedKeyConcurrencyCalls)
	var wg sync.WaitGroup
	for range cafeManagedKeyConcurrencyCalls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			active := service.StatusAPIKeyActive
			_, err := fixture.service.Update(ctx, fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Status: &active})
			results <- err
		}()
	}
	close(start)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatal("concurrent managed-Key enables did not finish before deadline")
	}
	close(results)
	for err := range results {
		require.NoError(t, err)
	}

	assertCafeManagedKeyPostgresState(t, fixture, "Managed Cafe key", service.StatusAPIKeyActive, nil, nil)
	keyCount, err := fixture.client.APIKey.Query().Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, keyCount)
	bindingCount, err := fixture.client.APIKeyAccountBinding.Query().Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, bindingCount)
}

func newCafeManagedKeyPostgresFixture(t *testing.T, initialStatus string) cafeManagedKeyPostgresFixture {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:18.1-alpine3.23",
		tcpostgres.WithDatabase("sub2api_cafe_managed_key_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "TimeZone=UTC")
	require.NoError(t, err)
	db := openCafeManagedKeyPostgres(t, ctx, dsn)
	t.Cleanup(func() { _ = db.Close() })
	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(ctx))

	now := time.Now().UTC().Truncate(time.Second)
	expiresAt := now.Add(24 * time.Hour)
	user, err := client.User.Create().
		SetEmail("cafe-managed-key@example.com").
		SetPasswordHash("hash").
		SetUsername("cafe-managed-key").
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().
		SetName("Cafe managed group").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetAccessMode(service.CafeRoomGroupAccessMode).
		SetSubscriptionType(domain.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)
	account, err := client.Account.Create().
		SetName("Cafe managed account").
		SetPlatform(service.PlatformOpenAI).
		SetType("api_key").
		SetStatus(service.StatusActive).
		AddGroupIDs(group.ID).
		Save(ctx)
	require.NoError(t, err)
	plan, err := client.GroupBuyPlan.Create().
		SetTitle("Cafe managed plan").
		SetTotalShares(1).
		SetSeatCount(1).
		SetPricePerShare(10).
		SetPricePerSeat(10).
		SetTargetGroupID(group.ID).
		SetStatus(service.GroupBuyPlanStatusActive).
		Save(ctx)
	require.NoError(t, err)
	room, err := client.CafeRoom.Create().
		SetCode("CAFE-MANAGED-KEY").
		SetName("Cafe managed room").
		SetPlanID(plan.ID).
		SetAccountID(account.ID).
		SetStatus(service.CafeRoomStatusEnabled).
		Save(ctx)
	require.NoError(t, err)
	round, err := client.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetCafeRoomID(room.ID).
		SetAssignedAccountID(account.ID).
		SetStatus(service.GroupBuyRoundStatusActive).
		SetTotalShares(1).
		SetPaidShares(1).
		SetTotalSeats(1).
		SetPaidSeats(1).
		SetDeadlineAt(now.Add(time.Hour)).
		SetActivatedAt(now).
		SetEntitlementExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetSeatNo(1).
		SetStatus(service.GroupBuySeatStatusActive).
		SetShareCount(1).
		SetPaidAt(now).
		SetActivatedAt(now).
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	key, err := client.APIKey.Create().
		SetUserID(user.ID).
		SetKey("sk-cafe-managed-integration-0000000000000001").
		SetName("Managed Cafe key").
		SetGroupID(group.ID).
		SetStatus(initialStatus).
		SetExpiresAt(expiresAt).
		SetManagedSourceType(service.APIKeyManagedSourceCafeRoomSeat).
		SetManagedSourceID(seat.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.UpdateOneID(seat.ID).SetBoundAPIKeyID(key.ID).SetBoundAt(now).Save(ctx)
	require.NoError(t, err)
	binding, err := client.APIKeyAccountBinding.Create().
		SetAPIKeyID(key.ID).
		SetUserID(user.ID).
		SetGroupID(group.ID).
		SetAccountID(account.ID).
		SetCafeRoomID(room.ID).
		SetRoundID(round.ID).
		SetSeatID(seat.ID).
		SetStatus("active").
		SetStrictMode(true).
		SetStartsAt(now).
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)

	apiKeyRepo := repository.NewAPIKeyRepository(client, db)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, nil, nil, nil, nil, nil, nil)
	return cafeManagedKeyPostgresFixture{
		client:    client,
		service:   apiKeyService,
		userID:    user.ID,
		groupID:   group.ID,
		accountID: account.ID,
		roomID:    room.ID,
		roundID:   round.ID,
		seatID:    seat.ID,
		keyID:     key.ID,
		bindingID: binding.ID,
		expiresAt: expiresAt,
	}
}

func openCafeManagedKeyPostgres(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		db, err := sql.Open("postgres", dsn)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err = db.PingContext(pingCtx)
			cancel()
			if err == nil {
				return db
			}
			_ = db.Close()
		}
		lastErr = err
		time.Sleep(250 * time.Millisecond)
	}
	require.NoError(t, fmt.Errorf("PostgreSQL did not become ready: %w", lastErr))
	return nil
}

func assertCafeManagedKeyPostgresState(t *testing.T, fixture cafeManagedKeyPostgresFixture, name, status string, whitelist, blacklist []string) {
	t.Helper()
	key, err := fixture.client.APIKey.Get(context.Background(), fixture.keyID)
	require.NoError(t, err)
	require.Equal(t, name, key.Name)
	require.Equal(t, status, key.Status)
	require.Equal(t, whitelist, key.IPWhitelist)
	require.Equal(t, blacklist, key.IPBlacklist)
}

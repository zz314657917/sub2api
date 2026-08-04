//go:build integration

package routes

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type cafeAuthenticatedHTTPSettings struct {
	enabled bool
}

func (s *cafeAuthenticatedHTTPSettings) GetPublicSettings(context.Context) (*service.PublicSettings, error) {
	return &service.PublicSettings{PixelCafeEnabled: s.enabled}, nil
}

type cafeAuthenticatedHTTPFixture struct {
	client      *dbent.Client
	router      *gin.Engine
	authService *service.AuthService
	settings    *cafeAuthenticatedHTTPSettings
	userA       *service.User
	userB       *service.User
	roomID      int64
	userASeatID int64
	userBSeatID int64
}

func TestCafeAuthenticatedHTTPPostgresIntegration(t *testing.T) {
	fixture := newCafeAuthenticatedHTTPFixture(t)
	tokenA := cafeAuthenticatedHTTPToken(t, fixture.authService, fixture.userA)
	tokenB := cafeAuthenticatedHTTPToken(t, fixture.authService, fixture.userB)

	t.Run("authenticated users receive redacted public rooms and only their own memberships", func(t *testing.T) {
		overview := cafeAuthenticatedHTTPRequest(t, fixture.router, http.MethodGet, "/api/v1/cafe/overview", tokenA, nil)
		require.Equal(t, http.StatusOK, overview.Code)
		require.Contains(t, overview.Body.String(), "CAFE-S169")
		assertCafeAuthenticatedHTTPPublicResponse(t, overview.Body.String())

		rooms := cafeAuthenticatedHTTPRequest(t, fixture.router, http.MethodGet, "/api/v1/cafe/rooms?zone=claude", tokenA, nil)
		require.Equal(t, http.StatusOK, rooms.Code)
		require.Contains(t, rooms.Body.String(), "CAFE-S169")
		assertCafeAuthenticatedHTTPPublicResponse(t, rooms.Body.String())

		userARooms := cafeAuthenticatedHTTPRequest(t, fixture.router, http.MethodGet, "/api/v1/cafe/my-rooms", tokenA, nil)
		require.Equal(t, http.StatusOK, userARooms.Code)
		require.Contains(t, userARooms.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.userASeatID))
		require.NotContains(t, userARooms.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.userBSeatID))
		require.NotContains(t, userARooms.Body.String(), fixture.userB.Email)

		userBRooms := cafeAuthenticatedHTTPRequest(t, fixture.router, http.MethodGet, "/api/v1/cafe/my-rooms", tokenB, nil)
		require.Equal(t, http.StatusOK, userBRooms.Code)
		require.Contains(t, userBRooms.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.userBSeatID))
		require.NotContains(t, userBRooms.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.userASeatID))
		require.NotContains(t, userBRooms.Body.String(), fixture.userA.Email)
	})

	t.Run("missing authorization is rejected before cafe data is returned", func(t *testing.T) {
		response := cafeAuthenticatedHTTPRequest(t, fixture.router, http.MethodGet, "/api/v1/cafe/rooms", "", nil)
		require.Equal(t, http.StatusUnauthorized, response.Code)
		require.Contains(t, response.Body.String(), "UNAUTHORIZED")
		require.NotContains(t, response.Body.String(), "CAFE-S169")
	})

	t.Run("disabled feature fails closed after authentication", func(t *testing.T) {
		fixture.settings.enabled = false
		t.Cleanup(func() { fixture.settings.enabled = true })
		response := cafeAuthenticatedHTTPRequest(t, fixture.router, http.MethodGet, "/api/v1/cafe/rooms", tokenA, nil)
		require.Equal(t, http.StatusNotFound, response.Code)
		require.Contains(t, response.Body.String(), "CAFE_DISABLED")
		require.NotContains(t, response.Body.String(), "CAFE-S169")
	})

	t.Run("agreement rejection cannot create a payment order or mutate a seat", func(t *testing.T) {
		beforeOrders, err := fixture.client.PaymentOrder.Query().Count(context.Background())
		require.NoError(t, err)
		beforeSeats, err := fixture.client.GroupBuySeat.Query().Count(context.Background())
		require.NoError(t, err)

		response := cafeAuthenticatedHTTPRequest(t, fixture.router, http.MethodPost, fmt.Sprintf("/api/v1/cafe/rooms/%d/orders", fixture.roomID), tokenA, []byte(`{"seat_no":3,"payment_type":"alipay","agreement_accepted":false}`))
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "CAFE_AGREEMENT_REQUIRED")

		afterOrders, err := fixture.client.PaymentOrder.Query().Count(context.Background())
		require.NoError(t, err)
		afterSeats, err := fixture.client.GroupBuySeat.Query().Count(context.Background())
		require.NoError(t, err)
		require.Equal(t, beforeOrders, afterOrders)
		require.Equal(t, beforeSeats, afterSeats)
	})
}

func newCafeAuthenticatedHTTPFixture(t *testing.T) cafeAuthenticatedHTTPFixture {
	t.Helper()
	ctx := context.Background()
	client, sqlDB := newCafeAuthenticatedHTTPPostgresClient(t)
	settings := &cafeAuthenticatedHTTPSettings{enabled: true}

	userAEntity, err := client.User.Create().
		SetEmail("cafe-s169-user-a@example.test").
		SetUsername("cafe-s169-user-a").
		SetPasswordHash("not-a-real-password").
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	userBEntity, err := client.User.Create().
		SetEmail("cafe-s169-user-b@example.test").
		SetUsername("cafe-s169-user-b").
		SetPasswordHash("not-a-real-password").
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	group, err := client.Group.Create().
		SetName("Cafe S169 Group").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType("subscription").
		SetMonthlyLimitUsd(100).
		Save(ctx)
	require.NoError(t, err)
	plan, err := client.GroupBuyPlan.Create().
		SetTitle("Cafe S169 Plan").
		SetProductKey(service.GroupBuyProductTokenPinPinPin).
		SetTotalShares(3).
		SetSeatCount(3).
		SetPricePerShare(12).
		SetPricePerSeat(12).
		SetQuotaPerShareLabel("synthetic quota").
		SetQuotaLabel("synthetic quota").
		SetMaxSharesPerUser(1).
		SetTargetGroupID(group.ID).
		SetTierGroupIds(map[string]int64{"3": group.ID}).
		SetValidityDays(30).
		SetTimeoutMinutes(60).
		SetLaunchMode(service.GroupBuyLaunchModeManual).
		SetRefundMode(service.GroupBuyRefundModeBalanceCredit).
		SetFulfillmentMode(service.CafeRoomFulfillmentMode).
		SetStatus(service.GroupBuyPlanStatusActive).
		Save(ctx)
	require.NoError(t, err)
	room, err := client.CafeRoom.Create().
		SetCode("CAFE-S169").
		SetName("S169 Claude Room").
		SetPlanID(plan.ID).
		SetZoneKey("claude").
		SetThemeKey("warm_wood").
		SetStatus(service.CafeRoomStatusEnabled).
		SetFeatured(true).
		SetMetadata(map[string]any{
			"account_email": "private-cafe-s169@example.test",
			"account_id":    "s169-private-account-id",
			"api_key":       "s169-private-key-value",
		}).
		Save(ctx)
	require.NoError(t, err)
	now := time.Now().UTC()
	round, err := client.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetCafeRoomID(room.ID).
		SetStatus("active").
		SetTotalShares(3).
		SetPaidShares(2).
		SetTotalSeats(3).
		SetPaidSeats(2).
		SetDeadlineAt(now.Add(time.Hour)).
		SetStartedAt(now.Add(-time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	userASeat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(userAEntity.ID).
		SetSeatNo(1).
		SetStatus(service.GroupBuySeatStatusActive).
		SetActivatedAt(now.Add(-time.Hour)).
		SetExpiresAt(now.Add(time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	userBSeat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(userBEntity.ID).
		SetSeatNo(2).
		SetStatus(service.GroupBuySeatStatusActive).
		SetActivatedAt(now.Add(-time.Hour)).
		SetExpiresAt(now.Add(time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(client, sqlDB)
	userService := service.NewUserService(userRepo, nil, nil, nil)
	authConfig := &config.Config{JWT: config.JWTConfig{Secret: "s169-local-jwt-secret", AccessTokenExpireMinutes: 30}}
	authService := service.NewAuthService(client, userRepo, nil, nil, authConfig, nil, nil, nil, nil, nil, nil, nil)
	userA, err := userService.GetByID(ctx, userAEntity.ID)
	require.NoError(t, err)
	userB, err := userService.GetByID(ctx, userBEntity.ID)
	require.NoError(t, err)

	previousCoordinator := service.DefaultIdempotencyCoordinator()
	service.SetDefaultIdempotencyCoordinator(nil)
	t.Cleanup(func() { service.SetDefaultIdempotencyCoordinator(previousCoordinator) })

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterUserRoutes(
		v1,
		&handler.Handlers{
			Cafe: handler.NewCafeHandler(
				service.NewCafePublicService(client, settings),
				service.NewCafeRoomOrderService(client, nil, nil, settings),
			),
			Admin: &handler.AdminHandlers{},
		},
		middleware.NewJWTAuthMiddleware(authService, userService),
		middleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() }),
		nil,
	)

	return cafeAuthenticatedHTTPFixture{
		client:      client,
		router:      router,
		authService: authService,
		settings:    settings,
		userA:       userA,
		userB:       userB,
		roomID:      room.ID,
		userASeatID: userASeat.ID,
		userBSeatID: userBSeat.ID,
	}
}

func cafeAuthenticatedHTTPToken(t *testing.T, authService *service.AuthService, user *service.User) string {
	t.Helper()
	token, err := authService.GenerateToken(user)
	require.NoError(t, err)
	return token
}

func cafeAuthenticatedHTTPRequest(t *testing.T, router *gin.Engine, method, path, token string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if len(body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if method == http.MethodPost {
		request.Header.Set("Idempotency-Key", "s169-local-idempotency-key")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertCafeAuthenticatedHTTPPublicResponse(t *testing.T, body string) {
	t.Helper()
	for _, prohibited := range []string{
		"private-cafe-s169@example.test",
		"s169-private-account-id",
		"s169-private-key-value",
		"account_email",
		"assigned_account_id",
		"metadata",
	} {
		require.NotContains(t, body, prohibited)
	}
}

func newCafeAuthenticatedHTTPPostgresClient(t *testing.T) (*dbent.Client, *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:18.1-alpine3.23",
		tcpostgres.WithDatabase("sub2api_cafe_s169_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "TimeZone=UTC")
	require.NoError(t, err)
	sqlDB := openCafeAuthenticatedHTTPPostgres(t, ctx, dsn)
	t.Cleanup(func() { _ = sqlDB.Close() })

	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(ctx))
	ensureCafeAuthenticatedHTTPUserAvatarTable(t, ctx, sqlDB)
	return client, sqlDB
}

func ensureCafeAuthenticatedHTTPUserAvatarTable(t *testing.T, ctx context.Context, sqlDB *sql.DB) {
	t.Helper()
	_, err := sqlDB.ExecContext(ctx, `
CREATE TABLE user_avatars (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    storage_provider VARCHAR(20) NOT NULL DEFAULT 'database',
    storage_key TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    content_type VARCHAR(100) NOT NULL DEFAULT '',
    byte_size BIGINT NOT NULL DEFAULT 0,
    sha256 VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)
	require.NoError(t, err)
}

func openCafeAuthenticatedHTTPPostgres(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		sqlDB, err := sql.Open("postgres", dsn)
		if err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = sqlDB.PingContext(pingCtx)
		cancel()
		if err == nil {
			return sqlDB
		}
		lastErr = err
		_ = sqlDB.Close()
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("PostgreSQL did not become ready: %v", lastErr)
	return nil
}

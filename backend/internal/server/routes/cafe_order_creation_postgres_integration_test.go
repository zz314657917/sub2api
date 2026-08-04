//go:build integration

package routes

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const cafeOrderCreationSyntheticPKey = "s170-synthetic-pkey-never-external"

type cafeOrderCreationFixture struct {
	client             *dbent.Client
	sqlDB              *sql.DB
	router             *gin.Engine
	token              string
	roomID             int64
	roundID            int64
	providerInstanceID string
}

type cafeOrderCreationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		OrderID     int64  `json:"order_id"`
		RoomID      int64  `json:"room_id"`
		RoundID     int64  `json:"round_id"`
		SeatNo      int    `json:"seat_no"`
		Status      string `json:"status"`
		PaymentType string `json:"payment_type"`
		PayURL      string `json:"pay_url"`
		OutTradeNo  string `json:"out_trade_no"`
	} `json:"data"`
}

func TestCafeOrderCreationPostgresIntegration(t *testing.T) {
	t.Run("successful order persists one seat lock and durable idempotency replay", func(t *testing.T) {
		fixture := newCafeOrderCreationFixture(t, cafeOrderCreationValidEasyPayConfig())
		body := []byte(`{"seat_no":1,"payment_type":"alipay","agreement_accepted":true,"return_url":"https://api.pixel-cafe-s170.invalid/payment/result"}`)

		first := cafeOrderCreationRequest(t, fixture.router, fixture.token, fixture.roomID, "s170-success-key", body)
		require.Equal(t, http.StatusOK, first.Code)
		require.NotContains(t, first.Body.String(), cafeOrderCreationSyntheticPKey)
		firstData := decodeCafeOrderCreationResponse(t, first)
		require.Equal(t, 0, firstData.Code)
		require.Equal(t, "success", firstData.Message)
		require.NotZero(t, firstData.Data.OrderID)
		require.Equal(t, fixture.roomID, firstData.Data.RoomID)
		require.Equal(t, fixture.roundID, firstData.Data.RoundID)
		require.Equal(t, 1, firstData.Data.SeatNo)
		require.Equal(t, service.OrderStatusPending, firstData.Data.Status)
		require.Equal(t, payment.TypeAlipay, firstData.Data.PaymentType)
		require.NotEmpty(t, firstData.Data.OutTradeNo)
		assertCafeOrderCreationHostedEasyPayURL(t, firstData.Data.PayURL)

		replay := cafeOrderCreationRequest(t, fixture.router, fixture.token, fixture.roomID, "s170-success-key", body)
		require.Equal(t, http.StatusOK, replay.Code)
		require.Equal(t, "true", replay.Header().Get("X-Idempotency-Replayed"))
		replayData := decodeCafeOrderCreationResponse(t, replay)
		require.Equal(t, firstData.Code, replayData.Code)
		require.Equal(t, firstData.Message, replayData.Message)
		require.Equal(t, firstData.Data, replayData.Data)

		ctx := context.Background()
		order, err := fixture.client.PaymentOrder.Get(ctx, firstData.Data.OrderID)
		require.NoError(t, err)
		require.Equal(t, service.OrderStatusPending, order.Status)
		require.Equal(t, payment.OrderTypeGroupBuy, order.OrderType)
		require.Equal(t, payment.TypeAlipay, order.PaymentType)
		require.NotNil(t, order.ProviderInstanceID)
		require.Equal(t, fixture.providerInstanceID, *order.ProviderInstanceID)
		require.NotNil(t, order.ProviderKey)
		require.Equal(t, payment.TypeEasyPay, *order.ProviderKey)
		require.NotNil(t, order.PayURL)
		assertCafeOrderCreationHostedEasyPayURL(t, *order.PayURL)
		require.NotContains(t, *order.PayURL, cafeOrderCreationSyntheticPKey)
		require.NotContains(t, order.ProviderSnapshot, "pkey")

		seats, err := fixture.client.GroupBuySeat.Query().Where(groupbuyseat.OrderIDEQ(order.ID)).All(ctx)
		require.NoError(t, err)
		require.Len(t, seats, 1)
		require.Equal(t, service.GroupBuySeatStatusLocked, seats[0].Status)
		require.NotNil(t, seats[0].SeatNo)
		require.Equal(t, 1, *seats[0].SeatNo)

		round, err := fixture.client.GroupBuyRound.Get(ctx, fixture.roundID)
		require.NoError(t, err)
		require.Equal(t, 1, round.ReservedSeats)
		require.Equal(t, 1, round.ReservedShares)
		require.Zero(t, round.PaidSeats)

		eventCount, err := fixture.client.GroupBuyEvent.Query().Where(
			groupbuyevent.RoundIDEQ(fixture.roundID),
			groupbuyevent.EventTypeEQ("shares_locked"),
		).Count(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, eventCount)
		auditCount, err := fixture.client.PaymentAuditLog.Query().Where(
			paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)),
			paymentauditlog.ActionEQ("ORDER_CREATED"),
		).Count(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, auditCount)
		require.Equal(t, 1, cafeOrderCreationIdempotencyCount(t, ctx, fixture.sqlDB, "s170-success-key", service.IdempotencyStatusSucceeded))
	})

	t.Run("provider construction failure marks order failed and releases its seat", func(t *testing.T) {
		fixture := newCafeOrderCreationFixture(t, cafeOrderCreationInvalidEasyPayConfig())
		body := []byte(`{"seat_no":1,"payment_type":"alipay","agreement_accepted":true,"return_url":"https://api.pixel-cafe-s170.invalid/payment/result"}`)

		response := cafeOrderCreationRequest(t, fixture.router, fixture.token, fixture.roomID, "s170-provider-failure-key", body)
		require.Equal(t, http.StatusServiceUnavailable, response.Code)
		require.Contains(t, response.Body.String(), "PAYMENT_PROVIDER_MISCONFIGURED")

		ctx := context.Background()
		orders, err := fixture.client.PaymentOrder.Query().Where(paymentorder.OrderTypeEQ(payment.OrderTypeGroupBuy)).All(ctx)
		require.NoError(t, err)
		require.Len(t, orders, 1)
		require.Equal(t, service.OrderStatusFailed, orders[0].Status)

		seats, err := fixture.client.GroupBuySeat.Query().Where(groupbuyseat.OrderIDEQ(orders[0].ID)).All(ctx)
		require.NoError(t, err)
		require.Len(t, seats, 1)
		require.Equal(t, service.GroupBuySeatStatusReleased, seats[0].Status)

		round, err := fixture.client.GroupBuyRound.Get(ctx, fixture.roundID)
		require.NoError(t, err)
		require.Zero(t, round.ReservedSeats)
		require.Zero(t, round.ReservedShares)

		releaseCount, err := fixture.client.GroupBuyEvent.Query().Where(
			groupbuyevent.RoundIDEQ(fixture.roundID),
			groupbuyevent.EventTypeEQ("shares_released"),
		).Count(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, releaseCount)
		auditCount, err := fixture.client.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(orders[0].ID, 10))).Count(ctx)
		require.NoError(t, err)
		require.Zero(t, auditCount)
		keyCount, err := fixture.client.APIKey.Query().Count(ctx)
		require.NoError(t, err)
		require.Zero(t, keyCount)
		bindingCount, err := fixture.client.APIKeyAccountBinding.Query().Count(ctx)
		require.NoError(t, err)
		require.Zero(t, bindingCount)
		require.Equal(t, 1, cafeOrderCreationIdempotencyCount(t, ctx, fixture.sqlDB, "s170-provider-failure-key", service.IdempotencyStatusFailedRetryable))
	})
}

func newCafeOrderCreationFixture(t *testing.T, providerConfig map[string]string) cafeOrderCreationFixture {
	t.Helper()
	ctx := context.Background()
	client, sqlDB := newCafeAuthenticatedHTTPPostgresClient(t)
	ensureCafeOrderCreationIdempotencyTimestampDefaults(t, ctx, sqlDB)
	settings := &cafeAuthenticatedHTTPSettings{enabled: true}

	userEntity, err := client.User.Create().
		SetEmail("cafe-s170-user@example.test").
		SetUsername("cafe-s170-user").
		SetPasswordHash("not-a-real-password").
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().
		SetName("Cafe S170 Group").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType("subscription").
		SetMonthlyLimitUsd(100).
		Save(ctx)
	require.NoError(t, err)
	plan, err := client.GroupBuyPlan.Create().
		SetTitle("Cafe S170 Plan").
		SetProductKey(service.GroupBuyProductTokenPinPinPin).
		SetTotalShares(2).
		SetSeatCount(2).
		SetPricePerShare(12).
		SetPricePerSeat(12).
		SetQuotaPerShareLabel("synthetic quota").
		SetQuotaLabel("synthetic quota").
		SetMaxSharesPerUser(1).
		SetTargetGroupID(group.ID).
		SetTierGroupIds(map[string]int64{"2": group.ID}).
		SetValidityDays(30).
		SetTimeoutMinutes(60).
		SetLaunchMode(service.GroupBuyLaunchModeManual).
		SetRefundMode(service.GroupBuyRefundModeBalanceCredit).
		SetFulfillmentMode(service.CafeRoomFulfillmentMode).
		SetStatus(service.GroupBuyPlanStatusActive).
		Save(ctx)
	require.NoError(t, err)
	account, err := client.Account.Create().
		SetName("cafe-s170-account").
		SetPlatform(service.PlatformOpenAI).
		SetType("api_key").
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	room, err := client.CafeRoom.Create().
		SetCode("CAFE-S170").
		SetName("S170 Claude Room").
		SetPlanID(plan.ID).
		SetAccountID(account.ID).
		SetZoneKey("claude").
		SetThemeKey("warm_wood").
		SetStatus(service.CafeRoomStatusEnabled).
		SetFeatured(true).
		Save(ctx)
	require.NoError(t, err)
	round, err := client.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetCafeRoomID(room.ID).
		SetAssignedAccountID(account.ID).
		SetStatus(service.CafeRoundStatusOpen).
		SetTotalShares(2).
		SetTotalSeats(2).
		SetDeadlineAt(time.Now().UTC().Add(time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	configJSON, err := json.Marshal(providerConfig)
	require.NoError(t, err)
	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeEasyPay).
		SetName("cafe-s170-easypay").
		SetConfig(string(configJSON)).
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		SetPaymentMode("popup").
		Save(ctx)
	require.NoError(t, err)

	settingRepo := repository.NewSettingRepository(client)
	require.NoError(t, settingRepo.SetMultiple(ctx, map[string]string{
		service.SettingPaymentEnabled:                    "true",
		service.SettingOrderTimeoutMinutes:               "30",
		service.SettingMaxPendingOrders:                  "3",
		service.SettingLoadBalanceStrategy:               "round-robin",
		service.SettingPaymentVisibleMethodAlipayEnabled: "true",
		service.SettingPaymentVisibleMethodAlipaySource:  service.VisibleMethodSourceEasyPayAlipay,
	}))
	userRepo := repository.NewUserRepository(client, sqlDB)
	groupRepo := repository.NewGroupRepository(client, sqlDB)
	paymentConfigSvc := service.NewPaymentConfigService(client, settingRepo, nil)
	paymentSvc := service.NewPaymentService(
		client,
		payment.NewRegistry(),
		payment.NewDefaultLoadBalancer(client, nil),
		nil,
		nil,
		paymentConfigSvc,
		userRepo,
		groupRepo,
		nil,
	)
	groupBuySvc := service.NewGroupBuyService(client, paymentSvc, nil, nil, nil, userRepo, groupRepo, nil, nil)
	orderSvc := service.NewCafeRoomOrderService(client, paymentSvc, groupBuySvc, settings)

	userSvc := service.NewUserService(userRepo, nil, nil, nil)
	authSvc := service.NewAuthService(
		client,
		userRepo,
		nil,
		nil,
		&config.Config{JWT: config.JWTConfig{Secret: "s170-local-jwt-secret", AccessTokenExpireMinutes: 30}},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	user, err := userSvc.GetByID(ctx, userEntity.ID)
	require.NoError(t, err)
	token := cafeAuthenticatedHTTPToken(t, authSvc, user)

	previousCoordinator := service.DefaultIdempotencyCoordinator()
	idempotencyConfig := service.DefaultIdempotencyConfig()
	idempotencyConfig.ObserveOnly = false
	service.SetDefaultIdempotencyCoordinator(service.NewIdempotencyCoordinator(
		repository.NewIdempotencyRepository(client, sqlDB),
		idempotencyConfig,
	))
	t.Cleanup(func() { service.SetDefaultIdempotencyCoordinator(previousCoordinator) })

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterUserRoutes(
		v1,
		&handler.Handlers{
			Cafe:  handler.NewCafeHandler(service.NewCafePublicService(client, settings), orderSvc),
			Admin: &handler.AdminHandlers{},
		},
		middleware.NewJWTAuthMiddleware(authSvc, userSvc),
		middleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() }),
		nil,
	)

	return cafeOrderCreationFixture{
		client:             client,
		sqlDB:              sqlDB,
		router:             router,
		token:              token,
		roomID:             room.ID,
		roundID:            round.ID,
		providerInstanceID: strconv.FormatInt(instance.ID, 10),
	}
}

func cafeOrderCreationRequest(t *testing.T, router *gin.Engine, token string, roomID int64, idempotencyKey string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/cafe/rooms/%d/orders", roomID), bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", idempotencyKey)
	request.Host = "api.pixel-cafe-s170.invalid"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func decodeCafeOrderCreationResponse(t *testing.T, response *httptest.ResponseRecorder) cafeOrderCreationResponse {
	t.Helper()
	var decoded cafeOrderCreationResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &decoded))
	return decoded
}

func assertCafeOrderCreationHostedEasyPayURL(t *testing.T, rawURL string) {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	require.Equal(t, "https", parsed.Scheme)
	require.Equal(t, "easypay-s170.invalid", parsed.Host)
	require.Equal(t, "/submit.php", parsed.Path)
	require.NotEmpty(t, parsed.Query().Get("sign"))
	require.Equal(t, "MD5", parsed.Query().Get("sign_type"))
}

func cafeOrderCreationIdempotencyCount(t *testing.T, ctx context.Context, sqlDB *sql.DB, key, status string) int {
	t.Helper()
	var count int
	err := sqlDB.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM idempotency_records
WHERE scope = $1 AND idempotency_key_hash = $2 AND status = $3
`, "cafe_room_order", service.HashIdempotencyKey(key), status).Scan(&count)
	require.NoError(t, err)
	return count
}

// Ent's TimeMixin creates non-null timestamp columns without the database
// defaults that migration 057 supplies for the native idempotency repository.
func ensureCafeOrderCreationIdempotencyTimestampDefaults(t *testing.T, ctx context.Context, sqlDB *sql.DB) {
	t.Helper()
	for _, statement := range []string{
		"ALTER TABLE idempotency_records ALTER COLUMN created_at SET DEFAULT NOW()",
		"ALTER TABLE idempotency_records ALTER COLUMN updated_at SET DEFAULT NOW()",
	} {
		_, err := sqlDB.ExecContext(ctx, statement)
		require.NoError(t, err)
	}
}

func cafeOrderCreationValidEasyPayConfig() map[string]string {
	return map[string]string{
		"pid":       "s170-synthetic-pid",
		"pkey":      cafeOrderCreationSyntheticPKey,
		"apiBase":   "https://easypay-s170.invalid",
		"notifyUrl": "https://pixel-cafe-s170.invalid/payment/notify",
		"returnUrl": "https://api.pixel-cafe-s170.invalid/payment/result",
	}
}

func cafeOrderCreationInvalidEasyPayConfig() map[string]string {
	return map[string]string{
		"pid":       "s170-synthetic-pid",
		"apiBase":   "https://easypay-s170.invalid",
		"notifyUrl": "https://pixel-cafe-s170.invalid/payment/notify",
		"returnUrl": "https://api.pixel-cafe-s170.invalid/payment/result",
	}
}

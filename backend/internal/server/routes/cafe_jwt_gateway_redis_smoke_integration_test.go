//go:build integration

package routes

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const cafeJWTGatewayRedisSmokeKey = "sk-cafe-s173-isolated-0000000000000001"

type cafeJWTGatewayRedisSmokeFixture struct {
	cafe      cafeAuthenticatedHTTPFixture
	userID    int64
	keyID     int64
	bindingID int64
	accountID int64
	keyName   string
}

func TestCafeJWTGatewayRedisSmokeIntegration(t *testing.T) {
	fixture := newCafeJWTGatewayRedisSmokeFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	assertCafeJWTGatewayMyRoomsBoundary(t, fixture)

	redisA, redisB := newCafeJWTGatewayRedisSmokeClients(t, ctx)
	cacheConfig := &config.Config{
		RunMode: config.RunModeSimple,
		APIKeyAuth: config.APIKeyAuthCacheConfig{
			L1Size:             128,
			L1TTLSeconds:       60,
			L2TTLSeconds:       60,
			NegativeTTLSeconds: 5,
			Singleflight:       true,
		},
	}
	apiKeyRepo := repository.NewAPIKeyRepository(fixture.cafe.client, cafeJWTGatewayRedisSQLDB(t, fixture.cafe.client))
	instanceA := service.NewAPIKeyService(apiKeyRepo, nil, nil, nil, nil, repository.NewAPIKeyCache(redisA), cacheConfig)
	instanceB := service.NewAPIKeyService(apiKeyRepo, nil, nil, nil, nil, repository.NewAPIKeyCache(redisB), cacheConfig)
	subscribeCtx, stopSubscriber := context.WithCancel(ctx)
	t.Cleanup(stopSubscriber)
	instanceB.StartAuthCacheInvalidationSubscriber(subscribeCtx)

	var terminalCalls atomic.Int32
	fixture.cafe.router.GET(
		"/v1/cafe-s173-preflight",
		gatewayAuthMiddleware(middleware.NewAPIKeyAuthMiddleware(instanceB, nil, cacheConfig), instanceB, nil),
		func(c *gin.Context) {
			apiKey, ok := middleware.GetAPIKeyFromContext(c)
			if !ok || apiKey == nil || apiKey.PinnedAccountID != fixture.accountID {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "missing expected cafe pin"})
				return
			}
			terminalCalls.Add(1)
			c.Header("X-S173-Key-Name", apiKey.Name)
			c.Header("X-S173-Pinned-Account", strconv.FormatInt(apiKey.PinnedAccountID, 10))
			c.Status(http.StatusNoContent)
		},
	)

	initial := cafeJWTGatewayPreflight(t, fixture.cafe.router)
	require.Equal(t, http.StatusNoContent, initial.Code)
	require.Equal(t, fixture.keyName, initial.Header().Get("X-S173-Key-Name"))
	require.Equal(t, strconv.FormatInt(fixture.accountID, 10), initial.Header().Get("X-S173-Pinned-Account"))
	require.EqualValues(t, 1, terminalCalls.Load())

	// The L2 record is removed without a Pub/Sub message. The old name can only
	// remain visible when instance B has actually retained its L1 auth snapshot.
	time.Sleep(200 * time.Millisecond)
	updatedName := "S173 stale L1 Cafe key"
	_, err := fixture.cafe.client.APIKey.UpdateOneID(fixture.keyID).SetName(updatedName).Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repository.NewAPIKeyCache(redisB).DeleteAuthCache(ctx, cafeJWTGatewayRedisAuthCacheKey(cafeJWTGatewayRedisSmokeKey)))
	staleL1 := cafeJWTGatewayPreflight(t, fixture.cafe.router)
	require.Equal(t, http.StatusNoContent, staleL1.Code)
	require.Equal(t, fixture.keyName, staleL1.Header().Get("X-S173-Key-Name"))

	inactive := "inactive"
	_, err = instanceA.Update(ctx, fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Status: &inactive})
	require.NoError(t, err)
	terminalBeforeInactive := terminalCalls.Load()
	require.Eventually(t, func() bool {
		response := cafeJWTGatewayPreflight(t, fixture.cafe.router)
		return response.Code == http.StatusUnauthorized && response.Body.String() != "" &&
			containsAll(response.Body.String(), "API_KEY_DISABLED")
	}, 8*time.Second, 40*time.Millisecond, "cross-instance invalidation must reject the stale active Key")
	require.Equal(t, terminalBeforeInactive, terminalCalls.Load(), "inactive Key must not reach the gateway preflight terminal")

	active := service.StatusAPIKeyActive
	_, err = instanceA.Update(ctx, fixture.keyID, fixture.userID, service.UpdateAPIKeyRequest{Status: &active})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		response := cafeJWTGatewayPreflight(t, fixture.cafe.router)
		return response.Code == http.StatusNoContent &&
			response.Header().Get("X-S173-Key-Name") == updatedName &&
			response.Header().Get("X-S173-Pinned-Account") == strconv.FormatInt(fixture.accountID, 10)
	}, 8*time.Second, 40*time.Millisecond, "valid re-enable must reach the preflight with its original fixed Account")

	shortBindingName := "S173 short binding Cafe key"
	expiresSoon := time.Now().UTC().Add(3 * time.Second)
	_, err = fixture.cafe.client.APIKeyAccountBinding.UpdateOneID(fixture.bindingID).SetExpiresAt(expiresSoon).Save(ctx)
	require.NoError(t, err)
	_, err = fixture.cafe.client.APIKey.UpdateOneID(fixture.keyID).SetName(shortBindingName).Save(ctx)
	require.NoError(t, err)
	instanceA.InvalidateAuthCacheByKey(ctx, cafeJWTGatewayRedisSmokeKey)
	require.Eventually(t, func() bool {
		response := cafeJWTGatewayPreflight(t, fixture.cafe.router)
		return response.Code == http.StatusNoContent && response.Header().Get("X-S173-Key-Name") == shortBindingName
	}, 8*time.Second, 40*time.Millisecond, "B must refresh after the binding-change invalidation before testing TTL expiry")

	require.Eventually(t, func() bool {
		response := cafeJWTGatewayPreflight(t, fixture.cafe.router)
		return response.Code == http.StatusForbidden && containsAll(response.Body.String(), "CAFE_ACCOUNT_UNAVAILABLE")
	}, 10*time.Second, 80*time.Millisecond, "binding expiry must not be extended by B's L1 or Redis L2 cache")
	terminalBeforeRejectedReplay := terminalCalls.Load()
	rejectedReplay := cafeJWTGatewayPreflight(t, fixture.cafe.router)
	require.Equal(t, http.StatusForbidden, rejectedReplay.Code)
	require.Contains(t, rejectedReplay.Body.String(), "CAFE_ACCOUNT_UNAVAILABLE")
	require.Equal(t, terminalBeforeRejectedReplay, terminalCalls.Load(), "expired Binding must not reach the gateway preflight terminal")
}

func newCafeJWTGatewayRedisSmokeFixture(t *testing.T) cafeJWTGatewayRedisSmokeFixture {
	t.Helper()
	ctx := context.Background()
	cafe := newCafeAuthenticatedHTTPFixture(t)

	room, err := cafe.client.CafeRoom.Get(ctx, cafe.roomID)
	require.NoError(t, err)
	plan, err := cafe.client.GroupBuyPlan.Get(ctx, room.PlanID)
	require.NoError(t, err)
	groupID := plan.TargetGroupID
	now := time.Now().UTC().Truncate(time.Second)
	expiresAt := now.Add(time.Hour)

	_, err = cafe.client.Group.UpdateOneID(groupID).SetAccessMode(service.CafeRoomGroupAccessMode).Save(ctx)
	require.NoError(t, err)
	account, err := cafe.client.Account.Create().
		SetName("S173 temporary Cafe account").
		SetPlatform(service.PlatformOpenAI).
		SetType("api_key").
		SetStatus(service.StatusActive).
		AddGroupIDs(groupID).
		Save(ctx)
	require.NoError(t, err)
	_, err = cafe.client.CafeRoom.UpdateOneID(cafe.roomID).SetAccountID(account.ID).Save(ctx)
	require.NoError(t, err)

	seat, err := cafe.client.GroupBuySeat.Get(ctx, cafe.userASeatID)
	require.NoError(t, err)
	_, err = cafe.client.GroupBuyRound.UpdateOneID(seat.RoundID).
		SetAssignedAccountID(account.ID).
		SetEntitlementExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	_, err = cafe.client.GroupBuySeat.UpdateOneID(seat.ID).SetExpiresAt(expiresAt).Save(ctx)
	require.NoError(t, err)

	keyName := "S173 active Cafe key"
	key, err := cafe.client.APIKey.Create().
		SetUserID(cafe.userA.ID).
		SetKey(cafeJWTGatewayRedisSmokeKey).
		SetName(keyName).
		SetGroupID(groupID).
		SetStatus(service.StatusAPIKeyActive).
		SetExpiresAt(expiresAt).
		SetManagedSourceType(service.APIKeyManagedSourceCafeRoomSeat).
		SetManagedSourceID(seat.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = cafe.client.GroupBuySeat.UpdateOneID(seat.ID).SetBoundAPIKeyID(key.ID).SetBoundAt(now).Save(ctx)
	require.NoError(t, err)
	binding, err := cafe.client.APIKeyAccountBinding.Create().
		SetAPIKeyID(key.ID).
		SetUserID(cafe.userA.ID).
		SetGroupID(groupID).
		SetAccountID(account.ID).
		SetCafeRoomID(cafe.roomID).
		SetRoundID(seat.RoundID).
		SetSeatID(seat.ID).
		SetStatus("active").
		SetStrictMode(true).
		SetStartsAt(now).
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)

	return cafeJWTGatewayRedisSmokeFixture{
		cafe:      cafe,
		userID:    cafe.userA.ID,
		keyID:     key.ID,
		bindingID: binding.ID,
		accountID: account.ID,
		keyName:   keyName,
	}
}

func assertCafeJWTGatewayMyRoomsBoundary(t *testing.T, fixture cafeJWTGatewayRedisSmokeFixture) {
	t.Helper()
	tokenA := cafeAuthenticatedHTTPToken(t, fixture.cafe.authService, fixture.cafe.userA)
	tokenB := cafeAuthenticatedHTTPToken(t, fixture.cafe.authService, fixture.cafe.userB)

	userAResponse := cafeAuthenticatedHTTPRequest(t, fixture.cafe.router, http.MethodGet, "/api/v1/cafe/my-rooms", tokenA, nil)
	require.Equal(t, http.StatusOK, userAResponse.Code)
	require.Contains(t, userAResponse.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.cafe.userASeatID))
	require.NotContains(t, userAResponse.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.cafe.userBSeatID))

	userBResponse := cafeAuthenticatedHTTPRequest(t, fixture.cafe.router, http.MethodGet, "/api/v1/cafe/my-rooms", tokenB, nil)
	require.Equal(t, http.StatusOK, userBResponse.Code)
	require.Contains(t, userBResponse.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.cafe.userBSeatID))
	require.NotContains(t, userBResponse.Body.String(), fmt.Sprintf("\"membership_id\":%d", fixture.cafe.userASeatID))

	anonymousResponse := cafeAuthenticatedHTTPRequest(t, fixture.cafe.router, http.MethodGet, "/api/v1/cafe/my-rooms", "", nil)
	require.Equal(t, http.StatusUnauthorized, anonymousResponse.Code)
	require.Contains(t, anonymousResponse.Body.String(), "UNAUTHORIZED")
}

func newCafeJWTGatewayRedisSmokeClients(t *testing.T, ctx context.Context) (*redisclient.Client, *redisclient.Client) {
	t.Helper()
	container, err := tcredis.Run(ctx, "redis:8.4-alpine")
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)
	options := &redisclient.Options{Addr: fmt.Sprintf("%s:%s", host, port.Port())}
	redisA := redisclient.NewClient(options)
	redisB := redisclient.NewClient(options)
	t.Cleanup(func() { _ = redisB.Close() })
	t.Cleanup(func() { _ = redisA.Close() })
	require.NoError(t, redisA.Ping(ctx).Err())
	require.NoError(t, redisB.Ping(ctx).Err())
	return redisA, redisB
}

func cafeJWTGatewayPreflight(t *testing.T, router *gin.Engine) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/v1/cafe-s173-preflight", nil)
	request.Header.Set("Authorization", "Bearer "+cafeJWTGatewayRedisSmokeKey)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func cafeJWTGatewayRedisAuthCacheKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func cafeJWTGatewayRedisSQLDB(t *testing.T, client *dbent.Client) *sql.DB {
	t.Helper()
	driver, ok := client.Driver().(*entsql.Driver)
	require.True(t, ok, "the disposable Cafe Ent client must use the PostgreSQL SQL driver")
	return driver.DB()
}

func containsAll(value string, tokens ...string) bool {
	for _, token := range tokens {
		if !strings.Contains(value, token) {
			return false
		}
	}
	return true
}

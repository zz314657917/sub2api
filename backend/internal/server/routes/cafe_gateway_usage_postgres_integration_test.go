//go:build integration

package routes

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const cafeGatewayUsageIntegrationKey = "sk-cafe-s174-isolated-0000000000000001"

type cafeGatewayUsageUpstream struct {
	calls atomic.Int64

	mu          sync.Mutex
	apiKeys     []string
	requestURIs []string
}

func (u *cafeGatewayUsageUpstream) record(r *http.Request) {
	if u == nil || r == nil {
		return
	}
	u.calls.Add(1)
	u.mu.Lock()
	defer u.mu.Unlock()
	u.apiKeys = append(u.apiKeys, r.Header.Get("x-api-key"))
	u.requestURIs = append(u.requestURIs, r.URL.RequestURI())
}

func (u *cafeGatewayUsageUpstream) snapshot() (apiKeys []string, requestURIs []string) {
	if u == nil {
		return nil, nil
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	return append([]string(nil), u.apiKeys...), append([]string(nil), u.requestURIs...)
}

func TestCafeGatewayUsagePostgresIntegration(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeJWTGatewayRedisSmokeFixture(t)
	sqlDB := cafeJWTGatewayRedisSQLDB(t, fixture.cafe.client)
	ensureCafeGatewayUsageLogSchema(t, ctx, sqlDB)

	upstreamState := &cafeGatewayUsageUpstream{}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamState.record(r)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", "s174-loopback-request")
		_, _ = w.Write([]byte(`{"id":"msg_s174","type":"message","role":"assistant","model":"claude-s174","content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":5,"output_tokens":3}}`))
	}))
	t.Cleanup(upstream.Close)

	pinnedAccountID, keyID, bindingID := setupCafeGatewayUsageFixture(t, fixture, upstream.URL)
	pinnedAccount, err := repository.NewAccountRepository(fixture.cafe.client, sqlDB, nil).GetByID(ctx, pinnedAccountID)
	require.NoError(t, err)
	require.True(t, pinnedAccount.IsSchedulable())
	require.Equal(t, service.PlatformAnthropic, pinnedAccount.Platform)
	require.Equal(t, service.AccountTypeAPIKey, pinnedAccount.Type)
	require.Equal(t, "s174-pinned-upstream-key", pinnedAccount.GetCredential("api_key"))
	require.Equal(t, upstream.URL, pinnedAccount.GetBaseURL())
	require.True(t, pinnedAccount.IsAnthropicAPIKeyPassthroughEnabled())
	router, apiKeyService := newCafeGatewayUsageRouter(t, fixture, sqlDB)

	response := cafeGatewayUsageRequest(router)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "msg_s174")
	require.EqualValues(t, 1, upstreamState.calls.Load())
	apiKeys, requestURIs := upstreamState.snapshot()
	require.Equal(t, []string{"s174-pinned-upstream-key"}, apiKeys)
	require.Equal(t, []string{"/v1/messages?beta=true"}, requestURIs)

	accountIDs := cafeGatewayUsageLogAccountIDs(t, ctx, sqlDB, keyID)
	require.Equal(t, []int64{pinnedAccountID}, accountIDs)

	upstreamCallsBeforeBindingExpiry := upstreamState.calls.Load()
	usageCountBeforeBindingExpiry := len(accountIDs)
	_, err = fixture.cafe.client.APIKeyAccountBinding.UpdateOneID(bindingID).
		SetExpiresAt(time.Now().UTC().Add(-time.Second)).
		Save(ctx)
	require.NoError(t, err)
	apiKeyService.InvalidateAuthCacheByKey(ctx, cafeGatewayUsageIntegrationKey)

	expiredResponse := cafeGatewayUsageRequest(router)
	require.Equal(t, http.StatusForbidden, expiredResponse.Code, expiredResponse.Body.String())
	require.Contains(t, expiredResponse.Body.String(), "CAFE_ACCOUNT_UNAVAILABLE")
	require.Equal(t, upstreamCallsBeforeBindingExpiry, upstreamState.calls.Load())
	require.Len(t, cafeGatewayUsageLogAccountIDs(t, ctx, sqlDB, keyID), usageCountBeforeBindingExpiry)

	_, err = fixture.cafe.client.APIKeyAccountBinding.UpdateOneID(bindingID).
		SetExpiresAt(time.Now().UTC().Add(time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	_, err = fixture.cafe.client.Account.UpdateOneID(pinnedAccountID).
		SetStatus(service.StatusDisabled).
		Save(ctx)
	require.NoError(t, err)
	apiKeyService.InvalidateAuthCacheByKey(ctx, cafeGatewayUsageIntegrationKey)

	unavailableResponse := cafeGatewayUsageRequest(router)
	require.Equal(t, http.StatusForbidden, unavailableResponse.Code, unavailableResponse.Body.String())
	require.Contains(t, unavailableResponse.Body.String(), "CAFE_ACCOUNT_UNAVAILABLE")
	require.Equal(t, upstreamCallsBeforeBindingExpiry, upstreamState.calls.Load())
	require.Len(t, cafeGatewayUsageLogAccountIDs(t, ctx, sqlDB, keyID), usageCountBeforeBindingExpiry)
}

func setupCafeGatewayUsageFixture(t *testing.T, fixture cafeJWTGatewayRedisSmokeFixture, upstreamURL string) (int64, int64, int64) {
	t.Helper()
	ctx := context.Background()

	room, err := fixture.cafe.client.CafeRoom.Get(ctx, fixture.cafe.roomID)
	require.NoError(t, err)
	plan, err := fixture.cafe.client.GroupBuyPlan.Get(ctx, room.PlanID)
	require.NoError(t, err)
	groupID := plan.TargetGroupID

	_, err = fixture.cafe.client.Group.UpdateOneID(groupID).
		SetPlatform(service.PlatformAnthropic).
		SetAllowMessagesDispatch(true).
		SetModelMatchPatterns([]string{"claude-s174"}).
		Save(ctx)
	require.NoError(t, err)
	_, err = fixture.cafe.client.User.UpdateOneID(fixture.userID).SetConcurrency(0).Save(ctx)
	require.NoError(t, err)

	pinned, err := fixture.cafe.client.Account.UpdateOneID(fixture.accountID).
		SetPlatform(service.PlatformAnthropic).
		SetType(service.AccountTypeAPIKey).
		SetSchedulable(true).
		SetConcurrency(0).
		SetPriority(100).
		SetCredentials(map[string]any{
			"api_key":  "s174-pinned-upstream-key",
			"base_url": upstreamURL,
		}).
		SetExtra(map[string]any{"anthropic_passthrough": true}).
		Save(ctx)
	require.NoError(t, err)

	_, err = fixture.cafe.client.Account.Create().
		SetName("S174 lower-priority fallback account").
		SetPlatform(service.PlatformAnthropic).
		SetType(service.AccountTypeAPIKey).
		SetStatus(service.StatusActive).
		SetSchedulable(true).
		SetConcurrency(0).
		SetPriority(0).
		SetCredentials(map[string]any{
			"api_key":  "s174-wrong-upstream-key",
			"base_url": upstreamURL,
		}).
		SetExtra(map[string]any{"anthropic_passthrough": true}).
		AddGroupIDs(groupID).
		Save(ctx)
	require.NoError(t, err)

	key, err := fixture.cafe.client.APIKey.UpdateOneID(fixture.keyID).
		SetKey(cafeGatewayUsageIntegrationKey).
		SetStatus(service.StatusAPIKeyActive).
		Save(ctx)
	require.NoError(t, err)
	binding, err := fixture.cafe.client.APIKeyAccountBinding.Get(ctx, fixture.bindingID)
	require.NoError(t, err)
	require.Equal(t, pinned.ID, binding.AccountID)

	return pinned.ID, key.ID, binding.ID
}

func newCafeGatewayUsageRouter(t *testing.T, fixture cafeJWTGatewayRedisSmokeFixture, sqlDB *sql.DB) (*gin.Engine, *service.APIKeyService) {
	t.Helper()
	cfg := &config.Config{
		RunMode: config.RunModeSimple,
		Gateway: config.GatewayConfig{MaxBodySize: 1024 * 1024},
		Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
			AllowInsecureHTTP: true,
		}},
	}
	apiKeyRepo := repository.NewAPIKeyRepository(fixture.cafe.client, sqlDB)
	accountRepo := repository.NewAccountRepository(fixture.cafe.client, sqlDB, nil)
	groupRepo := repository.NewGroupRepository(fixture.cafe.client, sqlDB)
	usageLogRepo := repository.NewUsageLogRepository(fixture.cafe.client, sqlDB)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, nil, nil, nil, nil, nil, cfg)
	concurrencyService := service.NewConcurrencyService(nil)
	gatewayService := service.NewGatewayService(
		accountRepo,
		groupRepo,
		usageLogRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		cfg,
		nil,
		concurrencyService,
		service.NewBillingService(cfg, nil),
		nil,
		nil,
		nil,
		repository.NewHTTPUpstream(cfg),
		&service.DeferredService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterGatewayRoutes(
		router,
		&handler.Handlers{
			Gateway: handler.NewGatewayHandler(
				gatewayService,
				nil,
				nil,
				nil,
				concurrencyService,
				nil,
				nil,
				nil,
				apiKeyService,
				nil,
				nil,
				nil,
				nil,
				cfg,
				nil,
			),
			OpenAIGateway: &handler.OpenAIGatewayHandler{},
			AsyncImage:    &handler.AsyncImageHandler{},
		},
		middleware.NewAPIKeyAuthMiddleware(apiKeyService, nil, cfg),
		apiKeyService,
		nil,
		nil,
		nil,
		cfg,
	)
	return router, apiKeyService
}

func cafeGatewayUsageRequest(router *gin.Engine) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"model":"claude-s174","max_tokens":64,"stream":false,"messages":[{"role":"user","content":"hello"}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+cafeGatewayUsageIntegrationKey)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func cafeGatewayUsageLogAccountIDs(t *testing.T, ctx context.Context, sqlDB *sql.DB, keyID int64) []int64 {
	t.Helper()
	rows, err := sqlDB.QueryContext(ctx, `SELECT account_id FROM usage_logs WHERE api_key_id = $1 ORDER BY id`, keyID)
	require.NoError(t, err)
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())
	return ids
}

// Ent does not model several current usage migrations or the native daily rollup
// table, while UsageLogRepository's INSERT expects that production schema shape.
func ensureCafeGatewayUsageLogSchema(t *testing.T, ctx context.Context, sqlDB *sql.DB) {
	t.Helper()
	for _, statement := range []string{
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS reasoning_effort VARCHAR(20)",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS media_type VARCHAR(16)",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS openai_ws_mode BOOLEAN NOT NULL DEFAULT FALSE",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS request_type SMALLINT NOT NULL DEFAULT 0",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS service_tier VARCHAR(16)",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS inbound_endpoint VARCHAR(128)",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS upstream_endpoint VARCHAR(128)",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS account_stats_cost NUMERIC(20, 10)",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS image_output_tokens INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS image_output_cost DECIMAL(20, 10) NOT NULL DEFAULT 0",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS image_input_tokens INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS image_input_cost DECIMAL(20, 10) NOT NULL DEFAULT 0",
		"ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS session_id VARCHAR(255)",
		`CREATE TABLE IF NOT EXISTS user_usage_daily_stats (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			usage_date DATE NOT NULL,
			requests BIGINT NOT NULL DEFAULT 0,
			input_tokens BIGINT NOT NULL DEFAULT 0,
			output_tokens BIGINT NOT NULL DEFAULT 0,
			cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
			cache_read_tokens BIGINT NOT NULL DEFAULT 0,
			tokens BIGINT NOT NULL DEFAULT 0,
			actual_cost NUMERIC(20, 10) NOT NULL DEFAULT 0,
			night_requests BIGINT NOT NULL DEFAULT 0,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (user_id, usage_date)
		)`,
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_logs_request_id_api_key_unique ON usage_logs (request_id, api_key_id)",
	} {
		_, err := sqlDB.ExecContext(ctx, statement)
		require.NoError(t, err)
	}
}

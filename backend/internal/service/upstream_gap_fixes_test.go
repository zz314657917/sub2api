package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/dgraph-io/ristretto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGapFixNonStreamImageCount(t *testing.T) {
	for _, tc := range []struct {
		name, reply string
		want        int
	}{
		{"empty", `{"data":[]}`, 1},
		{"one", `{"data":[{"b64_json":"aGVsbG8="}]}`, 1},
		{"two", `{"data":[{"b64_json":"aGVsbG8="},{"b64_json":"d29ybGQ="}]}`, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-1","prompt":"fixture","n":4}`)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}},
				Body: io.NopCloser(strings.NewReader(tc.reply)),
			}}}
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			result, err := svc.ForwardImages(context.Background(), c, &Account{ID: 6, Platform: PlatformOpenAI,
				Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture", "base_url": "https://fixture.invalid/v1"}}, body, parsed, "")
			require.NoError(t, err)
			require.Equal(t, tc.want, result.ImageCount, "empty response keeps the local single-image fallback, never request n")
		})
	}
}

func TestGapFixQuotaPauseReset(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	for _, window := range []string{"5h", "7d"} {
		for _, tc := range []struct {
			name, updated, absolute string
			relative                any
			want                    bool
		}{
			{"absolute_future", now.Add(-3 * time.Hour).Format(time.RFC3339), now.Add(time.Hour).Format(time.RFC3339), nil, true},
			{"relative_future", now.Add(-3 * time.Hour).Format(time.RFC3339), "", 4 * 3600, true},
			{"absolute_wins_future", now.Add(-3 * time.Hour).Format(time.RFC3339), now.Add(time.Hour).Format(time.RFC3339), 1, true},
			{"absolute_wins_past", now.Add(-3 * time.Hour).Format(time.RFC3339), now.Add(-time.Second).Format(time.RFC3339), 10 * 3600, false},
			{"absolute_equal", now.Add(-3 * time.Hour).Format(time.RFC3339), now.Format(time.RFC3339), nil, false},
			{"relative_equal", now.Add(-3 * time.Hour).Format(time.RFC3339), "", 3 * 3600, false},
			{"relative_past", now.Add(-3 * time.Hour).Format(time.RFC3339), "", 2 * 3600, false},
			{"invalid_absolute_relative_future", now.Add(-3 * time.Hour).Format(time.RFC3339), "bad", 4 * 3600, true},
			{"missing_reset", now.Add(-3 * time.Hour).Format(time.RFC3339), "", nil, false},
			{"invalid_reset", now.Add(-3 * time.Hour).Format(time.RFC3339), "bad", "bad", false},
			{"negative_reset", now.Add(-3 * time.Hour).Format(time.RFC3339), "", -1, false},
			{"missing_update", "", "", nil, true},
			{"invalid_update", "bad", "", nil, true},
			{"missing_update_relative", "", "", 3600, true},
			{"fresh_missing_reset", now.Add(-time.Minute).Format(time.RFC3339), "", nil, true},
		} {
			t.Run(window+"/"+tc.name, func(t *testing.T) {
				extra := map[string]any{"codex_" + window + "_used_percent": 99.0}
				if tc.updated != "" {
					extra["codex_usage_updated_at"] = tc.updated
				}
				if tc.absolute != "" {
					extra["codex_"+window+"_reset_at"] = tc.absolute
				}
				if tc.relative != nil {
					extra["codex_"+window+"_reset_after_seconds"] = tc.relative
				}
				ratio, ok := resolveOpenAIQuotaUtilization(extra, window, now)
				require.Equal(t, tc.want, ok)
				if ok {
					require.InDelta(t, .99, ratio, 1e-12)
				}
				if tc.name == "relative_future" {
					_, stillPaused := resolveOpenAIQuotaUtilization(extra, window, now.Add(time.Hour))
					require.False(t, stillPaused, "relative reset must keep the original update-time anchor")
				}
			})
		}
	}
}

func TestGapFixAccountStatsGatesAndPrecedence(t *testing.T) {
	for _, resolver := range []bool{false, true} {
		for _, accountOn := range []bool{false, true} {
			for _, groupOn := range []bool{false, true} {
				t.Run(fmt.Sprintf("resolver=%t/account=%t/group=%t", resolver, accountOn, groupOn), func(t *testing.T) {
					f := newGPT61SubscriptionFixture(resolver)
					f.input.Account.Type = AccountTypeAPIKey
					f.input.Account.Extra["openai_long_context_billing_enabled"] = accountOn
					f.input.APIKey.Group.LongContextPricingEnabled = groupOn
					f.input.Result.Model = "gpt-5.4-2026-03-05"
					f.input.Result.Usage = OpenAIUsage{InputTokens: 300000, OutputTokens: 2000}
					require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
					stats, charge := .78, .78
					if accountOn {
						stats = 1.545
						// Legacy callers without a resolver have always used the
						// account gate alone for customer billing.
						if groupOn || !resolver {
							charge = 1.545
						}
					}
					require.InDelta(t, stats, *f.usage.lastLog.AccountStatsCost, 1e-12)
					require.InDelta(t, charge, f.usage.lastLog.TotalCost, 1e-12)
				})
			}
		}
	}
	for _, tc := range []struct {
		name      string
		oauth     bool
		model     string
		custom    bool
		apply     bool
		wantStats float64
	}{
		{"oauth_sol_exemption", true, "gpt-6.1-sol", false, false, .62},
		{"oauth_other_model", true, "gpt-5.4-2026-03-05", false, false, 1.545},
		{"custom_before_model", false, "gpt-5.4-2026-03-05", true, false, 7},
		{"custom_before_apply", false, "gpt-5.4-2026-03-05", true, true, 7},
		{"apply_before_model", false, "gpt-5.4-2026-03-05", false, true, .78},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newGPT61SubscriptionFixture(true)
			f.input.Account.Type = AccountTypeAPIKey
			f.input.Account.Extra["openai_long_context_billing_enabled"] = true
			if tc.oauth {
				f.input.Account.Type = AccountTypeOAuth
			}
			f.input.APIKey.Group.LongContextPricingEnabled = false
			f.input.Result.Model = tc.model
			f.input.Result.Usage = OpenAIUsage{InputTokens: 300000, OutputTokens: 2000}
			channel := &Channel{ID: 61, ApplyPricingToAccountStats: tc.apply}
			if tc.custom {
				price := 7.0
				channel.AccountStatsPricingRules = []AccountStatsPricingRule{{AccountIDs: []int64{31},
					Pricing: []ChannelModelPricing{{Models: []string{tc.model}, BillingMode: BillingModePerRequest, PerRequestPrice: &price}}}}
			}
			f.svc.channelService = gpt61SubscriptionChannel(channel)
			f.svc.resolver = NewModelPricingResolver(f.svc.channelService, f.svc.billingService)
			require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
			require.InDelta(t, tc.wantStats, *f.usage.lastLog.AccountStatsCost, 1e-12)
		})
	}
}

// The fixture uses the real ent transaction and SQLmock SQL writes. Its row
// lock is held until the transaction commits/rolls back, modelling PostgreSQL
// serialization without a live database or a production application mutex.
type gapSubscriptionRepo struct {
	userSubRepoNoop
	t         *testing.T
	mu        sync.Mutex
	current   UserSubscription
	cache     *ristretto.Cache
	committed bool
	updateErr error
	statusErr error
	readErr   error
	skipCache bool
}

func (r *gapSubscriptionRepo) GetByIDForUpdate(ctx context.Context, id int64) (*UserSubscription, error) {
	tx := dbent.TxFromContext(ctx)
	require.NotNil(r.t, tx, "admin read must carry the real transaction context")
	r.mu.Lock()
	r.committed = false
	tx.OnCommit(func(next dbent.Committer) dbent.Committer {
		return dbent.CommitFunc(func(ctx context.Context, tx *dbent.Tx) error {
			if !r.skipCache {
				_, ok := r.cache.Get(subCacheKey(r.current.UserID, r.current.GroupID))
				require.True(r.t, ok, "cache must remain valid until commit succeeds")
			}
			err := next.Commit(ctx, tx)
			r.committed = err == nil
			r.mu.Unlock()
			return err
		})
	})
	tx.OnRollback(func(next dbent.Rollbacker) dbent.Rollbacker {
		return dbent.RollbackFunc(func(ctx context.Context, tx *dbent.Tx) error {
			defer r.mu.Unlock()
			return next.Rollback(ctx, tx)
		})
	})
	if r.readErr != nil {
		return nil, r.readErr
	}
	row, err := tx.Client().UserSubscription.Query().Where(usersubscription.IDEQ(id)).
		Select(usersubscription.FieldID, usersubscription.FieldExpiresAt).
		ForUpdate().Only(ctx)
	if err != nil {
		return nil, err
	}
	copy := r.current
	copy.ExpiresAt = row.ExpiresAt
	return &copy, nil
}

func (r *gapSubscriptionRepo) GetByID(ctx context.Context, _ int64) (*UserSubscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.committed {
		return nil, errors.New("unlocked read before transaction commit")
	}
	require.Nil(r.t, dbent.TxFromContext(ctx), "final reread must be outside transaction")
	if !r.skipCache {
		_, ok := r.cache.Get(subCacheKey(r.current.UserID, r.current.GroupID))
		require.False(r.t, ok, "cache must be invalidated before final reread")
	}
	copy := r.current
	return &copy, nil
}

func (r *gapSubscriptionRepo) ExtendExpiry(ctx context.Context, id int64, expiry time.Time) error {
	tx := dbent.TxFromContext(ctx)
	require.NotNil(r.t, tx)
	if r.updateErr != nil {
		return r.updateErr
	}
	if err := tx.Client().UserSubscription.Update().Where(usersubscription.IDEQ(id)).SetExpiresAt(expiry).Exec(ctx); err != nil {
		return err
	}
	// Publish the fixture state only on a successful commit.
	tx.OnCommit(func(next dbent.Committer) dbent.Committer {
		return dbent.CommitFunc(func(ctx context.Context, tx *dbent.Tx) error {
			err := next.Commit(ctx, tx)
			if err == nil {
				r.current.ExpiresAt = expiry
			}
			return err
		})
	})
	return nil
}

func (r *gapSubscriptionRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	tx := dbent.TxFromContext(ctx)
	require.NotNil(r.t, tx)
	if r.statusErr != nil {
		return r.statusErr
	}
	if err := tx.Client().UserSubscription.Update().Where(usersubscription.IDEQ(id)).SetStatus(status).Exec(ctx); err != nil {
		return err
	}
	tx.OnCommit(func(next dbent.Committer) dbent.Committer {
		return dbent.CommitFunc(func(ctx context.Context, tx *dbent.Tx) error {
			err := next.Commit(ctx, tx)
			if err == nil {
				r.current.Status = status
			}
			return err
		})
	})
	return nil
}

func gapSubscriptionFixture(t *testing.T) (*SubscriptionService, *gapSubscriptionRepo, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	cache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 100, MaxCost: 1 << 20, BufferItems: 64})
	require.NoError(t, err)
	t.Cleanup(cache.Close)
	repo := &gapSubscriptionRepo{t: t, cache: cache, current: UserSubscription{
		ID: 7, UserID: 11, GroupID: 13, ExpiresAt: time.Now().UTC().AddDate(0, 0, 10), Status: SubscriptionStatusActive,
	}}
	require.True(t, cache.Set(subCacheKey(11, 13), "cached", 1))
	cache.Wait()
	_, cached := cache.Get(subCacheKey(11, 13))
	require.True(t, cached)
	svc := NewSubscriptionService(nil, repo, nil, client, nil)
	svc.subCacheL1 = cache
	return svc, repo, mock
}

func expectGapSubscriptionRead(mock sqlmock.Sqlmock, expiry time.Time) {
	mock.ExpectQuery(`(?s)SELECT .* FROM "user_subscriptions".*FOR UPDATE`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at"}).AddRow(int64(7), expiry))
}

func TestGapFixAdminExtendTransaction(t *testing.T) {
	for _, mode := range []string{"success", "read_error", "expiry_error", "status_error", "commit_error", "expired_negative", "would_expire", "upper_bound"} {
		t.Run(mode, func(t *testing.T) {
			svc, repo, mock := gapSubscriptionFixture(t)
			initial := repo.current.ExpiresAt
			days := 5
			injected := errors.New("fixture failure")
			mock.ExpectBegin()
			switch mode {
			case "read_error":
				repo.readErr = injected
			case "expiry_error":
				repo.updateErr = injected
			case "status_error":
				repo.statusErr = injected
			case "expired_negative":
				repo.current.ExpiresAt, days = time.Now().Add(-time.Hour), -5
			case "would_expire":
				days = -20
			case "upper_bound":
				days = MaxValidityDays + 1
			}
			repo.current.Status = SubscriptionStatusExpired
			initial = repo.current.ExpiresAt
			if mode != "read_error" {
				expectGapSubscriptionRead(mock, initial)
			}
			writes := mode == "success" || mode == "status_error" || mode == "commit_error" || mode == "upper_bound"
			expectedExpiry := initial.AddDate(0, 0, days)
			if mode == "upper_bound" {
				expectedExpiry = MaxExpiresAt
			}
			if writes {
				mock.ExpectExec(`UPDATE "user_subscriptions" SET "updated_at" = \$1, "expires_at" = \$2.*`).
					WithArgs(sqlmock.AnyArg(), expectedExpiry, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			success := mode == "success" || mode == "upper_bound"
			if success || mode == "commit_error" {
				mock.ExpectExec(`UPDATE "user_subscriptions" SET "updated_at" = \$1, "status" = \$2.*`).
					WithArgs(sqlmock.AnyArg(), SubscriptionStatusActive, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
				if mode == "commit_error" {
					mock.ExpectCommit().WillReturnError(injected)
				} else {
					mock.ExpectCommit()
				}
			} else {
				mock.ExpectRollback()
			}
			sub, err := svc.ExtendSubscription(context.Background(), 7, days)
			if success {
				require.NoError(t, err)
				require.Equal(t, expectedExpiry, sub.ExpiresAt)
				require.Equal(t, SubscriptionStatusActive, sub.Status)
			} else {
				require.Error(t, err)
				require.Nil(t, sub)
				if mode == "read_error" {
					require.ErrorIs(t, err, ErrSubscriptionNotFound)
				} else if mode == "would_expire" {
					require.ErrorIs(t, err, ErrAdjustWouldExpire)
				} else if mode != "expired_negative" {
					require.ErrorIs(t, err, injected)
				}
				require.Equal(t, initial, repo.current.ExpiresAt)
				_, cached := repo.cache.Get(subCacheKey(11, 13))
				require.True(t, cached, "failed transaction must not invalidate cache")
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGapFixAdminConcurrentExtensions(t *testing.T) {
	svc, repo, mock := gapSubscriptionFixture(t)
	repo.skipCache = true // Concurrency test isolates serialization; lifecycle is checked above.
	initial := repo.current.ExpiresAt
	mock.MatchExpectationsInOrder(false)
	for i := 1; i <= 2; i++ {
		mock.ExpectBegin()
		expectGapSubscriptionRead(mock, initial.AddDate(0, 0, 5*(i-1)))
		mock.ExpectExec(`UPDATE "user_subscriptions" SET "updated_at" = \$1, "expires_at" = \$2.*`).
			WithArgs(sqlmock.AnyArg(), initial.AddDate(0, 0, 5*i), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start, errs := make(chan struct{}), make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := svc.ExtendSubscription(ctx, 7, 5)
			errs <- err
		}()
	}
	close(start)
	for range 2 {
		require.NoError(t, <-errs)
	}
	require.Equal(t, initial.AddDate(0, 0, 10), repo.current.ExpiresAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGapFixAdminExtendWithoutClient(t *testing.T) {
	initial := time.Now().AddDate(0, 0, 10)
	repo := &lockingRenewalRepo{current: UserSubscription{ID: 7, ExpiresAt: initial, Status: SubscriptionStatusActive}}
	svc := NewSubscriptionService(nil, repo, nil, nil, nil)
	sub, err := svc.ExtendSubscription(context.Background(), 7, 5)
	require.NoError(t, err)
	require.Equal(t, initial.AddDate(0, 0, 5), sub.ExpiresAt)
	require.Equal(t, 1, repo.lockReads)
}

type gapExtensionBillingCache struct {
	BillingCache
	calls atomic.Int32
	done  chan [2]int64
}

func (c *gapExtensionBillingCache) InvalidateSubscriptionCache(_ context.Context, userID, groupID int64) error {
	c.calls.Add(1)
	c.done <- [2]int64{userID, groupID}
	return nil
}

// This fixture uses actual ent SQL transactions, including the caller's final
// read. SQLmock permits only the caller's begin and only the later owner action.
type gapCallerSubscriptionRepo struct {
	userSubRepoNoop
	t       *testing.T
	outer   *dbent.Tx
	current UserSubscription
	pending UserSubscription
	steps   []string
	failAt  string
	failure error
}

func (r *gapCallerSubscriptionRepo) GetByIDForUpdate(ctx context.Context, id int64) (*UserSubscription, error) {
	require.Same(r.t, r.outer, dbent.TxFromContext(ctx), "lock read must use the caller's transaction")
	r.steps = append(r.steps, "lock")
	if r.failAt == "read_error" {
		return nil, r.failure
	}
	row, err := r.outer.Client().UserSubscription.Query().Where(usersubscription.IDEQ(id)).
		Select(usersubscription.FieldID, usersubscription.FieldExpiresAt).ForUpdate().Only(ctx)
	if err != nil {
		return nil, err
	}
	copy := r.current
	copy.ExpiresAt = row.ExpiresAt
	return &copy, nil
}

func (r *gapCallerSubscriptionRepo) ExtendExpiry(ctx context.Context, id int64, expiry time.Time) error {
	require.Same(r.t, r.outer, dbent.TxFromContext(ctx), "expiry write must use the caller's transaction")
	r.steps = append(r.steps, "expiry")
	if r.failAt == "expiry_error" {
		return r.failure
	}
	if err := r.outer.Client().UserSubscription.Update().Where(usersubscription.IDEQ(id)).SetExpiresAt(expiry).Exec(ctx); err != nil {
		return err
	}
	r.pending.ExpiresAt = expiry
	return nil
}

func (r *gapCallerSubscriptionRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	require.Same(r.t, r.outer, dbent.TxFromContext(ctx), "status write must use the caller's transaction")
	r.steps = append(r.steps, "status")
	if r.failAt == "status_error" {
		return r.failure
	}
	if err := r.outer.Client().UserSubscription.Update().Where(usersubscription.IDEQ(id)).SetStatus(status).Exec(ctx); err != nil {
		return err
	}
	r.pending.Status = status
	return nil
}

func (r *gapCallerSubscriptionRepo) GetByID(ctx context.Context, id int64) (*UserSubscription, error) {
	require.Same(r.t, r.outer, dbent.TxFromContext(ctx), "final read must still use the caller's transaction")
	r.steps = append(r.steps, "final")
	if r.failAt == "final_read_error" {
		return nil, r.failure
	}
	row, err := r.outer.Client().UserSubscription.Query().Where(usersubscription.IDEQ(id)).
		Select(usersubscription.FieldID, usersubscription.FieldExpiresAt, usersubscription.FieldStatus).Only(ctx)
	if err != nil {
		return nil, err
	}
	copy := r.pending
	copy.ExpiresAt, copy.Status = row.ExpiresAt, row.Status
	return &copy, nil
}

func TestGapFixAdminExtendCallerTransaction(t *testing.T) {
	for _, mode := range []string{"success", "rollback", "commit_error", "read_error", "expiry_error", "status_error", "final_read_error"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { db.Close() })
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
			mock.ExpectBegin()
			outer, err := client.Tx(ctx)
			require.NoError(t, err)
			cache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 100, MaxCost: 1 << 20, BufferItems: 64})
			require.NoError(t, err)
			t.Cleanup(cache.Close)
			require.True(t, cache.Set(subCacheKey(11, 13), "cached", 1))
			cache.Wait()
			billing := &gapExtensionBillingCache{done: make(chan [2]int64, 2)}
			initial := UserSubscription{
				ID: 7, UserID: 11, GroupID: 13, Status: SubscriptionStatusExpired,
				ExpiresAt: time.Now().UTC().AddDate(0, 0, 10),
			}
			repo := &gapCallerSubscriptionRepo{
				t: t, outer: outer, current: initial, pending: initial,
				failAt: mode, failure: errors.New("caller transaction fixture failure"),
			}
			outer.OnCommit(func(next dbent.Committer) dbent.Committer {
				return dbent.CommitFunc(func(ctx context.Context, tx *dbent.Tx) error {
					err := next.Commit(ctx, tx)
					if err == nil {
						repo.current = repo.pending
					}
					return err
				})
			})
			svc := NewSubscriptionService(nil, repo, &BillingCacheService{cache: billing}, client, nil)
			svc.subCacheL1 = cache
			assertCachesIntact := func() {
				t.Helper()
				_, cached := cache.Get(subCacheKey(11, 13))
				require.True(t, cached, "L1 must survive until the owner successfully commits")
				require.Zero(t, billing.calls.Load(), "billing invalidation must wait for the owner")
			}
			assertCachesIntact()
			expectedExpiry := initial.ExpiresAt.AddDate(0, 0, -5)
			if mode != "read_error" {
				expectGapSubscriptionRead(mock, initial.ExpiresAt)
			}
			if mode != "read_error" && mode != "expiry_error" {
				mock.ExpectExec(`UPDATE "user_subscriptions" SET "updated_at" = \$1, "expires_at" = \$2.*`).
					WithArgs(sqlmock.AnyArg(), expectedExpiry, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			if mode == "success" || mode == "rollback" || mode == "commit_error" || mode == "final_read_error" {
				mock.ExpectExec(`UPDATE "user_subscriptions" SET "updated_at" = \$1, "status" = \$2.*`).
					WithArgs(sqlmock.AnyArg(), SubscriptionStatusActive, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
				if mode != "final_read_error" {
					mock.ExpectQuery(`(?s)SELECT .* FROM "user_subscriptions".*`).
						WithArgs(int64(7)).
						WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at", "status"}).
							AddRow(int64(7), expectedExpiry, SubscriptionStatusActive))
				}
			}
			// No commit/rollback expectation exists while the service executes.
			sub, err := svc.ExtendSubscription(dbent.NewTxContext(ctx, outer), 7, -5)
			if mode == "success" || mode == "rollback" || mode == "commit_error" {
				require.NoError(t, err)
				require.Equal(t, expectedExpiry, sub.ExpiresAt)
				require.Equal(t, SubscriptionStatusActive, sub.Status)
				require.Equal(t, []string{"lock", "expiry", "status", "final"}, repo.steps)
			} else {
				require.Error(t, err)
				require.Nil(t, sub)
				if mode == "read_error" {
					require.ErrorIs(t, err, ErrSubscriptionNotFound)
				} else {
					require.ErrorIs(t, err, repo.failure)
				}
			}
			require.Equal(t, initial, repo.current, "the service must not commit the owner's writes")
			assertCachesIntact()
			require.NoError(t, mock.ExpectationsWereMet(), "only one begin and caller-transaction SQL are allowed")
			outer.OnCommit(func(next dbent.Committer) dbent.Committer {
				return dbent.CommitFunc(func(ctx context.Context, tx *dbent.Tx) error {
					assertCachesIntact()
					err := next.Commit(ctx, tx)
					assertCachesIntact() // The service's after-commit hook has not run yet.
					return err
				})
			})
			switch mode {
			case "success":
				mock.ExpectCommit()
				require.NoError(t, outer.Commit())
				_, cached := cache.Get(subCacheKey(11, 13))
				require.False(t, cached, "successful owner commit must invalidate L1")
				select {
				case ids := <-billing.done:
					require.Equal(t, [2]int64{11, 13}, ids)
				case <-time.After(time.Second):
					t.Fatal("successful owner commit did not invalidate billing cache")
				}
				require.EqualValues(t, 1, billing.calls.Load())
				require.Equal(t, expectedExpiry, repo.current.ExpiresAt)
				require.Equal(t, SubscriptionStatusActive, repo.current.Status)
			case "commit_error":
				mock.ExpectCommit().WillReturnError(repo.failure)
				require.ErrorIs(t, outer.Commit(), repo.failure)
			default:
				mock.ExpectRollback()
				require.NoError(t, outer.Rollback(), "only the caller may roll back")
			}
			if mode != "success" {
				require.Equal(t, initial, repo.current)
				assertCachesIntact()
				require.Never(t, func() bool { return billing.calls.Load() != 0 }, 50*time.Millisecond, time.Millisecond)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGapFixAdminExtendBillingCacheLifecycle(t *testing.T) {
	for _, mode := range []string{"success", "rollback", "commit_error"} {
		t.Run(mode, func(t *testing.T) {
			svc, repo, mock := gapSubscriptionFixture(t)
			billing := &gapExtensionBillingCache{done: make(chan [2]int64, 2)}
			svc.billingCacheService = &BillingCacheService{cache: billing}
			failure := errors.New("standalone cache fixture failure")
			initial := repo.current.ExpiresAt
			repo.current.Status = SubscriptionStatusExpired
			mock.ExpectBegin()
			expectGapSubscriptionRead(mock, initial)
			mock.ExpectExec(`UPDATE "user_subscriptions" SET "updated_at" = \$1, "expires_at" = \$2.*`).
				WithArgs(sqlmock.AnyArg(), initial.AddDate(0, 0, 5), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
			if mode == "rollback" {
				repo.statusErr = failure
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(`UPDATE "user_subscriptions" SET "updated_at" = \$1, "status" = \$2.*`).
					WithArgs(sqlmock.AnyArg(), SubscriptionStatusActive, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
				if mode == "commit_error" {
					mock.ExpectCommit().WillReturnError(failure)
				} else {
					mock.ExpectCommit()
				}
			}
			require.Zero(t, billing.calls.Load())
			_, err := svc.ExtendSubscription(context.Background(), 7, 5)
			if mode == "success" {
				require.NoError(t, err)
				select {
				case ids := <-billing.done:
					require.Equal(t, [2]int64{11, 13}, ids)
				case <-time.After(time.Second):
					t.Fatal("successful standalone commit did not invalidate billing cache")
				}
				require.EqualValues(t, 1, billing.calls.Load())
			} else {
				require.ErrorIs(t, err, failure)
				require.Never(t, func() bool { return billing.calls.Load() != 0 }, 50*time.Millisecond, time.Millisecond)
				require.Equal(t, initial, repo.current.ExpiresAt)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

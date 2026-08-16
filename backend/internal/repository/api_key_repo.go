package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

type apiKeyRepository struct {
	client  *dbent.Client
	sql     sqlExecutor
	dialect string
}

func NewAPIKeyRepository(client *dbent.Client, sqlDB *sql.DB) service.APIKeyRepository {
	return newAPIKeyRepositoryWithSQL(client, sqlDB)
}

func newAPIKeyRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *apiKeyRepository {
	return &apiKeyRepository{client: client, sql: sqlq, dialect: dialect.Postgres}
}

func (r *apiKeyRepository) executor(ctx context.Context) sqlExecutor {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	if r.sql != nil {
		return r.sql
	}
	return r.client
}

func (r *apiKeyRepository) isSQLite() bool {
	return r.dialect == dialect.SQLite
}

func (r *apiKeyRepository) activeQuery() *dbent.APIKeyQuery {
	// 默认过滤已软删除记录，避免删除后仍被查询到。
	return r.client.APIKey.Query().Where(apikey.DeletedAtIsNil())
}

func (r *apiKeyRepository) Create(ctx context.Context, key *service.APIKey) error {
	routeJSON, err := marshalAPIKeyMultiGroupRoutes(key.MultiGroupRoutes)
	if err != nil {
		return err
	}
	client := clientFromContext(ctx, r.client)
	builder := client.APIKey.Create().
		SetUserID(key.UserID).
		SetKey(key.Key).
		SetName(key.Name).
		SetStatus(key.Status).
		SetNillableGroupID(key.GroupID).
		SetNillableLastUsedAt(key.LastUsedAt).
		SetQuota(key.Quota).
		SetQuotaUsed(key.QuotaUsed).
		SetNillableExpiresAt(key.ExpiresAt).
		SetRateLimit5h(key.RateLimit5h).
		SetRateLimit1d(key.RateLimit1d).
		SetRateLimit7d(key.RateLimit7d)
	if key.ManagedSourceType != "" {
		builder.SetManagedSourceType(key.ManagedSourceType)
	}
	if key.ManagedSourceID != nil {
		builder.SetManagedSourceID(*key.ManagedSourceID)
	}

	if len(key.IPWhitelist) > 0 {
		builder.SetIPWhitelist(key.IPWhitelist)
	}
	if len(key.IPBlacklist) > 0 {
		builder.SetIPBlacklist(key.IPBlacklist)
	}

	created, err := builder.Save(ctx)
	if err == nil {
		key.ID = created.ID
		key.LastUsedAt = created.LastUsedAt
		key.CreatedAt = created.CreatedAt
		key.UpdatedAt = created.UpdatedAt
		err = r.setAPIKeyMultiGroupRoutes(ctx, key.ID, routeJSON)
		if err == nil {
			err = r.setAPIKeyAccountPoolStrategy(ctx, key.ID, service.NormalizeAccountPoolStrategy(key.AccountPoolStrategy))
		}
	}
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrAPIKeyExists)
	}
	return nil
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id int64) (*service.APIKey, error) {
	m, err := r.activeQuery().
		Where(apikey.IDEQ(id)).
		WithUser().
		WithGroup().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrAPIKeyNotFound
		}
		return nil, err
	}
	out := apiKeyEntityToService(m)
	if err := r.loadAPIKeyMultiGroupRoutes(ctx, out); err != nil {
		return nil, err
	}
	if err := r.hydrateAPIKeyMultiGroupRouteGroups(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetKeyAndOwnerID 根据 API Key ID 获取其 key 与所有者（用户）ID。
// 相比 GetByID，此方法性能更优，因为：
//   - 使用 Select() 只查询必要字段，减少数据传输量
//   - 不加载完整的 API Key 实体及其关联数据（User、Group 等）
//   - 适用于删除等只需 key 与用户 ID 的场景
func (r *apiKeyRepository) GetKeyAndOwnerID(ctx context.Context, id int64) (string, int64, error) {
	m, err := r.activeQuery().
		Where(apikey.IDEQ(id)).
		Select(apikey.FieldKey, apikey.FieldUserID).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return "", 0, service.ErrAPIKeyNotFound
		}
		return "", 0, err
	}
	return m.Key, m.UserID, nil
}

func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*service.APIKey, error) {
	m, err := r.activeQuery().
		Where(apikey.KeyEQ(key)).
		WithUser(func(q *dbent.UserQuery) {
			q.WithAllowedGroups(func(gq *dbent.GroupQuery) {
				gq.Select(group.FieldID)
			})
		}).
		WithGroup().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrAPIKeyNotFound
		}
		return nil, err
	}
	out := apiKeyEntityToService(m)
	if err := r.loadAPIKeyMultiGroupRoutes(ctx, out); err != nil {
		return nil, err
	}
	if err := r.hydrateAPIKeyMultiGroupRouteGroups(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *apiKeyRepository) GetByKeyForAuth(ctx context.Context, key string) (*service.APIKey, error) {
	now := time.Now()
	m, err := r.activeQuery().
		Where(apikey.KeyEQ(key)).
		Select(
			apikey.FieldID,
			apikey.FieldUserID,
			apikey.FieldGroupID,
			apikey.FieldName,
			apikey.FieldStatus,
			apikey.FieldIPWhitelist,
			apikey.FieldIPBlacklist,
			apikey.FieldQuota,
			apikey.FieldQuotaUsed,
			apikey.FieldExpiresAt,
			apikey.FieldRateLimit5h,
			apikey.FieldRateLimit1d,
			apikey.FieldRateLimit7d,
			apikey.FieldManagedSourceType,
			apikey.FieldManagedSourceID,
		).
		WithUser(func(q *dbent.UserQuery) {
			q.Select(
				user.FieldID,
				user.FieldEmail,
				user.FieldUsername,
				user.FieldStatus,
				user.FieldRole,
				user.FieldBalance,
				user.FieldConcurrency,
				user.FieldBalanceNotifyEnabled,
				user.FieldBalanceNotifyThresholdType,
				user.FieldBalanceNotifyThreshold,
				user.FieldBalanceNotifyExtraEmails,
				user.FieldTotalRecharged,
				user.FieldSignupSource,
				user.FieldLastLoginAt,
				user.FieldLastActiveAt,
				user.FieldRpmLimit,
			)
			q.WithAllowedGroups(func(gq *dbent.GroupQuery) {
				gq.Select(group.FieldID)
			})
		}).
		WithGroup(func(q *dbent.GroupQuery) {
			q.Select(
				group.FieldID,
				group.FieldName,
				group.FieldPlatform,
				group.FieldIsExclusive,
				group.FieldStatus,
				group.FieldSubscriptionType,
				group.FieldRoutingScope,
				group.FieldRateMultiplier,
				group.FieldDailyLimitUsd,
				group.FieldWeeklyLimitUsd,
				group.FieldMonthlyLimitUsd,
				group.FieldLongContextPricingEnabled,
				group.FieldModelPricing,
				group.FieldAllowImageGeneration,
				group.FieldImageRateIndependent,
				group.FieldImageRateMultiplier,
				group.FieldImagePrice1k,
				group.FieldImagePrice2k,
				group.FieldImagePrice4k,
				group.FieldClaudeCodeOnly,
				group.FieldFallbackGroupID,
				group.FieldFallbackGroupIDOnInvalidRequest,
				group.FieldModelRoutingEnabled,
				group.FieldModelRouting,
				group.FieldModelMatchPatterns,
				group.FieldMcpXMLInject,
				group.FieldSupportedModelScopes,
				group.FieldAllowMessagesDispatch,
				group.FieldDefaultMappedModel,
				group.FieldMessagesDispatchModelConfig,
				group.FieldModelsListConfig,
				group.FieldRpmLimit,
				group.FieldPeakRateEnabled,
				group.FieldPeakStart,
				group.FieldPeakEnd,
				group.FieldPeakRateMultiplier,
			)
		}).
		WithAccountBindings(func(q *dbent.APIKeyAccountBindingQuery) {
			q.Where(
				apikeyaccountbinding.StatusEQ("active"),
				apikeyaccountbinding.StartsAtLTE(now),
				apikeyaccountbinding.ExpiresAtGT(now),
			)
			q.WithAccount()
			q.WithGroup()
			q.WithSeat(func(sq *dbent.GroupBuySeatQuery) {
				sq.WithRound()
			})
		}).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrAPIKeyNotFound
		}
		return nil, err
	}
	out := apiKeyEntityToService(m)
	if out.IsCafeRoomManaged() && len(m.Edges.AccountBindings) == 1 {
		binding := m.Edges.AccountBindings[0]
		if binding != nil && binding.StrictMode && binding.Status == "active" &&
			binding.APIKeyID == out.ID && binding.UserID == out.UserID && out.GroupID != nil && binding.GroupID == *out.GroupID &&
			out.ManagedSourceID != nil && binding.SeatID == *out.ManagedSourceID &&
			binding.Edges.Account != nil && binding.Edges.Account.ID == binding.AccountID &&
			binding.Edges.Group != nil && binding.Edges.Group.ID == binding.GroupID &&
			binding.Edges.Seat != nil && binding.Edges.Seat.ID == binding.SeatID && binding.Edges.Seat.UserID == out.UserID && binding.Edges.Seat.RoundID == binding.RoundID && binding.Edges.Seat.Status == "active" &&
			binding.Edges.Seat.Edges.Round != nil && binding.Edges.Seat.Edges.Round.ID == binding.RoundID && binding.Edges.Seat.Edges.Round.Status == "active" &&
			binding.Edges.Seat.Edges.Round.CafeRoomID != nil && *binding.Edges.Seat.Edges.Round.CafeRoomID == binding.CafeRoomID &&
			binding.Edges.Seat.Edges.Round.AssignedAccountID != nil && *binding.Edges.Seat.Edges.Round.AssignedAccountID == binding.AccountID {
			out.PinnedAccountID = binding.AccountID
			out.ManagedBindingID = binding.ID
			expiresAt := binding.ExpiresAt
			out.ManagedBindingExpiresAt = &expiresAt
		}
	}
	if err := r.loadAPIKeyMultiGroupRoutes(ctx, out); err != nil {
		return nil, err
	}
	if err := r.hydrateAPIKeyMultiGroupRouteGroups(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, key *service.APIKey, fields service.APIKeyUpdateFields) error {
	if key == nil || fields.IsEmpty() {
		return nil
	}
	var routeJSON string
	var err error
	if fields.MultiGroupRoutes {
		routeJSON, err = marshalAPIKeyMultiGroupRoutes(key.MultiGroupRoutes)
		if err != nil {
			return err
		}
	}
	// 使用原子操作：将软删除检查与更新合并到同一语句，避免竞态条件。
	// 之前的实现先检查 Exist 再 UpdateOneID，若在两步之间发生软删除，
	// 则会更新已删除的记录。
	// 这里选择 Update().Where()，确保只有未软删除记录能被更新。
	// 同时显式设置 updated_at，避免二次查询带来的并发可见性问题。
	client := clientFromContext(ctx, r.client)
	now := time.Now()
	builder := client.APIKey.Update().Where(apikey.IDEQ(key.ID), apikey.DeletedAtIsNil()).SetUpdatedAt(now)
	if fields.Name {
		builder.SetName(key.Name)
	}
	if fields.Status {
		builder.SetStatus(key.Status)
	}
	if fields.Quota {
		builder.SetQuota(key.Quota)
	}
	if fields.QuotaUsed {
		builder.SetQuotaUsed(key.QuotaUsed)
	}
	if fields.RateLimits {
		builder.SetRateLimit5h(key.RateLimit5h).SetRateLimit1d(key.RateLimit1d).SetRateLimit7d(key.RateLimit7d)
	}
	if fields.RateLimitUsage {
		builder.SetUsage5h(key.Usage5h).SetUsage1d(key.Usage1d).SetUsage7d(key.Usage7d)
		if key.Window5hStart != nil {
			builder.SetWindow5hStart(*key.Window5hStart)
		} else {
			builder.ClearWindow5hStart()
		}
		if key.Window1dStart != nil {
			builder.SetWindow1dStart(*key.Window1dStart)
		} else {
			builder.ClearWindow1dStart()
		}
		if key.Window7dStart != nil {
			builder.SetWindow7dStart(*key.Window7dStart)
		} else {
			builder.ClearWindow7dStart()
		}
	}
	if fields.GroupID {
		if key.GroupID != nil {
			builder.SetGroupID(*key.GroupID)
		} else {
			builder.ClearGroupID()
		}
	}
	if fields.ExpiresAt {
		if key.ExpiresAt != nil {
			builder.SetExpiresAt(*key.ExpiresAt)
		} else {
			builder.ClearExpiresAt()
		}
	}
	if fields.IPRules {
		if len(key.IPWhitelist) > 0 {
			builder.SetIPWhitelist(key.IPWhitelist)
		} else {
			builder.ClearIPWhitelist()
		}
		if len(key.IPBlacklist) > 0 {
			builder.SetIPBlacklist(key.IPBlacklist)
		} else {
			builder.ClearIPBlacklist()
		}
	}

	affected, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		// 更新影响行数为 0，说明记录不存在或已被软删除。
		return service.ErrAPIKeyNotFound
	}
	if fields.MultiGroupRoutes {
		if err := r.setAPIKeyMultiGroupRoutes(ctx, key.ID, routeJSON); err != nil {
			return err
		}
	}
	if fields.AccountPoolStrategy {
		if err := r.setAPIKeyAccountPoolStrategy(ctx, key.ID, service.NormalizeAccountPoolStrategy(key.AccountPoolStrategy)); err != nil {
			return err
		}
	}

	// 使用同一时间戳回填，避免并发删除导致二次查询失败。
	key.UpdatedAt = now
	return nil
}

func (r *apiKeyRepository) UpdateCafeManagedAPIKey(ctx context.Context, key *service.APIKey, desiredStatus string, now time.Time) error {
	if key == nil || key.ID <= 0 || key.UserID <= 0 || key.ManagedSourceID == nil || *key.ManagedSourceID <= 0 ||
		key.ManagedSourceType != service.APIKeyManagedSourceCafeRoomSeat {
		return service.ErrCafeManagedKeyStateUnavailable
	}
	if desiredStatus != service.StatusAPIKeyActive && desiredStatus != "inactive" {
		return service.ErrCafeManagedKeyStatusInvalid
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin cafe managed key update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	// Read only the immutable round reference first, then acquire locks in the
	// same order as Cafe expiry: Round -> Seat -> Binding -> Key.
	seatRef, err := tx.GroupBuySeat.Query().
		Where(groupbuyseat.IDEQ(*key.ManagedSourceID)).
		Select(groupbuyseat.FieldID, groupbuyseat.FieldRoundID).
		Only(txCtx)
	if err != nil {
		if isEntLookupMiss(err) {
			return cafeManagedKeyStateError(key.ID, "managed seat is missing")
		}
		return fmt.Errorf("load managed seat %d for cafe key %d: %w", *key.ManagedSourceID, key.ID, err)
	}
	roundQuery := tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(seatRef.RoundID))
	if r.client.Driver().Dialect() != dialect.SQLite {
		roundQuery = roundQuery.ForUpdate()
	}
	round, err := roundQuery.Only(txCtx)
	if err != nil {
		if isEntLookupMiss(err) {
			return cafeManagedKeyStateError(key.ID, "managed round is missing")
		}
		return fmt.Errorf("lock managed round %d for cafe key %d: %w", seatRef.RoundID, key.ID, err)
	}
	seatQuery := tx.GroupBuySeat.Query().Where(groupbuyseat.IDEQ(seatRef.ID))
	if r.client.Driver().Dialect() != dialect.SQLite {
		seatQuery = seatQuery.ForUpdate()
	}
	seat, err := seatQuery.Only(txCtx)
	if err != nil {
		if isEntLookupMiss(err) {
			return cafeManagedKeyStateError(key.ID, "managed seat could not be locked")
		}
		return fmt.Errorf("lock managed seat %d for cafe key %d: %w", seatRef.ID, key.ID, err)
	}

	var binding *dbent.APIKeyAccountBinding
	if desiredStatus == service.StatusAPIKeyActive {
		bindingQuery := tx.APIKeyAccountBinding.Query().Where(
			apikeyaccountbinding.SeatIDEQ(seat.ID),
			apikeyaccountbinding.StatusEQ("active"),
		)
		if r.client.Driver().Dialect() != dialect.SQLite {
			bindingQuery = bindingQuery.ForUpdate()
		}
		binding, err = bindingQuery.Only(txCtx)
		if err != nil {
			if isEntLookupMiss(err) {
				return cafeManagedKeyEnableError(key.ID, "active binding is missing or ambiguous")
			}
			return fmt.Errorf("lock active managed binding for cafe key %d: %w", key.ID, err)
		}
	}

	persistedQuery := tx.APIKey.Query().Where(apikey.IDEQ(key.ID), apikey.DeletedAtIsNil())
	if r.client.Driver().Dialect() != dialect.SQLite {
		persistedQuery = persistedQuery.ForUpdate()
	}
	persisted, err := persistedQuery.Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrAPIKeyNotFound
		}
		return fmt.Errorf("lock cafe managed key %d: %w", key.ID, err)
	}
	if persisted.UserID != key.UserID || persisted.ManagedSourceType != service.APIKeyManagedSourceCafeRoomSeat ||
		persisted.ManagedSourceID == nil || *persisted.ManagedSourceID != seat.ID {
		return cafeManagedKeyStateError(key.ID, "managed key ownership or source does not match")
	}
	switch persisted.Status {
	case service.StatusAPIKeyActive, service.StatusAPIKeyDisabled, "inactive":
		// User-controlled transitions are allowed only from recoverable states.
	default:
		return cafeManagedKeyStateError(key.ID, "managed key is in a terminal state")
	}

	if desiredStatus == service.StatusAPIKeyActive {
		if err := validateCafeManagedKeyEnableFacts(txCtx, tx, persisted, round, seat, binding, now); err != nil {
			return err
		}
	}

	update := tx.APIKey.UpdateOneID(persisted.ID).
		SetName(key.Name).
		SetStatus(desiredStatus).
		SetUpdatedAt(now)
	if len(key.IPWhitelist) > 0 {
		update.SetIPWhitelist(key.IPWhitelist)
	} else {
		update.ClearIPWhitelist()
	}
	if len(key.IPBlacklist) > 0 {
		update.SetIPBlacklist(key.IPBlacklist)
	} else {
		update.ClearIPBlacklist()
	}
	if _, err := update.Save(txCtx); err != nil {
		return fmt.Errorf("update cafe managed key %d: %w", persisted.ID, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cafe managed key %d update: %w", persisted.ID, err)
	}
	return nil
}

func validateCafeManagedKeyEnableFacts(ctx context.Context, tx *dbent.Tx, key *dbent.APIKey, round *dbent.GroupBuyRound, seat *dbent.GroupBuySeat, binding *dbent.APIKeyAccountBinding, now time.Time) error {
	if key == nil || round == nil || seat == nil || binding == nil || key.GroupID == nil || key.ExpiresAt == nil ||
		round.CafeRoomID == nil || round.AssignedAccountID == nil || round.EntitlementExpiresAt == nil || seat.BoundAPIKeyID == nil || seat.ExpiresAt == nil {
		return cafeManagedKeyEnableError(cafeManagedKeyID(key), "managed entitlement facts are incomplete")
	}
	if round.Status != service.GroupBuyRoundStatusActive || seat.Status != service.GroupBuySeatStatusActive || binding.Status != "active" || !binding.StrictMode ||
		seat.RoundID != round.ID || seat.UserID != key.UserID || *seat.BoundAPIKeyID != key.ID ||
		binding.APIKeyID != key.ID || binding.UserID != key.UserID || binding.GroupID != *key.GroupID || binding.AccountID != *round.AssignedAccountID ||
		binding.CafeRoomID != *round.CafeRoomID || binding.RoundID != round.ID || binding.SeatID != seat.ID ||
		binding.StartsAt.After(now) || !binding.ExpiresAt.After(now) || !seat.ExpiresAt.After(now) || !key.ExpiresAt.After(now) || !round.EntitlementExpiresAt.After(now) ||
		!binding.ExpiresAt.Equal(*seat.ExpiresAt) || !binding.ExpiresAt.Equal(*key.ExpiresAt) || !binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
		return cafeManagedKeyEnableError(key.ID, "managed entitlement is inactive, expired or inconsistent")
	}
	managedGroup, err := tx.Group.Query().Where(group.IDEQ(binding.GroupID)).Only(ctx)
	if err != nil {
		if isEntLookupMiss(err) {
			return cafeManagedKeyEnableError(key.ID, "managed group is unavailable")
		}
		return fmt.Errorf("load managed group %d for cafe key %d: %w", binding.GroupID, key.ID, err)
	}
	if managedGroup.Status != service.StatusActive || managedGroup.AccessMode != service.CafeRoomGroupAccessMode {
		return cafeManagedKeyEnableError(key.ID, "managed group is unavailable")
	}
	assignedAccount, err := tx.Account.Query().Where(
		account.IDEQ(binding.AccountID),
		account.StatusEQ(service.StatusActive),
		account.HasGroupsWith(group.IDEQ(binding.GroupID)),
	).Only(ctx)
	if err != nil {
		if isEntLookupMiss(err) {
			return cafeManagedKeyEnableError(key.ID, "assigned account is unavailable")
		}
		return fmt.Errorf("load assigned account %d for cafe key %d: %w", binding.AccountID, key.ID, err)
	}
	if assignedAccount.ID != binding.AccountID {
		return cafeManagedKeyEnableError(key.ID, "assigned account is unavailable")
	}
	return nil
}

func isEntLookupMiss(err error) bool {
	return dbent.IsNotFound(err) || dbent.IsNotSingular(err)
}

func cafeManagedKeyID(key *dbent.APIKey) int64 {
	if key == nil {
		return 0
	}
	return key.ID
}

func cafeManagedKeyStateError(keyID int64, reason string) error {
	return service.ErrCafeManagedKeyStateUnavailable.WithMetadata(map[string]string{
		"key_id": fmt.Sprint(keyID),
		"reason": reason,
	})
}

func cafeManagedKeyEnableError(keyID int64, reason string) error {
	return service.ErrCafeManagedKeyEnableUnavailable.WithMetadata(map[string]string{
		"key_id": fmt.Sprint(keyID),
		"reason": reason,
	})
}

func (r *apiKeyRepository) Delete(ctx context.Context, id int64) error {
	// 存在唯一键约束 生成tombstone key 用来释放原key，长度远小于 128，满足 schema 限制
	tombstoneKey := fmt.Sprintf("__deleted__%d__%d", id, time.Now().UnixNano())
	// 显式软删除：避免依赖 Hook 行为，确保 deleted_at 一定被设置。
	affected, err := r.client.APIKey.Update().
		Where(apikey.IDEQ(id), apikey.DeletedAtIsNil()).
		SetKey(tombstoneKey).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrAPIKeyNotFound
		}
		return err
	}
	if affected == 0 {
		exists, err := r.client.APIKey.Query().
			Where(apikey.IDEQ(id)).
			Exist(mixins.SkipSoftDelete(ctx))
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		return service.ErrAPIKeyNotFound
	}
	return nil
}

// DeleteWithAudit 在同一事务内:
//  1. 把(明文 key、所有者、key 名称)写入 deleted_api_key_audits;
//  2. 软删除该 key(tombstone 覆盖 key 列以释放唯一约束)。
//
// 保证"被删除的 key 一定能反查到所有者"。事务模式与 group_repo.DeleteCascade 一致。
func (r *apiKeyRepository) DeleteWithAudit(ctx context.Context, id int64) error {
	tombstoneKey := fmt.Sprintf("__deleted__%d__%d", id, time.Now().UnixNano())

	if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
		return r.deleteWithAudit(ctx, existingTx.Client(), id, tombstoneKey)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return err
	}
	exec := r.client
	if err == nil {
		defer func() { _ = tx.Rollback() }()
		exec = tx.Client()
	}

	if err := r.deleteWithAudit(ctx, exec, id, tombstoneKey); err != nil {
		return err
	}

	if tx != nil {
		return tx.Commit()
	}
	return nil
}

func (r *apiKeyRepository) deleteWithAudit(ctx context.Context, exec *dbent.Client, id int64, tombstoneKey string) error {
	// 1. 审计:数据源即 api_keys 当前行;WHERE deleted_at IS NULL 保证只对未删除行写一次。
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO deleted_api_key_audits (key, api_key_id, user_id, key_name, deleted_at)
		SELECT key, id, user_id, name, NOW()
		FROM api_keys
		WHERE id = $1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}

	// 2. 软删除(tombstone 覆盖 key)。
	res, err := exec.ExecContext(ctx, `
		UPDATE api_keys
		SET key = $1, deleted_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`, tombstoneKey, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		// 并发/重复删除:记录已存在(已软删)则幂等返回 nil(defer 回滚空事务),否则 NotFound。
		exists, existErr := exec.APIKey.Query().
			Where(apikey.IDEQ(id)).
			Exist(mixins.SkipSoftDelete(ctx))
		if existErr != nil {
			return existErr
		}
		if exists {
			return nil
		}
		return service.ErrAPIKeyNotFound
	}
	return nil
}

func (r *apiKeyRepository) ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams, filters service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error) {
	q := r.activeQuery().Where(apikey.UserIDEQ(userID))

	// Apply filters
	if filters.Search != "" {
		q = q.Where(apikey.Or(
			apikey.NameContainsFold(filters.Search),
			apikey.KeyContainsFold(filters.Search),
		))
	}
	if filters.Status != "" {
		q = q.Where(apikey.StatusEQ(filters.Status))
	}
	if filters.GroupID != nil {
		if *filters.GroupID == 0 {
			q = q.Where(apikey.GroupIDIsNil())
		} else {
			ids, err := r.apiKeyIDsByRelatedGroup(ctx, *filters.GroupID)
			if err != nil {
				return nil, nil, err
			}
			if len(ids) == 0 {
				return []service.APIKey{}, paginationResultFromTotal(0, params), nil
			}
			q = q.Where(apikey.IDIn(ids...))
		}
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	keysQuery := q.
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range apiKeyListOrder(params) {
		keysQuery = keysQuery.Order(order)
	}

	keys, err := keysQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	outKeys := make([]service.APIKey, 0, len(keys))
	for i := range keys {
		out := apiKeyEntityToService(keys[i])
		if err := r.loadAPIKeyMultiGroupRoutes(ctx, out); err != nil {
			return nil, nil, err
		}
		if err := r.hydrateAPIKeyMultiGroupRouteGroups(ctx, out); err != nil {
			return nil, nil, err
		}
		outKeys = append(outKeys, *out)
	}

	return outKeys, paginationResultFromTotal(int64(total), params), nil
}

func (r *apiKeyRepository) VerifyOwnership(ctx context.Context, userID int64, apiKeyIDs []int64) ([]int64, error) {
	if len(apiKeyIDs) == 0 {
		return []int64{}, nil
	}

	ids, err := r.client.APIKey.Query().
		Where(apikey.UserIDEQ(userID), apikey.IDIn(apiKeyIDs...), apikey.DeletedAtIsNil()).
		IDs(ctx)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *apiKeyRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	count, err := r.activeQuery().Where(apikey.UserIDEQ(userID)).Count(ctx)
	return int64(count), err
}

// BackfillDefaultKeyFallbackGroup updates the lowest non-deleted API-key ID per
// user only when that default key is still ungrouped. RETURNING provides the
// exact cache keys changed by the guarded update.
func (r *apiKeyRepository) BackfillDefaultKeyFallbackGroup(ctx context.Context, groupID int64) ([]string, error) {
	if groupID <= 0 {
		return nil, service.ErrDefaultKeyFallbackGroupInvalid
	}
	placeholder := "$1"
	if r.isSQLite() {
		placeholder = "?"
	}
	rows, err := r.executor(ctx).QueryContext(ctx, `
		WITH default_keys AS (
			SELECT MIN(id) AS id
			FROM api_keys
			WHERE deleted_at IS NULL
			GROUP BY user_id
		)
		UPDATE api_keys
		SET group_id = `+placeholder+`, updated_at = CURRENT_TIMESTAMP
		WHERE deleted_at IS NULL
			AND group_id IS NULL
			AND id IN (SELECT id FROM default_keys)
		RETURNING key`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *apiKeyRepository) ExistsByKey(ctx context.Context, key string) (bool, error) {
	count, err := r.activeQuery().Where(apikey.KeyEQ(key)).Count(ctx)
	return count > 0, err
}

func (r *apiKeyRepository) ListByGroupID(ctx context.Context, groupID int64, params pagination.PaginationParams) ([]service.APIKey, *pagination.PaginationResult, error) {
	ids, err := r.apiKeyIDsByRelatedGroup(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	if len(ids) == 0 {
		return []service.APIKey{}, paginationResultFromTotal(0, params), nil
	}
	q := r.activeQuery().Where(apikey.IDIn(ids...))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	keysQuery := q.
		WithUser().
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range apiKeyListOrder(params) {
		keysQuery = keysQuery.Order(order)
	}

	keys, err := keysQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	outKeys := make([]service.APIKey, 0, len(keys))
	for i := range keys {
		out := apiKeyEntityToService(keys[i])
		if err := r.loadAPIKeyMultiGroupRoutes(ctx, out); err != nil {
			return nil, nil, err
		}
		if err := r.hydrateAPIKeyMultiGroupRouteGroups(ctx, out); err != nil {
			return nil, nil, err
		}
		outKeys = append(outKeys, *out)
	}

	return outKeys, paginationResultFromTotal(int64(total), params), nil
}

func apiKeyListOrder(params pagination.PaginationParams) []func(*entsql.Selector) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderDesc)

	var field string
	switch sortBy {
	case "name":
		field = apikey.FieldName
	case "status":
		field = apikey.FieldStatus
	case "expires_at":
		field = apikey.FieldExpiresAt
	case "last_used_at":
		field = apikey.FieldLastUsedAt
	case "created_at":
		field = apikey.FieldCreatedAt
	default:
		field = apikey.FieldID
	}

	if sortOrder == pagination.SortOrderAsc {
		return []func(*entsql.Selector){dbent.Asc(field), dbent.Asc(apikey.FieldID)}
	}
	return []func(*entsql.Selector){dbent.Desc(field), dbent.Desc(apikey.FieldID)}
}

// SearchAPIKeys searches API keys by user ID and/or keyword (name)
func (r *apiKeyRepository) SearchAPIKeys(ctx context.Context, userID int64, keyword string, limit int) ([]service.APIKey, error) {
	q := r.activeQuery()
	if userID > 0 {
		q = q.Where(apikey.UserIDEQ(userID))
	}

	if keyword != "" {
		q = q.Where(apikey.NameContainsFold(keyword))
	}

	keys, err := q.Limit(limit).Order(dbent.Desc(apikey.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}

	outKeys := make([]service.APIKey, 0, len(keys))
	for i := range keys {
		out := apiKeyEntityToService(keys[i])
		if err := r.loadAPIKeyMultiGroupRoutes(ctx, out); err != nil {
			return nil, err
		}
		if err := r.hydrateAPIKeyMultiGroupRouteGroups(ctx, out); err != nil {
			return nil, err
		}
		outKeys = append(outKeys, *out)
	}
	return outKeys, nil
}

// ClearGroupIDByGroupID 将指定分组的所有 API Key 的 group_id 设为 nil
func (r *apiKeyRepository) ClearGroupIDByGroupID(ctx context.Context, groupID int64) (int64, error) {
	n, err := r.client.APIKey.Update().
		Where(apikey.GroupIDEQ(groupID), apikey.DeletedAtIsNil()).
		ClearGroupID().
		Save(ctx)
	if err != nil {
		return int64(n), err
	}
	contains, err := apiKeyRouteContainsJSON(groupID)
	if err != nil {
		return int64(n), err
	}
	if r.isSQLite() {
		res, err := r.executor(ctx).ExecContext(ctx, `
			UPDATE api_keys
			SET
				multi_group_routes = COALESCE((
					SELECT json_group_array(value)
					FROM json_each(multi_group_routes)
					WHERE CAST(json_extract(value, '$.group_id') AS INTEGER) <> ?
				), '[]'),
				updated_at = CURRENT_TIMESTAMP
			WHERE deleted_at IS NULL
			  AND EXISTS (
				SELECT 1
				FROM json_each(multi_group_routes)
				WHERE CAST(json_extract(value, '$.group_id') AS INTEGER) = ?
			  )`,
			groupID, groupID)
		if err != nil {
			return int64(n), err
		}
		routeAffected, _ := res.RowsAffected()
		return int64(n) + routeAffected, nil
	}
	res, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE api_keys
		SET
			multi_group_routes = COALESCE((
				SELECT jsonb_agg(elem)
				FROM jsonb_array_elements(multi_group_routes) elem
				WHERE (elem->>'group_id')::bigint <> $1
			), '[]'::jsonb),
			updated_at = NOW()
		WHERE deleted_at IS NULL
		  AND multi_group_routes @> $2::jsonb`,
		groupID, contains)
	if err != nil {
		return int64(n), err
	}
	routeAffected, _ := res.RowsAffected()
	return int64(n) + routeAffected, nil
}

// UpdateGroupIDByUserAndGroup 将用户下绑定 oldGroupID 的所有 Key 迁移到 newGroupID
func (r *apiKeyRepository) UpdateGroupIDByUserAndGroup(ctx context.Context, userID, oldGroupID, newGroupID int64) (int64, error) {
	client := clientFromContext(ctx, r.client)
	n, err := client.APIKey.Update().
		Where(apikey.UserIDEQ(userID), apikey.GroupIDEQ(oldGroupID), apikey.DeletedAtIsNil()).
		SetGroupID(newGroupID).
		Save(ctx)
	return int64(n), err
}

// CountByGroupID 获取分组的 API Key 数量
func (r *apiKeyRepository) CountByGroupID(ctx context.Context, groupID int64) (int64, error) {
	ids, err := r.apiKeyIDsByRelatedGroup(ctx, groupID)
	if err != nil {
		return 0, err
	}
	return int64(len(ids)), nil
}

func (r *apiKeyRepository) ListKeysByUserID(ctx context.Context, userID int64) ([]string, error) {
	keys, err := r.activeQuery().
		Where(apikey.UserIDEQ(userID)).
		Select(apikey.FieldKey).
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *apiKeyRepository) ListKeysByGroupID(ctx context.Context, groupID int64) ([]string, error) {
	contains, err := apiKeyRouteContainsJSON(groupID)
	if err != nil {
		return nil, err
	}
	if r.isSQLite() {
		rows, err := r.executor(ctx).QueryContext(ctx, `
			SELECT key
			FROM api_keys
			WHERE deleted_at IS NULL
			  AND (
				group_id = ?
				OR EXISTS (
					SELECT 1
					FROM json_each(multi_group_routes)
					WHERE CAST(json_extract(value, '$.group_id') AS INTEGER) = ?
				)
			  )
			ORDER BY id DESC`,
			groupID, groupID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		keys := make([]string, 0)
		for rows.Next() {
			var key string
			if err := rows.Scan(&key); err != nil {
				return nil, err
			}
			keys = append(keys, key)
		}
		return keys, rows.Err()
	}
	rows, err := r.executor(ctx).QueryContext(ctx, `
		SELECT key
		FROM api_keys
		WHERE deleted_at IS NULL
		  AND (group_id = $1 OR multi_group_routes @> $2::jsonb)
		ORDER BY id DESC`,
		groupID, contains)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// IncrementQuotaUsed 使用 Ent 原子递增 quota_used 字段并返回新值
func (r *apiKeyRepository) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) (float64, error) {
	updated, err := r.client.APIKey.UpdateOneID(id).
		Where(apikey.DeletedAtIsNil()).
		AddQuotaUsed(amount).
		Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return 0, service.ErrAPIKeyNotFound
		}
		return 0, err
	}
	return updated.QuotaUsed, nil
}

// IncrementQuotaUsedAndGetState atomically increments quota_used, conditionally marks the key
// as quota_exhausted, and returns the latest quota state in one round trip.
func (r *apiKeyRepository) IncrementQuotaUsedAndGetState(ctx context.Context, id int64, amount float64) (*service.APIKeyQuotaUsageState, error) {
	query := `
		UPDATE api_keys
		SET
			quota_used = quota_used + $1,
			status = CASE
				WHEN quota > 0 AND quota_used + $1 >= quota THEN $2
				ELSE status
			END,
			updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING quota_used, quota, key, status
	`

	state := &service.APIKeyQuotaUsageState{}
	if err := scanSingleRow(ctx, r.sql, query, []any{amount, service.StatusAPIKeyQuotaExhausted, id}, &state.QuotaUsed, &state.Quota, &state.Key, &state.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrAPIKeyNotFound
		}
		return nil, err
	}
	return state, nil
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id int64, usedAt time.Time) error {
	affected, err := r.client.APIKey.Update().
		Where(apikey.IDEQ(id), apikey.DeletedAtIsNil()).
		SetLastUsedAt(usedAt).
		SetUpdatedAt(usedAt).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrAPIKeyNotFound
	}
	return nil
}

// IncrementRateLimitUsage atomically increments all rate limit usage counters and initializes
// window start times via COALESCE if not already set.
func (r *apiKeyRepository) IncrementRateLimitUsage(ctx context.Context, id int64, cost float64) error {
	_, err := r.sql.ExecContext(ctx, `
		UPDATE api_keys SET
			usage_5h = CASE WHEN window_5h_start IS NOT NULL AND window_5h_start + INTERVAL '5 hours' <= NOW() THEN $1 ELSE usage_5h + $1 END,
			usage_1d = CASE WHEN window_1d_start IS NOT NULL AND window_1d_start + INTERVAL '24 hours' <= NOW() THEN $1 ELSE usage_1d + $1 END,
			usage_7d = CASE WHEN window_7d_start IS NOT NULL AND window_7d_start + INTERVAL '7 days' <= NOW() THEN $1 ELSE usage_7d + $1 END,
			window_5h_start = CASE WHEN window_5h_start IS NULL OR window_5h_start + INTERVAL '5 hours' <= NOW() THEN NOW() ELSE window_5h_start END,
			window_1d_start = CASE WHEN window_1d_start IS NULL OR window_1d_start + INTERVAL '24 hours' <= NOW() THEN date_trunc('day', NOW()) ELSE window_1d_start END,
			window_7d_start = CASE WHEN window_7d_start IS NULL OR window_7d_start + INTERVAL '7 days' <= NOW() THEN date_trunc('day', NOW()) ELSE window_7d_start END,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`,
		cost, id)
	return err
}

// ResetRateLimitWindows resets expired rate limit windows atomically.
func (r *apiKeyRepository) ResetRateLimitWindows(ctx context.Context, id int64) error {
	_, err := r.sql.ExecContext(ctx, `
		UPDATE api_keys SET
			usage_5h = CASE WHEN window_5h_start IS NOT NULL AND window_5h_start + INTERVAL '5 hours' <= NOW() THEN 0 ELSE usage_5h END,
			window_5h_start = CASE WHEN window_5h_start IS NOT NULL AND window_5h_start + INTERVAL '5 hours' <= NOW() THEN NOW() ELSE window_5h_start END,
			usage_1d = CASE WHEN window_1d_start IS NOT NULL AND window_1d_start + INTERVAL '24 hours' <= NOW() THEN 0 ELSE usage_1d END,
			window_1d_start = CASE WHEN window_1d_start IS NOT NULL AND window_1d_start + INTERVAL '24 hours' <= NOW() THEN date_trunc('day', NOW()) ELSE window_1d_start END,
			usage_7d = CASE WHEN window_7d_start IS NOT NULL AND window_7d_start + INTERVAL '7 days' <= NOW() THEN 0 ELSE usage_7d END,
			window_7d_start = CASE WHEN window_7d_start IS NOT NULL AND window_7d_start + INTERVAL '7 days' <= NOW() THEN date_trunc('day', NOW()) ELSE window_7d_start END,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`,
		id)
	return err
}

// GetRateLimitData returns the current rate limit usage and window start times for an API key.
func (r *apiKeyRepository) GetRateLimitData(ctx context.Context, id int64) (result *service.APIKeyRateLimitData, err error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT usage_5h, usage_1d, usage_7d, window_5h_start, window_1d_start, window_7d_start
		FROM api_keys
		WHERE id = $1 AND deleted_at IS NULL`,
		id)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if !rows.Next() {
		return nil, service.ErrAPIKeyNotFound
	}
	data := &service.APIKeyRateLimitData{}
	if err := rows.Scan(&data.Usage5h, &data.Usage1d, &data.Usage7d, &data.Window5hStart, &data.Window1dStart, &data.Window7dStart); err != nil {
		return nil, err
	}
	return data, rows.Err()
}

func apiKeyEntityToService(m *dbent.APIKey) *service.APIKey {
	if m == nil {
		return nil
	}
	out := &service.APIKey{
		ID:                  m.ID,
		UserID:              m.UserID,
		Key:                 m.Key,
		Name:                m.Name,
		Status:              m.Status,
		IPWhitelist:         m.IPWhitelist,
		IPBlacklist:         m.IPBlacklist,
		LastUsedAt:          m.LastUsedAt,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
		GroupID:             m.GroupID,
		AccountPoolStrategy: service.AccountPoolStrategySharedOnly,
		Quota:               m.Quota,
		QuotaUsed:           m.QuotaUsed,
		ExpiresAt:           m.ExpiresAt,
		RateLimit5h:         m.RateLimit5h,
		RateLimit1d:         m.RateLimit1d,
		RateLimit7d:         m.RateLimit7d,
		Usage5h:             m.Usage5h,
		Usage1d:             m.Usage1d,
		Usage7d:             m.Usage7d,
		Window5hStart:       m.Window5hStart,
		Window1dStart:       m.Window1dStart,
		Window7dStart:       m.Window7dStart,
		ManagedSourceType:   m.ManagedSourceType,
		ManagedSourceID:     m.ManagedSourceID,
	}
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
		if allowed := m.Edges.User.Edges.AllowedGroups; len(allowed) > 0 {
			out.User.AllowedGroups = make([]int64, 0, len(allowed))
			for _, g := range allowed {
				if g != nil {
					out.User.AllowedGroups = append(out.User.AllowedGroups, g.ID)
				}
			}
		}
	}
	if m.Edges.Group != nil {
		out.Group = groupEntityToService(m.Edges.Group)
	}
	return out
}

func marshalAPIKeyMultiGroupRoutes(routes []domain.APIKeyMultiGroupRoute) (string, error) {
	if routes == nil {
		routes = []domain.APIKeyMultiGroupRoute{}
	} else {
		// Legacy model patterns remain readable for the explicit S91 migration,
		// but the repository must not write them back on ordinary API-key saves.
		sanitized := make([]domain.APIKeyMultiGroupRoute, len(routes))
		copy(sanitized, routes)
		for i := range sanitized {
			sanitized[i].ModelPatterns = nil
		}
		routes = sanitized
	}
	data, err := json.Marshal(routes)
	if err != nil {
		return "", fmt.Errorf("marshal api key multi-group routes: %w", err)
	}
	return string(data), nil
}

func apiKeyRouteContainsJSON(groupID int64) (string, error) {
	data, err := json.Marshal([]map[string]int64{{"group_id": groupID}})
	if err != nil {
		return "", fmt.Errorf("marshal api key route contains predicate: %w", err)
	}
	return string(data), nil
}

func (r *apiKeyRepository) setAPIKeyMultiGroupRoutes(ctx context.Context, id int64, routeJSON string) error {
	if r.isSQLite() {
		res, err := r.executor(ctx).ExecContext(ctx, `
			UPDATE api_keys
			SET multi_group_routes = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND deleted_at IS NULL`,
			routeJSON, id)
		if err != nil {
			return err
		}
		if affected, err := res.RowsAffected(); err == nil && affected == 0 {
			return service.ErrAPIKeyNotFound
		}
		return nil
	}
	res, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE api_keys
		SET multi_group_routes = $1::jsonb, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`,
		routeJSON, id)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return service.ErrAPIKeyNotFound
	}
	return nil
}

func (r *apiKeyRepository) loadAPIKeyMultiGroupRoutes(ctx context.Context, out *service.APIKey) error {
	if out == nil || out.ID <= 0 {
		return nil
	}
	var raw []byte
	err := scanSingleRow(ctx, r.executor(ctx), `
		SELECT multi_group_routes
		FROM api_keys
		WHERE id = $1 AND deleted_at IS NULL`,
		[]any{out.ID}, &raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return service.ErrAPIKeyNotFound
		}
		return err
	}
	if len(raw) == 0 {
		out.MultiGroupRoutes = []domain.APIKeyMultiGroupRoute{}
		return r.loadAPIKeyAccountPoolStrategy(ctx, out)
	}
	if err := json.Unmarshal(raw, &out.MultiGroupRoutes); err != nil {
		return fmt.Errorf("unmarshal api key multi-group routes: %w", err)
	}
	return r.loadAPIKeyAccountPoolStrategy(ctx, out)
}

func (r *apiKeyRepository) setAPIKeyAccountPoolStrategy(ctx context.Context, id int64, strategy string) error {
	strategy = service.NormalizeAccountPoolStrategy(strategy)
	if r.isSQLite() {
		res, err := r.executor(ctx).ExecContext(ctx, `
			UPDATE api_keys
			SET account_pool_strategy = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND deleted_at IS NULL`,
			strategy, id)
		if err != nil {
			return err
		}
		if affected, err := res.RowsAffected(); err == nil && affected == 0 {
			return service.ErrAPIKeyNotFound
		}
		return nil
	}
	res, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE api_keys
		SET account_pool_strategy = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`,
		strategy, id)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return service.ErrAPIKeyNotFound
	}
	return nil
}

func (r *apiKeyRepository) loadAPIKeyAccountPoolStrategy(ctx context.Context, out *service.APIKey) error {
	if out == nil || out.ID <= 0 {
		return nil
	}
	var raw string
	err := scanSingleRow(ctx, r.executor(ctx), `
		SELECT account_pool_strategy
		FROM api_keys
		WHERE id = $1 AND deleted_at IS NULL`,
		[]any{out.ID}, &raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return service.ErrAPIKeyNotFound
		}
		return err
	}
	out.AccountPoolStrategy = service.NormalizeAccountPoolStrategy(raw)
	return nil
}

func (r *apiKeyRepository) hydrateAPIKeyMultiGroupRouteGroups(ctx context.Context, out *service.APIKey) error {
	if out == nil || len(out.MultiGroupRoutes) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(out.MultiGroupRoutes))
	seen := make(map[int64]struct{}, len(out.MultiGroupRoutes))
	for _, route := range out.MultiGroupRoutes {
		if route.GroupID <= 0 {
			continue
		}
		if _, ok := seen[route.GroupID]; ok {
			continue
		}
		seen[route.GroupID] = struct{}{}
		ids = append(ids, route.GroupID)
	}
	if len(ids) == 0 {
		return nil
	}
	groups, err := clientFromContext(ctx, r.client).Group.Query().
		Where(group.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return err
	}
	byID := make(map[int64]*service.Group, len(groups))
	for i := range groups {
		g := groupEntityToService(groups[i])
		byID[g.ID] = g
	}
	out.MultiGroupRouteGroups = make([]*service.Group, 0, len(ids))
	for _, id := range ids {
		if g := byID[id]; g != nil {
			out.MultiGroupRouteGroups = append(out.MultiGroupRouteGroups, g)
		}
	}
	return nil
}

func (r *apiKeyRepository) apiKeyIDsByRelatedGroup(ctx context.Context, groupID int64) ([]int64, error) {
	contains, err := apiKeyRouteContainsJSON(groupID)
	if err != nil {
		return nil, err
	}
	if r.isSQLite() {
		rows, err := r.executor(ctx).QueryContext(ctx, `
			SELECT id
			FROM api_keys
			WHERE deleted_at IS NULL
			  AND (
				group_id = ?
				OR EXISTS (
					SELECT 1
					FROM json_each(multi_group_routes)
					WHERE CAST(json_extract(value, '$.group_id') AS INTEGER) = ?
				)
			  )
			ORDER BY id DESC`,
			groupID, groupID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		ids := make([]int64, 0)
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, rows.Err()
	}
	rows, err := r.executor(ctx).QueryContext(ctx, `
		SELECT id
		FROM api_keys
		WHERE deleted_at IS NULL
		  AND (group_id = $1 OR multi_group_routes @> $2::jsonb)
		ORDER BY id DESC`,
		groupID, contains)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func userEntityToService(u *dbent.User) *service.User {
	if u == nil {
		return nil
	}
	out := &service.User{
		ID:                         u.ID,
		Email:                      u.Email,
		Username:                   u.Username,
		Notes:                      u.Notes,
		PasswordHash:               u.PasswordHash,
		Role:                       u.Role,
		Balance:                    u.Balance,
		Concurrency:                u.Concurrency,
		Status:                     u.Status,
		ExcludeFromLeaderboard:     u.ExcludeFromLeaderboard,
		SignupSource:               u.SignupSource,
		RegisterIP:                 u.RegisterIP,
		LastLoginIP:                u.LastLoginIP,
		LastLoginAt:                u.LastLoginAt,
		LastActiveAt:               u.LastActiveAt,
		TotpSecretEncrypted:        u.TotpSecretEncrypted,
		TotpEnabled:                u.TotpEnabled,
		TotpEnabledAt:              u.TotpEnabledAt,
		BalanceNotifyEnabled:       u.BalanceNotifyEnabled,
		BalanceNotifyThresholdType: u.BalanceNotifyThresholdType,
		BalanceNotifyThreshold:     u.BalanceNotifyThreshold,
		TotalRecharged:             u.TotalRecharged,
		RPMLimit:                   u.RpmLimit,
		CreatedAt:                  u.CreatedAt,
		UpdatedAt:                  u.UpdatedAt,
	}
	// Parse extra emails JSON (supports both old []string and new []NotifyEmailEntry format)
	if u.BalanceNotifyExtraEmails != "" && u.BalanceNotifyExtraEmails != "[]" {
		out.BalanceNotifyExtraEmails = service.ParseNotifyEmails(u.BalanceNotifyExtraEmails)
	}
	return out
}

func groupEntityToService(g *dbent.Group) *service.Group {
	if g == nil {
		return nil
	}
	var modelPricing []service.ChannelModelPricing
	if len(g.ModelPricing) > 0 {
		if err := json.Unmarshal(g.ModelPricing, &modelPricing); err != nil {
			slog.Warn("group model_pricing unmarshal failed; falling back to channel/builtin pricing", "group_id", g.ID, "error", err)
			modelPricing = nil
		}
	}
	return &service.Group{
		ID:                              g.ID,
		Name:                            g.Name,
		Description:                     derefString(g.Description),
		Platform:                        g.Platform,
		RateMultiplier:                  g.RateMultiplier,
		IsExclusive:                     g.IsExclusive,
		Status:                          g.Status,
		Hydrated:                        true,
		DuplicateOperationID:            derefString(g.DuplicateOperationID),
		SubscriptionType:                g.SubscriptionType,
		RoutingScope:                    service.NormalizeGroupRoutingScope(g.RoutingScope, g.AllowImageGeneration),
		DailyLimitUSD:                   g.DailyLimitUsd,
		WeeklyLimitUSD:                  g.WeeklyLimitUsd,
		MonthlyLimitUSD:                 g.MonthlyLimitUsd,
		LongContextPricingEnabled:       g.LongContextPricingEnabled,
		ModelPricing:                    modelPricing,
		AllowImageGeneration:            g.AllowImageGeneration,
		ImageRateIndependent:            g.ImageRateIndependent,
		ImageRateMultiplier:             g.ImageRateMultiplier,
		ImagePrice1K:                    g.ImagePrice1k,
		ImagePrice2K:                    g.ImagePrice2k,
		ImagePrice4K:                    g.ImagePrice4k,
		DefaultValidityDays:             g.DefaultValidityDays,
		ClaudeCodeOnly:                  g.ClaudeCodeOnly,
		FallbackGroupID:                 g.FallbackGroupID,
		FallbackGroupIDOnInvalidRequest: g.FallbackGroupIDOnInvalidRequest,
		ModelRouting:                    g.ModelRouting,
		ModelRoutingEnabled:             g.ModelRoutingEnabled,
		ModelMatchPatterns:              service.NormalizeGroupModelMatchPatterns(g.ModelMatchPatterns),
		MCPXMLInject:                    g.McpXMLInject,
		SupportedModelScopes:            g.SupportedModelScopes,
		SortOrder:                       g.SortOrder,
		AllowMessagesDispatch:           g.AllowMessagesDispatch,
		AllowLive:                       g.AllowLive,
		RequireOAuthOnly:                g.RequireOauthOnly,
		RequirePrivacySet:               g.RequirePrivacySet,
		DefaultMappedModel:              g.DefaultMappedModel,
		MessagesDispatchModelConfig:     g.MessagesDispatchModelConfig,
		ModelsListConfig:                g.ModelsListConfig,
		RPMLimit:                        g.RpmLimit,
		PeakRateEnabled:                 g.PeakRateEnabled,
		PeakStart:                       g.PeakStart,
		PeakEnd:                         g.PeakEnd,
		PeakRateMultiplier:              g.PeakRateMultiplier,
		CreatedAt:                       g.CreatedAt,
		UpdatedAt:                       g.UpdatedAt,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

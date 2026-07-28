package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const stickySessionPrefix = "sticky_session:"
const liveCallPrefix = "live:call:"

type gatewayCache struct {
	rdb *redis.Client
}

var _ service.LiveCallStore = (*gatewayCache)(nil)

func NewGatewayCache(rdb *redis.Client) service.GatewayCache {
	return &gatewayCache{rdb: rdb}
}

// buildSessionKey 构建 session key，包含 groupID 实现分组隔离
// 格式: sticky_session:{groupID}:{sessionHash}
func buildSessionKey(groupID int64, sessionHash string) string {
	return fmt.Sprintf("%s%d:%s", stickySessionPrefix, groupID, sessionHash)
}

func (c *gatewayCache) GetSessionAccountID(ctx context.Context, groupID int64, sessionHash string) (int64, error) {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Get(ctx, key).Int64()
}

func (c *gatewayCache) SetSessionAccountID(ctx context.Context, groupID int64, sessionHash string, accountID int64, ttl time.Duration) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Set(ctx, key, accountID, ttl).Err()
}

func (c *gatewayCache) RefreshSessionTTL(ctx context.Context, groupID int64, sessionHash string, ttl time.Duration) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// DeleteSessionAccountID 删除粘性会话与账号的绑定关系。
// 当检测到绑定的账号不可用（如状态错误、禁用、不可调度等）时调用，
// 以便下次请求能够重新选择可用账号。
//
// DeleteSessionAccountID removes the sticky session binding for the given session.
// Called when the bound account becomes unavailable (e.g., error status, disabled,
// or unschedulable), allowing subsequent requests to select a new available account.
func (c *gatewayCache) DeleteSessionAccountID(ctx context.Context, groupID int64, sessionHash string) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Del(ctx, key).Err()
}

func liveCallKey(callHash string) string { return liveCallPrefix + callHash }

func HashLiveCallID(callID string) string {
	sum := sha256.Sum256([]byte(callID))
	return hex.EncodeToString(sum[:])
}

var liveClaimScript = redis.NewScript(`
local key = KEYS[1]
if redis.call('EXISTS', key) == 0 then return 0 end
local current = redis.call('HGET', key, 'controller')
local currentOwner = redis.call('HGET', key, 'controller_owner')
local target = ARGV[1]
local owner = ARGV[2]
if current == 'closed' then return 0 end
if target == 'observer' and current ~= 'pending' and current ~= 'observer' then return 0 end
if target == 'proxy' and current ~= 'pending' and current ~= 'observer' and (current ~= 'proxy' or currentOwner ~= owner) then return 0 end
redis.call('HSET', key, 'controller', target, 'controller_owner', owner)
return 1
`)

var liveReleaseControllerScript = redis.NewScript(`
local key = KEYS[1]
if redis.call('HGET', key, 'controller') ~= 'proxy' or redis.call('HGET', key, 'controller_owner') ~= ARGV[1] then return 0 end
redis.call('HSET', key, 'controller', 'pending', 'controller_owner', '')
return 1
`)

var liveCloseScript = redis.NewScript(`
local key = KEYS[1]
if redis.call('EXISTS', key) == 0 or redis.call('HGET', key, 'controller') == 'closed' then return 0 end
redis.call('HSET', key, 'controller', 'closed', 'controller_owner', '')
redis.call('EXPIRE', key, ARGV[1])
return 1
`)

func (c *gatewayCache) SaveLiveCall(ctx context.Context, record *service.LiveCallRecord, ttl time.Duration) error {
	if record == nil || record.CallHash == "" || record.CallID == "" {
		return fmt.Errorf("invalid live call record")
	}
	values := map[string]any{
		"call_id": record.CallID, "account_id": record.AccountID, "api_key_id": record.APIKeyID,
		"user_id": record.UserID, "group_id": record.GroupID, "subscription_id": record.SubscriptionID,
		"lease_id": record.LeaseID, "model": record.Model, "created_at": record.CreatedAt.UnixMilli(),
		"expires_at": record.ExpiresAt.UnixMilli(), "controller": record.Controller,
		"controller_owner": record.ControllerOwner, "user_agent": record.UserAgent,
		"ip_address": record.IPAddress, "inbound_endpoint": record.InboundEndpoint,
		"attestation_ciphertext": record.AttestationCiphertext,
	}
	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, liveCallKey(record.CallHash), values)
	pipe.Expire(ctx, liveCallKey(record.CallHash), ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *gatewayCache) GetLiveCall(ctx context.Context, callHash string) (*service.LiveCallRecord, error) {
	values, err := c.rdb.HGetAll(ctx, liveCallKey(callHash)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, service.ErrLiveCallNotFound
	}
	parseInt := func(name string) int64 { value, _ := strconv.ParseInt(values[name], 10, 64); return value }
	return &service.LiveCallRecord{
		CallID: values["call_id"], CallHash: callHash, AccountID: parseInt("account_id"), APIKeyID: parseInt("api_key_id"),
		UserID: parseInt("user_id"), GroupID: parseInt("group_id"), SubscriptionID: parseInt("subscription_id"),
		LeaseID: values["lease_id"], Model: values["model"], CreatedAt: time.UnixMilli(parseInt("created_at")),
		ExpiresAt: time.UnixMilli(parseInt("expires_at")), Controller: values["controller"], ControllerOwner: values["controller_owner"],
		UserAgent: values["user_agent"], IPAddress: values["ip_address"], InboundEndpoint: values["inbound_endpoint"],
		AttestationCiphertext: values["attestation_ciphertext"],
	}, nil
}

func (c *gatewayCache) ClaimLiveController(ctx context.Context, callHash, controller, owner string) (bool, error) {
	value, err := liveClaimScript.Run(ctx, c.rdb, []string{liveCallKey(callHash)}, controller, owner).Int()
	return value == 1, err
}

func (c *gatewayCache) ReleaseLiveController(ctx context.Context, callHash, owner string) (bool, error) {
	value, err := liveReleaseControllerScript.Run(ctx, c.rdb, []string{liveCallKey(callHash)}, owner).Int()
	return value == 1, err
}

func (c *gatewayCache) GetLiveController(ctx context.Context, callHash string) (string, error) {
	value, err := c.rdb.HGet(ctx, liveCallKey(callHash), "controller").Result()
	if err == redis.Nil {
		return "", service.ErrLiveCallNotFound
	}
	return value, err
}

func (c *gatewayCache) MarkLiveCallClosed(ctx context.Context, callHash string, ttl time.Duration) (bool, error) {
	value, err := liveCloseScript.Run(ctx, c.rdb, []string{liveCallKey(callHash)}, int64(ttl.Seconds())).Int()
	return value == 1, err
}

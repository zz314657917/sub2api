package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	apiKeyRouteBreakerPrefix       = "apikey:route:breaker:"
	apiKeyRouteBreakerInactiveTTL  = 30 * 60
	apiKeyRouteBreakerRetentionTTL = 24 * 60 * 60
	apiKeyRouteBreakerProbeTTL     = 30
)

func apiKeyRouteBreakerKey(key service.APIKeyRouteBreakerKey) string {
	return fmt.Sprintf("%s{%d:%s:%s}", apiKeyRouteBreakerPrefix, key.GroupID, key.RoutingScope, key.ModelDigest)
}

const apiKeyRouteBreakerAcquireScript = `
local now = tonumber(redis.call('TIME')[1])
local generation = tonumber(redis.call('HGET', KEYS[1], 'generation'))
if not generation then
  return {1, 0, 0, 0}
end
local open_until = tonumber(redis.call('HGET', KEYS[1], 'open_until')) or 0
if open_until > now then
  return {0, generation, 0, 0}
end
if open_until > 0 then
  local lease_until = tonumber(redis.call('HGET', KEYS[1], 'lease_until')) or 0
  if lease_until > now then
    return {0, generation, 0, 0}
  end
  local token = (tonumber(redis.call('HGET', KEYS[1], 'lease_token')) or 0) + 1
  redis.call('HSET', KEYS[1], 'lease_token', token, 'lease_until', now + ARGV[2])
  return {1, generation, token, 1}
end
return {1, generation, 0, 0}
`

const apiKeyRouteBreakerFailureScript = `
local now = tonumber(redis.call('TIME')[1])
local generation = tonumber(redis.call('HGET', KEYS[1], 'generation')) or 0
if redis.call('EXISTS', KEYS[1]) == 0 then
  if tonumber(ARGV[1]) ~= 0 or tonumber(ARGV[2]) == 1 then return 0 end
  redis.call('HSET', KEYS[1], 'generation', 0, 'failures', 1, 'level', 0, 'open_until', 0, 'lease_token', 0, 'lease_until', 0)
  redis.call('EXPIRE', KEYS[1], ARGV[8])
  return 1
end
if generation ~= tonumber(ARGV[1]) then return 0 end
local half_open = tonumber(ARGV[2])
if half_open == 1 then
  local token = tonumber(redis.call('HGET', KEYS[1], 'lease_token')) or 0
  local lease_until = tonumber(redis.call('HGET', KEYS[1], 'lease_until')) or 0
  if token ~= tonumber(ARGV[3]) or lease_until < now then return 0 end
end
local level = tonumber(redis.call('HGET', KEYS[1], 'level')) or 0
local failures = tonumber(redis.call('HGET', KEYS[1], 'failures')) or 0
if half_open == 1 then
  level = math.min(level + 1, 4)
else
  failures = failures + 1
  if failures < 3 then
    redis.call('HSET', KEYS[1], 'failures', failures)
    redis.call('EXPIRE', KEYS[1], ARGV[8])
    return 1
  end
  level = 1
end
local cooldowns = {tonumber(ARGV[4]), tonumber(ARGV[5]), tonumber(ARGV[6]), tonumber(ARGV[7])}
redis.call('HSET', KEYS[1], 'generation', generation + 1, 'failures', 0, 'level', level, 'open_until', now + cooldowns[level], 'lease_until', 0)
redis.call('EXPIRE', KEYS[1], ARGV[9])
return 1
`

const apiKeyRouteBreakerSuccessScript = `
local now = tonumber(redis.call('TIME')[1])
local generation = tonumber(redis.call('HGET', KEYS[1], 'generation')) or 0
if redis.call('EXISTS', KEYS[1]) == 0 then return 0 end
if generation ~= tonumber(ARGV[1]) then return 0 end
if tonumber(ARGV[2]) == 1 then
  local token = tonumber(redis.call('HGET', KEYS[1], 'lease_token')) or 0
  local lease_until = tonumber(redis.call('HGET', KEYS[1], 'lease_until')) or 0
  if token ~= tonumber(ARGV[3]) or lease_until < now then return 0 end
end
redis.call('HSET', KEYS[1], 'generation', generation + 1, 'failures', 0, 'level', 0, 'open_until', 0, 'lease_until', 0)
redis.call('EXPIRE', KEYS[1], ARGV[4])
return 1
`

const apiKeyRouteBreakerReleaseScript = `
local generation = tonumber(redis.call('HGET', KEYS[1], 'generation')) or 0
if redis.call('EXISTS', KEYS[1]) == 0 then return 0 end
if generation ~= tonumber(ARGV[1]) then return 0 end
local token = tonumber(redis.call('HGET', KEYS[1], 'lease_token')) or 0
if token ~= tonumber(ARGV[3]) then return 0 end
redis.call('HSET', KEYS[1], 'lease_until', 0)
redis.call('EXPIRE', KEYS[1], ARGV[4])
return 1
`

func (c *apiKeyCache) AcquireAPIKeyRouteBreaker(ctx context.Context, routeKey service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
	if c == nil || c.rdb == nil || routeKey.GroupID <= 0 || routeKey.RoutingScope == "" || routeKey.ModelDigest == "" {
		return nil, nil
	}
	values, err := c.rdb.Eval(ctx, apiKeyRouteBreakerAcquireScript, []string{apiKeyRouteBreakerKey(routeKey)}, apiKeyRouteBreakerRetentionTTL, apiKeyRouteBreakerProbeTTL).Slice()
	if err != nil {
		return nil, err
	}
	if len(values) != 4 || redisResultInt(values[0]) == 0 {
		return nil, nil
	}
	return &service.APIKeyRouteBreakerLease{
		Key:        routeKey,
		Generation: redisResultInt(values[1]),
		ProbeToken: redisResultInt(values[2]),
		HalfOpen:   redisResultInt(values[3]) == 1,
	}, nil
}

func (c *apiKeyCache) RecordAPIKeyRouteBreakerSuccess(ctx context.Context, lease service.APIKeyRouteBreakerLease) error {
	return c.evalAPIKeyRouteBreaker(ctx, apiKeyRouteBreakerSuccessScript, lease, apiKeyRouteBreakerInactiveTTL)
}

func (c *apiKeyCache) RecordAPIKeyRouteBreakerFailure(ctx context.Context, lease service.APIKeyRouteBreakerLease) error {
	return c.evalAPIKeyRouteBreaker(ctx, apiKeyRouteBreakerFailureScript, lease, 30, 120, 600, 1800, apiKeyRouteBreakerInactiveTTL, apiKeyRouteBreakerRetentionTTL)
}

func (c *apiKeyCache) ReleaseAPIKeyRouteBreakerProbe(ctx context.Context, lease service.APIKeyRouteBreakerLease) error {
	if !lease.HalfOpen {
		return nil
	}
	return c.evalAPIKeyRouteBreaker(ctx, apiKeyRouteBreakerReleaseScript, lease, apiKeyRouteBreakerRetentionTTL)
}

func (c *apiKeyCache) evalAPIKeyRouteBreaker(ctx context.Context, script string, lease service.APIKeyRouteBreakerLease, extraArgs ...int) error {
	if c == nil || c.rdb == nil || lease.Key.GroupID <= 0 || lease.Key.RoutingScope == "" || lease.Key.ModelDigest == "" {
		return nil
	}
	args := []interface{}{lease.Generation, boolToRedisInt(lease.HalfOpen), lease.ProbeToken}
	for _, value := range extraArgs {
		args = append(args, value)
	}
	return c.rdb.Eval(ctx, script, []string{apiKeyRouteBreakerKey(lease.Key)}, args...).Err()
}

func redisResultInt(value interface{}) int64 {
	switch value := value.(type) {
	case int64:
		return value
	case string:
		parsed, _ := strconv.ParseInt(value, 10, 64)
		return parsed
	case []byte:
		parsed, _ := strconv.ParseInt(string(value), 10, 64)
		return parsed
	default:
		return 0
	}
}

func boolToRedisInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

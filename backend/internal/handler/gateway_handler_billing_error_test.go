package handler

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBillingErrorDetails_MapsGroupRPMExceededToTooManyRequests(t *testing.T) {
	status, code, msg, retryAfter := billingErrorDetails(service.ErrGroupRPMExceeded)
	require.Equal(t, http.StatusTooManyRequests, status)
	require.Equal(t, "rate_limit_exceeded", code)
	require.NotEmpty(t, msg)
	require.Greater(t, retryAfter, 0, "RPM exceeded should return positive Retry-After")
	require.LessOrEqual(t, retryAfter, 60)
}

func TestBillingErrorDetails_MapsUserRPMExceededToTooManyRequests(t *testing.T) {
	status, code, msg, retryAfter := billingErrorDetails(service.ErrUserRPMExceeded)
	require.Equal(t, http.StatusTooManyRequests, status)
	require.Equal(t, "rate_limit_exceeded", code)
	require.NotEmpty(t, msg)
	require.Greater(t, retryAfter, 0, "RPM exceeded should return positive Retry-After")
	require.LessOrEqual(t, retryAfter, 60)
}

func TestBillingErrorDetails_APIKeyRateLimitStillMaps(t *testing.T) {
	// 回归保护：加 RPM 分支后不应影响已有 APIKey rate limit 的映射。
	for _, err := range []error{
		service.ErrAPIKeyRateLimit5hExceeded,
		service.ErrAPIKeyRateLimit1dExceeded,
		service.ErrAPIKeyRateLimit7dExceeded,
	} {
		status, code, _, _ := billingErrorDetails(err)
		require.Equal(t, http.StatusTooManyRequests, status, "status for %v", err)
		require.Equal(t, "rate_limit_exceeded", code)
	}
}

func TestBillingErrorDetails_BillingServiceUnavailableMapsTo503(t *testing.T) {
	status, code, _, retryAfter := billingErrorDetails(service.ErrBillingServiceUnavailable)
	require.Equal(t, http.StatusServiceUnavailable, status)
	require.Equal(t, "billing_service_error", code)
	require.Equal(t, 0, retryAfter, "non-RPM errors should not set Retry-After")
}

func TestCloneAPIKeyWithGroupKeepsRPMOverrideScoped(t *testing.T) {
	groupID := int64(3)
	zero := 0
	key := &service.APIKey{GroupID: &groupID, User: &service.User{
		RPMLimit: 60, UserGroupRPMOverride: &zero, UserGroupRPMOverrideGroupID: groupID,
	}}
	same := cloneAPIKeyWithGroup(key, &service.Group{ID: groupID})
	require.Same(t, key.User, same.User)

	fallback := cloneAPIKeyWithGroup(key, &service.Group{ID: 4})
	require.Equal(t, int64(4), *fallback.GroupID)
	require.NotSame(t, key.User, fallback.User)
	require.Nil(t, fallback.User.UserGroupRPMOverride)
	require.Zero(t, fallback.User.UserGroupRPMOverrideGroupID)
	require.Equal(t, 60, fallback.User.RPMLimit)
	require.Equal(t, groupID, key.User.UserGroupRPMOverrideGroupID)
	require.Equal(t, &zero, key.User.UserGroupRPMOverride)
}

func TestBillingErrorDetails_InsufficientBalanceIncludesRechargeURL(t *testing.T) {
	status, code, msg, _ := billingErrorDetails(service.ErrInsufficientBalance)
	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, "billing_error", code)
	require.Equal(t, "账户余额不足，充值请前往：https://ai.3zapi.com", msg)
}

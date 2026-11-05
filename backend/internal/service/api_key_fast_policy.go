package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	APIKeyOpenAIFastPolicyPassthrough = "passthrough"
	APIKeyOpenAIFastPolicyForce       = "force"
	APIKeyOpenAIFastPolicyFilter      = "filter"
)

var ErrAPIKeyOpenAIFastPolicyInvalid = infraerrors.BadRequest(
	"API_KEY_OPENAI_FAST_POLICY_INVALID", "openai_fast_policy must be passthrough, force or filter")

func IsValidAPIKeyOpenAIFastPolicy(policy string) bool {
	switch policy {
	case APIKeyOpenAIFastPolicyPassthrough, APIKeyOpenAIFastPolicyForce, APIKeyOpenAIFastPolicyFilter:
		return true
	default:
		return false
	}
}

// NormalizeAPIKeyOpenAIFastPolicy keeps old empty records compatible.
func NormalizeAPIKeyOpenAIFastPolicy(policy string) string {
	if IsValidAPIKeyOpenAIFastPolicy(policy) {
		return policy
	}
	return APIKeyOpenAIFastPolicyPassthrough
}

// openAIRequestFastOverride reads only authenticated request context. Group
// enforcement wins over the key; callers must apply global policy afterwards.
func openAIRequestFastOverride(ctx context.Context, account *Account, tier string) string {
	if ctx == nil || account == nil || account.Platform != PlatformOpenAI {
		return APIKeyOpenAIFastPolicyPassthrough
	}
	if openAIGroupForcesFast(ctx, account) {
		return APIKeyOpenAIFastPolicyForce
	}
	policy, _ := ctx.Value(ctxkey.APIKeyOpenAIFastPolicy).(string)
	switch policy {
	case APIKeyOpenAIFastPolicyForce:
		return APIKeyOpenAIFastPolicyForce
	case APIKeyOpenAIFastPolicyFilter:
		if normalizedOpenAIServiceTierValue(tier) == OpenAIFastTierPriority {
			return APIKeyOpenAIFastPolicyFilter
		}
	}
	return APIKeyOpenAIFastPolicyPassthrough
}

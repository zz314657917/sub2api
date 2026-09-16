package middleware

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const ContextKeyAPIKeyFirstResponseDispatcher ContextKey = "api_key_first_response_dispatcher"
const ContextKeyAPIKeyFirstResponsePristine ContextKey = "api_key_first_response_pristine"

func APIKeyFirstResponsePristine(c *gin.Context) (*service.APIKey, bool) {
	if c == nil {
		return nil, false
	}
	value, ok := c.Get(string(ContextKeyAPIKeyFirstResponsePristine))
	apiKey, ok := value.(*service.APIKey)
	return apiKey, ok && apiKey != nil
}

type APIKeyFirstResponseOutcome struct {
	Success bool
	Failure bool
}

// APIKeyFirstResponseOutcomeForContext uses the established stream marker and
// upstream-status classifier even when the HTTP status was already committed
// as 200. It is deliberately shared with the dispatcher so post-commit stream
// failures cool their own route without enabling replay.
func APIKeyFirstResponseOutcomeForContext(c *gin.Context) APIKeyFirstResponseOutcome {
	if c == nil {
		return APIKeyFirstResponseOutcome{}
	}
	if status := c.Writer.Status(); status >= 400 && status < 500 && status != 429 {
		return APIKeyFirstResponseOutcome{}
	}
	if routeBreakerFailureStatus(c) {
		return APIKeyFirstResponseOutcome{Failure: true}
	}
	if c.Writer.Status() < 400 {
		return APIKeyFirstResponseOutcome{Success: true}
	}
	return APIKeyFirstResponseOutcome{}
}

// PrepareAPIKeyFirstResponseAttempt rebuilds all route-dependent context from
// the pristine auth key. It never uses a failed attempt as its base.
func PrepareAPIKeyFirstResponseAttempt(c *gin.Context, apiKeyService *service.APIKeyService, pristine *service.APIKey, requestedModel string, attempted map[int64]struct{}, attemptCtx context.Context) (*service.APIKey, bool) {
	if c == nil || c.Request == nil || pristine == nil || apiKeyService == nil {
		return nil, false
	}
	delete(c.Keys, contextKeyAPIKeyModelResolution)
	delete(c.Keys, string(ContextKeySubscription))
	delete(c.Keys, string(ContextKeyAPIKeyRouteCooldown))
	delete(c.Keys, string(ContextKeyAPIKeyRouteBreakerStatus))
	forcePlatform, _ := GetForcePlatformFromContext(c)
	c.Request = c.Request.WithContext(ctxkey.WithKeyRouteFirstResponseAttempt(attemptCtx))
	resolved := apiKeyService.ResolveForFirstResponseAttempt(c.Request.Context(), pristine, c.Request.URL.Path, forcePlatform, requestedModel, attempted)
	if resolved == nil {
		return nil, false
	}
	// Selection may have acquired a half-open probe. Any local eligibility or
	// billing rejection must release it without recording an upstream failure.
	release := func() { FinalizeAPIKeyFirstResponseAttempt(apiKeyService, resolved, APIKeyFirstResponseOutcome{}) }
	if abortIfAPIKeyGroupUnavailable(c, resolved) || abortIfAPIKeyGroupNotAllowed(c, resolved) {
		release()
		return nil, false
	}
	SetAPIKeyContext(c, resolved)
	if shouldEnforceDeferredGroupBilling(c) {
		subscriptionService, _ := subscriptionServiceFromContext(c)
		if !enforceGroupBilling(c, resolved, nil, subscriptionService, false) {
			release()
			return nil, false
		}
	} else if !refreshSubscriptionContextForResolvedAPIKey(c, resolved) {
		release()
		return nil, false
	}
	cacheAPIKeyModelResolution(c, resolved, requestedModel)
	return resolved, true
}

// FinalizeAPIKeyFirstResponseAttempt records exactly one route-health outcome
// with an uncancelled bounded cleanup context.
func FinalizeAPIKeyFirstResponseAttempt(apiKeyService *service.APIKeyService, apiKey *service.APIKey, outcome APIKeyFirstResponseOutcome) {
	if apiKeyService == nil || apiKey == nil {
		return
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if outcome.Failure {
		if apiKey.GroupID != nil {
			// A default fallback need not have an explicit route record. Zero
			// selects the service's existing default cooldown in that case.
			seconds, _ := apiKey.RouteCooldownSeconds(*apiKey.GroupID)
			apiKeyService.MarkRouteGroupCooldown(cleanupCtx, apiKey, *apiKey.GroupID, seconds)
		}
		apiKeyService.RecordAPIKeyRouteBreakerFailure(cleanupCtx, apiKey)
		return
	}
	if outcome.Success {
		if apiKey.GroupID != nil {
			apiKeyService.ClearRouteGroupCooldown(cleanupCtx, apiKey, *apiKey.GroupID)
		}
		apiKeyService.RecordAPIKeyRouteBreakerSuccess(cleanupCtx, apiKey)
		return
	}
	apiKeyService.ReleaseAPIKeyRouteBreakerProbe(cleanupCtx, apiKey)
}

func IsAPIKeyFirstResponseDispatcher(c *gin.Context) bool {
	if c == nil {
		return false
	}
	value, _ := c.Get(string(ContextKeyAPIKeyFirstResponseDispatcher))
	marked, _ := value.(bool)
	return marked
}

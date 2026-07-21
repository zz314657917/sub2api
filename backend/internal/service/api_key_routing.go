package service

import (
	"math/rand/v2"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/domain"
)

// ResolveForRequest returns an API key copy whose Group/GroupID matches the best
// enabled multi-group route for this request. If no route matches, it returns k.
func (k *APIKey) ResolveForRequest(path, forcePlatform string) *APIKey {
	return k.ResolveForRequestWithGroupSkipper(path, forcePlatform, nil)
}

// ResolveForRequestWithGroupSkipper behaves like ResolveForRequest, but skips
// route candidates whose group ID matches skipGroup. It is used by the service
// layer to apply short-lived route cooldowns without mutating route config.
func (k *APIKey) ResolveForRequestWithGroupSkipper(path, forcePlatform string, skipGroup func(groupID int64) bool) *APIKey {
	return k.resolveForRequest(path, forcePlatform, "", false, false, skipGroup)
}

// ResolveForModelRequest returns an API key copy whose Group/GroupID matches
// the best route for the already-parsed request model and image intent.
func (k *APIKey) ResolveForModelRequest(path, forcePlatform, requestedModel string, imageIntent bool) *APIKey {
	return k.ResolveForModelRequestWithGroupSkipper(path, forcePlatform, requestedModel, imageIntent, nil)
}

// ResolveForModelRequestWithGroupSkipper is the model-aware variant used after
// gateway handlers parse the request body. If no model-specific route matches,
// it falls back only when the default group remains compatible with the request.
func (k *APIKey) ResolveForModelRequestWithGroupSkipper(path, forcePlatform, requestedModel string, imageIntent bool, skipGroup func(groupID int64) bool) *APIKey {
	return k.resolveForRequest(path, forcePlatform, requestedModel, imageIntent, true, skipGroup)
}

func (k *APIKey) resolveForRequest(path, forcePlatform, requestedModel string, imageIntent bool, modelAware bool, skipGroup func(groupID int64) bool) *APIKey {
	if k == nil {
		return k
	}
	if len(k.MultiGroupRoutes) == 0 {
		if modelAware && k.GroupID != nil && !k.canFallbackToDefaultGroup(path, forcePlatform, requestedModel, imageIntent) {
			return nil
		}
		return k
	}
	selected := k.selectRouteGroup(path, forcePlatform, requestedModel, imageIntent, modelAware, skipGroup)
	if selected == nil {
		if modelAware && !k.canFallbackToDefaultGroup(path, forcePlatform, requestedModel, imageIntent) {
			return nil
		}
		return k
	}
	if k.Group != nil && k.Group.ID == selected.ID {
		return k
	}
	clone := *k
	groupID := selected.ID
	clone.GroupID = &groupID
	clone.Group = selected
	if clone.User != nil {
		userCopy := *clone.User
		userCopy.UserGroupRPMOverride = nil
		clone.User = &userCopy
	}
	return &clone
}

func (k *APIKey) selectRouteGroup(path, forcePlatform, requestedModel string, imageIntent bool, modelAware bool, skipGroup func(groupID int64) bool) *Group {
	groups := k.routeGroupsByID()
	if len(groups) == 0 {
		return nil
	}
	platforms := preferredPlatformsForRequest(path, forcePlatform, requestedModel, modelAware)
	routingScope := RoutingScopeForRequest(path, requestedModel, imageIntent)
	if modelAware && (strings.TrimSpace(requestedModel) != "" || imageIntent) {
		if imageIntent {
			candidates := k.routeCandidates(groups, platforms, skipGroup, requestedModel, imageIntent, routingScope, true, true, true)
			if len(candidates) == 0 {
				candidates = k.routeCandidates(groups, nil, skipGroup, requestedModel, imageIntent, routingScope, true, true, true)
			}
			if len(candidates) > 0 {
				return selectBestRouteGroup(candidates)
			}
		}
		candidates := k.routeCandidates(groups, platforms, skipGroup, requestedModel, imageIntent, routingScope, true, true, false)
		if len(candidates) == 0 {
			candidates = k.routeCandidates(groups, nil, skipGroup, requestedModel, imageIntent, routingScope, true, true, false)
		}
		if len(candidates) > 0 {
			return selectBestRouteGroup(candidates)
		}
	}
	candidates := k.routeCandidates(groups, platforms, skipGroup, requestedModel, imageIntent, routingScope, modelAware, false, false)
	if len(candidates) == 0 {
		candidates = k.routeCandidates(groups, nil, skipGroup, requestedModel, imageIntent, routingScope, modelAware, false, false)
	}
	if len(candidates) == 0 {
		return nil
	}
	return selectBestRouteGroup(candidates)
}

func (k *APIKey) canFallbackToDefaultGroup(path, forcePlatform, requestedModel string, imageIntent bool) bool {
	if k == nil || k.GroupID == nil || !IsGroupContextValid(k.Group) || !k.Group.IsActive() || *k.GroupID != k.Group.ID {
		return false
	}
	platforms := preferredPlatformsForRequest(path, forcePlatform, requestedModel, true)
	if len(platforms) > 0 && !containsString(platforms, k.Group.Platform) {
		return false
	}
	routingScope := RoutingScopeForRequest(path, requestedModel, imageIntent)
	if !apiKeyRouteMatchesGroupScope(k.Group, routingScope) {
		return false
	}
	if !k.Group.MatchesModel(requestedModel) {
		return false
	}

	hasExplicitRoute := false
	for _, route := range k.MultiGroupRoutes {
		if !route.Enabled || route.GroupID != k.Group.ID {
			continue
		}
		if apiKeyRouteMatchesModelRequest(route, k.Group, requestedModel, imageIntent) {
			return true
		}
		hasExplicitRoute = true
	}
	return !hasExplicitRoute
}

func preferredPlatformsForRequest(path, forcePlatform, requestedModel string, modelAware bool) []string {
	platforms := preferredPlatformsForPath(path, forcePlatform)
	if modelAware && strings.TrimSpace(forcePlatform) == "" {
		if modelPlatforms := preferredPlatformsForModel(requestedModel); len(modelPlatforms) > 0 {
			return modelPlatforms
		}
	}
	return platforms
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func selectBestRouteGroup(candidates []apiKeyRouteCandidate) *Group {
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].priority < candidates[j].priority
	})
	bestPriority := candidates[0].priority
	best := candidates[:0]
	for _, candidate := range candidates {
		if candidate.priority != bestPriority {
			break
		}
		best = append(best, candidate)
	}
	return pickWeightedRouteGroup(best)
}

func (k *APIKey) routeGroupsByID() map[int64]*Group {
	groups := make(map[int64]*Group, len(k.MultiGroupRouteGroups)+1)
	if IsGroupContextValid(k.Group) {
		groups[k.Group.ID] = k.Group
	}
	for _, group := range k.MultiGroupRouteGroups {
		if IsGroupContextValid(group) {
			groups[group.ID] = group
		}
	}
	return groups
}

func (k *APIKey) RouteCooldownSeconds(groupID int64) (int, bool) {
	if k == nil || groupID <= 0 {
		return 0, false
	}
	cooldownSeconds := 0
	found := false
	for _, route := range k.MultiGroupRoutes {
		if route.GroupID != groupID {
			continue
		}
		if !route.Enabled {
			continue
		}
		found = true
		if route.CooldownSeconds <= 0 {
			cooldownSeconds = max(cooldownSeconds, apiKeyRouteDefaultCooldown)
			continue
		}
		cooldownSeconds = max(cooldownSeconds, route.CooldownSeconds)
	}
	if !found {
		return 0, false
	}
	return cooldownSeconds, true
}

type apiKeyRouteCandidate struct {
	group    *Group
	priority int
	weight   int
}

func (k *APIKey) routeCandidates(groups map[int64]*Group, platforms []string, skipGroup func(groupID int64) bool, requestedModel string, imageIntent bool, routingScope string, modelAware bool, explicitRulesOnly bool, imageOnlyRulesOnly bool) []apiKeyRouteCandidate {
	platformSet := make(map[string]struct{}, len(platforms))
	for _, platform := range platforms {
		if platform != "" {
			platformSet[platform] = struct{}{}
		}
	}
	candidates := make([]apiKeyRouteCandidate, 0, len(k.MultiGroupRoutes))
	for _, route := range k.MultiGroupRoutes {
		if !route.Enabled {
			continue
		}
		if skipGroup != nil && skipGroup(route.GroupID) {
			continue
		}
		group := groups[route.GroupID]
		if group == nil || !group.IsActive() {
			continue
		}
		if !apiKeyRouteMatchesGroupScope(group, routingScope) {
			continue
		}
		if len(platformSet) > 0 {
			if _, ok := platformSet[group.Platform]; !ok {
				continue
			}
		}
		if modelAware {
			hasModelRules := apiKeyRouteHasModelRules(route, group)
			if explicitRulesOnly && !hasModelRules {
				continue
			}
			if imageOnlyRulesOnly && !route.ImageOnly {
				continue
			}
			if !apiKeyRouteMatchesModelRequest(route, group, requestedModel, imageIntent) {
				continue
			}
		}
		priority := route.Priority
		if priority <= 0 {
			priority = apiKeyRouteDefaultPriority
		}
		weight := route.Weight
		if weight <= 0 {
			weight = apiKeyRouteDefaultWeight
		}
		candidates = append(candidates, apiKeyRouteCandidate{
			group:    group,
			priority: priority,
			weight:   weight,
		})
	}
	return candidates
}

func apiKeyRouteHasModelRules(route domain.APIKeyMultiGroupRoute, group *Group) bool {
	return (group != nil && len(group.ModelMatchPatterns) > 0) || route.ImageOnly || route.TextOnly
}

func apiKeyRouteMatchesGroupScope(group *Group, routingScope string) bool {
	if group == nil {
		return true
	}
	scope := NormalizeGroupRoutingScope(routingScope, false)
	if scope == "" {
		scope = GroupRoutingScopeInference
	}
	return group.EffectiveRoutingScope() == scope
}

func apiKeyRouteMatchesModelRequest(route domain.APIKeyMultiGroupRoute, group *Group, requestedModel string, imageIntent bool) bool {
	groupRoutingScope := ""
	if group != nil {
		groupRoutingScope = group.EffectiveRoutingScope()
	}
	videoIntent := IsVideoGenerationIntent("", requestedModel, nil) || groupRoutingScope == GroupRoutingScopeVideo
	embeddingIntent := isEmbeddingModel(requestedModel) || groupRoutingScope == GroupRoutingScopeEmbedding
	if route.ImageOnly && !imageIntent {
		return false
	}
	if route.TextOnly && (imageIntent || videoIntent || embeddingIntent) {
		return false
	}
	if imageIntent && (group == nil || group.Platform != PlatformOpenAI || !group.AllowImageGeneration) {
		return false
	}
	return group != nil && group.MatchesModel(requestedModel)
}

func RoutingScopeForRequest(path, requestedModel string, imageIntent bool) string {
	if imageIntent || IsImageGenerationIntent(path, requestedModel, nil) {
		return GroupRoutingScopeImage
	}
	if IsVideoGenerationIntent(path, requestedModel, nil) {
		return GroupRoutingScopeVideo
	}
	if IsEmbeddingEndpoint(path) || isEmbeddingModel(requestedModel) {
		return GroupRoutingScopeEmbedding
	}
	return GroupRoutingScopeInference
}

func IsEmbeddingEndpoint(endpoint string) bool {
	endpoint = strings.TrimSpace(strings.ToLower(endpoint))
	if endpoint == "" {
		return false
	}
	if idx := strings.IndexByte(endpoint, '?'); idx >= 0 {
		endpoint = endpoint[:idx]
	}
	endpoint = strings.TrimRight(endpoint, "/")
	return endpoint == "/v1/embeddings" || endpoint == "/embeddings"
}

func isEmbeddingModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(model, "embedding") || strings.Contains(model, "embeddings")
}

func pickWeightedRouteGroup(candidates []apiKeyRouteCandidate) *Group {
	if len(candidates) == 0 {
		return nil
	}
	total := 0
	for _, candidate := range candidates {
		total += candidate.weight
	}
	if total <= 0 {
		return candidates[0].group
	}
	n := rand.IntN(total)
	for _, candidate := range candidates {
		if n < candidate.weight {
			return candidate.group
		}
		n -= candidate.weight
	}
	return candidates[len(candidates)-1].group
}

func preferredPlatformsForPath(path, forcePlatform string) []string {
	if forcePlatform != "" {
		return []string{forcePlatform}
	}
	path = strings.ToLower(strings.TrimSpace(path))
	switch {
	case strings.HasPrefix(path, "/antigravity/"):
		return []string{PlatformAntigravity}
	case strings.HasPrefix(path, "/v1beta"):
		return []string{PlatformGemini}
	case strings.HasPrefix(path, "/v1/images/") || strings.HasPrefix(path, "/images/"):
		return []string{PlatformOpenAI}
	case strings.HasPrefix(path, "/v1/midjourney/") || strings.HasPrefix(path, "/midjourney/"):
		return []string{PlatformOpenAI}
	case strings.HasPrefix(path, "/v1/videos/") || strings.HasPrefix(path, "/videos/") ||
		strings.HasPrefix(path, "/v1/tasks/") || strings.HasPrefix(path, "/tasks/"):
		return []string{PlatformOpenAI}
	case strings.HasPrefix(path, "/v1/chat/completions") || strings.HasPrefix(path, "/chat/completions"):
		return []string{PlatformOpenAI}
	case strings.HasPrefix(path, "/v1/responses") || strings.HasPrefix(path, "/responses") || strings.HasPrefix(path, "/backend-api/codex/responses"):
		return []string{PlatformOpenAI, PlatformAnthropic}
	default:
		return nil
	}
}

func preferredPlatformsForModel(model string) []string {
	model = strings.ToLower(strings.TrimSpace(model))
	switch {
	case model == "":
		return nil
	case strings.HasPrefix(model, "claude-") ||
		strings.HasPrefix(model, "anthropic.claude-") ||
		strings.Contains(model, ".anthropic.claude-"):
		return []string{PlatformAnthropic, PlatformAntigravity}
	case strings.HasPrefix(model, "gemini-"):
		return nil
	case strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4") ||
		strings.HasPrefix(model, "text-embedding-"):
		return []string{PlatformOpenAI}
	default:
		return nil
	}
}

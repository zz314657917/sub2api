package service

import (
	"math/rand/v2"
	"sort"
	"strings"
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
	if k == nil || len(k.MultiGroupRoutes) == 0 {
		return k
	}
	selected := k.selectRouteGroup(path, forcePlatform, skipGroup)
	if selected == nil {
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

func (k *APIKey) selectRouteGroup(path, forcePlatform string, skipGroup func(groupID int64) bool) *Group {
	groups := k.routeGroupsByID()
	if len(groups) == 0 {
		return nil
	}
	candidates := k.routeCandidates(groups, preferredPlatformsForPath(path, forcePlatform), skipGroup)
	if len(candidates) == 0 {
		candidates = k.routeCandidates(groups, nil, skipGroup)
	}
	if len(candidates) == 0 {
		return nil
	}
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
	for _, route := range k.MultiGroupRoutes {
		if route.GroupID != groupID {
			continue
		}
		if !route.Enabled {
			return 0, false
		}
		if route.CooldownSeconds <= 0 {
			return apiKeyRouteDefaultCooldown, true
		}
		return route.CooldownSeconds, true
	}
	return 0, false
}

type apiKeyRouteCandidate struct {
	group    *Group
	priority int
	weight   int
}

func (k *APIKey) routeCandidates(groups map[int64]*Group, platforms []string, skipGroup func(groupID int64) bool) []apiKeyRouteCandidate {
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
		if len(platformSet) > 0 {
			if _, ok := platformSet[group.Platform]; !ok {
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
	case strings.HasPrefix(path, "/v1/chat/completions") || strings.HasPrefix(path, "/chat/completions"):
		return []string{PlatformOpenAI}
	case strings.HasPrefix(path, "/v1/responses") || strings.HasPrefix(path, "/responses") || strings.HasPrefix(path, "/backend-api/codex/responses"):
		return []string{PlatformOpenAI, PlatformAnthropic}
	default:
		return nil
	}
}

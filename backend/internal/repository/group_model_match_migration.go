package repository

import (
	"context"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// GroupModelMatchMigrationReport is the audit output for the S91 switch.
// Empty ModelMatchPatterns are deliberately reported instead of guessed.
type GroupModelMatchMigrationReport struct {
	UnconfiguredGroups []GroupModelMatchUnconfiguredGroup `json:"unconfigured_groups"`
	LegacyAPIKeys      []GroupModelMatchLegacyAPIKey      `json:"legacy_api_keys"`
	Conflicts          []GroupModelMatchConflict          `json:"conflicts"`
}

type GroupModelMatchUnconfiguredGroup struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type GroupModelMatchLegacyAPIKey struct {
	ID       int64   `json:"id"`
	UserID   int64   `json:"user_id"`
	GroupIDs []int64 `json:"group_ids"`
}

type GroupModelMatchConflict struct {
	GroupID  int64    `json:"group_id"`
	Patterns []string `json:"patterns"`
}

// GroupModelMatchMigration owns the explicit, administrator-triggered switch.
// It is intentionally separate from automatic startup migrations: adding the
// column is safe at startup, but clearing user-authored rules is not.
type GroupModelMatchMigration struct {
	client      *dbent.Client
	invalidator service.APIKeyAuthCacheInvalidator
}

func NewGroupModelMatchMigration(client *dbent.Client, invalidator service.APIKeyAuthCacheInvalidator) *GroupModelMatchMigration {
	return &GroupModelMatchMigration{client: client, invalidator: invalidator}
}

func (m *GroupModelMatchMigration) Preflight(ctx context.Context) (GroupModelMatchMigrationReport, error) {
	if m == nil || m.client == nil {
		return GroupModelMatchMigrationReport{}, fmt.Errorf("group model match migration client is nil")
	}
	return collectGroupModelMatchMigrationReport(ctx, m.client)
}

// Switch performs the final migration atomically. dryRun returns the audit
// report without writing; any unconfigured active group blocks the real switch.
func (m *GroupModelMatchMigration) Switch(ctx context.Context, dryRun bool) (GroupModelMatchMigrationReport, error) {
	report, err := m.Preflight(ctx)
	if err != nil || dryRun {
		return report, err
	}
	if len(report.UnconfiguredGroups) > 0 {
		return report, fmt.Errorf("group model match migration blocked: %s", formatUnconfiguredGroups(report.UnconfiguredGroups))
	}

	tx, err := m.client.Tx(ctx)
	if err != nil {
		return report, fmt.Errorf("begin group model match migration: %w", err)
	}
	// Re-read the audit inside the write transaction so a group enabled or
	// cleared after the initial preflight cannot slip through the cutover.
	report, err = collectGroupModelMatchMigrationReport(ctx, tx.Client())
	if err != nil {
		_ = tx.Rollback()
		return report, fmt.Errorf("recheck group model match migration: %w", err)
	}
	if len(report.UnconfiguredGroups) > 0 {
		_ = tx.Rollback()
		return report, fmt.Errorf("group model match migration blocked: %s", formatUnconfiguredGroups(report.UnconfiguredGroups))
	}
	keys, err := tx.APIKey.Query().Where().All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return report, fmt.Errorf("load API keys for group model match migration: %w", err)
	}
	invalidatedKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		routes := make([]domain.APIKeyMultiGroupRoute, len(key.MultiGroupRoutes))
		copy(routes, key.MultiGroupRoutes)
		changed := false
		for i := range routes {
			if routes[i].ModelPatterns != nil {
				routes[i].ModelPatterns = nil
				changed = true
			}
		}
		if !changed {
			continue
		}
		if _, err := tx.APIKey.UpdateOneID(key.ID).SetMultiGroupRoutes(routes).Save(ctx); err != nil {
			_ = tx.Rollback()
			return report, fmt.Errorf("clear legacy model patterns for API key %d: %w", key.ID, err)
		}
		invalidatedKeys = append(invalidatedKeys, key.Key)
	}
	if err := tx.Commit(); err != nil {
		return report, fmt.Errorf("commit group model match migration: %w", err)
	}
	if m.invalidator != nil {
		for _, key := range invalidatedKeys {
			m.invalidator.InvalidateAuthCacheByKey(ctx, key)
		}
	}
	return report, nil
}

func collectGroupModelMatchMigrationReport(ctx context.Context, client *dbent.Client) (GroupModelMatchMigrationReport, error) {
	report := GroupModelMatchMigrationReport{}
	groups, err := client.Group.Query().Where(group.StatusEQ(service.StatusActive), group.DeletedAtIsNil()).All(ctx)
	if err != nil {
		return report, fmt.Errorf("list active groups for model match migration: %w", err)
	}
	for _, g := range groups {
		if len(service.NormalizeGroupModelMatchPatterns(g.ModelMatchPatterns)) == 0 {
			report.UnconfiguredGroups = append(report.UnconfiguredGroups, GroupModelMatchUnconfiguredGroup{ID: g.ID, Name: g.Name, Status: g.Status})
		}
	}
	keys, err := client.APIKey.Query().All(ctx)
	if err != nil {
		return report, fmt.Errorf("list API keys for model match migration: %w", err)
	}
	patternsByGroup := make(map[int64]map[string]struct{})
	ruleSetsByGroup := make(map[int64]map[string]struct{})
	for _, key := range keys {
		var groupIDs []int64
		legacy := false
		for _, route := range key.MultiGroupRoutes {
			if route.GroupID <= 0 {
				continue
			}
			groupIDs = append(groupIDs, route.GroupID)
			if route.ModelPatterns == nil {
				continue
			}
			legacy = true
			normalizedRoutePatterns := make([]string, 0, len(route.ModelPatterns))
			patterns := patternsByGroup[route.GroupID]
			if patterns == nil {
				patterns = make(map[string]struct{})
				patternsByGroup[route.GroupID] = patterns
			}
			for _, pattern := range route.ModelPatterns {
				pattern = strings.ToLower(strings.TrimSpace(pattern))
				if pattern != "" {
					patterns[pattern] = struct{}{}
					normalizedRoutePatterns = append(normalizedRoutePatterns, pattern)
				}
			}
			if len(normalizedRoutePatterns) > 0 {
				sortStrings(normalizedRoutePatterns)
				signature := strings.Join(uniqueStrings(normalizedRoutePatterns), "\n")
				if ruleSetsByGroup[route.GroupID] == nil {
					ruleSetsByGroup[route.GroupID] = make(map[string]struct{})
				}
				ruleSetsByGroup[route.GroupID][signature] = struct{}{}
			}
		}
		if legacy {
			report.LegacyAPIKeys = append(report.LegacyAPIKeys, GroupModelMatchLegacyAPIKey{ID: key.ID, UserID: key.UserID, GroupIDs: uniqueInt64s(groupIDs)})
		}
	}
	for groupID, ruleSets := range ruleSetsByGroup {
		if len(ruleSets) < 2 {
			continue
		}
		patterns := patternsByGroup[groupID]
		values := make([]string, 0, len(patterns))
		for pattern := range patterns {
			values = append(values, pattern)
		}
		sortStrings(values)
		report.Conflicts = append(report.Conflicts, GroupModelMatchConflict{GroupID: groupID, Patterns: values})
	}
	return report, nil
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	out := values[:0]
	for _, value := range values {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

func formatUnconfiguredGroups(groups []GroupModelMatchUnconfiguredGroup) string {
	parts := make([]string, 0, len(groups))
	for _, group := range groups {
		parts = append(parts, fmt.Sprintf("%d:%s", group.ID, group.Name))
	}
	return strings.Join(parts, ", ")
}

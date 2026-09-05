package securityaudit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const maxPromptPolicyHistory = 20

func defaultPolicyHistory() PolicyHistory {
	return PolicyHistory{Versions: []PolicyVersionRecord{}}
}

func parsePolicyHistory(raw string) (PolicyHistory, error) {
	history := defaultPolicyHistory()
	if strings.TrimSpace(raw) == "" {
		return history, nil
	}
	if err := json.Unmarshal([]byte(raw), &history); err != nil {
		return PolicyHistory{}, fmt.Errorf("decode prompt audit policy history: %w", err)
	}
	if history.ActiveVersion < 0 {
		return PolicyHistory{}, errors.New("prompt audit policy history active version is invalid")
	}
	if history.Draft != nil {
		if history.Draft.DraftVersion < 1 || history.Draft.BaseConfigVersion < 1 {
			history.Draft = nil
		} else {
			normalizeRiskActionRules(&history.Draft.Rules)
			if err := validateRiskActionRules(history.Draft.Rules); err != nil {
				history.Draft = nil
			}
		}
	}
	valid := make([]PolicyVersionRecord, 0, len(history.Versions))
	seen := map[int]struct{}{}
	for _, version := range history.Versions {
		if version.PolicyVersion < 1 || version.PolicyVersion > 1_000_000 || strings.TrimSpace(version.PolicyID) == "" {
			continue
		}
		if _, ok := seen[version.PolicyVersion]; ok {
			continue
		}
		version.Rules.PolicyID = strings.TrimSpace(version.PolicyID)
		version.Rules.PolicyVersion = version.PolicyVersion
		normalizeRiskActionRules(&version.Rules)
		if err := validateRiskActionRules(version.Rules); err != nil {
			continue
		}
		version.PolicyID = version.Rules.PolicyID
		seen[version.PolicyVersion] = struct{}{}
		valid = append(valid, version)
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].PolicyVersion > valid[j].PolicyVersion })
	if len(valid) > maxPromptPolicyHistory {
		valid = valid[:maxPromptPolicyHistory]
	}
	history.Versions = valid
	if history.ActiveVersion > 0 {
		if _, ok := findPolicyVersion(history, history.ActiveVersion); !ok {
			history.ActiveVersion = 0
		}
	}
	return history, nil
}

func marshalPolicyHistory(history PolicyHistory) (string, error) {
	if len(history.Versions) > maxPromptPolicyHistory {
		history.Versions = history.Versions[:maxPromptPolicyHistory]
	}
	for i := range history.Versions {
		normalizeRiskActionRules(&history.Versions[i].Rules)
	}
	raw, err := json.Marshal(history)
	return string(raw), err
}

func readPolicyHistoryTx(ctx context.Context, tx *sql.Tx) (PolicyHistory, error) {
	var raw string
	err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key=$1 FOR UPDATE`, SettingKeyPromptAuditPolicyHistory).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultPolicyHistory(), nil
	}
	if err != nil {
		return PolicyHistory{}, err
	}
	return parsePolicyHistory(raw)
}

func writePolicyHistoryTx(ctx context.Context, tx *sql.Tx, history PolicyHistory) error {
	raw, err := marshalPolicyHistory(history)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO settings (key,value,updated_at) VALUES ($1,$2,NOW())
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`,
		SettingKeyPromptAuditPolicyHistory, raw)
	return err
}

func appendPolicyVersion(history PolicyHistory, rules RiskActionRules, configVersion int64, actorID int64, now time.Time) PolicyHistory {
	normalizeRiskActionRules(&rules)
	if rules.PolicyVersion < 1 {
		rules.PolicyVersion = 1
	}
	if strings.TrimSpace(rules.PolicyID) == "" {
		rules.PolicyID = "priority"
	}
	history.Versions = append(history.Versions, PolicyVersionRecord{
		PolicyVersion: rules.PolicyVersion,
		PolicyID:      rules.PolicyID,
		Rules:         rules,
		ConfigVersion: configVersion,
		CreatedAt:     now,
		CreatedBy:     actorID,
	})
	history.ActiveVersion = rules.PolicyVersion
	sort.SliceStable(history.Versions, func(i, j int) bool { return history.Versions[i].PolicyVersion > history.Versions[j].PolicyVersion })
	if len(history.Versions) > maxPromptPolicyHistory {
		history.Versions = history.Versions[:maxPromptPolicyHistory]
	}
	return history
}

func nextPolicyVersion(history PolicyHistory, current RiskActionRules) int {
	next := current.PolicyVersion
	for _, version := range history.Versions {
		if version.PolicyVersion > next {
			next = version.PolicyVersion
		}
	}
	if next < 1 {
		next = 0
	}
	return next + 1
}

func (m *ConfigManager) ListPolicyVersions(ctx context.Context) (PolicyHistory, error) {
	if m == nil || m.settings == nil {
		return PolicyHistory{}, infraerrors.ServiceUnavailable(ErrorCodeConfigUnavailable, "提示词审计策略历史暂不可用")
	}
	raw, err := m.settings.GetValue(ctx, SettingKeyPromptAuditPolicyHistory)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, service.ErrSettingNotFound) {
			return defaultPolicyHistory(), nil
		}
		return PolicyHistory{}, err
	}
	history, err := parsePolicyHistory(raw)
	if err != nil {
		return PolicyHistory{}, infraerrors.ServiceUnavailable(ErrorCodeConfigUnavailable, "提示词审计策略历史无效")
	}
	if active, ok := m.Active(); ok && history.ActiveVersion == 0 && active.Rules.PolicyVersion > 0 {
		history.ActiveVersion = active.Rules.PolicyVersion
	}
	return history, nil
}

func buildPolicyPreview(rules RiskActionRules) (PolicyPreview, error) {
	return buildPolicyPreviewWithActive(rules, ActiveConfig{Scanners: AllScannerIDs})
}

func buildPolicyPreviewWithActive(rules RiskActionRules, active ActiveConfig) (PolicyPreview, error) {
	rules = cloneRiskActionRules(rules)
	normalizeRiskActionRules(&rules)
	if err := validateRiskActionRules(rules); err != nil {
		return PolicyPreview{}, err
	}
	preview := PolicyPreview{PolicyID: rules.PolicyID, RuleCount: len(rules.Rules), CategoryCount: len(rules.Categories)}
	seenScopes := map[string]struct{}{}
	for _, rule := range rules.Rules {
		switch rule.Action {
		case ActionBlock:
			preview.BlockingRuleCount++
		case ActionWarn:
			preview.WarningRuleCount++
		}
		scoped := false
		if len(rule.Groups) > 0 {
			scoped = true
			seenScopes["groups"] = struct{}{}
		}
		if len(rule.Models) > 0 {
			scoped = true
			seenScopes["models"] = struct{}{}
		}
		if len(rule.Providers) > 0 {
			scoped = true
			seenScopes["providers"] = struct{}{}
		}
		if len(rule.Safety) > 0 {
			seenScopes["safety"] = struct{}{}
		}
		if len(rule.Categories) > 0 {
			seenScopes["categories"] = struct{}{}
		}
		if scoped {
			preview.ScopedRuleCount++
		}
	}
	for scope := range seenScopes {
		preview.AffectedScopes = append(preview.AffectedScopes, scope)
	}
	sort.Strings(preview.AffectedScopes)
	samples := []struct {
		name string
		text string
	}{
		{name: "safe", text: "Safety: Safe\nCategories: None"},
		{name: "controversial_jailbreak", text: "Safety: Controversial\nCategories: Jailbreak"},
		{name: "controversial_violent", text: "Safety: Controversial\nCategories: Violent"},
		{name: "unsafe_pii", text: "Safety: Unsafe\nCategories: PII"},
	}
	preview.Examples = make([]PolicyPreviewExample, 0, len(samples))
	for _, sample := range samples {
		shadow, err := compareGuardPolicies(sample.text, active, rules, PolicyMatchContext{})
		if err != nil {
			return PolicyPreview{}, err
		}
		preview.Examples = append(preview.Examples, PolicyPreviewExample{
			Name:               sample.name,
			Safety:             shadow.Current.Safety,
			Categories:         append([]string(nil), shadow.Current.Categories...),
			CurrentAction:      shadow.Current.Action,
			CurrentRiskLevel:   shadow.Current.RiskLevel,
			CandidateAction:    shadow.Candidate.Action,
			CandidateRiskLevel: shadow.Candidate.RiskLevel,
			MatchedRuleID:      shadow.Candidate.MatchedRuleID,
			OWASPTags:          append([]string(nil), shadow.Candidate.OWASPTags...),
			WouldEscalate:      shadow.WouldEscalate,
		})
	}
	return preview, nil
}

func (m *ConfigManager) PreviewPolicy(ctx context.Context, rules RiskActionRules) (PolicyPreview, error) {
	if m == nil {
		return PolicyPreview{}, infraerrors.ServiceUnavailable(ErrorCodeConfigUnavailable, "prompt audit active policy unavailable")
	}
	active, ok := m.Active()
	if !ok {
		return PolicyPreview{}, infraerrors.ServiceUnavailable(ErrorCodeConfigUnavailable, "prompt audit active policy unavailable")
	}
	return buildPolicyPreviewWithActive(rules, active)
}

func compareGuardPolicies(output string, active ActiveConfig, candidateRules RiskActionRules, matchContext PolicyMatchContext) (PolicyShadowResult, error) {
	if strings.TrimSpace(output) == "" || len(output) > 4096 {
		return PolicyShadowResult{}, infraerrors.BadRequest("prompt_audit_invalid_policy_shadow", "invalid synthetic Guard output")
	}
	candidateRules = cloneRiskActionRules(candidateRules)
	normalizeRiskActionRules(&candidateRules)
	if err := validateRiskActionRules(candidateRules); err != nil {
		return PolicyShadowResult{}, err
	}
	baseline, err := ParseQwen3Guard(output, active.Scanners)
	if err != nil {
		return PolicyShadowResult{}, infraerrors.BadRequest("prompt_audit_invalid_policy_shadow", "invalid synthetic Guard output")
	}
	// Evaluate both policies from the parser baseline, never from an already
	// escalated current result, so removing a custom escalation is observable.
	current, candidate := *baseline, *baseline
	ApplyRiskPolicy(&current, active.Rules, matchContext)
	ApplyRiskPolicy(&candidate, candidateRules, matchContext)
	return PolicyShadowResult{
		Current: current, Candidate: candidate,
		ActionChanged: current.Action != candidate.Action,
		RiskChanged:   current.RiskLevel != candidate.RiskLevel,
		WouldEscalate: actionRank(candidate.Action) > actionRank(current.Action) ||
			(actionRank(candidate.Action) == actionRank(current.Action) && riskRank(candidate.RiskLevel) > riskRank(current.RiskLevel)),
	}, nil
}

func (m *ConfigManager) SavePolicyDraft(ctx context.Context, req PolicyDraftRequest, actorID int64) (PolicyHistory, error) {
	if m == nil || m.db == nil {
		return PolicyHistory{}, errors.New("prompt audit policy persistence unavailable")
	}
	if req.ExpectedConfigVersion < 1 {
		return PolicyHistory{}, infraerrors.BadRequest("prompt_audit_invalid_policy_draft", "配置版本无效")
	}
	if _, err := buildPolicyPreview(req.Rules); err != nil {
		return PolicyHistory{}, err
	}
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return PolicyHistory{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, promptAuditConfigLockKey); err != nil {
		return PolicyHistory{}, err
	}
	current, err := readStorageConfigTx(ctx, tx)
	if err != nil {
		return PolicyHistory{}, err
	}
	if current.ConfigVersion != req.ExpectedConfigVersion {
		return PolicyHistory{}, infraerrors.Conflict(ErrorCodeConfigConflict, "提示词审计配置已被其他管理员更新")
	}
	history, err := readPolicyHistoryTx(ctx, tx)
	if err != nil {
		return PolicyHistory{}, err
	}
	if history.Draft == nil && req.ExpectedDraftVersion != 0 {
		return PolicyHistory{}, infraerrors.Conflict(ErrorCodeConfigConflict, "提示词审计草稿已不存在或已被更新")
	}
	if history.Draft != nil && history.Draft.DraftVersion != req.ExpectedDraftVersion {
		return PolicyHistory{}, infraerrors.Conflict(ErrorCodeConfigConflict, "提示词审计草稿已被其他管理员更新")
	}
	draftVersion := req.ExpectedDraftVersion + 1
	normalizeRiskActionRules(&req.Rules)
	if req.Rules.PolicyID == "" {
		req.Rules.PolicyID = current.Rules.PolicyID
	}
	if req.Rules.PolicyVersion < 1 {
		req.Rules.PolicyVersion = nextPolicyVersion(history, current.Rules)
	}
	history.Draft = &PolicyDraft{DraftVersion: draftVersion, BaseConfigVersion: current.ConfigVersion, Rules: cloneRiskActionRules(req.Rules), UpdatedAt: m.clock.Now(), UpdatedBy: actorID}
	if err := writePolicyHistoryTx(ctx, tx, history); err != nil {
		return PolicyHistory{}, err
	}
	if err := tx.Commit(); err != nil {
		return PolicyHistory{}, err
	}
	return history, nil
}

func readStorageConfigTx(ctx context.Context, tx *sql.Tx) (storageConfig, error) {
	var raw string
	err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key=$1 FOR UPDATE`, SettingKeyPromptAuditConfig).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultStorageConfig(), nil
	}
	if err != nil {
		return storageConfig{}, err
	}
	return ParseStorageConfig(raw)
}

func (m *ConfigManager) PublishPolicyDraft(ctx context.Context, req PolicyPublishRequest, actorID int64) (PublicConfig, error) {
	if m == nil || m.db == nil || m.encryptor == nil {
		return PublicConfig{}, errors.New("prompt audit policy persistence unavailable")
	}
	if req.ExpectedConfigVersion < 1 || req.ExpectedDraftVersion < 1 {
		return PublicConfig{}, infraerrors.BadRequest("prompt_audit_invalid_policy_publish", "发布版本无效")
	}
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return PublicConfig{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, promptAuditConfigLockKey); err != nil {
		return PublicConfig{}, err
	}
	current, err := readStorageConfigTx(ctx, tx)
	if err != nil {
		return PublicConfig{}, err
	}
	if current.ConfigVersion != req.ExpectedConfigVersion {
		return PublicConfig{}, infraerrors.Conflict(ErrorCodeConfigConflict, "提示词审计配置已被其他管理员更新")
	}
	history, err := readPolicyHistoryTx(ctx, tx)
	if err != nil {
		return PublicConfig{}, err
	}
	if history.Draft == nil || history.Draft.DraftVersion != req.ExpectedDraftVersion {
		return PublicConfig{}, infraerrors.Conflict(ErrorCodeConfigConflict, "提示词审计草稿已不存在或已被更新")
	}
	if history.Draft.BaseConfigVersion != current.ConfigVersion {
		return PublicConfig{}, infraerrors.Conflict(ErrorCodeConfigConflict, "提示词审计草稿基于旧配置，需重新保存")
	}
	next := cloneStorageConfig(current)
	next.ConfigVersion = current.ConfigVersion + 1
	next.Rules = cloneRiskActionRules(history.Draft.Rules)
	next.Rules.PolicyVersion = nextPolicyVersion(history, current.Rules)
	normalizeRiskActionRules(&next.Rules)
	if strings.TrimSpace(next.Rules.PolicyID) == "" {
		next.Rules.PolicyID = "priority"
	}
	next.UpdatedAt = m.clock.Now()
	next.UpdatedBy = actorID
	next.ChangeSummary = changeSummary(next)
	history = appendPolicyVersion(history, next.Rules, next.ConfigVersion, actorID, m.clock.Now())
	history.Draft = nil
	rawNext, err := json.Marshal(next)
	if err != nil {
		return PublicConfig{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO settings (key,value,updated_at) VALUES ($1,$2,NOW())
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`,
		SettingKeyPromptAuditConfig, string(rawNext)); err != nil {
		return PublicConfig{}, err
	}
	if err := writePolicyHistoryTx(ctx, tx, history); err != nil {
		return PublicConfig{}, err
	}
	if err := tx.Commit(); err != nil {
		return PublicConfig{}, err
	}
	return m.installPublishedConfig(ctx, next, actorID, "publish")
}

func (m *ConfigManager) installPublishedConfig(ctx context.Context, next storageConfig, _ int64, status string) (PublicConfig, error) {
	riskControlEnabled := m.currentRiskControlEnabled()
	if values, getErr := m.settings.GetMultiple(ctx, []string{SettingKeyRiskControl}); getErr == nil {
		riskControlEnabled = values[SettingKeyRiskControl] == "true"
	}
	active, err := ActiveFromStorage(next, riskControlEnabled, m.encryptor)
	if err != nil {
		return PublicConfig{}, err
	}
	previous, installed := m.installConfigSnapshot(next, active)
	if !installed {
		return PublicFromStorage(next, active.RiskControlEnabled, active.InvalidTokenEndpointIDs()), nil
	}
	m.configUntrusted.Store(false)
	m.clearLoadError()
	m.logInvalidTokenEndpoints(previous, active)
	LogInfo(EventConfigUpdated, map[string]any{"config_version": next.ConfigVersion, "policy_version": next.Rules.PolicyVersion, "status": status})
	if m.redis != nil {
		if err := m.redis.Publish(ctx, ConfigInvalidationChannel, strconv.FormatInt(next.ConfigVersion, 10)).Err(); err != nil {
			LogWarn(EventConfigReloadDegraded, map[string]any{"config_version": next.ConfigVersion, "status": "degraded", "error_code": "config_invalidation_publish_failed"})
		}
	}
	return PublicFromStorage(next, active.RiskControlEnabled, active.InvalidTokenEndpointIDs()), nil
}

func findPolicyVersion(history PolicyHistory, version int) (PolicyVersionRecord, bool) {
	for _, item := range history.Versions {
		if item.PolicyVersion == version {
			return item, true
		}
	}
	return PolicyVersionRecord{}, false
}

func (m *ConfigManager) RollbackPolicy(ctx context.Context, policyVersion int, expectedConfigVersion, actorID int64) (PublicConfig, error) {
	if m == nil || m.db == nil || m.encryptor == nil {
		return PublicConfig{}, errors.New("prompt audit config persistence unavailable")
	}
	if policyVersion < 1 || expectedConfigVersion < 1 {
		return PublicConfig{}, infraerrors.BadRequest("prompt_audit_invalid_policy_rollback", "策略版本或配置版本无效")
	}
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return PublicConfig{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, promptAuditConfigLockKey); err != nil {
		return PublicConfig{}, err
	}
	var raw string
	err = tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key=$1 FOR UPDATE`, SettingKeyPromptAuditConfig).Scan(&raw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PublicConfig{}, err
	}
	current := DefaultStorageConfig()
	if err == nil {
		current, err = ParseStorageConfig(raw)
		if err != nil {
			return PublicConfig{}, err
		}
	}
	if current.ConfigVersion != expectedConfigVersion {
		return PublicConfig{}, infraerrors.Conflict(ErrorCodeConfigConflict, "提示词审计配置已被其他管理员更新")
	}
	history, err := readPolicyHistoryTx(ctx, tx)
	if err != nil {
		return PublicConfig{}, err
	}
	target, ok := findPolicyVersion(history, policyVersion)
	if !ok {
		return PublicConfig{}, infraerrors.NotFound("prompt_audit_policy_version_not_found", "提示词审计策略版本不存在")
	}
	next := cloneStorageConfig(current)
	next.Rules = cloneRiskActionRules(target.Rules)
	next.ConfigVersion = current.ConfigVersion + 1
	next.UpdatedAt = m.clock.Now()
	next.UpdatedBy = actorID
	next.ChangeSummary = changeSummary(next)
	rawNext, err := json.Marshal(next)
	if err != nil {
		return PublicConfig{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO settings (key,value,updated_at) VALUES ($1,$2,NOW())
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`,
		SettingKeyPromptAuditConfig, string(rawNext)); err != nil {
		return PublicConfig{}, err
	}
	for index := range history.Versions {
		if history.Versions[index].PolicyVersion == policyVersion {
			history.Versions[index].RollbackCount++
		}
	}
	history.ActiveVersion = policyVersion
	if err := writePolicyHistoryTx(ctx, tx, history); err != nil {
		return PublicConfig{}, err
	}
	if err := tx.Commit(); err != nil {
		return PublicConfig{}, err
	}

	riskControlEnabled := m.currentRiskControlEnabled()
	if values, getErr := m.settings.GetMultiple(ctx, []string{SettingKeyRiskControl}); getErr == nil {
		riskControlEnabled = values[SettingKeyRiskControl] == "true"
	}
	active, err := ActiveFromStorage(next, riskControlEnabled, m.encryptor)
	if err != nil {
		return PublicConfig{}, err
	}
	previous, installed := m.installConfigSnapshot(next, active)
	if !installed {
		return PublicFromStorage(next, active.RiskControlEnabled, active.InvalidTokenEndpointIDs()), nil
	}
	m.configUntrusted.Store(false)
	m.clearLoadError()
	m.logInvalidTokenEndpoints(previous, active)
	LogInfo(EventConfigUpdated, map[string]any{
		"config_version": next.ConfigVersion, "policy_version": policyVersion, "status": "rollback",
	})
	if m.redis != nil {
		if err := m.redis.Publish(ctx, ConfigInvalidationChannel, strconv.FormatInt(next.ConfigVersion, 10)).Err(); err != nil {
			LogWarn(EventConfigReloadDegraded, map[string]any{
				"config_version": next.ConfigVersion, "policy_version": policyVersion,
				"status": "degraded", "error_code": "config_invalidation_publish_failed",
			})
		}
	}
	return PublicFromStorage(next, active.RiskControlEnabled, active.InvalidTokenEndpointIDs()), nil
}

### BLOCKED: prompt-audit-policy-matrix-s293-remediation

## Findings

2026-09-05 continuation, final evaluator: code-level fixes and focused regressions
are complete for this batch; runtime acceptance remains blocked. This is not a
deployment approval and does not replace S294's independent status.

- Fixed strict no-draft CAS, stale draft base publication, and out-of-order snapshot
  installation across Reload/Save/Publish/Rollback. Same-version global gate updates
  and invalid-config fail-closed behavior are preserved.
- Fixed historical record ID validation, trimmed-history active membership, rule
  attribution after defaults/maps, and equal-action/risk chunk priority/ID selection.
  Weaker chunks cannot supply attribution for a stronger baseline-only decision.
- Fixed saved-draft/editor separation, edits during saving, stale preview/shadow
  responses, publication/rollback version continuity, and ordinary-save stale drafts.
- Admin shadow now submits bounded synthetic Guard output. Active and candidate
  policies use the same parser baseline and current scanners; missing active config
  is 503. Legacy normalized current_result shadow remains compatible. OWASP labels
  remain explanatory metadata for matching rules, not just escalation winners.

## Independent Review

- Terra `r2_r3_boundary`: approved amended R2/R3 contracts after review.
- Terra `r1_final_qa`: confirmed history ID and chunk attribution defects, then
  verified their exact reproductions no longer fail. Withdrawn winner-only OWASP
  concern after the contract clarified existing explanatory-tag compatibility.
- Terra `r2_r3_final_qa`: confirmed retained-history active pointer and ordinary-save
  stale-draft UI gaps; both were fixed and independently retested. No remaining
  concrete blocker reported in the reviewed code scope.
- Terra workers independently implemented lifecycle tests, snapshot fixes and UI
  changes. Controller reviewed results and performed the final checks below.

## Executed Checks

- Backend: `go test ./internal/securityaudit ./internal/server/routes ./internal/server/middleware ./migrations -count=1` PASS.
- Backend: `go build ./...` PASS.
- Focused policy lifecycle SQL/Redis tests PASS with sqlmock and private miniredis.
  They assert CAS, settings payloads, failed writes/commits, and post-commit notices.
- Deterministic delayed Reload and monotonic-install tests PASS; these are process
  tests, not a real two-backend deployment.
- Frontend: `npm.cmd run test:run -- src/features/prompt-audit/__tests__` 43/43 PASS.
- Frontend typecheck PASS; latest `npm.cmd run build` PASS, Vite completed in 23.32s.
  Existing Browserslist age, DEP0190 and bundle-size warnings remain non-fatal.
- Independent QA also ran Go vet and targeted frontend/history tests successfully.
- `go test ./... -count=1` FAIL: existing repository fixture mismatch at
  `account_repo_upstream_billing_probe_update_test.go:559` (32 columns / 34 values),
  plus OS Access is denied while starting the usagestats test executable. No
  unrelated fixture, dependency, or host-security change was made.
- New `TestPromptAuditPolicyExplanationRoundTrip` compiles but explicitly SKIPs
  because PROMPT_AUDIT_TEST_POSTGRES_DSN is absent. The isolated fixture now includes
  migration 239 and synchronous/asynchronous pass/warn/block, empty/nonempty tags,
  readback field mapping, and 128/129-character rule ID checks. This is not PG PASS.
- Browser preparation used the Playwright workflow; the task-owned Vite launch was
  rejected by tool policy before starting. No browser was opened. Port 5278 and
  task session/profile process checks found no task-owned running resources.

## Scope And Unverified Risks

- Source edits remain within securityaudit and Prompt Audit frontend/i18n owners;
  workflow/current-task records document this batch. S294, billing, routing,
  lockfile and outputs changes were preserved. No commit, push, deployment, shared
  container update, migration execution or real provider request occurred.
- Dedicated PostgreSQL event roundtrip/migration/real concurrent transactions,
  Redis multi-instance convergence and notification-loss recovery, authenticated
  real-backend browser workflows, and Qwen3Guard 0.6b smoke remain unverified.
- Installing the fixed program requires one build/update. Subsequent supported
  rule edits publish as configuration; no additional build should be required.
  Real multi-instance no-restart proof remains an R4 acceptance requirement.

## Recommendation

BLOCKED for release. Provision and explicitly identify dedicated test resources,
then execute R4. Never run the existing destructive PostgreSQL fixture against an
unverified/shared DSN, and never count a skipped integration test as a pass.

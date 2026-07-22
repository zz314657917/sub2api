# Task Contract: upstream-openai-inactive-workspace-plan-s104

- Task ID: `upstream-openai-inactive-workspace-plan-s104`
- Role: Planner / Generator / Evaluator
- Goal: Port upstream `d0b8760eb` so inactive, deactivated, disabled, deleted,
  suspended, or expired ChatGPT workspaces cannot overwrite the valid OpenAI
  plan type already decoded from an ID/access token.
- Success Criteria:
  - An existing token-derived `plan_type` is preserved when
    `accounts/check` returns a different workspace billing value.
  - `accounts/check` remains a fallback when the token does not contain a
    plan type.
  - Explicitly inactive workspace candidates and candidates with expired
    entitlements are skipped during fallback selection.
  - Active candidates and malformed/missing entitlement expiry values retain
    the existing best-effort behavior.
  - Existing subscription-expiry and privacy-mode enrichment behavior remains
    unchanged.
- Allowed Paths:
  - `backend/internal/service/openai_oauth_service.go`
  - `backend/internal/service/openai_privacy_service.go`
  - `backend/internal/service/openai_subscription_test.go`
  - `docs/workflow/tasks/upstream-openai-inactive-workspace-plan-s104.md`
  - `docs/workflow/qa-reports/upstream-openai-inactive-workspace-plan-s104-qa.md`
  - `docs/workflow/spec.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: Codex session identity matching, PAT or Agent Identity auth,
  account persistence, scheduler routing, gateway requests, frontend, Ent,
  migrations, billing, deployment, containers, VERSION, and unrelated
  workflow history.
- Constraints:
  - Adapt the upstream behavior manually to the local fork; do not merge or
    cherry-pick upstream history.
  - Preserve token-derived plan types regardless of their concrete value,
    including K12/Team/custom plan names.
  - Treat missing or malformed entitlement expiry as usable to preserve the
    current best-effort fallback rather than failing account enrichment.
  - Keep `accounts/check` and subscription requests best-effort and retain
    their current timeouts, headers, logging, and error handling.
- Acceptance Commands:
  - `go test ./internal/service -run "TestFetchChatGPTAccountInfo|TestShouldApplyChatGPTAccountInfoPlanType|TestFetchChatGPTSubscriptionExpiresAt" -count=1`
  - `gofmt -d` on the three allowed Go files.
  - `git diff --check`, conflict-marker scan, exact allowlist audit, and
    unmerged-index check.
- Output: Scoped source diff, focused regressions, QA report, and final
  `PASS`, `FAIL`, or `BLOCKED` evidence.
- Stop Rules: Stop on any need to change import identity, PAT/Agent Identity,
  persistence, scheduler behavior, gateway behavior, schema, deployment, or a
  path outside the exact allowlist.

## Contract Review

`PASS`: Upstream `d0b8760eb` is an isolated three-file service/test patch. The
local OAuth enrichment and account-info fallback topology matches the upstream
pre-patch structure, and no schema, migration, frontend, auth-mode, or runtime
deployment prerequisite is required.

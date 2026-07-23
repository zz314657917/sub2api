# Task Contract: upstream-v0164-small-fixes-s111

- Task ID: `upstream-v0164-small-fixes-s111`
- Role: Planner / Generator / Evaluator
- Goal: Manually port four isolated fixes from upstream `v0.1.164` without
  merging upstream history: Grok CC Switch import (`a3a1575e9`), Grok 402
  cooldown (`ca0d3314c`), long model-rate-limit display (`48d58d72f`), and
  concrete GPT-5.6 Sol default ordering (`dd5956be5`).
- Success Criteria:
  - Grok CC Switch imports target `grokbuild`, use `grok-4.5`, and append
    exactly one `/v1` suffix while preserving the normalized homepage.
  - Grok HTTP 402 responses temporarily unschedule the account for 30 minutes
    with a non-sensitive reason; existing 401, 403, 429, and 5xx behavior is
    unchanged.
  - Model rate-limit badges use the shared localized countdown so durations
    over 24 hours display in days, and their tooltips include the complete
    local date and time while preserving local console-theme classes.
  - The default OpenAI model list keeps both `gpt-5.6-sol` and the bare
    `gpt-5.6` alias exactly once, but exposes `gpt-5.6-sol` first for account
    tests and available-model selection.
- Allowed Paths:
  - `backend/internal/service/openai_gateway_grok.go`
  - `backend/internal/service/openai_gateway_grok_test.go`
  - `backend/internal/service/openai_gateway_grok_s111_test.go`
  - `backend/internal/pkg/openai/constants.go`
  - `backend/internal/pkg/openai/constants_test.go`
  - `backend/internal/handler/admin/account_handler_available_models_test.go`
  - `frontend/src/utils/ccswitchImport.ts`
  - `frontend/src/utils/__tests__/ccswitchImport.spec.ts`
  - `frontend/src/components/account/AccountStatusIndicator.vue`
  - `frontend/src/components/account/__tests__/AccountStatusIndicator.spec.ts`
  - `docs/workflow/tasks/upstream-v0164-small-fixes-s111.md`
  - `docs/workflow/qa-reports/upstream-v0164-small-fixes-s111-qa.md`
  - `docs/workflow/spec.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
  - `knowledge/tasks/timeline.md`
- Denied Paths: Schema and migrations, billing, payment execution, proxy
  quarantine, composite groups, Ollama Cloud usage, account import indexing,
  dependencies, deployment, containers, VERSION, and unrelated upstream
  changes.
- Constraints:
  - Adapt behavior to the local topology; do not cherry-pick or merge upstream
    history.
  - Preserve local CC Switch base-URL normalization and usage script behavior.
  - Preserve local console theme classes and all model-status classification.
  - Do not change the bare `gpt-5.6` alias semantics or
    `openai.DefaultTestModel`; only change default list ordering.
  - Do not modify the separate S110 group-buy worktree or branch.
  - Do not commit, push, deploy, or update containers without separate
    authorization.
- Acceptance Commands:
  - `go test ./internal/service -run TestHandleGrokAccountUpstreamErrorPaymentRequiredPausesAccount -count=10`
  - `go test ./internal/pkg/openai -run TestDefaultModels -count=1`
  - `go test ./internal/handler/admin -run TestAccountHandlerGetAvailableModels_OpenAIAPIKeyDefaultsToConcreteGPT56Sol -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/utils/__tests__/ccswitchImport.spec.ts src/components/account/__tests__/AccountStatusIndicator.spec.ts"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"`
  - Targeted ESLint, `gofmt -d`, `git diff --check`, conflict-marker scan,
    exact allowlist audit, and unmerged-index check.
- Output: Scoped source changes, focused regressions, QA report, and final
  `PASS`, `FAIL`, or `BLOCKED` evidence. No automatic commit or push.
- Stop Rules: Stop on any need for schema/migration changes, external runtime
  credentials, payment or billing changes, proxy circuit work, dependency
  updates, deployment/container work, or a business path outside the approved
  allowlist.

## Contract Review

`PASS`: All four behaviors have narrow existing local boundaries and focused
test seams. The CC Switch and display patches require local presentation
adaptation, while the two backend changes are isolated ordering/error-policy
ports. No schema, API, dependency, deployment, or cross-worktree prerequisite
is required.

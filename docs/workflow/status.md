---
phase: done
current_sprint: upstream-v0184-compat-fixes-s276
total_sprints: 276
pending_action: S276 is committed locally after independent QA PASS. Preserve unrelated dirty paths and outputs; do not push unless explicitly requested. Next legal action is a new Planner contract for S277.
project_type: fullstack
qa_mode: runtime
approval_required: true
last_verified: 2026-08-31 22:18 +08:00
---

# Upstream v0.1.184 Compatibility Fixes S276

- `contract-draft`: port the Anthropic-to-Responses stream lifecycle/content
  index repair, streamed Anthropic tool argument placeholder repair, saved SMTP
  TLS fallback for test endpoints, and custom version suffix comparison.
- The work must use local owners and focused regressions. Whole-history merge,
  cherry-pick, frontend, schema, migrations, billing, provider traffic,
  dependencies, VERSION, containers, deployment, shared data, push and
  `outputs/**` are excluded. Contract:
  `docs/workflow/tasks/upstream-v0184-compat-fixes-s276.md`.
- Protected user changes remain in API-key route breaker/auth files,
  `backend/internal/service/admin_service.go`, the Pixel Cafe admin view and
  `outputs/**`.
- `PASS / contract-review`: all four behaviors map to existing local owners,
  the SMTP split-file topology is explicitly adapted, tests cover the critical
  positive/negative cases, and the approved allowlist does not overlap current
  protected dirty files. Terra Developer implementation is authorized.
- `BLOCKED / generator-dispatch`: `gpt-5.6-terra` returned HTTP 403 with
  `No available group route matches the requested model or request type`
  before processing the prompt. Cost and token usage are zero; the task
  worktree is clean and no business source or external state changed. Report:
  `docs/workflow/worker-results/upstream-v0184-compat-fixes-s276-result.md`.
- `AUTHORIZED / model-exception`: the user explicitly approved
  `gpt-5.6-sol` for both the S276 Developer Worker and independent QA. The
  contract and review record this S276-only exception; the global Agent Matrix
  remains unchanged.
- `BLOCKED / generator-dispatch-retry`: the Sol retry also returned HTTP 403
  with `No available group route matches the requested model or request type`
  before processing the prompt. Both attempts used zero tokens/cost; the
  task worktree remained clean and no business source or external state
  changed. Worker report records both attempts.
- `DONE / generator`: after the user authorized a collaboration sub-agent as
  the alternate execution path, the Developer implemented all four bounded
  behaviors in the approved allowlist. Controller findings added the upstream
  current-buffer stream state and real Chat/Responses/native tool-argument
  path regressions. Worker report:
  `docs/workflow/worker-results/upstream-v0184-compat-fixes-s276-result.md`.
- `PASS / controller-review`: seven stream lifecycle regressions, five gateway
  tool-argument paths, both native bridges, both SMTP request types, version
  suffix and ordinary-version comparisons passed ten iterations. Default-tag
  affected packages and server compile passed; gofmt, diff, conflict, exact
  allowlist and protected-file hash gates passed. Full unit-tag service
  compilation remains blocked by pre-existing tests outside the S276 diff.
- `PASS / final-qa`: independent `gpt-5.6-sol` QA passed the same focused
  regressions, default affected packages, server compile, formatting, diff,
  conflict, exact allowlist and protected-hash gates. The full unit-tag service
  baseline failure remains unowned by S276. QA report:
  `docs/workflow/qa-reports/upstream-v0184-compat-fixes-s276-qa.md`.
- `PASS / local-commit`: the exact S276 product/test scope and worker/QA
  evidence are committed locally; no push, provider, database, container,
  deployment or shared-data action was performed.

# Usage Billing Multiplier Breakdown S275

- `contract-draft`: preserve the current APIMart `7 * 1.2` charge and persisted
  composite `rate_multiplier`, but project separate pricing and balance-
  conversion multipliers through user/admin usage APIs.
- User/admin badges, tooltips, details and exports will use the pricing factor
  as the visible configured rate and show balance conversion separately.
  Legacy API responses keep the current fallback behavior.
- Schema, migrations, repository SQL, billing calculation, historical rewrite,
  provider traffic, containers, deployment, shared data, commit, push and
  `outputs/**` are excluded. Contract:
  `docs/workflow/tasks/usage-billing-multiplier-breakdown-s275.md`.
- `PASS / contract-review`: the existing enriched usage projection can identify
  official-model records from immutable model fields and current APIMart API
  Key accounts from the associated Account without changing persistence. The
  additive DTO fields, legacy fallback, exact UI/export owners and explicit
  historical account-configuration limitation are decision-complete; Terra
  Developer implementation is authorized.
- `DONE / generator`: additive DTO fields now project pricing and balance-
  conversion multipliers; user/admin tooltips, details, badges and exports use
  the split with legacy fallback. Persisted composite rate, costs and billing
  paths are unchanged. Worker report:
  `docs/workflow/worker-results/usage-billing-multiplier-breakdown-s275-result.md`.
- `PASS / controller-review`: the first implementation duplicated the APIMart
  trigger; the bounded fix now calls the existing
  `apimartImageUsageMultiplierForModels` helper directly. Focused service/DTO
  tests pass x10, full service/DTO packages, server compile, 43 frontend tests,
  typecheck and diff checks passed. Independent Terra QA remains required.
- `PASS / final-qa`: independent `gpt-5.6-terra` QA passed focused service/DTO
  tests x10, full affected Go packages, server compile, 43 frontend tests,
  typecheck, formatting, exact allowlist, conflict and protected-file hash
  gates. Persisted billing fields and denied billing/schema/repository paths are
  unchanged. Report:
  `docs/workflow/qa-reports/usage-billing-multiplier-breakdown-s275-qa.md`.
- `PASS / local-commit`: the exact 13-file product/test scope is committed as
  `27daa1f2a` after focused Go tests, 43 frontend Vitest cases, typecheck and
  staged diff checks passed. It has not been pushed. The user and admin
  list-driven usage surfaces display configured pricing separately from
  APIMart balance conversion. A direct repository `GetByID` does not hydrate an
  Account, so a future independent single-row UI would need an association
  follow-up for non-official APIMart account rows; the current details drawer
  uses the hydrated list row and is unaffected.

- Earlier workflow status was archived by pge-compact at 20260831T062627697Z.

---
phase: done
current_sprint: upstream-v0184-channel-pricing-s278
total_sprints: 278
pending_action: S278 is locally integrated and independently QA PASS. Preserve unrelated commit f81bb2a55 and all dirty paths/outputs; the next selective-upstream action is a new S279 contract for group partial-update limits. Do not push unless explicitly requested.
project_type: fullstack
qa_mode: runtime
approval_required: true
last_verified: 2026-09-01 11:52 +08:00
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

# Upstream v0.1.184 Frontend Compatibility S277

- `contract-draft`: manually adapt three non-overlapping frontend behaviors
  from the diverged v0.1.179--v0.1.184 history: strict local parsing for
  `datetime-local` values, redeem batch-expiry conversion through that parser,
  and removal of the Claude attribution-header override from generated Claude
  Code settings. The existing shell snippets already omit the override; the
  remaining settings JSON must match them.
- Protected out-of-scope dirty paths include all `backend/**` edits,
  `frontend/pnpm-lock.yaml`, route breaker/auth, `admin_service.go`, Pixel
  Cafe, `outputs/**`, and any deployment/external state. No whole-history
  merge, cherry-pick, dependency refresh, provider, database, container,
  deployment, commit or push is authorized.
- Contract: `docs/workflow/tasks/upstream-v0184-frontend-compat-s277.md`.
- `DONE / generator`: strict local parser, Redeem custom expiry conversion and
  Claude settings cleanup implemented in the six-file allowlist; worker report
  is `docs/workflow/worker-results/upstream-v0184-frontend-compat-s277-result.md`.
- `PASS / controller-review`: initial test expectation was corrected to the
  contract's second truncation (`.000Z`); focused Vitest 31/31, typecheck,
  build, diff and conflict checks passed.
- `PASS / final-qa`: independent Terra QA confirmed the same tests, typecheck,
  build and protected-path digests; QA report is
  `docs/workflow/qa-reports/upstream-v0184-frontend-compat-s277-qa.md`.
- `PASS / local-commit`: exact S277 product/test scope is committed locally;
  no backend, lockfile, Pixel Cafe, knowledge, outputs or external-state
  changes are included and no push was performed.

# Upstream v0.1.184 Channel Pricing Normalization S278

- `contract-draft`: adapt `eb4237a2b` so channel pricing lookup retries with
  `normalizeKnownOpenAICodexModel` after a literal model miss. Literal and
  exact-variant entries remain first; only known OpenAI/Codex aliases with a
  changed normalized name are retried. Non-OpenAI models and unrelated channel
  entries remain untouched.
- The local owner is `backend/internal/service/model_pricing_resolver.go`; the
  existing `admin_service.go`, gateway, billing-service, repository, schema,
  migration and provider traffic changes are denied. Focused usage pricing
  regressions use the existing in-memory channel cache and usage-log stub.
- Contract: `docs/workflow/tasks/upstream-v0184-channel-pricing-s278.md`.
- `BLOCKED / generator-dispatch`: the initial Terra Developer failed at the
  model gateway with HTTP 524; a same-contract retry failed with HTTP 503.
  Neither attempt produced a valid worker report or completed QA evidence.
- `AUTHORIZED / model-exception`: the user explicitly approved
  `gpt-5.6-sol` for both the S278 Developer Worker and independent QA. This
  exception is limited to S278; the global Agent Matrix remains unchanged.
- `DONE / generator`: the Sol Developer reviewed local S278 commit `43d109581`
  and added missing unknown-OpenAI and non-OpenAI negative regressions without
  changing the resolver. Eight focused cases x10, the complete default-tag
  service package, server compile, format, diff and conflict gates passed.
- `PASS / controller-review`: the implementation is literal-first and retries
  only changed non-empty known OpenAI/Codex normalization results. The actual
  S278 base is corrected to `f81bb2a55`; that concurrent 17-file commit is
  excluded. Independent Sol QA is authorized against `43d109581` plus the
  two-file allowlisted follow-up.
- `PASS / final-qa`: independent `gpt-5.6-sol` QA passed eight focused cases
  x10, an uncached complete service run, server compile, format, diff,
  conflict, exact-scope and protected-hash gates. The concurrent parent scope
  and all current dirty paths/outputs remained unchanged.
- `PASS / local-integration`: S278 product behavior is in local commit
  `43d109581`; the missing negative regressions and complete worker/QA/workflow
  evidence are included in this closeout. No push or external-state action was
  performed.

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

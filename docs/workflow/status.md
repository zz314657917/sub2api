---
phase: done
current_sprint: usage-billing-multiplier-breakdown-s275
total_sprints: 275
pending_action: S275 product change is committed locally as 27daa1f2a; workflow evidence is recorded in the companion documentation commit and the handoff follows separately. No push, billing, schema, history, provider, container, deployment or shared-data action is authorized; preserve all unrelated dirty paths and outputs.
project_type: fullstack
qa_mode: runtime
approval_required: true
last_verified: 2026-08-31 17:21 +08:00
---

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

### DONE: upstream-cn-providers-s226-d

## Summary

Controller takeover completed after two approved Developer Worker attempts
produced no valid artifact. The frontend delta was implemented from the exact
S226-C report base plus the task-local account-modal baseline.

## Commits

- Approved base: `5bb985cb69b5db2bf27efa957bc2406cdef9e0c1`
- Non-integrated user modal baseline: `d7158e916`
- Business commit: `a559956f7`
- Report commit: recorded separately after the business commit

## Changed Files

- Added CN provider admin API, balance/quota status cells, and Base URL presets.
- Added Kimi, Zhipu, and DeepSeek create/edit controls with legal mode/protocol
  combinations and credential persistence.
- Preserved custom Base URLs across mode/protocol watcher changes.
- Added CN platform icons, badges, colors, translations, account types, and
  Kimi model whitelist support.
- Added focused coverage for credentials, modal persistence, status cells,
  account-row rendering, and Kimi model suggestions.

## Acceptance Evidence

- `npm.cmd run test:run -- src/components/account/__tests__/AccountUsageCell.spec.ts src/components/account/__tests__/CNProviderBalanceCell.spec.ts src/components/account/__tests__/CNProviderQuotaCell.spec.ts src/components/account/__tests__/CreateAccountModal.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts src/components/account/__tests__/ModelWhitelistSelector.spec.ts src/components/account/__tests__/credentialsBuilder.spec.ts`
  -> 7 files, 87 tests passed.
- `npm.cmd run typecheck` -> passed.
- `npm.cmd run build` -> passed (`vue-tsc -b` and `vite build`). Existing Vite
  dynamic-import and chunk-size warnings were non-fatal.
- `git diff --check` -> passed.
- No unresolved conflicts or index entries.
- S226 upstream provenance checks for `901a0439f`, `4b667ccd4`, and `e72854538`
  against `upstream/main` -> passed.
- Exact D allowlist check -> passed.
- No frontend manifest or lockfile changed.

## Contract Compliance

- Only S226-D allowlisted frontend paths and this report are changed.
- No database, dependency, provider, deployment, push, or shared resource was
  used.
- The user-owned modal baseline remains outside the business diff and is not
  included in the business commit.
- Independent S226-E QA remains required; no main integration is authorized by
  this result.

## Risks

- No real provider endpoint or browser session was used in D; provider behavior
  remains covered by the backend batches and mocked frontend probes. UI/browser
  inspection is deferred to S226-E.

## knowledge_candidates

- None.

### DONE: upstream-v0177-codex-fingerprint-s221

## Scope

- Temporary user baseline: `c6d4ee2304c12497d55098ba07fc91b6087f20e2`.
- This result and the following worker commit contain only S221 deltas above
  that baseline. The baseline carries the separate user patch-id
  `5d316e5b6935fdc5dbf825f940feaf231d79ac0f` and is not amended or squashed.
- Amendment 1 approved the minimal bulk-off deletion path in
  `backend/internal/repository/account_repo.go` and its default-tag test.

## Delivered

- Added explicit opt-in Codex fingerprint convergence modes. Missing, invalid,
  and `off` modes preserve client identifiers; `device`, `session`, and `full`
  converge the configured carriers.
- Normal and passthrough OpenAI request construction stage IDs per attempt,
  including nil, rewrite raw passthrough `client_metadata`, and apply staged
  headers before identity enforcement. The turn-start timestamp is resolved
  once per ID set so headers, map bodies, and raw bodies remain identical.
- Added Create/Edit/Bulk OpenAI OAuth controls and localized labels. Create and
  Edit delete the mode key for `off`; Bulk sends the narrow null delete sentinel
  for `off`, which the repository removes only as `codex_fingerprint_mode`.
- Added the S220 OpenAI long-context billing toggle to EditAccountModal with
  default-off loading/persistence and Spark-shadow exclusion, preserving the
  existing `extra-wide` and asymmetric availability layout.

## Validation

- `go test ./internal/service -run $focused -count=10` PASS (5.493s).
- `go test ./internal/repository -run '^TestBulkUpdateCodexFingerprint' -count=10` PASS (0.124s).
- `go test ./internal/service -count=1` PASS (62.000s).
- `go test ./internal/handler -count=1` PASS (27.483s).
- `go test ./internal/server -count=1` PASS (0.724s).
- `go test ./cmd/server -run '^$' -count=0` PASS (11.031s).
- Account-modal Vitest suite PASS: 4 files, 76 tests.
- `vue-tsc --noEmit` PASS.
- `vite build` PASS (21.13s); existing dynamic-import and chunk-size warnings
  remain non-fatal.
- `git diff --check`, unmerged-index check, and upstream provenance checks for
  `c0ab3a00e` and `fce41e318` PASS.

## Constraints And Risks

- No migration, dependency, lockfile, provider, deployment, container, push,
  or `outputs/` change was made. A pnpm executable test run regenerated the
  lockfile; that tool-only change was immediately reversed before validation
  and is absent from the final diff.
- Bulk deletion is deliberately limited to one null sentinel and one JSONB key;
  unrelated null-valued bulk extra fields retain existing merge behavior.

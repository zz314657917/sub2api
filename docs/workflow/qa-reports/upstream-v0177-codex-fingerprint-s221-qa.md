### PASS: upstream-v0177-codex-fingerprint-s221

# Independent QA Report

## Scope And Boundary

- Reviewed the strict S221 increment `c6d4ee2304c12497d55098ba07fc91b6087f20e2..6be50cc0dd94750a98cb06bfaa691e918a95726a`.
- The increment contains exactly 14 amended-allowlist files. No migration,
  dependency, provider, database, deployment, `outputs/`, or unrelated file
  appears in the range.
- Temporary baseline patch-id remains exactly
  `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`; it is an ancestor and is not
  included in the S221 increment. `frontend/pnpm-lock.yaml` hash equals HEAD
  after pnpm's mechanical change was restored.

## Result

No implementation defect found. Default/off fingerprint behavior is a no-op;
device/session/full are explicit OAuth opt-ins. Both normal and passthrough
paths stage IDs per attempt, including nil, then apply headers after session
isolation and before identity enforcement. Raw JSON rewriting uses gjson/sjson
and shares the same IDs as headers and decoded-body behavior. Create, Edit, and
Bulk controls enforce the four OAuth modes; bulk off removes only
`codex_fingerprint_mode`; Edit preserves the user baseline layout while adding
the S220 long-context control and Spark-shadow exclusion.

## Acceptance Evidence

- All 9 focused backend names were discovered with `go test -list`; the focused
  suite passed with `-count=10`. The bulk null-sentinel repository regression
  also passed with `-count=10`.
- Full backend QA: service exit 0 (`61.611s`), handler exit 0 (`27.561s`),
  server PASS (`0.093s`), and `cmd/server` compile PASS (`0.070s`).
- Frontend focused suite: 4 files and 76 tests PASS. `pnpm.cmd run typecheck`
  passed. Production Vite build exited 0 (`20.34s`); only existing
  Browserslist, dynamic-import, and chunk-size warnings were emitted.
- An independent temporary 4 MiB raw JSON probe passed (`5.481s`): the payload
  and unrelated `client_metadata.trace` remained unchanged while the staged
  fingerprint session value was inserted. The probe was removed before final
  Git checks.
- `git diff --check`, conflict/unmerged-index checks, and upstream provenance
  for `c0ab3a00e` and `fce41e318` passed.

## Residual Risk

No browser/provider/shared-runtime test was run, by contract. UI behavior is
covered by modal Vitest, typecheck, and production build. The requested
`docs/workflow/agent-matrix.md` is absent from both the isolated worktree and
the main-repository file search; this is workflow-documentation drift, not a
product verification failure.

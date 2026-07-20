### PASS: usage-model-reasoning-effort-s82

## Findings

- No blocking finding remains.
- The shared helper appends a normalized suffix only for meaningful effort
  values and preserves existing model-label sanitization.
- User rows and detail rendering use the record's `reasoning_effort` field.
- Admin simple rows, requested/upstream pairs, and mapping chains annotate only
  the requested model; the focused tests assert a single suffix occurrence.
- The standalone reasoning-effort column, filters, exports, API contracts, and
  backend capture paths are unchanged.

## Executed Checks

- Focused Vitest: `modelDisplay` 4/4, admin `UsageTable` 10/10, user `UsageView`
  19/19; total 33/33 PASS.
- Frontend `vue-tsc --noEmit`: PASS.
- Frontend production build (`vue-tsc -b && vite build`): PASS.
- `git diff --check`: PASS.
- Changed business paths match the six frontend paths in the approved
  contract; workflow files are limited to the S82 evidence set.
- `git status` still shows the same three pre-existing user `knowledge/**`
  modifications and no generated `backend/internal/web/dist` changes.
- No conflict markers or unmerged index entries were introduced.

## Unverified Risks

- No authenticated administrator browser smoke was run because this UI-only
  change is covered by component tests and no live service was started.
- Existing build warnings about stale `caniuse-lite`, dynamic imports, and
  large chunks remain non-blocking and unrelated.

## Recommendation

`PASS` — close S82 locally. Do not deploy, update containers, or include the
pre-existing `knowledge/**` modifications in this change.

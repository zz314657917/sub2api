### PASS: upstream-main-compat-s82

## Findings

- No blocking finding remains. The five-file business diff ports the intent of
  upstream `8b75dd557` without advertising upstream-only account-level
  `http_bridge` support.
- The local protocol resolver consults account mode only inside
  `ModeRouterV2Enabled`; with the switch disabled it follows the existing
  legacy account-enabled branch. The new documentation matches this behavior.
- Backend validation, service constants, frontend utilities, and the account
  UI vocabulary remain `off / ctx_pool / passthrough` (with legacy
  shared/dedicated normalization). No runtime or enum file changed.
- `http_bridge_enabled` remains the independent oversized-first-message
  fallback. README explicitly distinguishes it from account mode selection.
- The example config changes comments only; comment-stripped values are byte-
  equivalent by line to the S82 baseline.

## Executed Checks

- Generator focused Vitest: PASS, 1 file / 1 test.
- Generator `npm.cmd run typecheck`: PASS.
- Generator `npm.cmd run build`: PASS, 1080 modules transformed.
- Fresh Evaluator focused Vitest: PASS, 1 file / 1 test.
- Fresh Evaluator `npm.cmd run typecheck`: PASS.
- Fresh Evaluator `npm.cmd run build`: PASS, 1080 modules transformed.
- Local runtime vocabulary and protocol-resolver source review: PASS.
- README/config prerequisite and environment-form assertions: PASS.
- Comment-stripped config-value comparison against baseline: PASS.
- Pre-QA exact ten-path audit, unmerged-index scan, real conflict-marker scan,
  `git diff --check`, and all three primary-checkout protected hashes: PASS.

## Unverified Risks

- No authenticated browser smoke was run. The changed UI surface is static help
  text, and direct locale import, TypeScript, and production compilation pass.
- Existing Vite dynamic-import and large-chunk warnings remain non-blocking and
  unrelated to the S82 diff.
- Account-level `http_bridge` remains unavailable locally. Adding that runtime
  feature would require a separate contract covering config validation,
  service/frontend enums, account form behavior, routing, and regressions.

## Recommendation

`PASS` — create the scoped S82 commit after the exact eleven-path tracking,
conflict-marker, cached-diff, and protected-hash gates pass. Keep the branch
isolated; do not merge, push, deploy, or update containers automatically.

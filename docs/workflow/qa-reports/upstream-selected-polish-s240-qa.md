### PASS: upstream-selected-polish-s240

## Independent QA identity

- Independent QA was executed with native `gpt-5.6-sol` after the required
  `gpt-5.6-terra` route returned `403 NO_MATCHING_GROUP_ROUTE` before inference.
- The user explicitly authorized `gpt-5.6-sol` as the named replacement model
  for this S240 QA gate; this is not a silent fallback.
- QA worktree:
  `E:/codex-worktrees/sub2api/upstream-selected-polish-s240-qa`.
- Frozen product range:
  `771af745c5387f8b3d829d152b2b9e6062990233..c444c1168`.
- The pre-existing QA branch head `265f1386f` contained only the earlier routing
  blocker report and was excluded from the product diff under review.

## Findings

- No blocking or non-blocking product finding was identified in the frozen
  candidate diff.
- The exact allowlist matched 18/18 paths: all 17 approved frontend
  implementation/test paths plus the approved Controller result report, with
  zero extra and zero missing paths.
- Dashboard token detail preserves the active-key note and exposes input,
  output, and combined cache-creation/cache-read totals.
- Native dark controls, empty-window SLA neutrality, Fast/Flex target/other
  model wording and summary behavior, announcement creation copy, and
  independent proxy/group helper loading match the contract and focused tests.
- The apparent conflict-marker grep hit in
  `backend/internal/pkg/antigravity/request_transformer.go` is an existing
  separator inside a string literal, not an unmerged marker; `git ls-files -u`
  was empty.

## Commands and results

- Focused Vitest over the five contract files: 5 files / 14 tests PASS in each
  of 10 independent consecutive runs.
- `vitest run src/views/admin/__tests__/SettingsView.spec.ts
  src/__tests__/style-theme.spec.ts`: 2 files / 29 tests PASS
  (`SettingsView.spec.ts` 27, `style-theme.spec.ts` 2).
- `vue-tsc --noEmit`: PASS.
- `vite build`: PASS, 1880 modules transformed. Existing dynamic-import,
  Browserslist-age, and chunk-size warnings remain non-fatal.
- `git diff --check 771af745c..c444c1168`: PASS.
- Exact allowlist comparison: expected 18, actual 18, delta 0.
- Unmerged index: 0 entries. QA worktree was clean after validation cleanup.
- All six source commits (`a6d868f27`, `7d796f111`, `0d5e3ca9b`,
  `22fc0cdbf`, `35e8ba2a3`, `1977810cf`) are ancestors of
  `upstream/main@67380eafd`.
- Frozen base is an ancestor of both candidate `c444c1168` and current
  `main@01ad28d5b`; `main` and `origin/main` both remained at `01ad28d5b`.

## Dependency and workspace protection

- QA had no local `node_modules`. Direct executables came from
  `E:/codex-worktrees/sub2api/upstream-selected-polish-s240/frontend/node_modules/.bin`
  while all commands ran with the QA `frontend/` as cwd.
- Because Node ESM resolves imports from `vitest.config.ts` relative to the QA
  checkout, a temporary QA `frontend/node_modules` directory junction pointed
  to that same validated dependency tree. It was removed after testing. No
  install ran and no dependency or lockfile changed.
- The protected main worktree HEAD and its pre-existing dirty/untracked path
  list were unchanged before and after QA. No main file was staged or edited,
  and no provider, database, container, deployment, remote-ref, or push action
  occurred.

## Residual risk

- This low-risk polish batch was validated by component/static tests,
  typecheck, production build, diff review, and scope/provenance gates. No
  authenticated browser acceptance was required by the contract, so
  browser-only visual nuances remain the principal residual risk.
- Final integration must replay onto current `main@01ad28d5b` and rerun the
  relevant combined-main checks; this QA verdict covers only the frozen
  `771af745c..c444c1168` candidate.

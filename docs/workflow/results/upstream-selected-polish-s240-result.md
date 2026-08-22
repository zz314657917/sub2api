### PASS: upstream-selected-polish-s240

## Scope

- Frozen base: `771af745c5387f8b3d829d152b2b9e6062990233`.
- Candidate head: `101151c49`.
- Selectively adapted sources: `a6d868f27`, `7d796f111`, `0d5e3ca9b`,
  `22fc0cdbf`, `35e8ba2a3`, and `1977810cf`.
- The candidate contains only the 17 approved frontend implementation/test paths.
  No backend, migration, dependency, generated, provider, database, container,
  deployment, push, `knowledge/*`, or `outputs/*` change is included.

## Implemented behavior

- User dashboard total-token card now keeps the active API-key note and adds
  total input, output, and combined cache-creation/cache-read token detail.
- Native date/time/select controls opt into the active light/dark color scheme.
- Ops SLA is neutral and displays `-` when the selected window has no SLA
  requests; the exception count no longer reports a fixed red state.
- OpenAI Fast/Flex rules use target/other-model language and render an explicit
  action summary while preserving the local experimental scheduler controls.
- Announcement empty state uses a creation prompt instead of a load-error copy,
  with the current split `admin/announcements.ts` locale topology.
- Account helper proxy and group requests load independently, so one failure
  does not suppress the other result.

## Verification

- Focused Vitest: 5 files / 14 tests PASS; repeated 10 times PASS.
- Adjacent regression: `SettingsView.spec.ts` 27/27 PASS and
  `style-theme.spec.ts` 2/2 PASS. The first run exposed a missing test helper;
  `openGatewayTab` was added and the full pair then passed.
- `vue-tsc --noEmit`: PASS.
- `vite build`: PASS (1880 modules; existing dynamic-import and chunk-size
  warnings only).
- `git diff --check 771af745c..101151c49`: PASS.
- Exact allowlist: 17 changed, 0 extra, 0 missing.
- Unmerged index: empty. Candidate worktree: clean.
- All six source commits exist and are ancestors of
  `upstream/main@67380eafd`.

## Environment note

- The first `pnpm exec` attempted dependency bootstrap and stopped at pnpm 11's
  ignored-build policy for `esbuild`/`vue-demi`. Dependencies were present, so
  validation used the checked-out worktree's direct `.bin` executables. This
  did not modify dependencies or the committed lockfile; generated pnpm files
  were removed before scope validation.

## Protected main

- The isolated candidate did not edit the main worktree or stage its dirty
  files. During validation another local operation advanced `main` and
  `origin/main` from `771af745c` to `01ad28d5b`; the new commit only changes the
  admin usage table and remains a descendant of the frozen base.
- Final integration must therefore be replayed and reverified on
  `main@01ad28d5b` after independent QA. No push is authorized.

## Residual risk

- UI behavior is covered by component/static tests, typecheck, and production
  build; no authenticated browser session was used for this low-risk polish
  batch.

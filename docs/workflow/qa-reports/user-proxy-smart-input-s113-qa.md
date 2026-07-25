### PASS: user-proxy-smart-input-s113

## Findings

- The shared frontend parser accepts standard proxy URLs and the compact
  `scheme://host:port:user:password` form, including `socks5h`, IPv6 brackets,
  encoded credentials, and passwords containing additional colons.
- The user proxy modal fills the existing structured create/update payload and
  supports validated multi-line creation; the admin batch importer reuses the
  same parser with an explicit scheme.
- No backend, schema, deployment, or container files were changed.

## Executed Checks

- `npm.cmd run test:run -- src/utils/__tests__/proxyInput.spec.ts src/views/user/__tests__/MyAccountsView.importFile.spec.ts src/__tests__/integration/proxy-data-import.spec.ts -- --pool=threads --maxWorkers=1 --minWorkers=1` -> 3 files / 20 tests passed, including multi-line creation and invalid-batch rejection.
- `npm.cmd run typecheck` -> passed.
- `NODE_OPTIONS=--max-old-space-size=4096 npm.cmd run build` -> passed, 1091 modules transformed (final rerun after multi-line user proxy support).
- Targeted ESLint over all changed frontend source/test/locale files -> passed.
- `git diff --check` -> passed.
- Local Vite server started on `127.0.0.1:5173`; Playwright reached the login redirect. An authenticated proxy-modal save smoke was unavailable.

## Unverified Risks

- Full frontend ESLint remains blocked by three pre-existing errors outside this
  change (`AccountTableFilters.spec.ts` reserved `Select`, and two
  `TutorialView.vue` `ScrollBehavior` errors).
- Real authenticated API persistence and a live proxy connection were not run.
- Existing backend debug/legacy logs that print complete proxy URLs remain
  outside this frontend-only contract and should be handled in a separate
  credential-redaction task.

## Recommendation

PASS / source-only. The requested input format is ready for the existing
structured proxy APIs; no deployment or container refresh was performed.

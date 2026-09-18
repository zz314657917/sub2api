### PASS: upstream-api-key-provider-filter

# Developer Result

## Changed Files

- `frontend/src/utils/keyGroupProviders.ts`: local-platform provider classifier with an own-property guard; unknown runtime values, including prototype property names, fall back to `other`.
- `frontend/src/utils/__tests__/keyGroupProviders.spec.ts`: classification and unknown-platform coverage.
- `frontend/src/views/user/KeysView.vue`: create-only accessible provider fieldset, filtered default-group selector, reopen/delayed-load reconciliation. Edit and route selectors keep the full authorized group list.
- `frontend/src/views/user/__tests__/KeysView.spec.ts`: mounted component tests for filtering, stale selection clearing, delayed groups, dialog reopen, create payload with cross-provider routes/timeouts, and edit route/timeouts.
- `frontend/src/i18n/locales/en/keys.ts` and `frontend/src/i18n/locales/zh/keys.ts`: localized labels and hints using only local supported platforms.

## Acceptance Evidence

- `node_modules/.bin/vitest.cmd run src/views/user/__tests__/KeysView.spec.ts src/utils/__tests__/keyGroupProviders.spec.ts` — PASS, 2 files / 11 tests.
- `node_modules/.bin/vue-tsc.cmd --noEmit` — PASS.
- `node_modules/.bin/eslint.cmd src/views/user/KeysView.vue src/views/user/__tests__/KeysView.spec.ts src/utils/keyGroupProviders.ts src/utils/__tests__/keyGroupProviders.spec.ts src/i18n/locales/en/keys.ts src/i18n/locales/zh/keys.ts` — PASS.
- `node_modules/.bin/vite.cmd build --outDir E:/codex-runtime/pge/sub2api/upstream-api-key-provider-filter/dist` — PASS. Existing Browserslist age and Rollup chunk-size warnings only.
- `git diff --check` — PASS.

## Contract Compliance

- No backend/API payload field was added for provider selection. The provider is UI state only.
- Filtering is limited to the create dialog default group. `groupOptions`, route options, editing, multi-group ordering and first-response timeout serialization stay on their existing paths.
- No commit, push, deployment, package/lockfile or controller-only workflow file was changed.

## Limits

- Browser visual smoke and independent QA are pending the separately assigned QA worker.

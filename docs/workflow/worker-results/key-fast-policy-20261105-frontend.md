### DONE: key-fast-policy-20261105

Frontend Developer implementation is complete and frozen. Independent QA and the
Controller's built-page browser acceptance remain separate gates. The full
acceptance command is **not all-green**: its two failures exactly match the
pre-task baseline and the verified baseline-source overlay.

## Contract and ownership

- Candidate: `E:/codex-worktrees/sub2api/key-fast-policy-20261105`.
- Contract SHA256: `f54335d5d9cc8199c6b4de1b52aa57d8193de94e19b31c6442f1894d48b1cbce`,
  checked against the independent round-2 PASS and rechecked after implementation.
- Only the following Frontend Developer allowlist paths were edited or added:
  1. `frontend/src/types/index.ts`
  2. `frontend/src/api/keys.ts`
  3. `frontend/src/views/user/KeysView.vue`
  4. `frontend/src/i18n/locales/zh/keys.ts`
  5. `frontend/src/i18n/locales/en/keys.ts`
  6. `frontend/src/api/__tests__/keyFastPolicy.spec.ts` (new)
  7. `frontend/src/views/user/__tests__/keyFastPolicy.spec.ts` (new)
  8. `docs/workflow/worker-results/key-fast-policy-20261105-frontend.md` (this report)
- No edits to existing tests, backend, shared status/spec, dependencies, or main
  product files. No installation, commit, push, provider call, migration,
  container/service change, or browser session was initiated by this Developer.

## Implementation

- Added optional `ApiKey.openai_fast_policy` for old responses without the field,
  the exact three-value `OpenAIFastPolicy` union, and create/update request types.
- `keysAPI.create` appends the optional policy after all eleven existing
  parameters. Axios serializes an explicitly supplied policy, including
  `passthrough`. Existing calls that omit it retain their existing payload.
- The existing KeysView form uses the existing Select inside a visible labelled
  fieldset with `data-testid="key-openai-fast-policy"`. Chinese choices are exactly
  `透传（默认）`, `强制启用`, `过滤`; matching English choices are local to keys locale.
- Form initialization, old/missing response fallback, and close/success reset
  use `passthrough`. Edit loads the saved policy and sends an explicit policy only
  when changed, including switching force/filter back to passthrough.
- Default create preserves the existing eleven-argument call and relies on the
  server's passthrough default; force/filter append argument twelve. The original
  exact-argument test passes without alteration.
- Guidance explains OpenAI-only scope, Fast = priority, group force precedence,
  global final authority, existing priority prices, and WS reconnect.
- No managed-key endpoint, entitlement bypass, or fallback write was added.
  Ordinary unchanged-policy edits omit the field. Mounted/API tests verify that
  a managed-key denial is surfaced without an alternate write. Actual server
  ownership enforcement remains the backend Developer/independent QA scope.

## Frozen source and dist

- Frozen at `2026-11-05T01:49:18.854Z`. No source or dist changes after freezing.
- Manifest:
  `E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-frontend/frontend-freeze.json`
- Manifest SHA256:
  `bbf2334ba576922396474536a9a54eb4696811675ebbc078ce3fd27acf94fc40`
- Seven source/test entries, set SHA256:
  `78b9876f5ffa97a776b23c852de7f5806eea59a8ec7d612e0f6b340fdeace4ce`
- Completed production dist:
  `E:/codex-runtime/sub2api/key-fast-policy-20261105/dist`
- 259 dist files, set SHA256:
  `2f2e93c6288572429e7760f9463ba79bcade7f5e6c451990c0b0fee1294a6b93`
- Set digests are SHA256 of UTF-8 `JSON.stringify` of the corresponding ordered
  manifest entry arrays. Entries also contain individual raw-file SHA256/size.
- Report is excluded from the source freeze so delivery documentation can finish
  while the Controller uses the frozen build. Notify Controller before any
  subsequent source change; do not silently rebuild or replace this dist.

## Findings and baseline attribution

The following two original failures were present before edits (369 tests,
367 passed / 2 failed), remain the only failures after edits (404 tests,
402 passed / 2 failed), and reproduce with baseline source supplied through a
Vite in-memory overlay (26 tests, 24 passed / the same 2 failed):

1. `frontend/src/views/user/__tests__/GroupBuyView.spec.ts:263`
   — `GroupBuyView switches to the Pixel Cafe shell when the opt-in flag is enabled`.
   Mount at line 266 throws `SyntaxError: Need to install with \`app.use\` function`
   from existing `frontend/src/components/common/ConfirmDialog.vue:39`
   (`useI18n`). The text assertion at line 269 (`像素网吧`) is not reached.
   The existing test mount does not install i18n or stub this nested dialog.
2. `frontend/src/views/user/__tests__/KeysView.createQuery.spec.ts:445`
   — `KeysView create query does not overwrite a touched default group when rates load after the dialog opens`.
   Exact failure at line 465:
   `expect(setupState.formData.group_id).toBe(2)` receives `null`.
   This concerns the existing create-provider/default-group watcher and delayed
   group-rate test setup; it is not a create-argument-count failure.

No assertion was weakened, test skipped, original test edited, or baseline defect
fixed in this task. The existing exact eleven-argument assertion at
`frontend/src/views/user/__tests__/KeysView.spec.ts:376` passed:
`user KeysView create provider selection clears a stale default group when switching provider and does not put provider in create payload`.

For overlay reproduction, each of the five pre-task product sources was extracted
from base commit `0f6b739881ccf9e4ad423f464aeedcc9abd7f48a` and verified against
`before.json` raw-file SHA256 before use. Vite's pre-load plugin supplied those
bytes without editing frozen files. The exercised runtime modules were
`src/views/user/KeysView.vue` and `src/api/keys.ts`; type-only and unused locale
modules did not need runtime loading. Both original test files,
ConfirmDialog.vue, and PixelCafePage.vue were also verified unchanged.

## Executed checks

All frontend commands ran from the candidate's `frontend` directory using
existing junction dependencies. Vitest cache was disabled to avoid writing its
cache through the shared node_modules junction.

| Check | Result | Evidence under developer-frontend |
| --- | --- | --- |
| Pre-edit existing user + Group Fast suites | 369 total: 367 passed, 2 original failures | `baseline-vitest.log` |
| New mounted and API tests | 35/35: mounted 20, Axios serialization/API 15 | `new-tests.log`, `new-tests.json` |
| Complete contract frontend Vitest selection | 404 total: 402 passed, same 2 original failures | `acceptance-vitest.log`, `acceptance-vitest.json` |
| Existing KeysView suites within acceptance | 40 total: 39 passed, same 1 original failure | `acceptance-vitest.json` |
| Existing Group Fast regressions within acceptance | 26/26: mounted 14, group/type 12 | `acceptance-vitest.json` |
| Verified baseline source overlay | 26 total: 24 passed, exact matching 2 original failures | `baseline-overlay-vitest.log`, `baseline-overlay-vitest.json`, `baseline-overlay-inputs.json`, `baseline-overlay-loads.json` |
| `vue-tsc --noEmit` | exit 0 | `vue-tsc.log` |
| Scoped eslint on seven source/test files | exit 0 | `eslint.log` |
| Vite production build to approved external dist | exit 0; 259 files | `vite-build.log`, `frontend-freeze.json` |
| Scoped `git diff --check` | exit 0 | `scope-verification.json` |
| Source/dist freeze recheck and failure-name comparison | 7 + 259 hashes match; failure names/messages match overlay | `freeze-and-failure-verification.json` |

Exact commands:

```powershell
./node_modules/.bin/vitest.cmd run src/views/user/__tests__ src/views/admin/__tests__/GroupsView.openaiFast.spec.ts src/views/admin/__tests__/groupsOpenAIFast.spec.ts --cache=false
./node_modules/.bin/vitest.cmd run src/views/user/__tests__/keyFastPolicy.spec.ts src/api/__tests__/keyFastPolicy.spec.ts --cache=false --reporter=default --reporter=json --outputFile.json=E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-frontend/new-tests.json
./node_modules/.bin/vitest.cmd run src/views/user/__tests__ src/api/__tests__/keyFastPolicy.spec.ts src/views/admin/__tests__/GroupsView.openaiFast.spec.ts src/views/admin/__tests__/groupsOpenAIFast.spec.ts --cache=false --reporter=default --reporter=json --outputFile.json=E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-frontend/acceptance-vitest.json
./node_modules/.bin/vitest.cmd run --config E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-frontend/baseline-overlay.config.mjs --cache=false
./node_modules/.bin/vue-tsc.cmd --noEmit
./node_modules/.bin/eslint.cmd src/types/index.ts src/api/keys.ts src/views/user/KeysView.vue src/i18n/locales/zh/keys.ts src/i18n/locales/en/keys.ts src/api/__tests__/keyFastPolicy.spec.ts src/views/user/__tests__/keyFastPolicy.spec.ts
./node_modules/.bin/vite.cmd build --outDir E:/codex-runtime/sub2api/key-fast-policy-20261105/dist
```

From candidate root:

```powershell
git diff --check -- frontend/src/types/index.ts frontend/src/api/keys.ts frontend/src/views/user/KeysView.vue frontend/src/i18n/locales/zh/keys.ts frontend/src/i18n/locales/en/keys.ts
```

## Unverified risks and recommendation

- New mounted tests exercise the actual Select, form submit, create/edit/reset,
  both locale strings, saved response replay, missing legacy fields, retry state,
  and permission-denial handling. API tests exercise real Axios JSON transforms
  at an intercepted adapter; no external service is contacted by these new tests.
- No Developer browser claim is made. Controller may run the prepared built-page
  CRUD/browser smoke in parallel against this immutable dist and must record its
  own accessibility, responsive, theme and process-cleanup evidence.
- Backend persistence, policy precedence, billing and managed-key enforcement
  require their separately owned tests and independent QA.
- Build reported existing large-chunk/Browserslist-age warnings. Dependencies
  were not installed or updated.
- Ready for independent frontend QA against the recorded freeze, with the exact
  two baseline failures retained and disclosed; no all-green full-suite claim.

knowledge_candidates: none

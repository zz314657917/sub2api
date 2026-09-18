### PASS: upstream-api-key-provider-filter

# Independent QA report

## Verdict

Independent `gpt-5.6-terra` QA passed the frozen behavior port. The implementation follows the amended local locale owners (`en/keys.ts`, `zh/keys.ts`), and it is limited to the approved frontend/test paths. No provider request, real credential, real key write, commit, merge, deployment, or push occurred.

## Findings

- `getKeyGroupProvider` classifies only configured `GroupPlatform` values: Anthropic and OpenAI retain dedicated categories; Kimi/Zhipu/DeepSeek use domestic; Gemini/Grok/Antigravity and unknown runtime strings use other. The own-property guard correctly rejects `constructor`, `toString`, and `__proto__` as unknown.
- The real mounted `KeysView` tests cover misleading display names, provider switching and stale default clearing, disabled/empty and delayed group loading, dialog reopen, and payload isolation. Both real `keysAPI.create` and `keysAPI.update` calls are asserted without a provider/category payload property.
- Cross-provider routing remained intentionally unfiltered. The create test changes the default provider while retaining the route group and `first_response_timeout_seconds: 12`; edit reconstructs all route groups and asserts the ordered `12`/`0` route-timeout representation.
- Isolated visual mock data rendered an Anthropic group named `Misleading GPT name`, an OpenAI group named `Misleading Claude name`, domestic, other and unknown-platform groups, plus an edited cross-provider routed key. Create exposes the provider fieldset and category hint; a real click on the domestic label selected it. Edit has no provider fieldset and visibly preserves the default group, two cross-provider route rows, and `12`/`0` second timeout values.
- Screenshots and DOM layout checks at 1440x900 and 390x844 show no horizontal document overflow. The mobile create and edit dialogs fit the 390px viewport; the provider radio grid remains four visible choices and the edit dialog preserves route controls.

## Executed evidence

All commands below were run independently after the Developer freeze.

| Command | Exit | Result |
| --- | ---: | --- |
| `node_modules/.bin/vitest.cmd run src/views/user/__tests__/KeysView.spec.ts src/utils/__tests__/keyGroupProviders.spec.ts` | 0 | 2 files, 11 tests passed. |
| `node_modules/.bin/vue-tsc.cmd --noEmit` | 0 | Passed. |
| `node_modules/.bin/eslint.cmd src/views/user/KeysView.vue src/views/user/__tests__/KeysView.spec.ts src/utils/keyGroupProviders.ts src/utils/__tests__/keyGroupProviders.spec.ts src/i18n/locales/en/keys.ts src/i18n/locales/zh/keys.ts` | 0 | Passed. |
| `node_modules/.bin/vite.cmd build --outDir E:/codex-runtime/pge/sub2api/upstream-api-key-provider-filter/dist` | 0 | Passed; only existing Browserslist age, dynamic-import topology and chunk-size warnings. |
| `git diff --check` | 0 | Passed. |
| `git diff --cached --check` | 0 | Passed; no staged candidate. |
| Task Vite mock server | 0 | `127.0.0.1:4173`, task-owned launcher PID `63640`, listener PID `67068`. |
| Playwright mock UI | 0 | Named session `pge-provider-filter-qa-20260918`; create/edit at 1440x900 and 390x844. |
| Browser/Vite cleanup recheck | 0 | Session close succeeded; no listener on `4173`, no owned `cliDaemon`/Chrome process, and task profile no longer exists. |

## Mock UI isolation and artifacts

- Browser session: `pge-provider-filter-qa-20260918`.
- Browser process at start: CLI daemon PID `58268`, Chrome PID `48872`, profile `C:/Users/Administrator/AppData/Local/Temp/playwright_chromiumdev_profile-Swfyns`. This was a newly created Playwright profile, not a user Chrome profile.
- The only API fixtures were Playwright route responses for local `http://127.0.0.1:4173/api/v1/**`; the handler returned synthetic keys/groups/settings and fail-closed `503` for unexpected local API paths. The fixture never sent a request outside localhost and never submitted a mutation.
- Screenshot artifacts: `E:/codex-runtime/pge/sub2api/upstream-api-key-provider-filter/.playwright-cli/page-2026-09-18T04-39-34-164Z.png` (desktop create), `page-2026-09-18T04-40-57-635Z.png` (mobile create), `page-2026-09-18T04-41-33-530Z.png` (mobile edit), `page-2026-09-18T04-41-50-510Z.png` (desktop edit).
- The Windows Playwright CLI was invoked through existing `npx.cmd --package @playwright/cli`; the provided Bash wrapper could not run because this host has no WSL. This changed only the launcher, not the isolated session or browser behavior.

## Diff and scope audit

The worktree contains only controller workflow records plus the contract allowlist: `KeysView.vue`, `KeysView.spec.ts`, `keyGroupProviders.ts`, `keyGroupProviders.spec.ts`, and the two local `keys.ts` locale files. No backend, package/lock, store, shared-main, knowledge, or runtime configuration path was modified. The untracked utility/spec are both approved allowlist paths.

## Unverified scope

This is local component/build and mock-browser acceptance only. Real provider traffic, authenticated server behavior, persistent API-key mutation, database state, container startup, deployment, commit/merge and push are out of scope and unverified.

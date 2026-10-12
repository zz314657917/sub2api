### PASS: upstream-v0215-20261009 — scoped local integration only

## Findings / Recommendation
- Three independent child QA reports PASS; controller reviewed candidate deltas, approved only minimal in-scope integration, and reran tests/build in the original shared working tree.
- Integrated **64 business/test files**: 36 production files, 26 new regression-test files and 2 narrowly updated existing assertions. Protocol 14 paths, frontend 38, WS/OAuth 12.
- No whole-release merge, commit, push, deployment, dependency upgrade, database migration, production data operation or paid provider call.
- Full frontend typecheck/production bundling are **not PASS**: baseline Airwallex SDK is absent. Full unoverlaid service tagged tests are **not PASS**: pre-existing test-fixture compilation drift remains.

## Delivered Behavior
### Protocol
- Preserve developer/system roles; flatten named Chat function tool_choice.
- Deterministically pair legacy function calls/results without modern-ID collisions.
- Preserve Responses refusals in Chat streaming, buffered and nonstreamed results.
- Include signature even when empty in Anthropic thinking blocks, preserving ordinary block/cache-control behavior.
- Guard negative buffered content-block indexes in four local owners.
- Forward Anthropic keepalive as valid SSE heartbeat comments rather than incompatible events.

### Frontend
- Preserve integer/zero RPM edits, reject invalid/fractional values; isolate stale group-multiplier responses.
- Autofill channel prices only for a new empty entry's first single model.
- Keep scientific-notation magnitudes and sub-byte units; replace misleading fixed x2 with loaded bilingual long-context labels; preserve zero pricing multipliers.
- Escape closes only the top dialog; disabled selectors close; backup polling cannot restart after unmount.
- Isolate stale cleanup, scheduled-test-result and monitor-template responses.
- Prevent saving Ops settings until successful current-generation load; validate whole-minute alert durations.
- Correct failed announcement-read confirmation and share concurrent step-up verification.
- Independent QA reproduced three real OpsSettingsDialog overlapping-request failures. Original Developer added generation/show/unmount guards and formal regressions; the same QA reran all original external reproductions without changing their assertions.

### WS / OAuth
- Refresh same-identity group billing prices per turn through the existing auth cache; freeze pricing-only snapshots for asynchronous accounting while preserving routing/user/trial/subscription/channel ownership.
- Cover passthrough's actual subsequent BeforeRequest hook, not only BeforeTurn; preserve first-turn fallback.
- Propagate image input usage through relay, per-turn/session totals and adapter; prevent repeated explicit-ID terminal frames from double accounting while still forwarding wire frames.
- Precisely and once recover HTTP encrypted reasoning signature rejection using the existing replay/trim safeguards.
- Declare cached-only web_search for OAuth history replay, including Lite additional_tools; preserve existing tool choices and API-key/compact behavior.

## Independent Gates
- Task-scoped inherited-current-session-model exception was explicitly offered after gpt-6.1-sol CLI route returned HTTP 403, and the user confirmed continuation. Global/repository matrix was not changed.
- Independent contract review PASS for each child, QA contract and the two narrowly necessary existing-test assertion updates.
- Protocol QA: full two packages 381 top-level /136 subtests PASS; service relevant regression 15 top-level /9 subtests PASS.
- WS/OAuth QA: new boundary tests, complete WS relay package, 32 pricing/ledger/context protection regressions, repeated tests and loopback HTTP/WS simulations PASS.
- Frontend QA: 128 mounted/related tests PASS and six independent external checks PASS after retesting its initial three real failures. No browser/visual acceptance claim.
- Reports:
  - `F:/mcplugins/sub2api/docs/workflow/qa-reports/upstream-v0215-protocol-20261009-qa.md`
  - `F:/mcplugins/sub2api/docs/workflow/qa-reports/upstream-v0215-frontend-20261009-qa.md`
  - `F:/mcplugins/sub2api/docs/workflow/qa-reports/upstream-v0215-ws-oauth-20261009-qa.md`

## Original Working-Tree Integration Checks
- Before integration, rehashed all **4646 protected original files** against the pre-dispatch manifest; no drift.
- Candidate diff compared against copied dirty baseline, not merely HEAD; exactly one contract owner per changed business/test path, no out-of-scope changes/deletions.
- Existing dirty hunks, including Type:message, ThinkingDisabled and group-modal cleanup changes, retained.
- All 64 integrated digests equal independently reviewed candidate digests. Backups are external; no shared status/spec/main-log/current-task file overwritten.
- Staged index unchanged; HEAD remains `ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b`.
- Added 18 task-specific workflow documents and this integration report; no pre-existing workflow document overwritten.
- Whole-tree diff-check still reports pre-existing trailing whitespace in `knowledge/tasks/current-task.md:153-154`; that file's protected digest is unchanged. Scoped task business diff-check is the applicable gate; no unrelated whitespace cleanup.

## Fresh Post-Integration Commands
Working directory for Go: `F:/mcplugins/sub2api/backend`.

| Command | Exit / result |
| --- | --- |
| `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity ./internal/service/openai_ws_v2 -count=1 -json` | 0; **420 top-level +152 subtests PASS**, zero failures/skips |
| `go test -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/source-fixture-overlay.json ./internal/handler ./internal/service -run '^TestV0215' -count=1 -json` | 0; **14 top-level +52 subtests PASS** |
| Same six-fixture overlay, 32-test pricing/ledger/image/context preservation regex from WS QA report | 0; **32 top-level +11 subtests PASS** |
| `go build -o E:/codex-runtime/upstream-v0215-20261009/source-server.exe ./cmd/server` | 0; actual production build, no overlay |
| Source-root Vitest external config, all 17 new files plus 11 related existing files | 0; **28 files /128 tests PASS** |
| `node node_modules/vue-tsc/bin/vue-tsc.js --noEmit` from original frontend | 2; only baseline TS2307 missing `@airwallex/components-sdk` |

- First source-root frontend invocation reused the QA list's absolute candidate paths, so only 11 existing files/46 tests matched. This was not reported as complete coverage; controller corrected the source-root file expansion and reran **all 28 files/128 tests** successfully.
- Fixture overlay contains six pre-existing `_test.go` mappings only, retaining all 142 test-function names/order. No production replacements or suppressed new tests. It is external and **not integrated into the repository**.
- Controller additionally executed the three expanded WS QA failures against the **immutable original baseline source copy**, with only the same six fixture-compile mappings relocated to that baseline path. All three failed with the same original reasons. Actual fresh baseline evidence, not only static attribution:
  - `TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost`: missing long-image-model fixture price.
  - `TestApplyCodexOAuthTransform_PreservesAllowedTools`: original Modified expectation.
  - `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested`: original call-ID hash expectation.
- These unrelated failures remain unresolved and prevent any full-suite PASS claim.

## Evidence / Recovery
- Isolated candidate branch: `codex/upstream-v0215-20261009`.
- Isolated checkout: `E:/codex-worktrees/sub2api/upstream-v0215-20261009`.
- Runtime root: `E:/codex-runtime/upstream-v0215-20261009`.
- Source protection/deltas/backups: `source-before.json`, `baseline.json`, `integration-manifest.json`, `integration-applied.json`, `integration-backup/`.
- Source retest logs: `source-packages.jsonl`, `source-focused.jsonl`, `source-preservation.jsonl`, `source-frontend-final.log`, `source-typecheck.log`.
- Fresh baseline attribution: `controller-baseline-three-failures.jsonl`, `baseline-fixture-overlay.json`.
- Independent QA runtime logs preserved in `qa-protocol-20261010/`, `qa-frontend-20261010/`, `qa-ws-oauth-evidence/`.
- Worktree remains for inspect/recovery; no server/browser/helper process intentionally left running.

## Unverified / Deferred
- No full repository/service suite, race detector, browser visuals, live auth-cache invalidation, production database/provider or deployed acceptance.
- Airwallex missing dependency blocks complete frontend typecheck/build; no install, externalization or production stub used to fabricate PASS.
- Duplicate explicit-response-ID map grows with completed turns for a relay lifetime; no long-connection memory pressure test.
- Deferred: Cline, Command Code, provider/platform/profile architecture, platform constraint migrations, Go/dependency/CI upgrades, TPS feature batch, monitor V2, namespace performance refactor, payment lifecycle batch and remaining upstream model/catalog changes.
- Future authorized commit/release must stage only task-owned hunks and run production acceptance after resolving environment blockers; deployment requires a separate explicit request.

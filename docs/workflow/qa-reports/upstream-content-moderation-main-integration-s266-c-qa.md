### PASS: upstream-content-moderation-main-integration-s266-c

# QA Report

## Task ID
upstream-content-moderation-main-integration-s266-c

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/upstream-content-moderation-main-integration-s266-c.md`
- Retest amendment/review present in evidence history: `2e3e336a9`, `37fe41015`.

## Retest Context
- The first QA attempt was correctly `BLOCKED`: the original contract supplied an aggregate outputs digest but omitted its serialization algorithm, so QA could not verify the protected untracked tree.
- The amended contract now defines the exact PowerShell manifest algorithm. This retest executed that snippet verbatim from `F:/mcplugins/sub2api` before and after all acceptance checks.

## Evidence
- main state: `HEAD=f080bbd094e2afd196ada2c00fbcfe7b86275361`; tracked worktree and index clean; only `?? outputs/`; unmerged and staged indexes empty.
- outputs protection: both runs returned `count=20`, `manifest=2996311A4EC1458EEC9C2AE4327D5D5EAA695C878783DE984AF841BBF0A79145`.
- provenance: `6054b9266` retains `-x c2cd7a0a1`; `f080bbd09` retains `-x eeed2369f`; A is an ancestor of B and the integration baseline `2a3664747` is an ancestor of A.
- patch IDs: main A `922e3bc0fc4ccf5c1bd1fecc41e33e8894ac8d0a`; main B `1c0333b260eb320ffeb89982757756a9e8c26df1`; both match the reviewed product commits.
- scope: A=21 paths, B=56 paths, exact union=67 paths, `2a3664747..HEAD`=67 paths; no missing or extra paths, no overlap with the frozen main delta, and no denied path hit (including Pixel Cafe, wallet/billing, workflow, lockfile, or outputs).
- integrity: `git diff --check 2a3664747..HEAD` passed; `gofmt -d` was empty; `git ls-files -u` and `git diff --cached --name-only` were empty.

## Commands
```text
contract outputs PowerShell snippet (before acceptance) -> PASS, 20 files / 2996311A4EC1458EEC9C2AE4327D5D5EAA695C878783DE984AF841BBF0A79145
go test ./internal/service -list 'Cyber|ContentModeration|KeywordMatcher|OpenAI.*Policy' -> PASS; non-empty discovery
go test ./internal/handler -list 'Cyber|OpenAI' -> PASS; non-empty discovery
go test ./internal/handler/admin -list 'ContentModeration|Cyber|Settings|RiskControl' -> PASS; non-empty discovery
go test ./internal/repository -list 'Cyber|ContentModeration|OpsError' -> PASS; non-empty discovery
go test ./internal/service -run 'Cyber|ContentModeration|KeywordMatcher|OpenAI.*Policy' -count=10 -> PASS, 8.322s
go test ./internal/handler -run 'Cyber|OpenAI' -count=10 -> PASS, 10.495s
go test ./internal/handler/admin -run 'ContentModeration|Cyber|Settings|RiskControl' -count=10 -> PASS, 0.176s
go test ./internal/repository -run 'Cyber|ContentModeration|OpsError' -count=10 -> PASS, 0.185s
go test ./migrations -run 'ContentModerationMatchedKeyword' -count=1 -> PASS, 0.174s
go test ./internal/service ./internal/handler ./internal/handler/admin -count=1 -> PASS; service 65.303s, handler 27.390s, admin 0.216s
go test ./cmd/server -run '^$' -count=1 -> PASS
node node_modules/vitest/vitest.mjs run src/features/prompt-audit/__tests__/integrationSurface.spec.ts src/views/admin/__tests__/RiskControlView.spec.ts -> PASS, 2 files / 7 tests
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit -> PASS
node node_modules/vite/bin/vite.js build -> PASS, 1904 modules; existing browserslist, dynamic-import, and chunk-size warnings only
contract outputs PowerShell snippet (after acceptance) -> PASS, 20 files / 2996311A4EC1458EEC9C2AE4327D5D5EAA695C878783DE984AF841BBF0A79145
```

## Findings
- 未发现明确问题。

## Baseline Failures
- Contract-known tagged API snapshot drift and the full repository billing SQL-mock `updatedAccountRows` 32/34-column fixture drift remain outside this integration. The contract does not require those broader suites; all required focused repository tests and required three-package full backend command passed.

## Bug Owner Recommendation
`none`

## Root Cause
- `none`

## Retest Scope
- none

## Unverified Risks
- No live provider, SMTP, real Redis/PostgreSQL, container, deployment, staging, push, or browser-runtime session was run, as required by the contract.
- Vite emitted pre-existing browserslist freshness, dynamic-import chunking, and chunk-size warnings; build completed successfully.

## Knowledge Promotion
`none`

### PASS: upstream-v0173-selective-fixes-s207

# QA Report

## Task ID
`upstream-v0173-selective-fixes-s207`

## Verdict
`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0173-selective-fixes-s207.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go test ./internal/service -run '<S207 focused regex>' -count=1 -> PASS
go test ./internal/service -run 'TestAntigravityGatewayService_ForwardGemini|TestGeminiMessagesCompatService|TestWebSearch' -count=1 -> PASS
go test ./internal/service -run 'TestGemini(ChatCompletions|MessagesCompat|Native)_(TempUnscheduled429|PoolMode429)DoesNotSetAccountRateLimit' -count=1 -> PASS
go test ./internal/service -count=1 -> PASS (61.428s latest run)
go test ./cmd/server -run '^$' -count=0 -> PASS
corepack.cmd pnpm --dir frontend exec vitest run src/components/common/__tests__/BaseDialog.spec.ts -> PASS (1 file / 1 test)
corepack.cmd pnpm --dir frontend run typecheck -> PASS
corepack.cmd pnpm --dir frontend run build -> PASS (1871 modules transformed)
gofmt -w <changed Go files> -> PASS
git diff --check -> PASS
allowlist audit -> PASS
git diff --name-only --diff-filter=U -> PASS (empty)
conflict marker scan -> PASS
v0.1.173 ancestor checks for all five source commits -> PASS
db0bff82c is not an ancestor of S207 HEAD -> PASS
post-merge focused S207 service smoke on local main -> PASS (5.472s)
post-merge go test ./cmd/server -run '^$' -count=0 on local main -> PASS (10.469s)
```

- manual checks:

```text
Compared every approved upstream patch with the local adaptation -> behavior preserved within local topology
Reviewed Gemini policy propagation through Messages, Native, and Chat Completions -> no duplicate temp-unscheduled or account-rate-limit write
Reviewed Antigravity fallback request bodies -> first uses mapped model, second uses fallback, result reports fallback
Reviewed changed paths against contract -> all paths allowed; no schema, migration, repository, handler, dependency, or unrelated frontend change
Removed the temporary worktree node_modules junction and verified the source node_modules target still exists
```

## Findings

- 未发现明确问题。
- The first allowlist command produced a false failure because its PowerShell expression created nested arrays; the corrected flattened-array check passed and showed all changed paths inside the contract allowlist.
- Frontend build warnings are pre-existing repository warnings and do not affect this scoped PASS.
- Local `main` fast-forward and post-merge smoke passed; remote push and deployment were not performed.

## Bug Owner Recommendation
`integration-owner`

## Root Cause
`none`

## Retest Scope

- Provider, deployment, container, shared data services, remote push, and production validation remain separate authorization scopes.

## Knowledge Promotion
`none`

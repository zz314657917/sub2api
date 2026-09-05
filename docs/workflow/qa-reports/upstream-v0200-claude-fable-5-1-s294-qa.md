### BLOCKED: upstream-v0200-claude-fable-5-1-s294

# QA Report

## Task ID
upstream-v0200-claude-fable-5-1-s294

## Verdict
`BLOCKED`

## Contract Checked

- `docs/workflow/tasks/upstream-v0200-claude-fable-5-1-s294.md`
- `docs/workflow/contract-reviews/upstream-v0200-claude-fable-5-1-s294-review.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes` for S294 changes
- denied paths touched: `no` by S294; unrelated pre-existing dirty paths remain outside scope
- commands run:

```text
go test ./internal/pkg/claude ./internal/pkg/antigravity ./internal/service -run 'Test(Default|.*Fable|.*Anthropic.*Window|.*ModelRateLimit|.*ClaudeOAuth|.*Usage)' -count=1 -> PASS
go test ./internal/service -> PASS
go build ./... -> PASS (rechecked after the independent Prompt Audit remediation)
npm.cmd run test -- --run src/composables/__tests__/useModelWhitelist.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts -> PASS, 47/47
npm.cmd run typecheck -> PASS (rechecked after the independent Prompt Audit remediation)
npm.cmd run build -> PASS (rechecked after the independent Prompt Audit remediation)
go test -tags unit ./... -> BLOCKED by pre-existing compile/test fixture drift
git diff --check -> PASS
git diff --name-only --diff-filter=U -> PASS, empty
node C:/Users/Administrator/.codex/scripts/codex-workflow.mjs pge-doctor --repo . --strict -> 20 checks pass; 1 non-blocking status compaction warning
```

- manual checks:

```text
Fable 5.1 appears in Claude/Antigravity/Bedrock catalogs and OpenCode config -> covered by focused backend/frontend tests
7d_oi-only 429 -> model-level Fable scope; no account cooldown -> covered by focused service tests
Antigravity Claude usage with only claude-fable-5-1 quota -> aggregate bar renders -> covered by new frontend test
```

## Findings

- 未发现 Fable-specific implementation bug in the focused backend/frontend evidence。
- 后续重测已解除并发 Prompt Audit 改动造成的 backend compilation 与 frontend typecheck/build 阻断；全量 unit-tag tests 仍保留已知 repository fixture drift，真实 provider/runtime smoke 也未执行。

## Bug Owner Recommendation

`integration-owner`

## Root Cause

`environment-blocked`

## Retest Scope

- 单独处理 `go test -tags unit ./...` 的 repository fixture 列数与既有编译漂移。
- 若要发布，再补真实 Anthropic/Antigravity/Bedrock provider/runtime smoke。
- 保留并重跑 S294 定向 backend/frontend 测试、diff allowlist 和冲突索引检查。

## Knowledge Promotion

- `none`

## Recommendation

`BLOCKED`。Fable 5.1 的局部行为证据和 build/typecheck 已通过，但全量 unit fixture drift 与 provider/runtime smoke 仍未解除；不标记 S294 为 PASS，也不执行 push、部署、容器或 provider smoke。

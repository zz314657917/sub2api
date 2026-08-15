### PASS: upstream-v0177-remote-compaction-v2-s218

# QA Report

## Task ID

`upstream-v0177-remote-compaction-v2-s218`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0177-remote-compaction-v2-s218.md`
- Developer commits reviewed: `f07518322`, `1567b88c8`.
- Review base: `098b4bd82..HEAD`; the Amendment duplicate is not evaluated as a
  product implementation commit.

## Evidence

- Diff reviewed: yes. Native `stream:true` + `compaction_trigger` remains on
  bare `/responses`; explicit compact and non-streaming body-signal requests
  retain the legacy path. HTTP, passthrough, and WebSocket paths admit and
  apply `x-codex-beta-features`; native v2 always includes
  `remote_compaction_v2` without overwriting a non-empty OAuth client value.
- R1 reviewed: yes. `supportsOpenAIEndpointCapabilityForRequest` rejects
  API-key accounts explicitly marked `openai_responses_supported=false` or
  force-chat for Responses selection. `Forward` suppresses raw-chat fallback
  for a context-marked native-v2 request, preserving `/responses` and its
  `compaction_trigger` body.
- Probe reviewed: yes. The local probe uses streaming `/responses`, SSE
  `Accept`, beta feature, OAuth `store:false`, deterministic UUID-shaped
  session IDs, and accepts success only when SSE/final/JSON output contains a
  compaction item. A 2xx empty output records unsupported.
- Allowed paths checked: yes. Exactly 19 changed files relative to
  `098b4bd82..HEAD`, all within the contract allowlist; outside allowlist: 0.
- Denied paths touched: no. No frontend, `outputs/`, migration, dependency,
  deployment, container, provider, database/Redis, turn-state, fingerprint,
  or push operation was performed.
- Commands run:

```text
go test ./internal/handler -list 'Test.*' and ./internal/service -list 'Test.*'
  -> PASS; all 4 handler and 12 service acceptance/R1 tests discoverable by default tag
go test ./internal/handler -run <4 S218 handler tests> -count=10
  -> PASS (1.512s)
go test ./internal/service -run <12 S218 and R1 service tests> -count=10
  -> PASS (1.724s)
go test ./internal/service -run <legacy compact compatibility pattern> -count=1
  -> PASS (1.816s)
go test ./internal/service -count=1
  -> PASS (62.912s)
go test ./internal/handler -count=1
  -> PASS (59.715s)
go test ./internal/server -count=1
  -> PASS (0.543s)
go test ./cmd/server -run '^$' -count=0
  -> PASS (1.069s; compile gate, no tests selected)
git diff --check 098b4bd82..HEAD; git diff --name-only --diff-filter=U; git ls-files -u
  -> PASS; no whitespace errors or unmerged index entries
gofmt -d <all changed Go files>
  -> PASS
git merge-base --is-ancestor 9662cff2e/a8b9ea22b/8ae6d8f67 upstream/main
  -> PASS
```

## Findings

- 未发现明确问题。

## Changed Files / Allowlist

- 19 actual files: 18 allowlisted backend source/test files plus
  `docs/workflow/worker-results/upstream-v0177-remote-compaction-v2-s218-result.md`.
- Exact changed-file and formatting checks show `outside=0`.

## Risks And Boundaries

- 验收全部使用 Go in-process fakes、`httptest` 或 loopback fixtures；未发起
  外部 provider 或生产请求。
- 已知基线：unit-tag service 包存在无关编译缺陷，因此未将 unit-tag 测试当作
  本次通过证据。R1 的等价真实请求回归位于 default-tag
  `openai_gateway_service_test.go`，已实际执行并通过。
- 本次为代码与本地运行时验收，不包含真实上游协议联调。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- 无；集成至主线后应按相同默认标签 focused、完整 backend 包和 Git 完整性门禁
  复跑。

## Knowledge Promotion

`none`

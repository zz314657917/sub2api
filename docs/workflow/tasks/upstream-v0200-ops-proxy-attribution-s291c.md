---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v0200-ops-proxy-attribution-s291c
worker_model: gpt-5.6-terra
base_commit: fd203d8bd
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-04
---

# S291-C OpenAI and WS Proxy Attribution

## Goal

补齐本地 OpenAI 全链及 WebSocket/fallback 错误事件的事件时代理快照，覆盖 Chat Completions、Responses、Messages、Images、Embeddings、Grok、raw stream、compact fallback 和 first-output timeout。WS 在无托管代理时必须记录 `unknown`。

## Success Criteria

- 所有本地 OpenAI 生产错误事件构造点要么显式使用 `opsUpstreamProxyID/Name`，要么由 transport/fallback helper 统一写入。
- WebSocket 默认 HTTP client 路径不推断直连，使用 `opsUpstreamWSProxyAttribution` 的 unknown 语义。
- 不改变响应、重试、failover、计费、流协议或调度行为。
- OpenAI 定向测试、完整 service 测试、server 编译、`go build ./...`、diff/冲突检查通过。

## Allowed Paths

- `backend/internal/service/openai_*.go`
- `backend/internal/service/gateway_*.go`（仅与 OpenAI 共享 helper 的错误路径）
- `backend/internal/service/grok_*.go`
- `backend/internal/service/ops_upstream_context_test.go`
- `docs/workflow/**` evidence files for this task
- `knowledge/tasks/current-task.md`

## Denied Paths

- `backend/internal/pkg/apicompat/**`, `admin_service.go`, frontend, lockfiles,
  outputs, schema, migrations, repositories, billing, deployment and containers
- Any path not listed above; no merge/rebase/cherry-pick

## Acceptance Commands

```powershell
Set-Location backend
gofmt -w internal/service/openai_*.go internal/service/grok_*.go
go test ./internal/service -run 'Test(OpenAI|OpsUpstream|GatewayUpstream|Grok)' -count=1
go test ./internal/service
go build ./...
git diff --check
git diff --name-only --diff-filter=U
```

## Stop Rules

遇到需要修改 apicompat、schema、依赖或改变 WS 传输行为时停止并回报；不宣称上游完整功能已完成，除非所有本地生产事件点扫描无遗漏。

## Output

写入 `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291c-result.md`，首行为 `### DONE/### BLOCKED/### FAILED`。

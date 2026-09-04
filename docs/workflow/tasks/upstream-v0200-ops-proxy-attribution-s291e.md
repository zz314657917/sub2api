---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v0200-ops-proxy-attribution-s291e
worker_model: gpt-5.6-terra
base_commit: 9a03e6735
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-04
---

# S291-E Attribution Completeness Fix

## Task ID

`upstream-v0200-ops-proxy-attribution-s291e`

## Role

Developer Worker (`gpt-5.6-terra`)

## Goal

仅修复全局生产事件扫描发现的 3 个漏填代理快照事件，随后证明本地全部直接 Ops 事件构造点没有缺失归因。

## Success Criteria

- 全局生产构造点扫描的三个 Gateway/Gemini 事件均使用当前账号的 proxy ID/name 快照。
- 扫描结果中每个生产 `OpsUpstreamErrorEvent` 构造点都显式写入代理归因，或在 `appendOpsUpstreamError` 中受到已验证的规范化保护。
- 不改变请求传输、重试、故障转移、计费或任何发送行为。

## Allowed Paths

- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `docs/workflow/**` evidence files for this task
- `knowledge/tasks/current-task.md`

## Denied Paths

- 其余所有业务文件、apicompat、admin service、frontend、outputs、schema、迁移、依赖、部署和容器；禁止 merge/rebase/cherry-pick。

## Constraints

- 仅适配行为，禁止整体合并、rebase 或 cherry-pick 上游历史。
- 不操作真实 provider、数据库、容器、部署或共享数据。
- 必须保留工作树中未提交的 apicompat、admin service、frontend、repository 与 `outputs/**` 改动。
- WebSocket 无可靠代理快照时只能记录 `unknown`，不得推断为直连。

## Acceptance Commands

```powershell
Set-Location backend
go test ./internal/service -run 'Test(Gateway|Gemini|OpsUpstream)' -count=1
go test ./internal/service
go build ./...
git diff --check
```

## Output

- `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291e-result.md`
- 完整性扫描结果、定向测试、service 测试、构建、diff 和冲突检查证据。

## Stop Rules

- 若事件没有可证明的账号或上下文快照，或需改变传输行为，停止并报告。
- 若需触及 denied path，或在保护脏改中发现没有本轮所有权的变化，停止并报告。

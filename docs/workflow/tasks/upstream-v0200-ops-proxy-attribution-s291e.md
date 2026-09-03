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

## Goal

仅修复全局生产事件扫描发现的 3 个漏填代理快照事件，随后证明本地全部直接 Ops 事件构造点没有缺失归因。

## Allowed Paths

- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `docs/workflow/**` evidence files for this task
- `knowledge/tasks/current-task.md`

## Denied Paths

- 其余所有业务文件、apicompat、admin service、frontend、outputs、schema、迁移、依赖、部署和容器；禁止 merge/rebase/cherry-pick。

## Acceptance Commands

```powershell
Set-Location backend
go test ./internal/service -run 'Test(Gateway|Gemini|OpsUpstream)' -count=1
go test ./internal/service
go build ./...
git diff --check
```

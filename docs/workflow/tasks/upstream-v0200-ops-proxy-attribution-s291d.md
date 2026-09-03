---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v0200-ops-proxy-attribution-s291d
worker_model: gpt-5.6-terra
base_commit: 4a692587a
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-04
---

# S291-D Antigravity Proxy Attribution

## Goal

在本地单体 `antigravity_gateway_service.go` 的全部生产上游错误事件中写入事件时代理快照，完成上游代理归因功能在本地现存 Gateway owners 的覆盖。

## Success Criteria

- 每个 `appendOpsUpstreamError` 事件使用所属账户或 pipeline 账户的 proxy ID/name 快照。
- 不改变 Antigravity 重试、rate-limit、failover、计费或请求传输。
- 服务测试、构建、diff/冲突检查通过，且所有本地生产 Ops 事件构造点均有代理归因或由明确 helper 覆盖。

## Allowed Paths

- `backend/internal/service/antigravity_gateway_service.go`
- `backend/internal/service/antigravity_gateway_service_test.go`
- `docs/workflow/**` evidence files for this task
- `knowledge/tasks/current-task.md`

## Denied Paths

- apicompat、admin service、frontend、lockfiles、outputs、schema、migrations、repositories、billing、deployment、containers，以及未列路径；禁止 merge/rebase/cherry-pick。

## Acceptance Commands

```powershell
Set-Location backend
gofmt -w internal/service/antigravity_gateway_service.go
go test ./internal/service -run 'Test(Antigravity|OpsUpstream)' -count=1
go test ./internal/service
go build ./...
git diff --check
git diff --name-only --diff-filter=U
```

## Stop Rules

若事件没有可证明的账户快照、需变更传输行为或触及任何 denied path，停止并报告。

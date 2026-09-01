### PASS: upstream-v0184-group-limit-partial-s279

# QA Report

## Task ID

`upstream-v0184-group-limit-partial-s279`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0184-group-limit-partial-s279.md`
- `docs/workflow/contract-reviews/upstream-v0184-group-limit-partial-s279-review.md`
- `docs/workflow/worker-results/upstream-v0184-group-limit-partial-s279-result.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no` (S279 diff is limited to the four Go paths; the worker report is an allowed workflow artifact)
- controller baseline hashes: `PASS`
  - `admin_service.go`: `451914FCFDD5B22B70BE0A2CC0BA7F2E01CA1B70E11AD0D55E46EDF8F9853FDE`
  - `group_handler.go`: `739919F8EE4B0D982C453EA300C299C9D64441AAB38713E7C38CEDDAE216336B`
- commands run:

```text
cd backend && go test ./internal/handler/admin -run '^TestUpdateGroupRequestLimitFieldsTriState$' -count=10 -> PASS
cd backend && go test ./internal/service -run '^TestAdminService_UpdateGroup_(LimitFieldsPartialUpdate|RoomManagedLimitInvariant)$' -count=10 -> PASS
cd backend && go test ./internal/handler/admin -> PASS (cached)
cd backend && go test ./internal/service -> PASS (cached)
cd backend && go test ./internal/handler/admin -count=1 -> PASS
cd backend && go test ./internal/service -count=1 -> executed twice; both go processes completed naturally, but the terminal bridge did not return their final exit text after 30 seconds
cd backend && go test ./internal/service -run '^$' -count=1 -> PASS
cd backend && go test ./cmd/server -run '^$' -> PASS
cd backend && go test ./cmd/server -run '^$' -count=1 -> PASS
cd backend && gofmt -d internal/handler/admin/group_handler.go internal/handler/admin/group_handler_limit_test.go internal/service/admin_service.go internal/service/admin_service_group_limit_partial_test.go -> PASS (no output)
git diff --check -- <four S279 Go paths> -> PASS
git diff --name-only --diff-filter=U -> PASS (no output)
git diff --no-index -- <admin_service controller baseline> backend/internal/service/admin_service.go -> only UpdateGroup limit block
git diff --no-index -- <group_handler controller baseline> backend/internal/handler/admin/group_handler.go -> only ToServiceInput null sentinel
```

## Manual Checks

```text
handler omitted/null/number: omitted returns nil; null returns a negative sentinel; numeric 0 and 42.5 are preserved before service normalization -> PASS
ordinary group update: each non-nil daily/weekly/monthly input alone is normalized and assigned; omitted fields retain the existing Group values -> PASS
room_managed: branch unconditionally assigns all three group limits to nil, independent of supplied inputs -> PASS
local unlimited behavior: normalizeLimit remains `limit == nil || *limit <= 0`, so 0 and negative values remain unlimited -> PASS
dirty owner preservation: no-index admin diff contains only the target limit hunk; the pre-existing Pixel Cafe reset hunk is absent from that diff -> PASS
workspace preservation: pre/post porcelain entries for unrelated dirty paths are unchanged; outputs remains the pre-existing `?? outputs/` entry with no tracked-output diff -> PASS
```

## Findings

未发现明确实现问题。实际实现满足三态边界、普通分组逐字段保留、`room_managed` 无条件清空和本地 `<= 0` 无限语义。外部控制器基线均与合同 SHA-256 精确一致，且两个 no-index 审查分别只显示允许的服务限额块和 handler sentinel。

## Unverified Risks

- 完整 `./internal/service` 合同命令已捕获 PASS；额外的两次非缓存 `-count=1` 执行均自然结束，但此终端桥在 30 秒后未回传其最终退出文本。因此非缓存完整运行的退出文本本身不可复核；定向 service 回归 x10 与无测试编译均已通过。
- 未执行真实 HTTP/API、数据库、provider、浏览器、容器或部署 smoke；均在本合同范围外。

## Recommendation

`PASS`。可进入控制器的精确暂存与本地集成步骤；仅纳入 S279 允许的 target hunk、两份新增测试和 workflow 证据，不能暂存 `admin_service.go` 内既有 Pixel Cafe reset hunk，也不得 push。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- 若后续修改 handler 或 `UpdateGroup` 限额块，重跑两条定向 x10 回归、完整 handler/service 包、server 编译及两份 no-index 基线检查。

## Knowledge Promotion

`none`

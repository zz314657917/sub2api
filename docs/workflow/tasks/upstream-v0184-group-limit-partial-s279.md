---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0184-group-limit-partial-s279
worker_model: gpt-5.6-terra
base_commit: 408916129
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.184 Group Limit Partial Updates S279

## Task ID

`upstream-v0184-group-limit-partial-s279`

## Role

你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑适配上游 `9f1effd71`：管理员部分更新分组时，省略的日/周/月限额保持原值；显式 `null` 只清除对应限额；数字只更新对应限额。保留本地 `room_managed` 分组始终使用无限分组级额度的约束。禁止整体 merge/cherry-pick。

## Success Criteria

- Handler 三态明确：JSON 省略字段传 `nil`（不修改），显式 `null` 传负数（清除为无限），数字原值传递；`0` 在 handler 层不被改写。
- 普通分组非限额部分更新保留已有日/周/月限额；只更新/清除某一项时其余两项保持不变。
- `room_managed` 分组无论输入是否省略或携带限额，服务端仍强制把三项分组限额清空，避免与 Room 计划下发的 Key 限额叠加。
- 保留本地 `normalizeLimit` 的既有 `<=0` 为无限语义；不顺带引入上游“0 表示禁止用量”的更大语义变化。
- 定向 handler/service 回归、完整受影响包、`cmd/server` 编译、gofmt、diff/conflict 和脏文件基线检查通过。

## Context

- Repo: `F:\mcplugins\sub2api`
- Upstream provenance: `9f1effd71`，包含在 `v0.1.184`。
- Upstream owner `backend/internal/service/admin_group.go` 在本地已并入 `backend/internal/service/admin_service.go`；原 patch 不能按文件拓扑直接套用。
- 当前 `admin_service.go` 已有用户/并行 Pixel Cafe 配额重置脏改，绝不能覆盖或纳入 S279。
- Controller baseline snapshots:
  - `E:/codex-runtime/pge/sub2api/upstream-v0184-group-limit-partial-s279/controller-baseline/admin_service.go` SHA-256 `451914FCFDD5B22B70BE0A2CC0BA7F2E01CA1B70E11AD0D55E46EDF8F9853FDE`
  - `E:/codex-runtime/pge/sub2api/upstream-v0184-group-limit-partial-s279/controller-baseline/group_handler.go` SHA-256 `739919F8EE4B0D982C453EA300C299C9D64441AAB38713E7C38CEDDAE216336B`

## Allowed Paths

- `backend/internal/handler/admin/group_handler.go`（仅 `optionalLimitField.ToServiceInput`）
- `backend/internal/handler/admin/group_handler_limit_test.go`
- `backend/internal/service/admin_service.go`（仅 `UpdateGroup` 的日/周/月限额更新块）
- `backend/internal/service/admin_service_group_limit_partial_test.go`
- `docs/workflow/worker-results/upstream-v0184-group-limit-partial-s279-result.md`
- `docs/workflow/qa-reports/upstream-v0184-group-limit-partial-s279-qa.md`

## Denied Paths

- `backend/internal/service/admin_service.go` 中除 `UpdateGroup` 限额块外的所有行，尤其 imports 与 `AdminResetCafeRateLimitUsage`
- `backend/internal/service/admin_service_group_test.go`
- `backend/internal/repository/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/service/api_key_auth_cache*.go`
- `backend/internal/server/middleware/**`
- `frontend/**`
- `frontend/pnpm-lock.yaml`
- `VERSION`
- `docker-compose*.yml`
- `deploy/**`
- `knowledge/**`
- `outputs/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、配置、容器、部署或数据文件

## Constraints

- 普通分组逐字段应用限额更新；不得把三个字段重新作为全量替换。
- `room_managed` 分支必须继续一次性强制清空三项限额，即使三项输入均为 `nil`。
- Handler 的 `null` 使用负数 sentinel；不要改变 JSON 数字、字符串数字和非法输入的既有解析规则。
- 不修改 `normalizeLimit` 的本地 `<=0` 语义，不改 DTO、repository、schema、migration 或前端。
- `admin_service.go` 是受保护脏文件；修改前后必须与 controller baseline 做 no-index 对比，除目标限额块外不得出现 S279 新差异。禁止格式化整个文件。
- 不 commit、push、调用 provider、操作数据库、容器、部署或共享数据。

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/handler/admin -run '^TestUpdateGroupRequestLimitFieldsTriState$' -count=10
go test ./internal/service -run '^TestAdminService_UpdateGroup_(LimitFieldsPartialUpdate|RoomManagedLimitInvariant)$' -count=10
go test ./internal/handler/admin
go test ./internal/service
go test ./cmd/server -run '^$'
gofmt -w internal/handler/admin/group_handler.go internal/handler/admin/group_handler_limit_test.go internal/service/admin_service_group_limit_partial_test.go
```

From repo root:

```powershell
git diff --check -- backend/internal/handler/admin/group_handler.go backend/internal/handler/admin/group_handler_limit_test.go backend/internal/service/admin_service.go backend/internal/service/admin_service_group_limit_partial_test.go
git diff --name-only --diff-filter=U
git diff --no-index -- E:/codex-runtime/pge/sub2api/upstream-v0184-group-limit-partial-s279/controller-baseline/admin_service.go backend/internal/service/admin_service.go
```

Also inspect: S279 changes in `admin_service.go` are exactly one target limit block; the pre-existing Cafe reset diff is byte-for-byte preserved; handler diff is limited to the null sentinel; tests cover omitted/null/number, per-field preservation and room-managed clearing; protected dirty paths/outputs are unchanged.

## Output

- Write `docs/workflow/worker-results/upstream-v0184-group-limit-partial-s279-result.md` using `C:/Users/Administrator/.codex/templates/worker-result.md`.
- First line must be `### DONE: upstream-v0184-group-limit-partial-s279`, `### BLOCKED: upstream-v0184-group-limit-partial-s279` or `### FAILED: upstream-v0184-group-limit-partial-s279`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0184-group-limit-partial-s279-qa.md`.

## Stop Rules

- If the behavior requires repository/schema/migration/frontend changes or any non-target `admin_service.go` edit, stop with `BLOCKED`.
- If the baseline snapshot is missing or shows unexplained differences outside the target limit block, stop and report; do not overwrite the dirty file.
- Record failures truthfully; do not weaken assertions or absorb concurrent changes.

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

# Task Contract

## Task ID
upstream-low-risk-maintenance-s250

## Role
Codex Controller 在隔离 worktree 中实施并复核。不得扩展到 Pixel Cafe、Google One 管理接口或其他上游候选。

## Goal
选择性移植三项已确认、互不重叠的上游低风险修复：DOMPurify 安全升级、cgroup/宿主内存指标一致性，以及管理员编辑用户时的并发数 `0 = 不限` 入口。每项保持独立业务提交。

## Success Criteria
- `DOMPurify` 直接与传递依赖都被锁定在 `>= 3.4.14`，不改变其他前端依赖版本。
- 未设置 cgroup 内存上限时，Ops 指标使用完整的宿主机 used/total/percent 三元组；设置具体 cgroup 上限时继续使用完整 cgroup 三元组。
- 用户编辑弹窗允许保存整数 `0`，仍拒绝负数和非整数，并清楚提示 `0 = unlimited`。
- 三项分别以依赖安全、Ops 指标、用户编辑 UI 三个业务提交交付；不合并分叉上游历史。

## Context
- Repo: `F:/mcplugins/sub2api`
- Base: `main@7e326ac28`
- Upstream sources: `4a1da2950`, `cd05772e9`, `5dfad32b8` from `upstream/main@03e8ab413`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`

## Allowed Paths
- `frontend/package.json`
- `frontend/pnpm-lock.yaml`
- `backend/internal/service/ops_metrics_collector.go`
- `backend/internal/service/ops_metrics_collector_memory_test.go`
- `frontend/src/components/admin/user/UserEditModal.vue`
- `frontend/src/components/admin/user/__tests__/UserEditModal.spec.ts`
- `frontend/src/i18n/locales/en/admin/users.ts`
- `frontend/src/i18n/locales/zh/admin/users.ts`
- `docs/workflow/worker-results/upstream-low-risk-maintenance-s250-result.md`
- `docs/workflow/qa-reports/upstream-low-risk-maintenance-s250-qa.md`

## Denied Paths
- `frontend/src/features/pixelCafe/**`
- `backend/internal/service/admin_service*.go`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/i18n/locales/{en,zh}/admin/groups.ts`
- `knowledge/**`
- `outputs/**`
- 数据库 schema/migration、provider 调用、容器、部署、push，以及未在 Allowed Paths 中的任何文件。

## Constraints
- 每个候选仅移植本地可达的修复行为；不得直接 cherry-pick 或合并分叉历史。
- 依赖升级只允许使用已有 lockfile 与包管理器的冻结/锁定模式，不运行会改写额外 workspace 元数据的安装命令。
- 保持本地语言、组件、测试和 Ops collector 拓扑；不得做无关重构或全量格式化。
- 主工作区的所有既有改动和未跟踪文件必须保持原样；业务实现仅在任务 worktree 发生。
- 禁止 push、部署、容器操作、真实 provider 或共享/生产数据库访问。

## Acceptance Commands
```powershell
# backend worktree
go test ./internal/service -run "TestResolveMemoryStats" -count=10
go test ./internal/service -run "TestOpsMetricsCollector" -count=1
go test ./cmd/server -run '^$' -count=1

# frontend worktree; use existing locked dependencies only
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/admin/user/__tests__/UserEditModal.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"

# scope and hygiene
gofmt -w backend/internal/service/ops_metrics_collector.go backend/internal/service/ops_metrics_collector_memory_test.go
git diff --check
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" .
git diff --name-only <base>..HEAD
git diff --cached --name-only
git diff --name-only
git ls-files -u
```

## Output
- 三个独立业务提交：依赖安全、Ops 内存指标、用户并发 UI。
- Controller result 第一行必须是 `### DONE: upstream-low-risk-maintenance-s250`，列出 changed files、命令、结果、风险和上游来源。
- QA report 第一行必须是 `### PASS: upstream-low-risk-maintenance-s250`、`### FAIL: ...` 或 `### BLOCKED: ...`，并明确是否独立执行。

## Stop Rules
- 任一修复需要修改 Denied Paths、数据库/生产配置、额外依赖升级或真实外部状态时立即停止并回报。
- 若 lockfile 不能在允许依赖范围内再生，停止依赖切片，不以手工不一致的 lockfile 代替。
- 若 focused 或 build 验收失败，停止主线合入，先定位到相应独立切片。
- 若出现未合并 index、冲突标记，或主工作区保护路径改变，停止并报告。

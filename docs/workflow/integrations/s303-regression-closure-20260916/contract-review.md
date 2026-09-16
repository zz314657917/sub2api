### PASS: s303-regression-closure

---
type: contract-review
scope: repository
status: approved
task_id: s303-regression-closure
verdict: PASS
base_commit: 75c9ab270e266c11e5548be17abad0a891a90a15
reviewer: final-evaluator
last_verified: 2026-09-16
---

### PASS: s303-regression-closure

# Contract Review

## Task ID
s303-regression-closure

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/s303-regression-closure.md`

## Findings
- 后端目标是实际 `jwtAuthWithSessionBinding` 与 `adminAuthWithSessionBinding`，不是重写 mock：两条路径均经 `UserService.GetByID` 的包装错误并使用 `errors.Is` 将直接或包装的 `ErrUserNotFound` 归为 `401/USER_NOT_FOUND`，其他查询错误归为 `500/INTERNAL_ERROR`；Admin 的 HTTP `Authorization` 与 WebSocket `Sec-WebSocket-Protocol: jwt.<token>` 分支均进入同一 JWT 校验函数。现有 unit fixture 已有可复用的 `stubUserRepo`、JWT 生成器和 `httptest` 路由，不需要数据库或真实升级。
- 前端目标是实际 `apiClient` 拦截器和 `refreshAuthTokens`。当前实现先判定 refresh token/auth user 是否已变化，再把 Axios 网络错误（状态 0）、429、5xx 映射为 `TOKEN_REFRESH_UNAVAILABLE`；401/403 及非 Axios 的 malformed-success 错误走清理与 `TOKEN_REFRESH_FAILED`。因此合同要求的 changed-session 优先级、保留/清理分支和 Axios adapter fake 均可直接断言。
- 后端选择命令已实际执行并成功：`go test -tags unit ./internal/server/middleware -list 'TestS303'`（退出码 0；当前尚未添加 `TestS303`，符合 build 前状态）。
- 前端锁文件完整并固定了 Vitest 2.1.9、jsdom、Axios 等 fixture 依赖，但本 worktree 暂无 `frontend/node_modules`，直接执行现有 `npm.cmd run test:run -- --list ...` 报 `vitest is not recognized`。这不是业务或合同范围问题；Generator 在任何前端命令前必须按合同约束执行 `pnpm.cmd install --frozen-lockfile`（或使用等价的已安装本地运行时），且不得改动 lockfile/manifest。把这一步记入 worker/QA 实际命令记录即可；无需改写合同。

## Gate Checks
- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`（后端已实测；前端在锁定依赖安装后的既定本地脚本中执行）
- worker_model_confirmed: `yes`（`gpt-5.6-terra`）
- base_commit_confirmed: `yes`（当前 HEAD 为 `75c9ab270e266c11e5548be17abad0a891a90a15`）
- openspec_traceable: `yes`（`docs/workflow/plans/upstream-main-integration-roadmap-20260913.md`）

## Approval
- 合同仅允许两个新增/扩展回归测试与流程证据，生产鉴权、会话、依赖、部署及数据边界均被明确拒绝；可进入 Generator build。
- 最小执行约束：新增后端用例必须真实调用 middleware/HTTP handler，并同时断言受保护 handler 未到达；新增前端每例须在 `afterEach` 复原 Axios mock、localStorage/sessionStorage、location 和模块/单飞状态，不能以降低既有断言替代覆盖。

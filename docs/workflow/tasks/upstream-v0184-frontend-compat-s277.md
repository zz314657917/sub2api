---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0184-frontend-compat-s277
worker_model: gpt-5.6-terra
base_commit: 53484808e7b1cab0049c2066d1a53816848e8b3c
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-08-31
---

# Upstream v0.1.184 Frontend Compatibility S277

## Task ID

`upstream-v0184-frontend-compat-s277`

## Role

你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按当前本地拓扑手工适配三个上游前端兼容行为：严格解析 `datetime-local` 过期时间、兑换码批量更新使用该解析器、以及生成的 Claude Code 配置不再关闭 attribution header。禁止整段合并或 cherry-pick 上游历史。

## Success Criteria

- `parseDateTimeLocalInput` 只接受 `YYYY-MM-DDTHH:mm` 和可选秒/小数秒的无时区本地 datetime；拒绝空字符串、空格分隔、`Z`/offset、超出日历、小时/分钟/秒范围的值。有效值以本地 `Date` 组件转换为秒级 Unix timestamp，秒以下截断，现有 `formatDateTimeLocalInput` 的分钟级 round trip 不变。
- 兑换码批量更新 custom expiry 调用该解析器；无效输入不得调用批量 API，保留现有 `expiryDaysRequired` 错误；有效输入以解析结果生成 ISO 时间（秒以下按 parser 规则截断）。clear mode 与其他批量字段不变。
- `UseKeyModal` 生成的 Anthropic/Claude Code Unix、CMD、PowerShell、Grok Claude 和 settings JSON 均保留 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`，但不出现 `CLAUDE_CODE_ATTRIBUTION_HEADER` 覆盖。
- 为上述 parser、兑换码提交和各终端/JSON 配置补充定向 Vitest 覆盖；typecheck 与 production build 不新增错误。

## Context

- Repo: `F:\mcplugins\sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`
- Upstream provenance: `81e461f65`, `b7aca87fd`, `5778739cd`, `c03776604`.
- Local owners: `frontend/src/utils/format.ts`, `frontend/src/views/admin/RedeemView.vue`, `frontend/src/components/keys/UseKeyModal.vue`.
- `git apply --check` against current topology fails for both source bundles. Adapt behavior, do not import upstream path/layout wholesale.
- Existing uncommitted S277 product edits may be present when work begins; inspect and preserve them, then add only missing test coverage and bounded corrections.

## Allowed Paths

- `frontend/src/utils/format.ts`
- `frontend/src/utils/__tests__/formatDateTimeLocalInput.spec.ts`
- `frontend/src/views/admin/RedeemView.vue`
- `frontend/src/views/admin/__tests__/RedeemView.batchUpdate.spec.ts`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`
- `docs/workflow/worker-results/upstream-v0184-frontend-compat-s277-result.md`
- `docs/workflow/qa-reports/upstream-v0184-frontend-compat-s277-qa.md`

## Denied Paths

- `backend/**`
- `frontend/pnpm-lock.yaml`
- `frontend/src/views/admin/pixelCafe/**`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_route_breaker_test.go`
- `backend/internal/service/admin_service.go`
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/package.json`
- `VERSION`
- `docker-compose*.yml`
- `deploy/**`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、生产配置、容器、部署或数据文件

## Constraints

- 仅实现四个来源所表达的三项用户可见行为；不增加时区提示或 i18n 键，不改 API 契约，不刷新 lockfile/依赖。
- 不得把任意可被 `Date.parse` 接受的字符串当作有效 `datetime-local` 输入；仅组件级校验后创建本地 `Date`。
- Claude attribution 修复只删除明确的禁用 override，不得关闭 nonessential traffic 保护或修改其他客户端配置。
- 不回滚或覆盖既有改动；不得执行 commit、push、provider 请求、数据库、容器、部署或共享数据操作。

## Acceptance Commands

从 `F:\mcplugins\sub2api\frontend` 执行：

```powershell
pnpm.cmd exec vitest run src/utils/__tests__/formatDateTimeLocalInput.spec.ts src/views/admin/__tests__/RedeemView.batchUpdate.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts
pnpm.cmd run typecheck
pnpm.cmd run build
```

从仓库根目录执行：

```powershell
git diff --check -- frontend/src/utils/format.ts frontend/src/utils/__tests__/formatDateTimeLocalInput.spec.ts frontend/src/views/admin/RedeemView.vue frontend/src/views/admin/__tests__/RedeemView.batchUpdate.spec.ts frontend/src/components/keys/UseKeyModal.vue frontend/src/components/keys/__tests__/UseKeyModal.spec.ts
git diff --name-only --diff-filter=U
```

还必须人工核对：changed paths 精确属于 allowlist；`CLAUDE_CODE_ATTRIBUTION_HEADER` 不再出现在上述 modal 的任何生成路径或测试 fixture；`frontend/pnpm-lock.yaml`、所有 `backend/**` 脏改、Pixel Cafe 和 `outputs/**` 的状态/哈希未变化；没有外部状态操作。

## Output

- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 `docs/workflow/worker-results/upstream-v0184-frontend-compat-s277-result.md`。
- 报告第一行必须是 `### DONE: upstream-v0184-frontend-compat-s277`、`### BLOCKED: upstream-v0184-frontend-compat-s277` 或 `### FAILED: upstream-v0184-frontend-compat-s277`。
- 报告列出 changed files、commands run、关键测试摘要、四个上游来源的处理状态、风险和 `knowledge_candidates`。

## Stop Rules

- 如果任一行为需要修改 denied path、API、依赖、i18n、数据库、生产配置或外部状态，立即 `BLOCKED` 并交回 Codex 裁决。
- 如果现有脏改动与 allowlist 内的 source/test owner 无法安全共存，停止，不得覆盖或重置该文件。
- 任何测试失败记录真实命令与归因；不得删除测试、放宽断言或吸收用户已有改动。
- 不提交产品 commit，不 push，不更新容器或共享数据。

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

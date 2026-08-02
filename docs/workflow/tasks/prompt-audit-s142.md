# Task Contract

## Task ID

prompt-audit-s142

## Role

Generator worker：只按本合同实现 Prompt Audit 行为切片；Planner/Evaluator 由主 Codex 负责。不得把其它 v0.1.169 功能带入本 Sprint。

## Goal

将上游 OpenAI-compatible Prompt Audit/Qwen3Guard 行为适配到当前已发布本地基线，提供可配置的 off、async_audit、blocking 模式、管理员配置与事件工作台，并保留完整中文/英文界面。所有安全边界必须在当前本地架构和迁移序列中成立，而不是直接 cherry-pick 上游整提交。

## Success Criteria

- 支持管理员 Prompt Audit 配置、端点探测、事件列表/详情/删除与按筛选删除；路由只允许现有管理员权限，操作动作沿用本地审计动作键。
- OpenAI-compatible Chat/Responses、Gemini、Embeddings、Images、Live、Videos 以及异步图片任务等现有入口按合同接入扫描；off 模式保持原业务路径，async 模式不阻塞请求，blocking 模式在无法取得可信配置或命中阻断策略时 fail-closed。
- Prompt snapshot、scanner、worker、队列和存储不把原始提示词、Guard token、Cookie、Bearer/API key 或完整请求体写入 PostgreSQL、审计日志或普通错误响应；`202` 的 `full_prompt` 列仅为详情契约兼容列，运行时只能写入经过限制的脱敏文本。
- 端点 URL、DNS、重定向、代理和探测凭证遵循 SSRF/凭证隔离规则；同一端点 URL 之外不得复用已保存 token，localhost 只允许明确的 loopback 解析。
- 新迁移使用 `201_prompt_audit.sql` 与 `202_prompt_audit_full_prompt.sql`，顺序位于现有 `200_add_ops_error_logs_user_time_index_notx.sql` 之后，不改写既有迁移。
- 中文和英文 Prompt Audit locale 键集合一致；未知后端枚举有安全回退，页面不会出现未汉化的固定操作文本。
- 通过聚焦后端/前端回归、Go 编译、前端 typecheck/build、迁移顺序与内容审计、秘密/原始提示词静态扫描、冲突和路径门禁；已知基线失败必须单独记录且不得由本 Sprint 引入。

## Context

- Repo: `E:/codex-worktrees/sub2api/prompt-audit-s142`
- Base: `origin/main@d25800a97`（S141 已发布基线）
- Upstream references: `d11bdb13f`（初始 Prompt Audit）、`0f7f8a317`、`df9d9e2e4`、`56b5f0df6`、`fc495e087`、`ac685ccaf` 等后续安全/审阅修复；仅作行为参考。
- Candidate reference: `codex/v0169-behavior-wide@f674f8684` 的 Prompt Audit 91-file slice；不得合并该分支的 Passkey、部署、计费、邮箱、代理熔断或其它 release-wide 改动。
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`, this contract, `backend/internal/server/middleware/audit_log.go`, local migration README/runner, and current audit locale modules.

## Allowed Paths

- `backend/cmd/server/main.go`（仅 Prompt Audit 生命周期）
- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`
- `backend/internal/handler/gateway_handler*.go`
- `backend/internal/handler/gemini_v1beta_handler.go`
- `backend/internal/handler/image_task_handler.go`
- `backend/internal/handler/openai_*.go`
- `backend/internal/handler/security_audit_*.go`
- `backend/internal/handler/handler.go`
- `backend/internal/handler/wire.go`
- `backend/internal/securityaudit/**`
- `backend/internal/server/routes/admin.go`
- `backend/migrations/201_prompt_audit.sql`
- `backend/migrations/202_prompt_audit_full_prompt.sql`
- `frontend/src/features/prompt-audit/**`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/router/index.ts`
- `frontend/src/i18n/locales/en/admin/index.ts`
- `frontend/src/i18n/locales/en/admin/promptAudit.ts`
- `frontend/src/i18n/locales/en/nav.ts`
- `frontend/src/i18n/locales/zh/admin/index.ts`
- `frontend/src/i18n/locales/zh/admin/promptAudit.ts`
- `frontend/src/i18n/locales/zh/nav.ts`
- `frontend/package.json`
- `frontend/pnpm-lock.yaml`
- `docs/workflow/tasks/prompt-audit-s142.md`
- `docs/workflow/worker-results/prompt-audit-s142-result.md`
- `docs/workflow/qa-reports/prompt-audit-s142-qa.md`
- `docs/workflow/main-log.md`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`

## Denied Paths

- `F:/mcplugins/sub2api/**` primary worktree user changes
- Passkey/WebAuthn files, migrations, API, profile/login UI, and auth settings
- `backend/go.mod`, `backend/go.sum`, deploy/Docker/Compose/GoReleaser/VERSION files unless a dependency is proven strictly required and separately approved
- Billing, pricing, scheduler, account refresh, SMTP, proxy circuit, Model Plaza, unrelated frontend pages, and unrelated locale modules
- `knowledge/**`, `outputs/**`, `C:/Users/Administrator/.codex/memories/**`
- Production databases, deployment, containers, remote refs, commit, push, force-push, and release publication

## Constraints

- Preserve dirty primary worktree and existing local architecture; do not reset, checkout, stash, or overwrite unrelated changes.
- Port behavior, not upstream history or file layout. Resolve local route, handler, repository and wire seams explicitly.
- Keep migrations 201/202 append-only and verify the actual migration runner ordering.
- Default mode must remain off and non-blocking; only blocking mode may fail closed, and it must fail closed when persisted configuration cannot be trusted.
- Never store raw prompt text or Guard credentials in PostgreSQL/audit logs; tests must use canary values and assert absence.
- Do not claim runtime/provider/browser/deployment verification without executing it. No external Guard or production call is authorized in this Sprint.

## Acceptance Commands

```powershell
Set-Location backend
go test ./internal/securityaudit -count=1
go test ./internal/handler -run 'PromptAudit|SecurityAudit' -count=1
go test ./internal/server/routes -run 'PromptAudit' -count=1
go test ./internal/server/middleware -run 'PromptAudit|Audit' -count=1
go test ./... -run '^$'
go build ./...

Set-Location ../frontend
corepack.cmd pnpm exec vitest run src/features/prompt-audit src/components/layout/__tests__/AppSidebar.spec.ts
corepack.cmd pnpm run typecheck
corepack.cmd pnpm run build

Set-Location ..
rg -n "Bearer |sk-|api[_-]?key|raw_prompt|full_prompt" backend/internal/securityaudit backend/internal/handler/security_audit_*.go
Get-ChildItem backend/migrations -File | Sort-Object Name
git diff --check
git ls-files -u
rg -n '^(<<<<<<<|=======|>>>>>>>)' backend frontend docs/workflow
```

Also perform a manual diff/allowlist audit and record any pre-existing failures separately.

## Output

- Separate implementation commits by coherent Prompt Audit batch if needed; do not merge or push.
- Worker result: `docs/workflow/worker-results/prompt-audit-s142-result.md`, first line `### DONE: prompt-audit-s142`, `### BLOCKED: ...`, or `### FAILED: ...`.
- QA report: `docs/workflow/qa-reports/prompt-audit-s142-qa.md`, first line `### PASS/FAIL/BLOCKED: prompt-audit-s142`.
- Report changed files, exact commands, evidence, residual risks, and whether any knowledge candidates exist.

## Stop Rules

- Stop and return `BLOCKED` if the implementation needs Passkey/auth, a migration rewrite, a dependency upgrade, production/deployment changes, or a security semantic choice not specified here.
- Stop if any raw prompt or credential can reach persistence and the contract cannot resolve it without expanding scope.
- Stop on denied-path changes, unresolved wire/schema conflicts, or two consecutive failed verification rounds; return the evidence instead of broadening the merge.

## Budget

- worker_mode: `claude-bare-deepseek-v4-pro`
- qa_worker_mode: `claude-bare-deepseek-v4-pro`
- worktree_root: `E:/codex-worktrees`

### DONE: prompt-audit-s142

# Worker Result

## Task ID
prompt-audit-s142

## Status
`done`

## Summary
- 完成 Prompt Audit/Qwen3Guard 的本地行为适配：管理员配置、端点探测、运行态、事件列表/详情/删除、按筛选删除，以及 OpenAI Chat/Responses、Anthropic Messages、Gemini、Embeddings、Images、Live、Videos 和异步图片任务入口的审计门控。
- 支持 `off`、`async_audit`、`blocking` 和 `BlockingLatestTurnOnly`；blocking 在配置不可信任、Guard 不可用或响应非法时保持 fail-closed，off/async 不改变原业务放行路径。
- 增加 DNS/私网/metadata/重定向/代理隔离，Guard token 不跨端点复用；Redis 只保存短期扫描载荷，PostgreSQL、审计日志、错误响应和管理员详情只保留脱敏/有界数据。
- 完成中英文 Prompt Audit 页面、导航和 locale；详情查询不再读取兼容列 `202.full_prompt`，避免旧行原文进入服务内存。

## Changed Files
- `backend/internal/securityaudit/**`
- `backend/internal/handler/security_audit_*.go`、各现有 gateway handler 的审计入口、`backend/internal/server/routes/admin.go`
- `backend/cmd/server/{main.go,wire.go,wire_gen.go,wire_gen_test.go}`
- `backend/migrations/201_prompt_audit.sql`、`backend/migrations/202_prompt_audit_full_prompt.sql`
- `frontend/src/features/prompt-audit/**`、Prompt Audit locale、`AppSidebar.vue`、`router/index.ts`、前端依赖锁文件
- `docs/workflow/spec.md`

## Commands Run
```text
go test ./internal/securityaudit -count=1 -> PASS
go test ./internal/handler -run 'PromptAudit|SecurityAudit' -count=1 -> PASS
go test ./internal/server/routes -run 'PromptAudit' -count=1 -> PASS
go test ./internal/server/middleware -run 'PromptAudit|Audit' -count=1 -> PASS
go test ./... -run '^$' -> PASS
go build ./... -> PASS
corepack.cmd pnpm --dir frontend exec vitest run src/features/prompt-audit src/components/layout/__tests__/AppSidebar.spec.ts -> 6 files / 45 tests PASS
corepack.cmd pnpm --dir frontend run typecheck -> PASS
corepack.cmd pnpm --dir frontend run build -> 1119 modules PASS
gofmt on changed Go files -> PASS
git diff --check HEAD -> PASS
git ls-files -u -> empty
scoped secret/raw-prompt scan -> only expected test canaries, compatibility names, token-header construction and API-key identifiers
exact changed-diff conflict scan -> no conflict markers introduced
```

## Test Output
```text
internal/securityaudit: ok
internal/handler PromptAudit/SecurityAudit: ok
internal/server/routes PromptAudit: ok
frontend: 6 passed, 45 passed
frontend build: vite transformed 1119 modules; built in 20.24s
```

## Risks
- 未运行真实 PostgreSQL 迁移/事务集成、Redis ACL/TLS、真实 Qwen3Guard、浏览器认证态、部署或生产验证；这些不在本 Sprint 隔离 worktree 验收边界内。
- 隔离分支基于 `origin/main@d25800a97`，文件列表显示 `199 -> 201 -> 202`；发布时必须合入本地 `main` 已有的 `200_add_ops_error_logs_user_time_index_notx.sql`，再复核最终编号顺序。
- 全量前端 Vitest 仍有仓库既有 public pages、Chat/Image Studio、Canvas、ImageCreator、Monitor、Settings 等基线失败；S142 聚焦测试未引入这些文件改动。

## Knowledge Candidates
- none

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- none

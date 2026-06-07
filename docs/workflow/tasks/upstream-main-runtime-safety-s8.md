# 上游合成 Sprint：upstream-main-runtime-safety-s8

## Summary

- Task ID: upstream-main-runtime-safety-s8
- Role: Developer Worker
- Branch: codex/upstream-main-runtime-safety-s8
- Worktree: E:/codex-worktrees/sub2api/upstream-main-runtime-safety-s8
- Baseline: local main=fed704641, upstream/main=635ad81cd
- Goal: 合成下一批后端运行时安全和一致性修复，不直接 merge upstream/main。

## Scope

本轮只移植低耦合后端 hardening 补丁：

- DB pool connection lifetime / idle time clamp。
- content moderation auto-ban 跳过管理员账号。
- OpenAI/Gateway compatible handlers 校验 stream 字段类型。
- 账号状态清理后同步 scheduler snapshot。
- OpenAI HTTP response id 绑定到选中账号，增强 previous response 续链路由一致性。

不纳入 frontend、Ent/migration、leader lock、大网关重构、TTFT migration、Linux DO 前端登录修复、lint-only 扫尾。

## Candidate Commits

按顺序 cherry-pick -x 或手工等价移植：

1. 1b6a15b48 fix(db-pool): enforce connection lifetime floors
2. c40a74d98 fix(risk-control): exempt admins from moderation auto-ban
3. 3571b082f fix: validate stream field type
4. b6c0706e3 fix: sync scheduler snapshots after account state clears
5. 7513b7ea6 Bind OpenAI HTTP response IDs to selected accounts

已判定本轮跳过：

- 0a521f09f: 本地已等价存在 Gemini tool_use -> text 前关闭 content block 的逻辑。
- 362f9e77b: leader lock 跨 repo/service/server/wire，单独 Sprint。
- 69b465451: 需要 migration，禁止本轮。
- aea2950b1: 涉及前端，禁止本轮。
- 650981f2e: lint/gofmt-only 且依赖上游上下文，低优先。

## Allowed Paths

- backend/internal/repository/db_pool.go
- backend/internal/repository/db_pool_test.go
- backend/internal/repository/account_repo.go
- backend/internal/repository/account_repo_integration_test.go
- backend/internal/service/content_moderation.go
- backend/internal/service/content_moderation_test.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_gateway_service_test.go
- backend/internal/handler/gateway_handler_chat_completions.go
- backend/internal/handler/gateway_handler_responses.go
- backend/internal/handler/openai_chat_completions.go
- backend/internal/handler/openai_gateway_handler.go
- backend/internal/handler/openai_stream_validation.go
- backend/internal/handler/openai_stream_validation_test.go
- docs/workflow/tasks/upstream-main-runtime-safety-s8.md
- docs/workflow/worker-results/upstream-main-runtime-safety-s8-result.md
- docs/workflow/qa-reports/upstream-main-runtime-safety-s8-qa.md
- docs/workflow/main-log.md

## Denied Paths

- frontend/**
- backend/ent/**
- backend/migrations/**
- deploy/**
- knowledge/**
- .github/**
- assets/**
- README*
- docs/workflow/status.md
- docs/workflow/spec.md

## Constraints

- 不新增数据库字段、migration、Ent schema、前端 API 或配置项。
- 不直接 merge upstream/main。
- 优先保留本地更完整实现；若上游提交已被等价吸收，记录 APPLIED_EQUIVALENT，不重复覆盖。
- 若某候选需要 denied paths、migration、frontend、leader-lock 架构入口或新公开 contract 字段，标记 DEFERRED 并停止该候选，不扩大 Sprint。
- 当前主工作区有用户未提交 knowledge 改动；实施必须停留在隔离 worktree。

## Public APIs / Interfaces

- 不新增公开 DTO 字段或配置项。
- 行为变化限定为：
  - DB pool 对非正数或超过 24 小时的 lifetime/idle time 使用安全默认值。
  - 风控 auto-ban 不禁用管理员账号。
  - OpenAI/Gateway Chat Completions 与 Responses 入口拒绝非 boolean stream 字段。
  - 账号过载、临时不可调度和模型限流清理后，scheduler snapshot 及时同步。
  - OpenAI HTTP response id 绑定当前账号，previous_response_id 粘连路由更稳定。

## Acceptance Commands

在 backend/ 目录执行：

```powershell
go test ./internal/repository -run "DBPool|Pool|Connection|Lifetime|SetOverloaded|TempUnschedulable|ClearModelRateLimits|Scheduler" -count=1
go test ./internal/service -run "ContentModeration|AutoBan|Admin|OpenAI|ResponseID|BindHTTP" -count=1
go test ./internal/handler -run "Stream|OpenAI|Gateway|ChatCompletions|Responses" -count=1
go test ./internal/repository ./internal/service ./internal/handler -count=1
```

基础检查：

```powershell
git status --short --branch
git diff --check
```

路径审计必须确认无 denied paths 改动。

## Output

- docs/workflow/worker-results/upstream-main-runtime-safety-s8-result.md
- docs/workflow/qa-reports/upstream-main-runtime-safety-s8-qa.md
- docs/workflow/main-log.md 追加一行或多行 S8 记录

## Stop Rules

- 任何候选要求修改 denied paths，停止该候选并记录 DEFERRED。
- cherry-pick 冲突若涉及本地产品线覆盖或大范围重构，停止该候选并记录 DEFERRED。
- 测试失败必须先归因；不能在未解释失败原因时合回 main。

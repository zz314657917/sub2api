# API Key Adaptive Route Breaker S271

## Task ID

api-key-adaptive-route-breaker-s271

## Role

你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

在现有 API Key 多分组路由短冷却之上，增加多实例共享的自适应分组熔断：按“分组 + 路由类型 + 精确请求模型”隔离连续瞬时故障，达到阈值后跳过故障分组，并通过单探测半开状态安全恢复。保留置顶账号无 fallback、现有账号级瞬时隔离、订阅/计费/权限和流式已输出后不可重放等边界。

## Success Criteria

- 现有 `API Key + group` 首次瞬时故障短冷却保持兼容；`429`、`529`、`5xx` 和流式 HTTP 200 内已分类的同类上游错误仍能让该 Key 的后续请求跳过所选分组。
- 新增共享熔断只在模型感知的多分组路由上生效，键空间按 `group_id + routing_scope + normalized exact requested model` 隔离；不得把某模型或图片/视频/Embedding 故障扩大为整个分组的无差别熔断。
- 共享状态以 Redis 原子操作维护，并使用 Redis TIME 避免多实例时钟偏差。关闭状态下连续 3 次可熔断故障才打开；任一成功响应清零连续失败。
- 未达到阈值的连续失败状态在 30 分钟无活动后自动遗忘，避免相隔很久的零散故障累计开闸；冷却/半开状态保留到足以完成最长退避和租约回收。
- 打开后的冷却阶梯为 30 秒、2 分钟、10 分钟、30 分钟封顶。冷却到期进入半开，只允许一个探测请求；探测失败升级阶梯，成功完全关闭，业务 `4xx` 不计失败并释放探测资格。
- 熔断判断或 Redis 读写失败时 fail open，不得造成所有路由不可用；现有单 Key 本地/Redis 短冷却继续独立工作。
- 所有可用候选都在冷却/熔断时不得偷偷回落到已被 skipper 排除的默认分组；返回无匹配路由，等待下一次半开探测。
- `PinnedAccountID > 0` 继续绕过多分组路由与熔断，不得自动调用其他分组。
- 添加默认标签单元测试覆盖：范围归一化、三次阈值、成功重置、四档退避、同冷却窗口并发失败不重复升级、半开单探测、业务 4xx 释放、Redis 错误 fail-open、默认分组不绕过、流式错误状态保真。

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first (canonical primary checkout, read-only): `F:/mcplugins/sub2api/AGENTS.md`, `F:/mcplugins/sub2api/docs/workflow/status.md`, `F:/mcplugins/sub2api/docs/workflow/spec.md`, `F:/mcplugins/sub2api/docs/workflow/agent-matrix.md`
- The task worktree may be based on committed `HEAD` and therefore have stale or absent ignored workflow guidance. The absolute primary paths above are authoritative for phase/contract approval only; never modify them from the Worker.
- Existing route selection: `backend/internal/service/api_key_routing.go`
- Existing per-Key cooldown: `backend/internal/service/api_key_service.go`, `backend/internal/repository/api_key_cache.go`, `backend/internal/server/middleware/api_key_auth.go`
- Request model context: `backend/internal/pkg/ctxkey/ctxkey.go`, `backend/internal/handler/ops_error_logger.go`
- Existing account+model breaker remains authoritative for individual account faults and is not modified.
- Primary `main` now contains S272 source commit `8b84ccf34`, which removed the
  route-count authorization bypass from `ResolveAPIKeyForModelRequest`. The
  S271 worktree predates that commit and modifies the same middleware owner;
  Controller integration must preserve S272's nil-only early return and its
  single-group model-match regressions.

## Allowed Paths

- `backend/internal/pkg/ctxkey/ctxkey.go`
- `backend/internal/service/api_key.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_routing.go`
- `backend/internal/service/api_key_route_breaker*.go`
- `backend/internal/service/api_key_route_cooldown_test.go`
- `backend/internal/service/api_key_routing*_test.go`
- `backend/internal/repository/api_key_cache.go`
- `backend/internal/repository/api_key_route_breaker*.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/middleware.go`
- `backend/internal/server/middleware/api_key_auth*_test.go`
- `docs/workflow/worker-results/api-key-adaptive-route-breaker-s271-result.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, database schemas and migrations
- Account schedulers, account-model transient breaker, billing, quota, subscription, pricing and usage-record behavior outside the existing route resolution handoff
- Gateway handler request replay or same-request cross-group failover
- Admin/public APIs, frontend and route-breaker UI
- Dependencies, deployment, containers, shared/local database data, provider traffic, configuration defaults and secrets
- `knowledge/**`, `outputs/**`, `C:/Users/Administrator/.codex/memories/**`
- Git commit, staging, push, history rewriting or cleanup of existing worktrees/branches
- Any source or workflow path not listed in Allowed Paths

## Constraints

- 保持最小改动，不做无关重构或格式化；不得覆盖当前本地领先提交或未跟踪 `outputs/`。
- 新共享熔断通过可选的专用 cache capability 接入，避免强制扩张 `APIKeyCache` 后导致大量既有 test double 失效。
- Redis key 的模型部分必须有界且不可注入任意超长原文；使用稳定归一化/摘要。Lua 只能操作同一 key，保持 Redis Cluster 单 key 安全。
- 失败分类不得把普通 `400/401/403/404`、内容审核、用户余额/套餐/权限错误计入共享故障；共享熔断仅统计明确标记的上游瞬时错误，或最终 `502/503/504/529`。无法确认来源的普通 `429/500` 仍可维持现有单 Key 短冷却，但不得升级共享分组熔断。
- 半开探测租约必须有界并可在业务 4xx 后释放；选择器若为多个候选申请探测，必须释放未最终选中的租约。
- 本 Sprint 只保证“后续请求”跳过；同一请求跨分组重放需要统一 body replay、流式输出、计费与订阅重验，明确留待后续 contract。
- Controller review/integration must hand-port the S271 middleware changes onto
  current `main` without restoring the removed
  `len(apiKey.MultiGroupRoutes) == 0 && apiKey.PinnedAccountID <= 0` fast path.
  `TestS272` is a mandatory regression gate before independent S271 QA.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
gofmt -w internal/pkg/ctxkey/ctxkey.go internal/service/api_key.go internal/service/api_key_service.go internal/service/api_key_routing.go internal/service/api_key_route_breaker*.go internal/repository/api_key_cache.go internal/repository/api_key_route_breaker*.go internal/server/middleware/api_key_auth.go internal/server/middleware/middleware.go internal/server/middleware/api_key_auth*_test.go
go test ./internal/repository -run "APIKeyRouteBreaker" -count=10
go test ./internal/service -run "APIKey(RouteBreaker|RouteCooldown|Routing)" -count=10
go test ./internal/server/middleware -run "APIKeyRoute(Cooldown|Breaker)|ShouldCooldownAPIKeyRoute" -count=10
go test ./internal/server/middleware -run '^TestS272' -count=10
go test ./internal/repository ./internal/service ./internal/server/middleware -count=1
go test ./cmd/server -run '^$' -count=1
cd F:/mcplugins/sub2api
git diff --check
git diff --name-only
git diff --cached --name-only
git ls-files -u
git status --short
```

- Controller 已在不含 S271 的主工作区复现既有
  `TestUpdateWithAccountBillingSettingsRollsBackWhenOutboxFails` fixture 漂移：
  `expected 32, actual 34`。完整 repository 命令仍必须执行并记录；如果
  唯一失败与该基线一致，可作为范围外已知失败归因，不得修改 denied-path
  fixture。S271 自有 repository focused 测试、完整 service、完整 middleware
  和 server compile 仍必须真实通过。
- 如果完整 `internal/service` 被已知的 `-tags unit` 漂移阻断，不得添加该 tag；默认标签完整包必须实际通过，否则报告 FAIL/BLOCKED。
- 不运行真实 Redis/provider；repository Lua 通过 `miniredis` 或仓库既有无外部依赖测试验证原子状态机。

## Output

- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 `docs/workflow/worker-results/api-key-adaptive-route-breaker-s271-result.md`。
- Worker report 第一行必须是 `### DONE: api-key-adaptive-route-breaker-s271`、`### BLOCKED: api-key-adaptive-route-breaker-s271` 或 `### FAILED: api-key-adaptive-route-breaker-s271`。
- 必须列出 changed files、commands run、test output、risks、contract compliance 和 knowledge_candidates。
- 不粘贴无关长日志；不直接写长期知识库。

## Stop Rules

- Contract 字段、故障分类、scope key 或半开并发语义仍不明确时停止并报告。
- 需要修改 Denied Paths、数据库/生产配置、安全边界、计费或同请求重放时停止并请求 Codex 裁决。
- 不能证明半开单探测、退避不被并发尾请求重复升级、默认组不绕过或 Redis 失败 fail-open 时不得宣称完成。
- `gpt-5.6-terra` 不可用、worktree 已存在、E 盘可用空间低于 10GB 或现有工作区边界不安全时立即 `BLOCKED`，不得静默换模型。

## Budget

- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.20`
- worktree_root: `E:/codex-worktrees`

## Worker Output

- 内容同 `Output`。

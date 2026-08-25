# Task Contract

## Task ID

`upstream-openai-spark-shadow-auto-reset-s259-b`

## Role

你是 P/G/E 流程里的 `gpt-5.6-terra` Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

将上游 `bdf7ead15` 中本地尚缺的 Spark linked-shadow 调度、独立配额边界与最小后台生命周期 API 适配到当前拓扑。S259-A 已提供 `parent_account_id`、`quota_dimension` 和父凭据传播；本任务必须建立其余闭环，而不是重做或机械覆盖 S259-A 及本地已有 quota helper。

## Success Criteria

- 调度快照保留 `parent_account_id` 和 `quota_dimension`。影子候选只有在其父账号仍为 OpenAI OAuth 且凭据可用时才可调度；父账号缺失、非 OpenAI OAuth、失效、过期或临时不可调度时 fail-closed。父账号手动 `schedulable=false` 与 global `RateLimitResetAt`/overload 不得误伤独立 Spark 配额维度。
- Spark 影子只按自身 `model_mapping`（默认只含当前 Spark alias 的恒等映射）参与相应模型调度；普通账号已配置相同模型时的既有行为保持不变。不得为 Spark 新增硬编码的全局账号类型门。
- 影子查询 `/wham/usage` 必须使用 S259-A 的父 OAuth 凭据，但只从 `additional_rate_limits[].metered_feature == "codex_bengalfox"` 生成并持久化影子行自己的 canonical `codex_5h_*` / `codex_7d_*` 快照；不得写入母账号或使用 `codex_spark_*` 前缀。影子的快照 TTL 不受 WSv2 门控；普通 OpenAI 探测语义保持原样。
- Spark 影子的 `/responses` 429 不得写 global quota、持久化 global Codex header、进入全局 OAuth-429 storm 或运行时熔断；普通 OpenAI OAuth 的原 429 路径保持。影子发生 401 时，token cache、刷新/错误和临时不可调度状态必须归属父凭据账号；刷新器不得直接刷新影子。
- 增加受管理员鉴权保护的 `POST /api/v1/admin/accounts/:id/shadow`。仅可为正常 OpenAI OAuth 父账号创建一个 Spark 影子；影子使用父代理、继承未提供的优先级/并发、默认 Spark 映射，且没有认证凭据。重复或并发创建返回结构化冲突；分组绑定失败不得留下半成品影子。
- 管理更新、批量更新、删除、刷新和重置保持影子不变量：影子不能写认证凭据、改变 OAuth 类型或单独代理；父代理更新同步给影子；父账号删除先处理影子；影子刷新/通用 quota reset/OpenAI reset-credit 均在出站前拒绝并保留结构化 HTTP 状态。账号导出跳过影子并返回跳过数量，避免产出无法导入的空凭据记录。
- 管理账号 DTO/创建响应/列表可表达 `parent_account_id` 与 `quota_dimension`，但本切片不得额外暴露父账号邮箱、ChatGPT account ID、订阅或隐私信息。
- 每项上述行为都要有 default-tag 的 focused 回归。当前主线中已有的 Spark quota-reset guard 和父凭据 resolver 必须由这些回归验证，而不是删除或重复实现。

## Context

- Repo: `F:/mcplugins/sub2api`
- Implementation worktree must start from the approved contract commit on `main@83402d5e9`.
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`, this contract, then the current local owners and upstream `bdf7ead15` counterparts.
- Upstream source: `bdf7ead15 feat(spark-shadow): OpenAI Spark linked shadow account`; it is an ancestor of refreshed `upstream/main@aa2c4e8d1`. `git apply --check` is expected to fail because Ent/API/Gateway topology diverged.
- S259-A is integrated on `main` as `48a0344e1`, Controller evidence `d8eb21b76`, and independent QA evidence `83402d5e9`. Treat its schema, migrations, credential resolver and HTTP/WS propagation as required baseline; do not modify them.
- Local partial baseline already contains `OpenAIQuotaService` shadow reset/credential helpers. Reuse and test them; do not overwrite later Agent Identity or quota-reset recovery behavior.
- Frozen primary protection: `main@83402d5e9`, 122 porcelain entries, patch-id `4a908d78f03c7be845284f83e3db10659d5373d4`, staged=0, unmerged=0. This is the recorded pre-build snapshot for later integration review, not a lock on unrelated concurrent user editing. The Developer worktree must remain isolated.
- The user-owned dirty `backend/cmd/server/wire_gen.go` touch is only the Cafe handler construction near line 299, while S259-B's account-usage injection is near line 150. The dirty `admin_service.go` touch is only Group room-managed limit normalization near lines 2070/2405, while S259-B account lifecycle owners start near line 3537. These are function-disjoint evidence, not permission to overwrite: final integration must re-check the live primary patch before applying.

## Allowed Paths

- `backend/internal/service/account_service.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/scheduler_cache.go`
- `backend/internal/repository/account_repo_spark_shadow_test.go`
- `backend/internal/repository/scheduler_cache_unit_test.go`
- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/openai_quota_service.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/token_refresher.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/service/shadow_routing.go`
- `backend/internal/service/shadow_routing_test.go`
- `backend/internal/service/openai_spark_shadow_scheduler_test.go`
- `backend/internal/service/openai_spark_shadow_quota_test.go`
- `backend/internal/service/openai_spark_shadow_admin_test.go`
- `backend/internal/service/ratelimit_service_openai_test.go`
- `backend/internal/service/openai_quota_service_test.go`
- `backend/internal/handler/admin/openai_oauth_handler.go`
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/admin/account_data.go`
- `backend/internal/handler/admin/admin_service_stub_test.go`
- `backend/internal/handler/admin/openai_oauth_handler_spark_shadow_test.go`
- `backend/internal/handler/admin/account_handler_spark_shadow_test.go`
- `backend/internal/handler/admin/account_data_handler_test.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/account_spark_shadow_test.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/api_contract_test.go`
- `docs/workflow/status.md` (Planner-only; Developer read-only)
- `docs/workflow/main-log.md` (Planner-only; Developer read-only)
- `docs/workflow/tasks/upstream-openai-spark-shadow-auto-reset-s259-b.md` (Planner-only; Developer read-only)
- `docs/workflow/worker-results/upstream-openai-spark-shadow-auto-reset-s259-b-result.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, database schema, generated Ent code, and all CRS sync owners.
- `frontend/**`, including account modal, account list UI, reset-card UI, i18n and browser tests.
- Automatic reset-credit consumption and warning behavior from `6f972145b` / `96b160d9`; those are S259-C.
- S258 OAuth identity revalidation, all gateway HTTP/WS credential propagation owners covered by S259-A, provider traffic, production/shared database, container, deployment, push, secrets, dependencies, lockfiles, CI, `knowledge/**`, `outputs/**`, and global memories.
- Any primary-worktree Pixel Cafe, Group, Settings, knowledge, asset or existing user-owned dirty path.

## Constraints

- Apply behavior, not the upstream patch. Preserve current Agent Identity, quota-reset recovery, model scheduler, privacy and response contracts unless this contract explicitly changes a Spark boundary.
- No migration execution, no real provider calls and no live database. Use fakes/httptest only.
- Shadow credentials must never be persisted, redacted, exported, refreshed or used to overwrite a parent. Parent resolution must stay one-hop and fail-closed as S259-A established.
- Do not add parent PII to a shadow response. Do not create a background job, automatic reset flow, retry loop or frontend action.
- Keep interface impact minimal. If the current `AccountRepository` / `AdminService` change requires a test stub or server contract stub, modify only the explicitly allowed support files; if another owner becomes necessary, stop for Codex review.
- A parent proxy update must not leave a silent partial shadow state. If error/transaction semantics cannot be kept safe in current topology, stop rather than inventing a broad repository redesign.
- During isolated build and independent QA, re-check the primary worktree before terminal reporting: staged and unmerged indexes must remain empty. A changed primary patch-id or porcelain list alone is non-blocking because it may be a concurrent user edit; record the observed state, then continue only if any S259-B-owned file overlap remains function/hunk-disjoint after live review. Do not stage, restore, clean, or amend anything in `F:/mcplugins/sub2api`.

## Acceptance Commands

Run from `backend/` unless a command says otherwise:

```powershell
go test ./internal/repository -run 'Test.*(SparkShadow|Shadow.*Scheduler)' -count=10
go test ./internal/service -run 'Test.*(SparkShadow|Spark.*Shadow|Shadow.*Spark|ParentHealth)' -count=10
go test ./internal/handler/admin -run 'Test.*(SparkShadow|Spark.*Shadow|CreateShadow|Shadow.*Refresh|Shadow.*Reset)' -count=10
go test ./internal/handler/dto -run 'Test.*(SparkShadow|Spark.*Shadow)' -count=10
go test ./internal/server -run 'Test.*(SparkShadow|Spark.*Shadow)' -count=1
go test ./internal/service ./internal/handler/admin ./internal/handler/dto ./internal/server -count=1 -timeout=3m
go test ./internal/repository -run 'Test.*(SparkShadow|Spark.*Shadow|Shadow.*Spark|ParentHealth)' -count=10
go test ./cmd/server -run '^$' -count=1
Set-Location ..
$changedGoFiles = git diff --name-only 83402d5e9..HEAD | Where-Object { $_ -like '*.go' }
if ($changedGoFiles) { & gofmt -d $changedGoFiles }
git diff --check
```

Also verify all of the following before reporting DONE:

```powershell
git diff --name-only 83402d5e9..HEAD -- <Allowed Paths except worker report>
git diff --cached --name-only
git ls-files -u
rg -n '^(<<<<<<<|=======|>>>>>>>)' <changed source/test paths>
git -C F:/mcplugins/sub2api status --short
git -C F:/mcplugins/sub2api diff --binary | git patch-id --stable
```

During isolated build and QA, the primary result must have staged=0 and unmerged=0. Record its current porcelain list and patch-id for evidence; a changed patch-id alone is not a failure. For every S259-B-owned path that is also dirty in the primary worktree, review the live zero-context hunk and surrounding function before continuing. The last command may emit known CRLF conversion warnings from pre-existing user files; they are not evidence of a changed patch-id.

## Output

- Write `docs/workflow/worker-results/upstream-openai-spark-shadow-auto-reset-s259-b-result.md`.
- Its first line must be `### DONE: upstream-openai-spark-shadow-auto-reset-s259-b`, `### BLOCKED: upstream-openai-spark-shadow-auto-reset-s259-b`, or `### FAILED: upstream-openai-spark-shadow-auto-reset-s259-b`.
- Include changed files, each command/result, focused-test names, exact behavior retained from the local partial baseline, residual risks, knowledge candidates, and contract-compliance fields.
- Commit business code/tests separately from the worker result. Do not write knowledge or memory files.

## Stop Rules

- Stop if a schema/migration, generated Ent file, frontend, CRS owner, gateway credential propagation owner, automatic reset path, production operation, or a path outside `Allowed Paths` is required.
- Stop if a parent PII response field, direct child credential persistence, automatic reset behavior, or a second-hop shadow chain is needed to satisfy a test.
- Stop if current interface impact requires modifying a support stub outside the allowlist, if proxy propagation cannot be made atomic enough for the existing repository contract, or if any focused test cannot be made default-tag.
- Stop if the primary staged or unmerged index is non-empty, if a live primary dirty hunk/function actually intersects an S259-B change, or if the same implementation issue fails twice. Primary patch-id/porcelain drift alone is not a stop condition during isolated build or QA. Report the exact blocker; do not widen the task.

## Contract Amendment: Primary Protection Phase Boundary

- This amendment applies only to the isolated Developer and independent-QA phases. It recognizes that the primary worktree is actively user-maintained: isolated work must never modify it, but a changed patch-id is not evidence that isolated source behavior is unsafe.
- Before any final main integration, the Controller must take a fresh primary snapshot (porcelain paths, patch-id, staged and unmerged indexes), run `git apply --check` for every candidate commit against the live primary worktree, and manually review every overlapping dirty hunk with its enclosing function. If an S259-B application would replace, reorder, or co-own a user hunk, stop integration instead of using a broad merge or overwrite.
- Final integration is legal only after that fresh check passes, the relevant staged/unmerged indexes remain empty, and Controller records the before/after protection evidence. This amendment does not authorize direct cherry-pick, reset, checkout, clean, push, or any change to user-owned primary files.

## Contract Amendment: Repository Baseline Failure

- `TestUpdateWithAccountBillingSettingsRollsBackWhenOutboxFails` is a default-tag baseline failure outside this task's allowed owners: at `main@83402d5e9` it panics in `backend/internal/repository/account_repo_upstream_billing_probe_update_test.go:559` because sqlmock supplies 32 values for the current 34-column generated account query. The identical command fails in the S259-B candidate before any S259-B test is selected.
- Do not modify that test, its fixtures, Ent, or schema. The Worker must record one full `go test ./internal/repository -count=1 -timeout=3m` execution as known-baseline-failed evidence, then run the amended focused repository command plus the full service/admin-handler/DTO/server command above. This known baseline failure neither proves nor excuses a Spark regression.
- The Worker must continue the remaining lifecycle, DTO, handler, route, export, update/bulk-update/delete/refresh/reset guard implementation and default-tag regressions. A terminal `DONE` still requires every success criterion and all amended target commands to pass; otherwise report `FAILED` with the precise unimplemented criteria.

## Contract Amendment: Server Evidence Owner

- The only originally allowlisted server test owner, `backend/internal/server/api_contract_test.go`, is repository-owned `//go:build unit`. Making it default-tag would activate an unrelated large contract suite and is not a narrow Spark test. No new server test owner is added.
- The default-tag server evidence for S259-B is therefore the focused handler route/API regressions plus `go test ./internal/server -count=1` and `go test ./cmd/server -run '^$' -count=1`, with Controller source review confirming `POST /api/v1/admin/accounts/:id/shadow` remains inside the protected administrator account group. This amendment does not permit `-tags=unit`, route-auth changes, or broader server test edits.

## Budget

- worker_mode: `native-codex-agent-gpt-5.6-terra`
- qa_worker_mode: `native-codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees/sub2api`

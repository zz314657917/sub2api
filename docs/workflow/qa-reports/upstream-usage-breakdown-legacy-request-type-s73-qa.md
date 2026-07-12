### PASS: upstream-usage-breakdown-legacy-request-type-s73

# QA Report

- Task ID: `upstream-usage-breakdown-legacy-request-type-s73`
- Contract: `docs/workflow/tasks/upstream-usage-breakdown-legacy-request-type-s73.md`
- Integration HEAD: `40b710e73`
- Evaluated commit: `40b710e738d9dfac57f1478d9311585069c1c63e`

## Findings

- 未发现明确问题。
- Sync fallback 精确限制为 `request_type=0, stream=false, openai_ws_mode=false`；Stream fallback 精确限制为 `request_type=0, stream=true, openai_ws_mode=false`；WS v2 fallback 精确限制为 `request_type=0, openai_ws_mode=true`。
- 三个 fallback 均以 `request_type=0` 为前提，显式非零 request type 继续由 `request_type = $N` 分支权威匹配，不会被 legacy flags 重分类。
- 原 `buildRequestTypeFilterCondition` 仅委托空 alias helper，已有无前缀 SQL 保持不变；`GetUserBreakdownStats` 只切换为 `ul` alias helper。
- `RequestType` 与 `Stream` 同时存在时，独立 `AND ul.stream = $N` 仍会追加。七列 row scan、`ORDER BY actual_cost DESC`、正数 `LIMIT` 以及 ordinary breakdown 包含 `exclude_from_leaderboard` 用户的语义均保持不变。
- S73 实际提交仅修改合同允许的 3 个路径；生产代码差异只位于 `GetUserBreakdownStats`、`buildRequestTypeFilterCondition` 及其 alias-aware helper，未触碰 leaderboard functions。

## Executed Checks

- 原样执行完整 S73 Acceptance Commands：`PASS`，总耗时约 45 秒。
- Required discovery gate 精确发现 2/2 个测试；required matrix 与 request-type regression matrix 均 `PASS`。
- Leaderboard discovery gate 与 `^TestUsageLogRepositoryGetUserLeaderboard` 回归均 `PASS`。
- `go test ./internal/repository -run "^$" -count=1` repository compile-only：`PASS`。
- 合同 clean worktree、allowed-path、production-hunk 与 `git diff --check` 门禁均 `PASS`。
- 通过 `git show` 独立审查实际 S73 commit `40b710e73`：3 个 changed paths 全在 allowlist，业务 diff 仅为 alias-aware request-type fallback；worker result 首行为 `### DONE: upstream-usage-breakdown-legacy-request-type-s73`。
- SQLMock matrix 实际返回并扫描非空七列 row，且断言完整 dimensions、`RequestType+Stream` extra AND、`ORDER BY actual_cost DESC`、`LIMIT` 和 ordinary breakdown SQL 不含 `exclude_from_leaderboard`。
- S71 backend smoke：service discovery `4/4`、middleware `1/1`、DTO `1/1`，对应 Fast/Flex HTTP/WS 与 user-ID round-trip 测试全部 `PASS`。
- S72 backend smoke：bare GPT-5.6 catalog/normalization/billing/account-scope discovery `6/6`，目标测试全部 `PASS`。

## Unverified Risks

- 未连接真实 PostgreSQL；本轮 SQL 行为由 SQLMock、精确 query fragments、参数矩阵和静态代码审查验证。
- 未运行全仓库测试、race test 或 S71/S72 frontend tests；本轮严格执行 S73 完整合同门禁，并仅补充合同要求的 S71/S72 轻量 backend smoke。

## Recommendation

- `PASS`。S73 可与 S71/S72 一并进入最终组合回归，无需返回 Generator 修复。
- Bug owner recommendation: `none`；root cause: `none`；knowledge promotion: `none`。

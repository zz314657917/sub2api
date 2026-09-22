### PASS: upstream-v027-next-backend

## Final Verdict

最终冻结版本通过独立 QA。DeepSeek fallback 只在 `PlatformDeepseek` 或解析后的精确
`api.deepseek.com` hostname 选择 opt-in bridge；plaintext reasoning 保留，encrypted-only
assistant 使用单空格占位，userinfo/suffix lookalike 不触发。active 但管理员暂停的
OAuth account 进入刷新候选；成功持久化 credentials 后保持 `Schedulable=false`，没有
`SetSchedulable(true)`。

初始 QA 曾发现 JSON string user item 未清空 pending reasoning：
`[reasoning(secret), "new user", assistant]` 会泄漏 `secret` 到后续 assistant。原
Developer 已在该 user 分支清空 pending reasoning，并在允许的
`deepseek_reasoning_v027_test.go` 新增
`TestV027DeepSeekReasoning_OptInStringInputResetsUserBoundary`。静态 diff 与下述两次
独立运行确认该修复准确且无范围扩大。

## Passed Commands

- `go test ./internal/pkg/apicompat -count=1`: PASS。
- `go test ./internal/service -run 'TestV027(DeepSeekReasoning|PausedRefresh)' -count=1 -v`: PASS；4 个顶层测试通过。fake upstream 覆盖 exact host、userinfo/suffix lookalike、DeepSeek platform、plaintext、encrypted-only streaming placeholder；fake refresh 覆盖 paused/enabled OAuth、non-OAuth exclusion、repository error 与零 `SetSchedulable`。
- `go build ./...`: PASS。
- 外部最小复现副本 `E:/codex-runtime/pge/sub2api/upstream-v027-next-qa-boundary/backend`：`go test ./internal/pkg/apicompat -run TestQAV027DeepSeekReasoningJSONStringUserBoundary -count=1 -v`: PASS。该测试先在初始版本实际失败（assistant `ReasoningContent="secret"`），复制最终 bridge 后通过。
- 隔离 tagged 精确集合：以全部 production Go files 加相关 unit fixtures/stubs（排除已知损坏的无关 tagged fixtures）运行 `go test -tags unit <production-files> <related-fixtures> -run 'TestForwardResponses_ForceChatCompletions|TestTokenRefreshService|TestV027(DeepSeekReasoning|PausedRefresh)' -count=1 -v`: PASS，15.526s。实际覆盖两个既有 raw fallback tests、完整 `TestTokenRefreshService*`、V027 DeepSeek/paused tests。
- 7 个允许 Go 文件 `gofmt -d`: 无输出；`git diff HEAD --check`: PASS；unmerged-path query: 空；tracked diff allowlist: PASS。

## Tagged Unit Baseline Comparison

最终候选工作树与不可变基线
`E:/codex-worktrees/sub2api/upstream-v027-next-baseline`（HEAD
`38a5a6040faa065a795c2ee81c5b11e12902164c`）各自实际运行：

```text
go test -tags unit ./internal/service -run 'TestForwardResponses_ForceChatCompletions|TestTokenRefreshService|TestV027' -count=1 -v -gcflags=all=-e
```

二者均在 test binary 生成前以同一既有 fixture 漂移失败：`stringPtr` redeclare、过时的
`computeTokenBreakdown` / `calculateCostInternal` / `buildCountTokensRequest` signatures、
以及旧 Proxy `FallbackMode` / `ExpiryWarnDays` fields；`-gcflags=all=-e` 的额外
`billing_service_test.go` 未使用 `context` import 也一致。这不是 tagged 全包 PASS，
且不能归因给本 Sprint；隔离集合的通过仅为相关 regression evidence。

## Limitations

隔离 tagged 集合内的既有
`TestTokenRefreshService_RefreshWithRetry_Antigravity` 触发了 privacy fixture 对
`daily-cloudcode-pa.sandbox.googleapis.com/v1internal:setUserSettings` 的请求，10 秒
context deadline 后该测试仍 PASS。此为既有 fixture 的外部副作用，不是本 Sprint 代码
路径；本 QA 不再执行可能触网的同类测试。除该意外 fixture 请求外，V027 验收使用
local fake；未进行真实 DeepSeek/OAuth 验证、数据库、容器、部署、提交或推送。

审查的 7 个 Go 文件均处于合同 allowlist；`docs/workflow/main-log.md` 与
`docs/workflow/status.md` 是主控 workflow 改动。`bcc.patch`、`oauth.patch`、
`patch.diff`、`patch.txt` 四个未跟踪 patch 均保留未动。

### DONE: upstream-runtime-hotfixes-s66a

# Worker Result

## Task ID

`upstream-runtime-hotfixes-s66a`

## Status

`done`

## Summary

- Anthropic `setup-token` 账号在 `refresh_token` 非空时会被 `ClaudeTokenRefresher` 接受；普通 OAuth 的既有资格判断保持不变。
- 本地后台刷新候选入口使用 `ListActive`，原本就不按账号类型过滤，因此没有引入上游已被本地架构替代的 `ListOAuthRefreshCandidates` SQL。
- `opsCaptureWriter` 在释放后对全部 `gin.ResponseWriter` 委托方法提供 nil-safe 行为，并覆盖回归测试。
- OpenAI WS v2 会将 Windows `forcibly closed by the remote host` 重置文案识别为 disconnect。
- 已满足 contract 的功能准则；contract 中 repository 正则未命中本地测试，已如实记录。

## Changed Files

- `backend/internal/service/token_refresher.go`
- `backend/internal/service/token_refresher_test.go`
- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/handler/ops_capture_writer_nil_test.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay_internal_test.go`
- `docs/workflow/worker-results/upstream-runtime-hotfixes-s66a-result.md`

## Commands Run

```text
go test ./internal/repository -run "Test.*OAuthRefreshCandidates|Test.*RefreshCandidate" -count=1 -> PASS, no tests to run
go test ./internal/service -run "TestClaudeTokenRefresher|Test.*DisconnectError" -count=1 -> PASS; TestIsOpenAIWSClientDisconnectError matched
go test ./internal/handler -run "TestOpsCaptureWriter" -count=1 -> PASS
go test ./internal/service/openai_ws_v2 -run "Test.*Disconnect" -count=1 -> PASS
go test ./internal/service -run "^TestClaudeTokenRefresher" -count=1 -v -> PASS with only token_refresher_test.go unit tag temporarily removed, then restored
go test -tags=unit ./internal/service -run "^TestClaudeTokenRefresher" -count=1 -> BLOCKED by unrelated existing unit-test compile failures
git diff --check -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/repository [no tests to run]
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/handler
ok github.com/Wei-Shaw/sub2api/internal/service/openai_ws_v2
PASS TestClaudeTokenRefresher_CanRefresh/anthropic_setup-token_with_refresh_token_-_can_refresh
PASS TestClaudeTokenRefresher_CanRefresh/anthropic_setup-token_without_refresh_token_-_cannot_refresh
```

## Risks

- Windows reset 分类使用合成 error 文案验证，未在真实 Windows WebSocket 断开环境中端到端复现。
- 完整 `-tags=unit` service 套件仍被既有 billing/ops/Grok 测试编译错误阻断；已通过临时单文件 tag 解除单独验证本任务 Claude 测试，最终文件已恢复原 tag。

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.

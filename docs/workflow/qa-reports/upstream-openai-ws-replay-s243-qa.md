### PASS: upstream-openai-ws-replay-s243

Independent QA report.

Checks to execute in this worktree:

- focused replay, coverage, and bridge tests with `-count=10`
- complete `go test ./internal/service`
- `go test ./cmd/server -run '^$' -count=1`
- `gofmt`/`git diff --check`
- exact allowlist, no conflict markers, empty unmerged index
- upstream source ancestry and protected-main dirty-state comparison

Results: focused selector passed with `-count=10`; complete service passed in
`70.190s`; server compile passed in `0.064s`; formatting and diff checks passed.
Allowlist, ancestry, conflict/index, and protected-main checks passed. No
provider, database, container, deployment, or push operation was run.
Contract amendment `38825fa63` added `backend/internal/service/openai_tool_continuation_test.go` to the exact allowlist. No business files changed during the amendment rerun.

#### 已通过的行为与构建证据

- focused replay/coverage/bridge：`go test ./internal/service -run "TestBuildOpenAIWSReplayInputSequence|TestAnalyzeToolCallOutputContextCoverageBytes|TestOpenAIWSHTTPBridge.*Replay|TestOpenAIWSRawPayloadHasToolCallOutput" -count=10` 通过，selector 实际发现并执行测试。
- 完整服务包：`go test ./internal/service` 通过（68.292s）。
- server compile：`go test ./cmd/server -run '^$' -count=1` 通过。
- `gofmt -d` 对相关变更 Go 文件无输出；`git diff --check` 通过。
- focused 覆盖 array/object input coverage、orphan historical custom filtering、paired function/custom preservation、item-reference non-pairing、current-turn preservation 及 bridge replay 请求行为。

#### 完整性与保护检查

- `25da02ddd`、`66808413d` 均为 `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54` 的祖先。
- `git ls-files -u` 为空；未发现精确冲突标记。
- QA worktree 在报告写入前保持干净；未执行 provider、数据库、容器、部署或 push 操作。
- 主工作区 `F:/mcplugins/sub2api` 未被写入，既有 dirty/untracked 状态保持不变。复核 hash：
  - `backend/internal/service/api_key_auth_cache.go`: `536125c6859a866bb841145711976d2e8eee4a5b`
  - `backend/internal/service/group_buy.go`: `665b9911b9f78efec1b6a6385c4864f33da22827`

#### 结论

功能测试、构建、exact allowlist、provenance、冲突/index 和主线保护均通过。

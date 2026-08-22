### PASS: upstream-openai-client-tools-s242

独立 QA worktree：`E:/codex-worktrees/sub2api/upstream-openai-client-tools-s242`

#### 验收证据

- focused apicompat：`go test ./internal/pkg/apicompat -run "TestResponsesClientTool|TestAdaptResponsesClientTools" -count=10` 通过。
- focused service：`go test ./internal/service -run "TestOpenAIPassthroughAPIKey|TestOpenAIWSHTTPBridge.*ClientTool|TestProxyOpenAIWSHTTPBridgeTurnAPIKey.*ClientTools" -count=10` 通过。
- 完整包：`go test ./internal/pkg/apicompat` 通过；`go test ./internal/service` 通过（65.55s）。
- 编译：`go test ./cmd/server -run '^$' -count=1` 通过。
- `gofmt -d` 对 7 个变更 Go 文件无输出；`git diff --check` 通过。
- selector 可发现并实际执行测试：覆盖 custom-only lowering、ordinary function/namespace no-op、非流式/SSE restoration、WS 首轮及省略 tools 的 follow-up 继承、显式 tools 替换、trailing JSON rejection、request mapping 清理。

#### 范围与完整性

- 开发提交：`dd8693bbbd217449015868891d2d2da5b8db2c52`，相对 frozen local base `baa6541acb1ef909b85d6d3cdb4817d2da0564c9` 的业务文件均在 S242 allowlist 内；提交还包含 contract/result 文档，未发现 allowlist 外业务路径。
- upstream provenance：`44ef88f65`、`7e579cb28`、`b94e484e2` 均为 `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54` 的祖先。
- `git ls-files -u` 为空；无精确冲突标记；开发 worktree 无未提交变更。
- 未执行 provider、Redis/PostgreSQL、container、deployment 或 push 操作。

#### 主工作区保护

QA 期间未向 `F:/mcplugins/sub2api` 写入文件。主工作区保持用户既有 dirty 状态，且 dirty 路径均为 S242 denied/protected 范围之外的既有工作；本次 QA 不包含这些路径。当前受保护文件 hash（主工作区）：

- `backend/internal/service/api_key_auth_cache.go`: `536125c6859a866bb841145711976d2e8eee4a5b`
- `backend/internal/service/group_buy.go`: `665b9911b9f78efec1b6a6385c4864f33da22827`

残余风险：未连接真实 function-only provider；验证依据为本地 fake upstream、单元测试、包级 runtime 测试和 server compile。

### PASS: upstream-v027-next-backend — 修订合同已锁定当前 local owner、暂停不解锁边界与 DeepSeek plaintext 保留验收

# Contract Review

## Verdict

`PASS`。修订后的 Local owner clarification 已将两项行为准确绑定到现有 local owner，所有必要生产与测试文件均在 allowlist 内，验收能区分功能补齐与回归。

## Owner And Scope

- DeepSeek 逻辑的唯一生产 owner 为 `backend/internal/service/openai_gateway_responses_chat_fallback.go` 中的 `forwardResponsesViaRawChatCompletions`：`chatBody` 序列化之后、`http.NewRequestWithContext` 之前。该时点已完成本地 Responses 转换，又只影响此 fallback，不会扩大到 native Chat。
- Token refresh 的生产 owner 为 `backend/internal/service/token_refresh_service.go` 中的 `TokenRefreshService.listActiveAccounts`。`ListActive` 已过滤 `StatusActive`；任务只移除其后对 `Schedulable=false` 的二次排除。非 OAuth、无 refresh token、error/disabled 和重试冷却仍由现有 refresh loop/candidate 检查处理。
- `postRefreshActions` 仅清除临时不可调度状态和缓存；凭据持久化只写 credentials。合同要求保持该路径不调用 `SetSchedulable(true)`，所以管理员 pause 不会因成功刷新被解除。

## Upstream Behavior Mapping

- `bcc73f8d4` 的核心语义是仅在 DeepSeek 平台或精确 `api.deepseek.com` host 上，对缺失/空 assistant `reasoning_content` 填入单个空格；已有非空 plaintext、非 assistant 与其他上游保持不变。修订合同将该语义映射到本地实际 fallback 出站点，无需引入上游已不存在的 split pipeline owner。
- `8e34ca5e3` 的核心语义是 paused active OAuth 账号仍可刷新，但刷新只更新凭据，不能解除管理员 `Schedulable=false`。本地通过移除 `listActiveAccounts` 的二次过滤实现相同结果，不需要修改 repository。

## Acceptance Sufficiency

- `TestV027DeepSeekReasoning` 覆盖真实 `forwardResponsesViaRawChatCompletions` 出站体：encrypted-only/cache miss 补空格、缓存命中或已有 plaintext 保留、streaming 与 nonstreaming，以及非 DeepSeek/严格 host 边界不注入。
- `TestV027PausedRefresh` 使用 fake repo/refresher 覆盖 paused active 候选、成功后的 `Schedulable=false` 保持、active enabled 正常刷新、repo error 和 non-OAuth 排除；同时保留 invalid-grant、重试与临时冷却既有回归命令。
- `go test` 定向命令、服务回归命令、`go build ./...`、`gofmt` allowlist 检查、`git diff HEAD --check` 和未合并路径检查共同覆盖编译、行为与范围。测试仅使用本地 fake，不声明真实 OAuth 或 DeepSeek provider 已验证。

## Gate Checks

- success_criteria_testable: `PASS`
- local_owners_explicit: `PASS`
- allowed_paths_explicit: `PASS`
- denied_paths_explicit: `PASS`
- admin_pause_non_regression: `PASS`
- plaintext_and_host_boundaries: `PASS`
- acceptance_commands_executable: `PASS`
- base_commit_confirmed: `PASS` (`38a5a6040faa065a795c2ee81c5b11e12902164c`)

## Scope Boundary

本评审批准隔离 worktree 内的最小实现和本地 fake 验收；不批准主工作区写入、repository/migration 改动、真实账户刷新、真实 DeepSeek 请求、提交、推送或部署。实现与独立 QA 仍须分别完成。

## Opt-in Bridge Amendment Review

### PASS: upstream-v027-next-backend opt-in bridge amendment

该修订可实施，且比 fallback-local 的事后恢复更可靠。当前 `responsesInputToChatMessages` 在单次遍历中负责函数调用合并、无效调用跳过、tool-output media 提取和最终重排；在该遍历中把 plaintext reasoning 绑定到其后的 assistant（含 assistant tool-call batch），能够保持同一轮的关联，避免按转换后索引猜配造成 parallel calls、跳过无效调用或媒体插入后的错位。

- 新入口必须是明确命名的 opt-in；现有 `ResponsesToChatCompletionsRequest` 的签名、默认输出和所有非 DeepSeek 调用保持不变，只有 DeepSeek fallback 选择 opt-in。
- opt-in 遍历只累计 `summary_text`/`reasoning_text` 的 plaintext，遇到用户或 tool-output 边界即丢弃未消费的 pending reasoning；`encrypted_content` 绝不转发为 plaintext，也不引入缓存或服务。
- pending reasoning 必须在随后的 assistant content 或 assistant tool-call batch 上写入 `ReasoningContent`，并在 `normalizeChatMessagesWithToolOutputMedia` 后仍与该 assistant 保持同一位置。显式非空 `reasoning_content` 优先保留；工具调用无匹配 reply 时也不得因归并而丢失已绑定的 reasoning assistant。
- 合同新增的 plaintext/encrypted、多轮、并行工具、无效调用、缺失 reply、媒体插入、默认转换不变测试，正好覆盖上述关联和非干扰不变量；`go test ./internal/pkg/apicompat -count=1` 是公共 bridge 的必要回归门禁。

因此不建议 fallback-local 的按输出列表恢复方案；它无法在重排后可靠重建 Responses item 与 assistant/tool batch 的一一关系。该 amendment 的两条新增 allowlist 路径足够，且无需扩大到公共调用方或新架构。

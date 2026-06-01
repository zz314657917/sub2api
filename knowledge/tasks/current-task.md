# 当前任务快照

最后更新：2026-06-01 17:34 +08:00

## 背景

用户要求继续按“v0.1.133 关键修复移植计划”做下一批，不执行 `v0.1.133` 整体 merge。主工作区 `F:/mcplugins/sub2api` 仍有并行未提交改动，因此本批在独立 worktree `F:/mcplugins/.codex-worktrees/sub2api-v0133-batch2` 和分支 `codex/upstream-v0.1.133-batch2` 完成。

## 当前目标

选择性移植上游 v0.1.133 的第二批低/中风险关键修复，重点是 OpenAI WebSocket/Responses 兼容、计费长上下文 cache 价格修复，以及已存在 Opus 4.8 支持中的 Bedrock 模型 ID 小修。继续避开账号配额自动暂停、风控运行态、DingTalk OAuth、大迁移和前端大替换。

## 本次已完成

- 提交 `d41955c69 fix: port upstream websocket compatibility fixes`：移植 OpenAI WS rate-limit failover、Responses/Chat/Anthropic 兼容、`response.failed` 流式错误终态、裸 `/responses` 路由识别和最小 OAuth 429 storm 停止换号逻辑。
- 提交 `e6aa3a150 fix: apply long context multipliers to cache billing`：移植长上下文计费对 `cache_read` 和 `cache_creation` 的倍率修复及回归测试。
- 提交 `e676580b1 fix: correct bedrock opus 4.8 model id`：手工修正 `claude-opus-4-8` Bedrock 默认模型 ID 为 `us.anthropic.claude-opus-4-8-v1`，并补 regional resolve 测试。

## 已确认事实

- `b34cc71be` 和 `cff2f291b` 的核心行为已在当前分支等效存在：`ensureForwardErrorResponse` 在 `Writer.Written()` 后继续追加 SSE，`inboundIsResponses` 覆盖 `/v1/responses`、裸 `/responses` 和 `/backend-api/codex/responses`。
- `68901cbff` 是整份 `backend/resources/model-pricing/model_prices_and_context_window.json` 大规模替换，包含大量模型元数据删除/重排，本批未纳入。
- `514ac5c6a` 整体包含迁移 `144`、前端账号/用量页和 Bedrock beta 测试语义变化；本批只吸收已确认的小修，不 cherry-pick 整包。

## 待验证点

- 后续如果要继续追模型/定价元数据，需要单独审 `68901cbff`，验证是否会删除本地仍依赖的模型条目。
- 后续如果要完整纳入 Opus 4.8 相关迁移/前端白名单，需要另开迁移编号方案，避免与本地 `162/163` 冲突。
- 若继续移植 v0.1.133，下一批应重新从 clean worktree 状态开始筛选，不要回到主工作区直接操作。

## 当前结论

Batch2 已完成并验证通过，分支 `codex/upstream-v0.1.133-batch2` 当前 HEAD 为 `e676580b1`。代码 worktree 干净；仅 `knowledge/tasks/current-task.md` 和 `knowledge/tasks/timeline.md` 作为本地交接记录有未提交修改。

## 下一步

- 合并策略：如要纳入主工作线，优先把 `d41955c69..e676580b1` 作为 batch2 代码提交范围 review/merge。
- 下一批移植：继续按主题审 upstream v0.1.133 剩余小修，仍跳过账号配额、风控、DingTalk、迁移重排和前端大页替换。
- 记录维护：若后续切回主工作区，应先 `git status --short`，确认不会把主工作区 payment/affiliate/tutorial/i18n 等并行改动混入上游修复批次。

## 验证记录

- `git diff --check`：通过。
- `git diff HEAD~3..HEAD --check`：通过。
- `cd backend && go test ./internal/pkg/apicompat/... ./internal/service/... ./internal/handler/... ./internal/server/...`：通过。
- `cd backend && go test -tags unit ./internal/service -run "TestCalculateCost_(OpenAIGPT54LongContextAppliesMultiplierToCacheRead|OpenAIGPT54NoLongContextKeepsCacheReadAtBasePrice|OpenAIGPT54LongContextAppliesMultiplierToCacheCreation|OpenAIGPT54NoLongContextKeepsCacheCreationAtBasePrice|LongContextAppliesMultiplierToCacheCreation5mAnd1h)$"`：通过。
- `cd backend && go test ./internal/domain ./internal/service -run "TestDefaultBedrockModelMapping_ClaudeOpus48|TestResolveBedrockModelID"`：通过。

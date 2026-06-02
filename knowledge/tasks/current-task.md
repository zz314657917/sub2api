# 当前任务快照

最后更新：2026-06-03 03:06 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 用户要求：整理本地分支，合并并提交一波。
- 当前 `main` 已有大量本地提交，且曾存在一组未提交的模型广场参考价改动。
- `tmp-doubao.html`、`tmp-kling26.html`、`tmp-modelList.html`、`tmp-ui-check/` 仍是未跟踪临时采证文件，本轮不纳入提交。

## 当前目标

- 提交当前已跟踪工作区改动。
- 把仍未合入 `main` 的本地同步链合入主线。
- 确认本地分支合入状态，并保留验证结果。

## 本次已完成

- 已执行 `git fetch --all --prune`，刷新 `origin` 与 `upstream` 引用。
- 已确认这些本地分支已经合入 `main`：
  - `codex/archive-chat-images-studio-reshape`
  - `codex/fix-chat-image-embedded-session`
  - `codex/sub2api-studio-layout`
  - `codex/upstream-v0.1.133-batch2`
  - `codex/upstream-v0.1.133-batch3`
  - `codex/upstream-v0.1.133-critical-fixes`
  - `pge/sub2api-canvas-editor-core`
  - `pge/sub2api-canvas-run-control`
- 已提交模型广场参考价改动：`f1a9e87c6 feat(models): show reference pricing in plaza`。
- 已合并仍未进入 `main` 的上游同步链：`codex/upstream-main-openai-ws-usage-dedup-s2k`。
- 合并冲突仅发生在 `knowledge/tasks/current-task.md`，已重写为当前主工作区事实。

## 已确认事实

- `codex/upstream-main-openai-ws-usage-dedup-s2k` 包含此前分批同步的 S1/S2 系列改动、QA 报告和 workflow 记录。
- 本次没有直接 merge `upstream/main`；剩余上游差异仍需单独评估。
- 模型广场 `reference_pricing` 只用于展示，不写入渠道价格，不影响 billing、倍率和扣费链路。
- 前端对旧后端响应兼容：`supportedModel.reference_pricing ?? null`。

## 验证记录

- 提交模型广场改动前执行 `git diff --check -- <本轮相关文件>`：通过，仅 `knowledge/tasks/current-task.md` 有既有 CRLF warning。
- 模型广场改动历史验证：
  - `go test -tags=unit ./internal/service -run "TestFillReferencePricing|TestListAvailable|TestPricingNeedsFallback|TestSynthesizePricingFromLiteLLM|TestGetModelPricing_ClaudeDecimalAliasMatchesHyphenatedPricing" -count=1`：通过。
  - `go test -tags=unit ./internal/handler -run "TestToUserSupportedModels|TestUserAvailableChannel_FieldWhitelist|TestBuildPlatformSections" -count=1`：通过。
  - `go test ./internal/service ./internal/handler -run TestNonExistent -count=1`：通过，仅编译目标包。
  - `corepack.cmd pnpm --dir frontend exec vitest run src/__tests__/public-pages.spec.ts`：通过，9 个用例。
  - `corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/ChatStudioView.spec.ts`：通过，7 个用例。
  - `corepack.cmd pnpm --dir frontend run typecheck`：通过。
  - 后续多次 `npm.cmd run test:run -- public-pages`、`npm.cmd run typecheck`、`npm.cmd run build` 和 `/models` 浏览器检查均通过，build 仅有项目既有 Vite chunk、Browserslist 和 Node `DEP0190` 警告。
- 上游同步链原始 S2k 验证记录：
  - `git diff --check`：通过。
  - `go test ./internal/service -run "OpenAIGatewayServiceRecordUsage_(PrefersClientRequestIDOverUpstreamRequestID|WSModePrefersUpstreamRequestIDOverClientRequestID|GeneratesRequestIDWhenAllSourcesMissing)" -count=1`：通过。

## 下一步

- 完成本次 merge commit 后，重新检查 `git branch --no-merged main` 和 `git status --short --branch`。
- 运行合并后的关键验证，至少覆盖后端 service/handler 编译与前端 typecheck。
- 如需共享结果，push 当前 `main`；如需清理临时采证文件或已合并分支，需单独确认后删除。

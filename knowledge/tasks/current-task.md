# 当前任务快照

最后更新：2026-06-09 00:42 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前默认续做入口是上游低风险合成，不是模型市场主线。
- 本地 `main` 已推送到 `origin/main=cbdb69bed`。
- 上游 `upstream/main=be0174456`，最新相对 `v0.1.135` 之后主要是 README/sponsors 类更新；功能候选大多已在前序 Sprint 评估或合成。

## 当前主线

- OpenAI gateway/auth/session/prompt cache 稳态已经进入本地主线：
  - API Key 独占分组访问。
  - sticky session 分组校验。
  - 跨组失配 `previous_response_id` 剥离。
  - `/responses` transport failover + 持久故障账号临时摘除。
  - Chat Completions `prompt_cache_key` 传播并按 API Key 隔离 session。
- 用量窗口与 Ops 告警也已补齐：
  - 5h `ResetsAt` 已同步到 `SessionWindowEnd`，过期窗口展示归零。
  - Ops 告警新增 `account_temp_unscheduled_count`，可覆盖临时不可调度账号。
- 模型市场、APIMart 计费、工单、Canvas、Chat/Image Studio、OpenWebUI 仍是本地产品线，后续合上游时默认保护，不跟随上游删除或覆盖。

## 已完成的最近 Sprint

- `upstream-main-release135-gateway-auth-s11`
  - release `0.1.135` gateway/auth/session 修复已合流。
- `upstream-main-prompt-cache-s12`
  - `d251487da` 已合流，Chat Completions 传播 `prompt_cache_key`。
- `upstream-main-usage-window-s13`
  - `16bc87693` 已合流，5h usage window `ResetsAt` 语义修复已进入主线。
- `upstream-main-ops-alert-temp-unscheduled-s14`
  - `f20e6bf76` 已合流，Ops alert 支持 `account_temp_unscheduled_count`。
  - S14 验证已在 `main` 通过并推送：`git diff --check`、denied path audit、`go test -tags unit ./internal/service -run "ComputeRuleMetric|TempUnscheduled|OpsAlert" -count=1`、`go test ./internal/handler/admin -run "OpsAlert|Metric" -count=1`、`go test ./internal/service ./internal/handler/admin -count=1`、`corepack.cmd pnpm --dir frontend run typecheck`。

## 下一步候选

- 低风险可选：
  - `f5cecea5b`：纯前端 Select 下拉高度修复，可单独小 Sprint。
- 继续延后：
  - `af19d4432`：代理有效期与失败回退，涉及 schema/migration/API/frontend，需单独大 Sprint。
  - README/sponsors/VERSION/docs-only 提交，默认跳过。
- 如果继续上游合成，先 `git fetch upstream --tags --prune`，再从当前 `main` 开独立 `codex/` 分支或隔离 worktree；不要直接 merge `upstream/main`。

## 证据入口

- Sprint contracts：
  - `docs/workflow/tasks/upstream-main-release135-gateway-auth-s11.md`
  - `docs/workflow/tasks/upstream-main-prompt-cache-s12.md`
  - `docs/workflow/tasks/upstream-main-usage-window-s13.md`
  - `docs/workflow/tasks/upstream-main-ops-alert-temp-unscheduled-s14.md`
- QA / result：
  - `docs/workflow/qa-reports/upstream-main-ops-alert-temp-unscheduled-s14-qa.md`
  - `docs/workflow/worker-results/upstream-main-ops-alert-temp-unscheduled-s14-result.md`
- 汇总流水：`docs/workflow/main-log.md`

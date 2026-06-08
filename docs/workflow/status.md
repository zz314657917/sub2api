---
phase: done
current_sprint: upstream-main-prompt-cache-s12
total_sprints: 12
pending_action: wait-next-upstream-batch-or-local-followup
project_type: web
qa_mode: runtime
approval_required: false
last_verified: 2026-06-08
---

# Workflow Status

- 当前阶段：`done`
- 当前 Sprint：`upstream-main-prompt-cache-s12`
- 当前目标：把最近两轮已合流的 OpenAI gateway/auth/session 与 prompt cache 修复收口为当前默认续做语境，避免入口继续停留在模型市场 `version=11` 主线。
- 当前结论：`upstream-main-release135-gateway-auth-s11` 与 `upstream-main-prompt-cache-s12` 都已完成 implementation、QA 和 integration；现有 main 分支默认应先按 gateway auth / sticky session / prompt cache 的稳定约束理解。
- 当前已稳定进入默认主线的事实：
  - API Key exclusive group access 已成为默认鉴权边界。
  - sticky session 复用前要验证分组仍匹配；跨组切换时应剥离失配的 `previous_response_id`。
  - `/responses` 非流式 transport 错误已走 failover，并对持久故障账号做临时摘除。
  - release `0.1.135` 相关 gateway/auth/session 修复已经完成本地合成和验证。
  - Chat Completions `prompt_cache_key` 传播已合入，prompt cache session 默认需按 API Key 隔离理解。
- 目标验证入口：
  - `docs/workflow/tasks/upstream-main-release135-gateway-auth-s11.md`
  - `docs/workflow/worker-results/upstream-main-release135-gateway-auth-s11-result.md`
  - `docs/workflow/qa-reports/upstream-main-release135-gateway-auth-s11-qa.md`
  - `docs/workflow/tasks/upstream-main-prompt-cache-s12.md`
  - `docs/workflow/worker-results/upstream-main-prompt-cache-s12-result.md`
  - `docs/workflow/qa-reports/upstream-main-prompt-cache-s12-qa.md`
  - `docs/workflow/main-log.md`
- 下一合法动作：等待下一批上游合成候选，或在本地继续围绕 OpenAI gateway / auth / sticky session / prompt cache 做 follow-up；如果再开新 Sprint，不要复用旧的模型市场状态页文案。
- 状态推进规则：先 `spec-approved`，再进入当前 Sprint 的 `contract-draft -> contract-approved -> build -> qa -> fix -> retest -> done`。

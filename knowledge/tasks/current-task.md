# 当前任务快照

最后更新：2026-06-08 13:45 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 6 月上旬的模型市场公开目录、后台维护、`¥` 展示口径、分组倍率换算和 `gpt-image-2-official` 人民币余额扣费修正，已经从“当前主线”退成稳定背景层。
- 最近两轮上游合成已经把默认续做入口继续前移到 OpenAI gateway / auth / sticky session / prompt cache 这一层：
  - `upstream-main-release135-gateway-auth-s11`
  - `upstream-main-prompt-cache-s12`

## 当前主线

当前默认续做入口不再是模型市场，而是 OpenAI 网关认证、会话复用边界和 prompt cache 稳态：

- API Key 独占分组访问已经进入默认约束，跨组误用不再应被视为可接受行为。
- OpenAI sticky session 现在要同时满足“分组成员校验”“跨组切换时剥离失配的 `previous_response_id`”“WSv2 生效路径保护”。
- `/responses` 非流式传输错误已经进入 failover + 持久故障临时摘除账号的默认心智。
- release `0.1.135` 的 gateway/auth/session 修复已经合流，后续再看相关报错时，不应继续按旧 transport 假设排查。
- Chat Completions 的 `prompt_cache_key` 传播已经成为新默认约束；后续涉及 prompt cache session 复用时，应按“随 API Key 隔离”理解，而不是把所有请求混进同一缓存会话。

## 当前结论

- `knowledge/00-start-here.md`、`knowledge/05-current-focus.md`、`docs/ai/current-task.md` 已经跟到 6 月 7 日的 OpenAI 网关稳态 / account capability routing / 控制台入口语境，但 `knowledge/tasks/current-task.md` 与 `docs/workflow/status.md` 之前仍停在模型市场 `version=11` 语境，容易把接手者带回旧主线。
- 当前真正更值得优先沉淀的是：
  - gateway auth / sticky session / previous response 保护
  - API Key exclusive group access
  - prompt cache key 传播与隔离
  - 非流式 JSON 与 `/responses` transport failover 的稳态
- 模型市场相关能力仍然有效，但更适合作为稳定背景事实，不该继续占据当前任务快照的最前面。

## 已稳定事实

- `upstream-main-release135-gateway-auth-s11` 已完成 implementation、QA 和 integration 记录。
- `upstream-main-prompt-cache-s12` 已完成 implementation、QA 和 integration 记录。
- 当前排查 OpenAI 兼容问题时，优先检查：
  - `backend/internal/handler/openai_gateway_handler.go`
  - `backend/internal/handler/gateway_handler_chat_completions.go`
  - `backend/internal/handler/gateway_handler_responses.go`
  - `backend/internal/service/openai_gateway_service.go`
  - `backend/internal/service/openai_gateway_messages.go`
- 当前排查 session / prompt cache 问题时，不要只看请求参数；还要一起看 API Key、分组、sticky session 和 prompt cache key 是否落在同一条约束链上。

## 下一步

- 如果继续做 OpenAI 兼容、账号鉴权、会话复用或 transport 稳态，先从 `docs/workflow/tasks/upstream-main-release135-gateway-auth-s11.md` 与 `docs/workflow/tasks/upstream-main-prompt-cache-s12.md` 回读目标和边界。
- 如果继续补稳定知识，优先把“gateway auth + sticky session + prompt cache”作为当前默认主线，而不是继续扩写模型市场历史细节。
- 如果没有新的上游合成批次，本文件保持当前快照即可；新增阶段历史再写入 `knowledge/tasks/timeline.md`。

## 验证与证据入口

- Sprint 记录：
  - `docs/workflow/tasks/upstream-main-release135-gateway-auth-s11.md`
  - `docs/workflow/worker-results/upstream-main-release135-gateway-auth-s11-result.md`
  - `docs/workflow/qa-reports/upstream-main-release135-gateway-auth-s11-qa.md`
  - `docs/workflow/tasks/upstream-main-prompt-cache-s12.md`
  - `docs/workflow/worker-results/upstream-main-prompt-cache-s12-result.md`
  - `docs/workflow/qa-reports/upstream-main-prompt-cache-s12-qa.md`
- 汇总流水：`docs/workflow/main-log.md`
- 最近主线提交：
  - `15f01494f` `fix: enforce exclusive group access for api keys`
  - `064f35021` `fix: validate OpenAI sticky session groups`
  - `9d69c1c09` `fix(openai): /responses 传输层错误转 failover + 持久故障临时摘除账号`
  - `541fe39c1` `fix(openai): adapt release135 transport and sticky tests`
  - `d1812704c` `fix(openai): add raw tool continuation validation`
  - `69e2d54a8` `fix(openai): propagate prompt cache key for chat completions`

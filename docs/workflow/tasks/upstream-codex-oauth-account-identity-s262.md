# Task Contract

## Task ID

`upstream-codex-oauth-account-identity-s262`

## Role

你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

将已通过行为审查但未通过独立 QA 的 S258 Codex OAuth 凭据级身份隔离，基于当前 `main` 及已合入的 Spark shadow credential 链重新适配。每条 OAuth 出站路径都必须以实际凭据来源（shadow 时为 parent）建立命名空间，并用真实 HTTP/WebSocket 出站工件证明不会跨账号或在握手中二次映射。

## Success Criteria

- 对 OpenAI OAuth / Setup-token 账号，`client_metadata`、嵌套 turn metadata、`prompt_cache_key`、`session_id`、`conversation_id` 与白名单 Codex 身份头都按 `API key + resolved credential account + client-original value` 稳定隔离；不得使用本地 child 行 ID 或泄露 token。
- 所有 body 改写从客户端原始值导出。普通 WS 握手必须从原始 prompt-cache/session 值生成一次 scoped header；不得把已改写 body 值再次 scope。相同 `prompt_cache_key` 与 `client_metadata.session_id` 必须生成相同 scoped 值。
- 真实 fake-upstream 回归必须覆盖普通 HTTP Forward、Chat Completions、Anthropic Messages、raw passthrough、普通 WS，以及 v2 passthrough 的首帧、`session.update` 和后续 `response.create`。每条被测路径都断言实际出站 body/header/frame，而非只测 helper。
- 基于 S259 的 `resolveCredentialAccount`，shadow child 必须使用 parent 的凭据命名空间；同一 gin context 发生 shadow -> 另一 OAuth 账号切换时，后者的身份和 header/body 不能遗留 parent 来源。缺 parent 必须在任何 HTTP/WS 出站前失败。
- API Key、非 OpenAI、无稳定 OAuth 凭据命名空间路径维持既有 `isolateOpenAISessionID` 行为；fingerprint off/on 顺序、turn-state guard、工具名映射、Responses Lite、Fast Policy 与现有 S259 shadow 调度行为不得回归。

## Context

- Repo: `F:/mcplugins/sub2api`
- Worker worktree: `E:/codex-worktrees/sub2api/upstream-codex-oauth-account-identity-s262`
- Frozen base: `main@5cca17b1480bdbf5891bf09e79e5306f1903b5d7`
- Upstream behavior source: `d493ce0bb2959c51c8aa0da32269adccb22302c1`
- Prior unmerged implementation/reference: `pge/upstream-codex-oauth-account-identity-s258@7239eb489f5f161b2fe229747506519014c8ba46`
- Prior QA failure: ordinary WS outbound proof was missing at `74143e32e`; later S258 test-only commits partly addressed it but did not add a Messages capture and were never independently retested.
- Read first: this contract, `docs/workflow/status.md`, `backend/internal/service/openai_agent_identity.go`, `backend/internal/service/openai_spark_shadow_credential_outbound_test.go`, and the S258 candidate diff.

## Allowed Paths

- `backend/internal/service/openai_codex_account_identity.go`
- `backend/internal/service/openai_codex_account_identity_test.go`
- `backend/internal/service/openai_agent_identity_compat_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_session_isolation_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_messages_usage_test.go`
- `backend/internal/service/openai_compat_model_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_success_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter_effort_test.go`
- `backend/internal/service/openai_spark_shadow_credential_outbound_test.go`
- `docs/workflow/worker-results/upstream-codex-oauth-account-identity-s262-result.md`

## Denied Paths

- `knowledge/**`, `frontend/**`, `backend/ent/**`, `backend/migrations/**`, `backend/cmd/server/wire_gen.go`, handlers, config, dependencies, routes, containers, deployment, databases, providers, push, and all paths not in Allowed Paths.
- `F:/mcplugins/sub2api` primary worktree, including all Pixel Cafe, workflow, untracked `outputs/` and `sub2api` entries, is read-only.

## Constraints

- Do not cherry-pick the old S258 branch. Port only compatible behavior onto S259's current credential resolver; preserve scheduler semantics of the logical selected account.
- Resolve a shadow credential source once per selected attempt and overwrite any gin-context staged source. Never append state from a prior account attempt.
- Raw passthrough remains a hot path: only decode and splice the small `client_metadata` object; do not full-unmarshal a potentially large request body.
- Never place bearer tokens, credential IDs from real accounts, or raw token fingerprints in reports, logs, source comments, or test output.
- Keep scope minimal; no production/provider/DB/browser/container activity and no push.

## Acceptance Commands

```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
New-Item -ItemType Directory -Force '.tmp/s262-go-build' | Out-Null
$env:GOTMPDIR = (Resolve-Path '.tmp/s262-go-build').Path
Push-Location backend
go test ./internal/service -list 'Test.*(CodexAccountIdentity|AccountIdentity|SparkShadowOutbound)'
go test ./internal/service -run 'Test.*(CodexAccountIdentity|AccountIdentity|SparkShadowOutbound)' -count=10
go test ./internal/service -run 'Test(ForwardAsChatCompletions|ForwardAsAnthropic|OpenAIGatewayService_Forward_WSv2|OpenAIGatewayService_ProxyResponsesWebSocketFromClient).*' -count=1
go test ./internal/service -count=1 -timeout=3m
go test ./cmd/server -run '^$' -count=1
Pop-Location
gofmt -w <each changed Go file>
git diff --check
```

- Before reporting, inspect exact allowlist relative to the frozen base, conflict markers, cached/unmerged index, and record a read-only primary-worktree porcelain/patch-id snapshot before and after the task. The primary snapshot may change externally; worker changes must cause no drift.

## Output

- Commit one scoped implementation commit on the assigned worker branch; write and commit the result report separately.
- Use `C:/Users/Administrator/.codex/templates/worker-result.md`.
- The first report line must be `### DONE: upstream-codex-oauth-account-identity-s262`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Include changed files, actual outbound assertions for each required protocol, command results, residual risks, contract compliance, upstream provenance, and `knowledge_candidates`.

## Stop Rules

- Stop as `BLOCKED` if the current resolver cannot expose the resolved credential account without changing denied architecture, or if the fake upstream harness cannot capture a required outbound artifact.
- Stop before changing any denied path or the primary worktree.
- This is the replacement contract after S258's two incomplete attempts. Do not create a further low-cost worker retry on failure; return a precise failure to Codex.

## Budget

- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

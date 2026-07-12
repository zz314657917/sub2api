### DONE: upstream-openai-fast-flex-user-scope-s71

## Worker Deviation

- Agent Matrix 指定的 `deepseek-v4-pro` 外部 worker 在本仓库已实测返回 model 404。
- 经主控确认用户已明确授权多智能体后，本次改用当前可用协作 agent 作为 Generator fallback；contract、允许路径、stop rules 和 Final Evaluator 归属均未改变。

## Summary

- `OpenAIFastPolicyRule` 新增可选 `user_ids`，并贯通 service settings、DTO、bulk settings API 和管理端表单。
- evaluator 只读取认证中间件已有的 `ctxkey.APIKeyUserID`；用户专属规则整组优先于全局规则，组内仍保持配置顺序与首条匹配语义。
- 后端拒绝零、负数和重复用户 ID；前端保留无效非空值交由后端拒绝，不会静默省略成全局规则，只有显式移除全部 ID 后才保存为全局规则。
- HTTP 证据通过 API Key auth middleware、真实 `Forward` 和 fake HTTP server capture 覆盖 managed API Key、API-key passthrough 和 OAuth passthrough。
- WebSocket 证据通过真实 `ProxyResponsesWebSocketFromClient` relay 覆盖 parsed/passthrough ingress 的首帧和后续帧，并同时断言匹配用户、非匹配用户以及伪造 body/header 身份。
- 独立 review 发现首版中英文用户范围文案误放在 `betaPolicy` namespace；已移入 `openaiFastPolicy`，并增加直接导入两份 locale 的回归断言，避免 `t()` stub 掩盖 namespace 错误。

## Changed Files

- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/admin/admin_helpers_test.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_fast_policy_test.go`
- `backend/internal/service/openai_fast_policy_ws_test.go`
- `backend/internal/server/middleware/openai_fast_policy_forwarding_test.go`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `docs/workflow/worker-results/upstream-openai-fast-flex-user-scope-s71-result.md`

## Commands Run

- Required Go test discovery: service `4/4`, middleware `1/1`, DTO `1/1`, PASS.
- Required service tests: PASS.
- Required HTTP forwarding test: PASS.
- Required DTO test: PASS.
- `go test ./internal/service -run "Test.*OpenAIFastPolicy|Test.*PassthroughBilling" -count=1`: PASS.
- Required frontend test discovery: exactly `1` test, PASS.
- Required frontend test: `1/1` PASS.
- Full `SettingsView.spec.ts`: `22/22` PASS.
- `corepack.cmd pnpm --dir frontend run typecheck`: PASS.
- `git diff --check`: PASS.
- Review fix 后重新执行 exact S71 frontend test、完整 `SettingsView.spec.ts`、typecheck 和完整 clean-worktree Acceptance Commands：PASS。

## Test Evidence

- HTTP fake upstream captured both trusted user `42` and trusted user `43` for all three modes. User `42` retained `service_tier=priority`; user `43` had it removed despite body/header values claiming user `42`.
- Parsed and passthrough WS relays each captured two upstream `response.create` frames. Both first and model-less follow-up frames used the trusted session context, with the same matching/non-matching behavior.
- Frontend test preserved loaded IDs without aliasing, submitted an invalid added `0` as a non-global scoped rule, submitted edited IDs exactly, and only omitted `user_ids` after explicit removal of all entries.
- Frontend test directly verifies all five user-scope keys exist under both en/zh `openaiFastPolicy` objects and do not exist under `betaPolicy`.

## Risks

- No live OpenAI Priority/Flex request was issued; evidence uses deterministic local fake HTTP/WS upstreams.
- Full frontend test output retains the pre-existing unresolved `router-link` warning and stale `caniuse-lite` notice; both are non-failing and unrelated to S71.
- No race test or unrelated full backend suite was run; contract-specific and adjacent regressions passed.

## Knowledge Candidates

- None. The implementation is specific to the existing Sub2API Fast/Flex policy boundary.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

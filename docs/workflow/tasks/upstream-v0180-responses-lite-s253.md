# Upstream v0.1.180 compatibility and Responses Lite S253

## Task ID

upstream-v0180-responses-lite-s253

## Role

Controller Generator / independent Evaluator. This contract executes the user-approved selective port in an isolated worktree. No worker is dispatched for this task; Codex owns the implementation and final integration decision.

## Goal

Port the compatible v0.1.180 fixes for raw Chat Completions streamed tool-call identity, OpenAI OAuth model discovery, Ops detail navigation, and account-priority visibility. Also port the complete prerequisite chain required to normalize OpenAI OAuth Responses Lite HTTP and WebSocket requests safely.

## Success Criteria

- OpenAI OAuth requests explicitly marked as Responses Lite normalize namespace tools into `input.additional_tools`, preserve compatible top-level tools, merge duplicate declarations safely, enforce `reasoning.context=all_turns`, and set `parallel_tool_calls=false` exactly when tools remain present.
- The Lite normalizer rejects malformed tool/parallel fields before upstream I/O. HTTP returns a client validation error and WebSocket returns the existing policy-violation close error. Non-Lite and API-key paths remain unchanged.
- Raw Chat Completions SSE forwarding deletes only present empty-string `tool_calls[].id` and `tool_calls[].function.name`; it preserves non-empty identity, arguments, index, type, non-tool chunks, and terminal markers.
- OpenAI OAuth account model synchronization requests the existing Codex model manifest using the local credential-resolution/auth-header conventions and accepts manifest `slug` identifiers. API-key model synchronization remains unchanged.
- Ops error detail can return to its invoking error/request list without discarding that list's visible filter/page state. Fresh account-table preferences show `priority`; an explicitly saved hidden preference remains respected.
- Every behavior has a focused regression; Go and frontend checks below pass without modifying dependencies, migrations, provider state, or the dirty primary worktree.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-v0180-responses-lite-s253`
- Base: `main@c209e5ef1e7ab2993da1105f48eedc7781eda783`
- Upstream evidence: `bc6d9c47d`, `56650d6ae`, `7498d8fdc`, `1563db3f8`, `cc894ef578`, `913ec5d74`, `cfecc8d113`, and `616df479e8`.
- S252 is concurrently building in the primary worktree. Its Pixel Cafe, Groups, Settings, workflow, knowledge, untracked asset, and `outputs/` changes are protected and must not be read as a baseline or absorbed.

## Allowed Paths

- `backend/internal/service/openai_responses_lite_tools.go`
- `backend/internal/service/openai_responses_lite_tools_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `backend/internal/service/image_generation_intent.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_http_bridge_test.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter_effort_test.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/service/openai_gateway_cc_tool_call_identity.go`
- `backend/internal/service/openai_gateway_cc_tool_call_identity_test.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/upstream_models_test.go`
- `frontend/src/views/admin/ops/OpsDashboard.vue`
- `frontend/src/views/admin/ops/components/OpsErrorDetailModal.vue`
- `frontend/src/views/admin/ops/components/OpsErrorDetailsModal.vue`
- `frontend/src/views/admin/ops/components/OpsRequestDetailsModal.vue`
- `frontend/src/views/admin/ops/components/__tests__/OpsErrorDetailNavigation.spec.ts`
- `frontend/src/views/admin/AccountsView.vue`
- `frontend/src/views/admin/__tests__/AccountsView.priorityColumn.spec.ts`
- `docs/workflow/tasks/upstream-v0180-responses-lite-s253.md`
- `docs/workflow/worker-results/upstream-v0180-responses-lite-s253-result.md`
- `docs/workflow/qa-reports/upstream-v0180-responses-lite-s253-qa.md`

## Denied Paths

- All Pixel Cafe, Groups, Settings, knowledge, untracked sprite, and `outputs/` paths in the primary worktree.
- Schema/migrations/Ent, dependencies and lockfiles, configuration, billing, account scheduling, provider credentials/traffic, Docker/Compose, deployment, shared databases, push, history rewrites, and branch/worktree cleanup.
- Any v0.1.180 item previously classified as a large product chain, including auto-reset credit, Codex affinity, service-tier/billing redesign, configurable model-list settings, OAuth transport plugins, and rejected-field retry.

## Constraints

- Adapt behavior to the current monolithic gateway topology; do not create or reference the upstream-only `openai_gateway_forward.go` request pipeline.
- Lite handling applies only to local OpenAI OAuth-like account flows and explicit Lite markers. The marker must remain whitelisted for the matching HTTP/WS upstream request; existing API-key manifest behavior, image-tool controls, custom client-tool lowering, WS replay, and S242/S243 semantics must remain intact.
- The account-model sync must reuse existing local Codex manifest constants and authentication helpers; do not invent new headers, clients, endpoints, or provider calls in tests.
- The raw SSE sanitizer is stateless and must be fail-open for invalid/non-tool payloads.
- Frontend behavior must retain saved user choices. No broad style, locale, component, or preference-storage refactor.
- Produce separately reviewable commits for: (1) Lite normalization, (2) raw identity/OAuth model sync, (3) Ops/account UI. Do not commit or push from the primary worktree.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/upstream-v0180-responses-lite-s253/backend
go test ./internal/service -run "Test(NormalizeOpenAIResponsesLite|OpenAIResponsesLite|StripEmptyChatToolCallIdentity|BuildUpstreamModelsRequestSupportsOpenAIOAuth|FetchUpstreamSupportedModelsParsesOpenAIOAuthManifest)" -count=10
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1

Set-Location E:/codex-worktrees/sub2api/upstream-v0180-responses-lite-s253
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops/components/__tests__/OpsErrorDetailNavigation.spec.ts src/views/admin/__tests__/AccountsView.priorityColumn.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
git diff --check
git ls-files -u
```

## Output

- Controller result: `docs/workflow/worker-results/upstream-v0180-responses-lite-s253-result.md` with a first-line verdict.
- Independent QA report: `docs/workflow/qa-reports/upstream-v0180-responses-lite-s253-qa.md` with a first-line verdict.
- Record changed files, executed commands, scope/provenance checks, protected-primary status, risks, and `knowledge_candidates`; do not write long-term knowledge.

## Stop Rules

- Stop and return to contract review if Lite normalization requires a missing upstream-only gateway subsystem, a migration, provider traffic, or an API contract not represented locally.
- Stop on any denied-path change, conflict with S252 ownership, two failed implementation attempts for one slice, or a validation command whose baseline cannot be isolated.
- Do not integrate any commit until focused checks, full service compile, frontend checks for the UI slice, diff review, and independent QA evidence pass.

## Budget

- controller_mode: direct isolated implementation
- qa_worker_model: `gpt-5.6-terra`
- worktree_root: `E:/codex-worktrees`

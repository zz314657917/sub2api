# Task Contract

## Task ID

upstream-grok-official-ua-s255

## Role

Developer Worker (`gpt-5.6-terra`) implements the approved behavior in this
isolated worktree. Codex Controller reviews it, then an independent QA Worker
(`gpt-5.6-terra`) reruns the gates in a separate worktree.

## Goal

Behavior-level port of upstream `9fb260439` from `upstream/main@e2d9b823f`:
use the pinned official Grok CLI User-Agent on all local Grok OAuth egress
paths. This replaces the obsolete internal `sub2api-grok/1.0` identity without
changing non-Grok custom User-Agent handling or API-key egress behavior.

## Success Criteria

- One package-local, documented official Grok OAuth User-Agent constant is
  reused by the Responses gateway, raw Chat Completions OAuth fallback, and
  account-connection test request.
- The value is the upstream-pinned `xai-grok-workspace/0.2.120`.
- Raw Chat Completions assigns this default only to Grok OAuth accounts; Grok
  API-key requests retain the caller/default transport User-Agent rather than
  being mislabeled as CLI traffic.
- The existing WebSocket-to-HTTP Grok bridge observes the same Responses
  request identity through the shared request builder.
- Runtime behavior, model routing, quota parsing, scheduling, billing,
  credentials, headers other than User-Agent, configuration, and API-key
  request semantics remain unchanged.

## Context

- Repo: `F:/mcplugins/sub2api`
- Isolated worktree: `E:/codex-worktrees/sub2api/upstream-grok-official-ua-s255`
- Base: `main@249cbc223`
- Upstream source: `9fb260439`; direct `git apply --check` fails because the
  local xAI package has no upstream CLI-identity module and local call sites
  have evolved independently.

## Allowed Paths

- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_gateway_grok_test.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_grok_s125_test.go`
- `backend/internal/service/openai_ws_http_bridge_test.go`
- `docs/workflow/worker-results/upstream-grok-official-ua-s255-result.md`
- `docs/workflow/qa-reports/upstream-grok-official-ua-s255-qa.md`

## Denied Paths

- `backend/internal/pkg/xai/**`, `frontend/**`, `knowledge/**`, `outputs/**`
- all Pixel Cafe and GroupBuy paths; schema, migrations, Ent generated files,
  dependency/configuration files, provider credentials, containers, deployment,
  push, and every path not explicitly Allowed

## Constraints

- Keep the local xAI package and its public API unchanged; the shared identity
  helper belongs in the existing service owner, not a new upstream subsystem.
- Do not add CLI token-auth/version headers: they are not part of the local
  source behavior and are outside this User-Agent-only port.
- All product edits occur only in this worktree. Preserve primary-worktree
  dirty paths byte-for-byte. No real provider, shared/production database,
  container, deployment, or push action is authorized.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "Test(BuildGrokResponsesRequestUsesOfficialCLIUserAgent|ForwardGrok.*|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge|AccountTestService_GrokOAuthUsesOfficialCLIUserAgent|ForwardAsRawChatCompletions_Grok(OAuthUsesOfficialCLIUserAgent|APIKeyDoesNotUseOfficialCLIUserAgent))" -count=10
go test ./internal/service -run "Test(BuildGrokResponsesRequest|ForwardGrok|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge|AccountTestService.*Grok|ForwardAsRawChatCompletions.*Grok)" -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
Pop-Location

gofmt -w backend/internal/service/openai_gateway_grok.go backend/internal/service/openai_gateway_chat_completions_raw.go backend/internal/service/account_test_service.go backend/internal/service/openai_gateway_grok_test.go backend/internal/service/openai_gateway_chat_completions_raw_test.go backend/internal/service/account_test_service_grok_s125_test.go backend/internal/service/openai_ws_http_bridge_test.go
git diff --check
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" backend/internal/service/openai_gateway_grok.go backend/internal/service/openai_gateway_chat_completions_raw.go backend/internal/service/account_test_service.go backend/internal/service/openai_gateway_grok_test.go backend/internal/service/openai_gateway_chat_completions_raw_test.go backend/internal/service/account_test_service_grok_s125_test.go backend/internal/service/openai_ws_http_bridge_test.go
git diff --name-only <base>..HEAD
git diff --cached --name-only
git diff --name-only
git ls-files -u
```

## Output

- One business commit limited to the product/test owners above.
- Developer result beginning `### DONE: upstream-grok-official-ua-s255`,
  `### FAILED: ...`, or `### BLOCKED: ...`, with changed paths, commands,
  upstream mapping, risks, and `knowledge_candidates`.
- Independent QA report beginning `### PASS: upstream-grok-official-ua-s255`,
  `### FAIL: ...`, or `### BLOCKED: ...`; QA may write only that report.

## Stop Rules

- Stop if implementation requires xAI package changes, CLI auth/version headers,
  an API contract, dependencies, configuration, external state, or any denied
  path.
- Stop before mainline integration if a focused/default-tag/service/server gate
  fails, if a non-Grok custom or API-key identity regresses, if conflict
  markers/unmerged index entries appear, or if the primary worktree changes.

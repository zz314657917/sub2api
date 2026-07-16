### DONE: upstream-main-compat-s77-grok

## Summary

Implemented the platform-aware image intent slice from upstream `410ea849`.
Grok passive `image_gen` namespace declarations now remain text intent when
the request does not explicitly select image generation. Native
`image_generation` tools, image models, and explicit namespace/function tool
choices remain image intent. Callers without a platform retain the existing
OpenAI declaration semantics.

## Changed Files

- `backend/internal/service/image_generation_intent.go`
- `backend/internal/service/image_generation_intent_grok_test.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_test.go`
- `backend/internal/service/openai_gateway_service.go`

The WS-owned files (`openai_gateway_handler.go` and
`openai_ws_forwarder.go`) were intentionally not modified. Root must apply the
same platform argument at their Responses WS ingress/permission call sites in
the approved sequential integration phase.

## Verification

All commands were run in this isolated worktree with
`GOTMPDIR=F:/mcplugins/sub2api/.tmp/go-build` where applicable:

- `go test ./internal/service -run 'ImageGenerationIntent|ImageIntent' -count=1` - PASS
- `go test ./internal/server/routes -run 'ImageIntent' -count=1 -v` - PASS
- `go test ./internal/config ./internal/handler ./internal/service ./internal/server/routes -run 'OpenAIWS|ImageGenerationIntent|ImageIntent' -count=1` - PASS
- `git diff --check` - PASS

The tests cover top-level and Responses Lite passive namespace declarations,
Chat Completions classification, native image tools, image models, explicit
tool choices, OpenAI legacy behavior, and image-only versus text-only route
selection.

## Scope and Risks

- No Ent, migration, billing schema, payment, scheduler policy, deployment,
  VERSION, container, knowledge, or WS-owned path was changed.
- The service call-site updates use the effective account platform and are a
  no-op for non-Grok callers because the legacy classifier remains the
  fallback.
- No authenticated or live upstream Grok smoke was run; verification is
  unit/route-level only.
- The user-authorized inherited-model worker fallback was used; no push or
  deployment was performed.

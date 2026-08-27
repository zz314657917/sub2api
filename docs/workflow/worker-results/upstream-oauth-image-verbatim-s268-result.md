### DONE: upstream-oauth-image-verbatim-s268

## Task ID
upstream-oauth-image-verbatim-s268

## Status
`done`

## Summary

Manually ported the prompt-preservation behavior from upstream `329b92ef0`.
The OAuth image Responses request now carries one stable top-level instruction
requiring verbatim image prompts, while the existing input builder continues
to forward the user prompt unchanged.

## Changed Files

- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_images_test.go`
- `docs/workflow/tasks/upstream-oauth-image-verbatim-s268.md`

## Commands Run

```text
go test ./internal/service -run '^TestBuildOpenAIImagesResponsesRequest_(StripsInputFidelity|RequiresVerbatimUserPrompt)$' -count=10 -> PASS
go test ./internal/service -run '^$' -count=1 -> PASS (compile-only)
go test ./cmd/server -run '^$' -count=1 -> PASS (compile-only)
gofmt -w changed Go files -> PASS
git diff --check -> PASS
git ls-files -u -> empty
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 30.096s
ok github.com/Wei-Shaw/sub2api/internal/service 25.880s [no tests to run]
ok github.com/Wei-Shaw/sub2api/cmd/server 32.365s [no tests to run]
```

## Risks

- No live OpenAI OAuth provider call was made; provider interpretation of the
  instruction remains an external runtime concern.
- No database, container, deployment, staging, or push operation was made.

## Knowledge Candidates

- The OAuth image Responses payload can preserve the client prompt by adding a
  stable instruction at the request builder boundary without changing input
  or tool parameter handling.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Upstream Provenance

- `329b92ef045fd24b49b33e719e42facc7b26e1b2` (`fix(openai): preserve OAuth image prompts verbatim`)

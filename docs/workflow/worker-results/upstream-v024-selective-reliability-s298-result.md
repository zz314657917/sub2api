### DONE: upstream-v024-selective-reliability-s298

# Worker Result

## Task ID
upstream-v024-selective-reliability-s298

## Status
`done`

## Summary
- Required Terra Generator dispatch stopped before execution with `403 No available group route matches the requested model or request type`, zero input/output tokens, and no worktree diff.
- Controller takeover then implemented the already approved contract in the primary worktree: user Usage API-key filters load every page, and native/passthrough Responses streams record `stream_failed` operations diagnostics only after output has committed.

## Changed Files
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`

## Commands Run
```text
gofmt -w backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_test.go -> PASS
git diff --check -- <four allowed code/test paths> -> PASS
go test ./internal/service -run 'TestOpenAIStreaming(ResponseFailedAfterOutputSanitizesVerboseResponseForClient|PassthroughResponseFailedAfterOutputSanitizesVerboseResponseForClient)$' -count=1 -> PASS
npm.cmd exec vitest run src/views/user/__tests__/UsageView.spec.ts -> PASS (31 tests)
go build ./... -> PASS
npm.cmd run typecheck -> PASS
```

## Risks
- Independent Terra QA did not run because the configured route is unavailable; see the S298 QA report.
- No provider, database, container, browser, deployment, commit or push action was performed.

## Knowledge Candidates
- none

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `yes: Terra worker route unavailable; controller takeover only for approved paths`

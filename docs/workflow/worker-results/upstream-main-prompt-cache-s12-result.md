### DONE: upstream-main-prompt-cache-s12

## Summary

- Created branch `codex/upstream-main-prompt-cache-s12` from local `main@c1cb19951`.
- Reviewed upstream `main=b7cfe2462` after fetching `v0.1.135` and post-release commits.
- Ported the approved Chat Completions prompt cache candidate without directly merging `upstream/main`.
- Kept the implementation inside approved backend service and workflow paths.

## Candidate Results

- `d251487da`: `CHERRY_PICKED` as `69e2d54a8`. Propagates non-empty `prompt_cache_key` into the Responses body for OpenAI API Key accounts, preserves existing explicit body values, and isolates generated `session_id` by current API Key ID.

## Deferred / Skipped

- `16bc87693`: `DEFERRED`. 5h `ResetsAt` sync touches backend API contract plus frontend/i18n and should be a separate Sprint.
- `f20e6bf76`: `DEFERRED`. `account_temp_unscheduled_count` alert metric touches backend and frontend ops UI; best handled as an ops Sprint.
- `f5cecea5b`: `DEFERRED`. Pure frontend Select dropdown height fix; keep separate from backend gateway work.
- README/VERSION-only upstream commits: `SKIPPED`.

## Commits

- `bccf633d5` docs: add prompt cache s12 contract
- `69e2d54a8` fix(openai): propagate prompt cache key for chat completions

## Changed Files

- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `docs/workflow/tasks/upstream-main-prompt-cache-s12.md`
- `docs/workflow/worker-results/upstream-main-prompt-cache-s12-result.md`
- `docs/workflow/qa-reports/upstream-main-prompt-cache-s12-qa.md`
- `docs/workflow/main-log.md`

## Verification

- `git status --short --branch`
- `git diff --check main..HEAD`
- denied path audit against `main..HEAD`
- `go test ./internal/service -run "ForwardAsChatCompletions|PromptCache|Session|OpenAI|ChatCompletions" -count=1`
- `go test ./internal/service -count=1`

## Notes

- No `frontend/`, `backend/ent/`, `backend/migrations/`, `skills/`, `deploy/`, `knowledge/`, `docs/workflow/status.md`, or `docs/workflow/spec.md` changes were made.
- Local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, and workflow docs were preserved.

### PASS: upstream-main-gateway-compat-s4

## Findings
- PASS: All changed paths are within the Sprint allowed path set when audited against fixed baseline `34d02457b`.
- PASS: No `frontend/`, `backend/ent/`, `backend/migrations/`, `deploy/`, `knowledge/`, `docs/workflow/status.md`, `docs/workflow/spec.md`, `.github/`, `assets/`, or `README*` changes are present in `34d02457b..HEAD`.
- PASS: All five candidate fixes were ported. `9b99f6c1f` was implemented as an equivalent minimal port to avoid importing a broader stream lifecycle refactor.

## Executed Checks
- `git status --short --branch` -> clean on `codex/upstream-main-gateway-compat-s4`.
- `git diff --check` -> PASS.
- `git diff --name-status 34d02457b..HEAD` -> only allowed backend gateway/apicompat/image files and workflow artifacts.
- `git diff --name-only 34d02457b..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|deploy/|knowledge/|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"` -> no matches.
- `go test ./internal/pkg/apicompat -run "Responses|ChatCompletions|Anthropic|Tool|DeepSeek|Reasoning" -count=1` -> PASS.
- `go test ./internal/service -run "OpenAIImages|ChatCompletions|Responses|Failed|Tool|DeepSeek" -count=1` -> PASS.
- `go test ./internal/handler -run "OpenAI|Gateway|Images|Failed" -count=1` -> PASS.
- `go test ./internal/pkg/apicompat ./internal/service ./internal/handler -count=1` -> PASS.

## Not Run
- `go test ./internal/server/routes ./cmd/server -count=1` was not run because this Sprint did not touch route/server wiring.

## Risks
- The `DeepSeek reasoning-only` fix is intentionally a minimal equivalent port for the current local stream bridge, not a full import of upstream's broader lifecycle state machine. This keeps Sprint scope controlled but means future gateway lifecycle parity should be handled in a separate Sprint if needed.

## Recommendation
PASS. The branch is ready for review/merge into the intended integration target.

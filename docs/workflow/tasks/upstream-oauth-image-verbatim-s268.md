---
task_id: upstream-oauth-image-verbatim-s268
phase: contract-approved
qa_mode: runtime
---

# Upstream OAuth Image Prompt Verbatim S268

## Role

Primary Codex performs the bounded Planner/Generator/Evaluator flow in this
isolated worktree. This is a manual behavior port, not a broad upstream merge.

## Goal

Port the behavior from upstream `329b92ef0`: OpenAI OAuth image Responses
requests must explicitly instruct the image tool to use the user's prompt
verbatim, while retaining the original prompt bytes in the input message.

## Success Criteria

- The generated OAuth image Responses payload contains the stable verbatim
  instruction in its top-level `instructions` field.
- The user prompt remains unchanged in `input[0].content[0].text`, including
  non-ASCII text, capitalization, quotes, punctuation, and constraints.
- Existing image edit inputs, tool parameters, model selection, streaming
  behavior, and non-OAuth/APIMart image paths remain unchanged.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen base: `main@cd42eebf1`
- Upstream behavior source: `329b92ef045fd24b49b33e719e42facc7b26e1b2`
- Worker worktree: `E:/codex-worktrees/sub2api/upstream-oauth-image-verbatim-s268`

## Allowed Paths

- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_images_test.go`
- `docs/workflow/tasks/upstream-oauth-image-verbatim-s268.md`
- `docs/workflow/worker-results/upstream-oauth-image-verbatim-s268-result.md`
- `docs/workflow/qa-reports/upstream-oauth-image-verbatim-s268-qa.md`

## Denied Paths

- `F:/mcplugins/sub2api` primary worktree and all Pixel Cafe/user dirty files
- Codex transform/instructions, gateway routing, account scheduling, billing,
  handlers, routes, frontend, schema/migrations, dependencies, providers,
  containers, deployment, databases, `knowledge/**`, `outputs/**`, staging,
  and push

## Constraints

- Reuse the existing `sjson` request builder and prompt field; do not add a
  second prompt normalization or translation path.
- Keep the instruction text as one stable local constant and do not include
  user data, credentials, or account identifiers in it.
- Do not alter input trimming beyond the existing builder behavior; the test
  must assert the current prompt contract rather than introduce a new one.

## Acceptance Commands

```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Set-Location backend
go test ./internal/service -run '^TestBuildOpenAIImagesResponsesRequest_(StripsInputFidelity|RequiresVerbatimUserPrompt)$' -count=10
go test ./internal/service -run '^$' -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -d internal/service/openai_images.go internal/service/openai_images_responses.go internal/service/openai_images_test.go
Set-Location ..
git diff --check -- backend/internal/service/openai_images.go backend/internal/service/openai_images_responses.go backend/internal/service/openai_images_test.go docs/workflow/tasks/upstream-oauth-image-verbatim-s268.md
git ls-files -u
```

## Stop Rules

- Stop if preserving the prompt requires changes to shared Codex instruction
  transforms, routing, billing, or any denied path.
- Stop before touching the primary worktree or any user-owned dirty/untracked
  file.
- Stop if the focused regression cannot run without a pre-existing unrelated
  package failure; report the baseline separately.

## Output

- Commit one scoped implementation commit and one worker-result report commit.
- Independent Terra QA must write a separate report before integration.
- Final evaluation must include exact scope, commands, provenance, and
  residual provider/runtime risk.

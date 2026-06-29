---
task_id: upstream-main-v0139-codex-json-input-s25
phase: contract-approved
owner: Codex
qa_mode: runtime
created_at: 2026-06-29 22:31 +08:00
---

# Task Contract: upstream v0.1.139 Codex JSON input S25

## Goal

Port upstream `df51edfb0` / `b105cc0fd` Codex OAuth transform behavior so system guidance remains visible in Responses `input` while still satisfying the ChatGPT internal Codex endpoint's `instructions` requirement.

## Success Criteria

- `role:"system"` items in Codex OAuth `input` are converted to `role:"developer"` instead of being removed.
- Extracted system text is still mirrored into `reqBody["instructions"]`, prepended before existing non-empty instructions.
- Responses JSON mode requests keep JSON-only guidance inside `input`, preventing upstream JSON mode rejection when the user message lacks the JSON keyword.
- Targeted service tests and `git diff --check` pass.
- Staged denied-path audit excludes existing dirty `knowledge/*`, frontend, generated/schema, migrations, deploy, README, and unrelated upstream batches.

## Allowed Paths

- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `docs/workflow/tasks/upstream-main-v0139-codex-json-input-s25.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/worker-results/upstream-main-v0139-codex-json-input-s25-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-codex-json-input-s25-qa.md`

## Denied Paths

- Existing dirty `knowledge/*` files, unless separately requested.
- `frontend/*`, public pages, Studio/Canvas, Model Plaza, payment pages, Ops/Keys UI.
- `backend/ent/*`, `backend/migrations/*`, wire generation, VERSION, README, sponsors, CI/deploy.
- OpenAI count_tokens bridge from `7a38c6621`; keep for a later Sprint because it adds handler, service, route, and endpoint surface.
- Subscription/payment/order display, API base fetch, keys column settings, Ops system log key id, openai quota headroom scheduler, and other product/UI batches.

## Constraints

- Do not cherry-pick wholesale; adapt only the local transform behavior and nearby tests.
- Do not change Codex tool normalization, image bridge, Spark image restrictions, billing, routing, or auth behavior.
- Preserve existing `instructions` fallback behavior when no system/developer guidance is present.
- Do not stage or commit unrelated dirty files.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestExtractSystemMessagesFromInput|TestApplyCodexOAuthTransform_ExtractsSystemMessages|TestApplyCodexOAuthTransform_JsonObjectKeepsJsonInstructionInInput" -count=1
go test ./internal/service -run "TestApplyCodexOAuthTransform_.*Instructions|TestExtractSystemMessagesFromInput" -count=1

cd F:/mcplugins/sub2api
git diff --check
git diff --cached --name-only | rg "^(knowledge/|frontend/|backend/ent/|backend/migrations/|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|deploy/|README|README_|assets/partners/)" || echo NO_DENIED_PATHS
```

## Output

- Code diff in allowed backend and workflow paths only.
- Worker-style result and QA report.
- Final report with commit hash, pushed status, tests run, and remaining dirty file note.

## Stop Rules

- Stop if the patch requires adding the OpenAI count_tokens bridge, route changes, or endpoint constants.
- Stop if the behavior would remove local Codex compatibility transformations unrelated to system/developer input.
- Stop if tests require real external OpenAI/OAuth upstreams.

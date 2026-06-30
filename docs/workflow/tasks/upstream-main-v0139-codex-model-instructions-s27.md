---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-06-30 13:18 +08:00
---

# Task Contract

## Task ID
upstream-main-v0139-codex-model-instructions-s27

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small follow-up port. No external worker is used.

## Goal
Port the low-risk OpenAI/Codex instructions fix from upstream so empty Codex OAuth `instructions` are synthesized from the model-specific Codex base prompt instead of the generic placeholder. Keep the change limited to backend OpenAI instruction selection and regression tests.

## Success Criteria
- `gpt-5.5` empty/blank Codex OAuth `instructions` receives the GPT-5.5 Codex base prompt.
- `gpt-5.1`, `gpt-5.2`, codex-family models, and unknown GPT-5 models choose the expected fallback prompt family.
- Existing non-empty `instructions` remain preserved.
- No frontend, migrations, Ent, wire, deploy, VERSION, README, or `knowledge/*` files are included in the commit.
- Targeted Go tests and diff checks pass.

## Context
- Repo: `F:/mcplugins/sub2api`
- Upstream reference: `709cf6185 修复 OpenAI GPT-5.5 的 Codex 指令选择`
- Supporting upstream references: `5e6effd79`, `00d68ff6d` for model-aware prompt resources.
- Already equivalent/skipped before this sprint: `27600b1d2c`, `1d47fd6300`, `2c14efeaa0`, `888cd8092d`, `32ea9cfe1f`, `89dffdd2e1`, `6aec505016`, `be3613593b`, `c10598dfe5`.

## Allowed Paths
- `backend/internal/pkg/openai/constants.go`
- `backend/internal/pkg/openai/instructions_gpt5_1.txt`
- `backend/internal/pkg/openai/instructions_gpt5_2.txt`
- `backend/internal/pkg/openai/instructions_gpt5_5.txt`
- `backend/internal/pkg/openai/instructions_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `docs/workflow/tasks/upstream-main-v0139-codex-model-instructions-s27.md`
- `docs/workflow/worker-results/upstream-main-v0139-codex-model-instructions-s27-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-codex-model-instructions-s27-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `knowledge/**`
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/service/wire.go`
- `deploy/**`
- `README*`
- `assets/partners/**`
- Payment, subscription, keys UI, ops UI, Grok routing, risk-control, OAuth email flow, and production configuration paths.

## Constraints
- Do not merge or rebase `upstream/main`.
- Keep this sprint to model-aware Codex default instructions only.
- Use upstream prompt resources directly; do not hand-edit prompt text.
- Do not modify or stage existing dirty `knowledge/*` files.
- Preserve local Studio Bridge, payment, Canvas, public page, and gateway customizations.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/pkg/openai -run "TestCodexBaseInstructionsForModel" -count=1
go test ./internal/service -run "TestDefaultCodexSynthInstructionsModelAware|TestApplyCodexOAuthTransform_GPT55SuppliesModelSpecificInstructions|TestApplyCodexOAuthTransform_CodexCLI_SuppliesDefaultWhenEmpty|TestApplyCodexOAuthTransform_NonCodexCLI_PreservesExistingInstructions" -count=1
go test ./internal/service -run "TestOpenAIGatewayServiceForwardGPT55InjectsModelSpecificInstructions" -count=1
cd F:/mcplugins/sub2api
git diff --check
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0139-codex-model-instructions-s27-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0139-codex-model-instructions-s27-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementing this requires frontend, migration, Ent, wire, deploy, README, or `knowledge/*` changes.
- Stop if tests require live OpenAI OAuth credentials.
- Stop if the local transform semantics conflict with existing system-message preservation from S25.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`

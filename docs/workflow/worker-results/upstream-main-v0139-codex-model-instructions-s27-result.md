### DONE: upstream-main-v0139-codex-model-instructions-s27

# Worker Result

## Task ID
upstream-main-v0139-codex-model-instructions-s27

## Status
done

## Summary
- Ported the upstream model-aware Codex default instructions behavior for empty OpenAI OAuth `instructions`.
- Added embedded GPT-5.1, GPT-5.2, and GPT-5.5 Codex base prompt resources and `CodexBaseInstructionsForModel`.
- Updated both the Codex OAuth transform helper and the `Forward` hot path so `gpt-5.5` blank instructions use the GPT-5.5 Codex base prompt instead of the generic placeholder.

## Changed Files
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

## Commands Run
```text
gofmt -w backend/internal/pkg/openai/constants.go backend/internal/pkg/openai/instructions_test.go backend/internal/service/openai_codex_transform.go backend/internal/service/openai_codex_transform_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_test.go -> pass
go test ./internal/pkg/openai -run "TestCodexBaseInstructionsForModel" -count=1 -> pass
go test ./internal/service -run "TestDefaultCodexSynthInstructionsModelAware|TestApplyCodexOAuthTransform_GPT55SuppliesModelSpecificInstructions|TestApplyCodexOAuthTransform_CodexCLI_SuppliesDefaultWhenEmpty|TestApplyCodexOAuthTransform_NonCodexCLI_PreservesExistingInstructions|TestOpenAIGatewayServiceForwardGPT55InjectsModelSpecificInstructions" -count=1 -> pass
git diff --check -> pass
```

## Test Output
```text
ok  	github.com/Wei-Shaw/sub2api/internal/pkg/openai	0.044s
ok  	github.com/Wei-Shaw/sub2api/internal/service	5.695s
git diff --check -> only existing CRLF warnings for dirty knowledge files; no whitespace errors
```

## Risks
- The large prompt resources are copied from `upstream/main`; future upstream prompt changes still require a separate sync.
- Live OpenAI OAuth was not exercised; verification is unit-level and request-body-level.
- Existing unrelated dirty files remain outside this sprint.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

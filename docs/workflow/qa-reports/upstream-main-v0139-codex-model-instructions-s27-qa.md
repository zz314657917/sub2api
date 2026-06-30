### PASS: upstream-main-v0139-codex-model-instructions-s27

# QA Report

## Task ID
upstream-main-v0139-codex-model-instructions-s27

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-v0139-codex-model-instructions-s27.md`

## Evidence
- diff reviewed: yes
- allowed paths checked: yes
- denied paths touched: no
- commands run:
```text
go test ./internal/pkg/openai -run "TestCodexBaseInstructionsForModel" -count=1 -> pass
go test ./internal/service -run "TestDefaultCodexSynthInstructionsModelAware|TestApplyCodexOAuthTransform_GPT55SuppliesModelSpecificInstructions|TestApplyCodexOAuthTransform_CodexCLI_SuppliesDefaultWhenEmpty|TestApplyCodexOAuthTransform_NonCodexCLI_PreservesExistingInstructions|TestOpenAIGatewayServiceForwardGPT55InjectsModelSpecificInstructions" -count=1 -> pass
git diff --check -> pass
```
- manual checks:
```text
Reviewed S27 diff paths against contract allowed paths -> pass
Confirmed existing dirty knowledge/OAuth/frontend files are outside S27 staging scope -> pass
Confirmed upstream helper behavior for " GPT-5.5 " and fallback models is covered by TestCodexBaseInstructionsForModel -> pass
```

## Findings
- 未发现明确问题。

## Bug Owner Recommendation
codex-planner

## Root Cause
none

## Retest Scope
- If this area changes again, rerun `TestCodexBaseInstructionsForModel`, `TestDefaultCodexSynthInstructionsModelAware`, and `TestOpenAIGatewayServiceForwardGPT55InjectsModelSpecificInstructions`.

## Knowledge Promotion
none

### PASS: upstream-main-v0139-codex-json-input-s25

# QA Report

## Findings

- PASS: system messages are preserved in `input` as developer messages instead of being removed.
- PASS: extracted system text is still mirrored into `instructions`, including existing-instructions prepend behavior.
- PASS: JSON mode regression confirms JSON guidance stays in the input message surface.
- PASS: `git diff --check` returned no whitespace errors; only existing `knowledge/*` CRLF warnings were printed.

## Executed Checks

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestExtractSystemMessagesFromInput|TestApplyCodexOAuthTransform_ExtractsSystemMessages|TestApplyCodexOAuthTransform_JsonObjectKeepsJsonInstructionInInput" -count=1
go test ./internal/service -run "TestApplyCodexOAuthTransform_.*Instructions|TestExtractSystemMessagesFromInput" -count=1

cd F:/mcplugins/sub2api
git diff --check
```

## Unverified Risks

- No live Codex OAuth request was sent to upstream.
- OpenAI count_tokens bridge remains unmerged and untested in this Sprint by design.

## Recommendation

PASS for S25 as a scoped backend compatibility port. Commit only S25 allowed paths and keep existing `knowledge/*` dirty files unstaged.

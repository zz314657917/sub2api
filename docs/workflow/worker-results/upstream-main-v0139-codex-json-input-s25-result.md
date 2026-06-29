### DONE: upstream-main-v0139-codex-json-input-s25

# Worker Result

## Summary

- Ported upstream `df51edfb0` / `b105cc0fd` as a local Codex OAuth transform adaptation.
- `role:"system"` input messages are now converted to `role:"developer"` and preserved in `input`.
- System guidance text is still mirrored into `instructions`, including prepending before existing non-empty instructions.
- Added a JSON mode regression test to ensure JSON-only guidance remains visible in Responses `input`.

## Changed Files

- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `docs/workflow/tasks/upstream-main-v0139-codex-json-input-s25.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Notes

- This Sprint intentionally excludes upstream `7a38c6621` OpenAI count_tokens bridge because it adds new handler, service, route, endpoint URL, and tests. It should remain a separate S26 contract.
- No billing, routing, Spark image, image bridge, auth, frontend, migration, or wire-generation behavior was changed.

## Commands Run

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestExtractSystemMessagesFromInput|TestApplyCodexOAuthTransform_ExtractsSystemMessages|TestApplyCodexOAuthTransform_JsonObjectKeepsJsonInstructionInInput" -count=1
go test ./internal/service -run "TestApplyCodexOAuthTransform_.*Instructions|TestExtractSystemMessagesFromInput" -count=1

cd F:/mcplugins/sub2api
git diff --check
```

## Risks

- No real OpenAI/Codex OAuth upstream smoke was run; validation is by local transform tests.
- The change mutates existing map items in `input` from `system` to `developer`, matching the previous in-place transform style used in this file.

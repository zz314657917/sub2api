### DONE: upstream-main-v0143-codex-compact-skip-image-bridge-s49

## Summary

- Ported upstream `c797159bf` into the isolated S49 worktree.
- Precomputed `isCompactRequest` before Codex image-generation bridge injection in `OpenAIGatewayService.Forward`.
- Skipped Codex image bridge tool, `tool_choice`, and bridge instruction injection for `/responses/compact` requests.
- Added regression coverage proving compact requests do not receive image bridge fields while existing non-compact bridge behavior remains covered by the existing tests.

## Changed Files

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `docs/workflow/tasks/upstream-main-v0143-codex-compact-skip-image-bridge-s49.md`
- `docs/workflow/worker-results/upstream-main-v0143-codex-compact-skip-image-bridge-s49-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-codex-compact-skip-image-bridge-s49-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run

```powershell
gofmt -w backend/internal/service/openai_gateway_service.go backend/internal/service/openai_image_generation_controls_test.go
go test ./internal/service -run "TestOpenAIGatewayServiceForward_CodexBridge|TestOpenAIGatewayServiceForward_.*Image|TestOpenAIGatewayService_CodexImageGenerationBridge" -count=1
git diff --check -- backend/internal/service/openai_gateway_service.go backend/internal/service/openai_image_generation_controls_test.go docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0143-codex-compact-skip-image-bridge-s49.md
```

## Test Output

- `internal/service`: PASS.
- `git diff --check`: PASS.

## Risks

- The local `Forward` layout differs from upstream: bridge injection happens before the old local compact mapping block. The implementation intentionally only precomputes the compact flag and guards bridge injection, without reordering account selection, model mapping, billing, or OAuth transform logic.
- Explicit image-generation tools on compact requests are still normalized by the existing generic normalization path if present in the client payload; this Sprint only prevents Codex bridge injection into compact requests.

## Knowledge Candidates

- None.

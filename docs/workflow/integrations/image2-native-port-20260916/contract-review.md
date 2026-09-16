### PASS: image2-native-port

# Contract Review

## Contract Checked

- `docs/workflow/tasks/image2-native-port.md`
- `docs/workflow/plans/image2-selective-integration-20260915.md`
- Baseline: `bb3dde9e410c41bde0c4f4dce17e7f59a44c86a5`.

## Findings

No blocking contract defect found after revision.

- The plan is present, explicitly limits the work to I-03/I-04, retains the experimental adapter as non-compiling reference only, and excludes MiniMax, database, deployment, and wholesale upstream integration.
- Criteria 7 and 8 now make the prior ambiguous public-output and billing claims independently falsifiable: separate JSON/SSE assertions cover upstream/public model separation, requested URL/base64 format, edit event naming, and no retry after partial output; the usage test must assert cached-image metadata, exact one-time charging, dedupe, and caller-map immutability.
- The required PowerShell gate was executed from the worktree root against `bb3dde9e4`. It enumerated `docs/workflow/status.md` only (existing controller workflow state), reported no denied business path, no formatting issue, a clean diff check, and no unmerged index. The broad workflow-document exclusion is expressly limited to controller/review evidence; Generator remains bound to the three specifically allowed task/result/QA document paths and manual hunk review.
- The focused backend selector was independently executed from `backend` and passed: `go test ./internal/service -run 'Test.*(OpenAIImages|CodexDirectImages|ImageOutput|ImageCache|RecordUsage)' -count=1` (`ok`, 0.178s). It selects the required existing image/output/usage families; new native tests must use the contracted `TestCodexDirectImages` or `TestOpenAIImages` prefix. No `unit` tag is presently needed.
- Source ownership matches the contract: `openai_gateway_service.go` persists `OpenAIForwardResult` image input/output usage and costs; no `gateway_service.go`/Claude owner change is authorized. Account-test parity remains an explicitly deferred, denied follow-up.

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes` — `gpt-5.6-terra`.
- base_commit_confirmed: `yes`
- openspec_traceable: `yes`

## Approval

PASS. The contract is approved for Terra Generator dispatch from baseline `bb3dde9e4`. Generator must retain the declared allowlist, execute the required path/format/conflict gate, and report only local fake-transport/mock evidence; real-provider, account-test, database, deployment, commit, and push acceptance remain out of scope.

---

## Addendum Review: existing Responses fixtures and forwarding selector

### PASS: image2-native-port

The small contract amendment is approved.

- `backend/internal/service/openai_images_test.go` is now explicitly allowed only for existing Responses-protocol fixtures, while the existing `openai_images_*_test.go` glob remains unchanged. This correctly acknowledges that the glob does not match the exact filename.
- The current diff is limited to adding `withOpenAIImagesForceResponses(context.Background())` to eleven existing OAuth/Responses fixtures, plus one explanatory comment. Assertions, API-key coverage, APIMart coverage, production code, and unrelated fixtures are untouched. The force context preserves their stated Responses-transform subject instead of accidentally routing them through the new native default.
- The contract now requires `go test ./internal/service -run 'TestOpenAIGatewayServiceForwardImages' -count=1`; this is necessary because the primary selector does not select that family. It is a new required gate, not retroactive evidence that the earlier selector covered it.
- `git diff --check` over the amended contract and fixture file is clean. No build or test was run in this review because the active Generator is concurrently changing the same implementation/test area.

Final QA must execute both focused selectors after the Generator stops and must review this file under the new exact-path allowlist.

---
status: approved
review_verdict: PASS
task_id: image2-account-parity
worker_model: gpt-5.6-terra
base_commit: dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779
spec_ref: docs/workflow/integrations/image2-native-port-20260916/plan.md
---

## Task ID
image2-account-parity

## Role
Terra Developer, independent Terra contract review and QA; controller owns integration.

## Goal
Complete the remaining account-test portion of I-03/I-04 on the accepted native
Images foundation. Account tests must use the same model selection, native
request builder and native result validation as forwarding while retaining the
existing admin test SSE protocol and credential/proxy behavior.

## Success Criteria
1. Account mapped supported native models use Codex images/generations, with
   Accept application/json; legacy/unknown supported image snapshots keep the
   configured Responses driver. Public test_start model remains requested model,
   captured upstream payload uses the mapped model. Preserve supplied prompt.
2. Native JSON returns correct MIME data URL (item output_format, then root,
   then default), revised prompt, test_start/content/image/test_complete events;
   malformed/empty/error/native transport failures never emit success:true.
3. Native HTTP 404/405 falls back to Responses at most once, closing each body;
   a second 404/405 terminates, other errors do not fallback. Preserve errors and
   existing credential redaction; no credentials in downstream error messages.
4. OAuth/SetupToken/Agent Identity/shadow credential resolution, account ID,
   proxy, custom/default User-Agent, identity/ChatGPT-account headers remain as
   before for each supported route. Fake transport/header tests prove applicable
   branches; no real credential, identity recovery or provider request.
5. Extract a shared native JSON result parser in openai_images_direct.go and
   use it in both the nonstream native forwarder and account tests. Preserve
   native forwarding's public format/model/dimensions/usage and stream behavior;
   run the accepted native regression suite. No independent duplicate parser.
6. Preserve main-tree Pelican prompt/error-sanitization hunks entirely. Only the
   testOpenAIImageOAuth function/comment may change in account_test_service.go.
   Controller applies/stages only this target hunk, never replaces/stages the
   entire dirty main file. Workers never modify main or other worktrees.
7. Existing account Responses fixtures retain assertions and explicitly use
   legacy models or genuine 404/405 fallback to exercise Responses; native model
   assertions must not be weakened simply to obtain green tests. APIKey tests
   remain unchanged. New tests call production testOpenAIImageOAuth or the
   existing account-test entrypoint, and assert endpoint, payload, headers,
   body-close count, request count and downstream event sequence.

## Allowed Paths
- backend/internal/service/account_test_service.go (testOpenAIImageOAuth function and its comment only)
- backend/internal/service/account_test_service_openai_image_test.go (OAuth Responses fixtures only)
- backend/internal/service/account_test_service_openai_native_test.go
- backend/internal/service/openai_images_direct.go (shared JSON parser extraction/use only; no streaming changes)
- backend/internal/service/openai_images_native_test.go (shared parser regressions only)
- docs/workflow/tasks/image2-account-parity.md
- docs/workflow/worker-results/image2-account-parity-result.md
- docs/workflow/qa-reports/image2-account-parity-qa.md

## Denied Paths
All other business paths, generic account prompt/error helpers, Pelican, handler,
wire, frontend, auth/recovery architecture, billing, config, schema/migrations,
dependencies, secrets, real providers/databases, containers, deployment, push,
worker commits, merge/rebase/cherry-pick and all other worktrees.

## Constraints
Use apply_patch, preserve local topology and existing credentials semantics.
Native foundation is committed at dc851ea3f, not an unverified experimental file.
Only behavior-preserving parser extraction is newly authorized in that adapter.
Main account-test dirty changes were inspected before dispatch and belong to
another task. Real-provider acceptance remains a separate open Epic requirement.

## Acceptance Commands
From backend:
`go test ./internal/service -run 'TestAccountTestService_OpenAIImage|TestCodexDirectImages|TestOpenAIGatewayServiceForwardImages' -count=1`
`go test ./internal/service -count=1`
`go build ./...`

From worktree root (PowerShell):
```powershell
$parityPaths = @((git diff --name-only dc851ea3f); (git ls-files --others --exclude-standard)) | Sort-Object -Unique
$parityAllowed = '^backend/internal/service/(account_test_service\.go|account_test_service_openai_image_test\.go|account_test_service_openai_native_test\.go|openai_images_direct\.go|openai_images_native_test\.go)$'
$parityDenied = @($parityPaths | Where-Object { $_ -notmatch '^docs/workflow/' -and $_ -notmatch $parityAllowed })
if ($parityDenied.Count) { throw ('Denied paths: ' + ($parityDenied -join ', ')) }
$parityGo = @($parityPaths | Where-Object { $_ -match '\.go$' })
if ($parityGo.Count) { $bad = @(gofmt -l $parityGo); if ($LASTEXITCODE -ne 0 -or $bad.Count) { throw 'gofmt failed' } }
git diff --check dc851ea3f
if ($LASTEXITCODE -ne 0) { throw 'diff check failed' }
if (@(git ls-files -u).Count) { throw 'unmerged index' }
```
Additionally review target hunks manually and enumerate ignored task evidence
with git status --short --ignored docs/workflow. Controller evidence exceptions
do not broaden worker documentation permissions.

## Output
Worker report starts ### DONE/FAILED/BLOCKED: image2-account-parity with changed
paths, exact commands, criterion evidence and gaps. Independent QA begins
### PASS/FAIL/BLOCKED: image2-account-parity, including parser equivalence and
main-dirty preservation requirements for final integration. No runtime claims.

## Stop Rules
Baseline exception amendment approved by independent re-review, exhaustive list:
- TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation:
  claude_code_validator_test.go:52, Should be false.
- TestApplyCodexOAuthTransform_PreservesAllowedTools:
  openai_codex_transform_test.go:18, Should be false.
- TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested:
  openai_codex_transform_test.go:275, expected
  fc_6336c46cf80eccddb18637bafbdab3b1de6137977fd8f941d2ccc9763b32d, actual
  fc_477489beb020a16331ccd2917b0310e7bb315418ee7f92d3781a5e28eb761.
- TestParseSSEUsage_SelectiveParsing:
  openai_gateway_service_test.go:3259, expected 9, actual 1.

Evidence: docs/workflow/baseline-validator-exception.md. Run the unfiltered
full service suite and preserve/report original exit FAIL. Additionally enumerate
every go test -json Action=fail event completely, with no truncated output.
Only these exact baseline fingerprints may remain for local scoped integration
after independent approval. Any extra failure or changed fingerprint blocks.
Every account identity/UA/shadow/proxy/header or new candidate failure remains
blocking. No assertion weakening, validator fix, full-suite green or release
approval is authorized.

No dispatch before independent contract PASS and approved frontmatter. Stop if
credential semantics or another denied owner must change; give exact evidence.
Do not weaken tests or omit identity/shadow branches to claim completion.

### PASS: image2-account-parity

# Contract Review

## Contract Checked

- `docs/workflow/tasks/image2-account-parity.md`
- `docs/workflow/integrations/image2-native-port-20260916/plan.md`
- Accepted native contract/review: `docs/workflow/integrations/image2-native-port-20260916/contract.md` and `contract-review.md`
- Baseline: `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779`

## Findings

No blocking contract defect found.

- The task is a bounded parity follow-up to the accepted native foundation. The plan explicitly leaves account-test parity open; this contract makes only the necessary account owner and behavior-preserving shared JSON parser extraction available. It does not reopen the native adapter, streaming, usage, billing, auth/recovery architecture, database, or deployment work.
- All seven behavior criteria are falsifiable with the contracted fake HTTP transport and production `testOpenAIImageOAuth` entrypoint: route/model/prompt/Accept, JSON item-root-default MIME selection and revised prompt event, exact fallback/error and body-close behavior, credential/header/proxy branches, and ordered downstream SSE events. Criterion 5 additionally retains the native forwarder assertions for public model, response format, dimensions, usage, and streaming behavior, so parser reuse cannot be accepted merely because account tests pass.
- The contract correctly requires one shared parser rather than a second account-only JSON parser. The current direct non-stream path already derives item `output_format`, then root `output_format`, then request default while preserving its response payload/model/dimensions/usage handling; the extracted helper must retain those semantics and the focused parser regressions must cover all three priority levels.
- Existing native edge tests already provide the complementary direct-adapter negative coverage (single 404/405 fallback stop, empty/native HTTP failure no replay, and pre-output retry handling). Account parity must add equivalent account-entrypoint evidence, including closed bodies, rather than treating direct-adapter evidence as account-test coverage.
- OAuth/SetupToken/Agent Identity/shadow resolution and proxy/custom/default User-Agent plus identity/ChatGPT-account headers remain inside the existing account function. The named new account-native test file is sufficient to use repository and transport fakes for each applicable branch; no real credential, recovery call, or provider is required. API-key tests are expressly unchanged.
- The focused selector was executed from `backend` with `go test ./internal/service -list 'TestAccountTestService_OpenAIImage|TestCodexDirectImages|TestOpenAIGatewayServiceForwardImages'`. It compiled and listed the existing account-image, direct-native, edge, and forwarding families, including `TestAccountTestService_OpenAIImageOAuthHandlesOutputItemDoneFallback`, the `TestCodexDirectImages...` cases, and the `TestOpenAIGatewayServiceForwardImages...` cases. New parity tests must retain the contracted `TestAccountTestService_OpenAIImage` or `TestCodexDirectImages` prefixes so the same selector actually includes them.
- Worktree inspection shows only the controller's `docs/workflow/status.md` phase/current-sprint hunk. Read-only main-tree review found three unrelated dirty `account_test_service.go` hunks: Pelican prompt propagation in normal account tests, its helper, and Pelican error sanitization. The contract restricts any target change to `testOpenAIImageOAuth` and its comment and requires controller hunk-only application/staging; this is adequate to preserve all three main-tree hunks without authorizing their modification.

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes` — `gpt-5.6-terra`
- base_commit_confirmed: `yes` — `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779`
- openspec_traceable: `yes` — accepted native plan identifies I-03/I-04 and explicitly defers account-test parity

## Approval

PASS. The contract is approved for the stated isolated worktree only. Final QA must inspect the exact parser extraction for direct-forwarder equivalence, rerun the selected account/native/forwarding tests plus the full service/build commands, and verify controller application preserves every non-target main-tree `account_test_service.go` hunk. This is fake-transport/local evidence only; real provider, database, container, deployment, push, and runtime credential recovery remain out of scope.

## Addendum Review — Baseline Validator Exception (2026-09-16)

### PASS: image2-account-parity

The pending exception is approved only with the following narrow interpretation.

- A clean, detached worktree at `E:/codex-worktrees/sub2api/integration-baseline-dc851ea3f` resolves to the exact declared baseline `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779` and has no candidate implementation. Independent reproduction from its `backend` directory of `go test ./internal/service -run '^TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation$' -count=1` exits 1 at `claude_code_validator_test.go:52` with `Should be false`.
- This exception covers exactly that named baseline validator test and no other result. QA must still run `go test ./internal/service -count=1`, record its non-zero result as `FAIL`, and identify this exact failure explicitly. It does not turn the full-suite command green and does not support a release claim.
- Local scoped-integration approval remains conditional on every other full-suite case passing, as well as all focused account/native/forwarding cases, build, formatting, path, conflict, parser-equivalence, and main-dirty-preservation gates. A new failure, a changed assertion/location for the named baseline failure, or a second failure is blocking until classified and fixed/replanned.
- In particular, existing or new account User-Agent, Agent Identity, shadow/credential, proxy, identity-header, or ChatGPT-account-header regressions are not exempt. The known account User-Agent defect remains a blocking candidate defect assigned to the Developer; it must be fixed and independently retested before local integration can pass.

## Addendum Gate Checks

- baseline_exact_commit_confirmed: `yes`
- baseline_failure_reproduced_cleanly: `yes`
- exception_single_test_only: `yes`
- full_suite_still_required_and_reported_fail: `yes`
- account_ua_or_identity_failure_exempted: `no`

## Addendum Approval

PASS for this contract-only exception. The workflow phase may proceed only under the conditions above; it does not waive any account behavior or UA regression, and final QA must preserve the distinction between a scoped local-integration decision and the intentionally failing full-suite command.

## Addendum Review — Complete Baseline Failure Enumeration (2026-09-16)

### FAIL: image2-account-parity

The three proposed additional exceptions are baseline facts, but this amendment is not approved yet because the task contract still states that *only* `TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation` may remain. That explicit singular rule conflicts with the four-test scope now proposed in `docs/workflow/baseline-validator-exception.md`; a review artifact cannot silently broaden a Generator/QA contract.

### Verified Baseline Facts

In the clean detached worktree at exact `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779`, independent execution of
`go test ./internal/service -run '^(TestApplyCodexOAuthTransform_PreservesAllowedTools|TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested|TestParseSSEUsage_SelectiveParsing)$' -count=1`
returned exit 1 and reproduced exactly:

- `TestApplyCodexOAuthTransform_PreservesAllowedTools` — `openai_codex_transform_test.go:18`, `Should be false`.
- `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested` — `openai_codex_transform_test.go:275`, expected `fc_6336c46cf80eccddb18637bafbdab3b1de6137977fd8f941d2ccc9763b32d`, actual `fc_477489beb020a16331ccd2917b0310e7bb315418ee7f92d3781a5e28eb761`.
- `TestParseSSEUsage_SelectiveParsing` — `openai_gateway_service_test.go:3259`, expected `9`, actual `1`.

Together with the previously reproduced `TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation` at `claude_code_validator_test.go:52` (`Should be false`), these are the proposed complete four-test baseline allowance. They do not authorize a production-code change, test change, assertion weakening, or an account/image behavior exemption.

### Required Contract Correction Before PASS

The controller must amend only the task's exception/stop-rule wording to enumerate all four exact test-name, source-line, and assertion-difference fingerprints, and state that this list is exhaustive. The amended contract must continue to require:

- the unfiltered `go test ./internal/service -count=1` command, its original non-zero exit, and a complete `go test -json` failure-event enumeration rather than truncated console output;
- report the full-suite command as `FAIL`, never green or release-ready;
- block on any missing/changed fingerprint or any extra failure event; and
- treat every account User-Agent, Agent Identity, shadow/credential, proxy, identity-header, and ChatGPT-account-header failure as blocking. The known UA defect is not an exception and requires Developer repair plus independent retest.

After that documentation-only correction, the exact four baseline fingerprints may be reviewed as a narrow local scoped-integration exception. No business or test-file modification is approved by this review.

## Addendum Review — Corrected Exhaustive Four-Fingerprint Exception (2026-09-16)

### PASS: image2-account-parity

The prior documentation conflict is resolved. The task Stop Rules now contain an exhaustive, individually fingerprinted four-test exception that matches the clean-baseline reproductions:

- `TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation` at `claude_code_validator_test.go:52`, `Should be false`.
- `TestApplyCodexOAuthTransform_PreservesAllowedTools` at `openai_codex_transform_test.go:18`, `Should be false`.
- `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested` at `openai_codex_transform_test.go:275`, with the exact expected and actual `fc_` hashes.
- `TestParseSSEUsage_SelectiveParsing` at `openai_gateway_service_test.go:3259`, expected `9`, actual `1`.

The revised wording correctly requires unfiltered `go test ./internal/service -count=1` execution, preserves and reports its original non-zero `FAIL` result, and requires a complete `go test -json` `Action=fail` enumeration. It blocks a missing/changed fingerprint and every extra failure, rather than accepting a truncated console excerpt.

The exception remains solely a scoped local-integration accounting rule. It authorizes no business/test/assertion change and no green-full-suite or release claim. Account User-Agent, Agent Identity, shadow credential, proxy, identity-header, ChatGPT-account-header, and every new candidate failure remain blocking; the UA defect must be fixed and independently retested.

## Corrected Addendum Gate Checks

- four_fingerprints_explicit_and_exhaustive: `yes`
- clean_baseline_fingerprints_independently_reproduced: `yes`
- unfiltered_full_suite_required_with_original_fail_exit: `yes`
- complete_json_failure_enumeration_required: `yes`
- extra_or_changed_failure_blocks: `yes`
- account_boundary_regressions_exempted: `no`

## Corrected Addendum Approval

PASS for the documentation-only exception amendment. Contract execution may proceed under the four exact fingerprints and no broader allowance; final QA remains responsible for the complete candidate full-suite JSON enumeration and all ordinary focused/path/build gates.

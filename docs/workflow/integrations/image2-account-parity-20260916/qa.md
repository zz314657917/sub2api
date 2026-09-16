### PASS: image2-account-parity

# Independent Terra QA — Final Scoped Retest (2026-09-16)

## Scoped Verdict

**PASS for the approved local scoped-integration contract only.** The original
OAuth identity regression is fixed, all candidate-focused, build, formatting,
scope, parser-equivalence, and main-dirty-preservation gates pass, and the
complete JSON full-suite failure set exactly matches the independently approved,
exhaustive four-fingerprint clean-baseline exception.

This is not a green full-suite or release verdict.

## Full Service Command Verdict

**FAIL:** `go test ./internal/service -count=1` completed with exit 1. A second
untruncated verification used `go test -json ./internal/service -count=1`,
exited 1 after 66.234s, and extracted every `Action=fail` event. The exact four
test failures are:

1. `TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation` —
   `claude_code_validator_test.go:52`, `Should be false`.
2. `TestApplyCodexOAuthTransform_PreservesAllowedTools` —
   `openai_codex_transform_test.go:18`, `Should be false`.
3. `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested` —
   `openai_codex_transform_test.go:275`, expected
   `fc_6336c46cf80eccddb18637bafbdab3b1de6137977fd8f941d2ccc9763b32d`, actual
   `fc_477489beb020a16331ccd2917b0310e7bb315418ee7f92d3781a5e28eb761`.
4. `TestParseSSEUsage_SelectiveParsing` —
   `openai_gateway_service_test.go:3259`, expected `9`, actual `1`.

The package-level `Action=fail` follows these four test events. No additional
test failure events appeared. Each fingerprint was independently reproduced in
the clean detached `dc851ea3f` baseline and is authorized only as the reviewed
scoped exception. A missing, changed, or extra failure remains blocking.

## Retest Findings

No candidate-specific blocking defect found.

- `enforceCodexIdentityHeaders(req.Header)` is again invoked after the
  custom/default User-Agent assignment. The prior failure
  `TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity`
  passes, preserving final Codex UA/identity normalization.
- Fake-transport tests now prove body closure for empty/malformed JSON, second
  404/405 fallback and other HTTP failures, and prove Agent Identity sensitive
  error content is absent from both returned errors and SSE output.
- The shared direct native JSON parser remains reused by the non-stream
  forwarder and account entrypoint. Focused cases cover item/root/default MIME
  precedence, malformed result rejection, mapped payload/prompt, single
  fallback/non-replay, SSE order, proxy, SetupToken, shadow, Agent Identity,
  ChatGPT account header, and unchanged APIKey paths.

## Executed Checks

- `go test ./internal/service -run '^(TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity|TestAccountTestService_OpenAIImageNative)' -count=1` — PASS (0.080s).
- `go test ./internal/service -run 'TestAccountTestService_OpenAIImage|TestCodexDirectImages|TestOpenAIGatewayServiceForwardImages' -count=1` — PASS (0.119s).
- `go test ./internal/service -count=1` — **FAIL** (66.086s), as required;
  JSON enumeration above is the complete failure evidence.
- `go test -json ./internal/service -count=1` with PowerShell `Action=fail`
  extraction — **FAIL** (exit 1, 66.234s), exactly the four reviewed baseline
  fingerprints plus package aggregate event.
- `go build ./...` — PASS.
- Contract allowlist, `gofmt -l`, `git diff --check dc851ea3f`, and unmerged
  index gate — PASS. Modified business paths are the five allowed service paths
  plus the new native account test; ignored workflow evidence was enumerated.
- Read-only main-tree inspection confirms the non-target Pelican prompt
  propagation and `sendErrorAndEnd` error-sanitization hunks remain present;
  this QA neither changed nor staged them.

## Historical Correction

The first QA run correctly found the OAuth identity regression, which is now
resolved. Its console output was truncated, so a later report statement that
the validator was the “only observed” full-suite failure was not complete
failure-set evidence and is superseded by the untruncated JSON enumeration
above. The contract and independent review now explicitly approve all four,
and only those four, clean-baseline fingerprints for scoped integration.

## Unverified Risks

No real provider, credential recovery, database, container, deployment, commit,
or push was performed. This is fake-transport/local source evidence only.

## Recommendation

The account-parity change may proceed through the approved local scoped
integration gate. Keep the full service command reported as **FAIL** and do not
claim a green suite, release readiness, or production validation until the four
independent baseline failures are resolved by their respective owners.

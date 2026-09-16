---
status: approved
review_verdict: PASS
task_id: claude-mid-output-config
worker_model: gpt-5.6-terra
base_commit: dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779
spec_ref: docs/workflow/tasks/claude-mid-output-config.md
---

## Task ID
claude-mid-output-config

## Role
Terra Developer and fresh independent Terra QA; controller owns final integration.

## Goal
Behavior-port upstream d8326fccfce4ba011f2fc1e148a0181912a9dfd1 into existing
gateway_request.go sanitizer. Preserve message output_config when its exact beta
is sent; otherwise strip only that message-level field and empty system control
messages. No change to top-level output_config or effort.

## Success Criteria
1. Add BetaMidConversationOutputConfig = mid-conversation-output-config-2026-07-01
   and include it in FullClaudeCodeMimicryBetas without altering other defaults.
2. Existing sanitizeAnthropicBodyForBetaTokens invokes the bounded sanitizer.
   Exact comma-token match makes the new helper a byte no-op; missing token strips
   only messages[].output_config. Top-level effort/output_config are preserved.
3. Only system messages carrying output_config and lacking content are removed:
   missing/null/empty string/empty array/empty text blocks count as no content.
   Preserve populated system, all user/assistant messages, unknown/nontext blocks,
   metadata and message ordering. Adjacent removals work. Without target fields
   return original bytes and changed=false. Malformed JSON fails conservatively.
4. Test actual buildUpstreamRequest with OAuth mimic (default and policy drop),
   buildUpstreamRequestAnthropicAPIKeyPassthrough (client token present/absent),
   buildCountTokensRequest with OAuth mimic (default and policy drop), and
   buildCountTokensRequestAnthropicAPIKeyPassthrough (client token present/absent).
   Each asserts outgoing exact beta token presence, message count/order and
   preserved top-level output_config.effort. Native count_tokens is excluded;
   do not claim all count_tokens routes are covered. Add an enabled-CCH actual
   builder test with target beta dropped and removed message output_config,
   proving the signature matches the final sanitized body. Existing context,
   thinking and fallback sanitizer behavior must be exercised in new untagged
   compatibility cases. gateway_service.go remains denied.
5. Run focused service tests, full service suite and full backend build. Evidence
   is local request-builder/fake testing, not live Anthropic acceptance.

## Allowed Paths
- backend/internal/pkg/claude/constants.go (new beta constant/list member only)
- backend/internal/service/gateway_request.go (new sanitizer/helper and invocation only)
- backend/internal/service/gateway_mid_conversation_output_config_test.go
- backend/internal/service/gateway_context_management_test.go (beta list expectation only if needed)
- docs/workflow/worker-results/claude-mid-output-config-result.md
- docs/workflow/qa-reports/claude-mid-output-config-qa.md

## Denied Paths
All other business paths, gateway_service.go, images/account testing, frontend,
auth/recovery, schemas/migrations, secrets, providers, databases, containers,
deployment, commits/push, merge/rebase/cherry-pick and all other worktrees.

## Constraints
Use apply_patch. Adapt behavior, never import upstream files wholesale. The local
sanitizer also strips fallbacks/fallback_credit_token: preserve those rules.
The new test file MUST NOT have a build constraint, and all its tests MUST contain
MidConversation in their names. Tests may not rely on helpers compiled only with
the unit tag. Byte-identical whole-body tests use fixtures without any other
legacy field requiring sanitization; malformed JSON helper cases return the
original bytes and changed=false. Existing sanitizer behavior is not replaced.
Stop for architectural/security changes or required edits outside the allowlist.
Independent contract PASS and approved frontmatter required before build.

## Acceptance Commands
From backend:
`go test ./internal/service -list 'Test.*MidConversation'`
The list MUST contain new sanitizer, all four builder, and CCH test families;
absence of any family is FAIL. Report the list with the result.
`go test ./internal/service -run 'Test.*MidConversation' -count=1`
`go test ./internal/service -count=1`
`go build ./...`
From root: `git diff --check`, `git ls-files -u`; gofmt -l on the four exact Go
paths (skip nonexistent optional file), and enumerate tracked/untracked changes
to prove only the business allowlist is modified. Report exact selected test
names so a selector with no new tests cannot pass acceptance.

## Output
Worker result begins ### DONE/FAILED/BLOCKED: claude-mid-output-config. QA begins
### PASS/FAIL/BLOCKED: claude-mid-output-config, with findings, changed paths,
executed commands, parser/body-preservation evidence and unverified risks.

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

Do not broaden scope or weaken tests. No real provider or runtime credential
request. Terra unavailable is BLOCKED, never silently substitute another model.

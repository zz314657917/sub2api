# Selective integration continuation — 2026-09-16

## Verified state

- Native Images foundation is committed locally at dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779.
- Account-test parity passed independent scoped QA and is committed as 6ed89efe2.
  Evidence: ../image2-account-parity-20260916/. The isolated worktree remains at
  E:/codex-worktrees/sub2api/image2-account-parity; do not reapply its dirty files.
- Claude message output_config passed independent scoped QA and is committed as
  d48fca7d3. Evidence: ../claude-mid-output-config-20260916/. The isolated worktree
  E:/codex-worktrees/sub2api/claude-mid-output-config is done; do not reapply it.
- Main worktree contains unrelated Pelican and first-response-timeout changes.
  Its shared status/current-task documents belong to concurrent work and are not
  overwritten by this continuation.

## Candidate triage

- 8ea4dc56f: do not copy the upstream Gemini response-signal implementation.
  Local source has no gemini_response_signal.go/detectGeminiResponseSignal or
  abnormal-stop event classifier. Local mapGeminiFinishReasonToClaudeStopReason
  defaults unknown reasons to end_turn. The specific upstream finishReason-to-502
  defect is not present in that local topology. Antigravity has separate logging
  for MALFORMED_FUNCTION_CALL; no changes to that path are authorized here.
  This is source triage, not full Gemini runtime parity or live-provider proof.
- d8326fccf: confirmed local gateway_request.go already owns the shared beta
  sanitizer and gateway_service.go invokes it before signing. A bounded port
  avoids the dirty gateway_service.go entirely. Port complete as d48fca7d3.

## Next gates

1. Reconcile remaining roadmap candidates without repeating completed commits.
   The 9eb120dd4 privacy/account-check fingerprint change still needs its own
   contract and provider acceptance boundary; no implementation done here.
2. Four full-service baseline failures remain: ClaudeCodeValidator strict
   validation, two ApplyCodexOAuthTransform cases, and ParseSSEUsage selective
   parsing. Both candidates and clean baseline reproduced exact fingerprints.
   Complete JSON failure enumeration was used after truncated console output
   hid three failures. Scoped exception approved independently; full suite FAIL.
3. Real-provider Image acceptance and broader architecture/database Epics remain
   open; do not infer authority from the selective migration request.

No push, deployment, database, container update or real provider call performed.

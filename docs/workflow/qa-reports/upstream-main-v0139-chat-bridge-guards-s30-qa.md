### PASS: upstream-main-v0139-chat-bridge-guards-s30

# QA Report

## Task ID
upstream-main-v0139-chat-bridge-guards-s30

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-v0139-chat-bridge-guards-s30.md`

## Evidence
- diff reviewed: yes
- allowed paths checked: yes
- denied paths touched: no
- commands run:
```text
go test ./internal/service -run "TestForwardAsChatCompletions_EnforcesCodexCLIOnlyRestriction|TestForwardAsChatCompletions_OAuthDoesNotInjectDefaultInstructions|TestForwardAsChatCompletions_APIKeyPropagatesPromptCacheKeyInResponsesBody|TestForwardAsChatCompletions_TransportErrorReturnsFailover" -count=1 -> pass
```
- manual checks:
```text
Reviewed upstream ae5e980dd and dbdbfb112 against local ForwardAsChatCompletions -> pass
Confirmed codex_cli_only rejected request leaves upstream.lastReq nil -> pass
Confirmed normal OAuth chat-completions bridge sends instructions as empty string and does not include default Codex prompt text -> pass
Confirmed staged scope will exclude existing Ent, migration, proxy/account, frontend, and knowledge dirty files -> pending staged audit
```

## Findings
- 未发现明确问题。

## Bug Owner Recommendation
codex-planner

## Root Cause
none

## Retest Scope
- If chat-completions bridge routing or Codex client detection changes again, rerun S30 tests plus the existing raw chat-completions fallback tests.

## Knowledge Promotion
none

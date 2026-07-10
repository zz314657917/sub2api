### PASS: upstream-v0151-first-batch-s60-qa

# QA Report

## Task ID
upstream-v0151-first-batch-s60-qa

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/upstream-v0151-first-batch-s60-qa.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:
```text
git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business status --short --branch
-> PASS: initial branch state was `## codex/upstream-v0151-first-batch-s60-business`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business merge-base HEAD 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca
-> PASS: returned baseline `3332c6883e7480f030fcffbccb6dc7ee0a3f69ca`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business log --oneline 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca..HEAD
-> PASS: exactly two commits: `a07c3e669 fix(openai): pair Codex Luna identity headers` and `7f47afa11 fix: port upstream protocol and dashboard compatibility`

mkdir -p E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/.tmp/go-build-qa
-> PASS

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/backend && GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/.tmp/go-build-qa go test ./internal/pkg/apicompat -run 'ParallelToolCalls|ChatCompletionsToResponses|ResponsesToChatCompletionsRequest' -count=1 -v
-> PASS

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/backend && GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/.tmp/go-build-qa go test ./internal/handler/admin -run 'TestGetUserBreakdown' -count=1 -v
-> PASS

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/backend && GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/.tmp/go-build-qa go test ./internal/pkg/openai -run 'Codex|PairCodexClientIdentity' -count=1 -v
-> PASS

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/backend && GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/.tmp/go-build-qa go test ./internal/service -run 'TestCodexVersionConstants_Consistency|TestEnforceCodexIdentityHeaders|Test.*Luna.*Identity|Test.*CompatMessagesBridge.*Originator|TestIsOpenAIWSClientDisconnectError' -count=1 -v
-> PASS

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/backend && GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/.tmp/go-build-qa go test ./internal/service -run 'Test.*(OpenAICodex|CodexIdentity|BuildOpenAIWSHeaders|AccountTest).*' -count=1 -v
-> PASS

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/backend && GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/.tmp/go-build-qa go test ./internal/service -run 'TestOpenAIGatewayServiceForwardAsAnthropicMappedNonCodexOmitsOriginator|TestOpenAIGatewayService_RecordLunaIdentityPairsOfficialCodexHeaders|TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity|TestOpenAIBuildUpstreamRequestOAuthOfficialClientOriginatorCompatibility|TestIsOpenAIWSClientDisconnectError|TestCodexVersionConstants_Consistency' -count=1 -v
-> PASS

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/frontend && npm.cmd install --no-package-lock
-> PASS: local dependency recovery only under `frontend/node_modules`

cd E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/frontend && rm -f package-lock.json && npm.cmd run typecheck
-> PASS

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business diff --check 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca..HEAD
-> PASS: no whitespace or patch-format errors

rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/backend E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business/frontend
-> PASS: exit 1, no conflict markers found

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business rev-parse HEAD
-> PASS: `a07c3e669cb3027a9eb8efa7c0063b3f0bd11025`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business rev-parse a07c3e669
-> PASS: `a07c3e669cb3027a9eb8efa7c0063b3f0bd11025`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business rev-parse HEAD^
-> PASS: `7f47afa1199b6a64c284040017eefc9893d7df02`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business rev-parse 7f47afa11
-> PASS: `7f47afa1199b6a64c284040017eefc9893d7df02`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business rev-parse HEAD^^
-> PASS: `3332c6883e7480f030fcffbccb6dc7ee0a3f69ca`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business diff --name-only 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca...HEAD
-> PASS: changed paths exactly match the contract list of 23 files

python allowed-path audit helper against git diff --name-only 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca...HEAD
-> PASS: `allowed_count=23`, `unexpected=<none>`, `missing=<none>`

git -C E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business status --short
-> PASS: final business worktree clean
```
- manual checks:
```text
no-tests audit -> PASS: none of the six required Go test groups emitted `[no tests to run]`
no live upstream credentials/network -> PASS: all evidence comes from local unit/typecheck commands only
```

## Findings
- 未发现明确问题。
- S60a protocol/admin coverage passed with verbose evidence for `ParallelToolCalls`, `ResponsesToChatCompletionsRequest`, `ChatCompletionsToResponses`, and `TestGetUserBreakdown*` cases.
- S60b contract coverage passed with verbose evidence for the required named behaviors: Luna identity pairing (`TestOpenAIGatewayService_RecordLunaIdentityPairsOfficialCodexHeaders`), Messages final-originator omission (`TestOpenAIGatewayServiceForwardAsAnthropicMappedNonCodexOmitsOriginator`), image account identity enforcement (`TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity`), official client pairing (`TestPairCodexClientIdentity` / `TestOpenAIBuildUpstreamRequestOAuthOfficialClientOriginatorCompatibility`), exact version floor `0.144.1` (`TestEnforceCodexIdentityHeaders/低于门槛的_version_提升为内置版本` and `/达标_version_原样保留`, plus `TestCodexVersionConstants_Consistency`), and Windows reset disconnect handling (`TestIsOpenAIWSClientDisconnectError/windows_remote_forced_close`).
- Allowed/denied-path audit passed: exactly the 23 contract-authorized source/test files changed from baseline, with no extra tracked business paths touched.

## Bug Owner Recommendation
`none`

## Root Cause
- `none`

## Retest Scope
- None. Full contract gate passed.

## Knowledge Promotion
- `none`

## Unverified Risks
- No live runtime smoke, browser interaction, or real upstream integration was exercised; verification stayed at local Go unit tests, frontend typecheck, and git/path hygiene as required by contract.

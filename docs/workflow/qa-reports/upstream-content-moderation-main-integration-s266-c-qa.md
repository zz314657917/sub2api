### BLOCKED: upstream-content-moderation-main-integration-s266-c

# QA Report

## Task ID
upstream-content-moderation-main-integration-s266-c

## Verdict
`BLOCKED`

## Contract Checked
- `docs/workflow/tasks/upstream-content-moderation-main-integration-s266-c.md`

## Findings
- The contract's protected outputs gate cannot be verified. `F:/mcplugins/sub2api/outputs/` contains the required 20 untracked files and the tracked/index state is clean, but a deterministic manifest of current SHA-256 entries (`<file-sha256><two spaces><relative path>`, path-sorted, UTF-8, no final newline) is `CB91CA1C6CC1DD00B180B64547E06387C651EA8CC8AEC2F10087B42B59C626E2`, not frozen `2996311A4EC1458EEC9C2AE4327D5D5EAA695C878783DE984AF841BBF0A79145`.
- Recomputations with slash/backslash paths, absolute/relative paths, hash/path order, optional size, LF/CRLF/final newline, and UTF-8/UTF-16 serializations did not yield the contract value. Neither the contract nor accessible workflow records state the original manifest serialization or per-file SHA-256 baseline. Therefore QA cannot prove that the protected untracked outputs are unchanged.
- The contract and user instruction require immediate BLOCKED on outputs-manifest drift. Acceptance test commands were intentionally not run after this protection gate failed.

## Scope and Provenance Evidence
```text
main HEAD -> f080bbd094e2afd196ada2c00fbcfe7b86275361 (expected f080bbd09)
git status --short -> ?? outputs/ only
git diff --quiet -> 0
git diff --cached --quiet -> 0
git ls-files -u -> empty
outputs file count -> 20
6054b9266 patch-id -> 922e3bc0fc4ccf5c1bd1fecc41e33e8894ac8d0a (matches c2cd7a0a1)
f080bbd09 patch-id -> 1c0333b260eb320ffeb89982757756a9e8c26df1 (matches eeed2369f)
both commits retain -x provenance; 6054b9266 is an ancestor of f080bbd09
21 S266-A paths + 56 S266-B paths -> exact 67-path union
main delta 2a3664747..HEAD -> 67 paths, exact union; frozen main delta overlap -> 0; denied-path hits -> 0
```

## Commands Not Run
- All backend/frontend Acceptance Commands, including compilation and build, were not run because the contract stop rule requires stopping on protected outputs drift.

## Bug Owner Recommendation
`codex-planner`

## Root Cause
- `contract-ambiguous`: the contract supplies only an aggregate outputs digest, without a reproducible manifest serialization or individual-file baseline, and the current reproducible manifest does not equal the supplied digest.

## Retest Scope
- Controller must provide the original manifest generation command and/or the 20-file relative-path/SHA-256 baseline, then confirm whether `outputs/` is unchanged. On an exact match, dispatch a fresh independent QA run from the frozen `main@f080bbd09` state and execute the full Acceptance Commands.

## Unverified Risks
- Combined backend/frontend regressions, server compilation, typecheck/build, gofmt, and final diff gate are unverified in this QA attempt because of the mandatory early stop.
- No provider, SMTP, Redis/PostgreSQL, container, deployment, staging, or push action was performed.

## Knowledge Promotion
`none`

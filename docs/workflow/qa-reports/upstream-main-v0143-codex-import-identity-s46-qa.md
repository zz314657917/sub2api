### PASS: upstream-main-v0143-codex-import-identity-s46

## Findings
- 未发现 S46 diff 的明确问题。
- PASS: `buildCodexIdentityKeys` 现在把 `user:` 放在最高优先级，`account:` 作为共享账号 fallback 放最后。
- PASS: `codexAccountIndex` 对共享 `account:` 键保留多候选，并跳过 `chatgpt_user_id` 冲突的候选。
- PASS: legacy account 缺少 `chatgpt_user_id` 时仍可通过 `account:` fallback 被更新和回填，并会给出人工确认 warning。
- PASS: in-batch dedup 复用同一冲突规则，同一 ChatGPT team 的不同 user 不会被误判成重复导入项。
- PASS: 前端只给 Codex session import API 调用设置 120s timeout，没有提高全局 API client timeout。

## Executed Checks
- `go test ./internal/handler/admin -run "TestCodexIdentity|TestParseCodexSessionImport|TestNormalizeCodexImport|TestResolveCodexImport|TestMergeCodexImport|TestImportCodex" -count=1`
  - Result: PASS.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - Result: PASS.
- `git diff --check`
  - Result: PASS.
- Two read-only explorer checks:
  - Codex import explorer confirmed local code was not equivalent before S46, upstream patch applied cleanly, and S46 should be independent.
  - Ops realtime explorer recommended `3f2ef6046` as a later independent Sprint, not mixed into S46.

## Unverified Risks
- No real browser/admin import flow was executed.
- Full backend and frontend suites were not run.
- Very large real Codex exports can still fail if deployment-level timeout is lower than 120s.

## Recommendation
- PASS S46 for scoped commit after staged denied-path audit.
- Next candidate: `3f2ef6046` ops realtime account stats performance, as S47. Keep its group-filter semantics risk explicit.

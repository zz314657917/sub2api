### PASS: upstream-api-key-validation-s209

## Findings

- 未发现阻断问题或范围漂移。
- Handler 校验位于 Create 幂等边界和 Update service 调用之前；service
  校验位于任何 repository 读取之前，满足双层 fail-closed 目标。
- 有效 `0 = unlimited`、大有限值、nil/正数 Create 有效期保持可用；本轮
  diff 未改 Update 的 RFC3339 解析或空字符串清除有效期路径。

## Executed Checks

- Committed-tree focused handler regressions:
  `go test ./internal/handler -run '^TestS209' -count=10` -> PASS.
- Committed-tree focused service regressions:
  `go test ./internal/service -run '^TestS209' -count=10` -> PASS.
- Complete committed-tree packages:
  `go test ./internal/handler ./internal/service -count=1` -> PASS
  (`handler` 27.245s, `service` 70.363s).
- Dependent server compilation:
  `go test ./cmd/server -run '^$' -count=0` -> PASS.
- `gofmt -d` over all four Go paths -> empty/PASS.
- `git diff --check`, exact nine-path allowlist, frozen parent
  `bb4b74e73df85962fa5fd75c3d73a9545b63c61d`, conflict-marker scan,
  unmerged-index check, and clean-worktree check -> PASS.
- Upstream provenance:
  `f5c108c836a849dd6b982e46238c42c6db611899` is an ancestor of fetched
  `upstream/main` -> PASS.
- Diff precision review: every product line maps to numeric/create-expiry
  validation or its regression coverage; no unrelated refactor was found.

## Unverified Risks

- No deployed HTTP service, shared PostgreSQL/Redis, container, staging,
  provider, production traffic, or performance test was exercised.
- S209 is based on `main@bb4b74e73d` and intentionally excludes the separate
  uncommitted S208 primary-worktree changes. It is not currently eligible for a
  direct fast-forward into local `main`; integration order must preserve S208.
- No remote push or deployment was performed.

## Recommendation

`PASS / local-regression`: retain the isolated S209 commits as the reviewed
behavior-level port. Close and commit S208 first, then integrate S209 with an
explicitly reviewed rebase/cherry-pick sequence and rerun focused post-merge
smoke checks. Do not push or deploy from this result alone.

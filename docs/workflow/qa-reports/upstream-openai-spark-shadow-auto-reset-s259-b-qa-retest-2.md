### PASS: upstream-openai-spark-shadow-auto-reset-s259-b

## Verdict

Independent retest at `c4dce97b8` passes the S259-B contract and the two
remediation commits.  The earlier concurrent-create failure is fixed: stale
preflight contenders yield exactly one successful create and one structured
`409 SPARK_SHADOW_ALREADY_EXISTS`; non-unique errors retain their original
identity.  The only full-repository failure is the contract-recorded 32/34
sqlmock baseline panic.

## Remediation audit

- `ae69b84ca` adds conflict normalization at the `CreateShadow` repository
  create boundary.  `25b06857a` narrows it to unique-key markers only.
- `TestSparkShadowConcurrentCreateReturnsStructuredConflict` starts two
  callers after the same gate.  Its fake always returns an empty shadow list
  (stale preflight) and makes the second create fail with a repository-style
  unique message.  It asserts one success, one HTTP 409,
  `SPARK_SHADOW_ALREADY_EXISTS`, and two create attempts.
- `TestSparkShadowCreatePreservesNonUniqueFailure` supplies a foreign-key
  style error and asserts `errors.Is` retains that original error and that it
  is not labeled `SPARK_SHADOW_ALREADY_EXISTS`.
- The helper accepts only `duplicate key`, `unique constraint`, or `duplicate
  entry` markers.  Ent's `ConstraintError.Error()` is
  `ent: constraint failed: ` plus its underlying message, so an Ent constraint
  is converted only when that underlying message has a unique marker; foreign
  key/check-style Ent constraint text remains unmodified.  This source check
  is the applicable Ent-path evidence because the contract forbids a live DB.

## Commands executed

All commands ran from `backend/` in
`E:/codex-worktrees/sub2api/upstream-openai-spark-shadow-auto-reset-s259-b-qa-retest-2`.

| Command | Result |
| --- | --- |
| `go test ./internal/service -run 'TestSparkShadow(ConcurrentCreateReturnsStructuredConflict|CreatePreservesNonUniqueFailure|BulkProxy|SingleProxy|RefreshQuotaAndParentDeleteGuards)' -count=10` | PASS (`0.078s`) |
| `go test ./internal/repository -run 'Test.*(SparkShadow|Shadow.*Scheduler)' -count=10` | PASS (`0.066s`) |
| `go test ./internal/repository -run 'Test.*(SparkShadow|Spark.*Shadow|Shadow.*Spark|ParentHealth)' -count=10` | PASS (`0.072s`) |
| `go test ./internal/service -run 'Test.*(SparkShadow|Spark.*Shadow|Shadow.*Spark|ParentHealth)' -count=10` | PASS (`2.181s`) |
| `go test ./internal/handler/admin -run 'Test.*(SparkShadow|Spark.*Shadow|CreateShadow|Shadow.*Refresh|Shadow.*Reset)' -count=10` | PASS (`3.906s`) |
| `go test ./internal/handler/dto -run 'Test.*(SparkShadow|Spark.*Shadow)' -count=10` | PASS (`0.653s`) |
| `go test ./internal/server -run 'Test.*(SparkShadow|Spark.*Shadow)' -count=1` | PASS; no matching tests (`0.664s`) |
| `go test ./internal/server -count=1` | PASS (`0.082s`) |
| `go test ./internal/service ./internal/handler/admin ./internal/handler/dto ./internal/server -count=1 -timeout=3m` | PASS: service `67.097s`, admin `0.223s`, DTO `0.092s`, server `0.094s` |
| `go test ./cmd/server -run '^$' -count=1` | PASS (`5.603s`; no tests to run) |

`go test ./internal/repository -count=1 -timeout=3m` failed only at the
recorded baseline `TestUpdateWithAccountBillingSettingsRollsBackWhenOutboxFails`:
`account_repo_upstream_billing_probe_update_test.go:559` panics because the
sqlmock row provides 32 values for the generated 34-column query.  No Spark
test or other repository failure was observed before this expected stop.

## Contract audit

- Prior independently verified boundaries remain intact: no shadow credentials
  or direct proxy/refresh/reset; parent single-proxy sync is atomic; bulk proxy
  lookup is limited to normal OpenAI OAuth parents; non-OAuth bulk targets do
  not look up shadows; exports omit children and report
  `skipped_shadow_accounts`; DTOs expose only `parent_account_id` and
  `quota_dimension`; and the shadow route stays in the administrator accounts
  group.
- `83402d5e9` is an ancestor of this candidate, `bdf7ead15` is an ancestor of
  `upstream/main`, and both remediation commits (`ae69b84ca`, `25b06857a`) are
  ancestors.  Relative to `084ac39e2`, the candidate changes only allowed
  `admin_service.go`, its allowed Spark admin test, and the workflow result.
- `gofmt -d` and `git diff --check` produced no output; no conflict markers
  were found; cached and unmerged indexes were empty.
- Read-only primary protection: 122 porcelain entries, patch-id
  `6203b9cb8ed3079b9f72287409aa5bf4cbd79c17`, staged `0`, unmerged `0`.
  Its S259-B overlaps remain function-disjoint: Cafe route hunks at
  `server/routes/admin.go:190-211` versus the account route, and Group limit
  hunks at `admin_service.go:2072-2414` versus S259-B lifecycle code after
  line 3531.

## Residual risk

No live provider, database, migration, container, deployment, or push action
was performed.  The unrelated repository SQL-mock baseline remains open.
Primary-worktree integration still requires the controller's separate fresh
apply/protection gate.

### DONE: upstream-openai-spark-shadow-auto-reset-s259-b

## Ownership and commits

- The Terra Developer ended twice without satisfying the amended contract. Per the P/G/E stop rule, the Controller completed only the existing S259-B allowlist; independent Terra QA remains mandatory before any main-worktree integration.
- Earlier in-scope business commits retained: `9c08de90e`, `a18ca0092`, `df2ecb9b2`.
- Controller lifecycle, export, DTO and default-tag regression completion: `3b0a7ad94 fix(openai): enforce spark shadow admin lifecycle`.
- This report is intentionally separate from the business commit.

## Completed behavior

- A Spark shadow remains credential-less: single and batch updates reject child credentials, child proxy changes, quota refreshes and generic quota resets with structured client errors.
- A normal OpenAI OAuth parent propagates a single-account proxy update to its Spark child in the repository transaction. If the atomic capability is unavailable, the parent write stops with `SPARK_SHADOW_PROXY_SYNC_UNSUPPORTED` and cannot leave a partial child proxy state.
- Bulk proxy updates reject loaded shadow targets and append only children of loaded normal OpenAI OAuth parents. Non-OAuth/API-key and unknown legacy bulk targets do not issue a Spark lookup, retaining the preceding bulk-update contract.
- Parent deletion removes children first; exports omit shadows and return `skipped_shadow_accounts`, so credential-less child records cannot be imported as standalone accounts.
- The protected create endpoint, shallow DTO mapping and generic reset handler retain the parent link/quota dimension without exposing parent PII.

## Changed files

- `backend/internal/service/admin_service.go`
- `backend/internal/service/openai_spark_shadow_admin_test.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/admin/account_data.go`
- `backend/internal/handler/admin/account_data_handler_test.go`
- `backend/internal/handler/admin/admin_service_stub_test.go`
- `backend/internal/handler/admin/openai_oauth_handler_spark_shadow_test.go`
- `backend/internal/handler/dto/account_spark_shadow_test.go`

## Verification

All commands ran from `backend/` unless noted otherwise.

```text
go test ./internal/repository -run 'Test.*(SparkShadow|Shadow.*Scheduler)' -count=10
PASS (5.495s)

go test ./internal/service -run 'Test.*(SparkShadow|Spark.*Shadow|Shadow.*Spark|ParentHealth)' -count=10
PASS (2.140s)

go test ./internal/handler/admin -run 'Test.*(SparkShadow|Spark.*Shadow|CreateShadow|Shadow.*Refresh|Shadow.*Reset)' -count=10
PASS (0.064s)

go test ./internal/handler/dto -run 'Test.*(SparkShadow|Spark.*Shadow)' -count=10
PASS (0.084s)

go test ./internal/server -count=1
PASS (0.083s)

go test ./internal/service -run 'Test(BulkUpdateAccountsInvalidatesProbeSnapshotForProxyUpdate|SparkShadowBulkProxy|SparkShadowSingleProxy)' -count=10 -timeout=3m
PASS (5.522s)

go test ./internal/service ./internal/handler/admin ./internal/handler/dto ./internal/server -count=1 -timeout=3m
PASS (service 67.036s; admin handler 0.213s; DTO 0.122s; server 0.088s)

go test ./cmd/server -run '^$' -count=1
PASS (0.065s)

go test ./internal/repository -run 'Test.*(SparkShadow|Spark.*Shadow|Shadow.*Spark|ParentHealth)' -count=10
PASS (0.070s)

gofmt -d <all nine changed Go files>; git diff --check; exact allowlist; conflict markers; cached/unmerged index
PASS
```

## Known baseline failure

`go test ./internal/repository -count=1 -timeout=3m` was rerun after the final code change and still fails only in the contract-amended, out-of-scope baseline test `TestUpdateWithAccountBillingSettingsRollsBackWhenOutboxFails`. Its SQL mock supplies 32 row values while the generated query selects 34, panicking at `account_repo_upstream_billing_probe_update_test.go:559`. No S259-B code, fixture, Ent or schema owner changes that file; the Spark-focused repository suite passes above.

## Contract and protection checks

- The business commit contains exactly the nine allowed service/repository/admin/DTO paths. No schema, migration, generated Ent, frontend, gateway credential propagation, automatic reset, dependency, knowledge, provider, database, container, deployment or push action occurred.
- Final primary-worktree observation was read-only: 122 porcelain entries, `staged=0`, `unmerged=0`, patch-id `4faacaec1685b65ef520edc6b3d65f41909e409a`. The only overlapping dirty owner is `admin_service.go`; its live Group room-managed limit hunks around lines 2070/2249/2405 remain disjoint from the S259-B account lifecycle functions around lines 3520-5120.
- S259-A parent resolution and the existing quota/reset protections were retained unchanged. No parent credentials, account ID, email, subscription or other parent PII is added to a shadow response.

## Residual risks and next gate

- Real provider, database, migration, container, deployment and push verification remain explicitly out of scope.
- The repository-wide SQL mock failure is a separate baseline defect and must not be silently fixed within S259-B.
- Controller implementation is ready for an independent `gpt-5.6-terra` QA worktree. Main-worktree integration is not authorized until QA passes and a fresh live primary protection/apply review succeeds.

knowledge_candidates: none.

## Independent-QA remediation

- Independent Terra QA report `e213ee53f` correctly returned `FAIL`: two creates that both pass the preflight lookup can race at the database unique index, and the raw `Create` error was not normalized to `SPARK_SHADOW_ALREADY_EXISTS` / HTTP 409.
- Controller remediation `ae69b84ca fix(openai): normalize spark shadow create conflicts` maps Ent constraint errors and repository-compatible unique-key messages to that same structured conflict. The new default-tag regression starts two concurrent creates against a fake that models the stale preflight plus unique-index winner; it proves exactly one create succeeds and the other returns HTTP 409 with reason `SPARK_SHADOW_ALREADY_EXISTS`.
- Remediation verification: focused Spark repository/service/admin-handler/DTO suites x10 PASS; explicit concurrent-create/bulk/single-proxy/refresh service subset x10 PASS (5.568s); full service/admin-handler/DTO/server PASS (service 66.242s); cmd/server compile PASS (0.064s); format/diff/conflict/index/allowlist PASS.
- The final full repository run still fails only at the same excluded `account_repo_upstream_billing_probe_update_test.go:559` 32/34 SQL mock panic. Spark-focused repository x10 PASS (0.075s).
- A fresh independent Terra QA retest is required; its report must supersede the failed report before any primary-worktree integration.

## Remediation correction

- Before retest dispatch, Controller narrowed `ae69b84ca`'s Ent constraint classification: only unique-key markers now map to `SPARK_SHADOW_ALREADY_EXISTS`. A non-unique foreign-key/check-style error remains its original error and cannot be mislabeled as an existing shadow.
- `25b06857a fix(openai): preserve nonunique shadow create errors` adds that default-tag regression. The focused concurrent-create/non-unique/bulk/single-proxy/refresh subset passed x10 (5.502s); all contract focused suites x10, combined service/admin-handler/DTO/server (service 65.995s), cmd/server (0.063s), format/diff/index and protection checks passed again.
- The full repository command continues to fail only in the recorded 32/34 SQL mock baseline at `account_repo_upstream_billing_probe_update_test.go:559`; final Spark-focused repository x10 passed (0.068s). A new QA retest must begin from this corrected evidence commit.

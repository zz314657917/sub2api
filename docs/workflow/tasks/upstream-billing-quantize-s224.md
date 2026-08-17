# Task Contract: upstream-billing-quantize-s224

## Task ID

`upstream-billing-quantize-s224`

## Status

`contract-approved`

## Role

Planner, Terra Developer Worker, independent Terra QA Worker, and Final Evaluator.
The Developer may start only after the Final Evaluator approves this contract.

## Goal

Behaviorally port upstream `e2652eb853c74c0054cec5a2fa4672ccd8652d01`:
quantize every `UsageBillingCommand` monetary amount to the PostgreSQL
`NUMERIC(20,8)` scale before persistence, without changing the raw-value request
fingerprint used for idempotency.

## Success Criteria

- `Normalize` creates or preserves the request fingerprint before quantization.
- Decimal half-away-from-zero rounding at eight fractional digits covers balance,
  prepaid balance, subscription, API-key quota, API-key rate-limit, and account
  quota amounts.
- Local-only `PrepaidBalanceCost` is covered. Non-monetary fields, percentages,
  request IDs, and payload hashes are unchanged.
- Tests prove rounding boundaries, exact balance/quota reconciliation, every
  monetary field, raw fingerprint ordering, explicit fingerprints, negative,
  zero, and nonfinite values.

## Context

- Repo: `F:/mcplugins/sub2api`
- Approved implementation base: `b7d10c957f09a42d29ab43b3d3fce2629350c045`
- Worktree: `E:/codex-worktrees/sub2api/upstream-billing-quantize-s224`
- Branch: `pge/upstream-billing-quantize-s224`
- Upstream head: `396a9d1130c9a8ab977e6a959a4fdd2d9f95dd27`
- Existing dependency: `github.com/shopspring/decimal v1.4.0`.

## Allowed Paths

- `backend/internal/service/usage_billing.go`
- `backend/internal/service/usage_billing_quantize_test.go`
- `docs/workflow/worker-results/upstream-billing-quantize-s224-result.md`

## Denied Paths

- Every path outside the allowlist, especially `frontend/**`, `knowledge/**`,
  `outputs/**`, `backend/migrations/**`, `backend/ent/**`, repositories,
  dependencies, lockfiles, Docker/deployment files, and VERSION.
- User-owned dirty files: `frontend/src/components/account/EditAccountModal.vue`,
  its test, `knowledge/00-start-here.md`, `knowledge/05-current-focus.md`, and
  `outputs/`.
- Database execution, provider calls, containers, deployment, remote push,
  release tagging, or wholesale upstream merge/cherry-pick.

## Constraints

- Adapt to the local command shape. Do not replace the file with upstream.
- Quantization must occur after fingerprint creation. Do not change SQL, schema,
  migration, repository, cost calculation, routing, or public API behavior.
- Use decimal rather than binary multiply/round. Preserve nonfinite inputs.

## Acceptance Commands

```powershell
Push-Location backend
$tests = @(
  'TestUsageBillingCommandQuantizesBalanceAndQuotaIdentically',
  'TestQuantizeUsageBillingAmountBoundaries',
  'TestQuantizedAmountsReconcileExactlyOverManyApplications',
  'TestNormalizeQuantizesEveryMonetaryField',
  'TestNormalizeKeepsFingerprintDerivedFromRawAmounts',
  'TestNormalizePreservesExplicitFingerprint',
  'TestQuantizeUsageBillingAmountPassesThroughNonFinite',
  'TestQuantizeUsageBillingAmountHandlesNegativeAmounts'
)
foreach ($test in $tests) {
  $listed = go test -list "^$test$" ./internal/service
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) { throw "S224 undiscoverable: $test" }
}
$pattern = '^(' + ($tests -join '|') + ')$'
go test ./internal/service -run $pattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S224 focused regressions failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S224 service regression failed' }
go test ./internal/repository -count=1
if ($LASTEXITCODE -ne 0) { throw 'S224 repository regression failed' }
Pop-Location

$base = 'b7d10c957f09a42d29ab43b3d3fce2629350c045'
$paths = @('backend/internal/service/usage_billing.go','backend/internal/service/usage_billing_quantize_test.go')
$formatDiff = gofmt -d $paths
if ($LASTEXITCODE -ne 0 -or $formatDiff) { throw 'S224 formatting check failed' }
git diff --check $base...HEAD
if ($LASTEXITCODE -ne 0) { throw 'S224 diff check failed' }
$unexpected = @(git diff --name-only $base...HEAD | Where-Object { $_ -notin $paths })
if ($unexpected.Count -gt 0) { throw "S224 changed paths outside allowlist: $unexpected" }
if (git diff --name-only --diff-filter=U) { throw 'S224 has unresolved conflicts' }
if (git ls-files -u) { throw 'S224 index has unresolved entries' }
git merge-base --is-ancestor e2652eb853c74c0054cec5a2fa4672ccd8652d01 upstream/main
if ($LASTEXITCODE -ne 0) { throw 'S224 upstream provenance failed' }
```

## Output

- Developer report: `docs/workflow/worker-results/upstream-billing-quantize-s224-result.md`.
- Independent QA report: `docs/workflow/qa-reports/upstream-billing-quantize-s224-qa.md`.
- No push or deployment. Integrate only after independent QA PASS.

## Stop Rules

- Stop if a denied path is required, the fingerprint observes quantized amounts,
  an amount remains above eight fractional digits, or a user-owned file changes.

## Budget

- developer_worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- worktree_root: `E:/codex-worktrees/sub2api`

## Contract Review

`PASS / contract-approved`: local `PrepaidBalanceCost` is a monetary field and
is explicitly covered; fingerprint calculation remains before quantization;
the existing decimal dependency is sufficient; the focused tests are default-tag
and discoverable after implementation; no denied path is needed.

## Amendment 1: Approved Base And Controller Evidence

- The implementation worktree was created after the contract-approval commit,
  so the exact allowlist base is `b7d10c957...`, not the earlier pre-contract
  product commit `6eb2bb564...`.
- Controller review found exactly the two allowed business files in
  `b68afce67`. Eight focused tests passed with `-count=10`, complete service
  passed in 60.207s, and complete repository passed in 1.670s.
- Independent QA must use `b7d10c957...` for range and allowlist checks and
  must not treat inherited workflow approval files as implementation changes.

### PASS: upstream-main-v0143-group-peak-rate-impl-s44

## Findings
- PASS: peak-rate service rules cover subscription-only enablement, non-subscription normalization, and `peak_rate_multiplier=0`.
- PASS: token billing applies membership/user/group multiplier before the peak multiplier.
- PASS: token-mode image output tokens remain token billed and logged with the token peak multiplier.
- PASS: image/per-request billing keeps image multiplier behavior and does not inherit peak multiplier.
- PASS: admin HTTP create no longer rejects standard groups before service normalization.
- PASS: frontend typecheck and focused component/view tests pass after peak fields were threaded into types and fixtures.
- PASS: subagent review issues were resolved or explicitly scoped:
  - backend P1 create-handler prevalidation fixed.
  - frontend P2 smart-route summary display fixed.
  - workflow P1 evidence now added.

## Executed Checks
- `go test ./internal/service -run "Test.*Peak.*|Test.*Group.*Peak.*|Test.*Billing.*Peak.*|Test.*Gateway.*Peak.*|Test.*RecordUsage.*Peak.*" -count=1`
  - Result: PASS.
- `go test ./internal/handler -run "Test.*AvailableChannel.*Peak.*|Test.*Payment.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1`
  - Result: PASS, no matching non-admin handler tests for the peak patterns.
- `go test ./internal/handler/admin -run "Test.*Group.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1`
  - Result: PASS.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - Result: PASS.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/KeysView.createQuery.spec.ts src/utils/apiKeyCapabilities.spec.ts"`
  - Result: PASS, 4 files / 37 tests.
- `git diff --check`
  - Result: PASS.
- Three read-only subagent reviews:
  - backend review found create-handler prevalidation and notification metadata risk.
  - frontend review found smart-routing summary display and fixture drift risks.
  - merge-boundary review found no denied path in `main..HEAD`, migration numbering correct, but workflow evidence missing before this fix.

## Unverified Risks
- Full repository tests were not run.
- Real browser visual smoke for admin group editor, payment card, available channels, and Keys route summary was not run.
- Group update notification metadata still does not include peak-specific changed fields; auth cache invalidation covers runtime auth correctness, but live UI consumers of notification metadata may need a separate enhancement if they depend on those keys.
- Commit authors on earlier worker commits are mixed; committer is normalized. This QA does not rewrite worker branch history.

## Recommendation
- PASS S44 for scoped merge after staged denied-path audit.
- Continue with a new S45 planning scan for the remaining upstream `v0.1.143` and post-`v0.1.143` candidates; do not combine that scan with the S44 implementation commit.

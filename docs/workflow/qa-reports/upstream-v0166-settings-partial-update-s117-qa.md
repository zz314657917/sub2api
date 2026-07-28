### PASS: upstream-v0166-settings-partial-update-s117

## Findings

- No S117 defect found in final diff review. The patch only changes the
  settings handler/service boundary plus focused tests and workflow evidence.
- Partial requests preserve unrelated stored values; explicit empty and false
  values remain intentional updates; SMTP's JSON alias is covered.

## Executed Checks

- `go test ./internal/handler/admin -run 'Test(UpdateSettings(PartialPayload|ExplicitEmptyValue|ExplicitFalseValue|SMTPFromAlias)|OmittedSettingKeys)' -count=1`
- `go test ./internal/service -run 'TestSettingServiceUpdateSettingsOmitting' -count=1`
- `go test ./internal/handler/admin -run 'TestSettingHandler_UpdateSettings' -count=1`
- `go test ./... -run '^$'`
- `gofmt -d` on all four S117 Go files
- `git diff --check` and allowlist audit

## Unverified Risks

- The `unit`-tag service aggregate cannot compile because of pre-existing
  duplicate helper and outdated billing/Grok test signatures outside S117.
- No authenticated admin API request against PostgreSQL or persistent runtime
  cache was run; evidence is source-level plus Gin handler regression tests.

## Recommendation

- PASS/source-only. Keep S117 isolated until the dirty primary worktree's
  S115/S116 changes are separately committed or otherwise reconciled; do not
  merge, deploy, or refresh containers as part of this task.

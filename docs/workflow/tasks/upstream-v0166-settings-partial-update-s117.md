---
task_id: upstream-v0166-settings-partial-update-s117
status: done
role: Developer Worker
qa_mode: runtime
---

# Task Contract

## Goal

Adapt upstream `0b5903d45` to the local settings topology so a partial
`PUT /api/v1/admin/settings` request preserves stored values for omitted
value-typed settings fields instead of persisting their Go zero values.

## Success Criteria

- A payload that sends only one value-typed setting updates that setting and
  preserves unrelated stored settings, including SMTP, site, and security
  values.
- A field explicitly sent with an empty string, `false`, `0`, or empty slice
  keeps its current deliberate-clear semantics.
- `smtp_from_email` maps to the persisted SMTP-from setting key.
- Existing pointer-typed settings preserve their current merge behavior.
- Partial writes refresh in-process settings caches from stored values; full
  writes retain the existing request-object cache refresh behavior.
- Existing system-settings and auth-source-default validation remains intact.

## Allowed Paths

- backend/internal/handler/admin/setting_handler.go
- backend/internal/handler/admin/setting_handler_partial_payload_test.go
- backend/internal/service/setting_service.go
- backend/internal/service/setting_service_partial_payload_test.go
- docs/workflow/status.md
- docs/workflow/spec.md
- docs/workflow/main-log.md
- docs/workflow/tasks/upstream-v0166-settings-partial-update-s117.md
- docs/workflow/worker-results/upstream-v0166-settings-partial-update-s117-result.md
- docs/workflow/qa-reports/upstream-v0166-settings-partial-update-s117-qa.md

## Denied Paths

- backend/ent/**
- backend/migrations/**
- backend/internal/server/**
- backend/internal/config/**
- backend/internal/service/** except backend/internal/service/setting_service.go and its listed test
- frontend/**
- deploy/**
- Dockerfile*
- knowledge/**
- outputs/**

## Constraints

- Capture JSON field presence before decoding the existing request DTO; do not
  infer omission from zero values after binding.
- Derive value-typed request JSON names from the DTO tags so new fields cannot
  silently bypass the preservation rule. Keep aliases explicit and covered by
  tests.
- Do not turn the endpoint into a new PATCH API or alter its route, auth,
  response, step-up, validation, or pointer-field semantics.
- Do not use the dirty primary worktree. All changes remain in this isolated
  S117 worktree until separately reviewed and merged.

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-v0166-settings-partial-update-s117/backend
go test ./internal/handler/admin -run "TestUpdateSettings(PartialPayload|FullPayload|SMTPFromAlias)" -count=1
go test ./internal/service -run "TestSettingServiceUpdateSettingsOmitting" -count=1
go test ./... -run "^$"
gofmt -d internal/handler/admin/setting_handler.go internal/handler/admin/setting_handler_partial_payload_test.go internal/service/setting_service.go internal/service/setting_service_partial_payload_test.go
cd E:/codex-worktrees/sub2api/upstream-v0166-settings-partial-update-s117
git diff --check
git diff --name-only HEAD
```

## Output

- Adapted handler/service behavior, focused regressions, developer result,
  QA report, and a path-constrained diff ready for final review.

## Stop Rules

- Stop if preserving omitted fields requires a schema/migration, route/API
  contract change, frontend payload change, or a broad settings refactor.
- Stop if existing pointer-field merges or fail-closed security defaults cannot
  be retained by a narrow handler/service adaptation.

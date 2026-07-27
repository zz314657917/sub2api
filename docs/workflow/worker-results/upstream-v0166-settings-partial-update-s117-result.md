### DONE: upstream-v0166-settings-partial-update-s117

## Changed Behavior

- The admin settings handler records the incoming top-level JSON field names
  before decoding the existing value-typed DTO.
- The settings service drops omitted keys from the generated update map and
  reloads persisted settings before refreshing in-process caches.
- Explicit empty strings and `false` values remain writes. The
  `smtp_from_email` request alias maps to the persisted SMTP-from key.

## Files

- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/admin/setting_handler_partial_payload_test.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/setting_service_partial_payload_test.go`
- S117 workflow artifacts only.

## Checks

- Focused handler and service regressions passed.
- Existing settings handler regressions and `go test ./... -run '^$'` passed.
- `gofmt -d` and `git diff --check` returned no source findings.

## Risks

- The repository's `unit`-tag service suite does not compile because of
  pre-existing test drift outside S117. No database-backed authenticated API
  smoke was run.

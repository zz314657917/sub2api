### DONE: upstream-main-v0144-safe-patches-s53

## Changed Files

- `backend/internal/service/token_refresh_service.go`
- `backend/internal/service/token_refresh_service_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_service_hotpath_test.go`
- `backend/internal/handler/admin/account_codex_import.go`
- `backend/internal/handler/admin/account_codex_import_test.go`
- `docs/workflow/tasks/upstream-main-v0144-safe-patches-s53.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Summary

- Cherry-picked `e5dc1f597` with a focused conflict resolution that adds `token_expired` to the existing local non-retryable token refresh list and test table.
- Cherry-picked `4dd3aee5c`, then scoped the imported hotpath tests back to the S53 mapped-billing assertions only. The upstream commit included extra hotpath request-view/image tests that depend on helpers not present in this local branch and are unrelated to mapped billing.
- Cherry-picked `6bd248fd1` without conflicts to prevent Codex access-only imports from being merged into existing full accounts.

## Contract Compliance

- Allowed paths only.
- Denied paths not touched: Ent, migrations, deploy, README, `.github`, frontend, payment, welfare, Docker/container files, and unrelated v0.1.144 feature paths.
- Larger v0.1.144 items remain deferred.

## Validation

- See `docs/workflow/qa-reports/upstream-main-v0144-safe-patches-s53-qa.md`.

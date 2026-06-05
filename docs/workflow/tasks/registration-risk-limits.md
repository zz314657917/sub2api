---
task_id: registration-risk-limits
role: codex-direct
qa_mode: runtime
status: contract-approved
---

# Registration Risk Limits

## Goal

Add backend registration abuse controls for email registration:

- Successful registrations per client IP in a configurable 24-hour window.
- Registration entry attempts per `IP + User-Agent` in a configurable short window.
- Registration email-domain attempts in a configurable short window.

## Allowed Paths

- `backend/internal/config/**`
- `backend/internal/server/routes/**`
- `backend/internal/middleware/**`
- `backend/internal/service/**`
- `backend/internal/pkg/**`
- `backend/internal/handler/**` only if handler wiring is needed
- `backend/cmd/server/wire_gen.go` only if constructor wiring is needed

## Denied Paths

- Database migrations / Ent schema.
- Frontend UI.
- Production secrets or deployment configs.

## Success Criteria

- Limits are configurable and can be disabled with zero or negative values.
- Existing request rate limits remain in place.
- Redis failures fail closed for registration risk checks.
- Register and send-verify-code paths are covered.
- Tests cover IP success limit, `IP + UA` short-window limit, email-domain short-window limit, disabled mode, and Redis failure behavior.

## Acceptance Commands

- `go test -tags=unit ./internal/server/routes ./internal/middleware ./internal/config -run "TestAuthRoutes|TestRegistrationRisk|TestConfig" -count=1`
- `go test -tags=unit ./internal/server/routes -run "TestRegistrationRisk" -count=1`
- `git diff --check`

### DONE: prompt-audit-policy-matrix-s293-r1-core

## Changed Files

- `backend/internal/securityaudit/prompt_repository.go`
- `backend/internal/securityaudit/prompt_qwen3guard.go`
- `backend/internal/securityaudit/prompt_rules.go`
- `backend/internal/securityaudit/prompt_scanner.go`
- R1 securityaudit regression tests

## Implemented

- Corrected `owasp_tags` / `config_version` INSERT placeholder mapping.
- Enforced Unsafe and unknown-category blocking floors in parser and common evaluator.
- Added action/risk, Unicode ID, rule-count, OWASP-count/length, and JSON-size validation.
- Restricted rule actions to Warn/Block and made matched-rule attribution escalation-aware.
- Made result aggregation compare decision, action, and risk consistently.

## Checks

- `go test ./internal/securityaudit -count=1` PASS
- `go test ./internal/server/routes ./internal/server/middleware ./migrations -count=1` PASS
- `go build ./...` PASS
- Repository-outside overlay reproductions for the five original failures PASS.
- Real PostgreSQL roundtrip BLOCKED: no dedicated PostgreSQL executable/DSN was available.

## Risks / Unverified

- Shared PostgreSQL/Redis and Qwen3Guard runtime were not used.
- Existing early-stop-on-Block remains; attribution is defined over evaluated chunks.
- No commit, push, deployment, migration, or shared data operation was performed.

## Continuation 2026-09-05

- Record-level history IDs are validated before use; empty/129-character records
  are excluded. Defaults/maps precede attribution and chunk ties use an unexported
  priority field followed by rule ID. Independent Terra exact-case retest PASS.
- Integration fixture now includes migration 239 and an explicit event-field
  roundtrip matrix. The new test compiles and SKIPs without a dedicated DSN.
- Latest evidence and release BLOCKED verdict:
  `docs/workflow/qa-reports/prompt-audit-policy-matrix-s293-remediation-continuation-qa.md`.

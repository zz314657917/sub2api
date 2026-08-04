# Upstream Email Ops Report S161

## Task ID

`upstream-email-ops-report-s161`

## Role

Primary Codex performs Planner, direct implementation, and final Evaluator gates in sequence. No worker is used for this isolated upstream behavior port.

## Goal

Adapt the structured daily/weekly Ops scheduled-report email behavior from upstream commit `361983cf4` onto the current notification-template topology. The existing report delivery, fallback, deduplication, and error/account-health HTML report paths must remain compatible.

## Success Criteria

- Daily and weekly scheduled reports expose a localized, structured metrics template with only explicitly allowed report placeholders.
- Error digest and account-health reports continue using their existing generated HTML through the report-detail section; no report is dropped merely because summary data is absent.
- The template runtime supplies deterministic display flags and safe placeholder defaults, so legacy customized templates and previews do not render stale sample report content in live emails.
- The editor shows the selected event's allowed placeholders when the API supplies them, while preserving its fallback list for incomplete responses.
- Focused service/frontend checks pass, or unrelated baseline failures are recorded separately. No migration, dependency, API/Wire, primary-worktree, merge, push, deployment, configured database, SMTP, or provider call is performed.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-email-alias-dedup-s157`
- Base: `20f54d50e`
- Upstream behavior source: `361983cf4` (`feat: 优化运维定时报表邮件模板`).
- The repository-wide P/G/E status belongs to concurrent Pixel Cafe S168 and must not be changed by this independent upstream-email task.
- The primary worktree `F:/mcplugins/sub2api` is dirty and remains untouched.

## Allowed Paths

- `backend/internal/service/notification_email_service.go`
- `backend/internal/service/notification_email_service_test.go`
- `backend/internal/service/ops_scheduled_report_service.go`
- `backend/internal/service/ops_scheduled_report_service_test.go`
- `frontend/src/views/admin/settings/EmailTemplateEditor.vue`
- `frontend/src/views/admin/settings/__tests__/EmailTemplateEditor.spec.ts`
- `docs/workflow/tasks/upstream-email-ops-report-s161.md`
- `docs/workflow/qa-reports/upstream-email-ops-report-s161-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `backend/migrations/**`, Ent schemas/generated code, `backend/go.mod`, `backend/go.sum`, dependencies, database configuration, API/DTO/Wire/routes, and production configuration.
- `knowledge/**`, `docs/workflow/status.md`, `docs/workflow/spec.md`, deployment, containers, the primary worktree, and unrelated frontend modules.
- SMTP transport, report scheduler frequency/eligibility, Ops data queries, email fallback/dedup policy, authentication/authorization policy, payment, billing, and provider protocols.

## Constraints

- Adapt behavior to the local S160 template topology; do not cherry-pick or blanket-merge upstream history.
- Preserve template rendering escape rules and keep `report_html` as the only raw HTML variable for the Ops scheduled-report event.
- Preserve current fallback and delivery semantics. Do not start a report scheduler or call an SMTP server.
- Keep the edit minimal and do not alter the dirty primary worktree.

## Acceptance Commands

```powershell
go test ./internal/service -run 'Test(NotificationEmail|OpsScheduledReport)' -count=1
go test ./internal/service -run 'Test.*(Email|Ops).*' -count=1
go test ./... -run '^$' -count=1
gofmt -d <changed Go files>
npx.cmd vitest run src/views/admin/settings/__tests__/EmailTemplateEditor.spec.ts
npm.cmd run typecheck
npx.cmd eslint src/views/admin/settings/EmailTemplateEditor.vue src/views/admin/settings/__tests__/EmailTemplateEditor.spec.ts
npm.cmd run build
git diff --check
git ls-files -u
```

## Output

- QA report: `docs/workflow/qa-reports/upstream-email-ops-report-s161-qa.md`
- Final verdict must be `PASS`, `FAIL`, or `BLOCKED`, with executed checks and unverified risks.

## Stop Rules

- Stop if the behavior requires a migration, dependency change, new API/Wire/routes, direct SMTP/report scheduler execution, or changes outside the allowed paths.
- Stop if the new template can render raw content from a variable other than the internally generated `report_html`, or if a legacy report type loses its fallback/template delivery path.
- Stop if validation requires a configured database, SMTP server, provider, deployment, or container mutation.

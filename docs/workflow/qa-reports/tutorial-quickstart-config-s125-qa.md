### PASS: tutorial-quickstart-config-s125

# QA Report

## Task ID
tutorial-quickstart-config-s125

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/tutorial-quickstart-config-s125.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:
```text
go test ./internal/service -run "QuickstartTutorial" -count=1 -> PASS
go test ./internal/handler -run "QuickstartTutorial" -count=1 -> PASS (no matching focused tests)
go test ./internal/handler/admin -run "QuickstartTutorial" -count=1 -> PASS (no matching focused tests)
go test ./... -run "^$" -> PASS
corepack.cmd pnpm --dir frontend exec vitest run src/views/public/__tests__/TutorialView.spec.ts src/api/__tests__/admin.tutorials.spec.ts -> PASS, 2 files / 9 tests
corepack.cmd pnpm --dir frontend run typecheck -> PASS
corepack.cmd pnpm --dir frontend run build -> PASS, 1100 modules transformed
git diff --check -> PASS
conflict-marker scan over S125 paths -> PASS, no markers
```
- manual checks:
```text
Playwright desktop /tutorial -> PASS: quick-start renders, Claude switch updates protocol and commands.
Playwright mobile 390x844 /tutorial -> PASS: no horizontal overflow observed; controls and code blocks remain readable.
Preview backend is the pre-change service, so /tutorial used the intentional local fallback config; the saved ai.3zapi.com flow is covered by the focused service and Vue tests, not by a live persistent API smoke.
```

## Follow-up: Base URL convention
- Quick-start defaults and `api_base_url` fallback now use the root URL for both
  ChatGPT / Codex and Claude. A configured `https://example.com/v1` is shown
  as `https://example.com`; the generated Codex cURL example calls
  `https://example.com/responses`.
- Regression checks after this correction:
```text
go test ./internal/service -run "QuickstartTutorial" -count=1 -> PASS
corepack.cmd pnpm --dir frontend exec vitest run src/views/public/__tests__/TutorialView.spec.ts -> PASS, 1 file / 8 tests
go test ./... -run "^$" -> PASS
corepack.cmd pnpm --dir frontend run typecheck -> PASS
git diff --check -> PASS
```
- The ChatGPT / Codex install step now includes the fixed official ChatGPT
  Desktop guide URL. It opens in a new tab with `noopener noreferrer`; the
  link disappears when the user selects Claude.
- The configuration-location step now states the full default `config.toml`
  path for each Codex terminal, creates and opens the Windows directory, and
  warns about the `.txt` extension. Claude correctly skips the `.codex`
  directory and explains its environment-variable configuration instead.

## Findings
- 未发现明确问题。
- Vite emitted pre-existing chunk-size and dynamic-import warnings during the
  production build; the build completed successfully and this task does not
  change chunking configuration.

## Bug Owner Recommendation
`none`

## Root Cause
`none`

## Retest Scope
- Authenticated administrator save/reset against a running S125 backend with
  a real settings store, followed by a public `/tutorial` refresh.

## Knowledge Promotion
- `none`

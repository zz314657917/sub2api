# Pelican preview reload regression verification

## Findings

Conditional HTML requests returned 304 using a template ETag, while the response CSP contained a fresh nonce. Browsers reused HTML carrying the previous nonce, blocked the preview measurement script, and left all previews at their initial 220px height.

Dynamic HTML now returns 200 with the current nonce and Cache-Control: no-store. The server-side template cache and static asset handling remain unchanged. The patch is limited to the embedded HTML handler and its regression tests.

## Executed checks

- Go embedded web package tests passed with `go test -tags embed ./internal/web -count=1`.
- Isolated Pelican frontend tests: 43 passed; typecheck, scoped ESLint, Vite production build and Linux embedded server build passed.
- The local deployment returned matching frontend asset hashes, healthy status and zero restarts. Environment variables and data mounts were preserved.
- Five consecutive browser reloads returned 200/no-store. Preview heights remained 257, 497 and 580px; no CSP or script errors occurred.
- At 390px viewport width, the gallery used one column with document scrollWidth=390; preview heights were 287, 536 and 580px. The updated plan overview was present.
- The task-owned browser was closed; matching browser/profile and CLI daemon process counts were zero. The Docker update guard was released, and the previous container was retained for rollback.

## Scope and limits

The local image contains the prior deployed application plus the plan panel update and nonce fix. It is not a deployment of unrelated concurrent billing commits. No real paid-provider requests were initiated by these checks. The 580px preview cap remains enabled.

## Recommendation

Accepted for the local deployment and source publication. Evidence was collected on 2026-09-23.

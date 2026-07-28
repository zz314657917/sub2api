### PASS: upstream-async-image-tasks-s123

## Findings

- No implementation defect was found in the approved source scope.
- The admin image-storage endpoints are protected by the existing Backup
  `AdminAuth` route group and preserve encrypted secret handling. The baseline
  does not provide a separate generic password/TOTP step-up guard for existing
  backup S3 settings; this is a residual security-design risk, not a new
  bypass introduced by S123.
- The local `sub2api-redis` server accepts unauthenticated connections, while
  its container environment defines `REDISCLI_AUTH`. The stock CLI therefore
  attempted an invalid `AUTH`; the smoke used a one-process unset of that
  CLI-only variable. No secret value was read or displayed, and application
  Redis configuration was not changed.

## Executed Checks

- `go test ./internal/config -run 'Test.*ImageStorage' -count=1`
- `go test ./internal/service -run 'TestImageTask|TestImageStorage|TestImageStorageSetting' -count=1`
- `go test ./internal/handler -run 'TestAsyncImageHandler|TestImageTaskAdminToggle' -count=1`
- `go test ./internal/repository -run 'Test.*Image(Task|Storage)' -count=1`
- `go test ./internal/server/middleware -run 'TestIsAsyncImageTaskRead' -count=1`
- `go test ./internal/server/routes -run 'TestGatewayRoutes' -count=1`
- `go test ./... -run '^$'`, `go mod verify`, and `go test ./cmd/server -run '^$'`
- `corepack.cmd pnpm --dir frontend run typecheck`
- `corepack.cmd pnpm --dir frontend run build`
- Wire regeneration, `gofmt -d`, `git diff --check`, conflict-marker scan,
  and exact allowlist audit.
- Local Redis runtime smoke against the running `sub2api-redis` container:
  `PING`, a unique-key `SET EX 60`, `TTL` (60 seconds), `GET`, exact `DEL`,
  and post-delete `EXISTS` (0) all passed. No existing Redis data was read,
  overwritten, or flushed.

## Unverified Risks

- No real S3 upload/presigned URL, real upstream image response, authenticated
  API-key/admin UI session, deployment, or container refresh.
- The S3 configuration action validates configuration/client construction only;
  it deliberately does not contact a real bucket in this source-only task.
- Redis storage primitives passed, but this does not prove the application
  configuration can connect to the same container or that the asynchronous
  image task flow works end to end.

## Recommendation

Source-only acceptance passes. Continue only with an isolated, non-production
runtime smoke that uses a real Redis and disposable S3-compatible bucket before
enabling the feature for users. Do not treat this result as deployment approval.

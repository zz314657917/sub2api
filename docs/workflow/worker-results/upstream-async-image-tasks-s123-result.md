### PASS: upstream-async-image-tasks-s123

## Changed Behavior

- Adds opt-in `POST /v1/images/generations/async`,
  `POST /v1/images/edits/async`, and owner-scoped
  `GET /v1/images/tasks/:task_id` routes.
- Reuses the local OpenAI image gateway in a bounded background context.
  Redis keeps task metadata only; result images are re-hosted in S3-compatible
  storage before a task is completed.
- Adds file/environment defaults and encrypted hot admin settings, including a
  Backup-page configuration card. The user-console `image_creator` path and
  synchronous Images endpoints are unchanged.

## Compatibility Decisions

- The local baseline has no matching Grok Images gateway handler, so async
  submission is limited to `PlatformOpenAI` rather than creating a new Grok
  execution path.
- A disabled store rejects new submissions but leaves polling active for Redis
  tasks that have not expired.
- Image URL re-hosting is HTTPS-only, bypasses proxy variables, limits
  redirects and size/time, and verifies resolved addresses before dialing.

## Evidence

- Focused config, service, handler, repository, middleware, and route tests
  passed; the repository regression runs against `miniredis`.
- `go test ./... -run '^$'`, `go mod verify`, `gofmt -d`, Wire regeneration,
  frontend typecheck, frontend production build, diff check, conflict-marker,
  and allowlist checks passed.

## Risks

- No live Redis, S3-compatible bucket, real OpenAI image request, or
  authenticated admin browser session was used.
- Admin image-storage settings inherit the existing Backup route's `AdminAuth`.
  This baseline has no separate general step-up middleware for S3 settings.

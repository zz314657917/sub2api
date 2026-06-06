### DONE: upstream-main-hardening-s3

## Summary
- Created branch `codex/upstream-main-hardening-s3` from local `main`.
- Added Sprint contract `docs/workflow/tasks/upstream-main-hardening-s3.md`.
- Ported all 8 approved upstream hardening commits by cherry-pick with minimal local conflict adaptation.

## Applied Commits
- `0042a1d18` from upstream `0ae332961`: escape API Key names with `html.EscapeString`.
- `74aab9e9b` from upstream `11b601717`: return `404` for unauthorized API Key access.
- `ec89a5ddb` from upstream `585ff0944`: Redis Lua `TIME` compatibility.
- `1e10208ed` from upstream `8a56c9fa0`: Postgres setup maintenance DB bootstrap.
- `b97859413` from upstream `04deb819b`: EasyPay query uses `trade_status` first.
- `5cbee92bd` from upstream `bba86f97d`: `userRepo.Delete` reuses caller transaction.
- `0836a8400` from upstream `705fe7d88`: admin user deletion deletes user API Keys.
- `d965f4900` from upstream `fb0195f3d`: normalize fixed quota windows on account edit.

## Conflict Adaptation
- `0ae332961`: preserved local API Key create fields and only changed `Name` to `html.EscapeString(req.Name)`.
- `04deb819b`: preserved local EasyPay GET-first query, POST fallback, and failed-response error handling; moved upstream `trade_status` priority into the existing parsed response path.
- `705fe7d88`: added `DeleteWithAudit` to the service interface and reused transaction client for audit deletion existence checks.

## Contract Compliance
- No direct merge of `upstream/main`.
- No changes to denied frontend, Ent schema, migrations, deploy, assets, README, knowledge, `docs/workflow/status.md`, or `docs/workflow/spec.md`.
- Deferred gateway, WS, Images, Responses, apicompat, DingTalk, email, and user-platform quota work remains out of this Sprint.

## Checks Run
- `git status --short --branch`
- `git diff --check`
- `go test ./internal/payment/provider -run "EasyPayQuery" -count=1`
- `go test ./internal/service ./internal/repository -run "DeleteUser|DeleteWithAudit|DeleteOwned|APIKey.*Delete|UserRepo.*Delete" -count=1`
- `go test ./internal/service ./internal/handler ./internal/repository ./internal/payment/provider ./internal/setup -run "APIKey|User|Delete|Redis|Concurrency|Session|EasyPay|Setup|Quota" -count=1`
- `go test ./internal/service ./internal/handler ./internal/repository -count=1`
- `go test ./internal/server/routes ./cmd/server -count=1`

## Risks
- No runtime container smoke was requested or run for this backend-only Sprint.
- EasyPay behavior intentionally keeps local failed-query responses as errors instead of adopting upstream's old pending fallback for `code=0`.

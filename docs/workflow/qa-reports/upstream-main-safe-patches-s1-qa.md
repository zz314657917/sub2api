### PASS: upstream-main-safe-patches-s1

# upstream-main-safe-patches-s1 QA Report

## Task ID
upstream-main-safe-patches-s1

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-safe-patches-s1.md`

## Evidence
- diff reviewed: yes
- allowed paths checked: yes
- denied paths touched: no
- commands run:
```text
git status --short --branch -> clean on codex/upstream-main-safe-patches-s1
git diff --name-status main..HEAD -> only allowed backend/service/handler/middleware/tests and workflow docs
git diff --check -> pass
go test ./internal/service -run "OpenAI|Codex|Ops|APIKey|Gateway|Billing|AccountUsage|Metrics|Ratelimit" -count=1 -> pass
go test ./internal/service ./internal/handler ./internal/repository -count=1 -> pass
go test ./internal/server/routes ./cmd/server -count=1 -> pass
corepack.cmd pnpm install --frozen-lockfile (frontend/) -> pass
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops/components/__tests__/OpsSettingsDialog.spec.ts -> pass
corepack.cmd pnpm --dir frontend run build -> pass
docker build -f tmp/Dockerfile.runtime-prebuilt -t sub2api:s1-runtime-smoke . -> pass
temporary Docker runtime smoke with Postgres/Redis/app containers -> /health returned {"status":"ok"}
```
- manual checks:
```text
Sprint 1 scope audit -> no Ent schema/codegen/migrations, no deploy changes, no README/assets/knowledge sync, no bridge redesign
temporary Docker containers/network -> cleaned up
```

## Findings
- 未发现当前 Sprint 1 补丁集引入的明确阻断问题。
- 后端核心验收命令全部通过。
- 前端 `typecheck`、`lint:check`、目标 Vitest 和生产 build 通过。
- Docker runtime 镜像构建通过，临时 Postgres/Redis/app 容器启动后 `/health` 返回 `{"status":"ok"}`。

## Bug Owner Recommendation
none

## Root Cause
- none

## Retest Scope
- None.

## Knowledge Promotion
- none

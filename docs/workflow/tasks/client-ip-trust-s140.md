---
task_id: client-ip-trust-s140
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Task ID

`client-ip-trust-s140`

## Role

Generator 只按本合同把上游 forwarded-client-IP 行为适配到当前架构；Evaluator
独立审查合同和验收证据。不得把上游提交历史或文件布局直接带入本地。

## Goal

在当前 `main` 的 HTTP、认证和安全审计边界内，建立可配置且 fail-closed 的
forwarded client IP 来源链。让 API-key ACL、Google/Gemini API-key ACL、session
binding 和操作审计在同一请求内使用一致的安全 IP，同时保留未显式启用时对原始
转发头的不信任默认值。

## Success Criteria

- 默认配置为 `false` 时，未被 Gin `TrustedProxies` 验证的原始转发头不会进入
  ACL、session binding、审计或其他安全决策；显式启用才允许配置的转发来源。
- `server.trusted_proxies` 继续支持显式空列表并保持 fail-closed；不会因旧设置
  或缺少代理配置自动打开不安全的 raw forwarded trust。
- 自定义 forwarded header 名称经过 trim、合法性校验、大小写规范化和去重，最多
  16 个；非法、空白、重复或超限输入被拒绝并有可测试的错误结果。
- 每个请求捕获一次 forwarded-IP 相关设置快照；同一请求后续中间件、ACL、审计和
  session binding 只读取该快照，不因并发设置热更新而前后使用不同解析模式。
- 主 API-key 鉴权路径与 Google/Gemini API-key 路径具有相同的 IP ACL 语义、拒绝
  状态和审计来源；未配置 ACL 时保留现有行为。
- 审计记录、session IP/User-Agent binding 与 API-key ACL 共享同一个安全客户端 IP
  helper；不得有调用方退回未经验证的 `GetClientIP` 原始头解析。
- 相关单元/handler/service 回归、格式、编译、冲突标记、未合并索引和允许路径
  检查通过；真实代理、生产设置、部署和容器行为明确列为未验证边界。

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/client-ip-trust-s140`
- Baseline: clean `origin/main` worktree；主工作树存在用户脏改动，不能直接操作。
- Relevant upstream behavior is distributed across trusted-proxy initialization, legacy
  toggle migration, settings model/persistence/audit, API contract/frontend types,
  request snapshot, session binding, audit, and both API-key authentication paths.
  These commits are reference material only and must be manually adapted.

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/pkg/ip/ip.go`
- `backend/internal/pkg/ip/ip_test.go`
- `backend/internal/server/http.go`
- `backend/internal/server/http_ingress_test.go`
- `backend/internal/server/router.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_google.go`
- `backend/internal/server/middleware/api_key_auth_test.go`
- `backend/internal/server/middleware/api_key_auth_google_test.go`
- `backend/internal/server/middleware/session_binding.go`
- `backend/internal/server/middleware/session_binding_test.go`
- `backend/internal/server/middleware/audit_log.go`
- `backend/internal/server/middleware/audit_log_test.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/setting_service_update_test.go`
- `backend/internal/service/setting_service_partial_payload_test.go`
- `backend/internal/service/wire.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/admin/setting_handler_partial_payload_test.go`
- `backend/internal/handler/dto/settings.go`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `deploy/config.example.yaml`
- `README.md`
- `README_CN.md`
- `docs/workflow/tasks/client-ip-trust-s140.md`
- `docs/workflow/qa-reports/client-ip-trust-s140-qa.md`
- `docs/workflow/worker-results/client-ip-trust-s140-result.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, new tables, schema changes, data backfills,
  or destructive persistence changes
- `backend/internal/repository/**`, generated Wire output, and any handler/service/
  middleware file not explicitly listed above
- unrelated authentication, billing, routing, provider, passkey, prompt-audit, or
  frontend feature work
- `deploy/**`, `Dockerfile*`, `docker-compose*.yml`, production configuration, live
  settings, deployment, container refresh, or external runtime changes
- `knowledge/**`, `knowledge/tasks/current-task.md`, and
  `C:/Users/Administrator/.codex/memories/**`
- `outputs/**` and all existing dirty paths outside this worktree
- direct cherry-pick/merge of the upstream commit chain, history rewriting, commit,
  push, or changes to the primary worktree

## Constraints

- Security default is `false`. Do not adopt the upstream default that trusts raw
  forwarded headers, and do not auto-enable it while migrating a legacy toggle when
  trusted proxies are absent or unspecified.
- Preserve explicit `server.trusted_proxies: []` semantics and fail closed. Proxy
  validation must remain centralized at the existing HTTP/Gin boundary.
- Resolve a request-level immutable settings snapshot before security consumers run;
  do not read mutable global settings independently in later middleware or handlers.
- Accept only valid HTTP header field names after trimming; normalize comparisons
  case-insensitively, deduplicate deterministically, and enforce the 16-header maximum.
  Never log raw credential or token values while reporting validation failures.
- Use one security client-IP helper for ACL, audit and session binding. Preserve the
  existing trusted-client behavior when the new option is disabled and keep malformed
  proxy chains fail-closed.
- Apply equivalent ACL behavior to the primary API-key and Google/Gemini API-key paths;
  do not weaken existing authentication, status codes, or account selection semantics.
- Adapt behavior to local file boundaries and tests; upstream commits are evidence, not
  a merge plan. Keep changes minimal and do not reformat unrelated files.

## Acceptance Commands

Run the Go commands from `backend`, and the frontend/Git commands from the worktree
root (set PowerShell output to UTF-8 first):

```powershell
Push-Location backend
go test ./internal/pkg/ip -count=1
go test ./internal/config -run "Test.*Trusted.*Proxy|Test.*Forwarded.*IP|Test.*Client.*IP" -count=1
go test ./internal/server/middleware -run "Test.*(IP|Session|Audit|APIKey)" -count=1
go test ./internal/handler/admin -run "Test.*(APIKey|Gemini|Google|Settings|Audit|Session)" -count=1
go test ./internal/service -run "Test.*(APIKey|ACL|ClientIP|Forwarded|Session|Audit)" -count=1
go test ./... -run '^$'
Pop-Location

corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts
corepack.cmd pnpm --dir frontend exec eslint src/api/admin/settings.ts src/views/admin/SettingsView.vue src/views/admin/__tests__/SettingsView.spec.ts src/i18n/locales/en/admin/settings.ts src/i18n/locales/zh/admin/settings.ts
corepack.cmd pnpm --dir frontend run build
git diff --check
git ls-files -u
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" .
git diff --name-only HEAD
```

Required review checks:

- prove default-disabled and explicit-enabled cases with forged headers and trusted
  proxy chains;
- prove custom-header trim/validation/normalization/deduplication and the 16-item limit;
- prove a request keeps one settings snapshot across a simulated hot update;
- prove primary API-key and Google/Gemini ACL parity, plus one shared source assertion
  for audit and session binding;
- audit the final path list against Allowed/Denied Paths and record any pre-existing
  unrelated failures without claiming a green full suite.

## Output

- Implementation and focused regressions only within the allowlist.
- `docs/workflow/worker-results/client-ip-trust-s140-result.md`, whose first line is
  `### DONE: client-ip-trust-s140`, `### BLOCKED: client-ip-trust-s140`, or
  `### FAILED: client-ip-trust-s140`; include changed files, commands, concise output,
  risks, and knowledge candidates.
- `docs/workflow/qa-reports/client-ip-trust-s140-qa.md` with `PASS`, `FAIL`, or `BLOCKED`
  on the first line and the evidence categories `Findings`, `Executed Checks`,
  `Unverified Risks`, and `Recommendation`.

## Stop Rules

- Stop and request evaluator/owner decision if the implementation requires a migration,
  schema rewrite, production setting, deployment/container action, or a broad auth/
  routing redesign.
- Stop if safety would require trusting raw forwarded headers by default or silently
  migrating an unset/ambiguous legacy toggle to an enabled state.
- Stop if a caller cannot be moved to the shared security-IP source without changing an
  unrelated business contract, or if API-key ACL parity would require weakening auth.
- Stop if an upstream commit cannot be behaviorally adapted without importing denied
  paths; do not cherry-pick around the boundary.
- Stop on out-of-scope diff paths, unresolved merge entries, or repeated test failures
  that cannot be isolated to this contract. Do not commit, push, deploy, or modify the
  primary dirty worktree.

## Contract Review

`PASS`: the scope is limited to existing configuration, settings, middleware, and
administrator UI seams; no schema or generated-code change is allowed. The contract
keeps raw forwarded-header trust disabled by default, rejects the upstream automatic
legacy enablement, requires request snapshots and Google/Gemini ACL parity, and has
executable focused backend/frontend and integrity checks. Implementation may proceed
in this isolated worktree; publication remains a separate gate.

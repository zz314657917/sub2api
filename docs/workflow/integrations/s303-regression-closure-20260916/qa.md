### PASS: s303-regression-closure

# QA Report

## Task ID
s303-regression-closure

## Verdict
`PASS`

## Contract Checked

- `docs/workflow/tasks/s303-regression-closure.md`
- `docs/workflow/contract-reviews/s303-regression-closure-review.md`
- `docs/workflow/worker-results/s303-regression-closure-result.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`（本 QA 只新增本报告；控制器已有未提交的 `docs/workflow/main-log.md` 与 `docs/workflow/status.md` 未改动。）

### Commands run

```text
backend: go test -tags unit ./internal/server/middleware -list 'TestS303'
  -> PASS; selected TestS303JWTAuthLookupOutcomes and TestS303AdminAuthLookupOutcomesHTTPAndWebSocket.
backend: go test -tags unit ./internal/server/middleware -run 'TestS303|TestJWTAuth|TestAdminAuth' -count=1
  -> PASS.
backend: go build ./...
  -> PASS.
frontend: npm.cmd run test:run -- src/api/__tests__/client.spec.ts src/api/__tests__/tokenRefresh.spec.ts
  -> PASS; 2 files, 28 tests (client 21, tokenRefresh 7).
frontend: npm.cmd run typecheck
  -> PASS.
frontend: npm.cmd run build
  -> PASS; exit 0, Vite built in 21.64s. Dynamic-import/chunk-size warnings only.
root: gofmt -l backend/internal/server/middleware/s303_auth_lookup_test.go
  -> PASS; no output.
root: git diff --check
  -> PASS; no findings.
root: git ls-files -u
  -> PASS; no unmerged entries.
```

### Manual checks

- Backend new coverage is selected by the `unit` build tag and uses real JWT generation, Gin `httptest` request/handler flow, and the existing `jwtAuth` / `adminAuth` wrappers, which delegate to the production session-binding middleware.
- The two new `TestS303` roots exercise 17 new subcases: JWT lookup has active/direct-wrapped-not-found/timeout/internal outcomes (5); Admin has six outcomes through both HTTP `Authorization` and WebSocket `Sec-WebSocket-Protocol: jwt.<token>` extraction (12). Every rejected path asserts that the protected handler is not reached.
- Existing selected JWT/Admin tests remain present in the verbose run, including invalid/tampered JWT and token-version rejection cases; no existing test was changed by this task.
- Frontend cases drive the actual imported `apiClient` interceptor with Axios adapter/post fakes. Network (status 0), 429, 500 and 503 retain all four auth storage keys plus the pre-existing session-storage value, write no `auth_expired`, and make zero `href` assignments.
- Refresh 401/403 and malformed-success paths clear the invalid session and assert the `/login` setter call. Changed-session network failure returns `AUTH_SESSION_CHANGED` before transient classification and preserves the replacement session including its exact `token_expires_at` value.
- `afterEach` restores the original `window.location` descriptor and `navigator.locks`; each case starts with cleared local/session storage. Redirect assertions use an explicit location setter spy, not jsdom's unimplemented navigation behavior.

## Findings

未发现明确问题。测试变更严格限于：

- `backend/internal/server/middleware/s303_auth_lookup_test.go`（新增、未跟踪）
- `frontend/src/api/__tests__/client.spec.ts`（已修改）
- `docs/workflow/qa-reports/s303-regression-closure-qa.md`（本报告）

未暂存变更中的 `docs/workflow/main-log.md` 与 `docs/workflow/status.md` 是控制器已有流程文档，不属于此开发/QA测试范围，未被本 QA 改动。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

不需要修复重测。若后续触及鉴权实现，至少重跑本合同的后端选择/回归命令和两份前端定向测试。

## Runtime Gaps

本结论仅覆盖本地 fake/单元回归、编译、类型检查与前端生产构建。未执行真实数据库故障注入、外部 provider、容器、已认证浏览器/UI、部署、提交或 push；这些仍是路线图中的独立运行态门禁，不能由本报告替代。

## Knowledge Promotion

`none`

### PASS: upstream-main-integration-s82-s86

## Findings

- No blocking integration finding remains across the local Usage S82 change
  and upstream compatibility S82-S86.
- The integration history contains four explicit merge commits for S82, S83,
  S84, and S85-S86. S86 is stacked on S85, so a separate S85 merge would be
  redundant.
- All 22 business paths match their owning source commit after integration.
  The remaining 21 paths are the expected workflow status/spec/log files and
  18 task, worker-result, and QA artifacts.
- The refreshed remote baseline remains `origin/main` `37e0b493c`; it is an
  ancestor of the integration head. The refreshed `upstream/main`
  `db4295d646` is audit input only and is not merged by this release.

## Executed Checks

- Combined frontend Vitest: PASS, 7 files / 55 tests, covering WS mode copy,
  subscription expiry minute precision, Usage reasoning effort, UseKeyModal,
  and the Channel Status baseline.
- Frontend `npm.cmd run typecheck`: PASS.
- Frontend `npm.cmd run build`: PASS, 1088 modules transformed.
- Broader service regressions for Anthropic forwarding/buffered content type
  and proxy-quality/Grok behavior: PASS.
- Broader handler `TestHandleFailoverError_` regressions: PASS.
- All five source commits are ancestors of the integration head; all 22
  business blobs match their owners and all 18 expected workflow artifacts
  exist.
- Exact 43-path audit, unmerged-index scan, real conflict-marker scan, and
  `git diff --check origin/main..HEAD`: PASS.

## Unverified Risks

- No authenticated browser smoke or live Anthropic, xAI, OpenAI, billing, or
  proxy request was run. Focused local regressions and compile/build checks
  cover the changed behavior.
- Existing Browserslist, Vite dynamic-import/chunk-size, and Node `DEP0190`
  warnings remain non-blocking.
- No deployment or container update was performed. This report does not claim
  parity with the current upstream branch beyond the reviewed S82-S86 ports.

## Recommendation

`PASS / publish-ready`. Fast-forward local `main`, push `origin/main`, and
verify the remote SHA under the user's explicit authorization. Keep deployment
and container operations out of scope.

# Worker Result: upstream-main-compat-s79

## Verdict

`IMPLEMENTED — ready for independent evaluation`

## Baseline and Scope

- Local baseline: `e50c512746f67e4b916b596c28b4f8010fd87dc5`
- Upstream snapshot: `d4b9797ff72024960a035cf22fdd8f213e149169`
- Behavior references: `72b29a7d4`, `1a855d3e0`, `2264a3308`, `c3f759d13`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s79`
- No upstream history was merged; every behavior was adapted to local topology.

## Implemented Behavior

- Antigravity subscription normalization now preserves known paid `Pro` and
  `Ultra` plan types when the response contains `IneligibleTiers`, while
  retaining abnormal status and the first reason. Free, unknown, and absent
  tiers remain `Abnormal`; nil responses remain `Free`.
- Channel monitor adapters support an optional response extractor. Anthropic
  monitoring joins all trimmed non-empty `content[]` text blocks in order and
  ignores thinking/tool/malformed blocks; other providers still use their
  existing gjson path.
- Anthropic parsing removes repeated, case-insensitive trailing `[1m]`
  selectors only when a non-empty base model remains. It updates both
  `ParsedRequest.Model` and the local `[]byte` `ParsedRequest.Body` before
  parsing system/messages.
- One shared handler helper returns the parsed request and normalized body.
  `GatewayHandler.Messages`, `GatewayHandler.CountTokens`, and
  `OpenAIGatewayHandler.CountTokens` all overwrite their local body with that
  result before downstream Claude detection, Ops/moderation, mapped-body, and
  session-hash consumers.
- Existing payment locale keys remain unchanged; their English and Chinese
  label/error values are now unit-neutral.

## Test Topology Amendment

The repository's legacy `gateway_request_test.go` and
`channel_monitor_checker_body_test.go` use `//go:build unit`. To prevent a
default-tag zero-test pass, S79 added dedicated default-tag test files and kept
both legacy files unchanged. Both contract amendments passed independent
review.

## Executed Checks

- Default-tag service discovery found exactly the three required S79 tests:
  `TestS79NormalizeAntigravitySubscription`,
  `TestS79ExtractAnthropicMonitorText`, and
  `TestS79ParseGatewayRequestClaudeCodeLongContext`.
- `go test ./internal/service -run '^TestS79' -count=1`: PASS.
- Default-tag handler discovery found
  `TestS79AnthropicNormalizedBodyPropagation`.
- `go test ./internal/handler -run '^TestS79AnthropicNormalizedBodyPropagation$' -count=1`: PASS.
- `npm.cmd run test:run -- src/i18n/__tests__/paymentValidityLocales.spec.ts`:
  PASS, 1 file / 2 named tests.
- `npm.cmd run typecheck`: PASS.
- `npm.cmd run build`: PASS with existing non-blocking Vite dynamic-import and
  chunk-size warnings.
- The first Go attempt hit the host's inaccessible default `GOCACHE`; all
  reported Go checks were rerun successfully with isolated workspace-local
  `GOCACHE` and `GOTMPDIR`.

## Not Performed

- No live Anthropic, Antigravity, or payment browser smoke.
- No push, deployment, container update, VERSION change, migration, Ent
  generation, dependency update, or merge into local `main`.
- No `knowledge/**` or handoff/timeline write.

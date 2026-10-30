### PASS: model-unavailable-failover-20261030

2026-10-30 final local acceptance. Exact eight-file R3 source is integrated in
F:/mcplugins/sub2api and freshly verified there. No task commit, push or deployment.
The historical R1 report below is superseded and retained as failure evidence.

## Findings

- No remaining feature blocker found in the frozen and integrated source.
- OpenAI HTTP authoritative missing-model400/404, including the reported Luna
  unsupported-by message, switches accounts before client output. The rejected
  account/model pair receives the existing30-minute cooldown under normal
  persistence policy; deterministic rejection never retries that same account.
- Actual upstream model keys and necessary scheduler lookup mirrors are kept
  consistent without applying non-idempotent model mapping twice.
- Model-cooled accounts are filtered before scheduler top-K truncation.
  A stale sticky binding cannot hide a healthy account at top-K1; seven cooled
  candidates cannot hide the healthy eighth account at the default top-K7.
- Other models remain eligible. Parameter/capability/endpoint negatives, explicit
  conflicting codes, image-driver guards, post-output replay barriers,
  passthrough rules, pool/custom opt-outs, usage and release ownership remain.
- One pre-existing WS test still fails in both tags:
  TestOpenAIResponsesWebSocket_FailoverOnUpstreamUsageLimitEvent at
  openai_gateway_handler_test.go:1506, StatusPolicyViolation,
  reason "API key group is unavailable". Baseline, independent QA and fresh main
  have the same assertion/meaning. Neither handler group is claimed wholly green.

## Executed Checks

Independent QA and fresh main use Go1.26.10, local toolchain and readonly/offline
dependencies. Counts are Go test events including parent tests.

| Check | Independent frozen QA | Fresh main |
| --- | --- | --- |
| Service contract default/unit | 108/117 PASS | 108/117 PASS |
| Handler contract default/unit | each145 PASS + same1 original WS FAIL | same |
| Model-limit/sticky/scheduler/image cross-unit | 225 PASS | 225 PASS |
| Complete handler/Forward/loopback probe default/unit | each30 PASS | exact QA source preserved |
| Controller screenshot404 three owners | controller candidate PASS | 4 PASS events, three owners |
| Full backend build, gofmt, scoped diff-check | PASS | PASS |
| Baseline inventory/outcomes | no missing/new failure/skip | same comparison PASS |

The unchanged priority four-case oracle passes. The same v3 handler probe first
reproduced top-K1/default7 failures on R2, then passes on R3: actual HTTP trace
[A,B,B], second request200, successful-account usage only, every acquired slot
released and upstream body closed. This did not enlarge test top-K to pass.
The31 policy/replay checks per tag are service-owner tests; they are not claimed
as a complete external handler policy matrix.

Source and evidence bindings:

- Original approved contract: da964c460e0b768d12b136ab5821c547bd037532b0ec181a771eba1b9199a6f3.
- Scheduler amendment: dc13191e86bf9b0cd9c20f4b9e3b412234597110c14654db14701cdbeef4d611.
- Source freeze8: 028480614bbced62ba62389a37ae5271e93d4d3fcdd2ab9630f3c1ff19fb40c0.
- Independent QA seal: f2ed28d1b2ff605d6c49023b0b526e2590e52613032102a81e3efcbb0127fb36.
- QA report: 2203c67131d9cfb61da28c5cea58daebec8e52f7f45833668cb85509012fa037.
- Controller verified all184 sealed artifacts, both2731-file backend trees,
  all eight task hashes and all four unchanged R2 production hashes.
- Whole-candidate audit4809 files: only eight approved differences.

Protected integration and final full-tree audit pass for12250 original paths,
with the four explicit workflow exclusions and eight independently observed
frontend byte snapshots. Concurrent commits132e980b, f88b54e0 and424b96a6 are
preserved; this task did not write HEAD/index or stage files. Current main
HEAD424b96a6e9602d2922a1b8f524644ca2a427fa5b, index
42a661f6910a5ca29dc2cb3713102bd7f9ead91f97ef10279c924d00db8e5c7c,
staged tree empty.

R8 protection2d4e84cf/observation3e4f4eae independently reviewed PASS.
Its actual identity block passes11 synthetic Git controls, including same-byte
commit/index-metadata acceptance and rejection of staged changes, deletion,
extra paths, intermediate unobserved content, backend-change/revert, merge and
non-descendant history. Old drift and R6/R7 review FAILs remain unchanged.

Evidence root: E:/codex-runtime/sub2api/model-unavailable-failover-20261030.
Key records: qa/qa-result.md, qa/qa-r3-seal.json,
controller-qa-r3-seal-verification.json, controller-r3-whole-candidate-source-audit.json,
main-r3-integration-r2-protection.json, main-r3-verification.json,
main-r3-comparison.json, main-r3-final-protection.json, controller-final-result.json,
reconciliation/review-r4.md and reconciliation/identity-probe-r3-results.json.

## Unverified Risks

No real provider, application database, production account configuration,
container/deployment or race-instrumentation validation. No whole-repository
all-tests-green claim. Explicit pool/custom persistence opt-outs remain:
current-request switching still works, but persistent next-request cooldown
is not promised for those opt-outs or failed persistence.

## Recommendation

PASS for local source adaptation. All required local gates are complete.
Commit, push, container update and deployment remain outside this task.

## Historical R1 report — superseded, original failure retained

### FAIL: model-unavailable-failover-20261030

## Findings

First frozen candidate needs a classifier correction before integration:

- Exact `unknown model` HTTP404 loses the legacy missing-model behavior.
- `unknown provider for model` HTTP404 lacks the relevant upstream matching.
- HTTP400 `Model 'gpt-6-luna' is not supported for streaming` and the equivalent
  tools rejection are request capability errors, but incorrectly switch and
  persist a model cooldown.

Independent QA reproduced all four cases through an external handler probe.
The original Developer owns the helper/new-test-only correction; original r1
source, probes and failure output remain preserved. No product integration yet.

## Executed Checks

- Independent contract review PASS on draft826f640c, followed by independent
  metadata-only binding to approved da964c46.
- Immutable baseline carries nine exact existing dirty backend prerequisites.
- Baseline focused service: default21/unit30 pass events, zero failures/skips.
- Baseline handler focused: each tag145 pass events and one original
  `TestOpenAIResponsesWebSocket_FailoverOnUpstreamUsageLimitEvent` failure.
- Baseline model-limit/sticky/scheduler/image cross-unit:219 pass events,
  zero failures/skips. Baseline full backend build PASS.
- Controller's original-code HTTP404 regression probe fails in all three
  Responses/Chat-compat/automatic-passthrough error owners.
- Developer r1 focused default102/unit111 and backend build PASS, zero skips;
  these are Developer results and do not replace independent acceptance.
- Candidate scope freeze includes seven allowed files. Pre-QA protection
  confirms4805 tracked plus7445 untracked original files and HEAD/index.
- Controller's same three-owner screenshot probe is green on r1.
- Independent r1 priority handler probe: four failing cases listed above.
- Scope review: no existing tests, dependencies or frozen prerequisites changed.

## Evidence

- Root: E:/codex-runtime/sub2api/model-unavailable-failover-20261030.
- `primary-before.json`, `primary-untracked-before.json`, `baseline-source.json`.
- `baseline-test-counts.json`, baseline logs and `baseline-summary.json`.
- `contract-review-r1.md`, `contract-approval-binding.md`.
- `source-freeze-r1.json`, `developer-result.md`, `developer-r1-summary.json`.
- `candidate-controller-probe-r1.jsonl`.
- Independent QA: `qa/frozen-r1`, `qa/probes/priority_probe_test.go`,
  `qa/logs/candidate-r1-priority-default/summary.json`.

## Unverified Risks

R2 correction and independent final acceptance remain pending. No real provider,
live account configuration, production database, container or deployment test.
Pool/default and custom exclusions intentionally preserve no-persistence policy;
write failures cannot establish subsequent-request cooldown.

## Recommendation

Do not integrate r1. Retest the same independent failures on a new frozen
candidate, then finish handler/policy/output/billing checks and exact protected
integration. Preserve the original WebSocket baseline failure separately.

# Native port recovery work order

Controller deep review, 2026-09-16. This is a sequencing aid for the approved
image2-native-port contract, not a replacement or reduced acceptance scope.
The parent contract and independent PASS review remain authoritative.

## Current evidence

- Generator stopped without a complete result twice; broad worker loop stopped.
- Controller reran the parent focused selector: PASS (0.171s). This does not
  prove new native behavior: no native test file exists yet.
- Existing Responses fixtures now explicitly force Responses; native default
  dispatch must have separate production-entrypoint coverage.
- Preview-only stream failure returns result with zero completed images, but
  RecordUsage still accepts handler preflight CostOverride. No billing proof.
- Stream reader non-EOF error bypasses the output-written wrapper; preview-only
  EOF also bypasses it. These need explicit regression cases.
- Usage map is cloned, but cached-image metadata and charge/dedupe tests missing.
- Controller explicitly ran `go test ./internal/service -run
  'TestOpenAIGatewayServiceForwardImages_OAuth' -count=1`: FAIL, seven Responses
  fixtures (5.437s). The parent regex does not match this naming family, so this
  extra selector is mandatory for step 1 and final QA. Failures are stream
  transform/too-large/edit/multipart/output-item/multiline/disconnect cases.
- The forwarder's `imageCount <= 0` fallback still replaces zero with requested
  N; preview-only tests must inspect the final ForwardImages result, not merely
  the adapter's internal count.

## Bounded steps (all must finish before independent QA)

1. Transport tests and associated adapter fixes: add actual ForwardImages tests
   for native generation/edit, JSON/SSE, public/upstream mapping, MIME/format,
   legacy/unknown fallback, 404/405 call counts, errors/empty output, plan gates,
   completed-then-error, preview-then-error/read-error/EOF, and actual dimensions.
   Do not change accounting owners in this step. Report passing and failing
   assertions honestly; preview accounting can remain explicitly open to step 2.
2. Accounting implementation and tests: persist cached-image metadata without
   map mutation; prove exact cost, nil/populated maps, one-time charging/dedupe,
   and preview-only accounting despite a nonzero input preflight override.
   Keep native partial-result semantics honest; never invent completed images.
3. Full approved contract checks, worker report, controller diff review,
   independent Terra QA, then local integration decision. Real-provider and
   account-test parity remain open outside this local-mock contract.

No new business paths, handler edits, schema, provider calls, deployment, commits
by workers, or main-worktree edits are authorized by this work order.

## Additional upstream comparison checklist

Controller compared the current adapter with local Git object c0d511937:

- JSON MIME fallback must consider root output_format as well as per-item and
  request format (currently root format is skipped).
- Native usage with cached_tokens_details but absent cached_tokens must retain
  bounded explicit text/image cached totals (currently fallback is missing).
- isOpenAIImagesMainModelError currently has no call sites. A helper's presence
  does not prove driver-vs-image plan-gate behavior; test actual forwarding.
- Current public-model test uses identical request/upstream models and thus
  cannot prove mapped/public separation. Use distinct supported mapped models.

## Worker ownership recovery

The original Generator repeatedly ended bounded follow-ups before adding the
requested mapping tests. Its run is confirmed completed. Controller retains
all its edits and assigns the mapping JSON/SSE + multipart edit/MIME slice to
a fresh independent-context Terra Developer `native_mapping_tests`, under the
same parent approved contract. Only that developer may modify adapter/tests
during this slice. `native_contract_review` separately examines preview billing
read-only; it is not final QA. Billing, remaining transport cases and final QA
remain required after this slice, with no reduced completion claim.

Read-only preview-accounting review found no need to expand the approved
allowlist. Controller now assigns the disjoint gateway usage/RecordUsage-test
slice to fresh Terra `native_usage_accounting` while `native_mapping_tests`
owns adapter/tests. No shared business-file ownership. Both remain subordinate
to the full parent contract; independent final QA waits for both plus remaining
transport coverage. Preview count zero must ignore preflight image costs and
must not enter the image/per-request fallback's default RequestCount=1; token
mode may charge only measured token usage. Preserve all non-image behavior.

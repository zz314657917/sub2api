# DeepSeek compatibility publication checks

Date: 2026-10-22.

## Scope

Publish only the eight source/test paths previously integrated and sealed by the
independent DeepSeek QA. All eight match
`E:/codex-runtime/deepseek-compat-20261022/qa-seal.json`.

The change repairs CN Chat account eligibility, respects explicit DeepSeek
Responses protocol configuration, normalizes image URLs and mapped `[1m]`
suffixes, and supplies the DeepSeek model fallback catalog. CN image generation
remains rejected at both the handler and scheduler.

Other working-tree changes are excluded. In particular, the local beta test
fixture file also contains an unpublished structured-output test; it is not
published independently of its separate production changes.

## Fresh verification of the exact publication snapshot

The isolated snapshot contains committed HEAD `cb1c1516a`, the eight sealed
DeepSeek paths, and the three account-filter paths published separately.
It contains none of the other working-tree source changes.

- Fifteen new DeepSeek tests: default and unit tags, three repetitions each,
  45 top-level passes per tag, zero failures/skips.
- DeepSeek/CN/scheduler/compaction/V027/V0217 service regressions:
  default 169 and unit 178 top-level passes, zero failures/skips.
- Protocol packages apicompat, openai_compat and antigravity:
  327 top-level passes, zero failures/skips.
- Model catalog, CN and compaction handlers: 31 top-level passes.
- Codex account import: 40 top-level passes.
- Complete backend production build and server build with the newly built
  frontend embedded: passed.
- Eight Go paths: gofmt clean. Selected source diff whitespace check: passed.
- Independent static review found no blocking issue or dependency on the other
  unpublished working-tree paths. Review: `independent-review.md`.

Evidence:
`E:/codex-runtime/sub2api-publish-20261022/verification.json`,
`publish-paths.json`, and the named command logs.

## Limits

This is scoped regression and build acceptance, not a claim that the full
repository test suite passes. Fresh broad Codex checks reproduce exactly the
previous independent QA's failures: 226 top-level passes and two failures with
default tags; 244 passes and three failures with unit tags. No new failing test
names or skipped tests were observed. The failing tests remain
`TestApplyCodexOAuthTransform_PreservesAllowedTools`,
`TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested`,
and, for unit tags, `TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy`.
Comparison results are retained in `known-failures.json`.

No deployment, real provider request, production database operation, or runtime
service acceptance was performed. The publication does not change pricing,
peak-hour billing, or provider architecture.

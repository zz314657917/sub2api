### DONE: upstream-main-openai-runtime-s6b

## Verdict
DONE

## Candidate Results
- `679c0865a` (`fix(openai): handle versioned compatible base URLs`): APPLIED_EQUIVALENT. `cherry-pick -x` conflicted in `backend/internal/service/openai_images.go`, but after resolving to the current branch implementation the patch was empty. Versioned compatible base URL support is already present through `buildOpenAIEndpointURL` and existing tests.
- `a61174291` (`fix(gateway): detach upstream context unconditionally for image generation`): APPLIED_EQUIVALENT. `cherry-pick -x` conflicted in `backend/internal/service/openai_images_responses.go`; current branch already uses `detachUpstreamContext(ctx)` for image generation paths.
- `2c14efeaa` (`fix(openai-images): 修复图片生成 n 参数透传`): APPLIED_EQUIVALENT. `cherry-pick -x` conflicted in `backend/internal/service/openai_images_responses.go` and `backend/internal/service/openai_images_test.go`; current branch already includes the `shouldPassOpenAIImagesN` behavior plus newer one-image-per-request protection for `gpt-image-2`.
- `87fac3045` (`fix: use tier cooldown for google one gemini 429`): APPLIED_EQUIVALENT. `cherry-pick -x` became empty; current branch already contains the Gemini cooldown behavior.
- `bec1e2b69` (`fix(openai): 永久禁用缺失 refresh_token 的 OAuth 账号`): APPLIED_EQUIVALENT. `cherry-pick -x` produced empty-equivalent behavior after conflict resolution; committed an empty source record as `fdadfcc6c`.

## Deferred
- None.

## Changed Files
- `docs/workflow/worker-results/upstream-main-openai-runtime-s6b-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-runtime-s6b-qa.md`

## Acceptance Commands
- `git status --short --branch`: PASS.
- `git diff --check`: PASS.
- `git diff --name-status f78566b5d..HEAD`: PASS; only workflow docs before report generation, then worker result / QA report are added.
- `git diff --name-only f78566b5d..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|backend/internal/pkg/apicompat/|backend/internal/repository/|backend/internal/server/|deploy/|knowledge/|docs/workflow/main-log\\.md|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"`: PASS; no denied path matches.
- `go test ./internal/service -run "OpenAI|Images|Gateway|Versioned|BaseURL|Gemini|RefreshToken|RateLimit" -count=1` from `backend/`: PASS (`ok github.com/Wei-Shaw/sub2api/internal/service 48.438s`).

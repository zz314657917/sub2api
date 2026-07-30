### PASS: openai-local-group-id-s126

# QA Report

## Task ID

openai-local-group-id-s126

## Verdict

`PASS / source-only`

## Contract Checked

- `docs/workflow/tasks/openai-local-group-id-s126.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go test ./internal/service -run "TestStripOpenAILocalGroupID_TopLevelOnly|TestOpenAIGatewayService_Forward_TextResponsesSetsBillingModelToMappedModel|TestForwardAsRawChatCompletions_ForcesStreamUsageUpstreamAndPassesUsageDownstream|TestOpenAIGatewayServiceParseOpenAIImagesRequest_TransparentBackgroundAlias(JSON|Multipart)" -count=1 -> PASS (5.562s)
go test ./internal/service -run "OpenAI|Images|ChatCompletions" -count=1 -> PASS (36.410s)
go test ./... -run "^$" -> PASS (all backend packages compile)
gofmt -d <six changed Go files> -> PASS (empty output)
git diff --check -> PASS (only existing line-ending warnings)
rg conflict-marker audit on S126 files -> PASS (no matches)
```

- manual checks:

```text
Responses API-key upstream recorder -> top-level group_id absent
Raw third-party Chat Completions upstream recorder -> top-level group_id absent
Images JSON compatibility rewrite -> top-level group_id absent
Images multipart compatibility rewrite -> group_id form field absent; image part preserved by existing parser regression
Nested metadata.group_id -> preserved
```

## Findings

- 未发现明确问题。
- The API key and its resolved group remain the only group-routing authority;
  request-body `group_id` is consumed only as invalid local metadata and is not
  interpreted as a routing override.
- A real external strict OpenAI-compatible upstream, deployment, container
  refresh, commit, and push were not executed.

## Bug Owner Recommendation

`original-worker`

## Root Cause

`implementation-bug`

The raw forwarding paths preserved unknown top-level client fields, allowing
sub2api-local `group_id` metadata to reach strict upstream request validators.

## Retest Scope

- None required for source-level acceptance. If deployed later, send one
  request with top-level `group_id` to each of `/v1/responses`,
  `/v1/chat/completions`, `/v1/images/generations`, and multipart
  `/v1/images/edits` against a strict upstream and confirm it no longer returns
  `Unknown parameter: 'group_id'`.

## Knowledge Promotion

- `none`

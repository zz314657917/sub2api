### DONE: image2-account-parity

Changed paths:

- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_openai_image_test.go`
- `backend/internal/service/account_test_service_openai_native_test.go`
- `backend/internal/service/openai_images_direct.go`
- `backend/internal/service/openai_images_native_test.go`

Implemented mapped native account-test routing with JSON Accept, one 404/405
Responses fallback with explicit body closing, and a shared native JSON parser
used by both the direct non-stream forwarder and account test entrypoint. The
native parser retains item/root/request output-format precedence and rejects
malformed JSON. Existing Responses fixtures now intentionally use the legacy
`gpt-image-2.5` route; API-key assertions are untouched.

Production-entrypoint fake coverage includes mapped payload/prompt, MIME
priority, ordered SSE output, 404/405 fallback, no second fallback, HTTP and
transport failures, closed bodies, OAuth/SetupToken/shadow/Agent Identity auth,
custom/default user-agent, ChatGPT account header and proxy. Native parser
regressions cover format precedence and malformed JSON.

Executed evidence:

- `go test ./internal/service -run 'TestAccountTestService_OpenAIImage|TestCodexDirectImages|TestOpenAIGatewayServiceForwardImages' -count=1` — PASS.
- `go test ./internal/service -run 'TestAccountTestService_OpenAIImageNative|TestCodexDirectImagesJSONParser' -count=1` — PASS after the final credential matrix additions.
- `go build ./...` — PASS.
- Contract scope, `gofmt`, `git diff --check dc851ea3f`, and unmerged-index gates — PASS.

`go test ./internal/service -count=1` was started twice but the local command
window returned without a final exit result after 30 seconds; it is not claimed
as passed. No provider, database, container, deployment, commit, or push ran.

## QA remediation

Independent QA found that the first implementation moved final Codex identity
normalization before the custom/default User-Agent assignment. This is fixed:
the established final `enforceCodexIdentityHeaders` call again follows User-Agent
selection. Native tests now assert normalized `codex_cli_rs/` output, not raw
custom UA passthrough. Failure-path tests now assert close count for second
fallback, other HTTP, empty and malformed responses, plus Agent Identity
sensitive-error redaction. The directed identity/native and contracted selector
both passed after remediation; the independent `ClaudeCodeValidator` full-suite
baseline failure remains outside this task.

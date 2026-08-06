### PASS: upstream-v0171-image-data-url-s189

## Findings

- Upstream `d6467f6eb` fixes an image-task offload incompatibility: a valid `data:image/...;base64,...` result was previously treated as a remote URL. The local uploader now recognizes the scheme before its HTTP downloader and decodes it with the existing byte-size and image-content validation boundary.
- The parser requires an `image/*` media type and final `base64` marker, rejects malformed/non-base64 inputs without an HTTP call, preserves `b64_json` precedence, and uses detected bytes rather than a conflicting declared type for storage extension/content type.

## Executed Checks

- `gofmt -w backend/internal/service/image_storage.go backend/internal/service/image_storage_data_url_test.go`: passed.
- `go test ./internal/service -run '^TestImageResultUploader' -count=1` from `backend/`: passed. Covers local decode/no-HTTP behavior, malformed data URLs, decoded-size limits, and `b64_json` precedence.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0` from `backend/`: passed.
- `git diff --check`: passed; conflict-marker scan and `git ls-files -u` were empty.

## Unverified Risks

- No object-storage provider, external URL, image generation provider, database, or production task was contacted. The conclusion is limited to the local uploader's result-rewrite path.

## Recommendation

Commit to the isolated branch `codex/upstream-v0171-integration-s183`; do not merge the primary worktree, push, or deploy.

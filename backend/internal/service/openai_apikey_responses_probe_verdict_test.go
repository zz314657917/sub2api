package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/stretchr/testify/require"
)

type responsesProbeVerdictAccountRepo struct {
	stubOpenAIAccountRepo
	updateExtraCalls []map[string]any
}

func (r *responsesProbeVerdictAccountRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	copied := make(map[string]any, len(updates))
	for key, value := range updates {
		copied[key] = value
	}
	r.updateExtraCalls = append(r.updateExtraCalls, copied)
	return nil
}

func TestProbeOpenAIAPIKeyResponsesSupport_InconclusiveResponseKeepsUnknown(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"failed":                    `{"status":"failed","output":[{"type":"function_call"}]}`,
		"max_output_tokens":         `{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":[{"type":"function_call"}]}`,
		"max_output_tokens_no_call": `{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &responsesProbeVerdictAccountRepo{stubOpenAIAccountRepo: stubOpenAIAccountRepo{accounts: []Account{{
				ID:       213,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key": "sk-test",
				},
			}}}}
			svc := &AccountTestService{
				accountRepo: repo,
				cfg:         &config.Config{},
				httpUpstream: &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
				}},
			}

			svc.ProbeOpenAIAPIKeyResponsesSupport(context.Background(), 213)

			require.Empty(t, repo.updateExtraCalls)
		})
	}
}

func TestProbeOpenAIAPIKeyResponsesSupport_ConclusiveResponsesStillPersist(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		status    int
		body      string
		supported bool
	}{
		"completed_with_function_call": {
			status:    http.StatusOK,
			body:      `{"status":"completed","output":[{"type":"function_call"}]}`,
			supported: true,
		},
		"completed_without_function_call": {
			status:    http.StatusOK,
			body:      `{"status":"completed","output":[]}`,
			supported: false,
		},
		"other_incomplete": {
			status:    http.StatusOK,
			body:      `{"status":"incomplete","incomplete_details":{"reason":"content_filter"},"output":[]}`,
			supported: false,
		},
		"missing_status": {
			status:    http.StatusOK,
			body:      `{"output":[]}`,
			supported: false,
		},
		"malformed_status": {
			status:    http.StatusOK,
			body:      `{`,
			supported: false,
		},
		"not_found": {
			status:    http.StatusNotFound,
			body:      `{}`,
			supported: false,
		},
		"method_not_allowed": {
			status:    http.StatusMethodNotAllowed,
			body:      `{}`,
			supported: false,
		},
		"other_non_2xx": {
			status:    http.StatusBadRequest,
			body:      `{}`,
			supported: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &responsesProbeVerdictAccountRepo{stubOpenAIAccountRepo: stubOpenAIAccountRepo{accounts: []Account{{
				ID:       213,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key": "sk-test",
				},
			}}}}
			svc := &AccountTestService{
				accountRepo: repo,
				cfg:         &config.Config{},
				httpUpstream: &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: tt.status,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				}},
			}

			svc.ProbeOpenAIAPIKeyResponsesSupport(context.Background(), 213)

			require.Len(t, repo.updateExtraCalls, 1)
			require.Equal(t, tt.supported, repo.updateExtraCalls[0][openai_compat.ExtraKeyResponsesSupported])
		})
	}
}

func TestResponsesProbeVerdictIsConclusive(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		status     int
		body       string
		conclusive bool
	}{
		"failed": {
			status:     http.StatusOK,
			body:       `{"status":"failed"}`,
			conclusive: false,
		},
		"max_output_tokens": {
			status:     http.StatusCreated,
			body:       `{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"}}`,
			conclusive: false,
		},
		"other_incomplete": {
			status:     http.StatusOK,
			body:       `{"status":"incomplete","incomplete_details":{"reason":"content_filter"}}`,
			conclusive: true,
		},
		"missing_status": {
			status:     http.StatusOK,
			body:       `{}`,
			conclusive: true,
		},
		"malformed_body": {
			status:     http.StatusOK,
			body:       `{`,
			conclusive: true,
		},
		"non_2xx": {
			status:     http.StatusInternalServerError,
			body:       `{"status":"failed"}`,
			conclusive: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.conclusive, responsesProbeVerdictIsConclusive(tt.status, []byte(tt.body)))
		})
	}
}

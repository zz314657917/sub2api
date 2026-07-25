package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIGatewayService_ResponsesInstructionsDependOnAccountType(t *testing.T) {
	tests := []struct {
		name                  string
		accountType           string
		model                 string
		body                  string
		userAgent             string
		wantInstructions      string
		wantInstructionsExist bool
		wantDefaultCodex      bool
	}{
		{
			name:                  "api key without instructions stays absent despite codex user agent",
			accountType:           AccountTypeAPIKey,
			model:                 "gpt-5.4",
			body:                  `{"model":"gpt-5.4","input":"hello","stream":false}`,
			userAgent:             "codex_cli_rs/0.1.0",
			wantInstructionsExist: false,
		},
		{
			name:                  "api key explicit empty instructions stays empty",
			accountType:           AccountTypeAPIKey,
			model:                 "gpt-5.4",
			body:                  `{"model":"gpt-5.4","instructions":"","input":"hello","stream":false}`,
			userAgent:             "codex_cli_rs/0.1.0",
			wantInstructions:      "",
			wantInstructionsExist: true,
		},
		{
			name:                  "api key custom instructions stay unchanged",
			accountType:           AccountTypeAPIKey,
			model:                 "gpt-5.4",
			body:                  `{"model":"gpt-5.4","instructions":"Keep this exactly.","input":"hello","stream":false}`,
			userAgent:             "codex_cli_rs/0.1.0",
			wantInstructions:      "Keep this exactly.",
			wantInstructionsExist: true,
		},
		{
			name:                  "oauth without instructions gets codex default",
			accountType:           AccountTypeOAuth,
			model:                 "gpt-5.5",
			body:                  `{"model":"gpt-5.5","input":"hello","stream":false}`,
			userAgent:             "curl/8.0",
			wantInstructionsExist: true,
			wantDefaultCodex:      true,
		},
		{
			name:                  "oauth custom instructions stay unchanged",
			accountType:           AccountTypeOAuth,
			model:                 "gpt-5.5",
			body:                  `{"model":"gpt-5.5","instructions":"Use the caller instructions.","input":"hello","stream":false}`,
			userAgent:             "curl/8.0",
			wantInstructions:      "Use the caller instructions.",
			wantInstructionsExist: true,
		},
		{
			name:                  "api key gpt 5.6 luna does not fall back to gpt 5.5 codex prompt",
			accountType:           AccountTypeAPIKey,
			model:                 "gpt-5.6-luna",
			body:                  `{"model":"gpt-5.6-luna","input":"hello","stream":false}`,
			userAgent:             "codex_cli_rs/0.1.0",
			wantInstructionsExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader([]byte(tt.body)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("User-Agent", tt.userAgent)

			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(
					`{"id":"resp_instructions_test","model":"` + tt.model + `","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`,
				)),
			}}
			svc := &OpenAIGatewayService{
				cfg:          &config.Config{},
				httpUpstream: upstream,
			}
			account := &Account{
				ID:          123,
				Name:        "instructions-test",
				Platform:    PlatformOpenAI,
				Type:        tt.accountType,
				Concurrency: 1,
				Credentials: map[string]any{
					"api_key":            "sk-test",
					"access_token":       "oauth-token",
					"chatgpt_account_id": "chatgpt-acc",
					"base_url":           "https://example.com",
				},
				Extra: map[string]any{
					"use_responses_api": true,
				},
				Status:      StatusActive,
				Schedulable: true,
			}

			result, err := svc.Forward(context.Background(), c, account, []byte(tt.body))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, upstream.lastBody)

			instructions := gjson.GetBytes(upstream.lastBody, "instructions")
			require.Equal(t, tt.wantInstructionsExist, instructions.Exists())
			if tt.wantInstructionsExist && !tt.wantDefaultCodex {
				require.Equal(t, tt.wantInstructions, instructions.String())
			}
			if tt.wantDefaultCodex {
				require.Contains(t, instructions.String(), "You are Codex, a coding agent based on GPT-5")
			}
			if tt.accountType == AccountTypeAPIKey {
				require.NotContains(t, string(upstream.lastBody), "You are Codex")
			}
		})
	}
}

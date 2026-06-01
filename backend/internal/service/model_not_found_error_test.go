//go:build unit

package service

import (
	"net/http"
	"testing"
)

func TestIsUpstreamModelNotFoundError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       bool
	}{
		{
			name:       "openai model_not_found payload",
			statusCode: http.StatusNotFound,
			body:       `{"error":{"message":"The model 'gpt-5.4' does not exist","type":"invalid_request_error","code":"model_not_found"}}`,
			want:       true,
		},
		{
			name:       "unknown model wording",
			statusCode: http.StatusNotFound,
			body:       `{"error":{"message":"unknown model: claude-sonnet-4.5"}}`,
			want:       true,
		},
		{
			name:       "endpoint not found is not model scoped",
			statusCode: http.StatusNotFound,
			body:       `{"error":{"message":"route not found"}}`,
			want:       false,
		},
		{
			name:       "non 404 ignored",
			statusCode: http.StatusBadRequest,
			body:       `{"error":{"message":"model not found"}}`,
			want:       false,
		},
		{
			name:       "empty body ignored",
			statusCode: http.StatusNotFound,
			body:       "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUpstreamModelNotFoundError(tt.statusCode, []byte(tt.body)); got != tt.want {
				t.Fatalf("isUpstreamModelNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

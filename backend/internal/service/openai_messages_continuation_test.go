package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAICompatPreviousResponseUnsupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		message    string
		body       []byte
		want       bool
	}{
		{
			name:       "user unavailable message",
			statusCode: http.StatusBadRequest,
			message:    "previous_response_id is not available for this user",
			want:       true,
		},
		{
			name:       "api key required message",
			statusCode: http.StatusBadRequest,
			message:    "previous_response_id requires an openai api-key account for http requests",
			want:       true,
		},
		{
			name:       "nested user unavailable message",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"error":{"message":"The previous_response_id is not available for this user"}}`),
			want:       true,
		},
		{
			name:       "existing unsupported wording",
			statusCode: http.StatusBadRequest,
			message:    "previous_response_id is not supported by this endpoint",
			want:       true,
		},
		{
			name:       "missing previous response id",
			statusCode: http.StatusBadRequest,
			message:    "resource is not available for this user",
			want:       false,
		},
		{
			name:       "wrong status code",
			statusCode: http.StatusInternalServerError,
			message:    "previous_response_id is not available for this user",
			want:       false,
		},
		{
			name:       "not found remains separate classification",
			statusCode: http.StatusNotFound,
			message:    "previous_response_id is not available for this user",
			want:       false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, isOpenAICompatPreviousResponseUnsupported(tt.statusCode, tt.message, tt.body))
		})
	}
}

package handler

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestShouldSubmitOpenAIPartialUsage(t *testing.T) {
	result := &service.OpenAIForwardResult{Usage: service.OpenAIUsage{InputTokens: 1}}
	failover := &service.UpstreamFailoverError{}

	tests := []struct {
		name   string
		err    error
		result *service.OpenAIForwardResult
		want   bool
	}{
		{name: "nil error", err: nil, result: result, want: false},
		{name: "nil result", err: errors.New("upstream failed"), result: nil, want: false},
		{name: "generic error with partial result", err: errors.New("stream interrupted"), result: result, want: true},
		{name: "failover error", err: failover, result: result, want: false},
		{name: "wrapped failover error", err: fmt.Errorf("wrapped: %w", failover), result: result, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSubmitOpenAIPartialUsage(tt.err, tt.result); got != tt.want {
				t.Fatalf("shouldSubmitOpenAIPartialUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

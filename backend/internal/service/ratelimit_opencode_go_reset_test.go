package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseOpenAIRateLimitResetTimeOpenCodeGoUsageLimit(t *testing.T) {
	tests := []struct {
		name string
		body string
		want time.Duration
	}{
		{
			name: "days",
			body: `{"type":"error","error":{"type":"GoUsageLimitError","message":"Weekly usage limit reached. Resets in 2 days."}}`,
			want: 48 * time.Hour,
		},
		{
			name: "compound hours and minutes",
			body: `{"type":"error","error":{"type":"GoUsageLimitError","message":"5-hour usage limit reached. Resets in 4hr 59min."}}`,
			want: 4*time.Hour + 59*time.Minute,
		},
		{
			name: "unknown error type stays ignored",
			body: `{"type":"error","error":{"type":"rate_limit_error","message":"Resets in 2 days."}}`,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			resetAt := parseOpenAIRateLimitResetTime([]byte(tt.body))
			after := time.Now()

			if tt.want == 0 {
				require.Nil(t, resetAt)
				return
			}
			require.NotNil(t, resetAt)
			actual := time.Unix(*resetAt, 0)
			require.False(t, actual.Before(before.Add(tt.want).Truncate(time.Second)))
			require.False(t, actual.After(after.Add(tt.want)))
		})
	}
}

func TestParseOpenCodeGoUsageLimitResetDuration(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    time.Duration
	}{
		{name: "case insensitive marker", message: "limit reached; resets in 1.5 HOURS", want: time.Hour + 30*time.Minute},
		{name: "weeks days and seconds", message: "Resets in 1 week 2 days 3sec", want: 9*24*time.Hour + 3*time.Second},
		{name: "missing marker", message: "usage limit reached in 2 days", want: 0},
		{name: "zero is rejected", message: "Resets in 0 minutes", want: 0},
		{name: "negative is rejected", message: "Resets in -2 hours", want: 0},
		{name: "unknown unit is rejected", message: "Resets in 2 parsecs", want: 0},
		{name: "overflow is rejected", message: "Resets in 999999999999999999999 hours", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, parseOpenCodeGoUsageLimitResetDuration(tt.message))
		})
	}
}

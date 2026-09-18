package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamResponseModelObservationTerminalAndProtocols(t *testing.T) {
	o := &upstreamResponseModelObserver{}
	o.ObserveOpenAI([]byte(`{"type":"response.created","response":{"model":"gpt-first"}}`), "response.created")
	o.ObserveOpenAI([]byte(`{"type":"response.completed","response":{"model":"gpt-final"}}`), "response.completed")
	require.Equal(t, "gpt-final", o.Model())
	require.True(t, o.Conflict())

	g := &upstreamResponseModelObserver{}
	g.ObserveGemini([]byte(`{"response":{"response":{"modelVersion":"gemini-a"}}}`))
	g.ObserveGemini([]byte(`{"modelVersion":"gemini-b"}`))
	require.Equal(t, "gemini-b", g.Model())

	a := &upstreamResponseModelObserver{}
	a.ObserveAnthropic([]byte(`{"message":{"model":"claude-first"}}`))
	a.ObserveAnthropic([]byte(`{"message":{"model":"claude-second"}}`))
	require.Equal(t, "claude-first", a.Model())
}

func TestUpstreamResponseModelRejectsMalformedAndCapsRunes(t *testing.T) {
	o := &upstreamResponseModelObserver{}
	o.ObserveOpenAI([]byte(`{"model":"ok"`), "response.completed")
	require.Empty(t, o.Model())
	o.Observe(strings.Repeat("测", 201), false)
	require.Len(t, []rune(o.Model()), 200)
}

func TestUpstreamModelMismatchGrokAliasesAndUnknown(t *testing.T) {
	for _, tc := range []struct {
		sent, observed string
		want           *bool
	}{
		{"grok-4.5-latest", "GROK-4.5-BUILD", upstreamResponseModelBoolPtr(false)},
		{"grok-4.5", "grok-4.6-build", upstreamResponseModelBoolPtr(true)},
		{"gpt-5.5", "", nil},
	} {
		got := upstreamModelMismatch(tc.sent, tc.observed)
		if tc.want == nil {
			require.Nil(t, got)
		} else {
			require.Equal(t, *tc.want, *got)
		}
	}
}

func upstreamResponseModelBoolPtr(v bool) *bool { return &v }

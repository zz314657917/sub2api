package service

import (
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func filterGrokPingTestInput(t *testing.T, input string) string {
	t.Helper()
	body := newGrokResponsesBillingPingFilterBody(
		io.NopCloser(strings.NewReader(input)),
		&Account{Platform: PlatformGrok},
		defaultMaxLineSize,
	)
	output, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NoError(t, body.Close())
	return string(output)
}

func TestGrokResponsesBillingPingFilterRewritesOnlyCompatiblePingFrames(t *testing.T) {
	input := strings.Join([]string{
		"event: ping",
		`data: {"type":"ping","cost":"0.25"}`,
		"",
		"event: ping",
		`data: {not-json}`,
		"",
		"event: ping",
		"",
		"event: ping",
		`data: {"type":"response.completed"}`,
		"",
		"event: response.output_text.delta",
		`data: {"type":"response.output_text.delta","delta":"ok"}`,
		"",
	}, "\n")
	expected := strings.Join([]string{
		": ping",
		"",
		": ping",
		"",
		": ping",
		"",
		"event: ping",
		`data: {"type":"response.completed"}`,
		"",
		"event: response.output_text.delta",
		`data: {"type":"response.output_text.delta","delta":"ok"}`,
		"",
	}, "\n")

	require.Equal(t, expected, filterGrokPingTestInput(t, input))
}

func TestGrokResponsesBillingPingFilterPassesOverLimitCandidatesThrough(t *testing.T) {
	lineBounded := []string{"event: ping"}
	for range grokResponsesPingFrameMaxLines {
		lineBounded = append(lineBounded, ": vendor keepalive")
	}
	lineBounded = append(lineBounded, "")
	lineInput := strings.Join(lineBounded, "\n")
	require.Equal(t, lineInput, filterGrokPingTestInput(t, lineInput))

	byteInput := "event: ping\ndata: " + strings.Repeat("x", grokResponsesPingFrameMaxBytes) + "\n\n"
	require.Equal(t, byteInput, filterGrokPingTestInput(t, byteInput))
}

func TestGrokResponsesBillingPingFilterLeavesNonGrokBodyUnwrapped(t *testing.T) {
	input := "event: ping\ndata: {\"type\":\"ping\"}\n\n"
	source := io.NopCloser(strings.NewReader(input))
	body := newGrokResponsesBillingPingFilterBody(source, &Account{Platform: PlatformOpenAI}, defaultMaxLineSize)

	output, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Equal(t, input, string(output))
	require.NoError(t, body.Close())
}

type grokPingFilterTestReadCloser struct {
	reader     io.ReadCloser
	closeErr   error
	closeCount atomic.Int32
}

func (r *grokPingFilterTestReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *grokPingFilterTestReadCloser) Close() error {
	r.closeCount.Add(1)
	_ = r.reader.Close()
	return r.closeErr
}

func TestGrokResponsesBillingPingFilterCloseCancelsSourceOnceAndPropagatesError(t *testing.T) {
	upstreamReader, upstreamWriter := io.Pipe()
	closeErr := errors.New("source close failed")
	source := &grokPingFilterTestReadCloser{reader: upstreamReader, closeErr: closeErr}
	body := newGrokResponsesBillingPingFilterBody(source, &Account{Platform: PlatformGrok}, defaultMaxLineSize)

	require.ErrorIs(t, body.Close(), closeErr)
	require.Equal(t, int32(1), source.closeCount.Load())
	require.Error(t, body.Close())
	require.Equal(t, int32(1), source.closeCount.Load())
	_, err := upstreamWriter.Write([]byte("blocked"))
	require.Error(t, err)
	require.NoError(t, upstreamWriter.Close())
}

func TestGrokResponsesBillingPingFilterFlushesCompleteNonPingFrameBeforeEOF(t *testing.T) {
	upstreamReader, upstreamWriter := io.Pipe()
	body := newGrokResponsesBillingPingFilterBody(upstreamReader, &Account{Platform: PlatformGrok}, defaultMaxLineSize)
	t.Cleanup(func() { require.NoError(t, body.Close()) })

	writeResult := make(chan error, 1)
	go func() {
		_, err := io.WriteString(upstreamWriter, "event: future.event\ndata: {\"type\":\"future.event\"}\n\n")
		writeResult <- err
	}()

	type readResult struct {
		value string
		err   error
	}
	readDone := make(chan readResult, 1)
	go func() {
		buffer := make([]byte, 128)
		n, err := body.Read(buffer)
		readDone <- readResult{value: string(buffer[:n]), err: err}
	}()

	select {
	case result := <-readDone:
		require.NoError(t, result.err)
		require.Contains(t, result.value, "future.event")
	case <-time.After(time.Second):
		t.Fatal("completed frame was buffered until upstream EOF")
	}
	require.NoError(t, <-writeResult)
	require.NoError(t, upstreamWriter.Close())
}

func TestGrokResponsesBillingPingFilterHandlesPartialAndCRFrames(t *testing.T) {
	partialPing := "event: ping\ndata: {\"type\":\"ping\",\"cost\":\"0\"}"
	require.Equal(t, ": ping\n\n", filterGrokPingTestInput(t, partialPing))

	bareCR := "event: ping\rdata: {\"type\":\"ping\",\"cost\":\"0\"}\r\r" +
		"event: future.event\rdata: {\"type\":\"future.event\"}\r\r"
	require.Equal(
		t,
		": ping\n\n"+"event: future.event\rdata: {\"type\":\"future.event\"}\r\r",
		filterGrokPingTestInput(t, bareCR),
	)
}

func TestGrokResponsesBillingPingFilterReportsOversizedLine(t *testing.T) {
	body := newGrokResponsesBillingPingFilterBody(
		io.NopCloser(strings.NewReader("data: 123456789\n\n")),
		&Account{Platform: PlatformGrok},
		8,
	)
	_, err := io.ReadAll(body)
	require.ErrorContains(t, err, "filter Grok Responses billing ping")
	require.NoError(t, body.Close())
}

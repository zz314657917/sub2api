package service

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const compactKeepaliveTestInterval = 5 * time.Millisecond

func waitForCompactKeepalive() {
	time.Sleep(5 * compactKeepaliveTestInterval)
}

func TestRemoteCompactKeepaliveCommitsSSEComment(t *testing.T) {
	c, rec := newCompactBridgeTestContext(t, true)
	stop := StartOpenAICompactSSEKeepalive(c, compactKeepaliveTestInterval)
	waitForCompactKeepalive()
	committed := StopOpenAICompactSSEKeepaliveCommitted(c)
	stop()

	require.True(t, committed)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), ": keepalive\n\n")
}

func TestRemoteCompactKeepaliveBytesDoNotSuppressFailover(t *testing.T) {
	c, rec := newCompactBridgeTestContext(t, true)
	before := OpenAICompactKeepaliveAdjustedWrittenSize(c)
	stop := StartOpenAICompactSSEKeepalive(c, compactKeepaliveTestInterval)
	waitForCompactKeepalive()
	require.Equal(t, before, OpenAICompactKeepaliveAdjustedWrittenSize(c))

	_, err := c.Writer.Write([]byte("business-response"))
	require.NoError(t, err)
	stop()
	require.Equal(t, len("business-response"), OpenAICompactKeepaliveAdjustedWrittenSize(c))
	require.Contains(t, rec.Body.String(), "business-response")
}

func TestRemoteCompactKeepaliveCommittedFailureUsesFailedEvent(t *testing.T) {
	c, rec := newCompactBridgeTestContext(t, true)
	stop := StartOpenAICompactSSEKeepalive(c, compactKeepaliveTestInterval)
	defer stop()
	waitForCompactKeepalive()

	require.True(t, writeOpenAICompactSSEBridge(c, http.StatusBadGateway, []byte(`{"error":{"message":"upstream exploded"}}`)))
	events := parseCompactBridgeEvents(t, rec.Body.String())
	require.Len(t, events, 1)
	require.Equal(t, "response.failed", events[0][0])
	require.Equal(t, "failed", gjson.Get(events[0][1], "response.status").String())
	require.Contains(t, gjson.Get(events[0][1], "response.error.message").String(), "upstream exploded")
	streamErr, ok := GetOpsStreamError(c)
	require.True(t, ok)
	require.Equal(t, http.StatusBadGateway, streamErr.IntendedStatus)
}

func TestRemoteCompactKeepaliveSerializesFinalWrite(t *testing.T) {
	c, rec := newCompactBridgeTestContext(t, true)
	stop := StartOpenAICompactSSEKeepalive(c, compactKeepaliveTestInterval)
	waitForCompactKeepalive()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = c.Writer.Write([]byte("final"))
		c.Writer.Flush()
	}()
	<-done
	stop()
	lengthAfterFinal := rec.Body.Len()
	waitForCompactKeepalive()

	require.Equal(t, lengthAfterFinal, rec.Body.Len())
	require.True(t, strings.HasSuffix(rec.Body.String(), "final"))
}

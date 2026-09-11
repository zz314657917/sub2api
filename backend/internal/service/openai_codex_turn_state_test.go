package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newS219TurnStateContext(t *testing.T, apiKeyID int64, sessionID string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	if sessionID != "" {
		c.Request.Header.Set("session_id", sessionID)
	}
	if apiKeyID > 0 {
		c.Set("api_key", &APIKey{ID: apiKeyID})
	}
	return c, recorder
}

func turnStateResponse(state, contentType, body string) *http.Response {
	h := http.Header{"Content-Type": []string{contentType}}
	if state != "" {
		h.Set(openAICodexTurnStateHeader, state)
	}
	return &http.Response{StatusCode: http.StatusOK, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

func TestOpenAICodexTurnStateSeedRequiresAPIKeyAndSession(t *testing.T) {
	c, _ := newS219TurnStateContext(t, 17, "underscore-session")
	require.Equal(t, "17\x00underscore-session", openAICodexTurnStateSeed(c))
	c.Request.Header.Set("session-id", "hyphen-session")
	require.Equal(t, "17\x00hyphen-session", openAICodexTurnStateSeed(c))

	noKey, _ := newS219TurnStateContext(t, 0, "session")
	noSession, _ := newS219TurnStateContext(t, 17, "")
	require.Empty(t, openAICodexTurnStateSeed(noKey))
	require.Empty(t, openAICodexTurnStateSeed(noSession))
	require.Empty(t, openAICodexTurnStateSeed(nil))
}

func TestOpenAICodexTurnStateRelayGuardAndExpiry(t *testing.T) {
	svc := &OpenAIGatewayService{}
	c, recorder := newS219TurnStateContext(t, 17, "session-a")
	origin := &Account{ID: 101}
	upstream := http.Header{}
	upstream.Set(openAICodexTurnStateHeader, "state-a")
	state := stageOpenAICodexTurnState(c.Writer.Header(), upstream)
	c.Data(http.StatusOK, "application/json", []byte(`{}`))
	svc.noteOpenAICodexTurnStateCommitted(c, origin, state)
	require.Equal(t, "state-a", recorder.Header().Get(openAICodexTurnStateHeader))

	sameAccount := http.Header{}
	sameAccount.Set(openAICodexTurnStateHeader, "state-a")
	svc.guardOpenAICodexTurnStateEcho(c, origin, sameAccount)
	require.Equal(t, "state-a", sameAccount.Get(openAICodexTurnStateHeader))
	foreignAccount := http.Header{}
	foreignAccount.Set(openAICodexTurnStateHeader, "state-a")
	svc.guardOpenAICodexTurnStateEcho(c, &Account{ID: 102}, foreignAccount)
	require.Empty(t, foreignAccount.Get(openAICodexTurnStateHeader))

	seed := openAICodexTurnStateSeed(c)
	svc.openaiCodexTurnStateOrigins.Store(seed, openAICodexTurnStateOrigin{accountID: 101, expiresAt: time.Now().Add(-time.Second)})
	expired := http.Header{}
	expired.Set(openAICodexTurnStateHeader, "state-a")
	svc.guardOpenAICodexTurnStateEcho(c, &Account{ID: 102}, expired)
	require.Equal(t, "state-a", expired.Get(openAICodexTurnStateHeader))
	_, exists := svc.openaiCodexTurnStateOrigins.Load(seed)
	require.False(t, exists)

	unknown := http.Header{}
	unknown.Set(openAICodexTurnStateHeader, "unknown-state")
	svc.guardOpenAICodexTurnStateEcho(c, &Account{ID: 102}, unknown)
	require.Equal(t, "unknown-state", unknown.Get(openAICodexTurnStateHeader))

	svc.openaiCodexTurnStateOrigins.Store(seed, "malformed")
	malformed := http.Header{}
	malformed.Set(openAICodexTurnStateHeader, "malformed-state")
	svc.guardOpenAICodexTurnStateEcho(c, &Account{ID: 102}, malformed)
	require.Equal(t, "malformed-state", malformed.Get(openAICodexTurnStateHeader))
	_, exists = svc.openaiCodexTurnStateOrigins.Load(seed)
	require.False(t, exists)

	svc.openaiCodexTurnStateOrigins.Store("sweep-malformed", "malformed")
	svc.openaiCodexTurnStateOrigins.Store("sweep-expired", openAICodexTurnStateOrigin{expiresAt: time.Now().Add(-time.Second)})
	svc.openaiCodexTurnStateWrites.Store(255)
	svc.sweepOpenAICodexTurnStateOrigins()
	_, exists = svc.openaiCodexTurnStateOrigins.Load("sweep-malformed")
	require.False(t, exists)
	_, exists = svc.openaiCodexTurnStateOrigins.Load("sweep-expired")
	require.False(t, exists)

	noAccount, _ := newS219TurnStateContext(t, 18, "no-account")
	svc.noteOpenAICodexTurnStateProvenance(noAccount, nil)
	svc.noteOpenAICodexTurnStateProvenance(noAccount, &Account{})
	_, exists = svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(noAccount))
	require.False(t, exists)

	stageOpenAICodexTurnState(recorder.Header(), http.Header{})
	require.Empty(t, recorder.Header().Get(openAICodexTurnStateHeader))
}

func TestOpenAIHTTPBuildersGuardCrossAccountTurnState(t *testing.T) {
	svc := &OpenAIGatewayService{}
	c, _ := newS219TurnStateContext(t, 22, "builder-session")
	c.Request.Header.Set(openAICodexTurnStateHeader, "foreign-state")
	foreign := &Account{ID: 202, Type: AccountTypeAPIKey}
	svc.noteOpenAICodexTurnStateProvenance(c, foreign)
	local := &Account{ID: 203, Type: AccountTypeAPIKey}

	normal, err := svc.buildUpstreamRequest(context.Background(), c, local, []byte(`{"model":"gpt-5"}`), "token", false, "", false)
	require.NoError(t, err)
	require.Empty(t, normal.Header.Get(openAICodexTurnStateHeader))

	passthrough, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, local, []byte(`{"model":"gpt-5"}`), "token")
	require.NoError(t, err)
	require.Empty(t, passthrough.Header.Get(openAICodexTurnStateHeader))

	svc.noteOpenAICodexTurnStateProvenance(c, local)
	same, err := svc.buildUpstreamRequest(context.Background(), c, local, []byte(`{"model":"gpt-5"}`), "token", false, "", false)
	require.NoError(t, err)
	require.Equal(t, "foreign-state", same.Header.Get(openAICodexTurnStateHeader))
}

func TestOpenAIStreamingTurnStateRecordsOnlyOnCommit(t *testing.T) {
	svc := &OpenAIGatewayService{cfg: &config.Config{}, toolCorrector: NewCodexToolCorrector()}
	account := &Account{ID: 301, Type: AccountTypeAPIKey}
	c, _ := newS219TurnStateContext(t, 31, "stream-session")
	noteWritten := make([]bool, 0, 1)
	svc.openaiCodexTurnStateNoteHook = func(noteContext *gin.Context) {
		noteWritten = append(noteWritten, noteContext.Writer.Written())
	}

	// No downstream flush happens before this failover result, so this account
	// must not become provenance merely because its upstream response had a header.
	noOutput := turnStateResponse("abandoned", "text/event-stream", "")
	_, err := svc.handleStreamingResponse(context.Background(), noOutput, c, account, time.Now(), "gpt-5", "gpt-5")
	require.Error(t, err)
	_, exists := svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(c))
	require.False(t, exists)

	committed := turnStateResponse("committed", "text/event-stream", strings.Join([]string{
		`data: {"type":"response.output_text.delta","delta":"ok"}`,
		`data: {"type":"response.completed","response":{"id":"resp-1","usage":{"input_tokens":1,"output_tokens":1}}}`,
	}, "\n"))
	_, err = svc.handleStreamingResponse(context.Background(), committed, c, account, time.Now(), "gpt-5", "gpt-5")
	require.NoError(t, err)
	origin, exists := svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(c))
	require.True(t, exists)
	require.Equal(t, account.ID, origin.(openAICodexTurnStateOrigin).accountID)
	require.Equal(t, []bool{true}, noteWritten, "multiple flushes must record normal-stream provenance once, after commit")
}

func TestOpenAINonStreamingTurnStateRelaysJSONAndSSE(t *testing.T) {
	svc := &OpenAIGatewayService{cfg: &config.Config{}}
	account := &Account{ID: 401, Type: AccountTypeAPIKey}
	noteWritten := make([]bool, 0, 2)
	svc.openaiCodexTurnStateNoteHook = func(noteContext *gin.Context) {
		noteWritten = append(noteWritten, noteContext.Writer.Written())
	}
	cJSON, recorderJSON := newS219TurnStateContext(t, 41, "json-session")
	jsonResp := turnStateResponse("json-state", "application/json", `{"id":"resp-json","usage":{"input_tokens":1,"output_tokens":1}}`)
	_, err := svc.handleNonStreamingResponse(context.Background(), jsonResp, cJSON, account, "gpt-5", "gpt-5")
	require.NoError(t, err)
	require.Equal(t, "json-state", recorderJSON.Header().Get(openAICodexTurnStateHeader))

	cSSE, recorderSSE := newS219TurnStateContext(t, 42, "sse-session")
	sseResp := turnStateResponse("sse-state", "text/event-stream", `data: {"type":"response.completed","response":{"id":"resp-sse","usage":{"input_tokens":1,"output_tokens":1}}}`)
	_, err = svc.handleNonStreamingResponse(context.Background(), sseResp, cSSE, account, "gpt-5", "gpt-5")
	require.NoError(t, err)
	require.Equal(t, "sse-state", recorderSSE.Header().Get(openAICodexTurnStateHeader))
	require.Equal(t, []bool{true, true}, noteWritten, "normal JSON and SSE-to-JSON must note only after c.Data/bridge commits")
}

func TestOpenAIPassthroughTurnStateRelayAndGuard(t *testing.T) {
	svc := &OpenAIGatewayService{cfg: &config.Config{}}
	account := &Account{ID: 501, Type: AccountTypeAPIKey}
	noteWritten := make([]bool, 0, 3)
	svc.openaiCodexTurnStateNoteHook = func(noteContext *gin.Context) {
		noteWritten = append(noteWritten, noteContext.Writer.Written())
	}
	c, recorder := newS219TurnStateContext(t, 51, "passthrough-session")
	resp := turnStateResponse("passthrough-state", "application/json", `{"id":"resp-pass","usage":{"input_tokens":1,"output_tokens":1}}`)
	_, err := svc.handleNonStreamingResponsePassthrough(context.Background(), resp, c, account, "gpt-5", "gpt-5")
	require.NoError(t, err)
	require.Equal(t, "passthrough-state", recorder.Header().Get(openAICodexTurnStateHeader))

	cSSE, recorderSSE := newS219TurnStateContext(t, 52, "passthrough-sse-session")
	sseResp := turnStateResponse("passthrough-sse-state", "text/event-stream", `data: {"type":"response.completed","response":{"id":"resp-pass-sse","usage":{"input_tokens":1,"output_tokens":1}}}`)
	_, err = svc.handleNonStreamingResponsePassthrough(context.Background(), sseResp, cSSE, account, "gpt-5", "gpt-5")
	require.NoError(t, err)
	require.Equal(t, "passthrough-sse-state", recorderSSE.Header().Get(openAICodexTurnStateHeader))
	require.Equal(t, []bool{true, true}, noteWritten, "passthrough JSON and SSE-to-JSON must note only after commit")

	c.Request.Header.Set(openAICodexTurnStateHeader, "passthrough-state")
	req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, &Account{ID: 502, Type: AccountTypeAPIKey}, []byte(`{"model":"gpt-5"}`), "token")
	require.NoError(t, err)
	require.Empty(t, req.Header.Get(openAICodexTurnStateHeader))

	streamContext, streamRecorder := newS219TurnStateContext(t, 53, "passthrough-stream-session")
	streamAccount := &Account{ID: 503, Type: AccountTypeAPIKey}
	stream := turnStateResponse("passthrough-stream-state", "text/event-stream", strings.Join([]string{
		`data: {"type":"response.output_text.delta","delta":"ok"}`,
		`data: {"type":"response.completed","response":{"id":"resp-pass-stream","usage":{"input_tokens":1,"output_tokens":1}}}`,
	}, "\n"))
	_, err = svc.handleStreamingResponsePassthrough(context.Background(), stream, streamContext, streamAccount, time.Now(), "gpt-5", "gpt-5")
	require.NoError(t, err)
	require.Equal(t, "passthrough-stream-state", streamRecorder.Header().Get(openAICodexTurnStateHeader))
	streamOrigin, exists := svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(streamContext))
	require.True(t, exists)
	require.Equal(t, streamAccount.ID, streamOrigin.(openAICodexTurnStateOrigin).accountID)
	require.Equal(t, []bool{true, true, true}, noteWritten, "multiple passthrough stream flushes must note exactly once")
}

func TestWriteOpenAIPassthroughResponseHeadersTurnState(t *testing.T) {
	dst := http.Header{http.CanonicalHeaderKey(openAICodexTurnStateHeader): []string{"stale"}}
	src := http.Header{http.CanonicalHeaderKey(openAICodexTurnStateHeader): []string{"fresh"}}
	writeOpenAIPassthroughResponseHeaders(dst, src, nil)
	require.Equal(t, "fresh", dst.Get(openAICodexTurnStateHeader))
	writeOpenAIPassthroughResponseHeaders(dst, http.Header{"Content-Type": []string{"application/json"}}, nil)
	require.Empty(t, dst.Get(openAICodexTurnStateHeader))
	dst.Set(openAICodexTurnStateHeader, "stale-nil-source")
	writeOpenAIPassthroughResponseHeaders(dst, nil, nil)
	require.Empty(t, dst.Get(openAICodexTurnStateHeader))
}

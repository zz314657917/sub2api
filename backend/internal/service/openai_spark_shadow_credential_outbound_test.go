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
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func sparkShadowAccountID(id int64) *int64 { return &id }

func newSparkShadowOAuthAccount(id int64, accessToken, chatGPTAccountID string) Account {
	return Account{
		ID:          id,
		Name:        "spark-shadow-outbound-test",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":       accessToken,
			"chatgpt_account_id": chatGPTAccountID,
			"user_agent":         "codex_cli_rs/0.99.0",
		},
		RateMultiplier: f64p(1),
	}
}

func newSparkShadowHTTPResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"spark-shadow-test"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`data: {"type":"response.completed","response":{"id":"resp_spark_shadow","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`,
			"",
			"data: [DONE]",
			"",
		}, "\\n"))),
	}
}

func newSparkShadowWSConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	cfg.Gateway.OpenAIWS.Enabled = true
	cfg.Gateway.OpenAIWS.OAuthEnabled = true
	cfg.Gateway.OpenAIWS.APIKeyEnabled = true
	cfg.Gateway.OpenAIWS.ResponsesWebsocketsV2 = true
	cfg.Gateway.OpenAIWS.ModeRouterV2Enabled = true
	cfg.Gateway.OpenAIWS.IngressModeDefault = OpenAIWSIngressModeCtxPool
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 1
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 1
	cfg.Gateway.OpenAIWS.QueueLimitPerConn = 8
	cfg.Gateway.OpenAIWS.DialTimeoutSeconds = 3
	cfg.Gateway.OpenAIWS.ReadTimeoutSeconds = 3
	cfg.Gateway.OpenAIWS.WriteTimeoutSeconds = 3
	return cfg
}

func newSparkShadowGinContext() *gin.Context {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.99.0")
	return c
}

func TestOpenAIGatewayService_SparkShadowOutboundHTTPUsesParentAndDoesNotLeakToNextAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	parent := newSparkShadowOAuthAccount(6101, "test-parent-http-access", "test-parent-http-chatgpt")
	child := newSparkShadowOAuthAccount(6102, "", "")
	child.ParentAccountID = sparkShadowAccountID(parent.ID)
	child.Credentials = nil
	child.Extra = map[string]any{"openai_passthrough": true}
	next := newSparkShadowOAuthAccount(6103, "test-next-http-access", "test-next-http-chatgpt")
	next.Extra = map[string]any{"openai_passthrough": true}

	upstream := &httpUpstreamRecorder{responses: []*http.Response{newSparkShadowHTTPResponse(), newSparkShadowHTTPResponse()}}
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{},
		httpUpstream: upstream,
		accountRepo:  stubOpenAIAccountRepo{accounts: []Account{parent, next}},
	}
	body := []byte(`{"model":"gpt-5.1","stream":false,"input":[{"type":"input_text","text":"hello"}]}`)

	_, err := svc.Forward(context.Background(), newSparkShadowGinContext(), &child, body)
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.True(t, upstream.requests[0].Header.Get("Authorization") == "Bearer test-parent-http-access")
	require.Equal(t, "test-parent-http-chatgpt", upstream.requests[0].Header.Get("chatgpt-account-id"))

	_, err = svc.Forward(context.Background(), newSparkShadowGinContext(), &next, body)
	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	require.True(t, upstream.requests[1].Header.Get("Authorization") == "Bearer test-next-http-access")
	require.Equal(t, "test-next-http-chatgpt", upstream.requests[1].Header.Get("chatgpt-account-id"))

	missingParent := child
	missingParent.ID = 6104
	missingParent.ParentAccountID = sparkShadowAccountID(6199)
	_, err = svc.Forward(context.Background(), newSparkShadowGinContext(), &missingParent, body)
	require.Error(t, err)
	require.Len(t, upstream.requests, 2, "missing parent must not create an outbound HTTP request")
}

func TestOpenAIGatewayService_SparkShadowOutboundWSUsesParentAndFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	parent := newSparkShadowOAuthAccount(6201, "test-parent-ws-access", "test-parent-ws-chatgpt")
	child := newSparkShadowOAuthAccount(6202, "", "")
	child.ParentAccountID = sparkShadowAccountID(parent.ID)
	child.Credentials = nil
	child.Extra = map[string]any{"responses_websockets_v2_enabled": true}
	next := newSparkShadowOAuthAccount(6203, "test-next-ws-access", "test-next-ws-chatgpt")
	next.Extra = map[string]any{"responses_websockets_v2_enabled": true}

	cfg := newSparkShadowWSConfig()
	conn := &openAIWSCaptureConn{events: [][]byte{
		[]byte(`{"type":"response.completed","response":{"id":"resp_spark_shadow_1","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`),
		[]byte(`{"type":"response.completed","response":{"id":"resp_spark_shadow_2","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`),
	}}
	dialer := &openAIWSCaptureDialer{conn: conn}
	pool := newOpenAIWSConnPool(cfg)
	pool.setClientDialerForTest(dialer)
	svc := &OpenAIGatewayService{
		cfg:              cfg,
		httpUpstream:     &httpUpstreamRecorder{},
		cache:            &stubGatewayCache{},
		accountRepo:      stubOpenAIAccountRepo{accounts: []Account{parent, next}},
		openaiWSResolver: NewOpenAIWSProtocolResolver(cfg),
		toolCorrector:    NewCodexToolCorrector(),
		openaiWSPool:     pool,
	}
	body := []byte(`{"model":"gpt-5.1","stream":false,"input":[{"type":"input_text","text":"hello"}]}`)

	result, err := svc.Forward(context.Background(), newSparkShadowGinContext(), &child, body)
	require.NoError(t, err)
	require.Equal(t, "resp_spark_shadow_1", result.RequestID)
	firstHeaders := cloneHeader(dialer.lastHeaders)
	require.True(t, firstHeaders.Get("Authorization") == "Bearer test-parent-ws-access")
	require.Equal(t, "test-parent-ws-chatgpt", firstHeaders.Get("chatgpt-account-id"))
	require.Equal(t, "response.create", gjson.Get(requestToJSONString(conn.writes[0]), "type").String())

	result, err = svc.Forward(context.Background(), newSparkShadowGinContext(), &next, body)
	require.NoError(t, err)
	require.Equal(t, "resp_spark_shadow_2", result.RequestID)
	require.True(t, dialer.lastHeaders.Get("Authorization") == "Bearer test-next-ws-access")
	require.Equal(t, "test-next-ws-chatgpt", dialer.lastHeaders.Get("chatgpt-account-id"))
	require.Len(t, conn.writes, 2)
	require.Equal(t, "response.create", gjson.Get(requestToJSONString(conn.writes[1]), "type").String())

	badChild := child
	badChild.ParentAccountID = sparkShadowAccountID(6299)
	decision := svc.getOpenAIWSProtocolResolver().Resolve(&badChild)
	recoveryTried := false
	_, err = svc.forwardOpenAIWSV2(
		context.Background(), newSparkShadowGinContext(), &badChild,
		map[string]any{"model": "gpt-5.1", "stream": false, "input": []any{}}, "test-child-token",
		decision, true, false, "gpt-5.1", "gpt-5.1", time.Now(), 1, "", &recoveryTried,
	)
	require.Error(t, err)
	require.Equal(t, 2, dialer.DialCount(), "ineligible shadow parent must not dial upstream websocket")
}

func TestOpenAIGatewayService_SparkShadowOutboundWSV2PassthroughUsesParentForAllFrames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	parent := newSparkShadowOAuthAccount(6301, "test-parent-passthrough-access", "test-parent-passthrough-chatgpt")
	child := newSparkShadowOAuthAccount(6302, "", "")
	child.ParentAccountID = sparkShadowAccountID(parent.ID)
	child.Credentials = nil
	child.Extra = map[string]any{"openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModePassthrough}

	cfg := newSparkShadowWSConfig()
	upstreamConn := &openAIWSCaptureConn{
		readDelays: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
		events: [][]byte{
			[]byte(`{"type":"response.completed","response":{"id":"resp_spark_passthrough_1","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`),
			[]byte(`{"type":"response.completed","response":{"id":"resp_spark_passthrough_2","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`),
		},
	}
	dialer := &openAIWSCaptureDialer{conn: upstreamConn}
	svc := &OpenAIGatewayService{
		cfg:                       cfg,
		httpUpstream:              &httpUpstreamRecorder{},
		cache:                     &stubGatewayCache{},
		accountRepo:               stubOpenAIAccountRepo{accounts: []Account{parent}},
		openaiWSResolver:          NewOpenAIWSProtocolResolver(cfg),
		toolCorrector:             NewCodexToolCorrector(),
		openaiWSPassthroughDialer: dialer,
	}
	token, _, err := svc.GetAccessToken(context.Background(), &child)
	require.NoError(t, err)

	serverErr := make(chan error, 1)
	wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientConn, acceptErr := websocket.Accept(w, r, &websocket.AcceptOptions{CompressionMode: websocket.CompressionContextTakeover})
		if acceptErr != nil {
			serverErr <- acceptErr
			return
		}
		defer func() { _ = clientConn.CloseNow() }()

		recorder := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(recorder)
		ginCtx.Request = r.Clone(r.Context())
		ginCtx.Request.Header = r.Header.Clone()
		ginCtx.Request.Header.Set("User-Agent", "codex_cli_rs/0.99.0")
		readCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		messageType, firstMessage, readErr := clientConn.Read(readCtx)
		cancel()
		if readErr != nil {
			serverErr <- readErr
			return
		}
		if messageType != websocket.MessageText {
			serverErr <- io.ErrUnexpectedEOF
			return
		}
		serverErr <- svc.ProxyResponsesWebSocketFromClient(r.Context(), ginCtx, clientConn, &child, token, firstMessage, nil)
	}))
	defer wsServer.Close()

	dialCtx, cancelDial := context.WithTimeout(context.Background(), 3*time.Second)
	clientConn, _, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(wsServer.URL, "http"), nil)
	cancelDial()
	require.NoError(t, err)
	defer func() { _ = clientConn.CloseNow() }()

	writeFrame := func(payload []byte) {
		writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		require.NoError(t, clientConn.Write(writeCtx, websocket.MessageText, payload))
	}
	readCompleted := func(wantID string) {
		readCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, event, readErr := clientConn.Read(readCtx)
		require.NoError(t, readErr)
		require.Equal(t, wantID, gjson.GetBytes(event, "response.id").String())
	}

	writeFrame([]byte(`{"type":"response.create","model":"gpt-5.1","stream":false}`))
	writeFrame([]byte(`{"type":"session.update","session":{"model":"gpt-5.1"}}`))
	writeFrame([]byte(`{"type":"response.create","stream":false,"input":[{"type":"input_text","text":"follow-up"}]}`))
	readCompleted("resp_spark_passthrough_1")
	readCompleted("resp_spark_passthrough_2")
	_ = clientConn.Close(websocket.StatusNormalClosure, "done")

	select {
	case err := <-serverErr:
		if err != nil {
			require.True(t, strings.Contains(err.Error(), "StatusNormalClosure") || strings.Contains(err.Error(), "openai ws connection closed"))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("等待 Spark shadow passthrough websocket 结束超时")
	}

	require.Equal(t, 1, dialer.DialCount())
	require.True(t, dialer.lastHeaders.Get("Authorization") == "Bearer test-parent-passthrough-access")
	require.Equal(t, "test-parent-passthrough-chatgpt", dialer.lastHeaders.Get("chatgpt-account-id"))
	require.Len(t, upstreamConn.writes, 3)
	require.Equal(t, "response.create", gjson.Get(requestToJSONString(upstreamConn.writes[0]), "type").String())
	require.Equal(t, "session.update", gjson.Get(requestToJSONString(upstreamConn.writes[1]), "type").String())
	require.Equal(t, "response.create", gjson.Get(requestToJSONString(upstreamConn.writes[2]), "type").String())

	missingParent := child
	missingParent.ParentAccountID = sparkShadowAccountID(6399)
	missingDialer := &openAIWSCaptureDialer{conn: &openAIWSCaptureConn{}}
	missingSvc := &OpenAIGatewayService{
		cfg:                       cfg,
		accountRepo:               stubOpenAIAccountRepo{},
		openaiWSResolver:          NewOpenAIWSProtocolResolver(cfg),
		openaiWSPassthroughDialer: missingDialer,
	}
	headers, _ := missingSvc.buildOpenAIWSHeaders(context.Background(), newSparkShadowGinContext(), &missingParent, "test-child-token", NewOpenAIWSProtocolResolver(cfg).Resolve(&missingParent), true, "", "", "", "gpt-5.1", "")
	require.Nil(t, headers)
	require.Equal(t, 0, missingDialer.DialCount(), "missing parent must not construct an outbound passthrough dial")

}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGroupFast20261103WSRealFirstAndFollowup(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, tier := range []string{"", "default", "flex", "ultrafast", "unknown"} {
			for _, action := range []string{BetaPolicyActionPass, BetaPolicyActionFilter} {
				t.Run(fmt.Sprintf("passthrough=%v/%s/%s", passthrough, tier, action), func(t *testing.T) {
					fragment := ""
					if tier != "" {
						fragment = fmt.Sprintf("\"service_tier\":\"%s\",", tier)
					}
					capture := runGroupFast20261103WSRelay(t, passthrough, fragment, action, true)
					require.Len(t, capture.writes, 2)
					for _, payload := range capture.writes {
						if action == BetaPolicyActionFilter {
							require.NotContains(t, payload, "service_tier")
						} else {
							require.Equal(t, "priority", payload["service_tier"])
						}
					}
				})
			}
		}
	}
}
func runGroupFast20261103WSRelay(t *testing.T, passthrough bool, firstTier string, action string, enabled bool) openAIWSUserScopedFastPolicyCapture {
	t.Helper()
	trustedUserID, spoofedUserID := int64(42), int64(42)
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	cfg.Gateway.OpenAIWS.Enabled = true
	cfg.Gateway.OpenAIWS.OAuthEnabled = true
	cfg.Gateway.OpenAIWS.APIKeyEnabled = true
	cfg.Gateway.OpenAIWS.ResponsesWebsocketsV2 = true
	cfg.Gateway.OpenAIWS.ModeRouterV2Enabled = passthrough
	cfg.Gateway.OpenAIWS.IngressModeDefault = OpenAIWSIngressModeCtxPool
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 1
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 1
	cfg.Gateway.OpenAIWS.QueueLimitPerConn = 8
	cfg.Gateway.OpenAIWS.DialTimeoutSeconds = 3
	cfg.Gateway.OpenAIWS.ReadTimeoutSeconds = 3
	cfg.Gateway.OpenAIWS.WriteTimeoutSeconds = 3

	upstreamConn := &openAIWSCaptureConn{
		readDelays: []time.Duration{0, 150 * time.Millisecond},
		events: [][]byte{
			[]byte(`{"type":"response.completed","response":{"id":"resp_user_scope_1","model":"gpt-5.5","usage":{"input_tokens":1,"output_tokens":1}}}`),
			[]byte(`{"type":"response.completed","response":{"id":"resp_user_scope_2","model":"gpt-5.5","usage":{"input_tokens":1,"output_tokens":1}}}`),
		},
	}
	captureDialer := &openAIWSCaptureDialer{conn: upstreamConn}
	pool := newOpenAIWSConnPool(cfg)
	pool.setClientDialerForTest(captureDialer)

	settings := &OpenAIFastPolicySettings{Rules: []OpenAIFastPolicyRule{{ServiceTier: OpenAIFastTierPriority, Action: action, Scope: BetaPolicyScopeAll}}}
	rawSettings, err := json.Marshal(settings)
	require.NoError(t, err)
	repo := &openAIFastPolicyRepoStub{values: map[string]string{SettingKeyOpenAIFastPolicySettings: string(rawSettings)}}
	svc := &OpenAIGatewayService{
		cfg:                       cfg,
		httpUpstream:              &httpUpstreamRecorder{},
		cache:                     &stubGatewayCache{},
		openaiWSResolver:          NewOpenAIWSProtocolResolver(cfg),
		toolCorrector:             NewCodexToolCorrector(),
		openaiWSPool:              pool,
		openaiWSPassthroughDialer: captureDialer,
		settingService:            NewSettingService(repo, cfg),
	}

	account := &Account{
		ID:          905,
		Name:        "openai-ws-user-scope",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "sk-test"},
		Extra: map[string]any{
			"responses_websockets_v2_enabled": true,
		},
	}
	if passthrough {
		account.Extra["openai_apikey_responses_websockets_v2_mode"] = OpenAIWSIngressModePassthrough
	}

	serverErrCh := make(chan error, 1)
	wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, acceptErr := coderws.Accept(w, r, &coderws.AcceptOptions{CompressionMode: coderws.CompressionContextTakeover})
		if acceptErr != nil {
			serverErrCh <- acceptErr
			return
		}
		defer func() { _ = conn.CloseNow() }()

		trustedCtx := context.WithValue(groupFast20261103Context(PlatformComposite, enabled), ctxkey.APIKeyUserID, trustedUserID)
		rec := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(rec)
		req := r.Clone(trustedCtx)
		req.Header = req.Header.Clone()
		req.Header.Set("User-Agent", "unit-test-agent/1.0")
		ginCtx.Request = req

		readCtx, cancel := context.WithTimeout(trustedCtx, 3*time.Second)
		msgType, firstMessage, readErr := conn.Read(readCtx)
		cancel()
		if readErr != nil {
			serverErrCh <- readErr
			return
		}
		if msgType != coderws.MessageText && msgType != coderws.MessageBinary {
			serverErrCh <- errors.New("unsupported websocket client message type")
			return
		}
		serverErrCh <- svc.ProxyResponsesWebSocketFromClient(trustedCtx, ginCtx, conn, account, "sk-test", firstMessage, nil)
	}))
	defer wsServer.Close()

	dialCtx, cancelDial := context.WithTimeout(context.Background(), 3*time.Second)
	clientConn, _, err := coderws.Dial(dialCtx, "ws"+strings.TrimPrefix(wsServer.URL, "http"), &coderws.DialOptions{
		HTTPHeader: http.Header{"x-sub2api-user-id": []string{fmt.Sprint(spoofedUserID)}},
	})
	cancelDial()
	require.NoError(t, err)
	defer func() { _ = clientConn.CloseNow() }()

	writeMessage := func(payload string) {
		writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		require.NoError(t, clientConn.Write(writeCtx, coderws.MessageText, []byte(payload)))
	}
	readMessage := func(responseID string) {
		readCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, event, readErr := clientConn.Read(readCtx)
		require.NoError(t, readErr)
		if action == BetaPolicyActionBlock {
			require.Equal(t, "error", gjson.GetBytes(event, "type").String())
			require.NotEmpty(t, gjson.GetBytes(event, "error.message").String())
		} else {
			require.Equal(t, responseID, gjson.GetBytes(event, "response.id").String())
		}
	}

	writeMessage(fmt.Sprintf(`{"type":"response.create","model":"gpt-5.5","stream":false,%s"user_id":%d,"metadata":{"user_id":%d}}`, firstTier, spoofedUserID, spoofedUserID))
	readMessage("resp_user_scope_1")
	if action != BetaPolicyActionBlock {
		writeMessage(fmt.Sprintf(`{"type":"response.create","stream":false,"user_id":%d,"metadata":{"user_id":%d},"previous_response_id":"resp_user_scope_1"}`, spoofedUserID, spoofedUserID))
		readMessage("resp_user_scope_2")
	}
	_ = clientConn.Close(coderws.StatusNormalClosure, "done")

	select {
	case serverErr := <-serverErrCh:
		if serverErr != nil && action != BetaPolicyActionBlock {
			require.Contains(t, serverErr.Error(), "StatusNormalClosure")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("waiting for user-scoped websocket relay to stop timed out")
	}
	svc.CloseOpenAIWSPool()

	captureDialer.mu.Lock()
	upstreamHeaders := cloneHeader(captureDialer.lastHeaders)
	captureDialer.mu.Unlock()
	return openAIWSUserScopedFastPolicyCapture{writes: upstreamConn.writes, headers: upstreamHeaders}
}

func TestGroupFast20261103WSBlockAndDisabled(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprintf("block/%v", passthrough), func(t *testing.T) {
			capture := runGroupFast20261103WSRelay(t, passthrough, "", BetaPolicyActionBlock, true)
			require.Empty(t, capture.writes)
		})
		t.Run(fmt.Sprintf("disabled/%v", passthrough), func(t *testing.T) {
			capture := runGroupFast20261103WSRelay(t, passthrough, "", BetaPolicyActionPass, false)
			require.Len(t, capture.writes, 2)
			for _, payload := range capture.writes {
				require.NotContains(t, payload, "service_tier")
			}
		})
	}
}

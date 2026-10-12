package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type v0215LoopbackWSDialer struct{ url string }

func (d *v0215LoopbackWSDialer) Dial(ctx context.Context, _ string, headers http.Header, _ string) (openAIWSClientConn, int, http.Header, error) {
	return newDefaultOpenAIWSClientDialer().Dial(ctx, d.url, headers, "")
}

func TestV0215WSLocalOAuthHistoryAndImageAdapter(t *testing.T) {
	for _, tc := range []struct {
		name, typ string
		lite      bool
	}{
		{"apikey", AccountTypeAPIKey, false},
		{"oauth", AccountTypeOAuth, false},
		{"oauth-lite", AccountTypeOAuth, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstreamBodies := make(chan []byte, 2)
			upstreamErrors := make(chan error, 1)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := coderws.Accept(w, r, nil)
				if err != nil {
					upstreamErrors <- err
					return
				}
				defer conn.CloseNow()
				ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
				defer cancel()
				for turn := 1; turn <= 2; turn++ {
					_, payload, err := conn.Read(ctx)
					if err != nil {
						upstreamErrors <- err
						return
					}
					upstreamBodies <- payload
					response := fmt.Sprintf(`{"type":"response.completed","response":{"id":"resp_local_%d","model":"gpt-5.4","usage":{"input_tokens":120,"output_tokens":70,"input_tokens_details":{"image_tokens":%d,"cached_tokens":5}},"tool_usage":{"image_gen":{"input_tokens_details":{"image_tokens":999},"output_tokens_details":{"image_tokens":55}}}}}`, turn, turn*20)
					if err := conn.Write(ctx, coderws.MessageText, []byte(response)); err != nil {
						upstreamErrors <- err
						return
					}
					if turn == 1 {
						duplicate := strings.Replace(response, "response.completed", "response.done", 1)
						if err := conn.Write(ctx, coderws.MessageText, []byte(duplicate)); err != nil {
							upstreamErrors <- err
							return
						}
					}
				}
				// Await the gateway's close without inventing another turn.
				_, _, _ = conn.Read(ctx)
			}))
			defer upstream.Close()
			cfg := &config.Config{}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			cfg.Gateway.OpenAIWS.Enabled = true
			cfg.Gateway.OpenAIWS.OAuthEnabled = true
			cfg.Gateway.OpenAIWS.APIKeyEnabled = true
			cfg.Gateway.OpenAIWS.ResponsesWebsocketsV2 = true
			cfg.Gateway.OpenAIWS.ModeRouterV2Enabled = true
			cfg.Gateway.OpenAIWS.DialTimeoutSeconds = 3
			cfg.Gateway.OpenAIWS.ReadTimeoutSeconds = 3
			cfg.Gateway.OpenAIWS.WriteTimeoutSeconds = 3
			svc := &OpenAIGatewayService{
				cfg: cfg, openaiWSResolver: NewOpenAIWSProtocolResolver(cfg),
				openaiWSPassthroughDialer: &v0215LoopbackWSDialer{url: "ws" + strings.TrimPrefix(upstream.URL, "http")},
			}
			account := v0215HTTPAccount(tc.typ)
			account.Status, account.Schedulable = StatusActive, true
			account.Extra["openai_apikey_responses_websockets_v2_mode"] = OpenAIWSIngressModePassthrough
			account.Extra["openai_oauth_responses_websockets_v2_mode"] = OpenAIWSIngressModePassthrough
			turns := make(chan *OpenAIForwardResult, 4)
			before := make(chan int, 4)
			finished := make(chan error, 1)
			hooks := &OpenAIWSIngressHooks{
				BeforeRequest: func(turn int, _ []byte, _ string) error { before <- turn; return nil },
				BeforeTurn:    func(_ int) error { return fmt.Errorf("passthrough must preserve its no-BeforeTurn lifecycle") },
				AfterTurn: func(_ int, result *OpenAIForwardResult, err error) {
					if result != nil && err == nil {
						turns <- result
					}
				},
			}
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := coderws.Accept(w, r, nil)
				if err != nil {
					finished <- err
					return
				}
				defer conn.CloseNow()
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = r
				ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
				defer cancel()
				_, first, err := conn.Read(ctx)
				if err != nil {
					finished <- err
					return
				}
				finished <- svc.ProxyResponsesWebSocketFromClient(ctx, c, conn, account, "fake-token", first, hooks)
			}))
			defer gateway.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			client, _, err := coderws.Dial(ctx, "ws"+strings.TrimPrefix(gateway.URL, "http"), nil)
			require.NoError(t, err)
			defer client.CloseNow()
			payload := `{"type":"response.create","model":"gpt-5.4","tools":[],"input":[{"type":"web_search_call","id":"ws_local","status":"completed"}]}`
			if tc.lite {
				payload = strings.Replace(payload, `"tools":[]`, `"client_metadata":{"`+responsesLiteWSMetadataKey+`":"true"},"tools":[]`, 1)
			}
			for turn := 1; turn <= 2; turn++ {
				require.NoError(t, client.Write(ctx, coderws.MessageText, []byte(payload)))
				_, _, err := client.Read(ctx)
				require.NoError(t, err)
				if turn == 1 {
					_, _, err = client.Read(ctx) // duplicate forwarded, never rebilled
					require.NoError(t, err)
				}
				var result *OpenAIForwardResult
				select {
				case result = <-turns:
				case <-ctx.Done():
					t.Fatal("missing billed turn")
				}
				require.Equal(t, turn*20, result.Usage.ImageInputTokens)
				require.Equal(t, 55, result.Usage.ImageOutputTokens)
				require.Equal(t, 5, result.Usage.CacheReadInputTokens)
				var sent []byte
				select {
				case sent = <-upstreamBodies:
				case <-ctx.Done():
					t.Fatal("missing local upstream payload")
				}
				path := "tools.0"
				if tc.lite {
					path = "input.1.tools.0"
					require.Empty(t, gjson.GetBytes(sent, "tools").Array())
				}
				expected := tc.typ == AccountTypeOAuth
				require.Equal(t, expected, gjson.GetBytes(sent, path+".type").String() == "web_search")
				if expected {
					require.Equal(t, "none", gjson.GetBytes(sent, "tool_choice").String())
					require.False(t, gjson.GetBytes(sent, path+".external_web_access").Bool())
				}
			}
			_ = client.Close(coderws.StatusNormalClosure, "done")
			select {
			case err := <-finished:
				require.NoError(t, err)
			case <-ctx.Done():
				t.Fatal("gateway did not finish")
			}
			require.Empty(t, turns, "duplicate terminal must not invoke another AfterTurn")
			select {
			case err := <-upstreamErrors:
				t.Fatal(err)
			default:
			}
			select {
			case turn := <-before:
				require.Equal(t, 2, turn, "passthrough subsequent turn reaches BeforeRequest")
			default:
				t.Fatal("second turn BeforeRequest not called")
			}
		})
	}
}

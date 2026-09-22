package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type v027HTTPUpstream struct{ client *http.Client }

func (u v027HTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return u.client.Do(req)
}

func (u v027HTTPUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.client.Do(req)
}

func TestV027OpenAIForwardsStrictDeveloperRoleToLocalhost(t *testing.T) {
	var upstreamBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		upstreamBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		require.Equal(t, "system", gjson.GetBytes(upstreamBody, "messages.0.role").String())
		require.Equal(t, "user", gjson.GetBytes(upstreamBody, "messages.1.role").String())
		require.Equal(t, "9007199254740993", gjson.GetBytes(upstreamBody, "unknown_numeric").Raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_v027","choices":[{"message":{"role":"assistant","content":"ok"}}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))
	}))
	defer upstream.Close()

	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"strict-model","unknown_numeric":9007199254740993,"messages":[{"role":"developer","content":"rules","unknown":{"keep":true}},{"role":"user","content":"ping"}]}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	account := &Account{
		ID: 2701,
		// A CN platform remains strict even when the deterministic fake upstream
		// runs on localhost; compatible OpenAI accounts use exact official hosts.
		Platform:    PlatformDeepseek,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "sk-v027", "base_url": upstream.URL},
	}
	svc := &OpenAIGatewayService{
		cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
			Enabled: false, AllowInsecureHTTP: true,
		}}},
		httpUpstream: v027HTTPUpstream{client: upstream.Client()},
	}

	result, err := svc.forwardAsRawChatCompletions(context.Background(), c, account, body, "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "developer", gjson.GetBytes(body, "messages.0.role").String())
	require.Equal(t, "system", gjson.GetBytes(upstreamBody, "messages.0.role").String())
}

func TestV027OpenAIChatRoleNormalizationScopeAndPreservation(t *testing.T) {
	body := []byte(`{"messages":[{"role":"system","content":"first"},{"role":"developer","content":[{"type":"text","text":"keep"}],"unknown":9007199254740993},{"role":"assistant","content":null},{"role":"tool","content":"done"}],"future":{"flag":true}}`)
	strict := &Account{Platform: PlatformDeepseek, Type: AccountTypeAPIKey}
	got, err := normalizeStrictChatDeveloperRoles(strict, "https://relay.example.test/v1/chat/completions", body)
	require.NoError(t, err)
	require.Equal(t, "system", gjson.GetBytes(got, "messages.1.role").String())
	require.Equal(t, "9007199254740993", gjson.GetBytes(got, "messages.1.unknown").Raw)
	require.Equal(t, "assistant", gjson.GetBytes(got, "messages.2.role").String())
	require.Equal(t, "developer", gjson.GetBytes(body, "messages.1.role").String())

	compatible := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	for _, targetURL := range []string{
		"https://api.openai.com/v1/chat/completions",
		"https://api.deepseek.com.example.test/v1/chat/completions",
		"https://api.deepseek.com@example.test/v1/chat/completions",
		"://invalid-url",
	} {
		untouched, normalizeErr := normalizeStrictChatDeveloperRoles(compatible, targetURL, body)
		require.NoError(t, normalizeErr)
		require.Equal(t, body, untouched)
	}
	for _, host := range []string{"api.deepseek.com", "api.kimi.com", "api.moonshot.cn", "api.moonshot.ai", "open.bigmodel.cn", "api.z.ai"} {
		normalized, normalizeErr := normalizeStrictChatDeveloperRoles(compatible, "https://"+host+"/v1/chat/completions", body)
		require.NoError(t, normalizeErr)
		require.Equal(t, "system", gjson.GetBytes(normalized, "messages.1.role").String())
	}
	oauth := &Account{Platform: PlatformDeepseek, Type: AccountTypeOAuth}
	untouched, err := normalizeStrictChatDeveloperRoles(oauth, "https://api.deepseek.com/v1/chat/completions", body)
	require.NoError(t, err)
	require.Equal(t, body, untouched)
}

func TestV027OpenAIChatRoleNormalizationRejectsInvalidStrictPayload(t *testing.T) {
	account := &Account{Platform: PlatformKimi, Type: AccountTypeAPIKey}
	for _, body := range [][]byte{[]byte(`null`), []byte(`[]`), []byte(`{"messages":"wrong"}`), []byte(`{"messages":[null]}`), []byte(`{"messages":[{"content":"missing role"}]}`)} {
		_, err := normalizeStrictChatDeveloperRoles(account, "https://api.kimi.com/v1/chat/completions", body)
		require.Error(t, err)
	}
}

func TestV027OpenAIManifestValidation(t *testing.T) {
	require.NoError(t, validateCodexModelsManifestEnvelope([]byte(`{"models":[]}`)))
	for _, body := range [][]byte{
		[]byte(`null`),
		[]byte(`{"Models":[]}`),
		[]byte(`{"models":null}`),
		[]byte(`{"models":{}}`),
		[]byte(`{"models":"[]"}`),
		[]byte(`{"models":[}`),
	} {
		require.Error(t, validateCodexModelsManifestEnvelope(body))
	}
}

type v027ResponseBindContextProbeCache struct {
	stubGatewayCache
	errs      []error
	values    []string
	deadlines []time.Time
}

func (c *v027ResponseBindContextProbeCache) SetSessionAccountID(ctx context.Context, groupID int64, sessionHash string, accountID int64, ttl time.Duration) error {
	c.errs = append(c.errs, ctx.Err())
	if value, _ := ctx.Value(v027ResponseBindContextKey{}).(string); value != "" {
		c.values = append(c.values, value)
	}
	if deadline, ok := ctx.Deadline(); ok {
		c.deadlines = append(c.deadlines, deadline)
	}
	return c.stubGatewayCache.SetSessionAccountID(ctx, groupID, sessionHash, accountID, ttl)
}

type v027ResponseBindContextKey struct{}

func TestV027OpenAIResponseAffinitySurvivesCancellationWithBoundedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	probe := &v027ResponseBindContextProbeCache{}
	svc := &OpenAIGatewayService{cache: probe}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	groupID := int64(2702)
	c.Set("api_key", &APIKey{GroupID: &groupID})
	requestCtx, cancelRequest := context.WithCancel(context.WithValue(context.Background(), v027ResponseBindContextKey{}, "trace-v027"))
	cancelRequest()
	startedAt := time.Now()

	svc.bindHTTPResponseAccount(requestCtx, c, &Account{ID: 2702}, "resp_v027")

	require.Len(t, probe.errs, 1)
	require.NoError(t, probe.errs[0])
	require.Equal(t, []string{"trace-v027"}, probe.values)
	require.Len(t, probe.deadlines, 1)
	require.True(t, probe.deadlines[0].After(startedAt))
	require.LessOrEqual(t, probe.deadlines[0].Sub(startedAt), openAIWSStateStoreRedisTimeout+100*time.Millisecond)
	boundAccountID, err := svc.getOpenAIWSStateStore().GetResponseAccount(context.Background(), groupID, "resp_v027")
	require.NoError(t, err)
	require.Equal(t, int64(2702), boundAccountID)

	require.NotPanics(t, func() { svc.bindHTTPResponseAccount(nil, nil, nil, "") })
}

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const deepSeekCompatModel = "deepseek-v4.1-flash"
const deepSeekCompatImage = "data:image/png;base64,QQ=="

func deepSeekCompatAccount(base, protocol string) *Account {
	return &Account{
		ID: 20261022, Name: "anonymous-deepseek", Platform: PlatformDeepseek,
		Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true,
		Concurrency: 1, GroupIDs: []int64{20261022},
		Credentials: map[string]any{
			"api_key": "synthetic-key", "base_url": base, "api_protocol": protocol,
			"model_mapping": map[string]any{deepSeekCompatModel: deepSeekCompatModel},
		},
	}
}

func deepSeekCompatService(account *Account, upstream HTTPUpstream, advanced bool) *OpenAIGatewayService {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	svc := newCNNativeGatewayTestService(upstream)
	svc.cfg.Security.URLAllowlist.Enabled = false
	svc.cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	svc.cfg.Gateway.Scheduling.LoadBatchEnabled = advanced
	svc.accountRepo = groupAwareStubOpenAIAccountRepo{stubOpenAIAccountRepo{accounts: []Account{*account}}}
	svc.cache = &schedulerTestGatewayCache{}
	svc.concurrencyService = NewConcurrencyService(schedulerTestConcurrencyCache{})
	if advanced {
		svc.rateLimitService = newOpenAIAdvancedSchedulerRateLimitService("true")
	}
	return svc
}

func deepSeekCompatSelect(svc *OpenAIGatewayService, capability OpenAIEndpointCapability, model string) (*AccountSelectionResult, error) {
	groupID := int64(20261022)
	selection, _, err := svc.SelectAccountWithSchedulerForCapability(
		context.Background(), &groupID, "", "", model, nil,
		OpenAIUpstreamTransportAny, capability, false, PlatformDeepseek,
	)
	return selection, err
}

func deepSeekCompatContext(path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

func deepSeekCompatReply(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/json")
	if strings.HasSuffix(path, "/chat/completions") {
		_, _ = io.WriteString(w, `{"id":"chatcmpl_anonymous","model":"deepseek-v4.1-flash","choices":[{"index":0,"message":{"role":"assistant","content":"anonymous-ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":2,"completion_tokens":1,"total_tokens":3}}`)
		return
	}
	_, _ = io.WriteString(w, `{"id":"resp_anonymous","object":"response","status":"completed","model":"deepseek-v4.1-flash","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"anonymous-ok"}]}],"usage":{"input_tokens":2,"output_tokens":1,"total_tokens":3}}`)
}

func TestDeepSeekCompat20261022AnonymousChatSchedulerForward(t *testing.T) {
	for _, advanced := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "advanced"}[advanced], func(t *testing.T) {
			var calls atomic.Int32
			var sent []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				sent, _ = io.ReadAll(r.Body)
				if r.URL.Path != "/v1/chat/completions" {
					http.Error(w, "expected chat", http.StatusBadRequest)
					return
				}
				deepSeekCompatReply(w, r.URL.Path)
			}))
			defer server.Close()
			account := deepSeekCompatAccount(server.URL, APIProtocolChatCompletions)
			account.Extra = map[string]any{"openai_responses_supported": false}
			svc := deepSeekCompatService(account, v027HTTPUpstream{client: server.Client()}, advanced)
			selection, err := deepSeekCompatSelect(svc, OpenAIEndpointCapabilityChatCompletions, deepSeekCompatModel)
			if err != nil {
				require.Zero(t, calls.Load(), "selection failure must precede upstream contact")
				t.Fatalf("anonymous screenshot symptom: scheduler excluded eligible DeepSeek; upstream_calls=%d: %v", calls.Load(), err)
			}
			defer selection.ReleaseFunc()
			require.Equal(t, account.ID, selection.Account.ID)
			body := []byte(`{"model":"deepseek-v4.1-flash","messages":[{"role":"user","content":"anonymous ping"}]}`)
			c, rec := deepSeekCompatContext("/v1/chat/completions", body)
			result, err := svc.ForwardAsChatCompletions(context.Background(), c, selection.Account, body, "", "")
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Contains(t, rec.Body.String(), "anonymous-ok")
			require.Equal(t, int32(1), calls.Load())
			require.Equal(t, deepSeekCompatModel, gjson.GetBytes(sent, "model").String())
			t.Logf("scheduler selected anonymous account; actualForward path=/v1/chat/completions upstream_calls=%d body=%s response=%s", calls.Load(), sent, rec.Body.String())
		})
	}
}

func TestDeepSeekCompat20261022NativeImagesSchedulerForward(t *testing.T) {
	for _, extra := range []struct {
		name string
		data map[string]any
	}{
		{"missing_metadata", nil},
		{"probe_false", map[string]any{"openai_responses_supported": false}},
		{"force_chat", map[string]any{"openai_responses_mode": "force_chat_completions", "openai_responses_supported": true}},
	} {
		for _, lifted := range []bool{false, true} {
			for _, advanced := range []bool{false, true} {
				t.Run(extra.name+map[bool]string{false: "/user", true: "/tool"}[lifted]+map[bool]string{false: "/legacy", true: "/advanced"}[advanced], func(t *testing.T) {
					var calls atomic.Int32
					var sent []byte
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						calls.Add(1)
						sent, _ = io.ReadAll(r.Body)
						valid := r.URL.Path == "/responses"
						input := gjson.GetBytes(sent, "input").Array()
						for _, item := range input {
							for _, part := range item.Get("content").Array() {
								if part.Get("type").String() == "input_image" {
									valid = valid && part.Get("image_url").Type == gjson.String && !part.Get("url").Exists()
								}
							}
						}
						if !valid {
							http.Error(w, "duplicate url, nonstring image_url, or wrong protocol", http.StatusBadRequest)
							return
						}
						deepSeekCompatReply(w, r.URL.Path)
					}))
					defer server.Close()
					account := deepSeekCompatAccount(server.URL, APIProtocolResponses)
					account.Extra = extra.data
					svc := deepSeekCompatService(account, v027HTTPUpstream{client: server.Client()}, advanced)
					selection, err := deepSeekCompatSelect(svc, OpenAIEndpointCapabilityResponses, deepSeekCompatModel)
					if err != nil {
						require.Zero(t, calls.Load())
						t.Fatalf("native image scheduler rejected account; upstream_calls=%d: %v", calls.Load(), err)
					}
					defer selection.ReleaseFunc()
					input := `[{"role":"user","content":[{"type":"input_text","text":"anonymous image"},{"type":"image_url","image_url":{"url":"` + deepSeekCompatImage + `"},"url":"https://unused.invalid/image"}]}]`
					if lifted {
						input = `[
						{"type":"function_call","call_id":"call_A","name":"view_image","arguments":"{}"},
						{"type":"function_call","call_id":"call_B","name":"view_image","arguments":"{}"},
						{"type":"function_call_output","call_id":"call_A","output":[{"type":"input_image","image_url":"` + deepSeekCompatImage + `","url":"https://unused.invalid/image"}]},
						{"role":"developer","content":"notice"},
						{"type":"function_call_output","call_id":"call_B","output":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"Qg=="}}]}
					]`
					}
					body := []byte(`{"model":"deepseek-v4.1-flash","store":true,"previous_response_id":"resp_old","trace_id":9007199254740993,"metadata":{"anonymous":true},"input":` + input + `}`)
					c, rec := deepSeekCompatContext("/v1/responses", body)
					result, err := svc.Forward(context.Background(), c, selection.Account, body)
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, http.StatusOK, rec.Code)
					require.Contains(t, rec.Body.String(), "anonymous-ok")
					require.Equal(t, int32(1), calls.Load())
					require.Equal(t, "9007199254740993", gjson.GetBytes(sent, "trace_id").Raw)
					require.True(t, gjson.GetBytes(sent, "metadata.anonymous").Bool())
					require.False(t, gjson.GetBytes(sent, "store").Bool())
					require.False(t, gjson.GetBytes(sent, "previous_response_id").Exists())
					if lifted {
						require.Equal(t, "function_call_output", gjson.GetBytes(sent, "input.2.type").String())
						require.Equal(t, "function_call_output", gjson.GetBytes(sent, "input.3.type").String())
						require.Equal(t, "call_A", gjson.GetBytes(sent, "input.2.call_id").String())
						require.Equal(t, "call_B", gjson.GetBytes(sent, "input.3.call_id").String())
						require.Equal(t, "user", gjson.GetBytes(sent, "input.4.role").String())
						require.Equal(t, deepSeekCompatImage, gjson.GetBytes(sent, "input.4.content.1.image_url").String())
						require.Equal(t, "data:image/png;base64,Qg==", gjson.GetBytes(sent, "input.4.content.3.image_url").String())
						require.Equal(t, "developer", gjson.GetBytes(sent, "input.5.role").String())
					} else {
						require.Equal(t, "input_image", gjson.GetBytes(sent, "input.0.content.1.type").String())
						require.Equal(t, deepSeekCompatImage, gjson.GetBytes(sent, "input.0.content.1.image_url").String())
					}
					t.Logf("scheduler->Forward->localhost path=/responses upstream_calls=%d body=%s response=%s", calls.Load(), sent, rec.Body.String())
				})
			}
		}
	}
}

func TestDeepSeekCompat20261022ChatProtocolOverridesStaleResponses(t *testing.T) {
	for _, extra := range []map[string]any{
		nil,
		{"openai_responses_supported": true},
		{"openai_responses_mode": "force_responses", "openai_responses_supported": true},
	} {
		for _, advanced := range []bool{false, true} {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.URL.Path != "/v1/chat/completions" {
					http.Error(w, "expected chat fallback", http.StatusBadRequest)
					return
				}
				deepSeekCompatReply(w, r.URL.Path)
			}))
			account := deepSeekCompatAccount(server.URL, APIProtocolChatCompletions)
			account.Extra = extra
			svc := deepSeekCompatService(account, v027HTTPUpstream{client: server.Client()}, advanced)
			selection, err := deepSeekCompatSelect(svc, OpenAIEndpointCapabilityResponses, deepSeekCompatModel)
			require.Error(t, err)
			require.Nil(t, selection)
			require.Zero(t, calls.Load())
			selection, err = deepSeekCompatSelect(svc, OpenAIEndpointCapabilityChatCompletions, deepSeekCompatModel)
			require.NoError(t, err)
			body := []byte(`{"model":"deepseek-v4.1-flash","input":"anonymous ping"}`)
			c, rec := deepSeekCompatContext("/v1/responses", body)
			result, err := svc.Forward(context.Background(), c, selection.Account, body)
			selection.ReleaseFunc()
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Contains(t, rec.Body.String(), "anonymous-ok")
			require.Equal(t, int32(1), calls.Load())
			server.Close()
		}
	}
}

func TestDeepSeekCompat20261022SchedulerNegativeZeroCalls(t *testing.T) {
	for _, advanced := range []bool{false, true} {
		for _, capability := range []OpenAIEndpointCapability{OpenAIEndpointCapabilityChatCompletions, OpenAIEndpointCapabilityResponses} {
			for _, tc := range []struct {
				name string
				edit func(*Account)
			}{
				{"endpoint_denial", func(a *Account) {
					a.Credentials["openai_capabilities"] = map[string]any{"chat_completions": false, "responses": false}
				}},
				{"missing_chat", func(a *Account) { a.Credentials["openai_capabilities"] = []string{"responses"} }},
				{"generic_denial", func(a *Account) { a.Extra = map[string]any{"supported_capabilities": []string{"image"}} }},
				{"wrong_platform", func(a *Account) { a.Platform = PlatformOpenAI }},
				{"wrong_group", func(a *Account) { a.GroupIDs = []int64{99} }},
				{"inactive", func(a *Account) { a.Status = "disabled" }},
				{"unschedulable", func(a *Account) { a.Schedulable = false }},
				{"rate_limited", func(a *Account) { until := time.Now().Add(time.Hour); a.RateLimitResetAt = &until }},
				{"unsupported_model", func(a *Account) { a.Credentials["model_mapping"] = map[string]any{"other": "other"} }},
				{"oauth", func(a *Account) { a.Type = AccountTypeOAuth }},
			} {
				t.Run(tc.name+"/"+string(capability)+map[bool]string{false: "/legacy", true: "/advanced"}[advanced], func(t *testing.T) {
					var calls atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
					defer server.Close()
					account := deepSeekCompatAccount(server.URL, APIProtocolResponses)
					tc.edit(account)
					svc := deepSeekCompatService(account, v027HTTPUpstream{client: server.Client()}, advanced)
					selection, err := deepSeekCompatSelect(svc, capability, deepSeekCompatModel)
					require.Error(t, err)
					require.Nil(t, selection)
					require.Zero(t, calls.Load())
				})
			}
		}
	}
}

func TestDeepSeekCompat20261022EndpointPolicy(t *testing.T) {
	for _, platform := range []string{PlatformDeepseek, PlatformKimi, PlatformZhipu} {
		for _, protocol := range []string{APIProtocolResponses, APIProtocolChatCompletions, APIProtocolAnthropic} {
			a := deepSeekCompatAccount("", protocol)
			a.Platform = platform
			require.True(t, a.IsOpenAICompatible())
			require.True(t, supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityChatCompletions))
			require.Equal(t, platform == PlatformDeepseek && protocol == APIProtocolResponses,
				supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityResponses))
			require.False(t, a.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityEmbeddings))
			require.False(t, a.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityLive))
			require.False(t, a.SupportsOpenAIImageCapability(OpenAIImagesCapabilityNative))
			a.Type = AccountTypeOAuth
			require.False(t, supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityChatCompletions))
			require.False(t, supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityResponses))
		}
	}
	for _, raw := range []any{[]any{}, []string{}, map[string]any{}, map[string]bool{}} {
		a := deepSeekCompatAccount("", APIProtocolResponses)
		a.Credentials["openai_capabilities"] = raw
		require.True(t, supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityResponses), "preserve empty-container semantics")
	}
	for _, raw := range []any{[]string{"chat_completions"}, map[string]any{"chat_completions": true, "responses": false}} {
		a := deepSeekCompatAccount("", APIProtocolResponses)
		a.Credentials["openai_capabilities"] = raw
		require.True(t, supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityChatCompletions))
		require.False(t, supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityResponses))
	}
	for _, extra := range []map[string]any{
		{"openai_responses_supported": false},
		{"openai_responses_mode": "force_chat_completions"},
	} {
		a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: extra}
		require.False(t, supportsOpenAIEndpointCapabilityForRequest(a, OpenAIEndpointCapabilityResponses))
	}
	require.False(t, supportsOpenAIEndpointCapabilityForRequest(&Account{Platform: PlatformGrok, Type: AccountTypeAPIKey}, OpenAIEndpointCapabilityResponses))
}

func TestDeepSeekCompat20261022ImageShapesAndScope(t *testing.T) {
	for _, part := range []string{
		`{"type":"input_image","image_url":"` + deepSeekCompatImage + `","url":"https://unused.invalid/image"}`,
		`{"type":"input_image","url":"` + deepSeekCompatImage + `"}`,
		`{"type":"input_image","image_url":{"url":"` + deepSeekCompatImage + `"}}`,
		`{"type":"image_url","image_url":{"url":"` + deepSeekCompatImage + `"}}`,
		`{"type":"image","source":{"type":"base64","media_type":"image/png","data":"QQ=="}}`,
		`{"type":"image","source":{"type":"url","url":"` + deepSeekCompatImage + `"}}`,
	} {
		body := []byte(`{"store":true,"trace_id":9007199254740993,"input":[{"role":"user","content":[` + part + `,{"type":"input_image","file_id":"file_anonymous"}]}]}`)
		a := deepSeekCompatAccount("", APIProtocolResponses)
		got := normalizeDeepSeekResponsesRequestBody(a, body)
		require.Equal(t, deepSeekCompatImage, gjson.GetBytes(got, "input.0.content.0.image_url").String())
		require.Equal(t, gjson.String, gjson.GetBytes(got, "input.0.content.0.image_url").Type)
		require.False(t, gjson.GetBytes(got, "input.0.content.0.url").Exists())
		require.Equal(t, "file_anonymous", gjson.GetBytes(got, "input.0.content.1.file_id").String())
		require.False(t, gjson.GetBytes(got, "input.0.content.1.image_url").Exists())
		require.Equal(t, "9007199254740993", gjson.GetBytes(got, "trace_id").Raw)
		require.Equal(t, got, normalizeDeepSeekResponsesRequestBody(a, got))
	}
	body := []byte(`{"store":true,"previous_response_id":"resp_keep","input":[{"role":"user","content":[{"type":"input_image","image_url":{"url":"` + deepSeekCompatImage + `"},"url":"unused"}]}]}`)
	for _, base := range []string{"https://api.deepseek.com", "https://API.DEEPSEEK.COM/v1"} {
		a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"base_url": base}}
		got := normalizeDeepSeekResponsesRequestBody(a, body)
		require.Equal(t, deepSeekCompatImage, gjson.GetBytes(got, "input.0.content.0.image_url").String())
		require.True(t, gjson.GetBytes(got, "store").Bool())
		require.Equal(t, "resp_keep", gjson.GetBytes(got, "previous_response_id").String())
	}
	for _, a := range []*Account{
		nil,
		{Platform: PlatformOpenAI, Credentials: map[string]any{"base_url": "https://api.deepseek.com.evil.invalid"}},
		{Platform: PlatformOpenAI, Credentials: map[string]any{"base_url": "https://sub.api.deepseek.com"}},
		{Platform: PlatformKimi, Credentials: map[string]any{"base_url": "https://api.deepseek.com"}},
		deepSeekCompatAccount("", APIProtocolChatCompletions),
	} {
		require.Equal(t, body, normalizeDeepSeekResponsesRequestBody(a, body))
	}
}

func TestDeepSeekCompat20261022OfficialHostForwardInMemory(t *testing.T) {
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(`{"id":"resp_anonymous","object":"response","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)),
	}}
	account := &Account{ID: 20261022, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "synthetic-key", "base_url": "https://api.deepseek.com"}}
	body := []byte(`{"model":"deepseek-v4.1-flash","store":true,"input":[{"role":"user","content":[{"type":"input_image","image_url":{"url":"` + deepSeekCompatImage + `"},"url":"unused"}]}]}`)
	svc := newCNNativeGatewayTestService(upstream)
	svc.cfg.Security.URLAllowlist.UpstreamHosts = []string{"api.deepseek.com"}
	c, _ := deepSeekCompatContext("/v1/responses", body)
	_, err := svc.Forward(context.Background(), c, account, body)
	require.NoError(t, err)
	require.Equal(t, "https://api.deepseek.com/v1/responses", upstream.lastReq.URL.String())
	require.Equal(t, deepSeekCompatImage, gjson.GetBytes(upstream.lastBody, "input.0.content.0.image_url").String())
	require.False(t, gjson.GetBytes(upstream.lastBody, "input.0.content.0.url").Exists())
	require.True(t, gjson.GetBytes(upstream.lastBody, "store").Bool())
}

func TestDeepSeekCompat20261022MappedSuffixForward(t *testing.T) {
	var sent []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sent, _ = io.ReadAll(r.Body)
		deepSeekCompatReply(w, r.URL.Path)
	}))
	defer server.Close()
	account := deepSeekCompatAccount(server.URL, APIProtocolChatCompletions)
	account.Extra = map[string]any{"openai_responses_supported": false}
	account.Credentials["model_mapping"] = map[string]any{"local-alias": deepSeekCompatModel + "[1M][1m]"}
	svc := deepSeekCompatService(account, v027HTTPUpstream{client: server.Client()}, false)
	selection, err := deepSeekCompatSelect(svc, OpenAIEndpointCapabilityChatCompletions, "local-alias")
	require.NoError(t, err)
	defer selection.ReleaseFunc()
	body := []byte(`{"model":"local-alias","messages":[{"role":"user","content":"ping"}]}`)
	c, _ := deepSeekCompatContext("/v1/chat/completions", body)
	result, err := svc.ForwardAsChatCompletions(context.Background(), c, selection.Account, body, "", "")
	require.NoError(t, err)
	require.Equal(t, deepSeekCompatModel, gjson.GetBytes(sent, "model").String())
	require.Equal(t, deepSeekCompatModel, result.UpstreamModel)
	require.Equal(t, deepSeekCompatModel, normalizeOpenAIModelForUpstream(account, " "+deepSeekCompatModel+"[1M][1m] "))
	require.Equal(t, "other[1m]", normalizeOpenAIModelForUpstream(&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, "other[1m]"))
	empty := deepSeekCompatAccount("", "")
	delete(empty.Credentials, "model_mapping")
	require.True(t, empty.IsModelSupported(deepSeekCompatModel))
	require.True(t, empty.IsModelSupported("local-future-alias"), "do not import upstream restrictive whitelist")
}

// Keep fixture JSON syntactically checked independently of gateway decoding.
func TestDeepSeekCompat20261022ImageFileIDUnrelatedFields(t *testing.T) {
	body := []byte(`{"input":[{"role":"user","content":[{"type":"input_image","file_id":"file_anonymous"},{"type":"input_text","text":"keep"}]}],"trace_id":9007199254740993,"extra":{"url":"keep"}}`)
	a := deepSeekCompatAccount("", APIProtocolResponses)
	got := normalizeDeepSeekResponsesRequestBody(a, body)
	var decoded map[string]any
	decoder := json.NewDecoder(bytes.NewReader(got))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&decoded))
	require.Equal(t, json.Number("9007199254740993"), decoded["trace_id"])
	require.Equal(t, "keep", gjson.GetBytes(got, "extra.url").String())
	require.False(t, gjson.GetBytes(got, "input.0.content.0.image_url").Exists())
}

func TestDeepSeekCompat20261022CNImageGenerationMediaSchedulerDenied(t *testing.T) {
	for _, target := range []struct{ name, platform, protocol string }{
		{"deepseek_chat", PlatformDeepseek, APIProtocolChatCompletions},
		{"deepseek_responses", PlatformDeepseek, APIProtocolResponses},
		{"kimi", PlatformKimi, APIProtocolChatCompletions},
		{"zhipu", PlatformZhipu, APIProtocolChatCompletions},
	} {
		for _, advanced := range []bool{false, true} {
			for _, capability := range []struct {
				name  string
				value any
			}{
				{"missing", nil}, {"empty", []string{}}, {"image", []string{"image"}},
			} {
				t.Run(fmt.Sprintf("%s/advanced=%t/%s", target.name, advanced, capability.name), func(t *testing.T) {
					var calls atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
					defer server.Close()
					account := deepSeekCompatAccount(server.URL, target.protocol)
					account.Platform = target.platform
					if capability.value != nil {
						account.Extra = map[string]any{"supported_capabilities": capability.value}
					}
					svc := deepSeekCompatService(account, v027HTTPUpstream{client: server.Client()}, advanced)
					groupID := int64(20261022)
					selection, _, err := svc.SelectAccountWithSchedulerForMediaCapabilityForUser(
						context.Background(), &groupID, "", "", deepSeekCompatModel, nil,
						OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityChatCompletions,
						AccountCapabilityImage, false, 0, target.platform,
					)
					selected := selection != nil && selection.Account != nil
					if selection != nil && selection.ReleaseFunc != nil {
						selection.ReleaseFunc()
					}
					require.Error(t, err, "CN Image generic requirement must be rejected regardless of metadata")
					require.False(t, selected)
					require.Zero(t, calls.Load())
					t.Logf("media scheduler generic=image platform=%s capabilities=%s selected=%t upstream_calls=%d", target.platform, capability.name, selected, calls.Load())
				})
			}
		}
	}
}

func TestDeepSeekCompat20261022ImageInputMediaSchedulerForward(t *testing.T) {
	for _, protocol := range []string{APIProtocolChatCompletions, APIProtocolResponses} {
		for _, advanced := range []bool{false, true} {
			for _, empty := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/advanced=%t/empty=%t", protocol, advanced, empty), func(t *testing.T) {
					var calls atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						calls.Add(1)
						deepSeekCompatReply(w, r.URL.Path)
					}))
					defer server.Close()
					account := deepSeekCompatAccount(server.URL, protocol)
					if empty {
						account.Extra = map[string]any{"supported_capabilities": []string{}}
					}
					svc := deepSeekCompatService(account, v027HTTPUpstream{client: server.Client()}, advanced)
					groupID := int64(20261022)
					body := []byte(`{"model":"deepseek-v4.1-flash","input":[{"role":"user","content":[{"type":"input_image","image_url":"data:image/png;base64,QQ=="}]}]}`)
					require.False(t, IsImageGenerationIntentForPlatform("/v1/responses", deepSeekCompatModel, body, PlatformDeepseek))
					selection, _, err := svc.SelectAccountWithSchedulerForMediaCapabilityForUser(
						context.Background(), &groupID, "", "", deepSeekCompatModel, nil,
						OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityChatCompletions,
						AccountCapabilityChat, false, 0, PlatformDeepseek,
					)
					require.NoError(t, err)
					require.NotNil(t, selection.Account)
					defer selection.ReleaseFunc()
					c, rec := deepSeekCompatContext("/v1/responses", body)
					_, err = svc.Forward(context.Background(), c, selection.Account, body)
					require.NoError(t, err)
					require.Contains(t, rec.Body.String(), "anonymous-ok")
					require.Equal(t, int32(1), calls.Load())
				})
			}
		}
	}
}

package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func newNonStreamingFailoverContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	return c, rec
}

func newNonStreamingFailoverService() *OpenAIGatewayService {
	return &OpenAIGatewayService{cfg: &config.Config{}}
}

func newNonStreamingFailoverAccount() *Account {
	return &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"pool_mode": true}}
}

func newNonStreamingSSEResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{
		"Content-Type": []string{"text/event-stream"}, "X-Request-Id": []string{"rid-nonstreaming-failed"},
	}}
}

func nonStreamingTerminalBody(eventType, data string) []byte {
	return []byte(strings.Join([]string{"event: " + eventType, "data: " + data, "", "data: [DONE]"}, "\n"))
}

func TestNonStreamingTerminalFailureFailover_ResponseFailedCapacity(t *testing.T) {
	c, rec := newNonStreamingFailoverContext(t)
	body := nonStreamingTerminalBody("response.failed", `{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model.","type":"invalid_request_error"}}`)
	result, err := newNonStreamingFailoverService().handleSSEToJSON(newNonStreamingSSEResponse(), c, newNonStreamingFailoverAccount(), body, "model", "model")
	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
	require.True(t, failoverErr.RetryableOnSameAccount)
	require.Contains(t, string(failoverErr.ResponseBody), "Selected model is at capacity")
	require.False(t, c.Writer.Written())
	require.Empty(t, rec.Body.String())
}

func TestNonStreamingTerminalFailureFailover_ErrorUsesConservativeClassifier(t *testing.T) {
	nonTransient := `{"type":"error","error":{"message":"upstream rejected request"}}`
	c, rec := newNonStreamingFailoverContext(t)
	result, err := newNonStreamingFailoverService().handleSSEToJSON(newNonStreamingSSEResponse(), c, newNonStreamingFailoverAccount(), nonStreamingTerminalBody("error", nonTransient), "model", "model")
	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.False(t, errors.As(err, &failoverErr))
	require.Equal(t, http.StatusBadGateway, rec.Code)

	transient := `{"type":"error","error":{"message":"Temporary upstream failure, please retry"}}`
	c, _ = newNonStreamingFailoverContext(t)
	_, err = newNonStreamingFailoverService().handleSSEToJSON(newNonStreamingSSEResponse(), c, newNonStreamingFailoverAccount(), nonStreamingTerminalBody("error", transient), "model", "model")
	require.ErrorAs(t, err, &failoverErr)
	require.False(t, c.Writer.Written())
}

func TestNonStreamingTerminalFailureFailover_PassthroughAndCommittedBoundaries(t *testing.T) {
	body := nonStreamingTerminalBody("response.failed", `{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model.","type":"invalid_request_error"}}`)
	c, rec := newNonStreamingFailoverContext(t)
	passResult, err := newNonStreamingFailoverService().handlePassthroughSSEToJSON(newNonStreamingSSEResponse(), c, newNonStreamingFailoverAccount(), body, "model", "model")
	require.Nil(t, passResult)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Empty(t, rec.Body.String())

	c, rec = newNonStreamingFailoverContext(t)
	MarkResponseCommitted(c)
	result, err := newNonStreamingFailoverService().handleSSEToJSON(newNonStreamingSSEResponse(), c, newNonStreamingFailoverAccount(), body, "model", "model")
	require.Nil(t, result)
	require.Error(t, err)
	require.False(t, errors.As(err, &failoverErr))
	require.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestNonStreamingTerminalFailureFailover_NilAccountProposesNothing(t *testing.T) {
	c, _ := newNonStreamingFailoverContext(t)
	payload := []byte(`{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model."}}`)
	require.Nil(t, newNonStreamingFailoverService().nonStreamingTerminalFailureFailover(c, newNonStreamingSSEResponse(), nil, false, "response.failed", payload, "Selected model is at capacity. Please try a different model."))
}

func TestOpenAIStreamErrorEventShouldFailover_UpstreamSemantics(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{"cyber policy is never retried", `{"type":"error","error":{"type":"safety_error","code":"cyber_policy","message":"blocked by policy"}}`, false},
		{"deactivated workspace is account state", `{"type":"error","error":{"code":"deactivated_workspace","message":"workspace unavailable"}}`, true},
		{"generic forbidden stays conservative", `{"type":"error","error":{"type":"permission_error","message":"forbidden"}}`, false},
		{"credential auth failure retries", `{"type":"error","error":{"type":"authentication_error","code":"invalid_api_key","message":"unauthorized"}}`, true},
		{"explicit 529 retries", `{"type":"error","error":{"status_code":529,"message":"overloaded"}}`, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := []byte(tc.data)
			require.Equal(t, tc.want, openAIStreamErrorEventShouldFailover(payload, extractOpenAISSEErrorMessage(payload)))
		})
	}
}

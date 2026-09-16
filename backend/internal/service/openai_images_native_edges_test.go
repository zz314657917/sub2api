package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const codexDirectImagesTestWebP = "UklGRiIAAABXRUJQVlA4IBYAAADQAQCdASoBAAEAAUAmJaQAA3AA/vuUAAA="

type codexDirectImagesErrorReader struct {
	data []byte
	err  error
}

func (r *codexDirectImagesErrorReader) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	return 0, r.err
}

type codexDirectImagesPlanRepo struct {
	AccountRepository
	calls []struct {
		scope string
		reset time.Time
	}
}

func (r *codexDirectImagesPlanRepo) SetModelRateLimit(_ context.Context, _ int64, scope string, reset time.Time) error {
	r.calls = append(r.calls, struct {
		scope string
		reset time.Time
	}{scope: scope, reset: reset})
	return nil
}

func (r *codexDirectImagesPlanRepo) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	return nil
}

func newCodexDirectImagesEdgeForward(t *testing.T, body []byte, upstream *httpUpstreamRecorder) (*OpenAIGatewayService, *Account, *gin.Context, *httptest.ResponseRecorder, *OpenAIImagesRequest) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, openAIImagesGenerationsEndpoint, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	account := &Account{ID: 81, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test"}}
	return svc, account, c, rec, parsed
}

func codexDirectImagesResponse(status int, contentType, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body))}
}

func TestCodexDirectImagesFallbackStopsAfterSingle404Or405(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusMethodNotAllowed} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"cat"}`)
			upstream := &httpUpstreamRecorder{responses: []*http.Response{
				codexDirectImagesResponse(status, "application/json", `{"error":{"message":"native unavailable"}}`),
				codexDirectImagesResponse(http.StatusNotFound, "application/json", `{"error":{"message":"responses unavailable"}}`),
			}}
			svc, account, c, _, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)

			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.Error(t, err)
			require.Nil(t, result)
			require.Len(t, upstream.requests, 2)
			require.Equal(t, "/backend-api/codex/images/generations", upstream.requests[0].URL.Path)
			require.Equal(t, "/backend-api/codex/responses", upstream.requests[1].URL.Path)
		})
	}
}

func TestCodexDirectImagesLegacyAndUnknownSnapshotsKeepResponsesDriver(t *testing.T) {
	t.Setenv("SUB2API_IMAGES_MAIN_MODEL", "gpt-5.6-sol")
	for _, model := range []string{"gpt-image-1", "gpt-image-future-snapshot"} {
		t.Run(model, func(t *testing.T) {
			body := []byte(`{"model":"` + model + `","prompt":"cat"}`)
			upstream := &httpUpstreamRecorder{resp: codexDirectImagesResponse(http.StatusOK, "text/event-stream", `data: {"type":"response.completed","response":{"output":[{"type":"image_generation_call","result":"eA=="}]}}`+"\n\n")}
			svc, account, c, _, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)

			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.NoError(t, err)
			require.Equal(t, model, result.UpstreamModel)
			require.Len(t, upstream.requests, 1)
			require.Equal(t, "/backend-api/codex/responses", upstream.lastReq.URL.Path)
			require.Equal(t, "gpt-5.6-sol", gjson.GetBytes(upstream.lastBody, "model").String())
			require.Equal(t, model, gjson.GetBytes(upstream.lastBody, "tools.0.model").String())
		})
	}
}

func TestCodexDirectImagesNativeEmptyAndHTTPFailureDoNotRetry(t *testing.T) {
	for _, tc := range []struct {
		name string
		resp *http.Response
	}{
		{"empty output", codexDirectImagesResponse(http.StatusOK, "application/json", `{"data":[]}`)},
		{"http error", codexDirectImagesResponse(http.StatusBadRequest, "application/json", `{"error":{"message":"invalid native option"}}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"cat"}`)
			upstream := &httpUpstreamRecorder{resp: tc.resp}
			svc, account, c, _, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)
			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.Error(t, err)
			require.Nil(t, result)
			require.Len(t, upstream.requests, 1)
		})
	}
}

func TestCodexDirectImagesRetryablePreOutputHTTPFailureIsDelegatedWithoutReplay(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2","prompt":"cat","stream":true}`)
	upstream := &httpUpstreamRecorder{resp: codexDirectImagesResponse(http.StatusServiceUnavailable, "application/json", `{"error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded."}}`)}
	svc, account, c, _, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusServiceUnavailable, failoverErr.StatusCode)
	require.True(t, failoverErr.RetryableOnSameAccount)
	require.Positive(t, failoverErr.SameAccountRetryLimit)
	require.Len(t, upstream.requests, 1, "one ForwardImages call must not replay before the gateway retry handler decides")
}

func TestCodexDirectImagesReadFailureBeforeOutputIsRetryable(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2","prompt":"cat","stream":true}`)
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}},
		Body:       io.NopCloser(&codexDirectImagesErrorReader{err: errors.New("upstream stream read failed")}),
	}}
	svc, account, c, _, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr, "pre-output stream transport failures need gateway failover/retry classification")
	require.Len(t, upstream.requests, 1)
}

func TestCodexDirectImagesDetectsPNGJPEGAndWebPDimensions(t *testing.T) {
	makeImage := func(t *testing.T, format string, width, height int) string {
		t.Helper()
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		img.Set(0, 0, color.RGBA{R: 0xff, A: 0xff})
		var out bytes.Buffer
		var err error
		switch format {
		case "png":
			err = png.Encode(&out, img)
		case "jpeg":
			err = jpeg.Encode(&out, img, nil)
		default:
			t.Fatalf("unsupported generated format %q", format)
		}
		require.NoError(t, err)
		return base64.StdEncoding.EncodeToString(out.Bytes())
	}

	for _, tc := range []struct {
		name   string
		format string
		b64    string
		size   string
	}{
		{"png", "png", makeImage(t, "png", 2, 3), "2x3"},
		{"jpeg", "jpeg", makeImage(t, "jpeg", 4, 5), "4x5"},
		{"webp fixture", "webp", codexDirectImagesTestWebP, "1x1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"cat","response_format":"url"}`)
			upstream := &httpUpstreamRecorder{resp: codexDirectImagesResponse(http.StatusOK, "application/json", `{"output_format":"`+tc.format+`","data":[{"b64_json":"`+tc.b64+`"}]}`)}
			svc, account, c, rec, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)
			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.NoError(t, err)
			require.Equal(t, 1, result.ImageCount)
			require.Equal(t, []string{tc.size}, result.ImageOutputSizes)
			require.Equal(t, tc.size, gjson.Get(rec.Body.String(), "data.0.size").String())
			require.Equal(t, "data:image/"+tc.format+";base64,"+tc.b64, gjson.Get(rec.Body.String(), "data.0.url").String())
		})
	}
}

func TestCodexDirectImagesPlanGateDistinguishesDriverAndImageCooldown(t *testing.T) {
	t.Setenv("SUB2API_IMAGES_MAIN_MODEL", "gpt-5.6-sol")
	repo := &codexDirectImagesPlanRepo{}
	for _, tc := range []struct {
		name          string
		requestModel  string
		rejectedModel string
		wantCooldown  bool
	}{
		{"responses driver", "gpt-image-1", "gpt-5.6-sol", false},
		{"native image", "gpt-image-2", "gpt-image-2", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo.calls = nil
			body := []byte(`{"model":"` + tc.requestModel + `","prompt":"cat"}`)
			upstream := &httpUpstreamRecorder{resp: codexDirectImagesResponse(http.StatusBadRequest, "application/json", `{"error":{"type":"invalid_request_error","message":"The '`+tc.rejectedModel+`' model is not supported when using Codex with a ChatGPT account."}}`)}
			svc, account, c, _, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)
			svc.rateLimitService = &RateLimitService{accountRepo: repo}

			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.Error(t, err)
			require.Nil(t, result)
			if tc.wantCooldown {
				require.Len(t, repo.calls, 1)
				require.Equal(t, "gpt-image-2", repo.calls[0].scope)
			} else {
				require.Empty(t, repo.calls)
			}
		})
	}
}

func TestCodexDirectImagesStreamingPreservesCompletedAndPartialUsageWithoutReplay(t *testing.T) {
	completed := `data: {"type":"image_generation.completed","b64_json":"` + nativeTestPNG + `","usage":{"input_tokens":3,"output_tokens":7,"output_tokens_details":{"image_tokens":7}}}` + "\n\n"
	partial := `data: {"type":"image_generation.partial_image","b64_json":"` + nativeTestPNG + `","usage":{"input_tokens":2,"output_tokens":5,"output_tokens_details":{"image_tokens":5}}}` + "\n\n"
	for _, tc := range []struct {
		name      string
		stream    string
		readerErr error
		wantCount int
		wantUsage int
	}{
		{"completed then upstream error", completed + `data: {"type":"error","error":{"message":"late failure"}}` + "\n\n", nil, 1, 7},
		{"partial then upstream error", partial + `data: {"type":"error","error":{"message":"late failure"}}` + "\n\n", nil, 0, 5},
		{"partial then read error", partial, errors.New("stream read failed"), 0, 5},
		{"partial then eof", partial, nil, 0, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"cat","stream":true}`)
			var reader io.Reader = strings.NewReader(tc.stream)
			if tc.readerErr != nil {
				reader = &codexDirectImagesErrorReader{data: []byte(tc.stream), err: tc.readerErr}
			}
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(reader)}}
			svc, account, c, rec, parsed := newCodexDirectImagesEdgeForward(t, body, upstream)

			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tc.wantCount, result.ImageCount)
			require.Equal(t, tc.wantUsage, result.Usage.OutputTokens)
			require.Len(t, upstream.requests, 1)
			if tc.wantCount == 0 {
				events := parseOpenAIImageTestSSEEvents(rec.Body.String())
				partialEvent, ok := findOpenAIImageTestSSEEvent(events, "image_generation.partial_image")
				require.True(t, ok, "partial output must already be written before its terminal failure")
				require.Equal(t, nativeTestPNG, gjson.Get(partialEvent.Data, "b64_json").String())
			}
		})
	}
}

package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const nativeTestPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScL9dwAAAABJRU5ErkJggg=="

type nativeImagesTestErrorReader struct{ err error }

func (r nativeImagesTestErrorReader) Read([]byte) (int, error) { return 0, r.err }

type nativeImagesDataThenErrorReader struct {
	data []byte
	err  error
}

func (r *nativeImagesDataThenErrorReader) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	return 0, r.err
}

type nativeImagesPartialFailWriter struct {
	gin.ResponseWriter
	err error
}

func (w *nativeImagesPartialFailWriter) Write(p []byte) (int, error) {
	n := len(p) / 2
	if n == 0 {
		n = 1
	}
	_, _ = w.ResponseWriter.Write(p[:n])
	return n, w.err
}

func newNativeImagesForward(t *testing.T, body []byte, response *http.Response) (*OpenAIForwardResult, error, *httptest.ResponseRecorder, *httpUpstreamRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	svc := &OpenAIGatewayService{httpUpstream: &httpUpstreamRecorder{resp: response}}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	account := &Account{ID: 11, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test"}}
	result, forwardErr := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
	return result, forwardErr, rec, svc.httpUpstream.(*httpUpstreamRecorder)
}

func TestCodexDirectImagesForwardJSONPublicModelMIMEAndDimensions(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2","prompt":"cat","response_format":"url"}`)
	result, err, rec, upstream := newNativeImagesForward(t, body, &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewBufferString(`{"model":"gpt-image-2","data":[{"b64_json":"` + nativeTestPNG + `","output_format":"png"}]}`))})
	require.NoError(t, err)
	require.Equal(t, "gpt-image-2", result.Model)
	require.Equal(t, "gpt-image-2", result.UpstreamModel)
	require.Equal(t, "/backend-api/codex/images/generations", upstream.lastReq.URL.Path)
	require.Equal(t, "gpt-image-2", gjson.GetBytes(upstream.lastBody, "model").String())
	require.Equal(t, "gpt-image-2", gjson.Get(rec.Body.String(), "model").String())
	require.Equal(t, "gpt-image-2", gjson.Get(rec.Body.String(), "data.0.model").String())
	require.Contains(t, gjson.Get(rec.Body.String(), "data.0.url").String(), "data:image/png;base64,")
	require.False(t, gjson.Get(rec.Body.String(), "data.0.b64_json").Exists())
	require.Equal(t, "1x1", gjson.Get(rec.Body.String(), "data.0.size").String())
}

func TestCodexDirectImagesForward404FallsBackOnce(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2","prompt":"cat"}`)
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	upstream := &httpUpstreamRecorder{responses: []*http.Response{{StatusCode: 404, Header: http.Header{}, Body: io.NopCloser(bytes.NewBufferString(`{"error":{"message":"missing"}}`))}, {StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(bytes.NewBufferString(`data: {"type":"response.completed","response":{"output":[{"type":"image_generation_call","result":"eA=="}]}}

`))}}}
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test"}}
	_, err = svc.ForwardImages(context.Background(), c, account, body, parsed, "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "/backend-api/codex/images/generations", upstream.requests[0].URL.Path)
	require.Equal(t, "/backend-api/codex/responses", upstream.requests[1].URL.Path)
}

func TestCodexDirectImagesForwardMappedModelJSONAndSSEOutputFormats(t *testing.T) {
	tests := []struct {
		name           string
		stream         bool
		responseFormat string
		upstream       string
	}{
		{
			name:           "json url uses root output format",
			responseFormat: "url",
			upstream:       `{"output_format":"webp","data":[{"b64_json":"` + nativeTestPNG + `"}]}`,
		},
		{
			name:           "json b64_json remains base64",
			responseFormat: "b64_json",
			upstream:       `{"output_format":"webp","data":[{"b64_json":"` + nativeTestPNG + `"}]}`,
		},
		{
			name:           "sse url uses event output format",
			stream:         true,
			responseFormat: "url",
			upstream:       `data: {"type":"image_generation.completed","b64_json":"` + nativeTestPNG + `","output_format":"webp"}` + "\n\n",
		},
		{
			name:           "sse b64_json remains base64",
			stream:         true,
			responseFormat: "b64_json",
			upstream:       `data: {"type":"image_generation.completed","b64_json":"` + nativeTestPNG + `","output_format":"webp"}` + "\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			contentType := "application/json"
			if tt.stream {
				contentType = "text/event-stream"
			}
			body := []byte(`{"model":"gpt-image-2","prompt":"cat","response_format":"` + tt.responseFormat + `","stream":` + strconv.FormatBool(tt.stream) + `}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {contentType}},
				Body:       io.NopCloser(strings.NewReader(tt.upstream)),
			}}
			svc := &OpenAIGatewayService{httpUpstream: upstream}
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			account := &Account{ID: 12, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{
				"access_token":  "test",
				"model_mapping": map[string]any{"gpt-image-2": "gpt-image-2.5-flare"},
			}}

			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.NoError(t, err)
			require.Equal(t, "gpt-image-2", result.Model)
			require.Equal(t, "gpt-image-2.5-flare", result.UpstreamModel)
			require.Equal(t, "/backend-api/codex/images/generations", upstream.lastReq.URL.Path)
			require.Equal(t, "gpt-image-2.5-flare", gjson.GetBytes(upstream.lastBody, "model").String())

			var output string
			if tt.stream {
				events := parseOpenAIImageTestSSEEvents(rec.Body.String())
				event, ok := findOpenAIImageTestSSEEvent(events, "image_generation.completed")
				require.True(t, ok)
				require.Equal(t, "gpt-image-2", gjson.Get(event.Data, "model").String())
				output = event.Data
			} else {
				require.Equal(t, "gpt-image-2", gjson.Get(rec.Body.String(), "model").String())
				require.Equal(t, "gpt-image-2", gjson.Get(rec.Body.String(), "data.0.model").String())
				output = gjson.Get(rec.Body.String(), "data.0").Raw
			}
			if tt.responseFormat == "url" {
				require.Equal(t, "data:image/webp;base64,"+nativeTestPNG, gjson.Get(output, "url").String())
				require.False(t, gjson.Get(output, "b64_json").Exists())
			} else {
				require.Equal(t, nativeTestPNG, gjson.Get(output, "b64_json").String())
				require.False(t, gjson.Get(output, "url").Exists())
			}
		})
	}
}

func TestCodexDirectImagesForwardMultipartEditMappedModelAndSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "gpt-image-2"))
	require.NoError(t, writer.WriteField("prompt", "replace the background"))
	require.NoError(t, writer.WriteField("stream", "true"))
	require.NoError(t, writer.WriteField("response_format", "url"))
	png, err := base64.StdEncoding.DecodeString(nativeTestPNG)
	require.NoError(t, err)
	imageHeader := make(textproto.MIMEHeader)
	imageHeader.Set("Content-Disposition", `form-data; name="image"; filename="source.png"`)
	imageHeader.Set("Content-Type", "image/png")
	image, err := writer.CreatePart(imageHeader)
	require.NoError(t, err)
	_, err = image.Write(png)
	require.NoError(t, err)
	maskHeader := make(textproto.MIMEHeader)
	maskHeader.Set("Content-Disposition", `form-data; name="mask"; filename="mask.png"`)
	maskHeader.Set("Content-Type", "image/png")
	mask, err := writer.CreatePart(maskHeader)
	require.NoError(t, err)
	_, err = mask.Write(png)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", bytes.NewReader(body.Bytes()))
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(`data: {"type":"image_generation.completed","b64_json":"` + nativeTestPNG + `","output_format":"png"}` + "\n\n")),
	}}
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body.Bytes())
	require.NoError(t, err)
	account := &Account{ID: 13, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{
		"access_token":  "test",
		"model_mapping": map[string]any{"gpt-image-2": "gpt-image-2.5-flare"},
	}}

	result, err := svc.ForwardImages(context.Background(), c, account, body.Bytes(), parsed, "")
	require.NoError(t, err)
	require.True(t, result.Stream)
	require.Equal(t, "gpt-image-2", result.Model)
	require.Equal(t, "gpt-image-2.5-flare", result.UpstreamModel)
	require.Equal(t, "/backend-api/codex/images/edits", upstream.lastReq.URL.Path)
	require.Equal(t, "gpt-image-2.5-flare", gjson.GetBytes(upstream.lastBody, "model").String())
	require.Equal(t, "data:image/png;base64,"+nativeTestPNG, gjson.GetBytes(upstream.lastBody, "images.0.image_url").String())
	require.Equal(t, "data:image/png;base64,"+nativeTestPNG, gjson.GetBytes(upstream.lastBody, "mask.image_url").String())
	events := parseOpenAIImageTestSSEEvents(rec.Body.String())
	event, ok := findOpenAIImageTestSSEEvent(events, "image_edit.completed")
	require.True(t, ok)
	require.Equal(t, "image_edit.completed", gjson.Get(event.Data, "type").String())
	require.Equal(t, "gpt-image-2", gjson.Get(event.Data, "model").String())
	require.Equal(t, "data:image/png;base64,"+nativeTestPNG, gjson.Get(event.Data, "url").String())
	require.False(t, gjson.Get(event.Data, "b64_json").Exists())
}

func TestCodexDirectImagesForwardUsageCacheDetailFallbackJSONAndSSE(t *testing.T) {
	tests := []struct {
		name      string
		stream    bool
		usage     string
		wantTotal int
		wantImage int
	}{
		{
			name:      "json derives total when explicit total is absent",
			usage:     `{"input_tokens":100,"input_tokens_details":{"image_tokens":40,"cached_tokens_details":{"text_tokens":20,"image_tokens":30}}}`,
			wantTotal: 50,
			wantImage: 30,
		},
		{
			name:      "sse derives total when explicit total is absent",
			stream:    true,
			usage:     `{"input_tokens":100,"input_tokens_details":{"image_tokens":40,"cached_tokens_details":{"text_tokens":20,"image_tokens":30}}}`,
			wantTotal: 50,
			wantImage: 30,
		},
		{
			name:      "json preserves explicit zero total",
			usage:     `{"input_tokens":100,"input_tokens_details":{"image_tokens":40,"cached_tokens":0,"cached_tokens_details":{"text_tokens":20,"image_tokens":30}}}`,
			wantTotal: 0,
			wantImage: 0,
		},
		{
			name:      "sse preserves explicit zero total",
			stream:    true,
			usage:     `{"input_tokens":100,"input_tokens_details":{"image_tokens":40,"cached_tokens":0,"cached_tokens_details":{"text_tokens":20,"image_tokens":30}}}`,
			wantTotal: 0,
			wantImage: 0,
		},
		{
			name:      "json bounds unreasonable image cache by explicit total and image input",
			usage:     `{"input_tokens":100,"input_tokens_details":{"image_tokens":999,"cached_tokens":25,"cached_tokens_details":{"text_tokens":999,"image_tokens":999}}}`,
			wantTotal: 25,
			wantImage: 25,
		},
		{
			name:      "sse bounds unreasonable image cache by explicit total and image input",
			stream:    true,
			usage:     `{"input_tokens":100,"input_tokens_details":{"image_tokens":999,"cached_tokens":25,"cached_tokens_details":{"text_tokens":999,"image_tokens":999}}}`,
			wantTotal: 25,
			wantImage: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"cache test","stream":` + strconv.FormatBool(tt.stream) + `}`)
			contentType := "application/json"
			upstreamBody := `{"usage":` + tt.usage + `,"data":[{"b64_json":"` + nativeTestPNG + `"}]}`
			if tt.stream {
				contentType = "text/event-stream"
				upstreamBody = `data: {"type":"image_generation.completed","b64_json":"` + nativeTestPNG + `","usage":` + tt.usage + `}` + "\n\n"
			}
			result, err, _, _ := newNativeImagesForward(t, body, &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {contentType}},
				Body:       io.NopCloser(strings.NewReader(upstreamBody)),
			})
			require.NoError(t, err)
			require.Equal(t, tt.wantTotal, result.Usage.CacheReadInputTokens)
			require.Equal(t, tt.wantImage, result.Usage.ImageCacheReadTokens)
		})
	}
}

func TestCodexDirectImagesForwardReadFailureBeforeOutputIsRetryableJSONAndSSE(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(strconv.FormatBool(stream), func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"read failure","stream":` + strconv.FormatBool(stream) + `}`)
			contentType := "application/json"
			if stream {
				contentType = "text/event-stream"
			}
			result, err, _, upstream := newNativeImagesForward(t, body, &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {contentType}},
				Body:       io.NopCloser(nativeImagesTestErrorReader{err: errors.New("upstream read failed")}),
			})
			require.Nil(t, result)
			var failoverErr *UpstreamFailoverError
			require.ErrorAs(t, err, &failoverErr)
			require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
			require.True(t, failoverErr.RetryableOnSameAccount)
			require.Len(t, upstream.requests, 1)
		})
	}
}

func TestCodexDirectImagesForwardDrainsCompletedUsageAfterClientDisconnect(t *testing.T) {
	for _, firstEvent := range []string{"image_generation.started", "image_generation.partial_image"} {
		t.Run(firstEvent, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			body := []byte(`{"model":"gpt-image-2","prompt":"drain","stream":true}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Writer = &nativeImagesPartialFailWriter{ResponseWriter: c.Writer, err: errors.New("client disconnected")}
			usage := `{"input_tokens":10,"output_tokens":7,"input_tokens_details":{"image_tokens":4,"cached_tokens_details":{"text_tokens":2,"image_tokens":3}},"output_tokens_details":{"image_tokens":7}}`
			firstPayload := `{"type":"` + firstEvent + `"}`
			if firstEvent == "image_generation.partial_image" {
				firstPayload = `{"type":"` + firstEvent + `","b64_json":"` + nativeTestPNG + `"}`
			}
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader(
					"data: " + firstPayload + "\n\n" +
						`data: {"type":"image_generation.completed","b64_json":"` + nativeTestPNG + `","usage":` + usage + `}` + "\n\n",
				)),
			}}
			svc := &OpenAIGatewayService{httpUpstream: upstream}
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			account := &Account{ID: 14, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test"}}

			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, 1, result.ImageCount)
			require.Equal(t, 7, result.Usage.OutputTokens)
			require.Equal(t, 5, result.Usage.CacheReadInputTokens)
			require.Equal(t, 3, result.Usage.ImageCacheReadTokens)
			require.Len(t, upstream.requests, 1)
		})
	}
}

func TestCodexDirectImagesForwardWritesErrorEventForNativeFailures(t *testing.T) {
	for _, failure := range []struct {
		name    string
		payload string
	}{
		{
			name:    "error",
			payload: `{"type":"error","error":{"type":"server_error","message":"native failure"}}`,
		},
		{
			name:    "response failed",
			payload: `{"type":"response.failed","response":{"id":"resp_native_failure","error":{"type":"server_error","message":"native failure"}}}`,
		},
	} {
		t.Run(failure.name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"failure","stream":true}`)
			stream := `data: {"type":"image_generation.partial_image","b64_json":"` + nativeTestPNG + `"}` + "\n\n" + "data: " + failure.payload + "\n\n"
			result, err, rec, upstream := newNativeImagesForward(t, body, &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"text/event-stream"}},
				Body:       io.NopCloser(strings.NewReader(stream)),
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Zero(t, result.ImageCount)
			require.Len(t, upstream.requests, 1)
			events := parseOpenAIImageTestSSEEvents(rec.Body.String())
			event, ok := findOpenAIImageTestSSEEvent(events, "error")
			require.True(t, ok)
			require.Equal(t, "error", gjson.Get(event.Data, "type").String())
			require.Equal(t, "native failure", gjson.Get(event.Data, "error.message").String())
		})
	}
}

func TestCodexDirectImagesStreamingPartialClientWritePreventsReadFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)
	clientWriteErr := errors.New("client partial write failed")
	c.Writer = &nativeImagesPartialFailWriter{ResponseWriter: c.Writer, err: clientWriteErr}
	upstreamReadErr := errors.New("upstream read failed after client write")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}},
		Body: io.NopCloser(&nativeImagesDataThenErrorReader{
			data: []byte(`data: {"type":"image_generation.started"}` + "\n\n"),
			err:  upstreamReadErr,
		}),
	}

	_, count, _, _, err := (&OpenAIGatewayService{}).handleCodexDirectImagesStreamingResponse(resp, c, time.Now(), &OpenAIImagesRequest{Model: "gpt-image-2", Stream: true})
	require.Zero(t, count)
	var writtenErr *codexDirectImagesOutputWrittenError
	require.ErrorAs(t, err, &writtenErr)
	require.ErrorIs(t, err, clientWriteErr)
	var failoverErr *UpstreamFailoverError
	require.NotErrorAs(t, err, &failoverErr)
	require.False(t, errors.Is(err, upstreamReadErr), "the earlier client write error must not be replaced by a later upstream read error")
}

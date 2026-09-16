package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type codexDirectImagesOutputWrittenError struct{ err error }

func (e *codexDirectImagesOutputWrittenError) Error() string { return e.err.Error() }
func (e *codexDirectImagesOutputWrittenError) Unwrap() error { return e.err }

// A transport failure is safe to fail over only before this adapter has written
// anything downstream. Request cancellation, client cancellation, and local
// response-size enforcement are not upstream failures.
func newCodexDirectImagesPreOutputReadFailover(err error, c *gin.Context) error {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrUpstreamResponseBodyTooLarge) {
		return err
	}
	if c != nil && c.Request != nil && c.Request.Context().Err() != nil {
		return err
	}
	return &UpstreamFailoverError{
		StatusCode:             http.StatusBadGateway,
		RetryableOnSameAccount: true,
	}
}

type openAIImagesForceResponsesContextKey struct{}

func withOpenAIImagesForceResponses(ctx context.Context) context.Context {
	return context.WithValue(ctx, openAIImagesForceResponsesContextKey{}, true)
}

func isOpenAIImagesForceResponses(ctx context.Context) bool {
	forced, _ := ctx.Value(openAIImagesForceResponsesContextKey{}).(bool)
	return forced
}

// Keep this allowlist intentionally narrow: an unknown account mapping remains
// on the established Responses driver instead of silently selecting a new API.
func usesCodexDirectImages(model string) bool {
	switch strings.TrimSpace(model) {
	case "gpt-image-1.5", "gpt-image-2", "gpt-image-2.5-flare", "gpt-image-2.5-sunburst", "gpt-image-2.5-flare-2026-09-08", "gpt-image-2.5-sunburst-2026-09-08":
		return true
	default:
		return false
	}
}

func buildOpenAIImagesOAuthPayload(parsed *OpenAIImagesRequest, model string) ([]byte, string, error) {
	if parsed == nil || strings.TrimSpace(parsed.Prompt) == "" {
		return nil, "", fmt.Errorf("prompt is required")
	}
	body := []byte(`{}`)
	if !parsed.Multipart && gjson.ValidBytes(parsed.Body) {
		body = append([]byte(nil), parsed.Body...)
	}
	body, _ = sjson.SetBytes(body, "model", model)
	body, _ = sjson.SetBytes(body, "prompt", parsed.Prompt)
	body, _ = sjson.SetBytes(body, "n", max(parsed.N, 1))
	for _, field := range []struct{ key, value string }{{"size", parsed.Size}, {"quality", parsed.Quality}, {"background", parsed.Background}, {"output_format", parsed.OutputFormat}, {"moderation", parsed.Moderation}, {"input_fidelity", parsed.InputFidelity}, {"style", parsed.Style}} {
		if field.value != "" {
			body, _ = sjson.SetBytes(body, field.key, field.value)
		}
	}
	if parsed.OutputCompression != nil {
		body, _ = sjson.SetBytes(body, "output_compression", *parsed.OutputCompression)
	}
	if parsed.PartialImages != nil {
		body, _ = sjson.SetBytes(body, "partial_images", *parsed.PartialImages)
	}
	if parsed.Stream {
		body, _ = sjson.SetBytes(body, "stream", true)
	} else {
		body, _ = sjson.DeleteBytes(body, "stream")
	}
	body, _ = sjson.DeleteBytes(body, "response_format")
	endpoint := "/images/generations"
	if parsed.IsEdits() {
		endpoint = "/images/edits"
		images := make([]map[string]string, 0, len(parsed.InputImageURLs)+len(parsed.Uploads))
		for _, value := range parsed.InputImageURLs {
			if value = strings.TrimSpace(value); value != "" {
				images = append(images, map[string]string{"image_url": value})
			}
		}
		for _, upload := range parsed.Uploads {
			value, err := openAIImageUploadToDataURL(upload)
			if err != nil {
				return nil, "", err
			}
			images = append(images, map[string]string{"image_url": value})
		}
		if len(images) == 0 {
			return nil, "", fmt.Errorf("image input is required")
		}
		body, _ = sjson.SetBytes(body, "images", images)
		mask := parsed.MaskImageURL
		if parsed.MaskUpload != nil {
			var err error
			mask, err = openAIImageUploadToDataURL(*parsed.MaskUpload)
			if err != nil {
				return nil, "", err
			}
		}
		if strings.TrimSpace(mask) != "" {
			body, _ = sjson.SetBytes(body, "mask.image_url", mask)
		}
	}
	return body, strings.TrimSuffix(chatgptCodexURL, "/responses") + endpoint, nil
}

func codexDirectImagesUsage(body []byte) (OpenAIUsage, bool) {
	value := gjson.GetBytes(body, "usage")
	usage, ok := openAIUsageFromGJSON(value)
	if !ok {
		return usage, false
	}
	if !value.Get("output_tokens_details.image_tokens").Exists() {
		usage.ImageOutputTokens = usage.OutputTokens
	}
	cached := value.Get("input_tokens_details.cached_tokens_details")
	if !value.Get("input_tokens_details.cached_tokens").Exists() && cached.IsObject() {
		imageTokens, _ := boundedJSONNonNegativeInt(cached.Get("image_tokens"))
		textTokens, _ := boundedJSONNonNegativeInt(cached.Get("text_tokens"))
		usage.CacheReadInputTokens = min(imageTokens, max(usage.InputTokens, 0))
		usage.CacheReadInputTokens += min(textTokens, max(usage.InputTokens-usage.CacheReadInputTokens, 0))
	}
	imageCached, _ := boundedJSONNonNegativeInt(cached.Get("image_tokens"))
	usage.ImageCacheReadTokens = min(imageCached, max(usage.ImageInputTokens, 0), max(usage.CacheReadInputTokens, 0))
	return usage, true
}

func isOpenAIImagesMainModelError(status int, body []byte) bool {
	message := extractUpstreamErrorMessage(body)
	return isOpenAIImagesCodexPlanGatedModelError(status, body, openAIImagesResponsesMainModelValue()) && (strings.Contains(message, "'"+openAIImagesResponsesMainModelValue()+"'") || strings.Contains(message, `"`+openAIImagesResponsesMainModelValue()+`"`))
}

// parseCodexDirectImagesJSONResult parses the non-stream native Images result
// shape shared by the public forwarder and account connection tests. Native
// output_format is item-scoped when available, otherwise response-scoped, then
// falls back to the requested output format.
func parseCodexDirectImagesJSONResult(body []byte, defaultOutputFormat string) ([]openAIResponsesImageResult, error) {
	if !gjson.ValidBytes(body) {
		return nil, &OpenAIImagesUpstreamError{StatusCode: http.StatusBadGateway, ErrorType: "upstream_error", Message: "Images API returned malformed JSON"}
	}
	if upstreamErr := openAIImagesUpstreamErrorFromSSEPayload(body); upstreamErr != nil {
		return nil, upstreamErr
	}
	items := gjson.GetBytes(body, "data").Array()
	rootOutputFormat := strings.TrimSpace(gjson.GetBytes(body, "output_format").String())
	results := make([]openAIResponsesImageResult, 0, len(items))
	for _, item := range items {
		result := strings.TrimSpace(item.Get("b64_json").String())
		if result == "" {
			continue
		}
		outputFormat := strings.TrimSpace(item.Get("output_format").String())
		if outputFormat == "" {
			outputFormat = rootOutputFormat
		}
		if outputFormat == "" {
			outputFormat = strings.TrimSpace(defaultOutputFormat)
		}
		results = append(results, openAIResponsesImageResult{
			Result:        result,
			RevisedPrompt: strings.TrimSpace(item.Get("revised_prompt").String()),
			OutputFormat:  outputFormat,
		})
	}
	if len(results) == 0 {
		return nil, &OpenAIImagesUpstreamError{StatusCode: http.StatusBadGateway, ErrorType: "upstream_error", Message: "Images API returned no image output"}
	}
	return results, nil
}

func (s *OpenAIGatewayService) handleCodexDirectImagesNonStreamingResponse(resp *http.Response, c *gin.Context, parsed *OpenAIImagesRequest) (OpenAIUsage, int, []string, error) {
	body, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	if err != nil {
		return OpenAIUsage{}, 0, nil, newCodexDirectImagesPreOutputReadFailover(err, c)
	}
	items := gjson.GetBytes(body, "data").Array()
	results, err := parseCodexDirectImagesJSONResult(body, parsed.OutputFormat)
	if err != nil {
		return OpenAIUsage{}, 0, nil, err
	}
	count := 0
	sizes := make([]string, 0, len(items))
	resultIndex := 0
	for i, item := range items {
		if strings.TrimSpace(item.Get("b64_json").String()) == "" {
			continue
		}
		result := results[resultIndex]
		resultIndex++
		b64 := result.Result
		count++
		body, _ = sjson.SetBytes(body, fmt.Sprintf("data.%d.model", i), parsed.Model)
		if size := detectOpenAIImageResultSize(b64); size != "" {
			body, _ = sjson.SetBytes(body, fmt.Sprintf("data.%d.size", i), size)
			sizes = append(sizes, size)
		}
		if parsed.ResponseFormat == "url" {
			body, _ = sjson.SetBytes(body, fmt.Sprintf("data.%d.url", i), "data:"+openAIImageOutputMIMEType(result.OutputFormat)+";base64,"+b64)
			body, _ = sjson.DeleteBytes(body, fmt.Sprintf("data.%d.b64_json", i))
		}
	}
	if count == 0 {
		return OpenAIUsage{}, 0, nil, &OpenAIImagesUpstreamError{StatusCode: http.StatusBadGateway, ErrorType: "upstream_error", Message: "Images API returned no image output"}
	}
	usage, _ := codexDirectImagesUsage(body)
	body, _ = sjson.SetBytes(body, "model", parsed.Model)
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	c.Data(resp.StatusCode, "application/json", body)
	return usage, count, sizes, nil
}

func (s *OpenAIGatewayService) handleCodexDirectImagesStreamingResponse(resp *http.Response, c *gin.Context, start time.Time, parsed *OpenAIImagesRequest) (OpenAIUsage, int, []string, *int, error) {
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	c.Header("Content-Type", "text/event-stream")
	c.Status(resp.StatusCode)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return OpenAIUsage{}, 0, nil, nil, fmt.Errorf("streaming is not supported by response writer")
	}
	var usage OpenAIUsage
	count := 0
	downstreamWriteAttempted := false
	clientDisconnected := false
	var clientWriteErr error
	var sizes []string
	var first *int
	var accumulator openAISSEDataAccumulator
	var processErr error
	process := func(data []byte) {
		if processErr != nil || bytes.Equal(bytes.TrimSpace(data), []byte("[DONE]")) {
			return
		}
		if first == nil {
			ms := int(time.Since(start).Milliseconds())
			first = &ms
		}
		if !gjson.ValidBytes(data) {
			processErr = fmt.Errorf("invalid image stream JSON")
			return
		}
		if directUsage, ok := codexDirectImagesUsage(data); ok {
			mergeOpenAIUsageNonZero(&usage, directUsage)
			if directUsage.ImageCacheReadTokens > 0 {
				usage.ImageCacheReadTokens = directUsage.ImageCacheReadTokens
			}
		}
		if upstreamErr := openAIImagesUpstreamErrorFromSSEPayload(data); upstreamErr != nil {
			if !clientDisconnected {
				downstreamWriteAttempted = true
				if writeErr := s.writeOpenAIImagesStreamEvent(c, flusher, "error", buildOpenAIImagesStreamErrorBodyFromUpstream(upstreamErr)); writeErr != nil {
					clientDisconnected = true
					clientWriteErr = writeErr
					logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Images stream client disconnected, continue draining upstream for billing")
				}
			}
			processErr = upstreamErr
			return
		}
		b64 := strings.TrimSpace(gjson.GetBytes(data, "b64_json").String())
		event := gjson.GetBytes(data, "type").String()
		if strings.HasPrefix(event, "image_generation.") && parsed.IsEdits() {
			event = strings.Replace(event, "image_generation.", "image_edit.", 1)
			data, _ = sjson.SetBytes(data, "type", event)
		}
		data, _ = sjson.SetBytes(data, "model", parsed.Model)
		if b64 != "" && strings.HasSuffix(event, ".completed") {
			count++
			if size := detectOpenAIImageResultSize(b64); size != "" {
				sizes = append(sizes, size)
				data, _ = sjson.SetBytes(data, "size", size)
			}
		}
		if parsed.ResponseFormat == "url" && b64 != "" {
			format := gjson.GetBytes(data, "output_format").String()
			if format == "" {
				format = parsed.OutputFormat
			}
			data, _ = sjson.SetBytes(data, "url", "data:"+openAIImageOutputMIMEType(format)+";base64,"+b64)
			data, _ = sjson.DeleteBytes(data, "b64_json")
		}
		if event != "" && !clientDisconnected {
			// A client write can partially succeed even when it returns an error.
			// From this point the response is irreversible and must not fail over.
			downstreamWriteAttempted = true
			if writeErr := s.writeOpenAIImagesStreamEvent(c, flusher, event, data); writeErr != nil {
				clientDisconnected = true
				clientWriteErr = writeErr
				logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Images stream client disconnected, continue draining upstream for billing")
			}
		}
	}
	downstreamOutputStarted := func() bool {
		return downstreamWriteAttempted || (c != nil && c.Writer != nil && c.Writer.Written())
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			accumulator.AddLine(strings.TrimRight(line, "\r\n"), process)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			if clientWriteErr != nil {
				return usage, count, sizes, first, &codexDirectImagesOutputWrittenError{err: clientWriteErr}
			}
			if processErr != nil {
				if downstreamOutputStarted() {
					return usage, count, sizes, first, &codexDirectImagesOutputWrittenError{err: processErr}
				}
				return usage, count, sizes, first, processErr
			}
			if downstreamOutputStarted() {
				return usage, count, sizes, first, &codexDirectImagesOutputWrittenError{err: err}
			}
			return usage, count, sizes, first, newCodexDirectImagesPreOutputReadFailover(err, c)
		}
	}
	accumulator.Flush(process)
	if clientWriteErr != nil {
		return usage, count, sizes, first, &codexDirectImagesOutputWrittenError{err: clientWriteErr}
	}
	if processErr != nil {
		if downstreamOutputStarted() {
			// The caller preserves the returned usage but must not bill a preview as
			// a completed image; count remains the actual completed-output count.
			return usage, count, sizes, first, &codexDirectImagesOutputWrittenError{err: processErr}
		}
		return usage, count, sizes, first, processErr
	}
	if count == 0 {
		if downstreamOutputStarted() {
			return usage, 0, sizes, first, &codexDirectImagesOutputWrittenError{err: io.ErrUnexpectedEOF}
		}
		return usage, 0, sizes, first, &OpenAIImagesUpstreamError{StatusCode: http.StatusBadGateway, ErrorType: "upstream_error", Message: "Images API returned no image output"}
	}
	return usage, count, sizes, first, nil
}

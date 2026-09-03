package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	openAIImagesGenerationsEndpoint = "/v1/images/generations"
	openAIImagesEditsEndpoint       = "/v1/images/edits"

	openAIImagesGenerationsURL = "https://api.openai.com/v1/images/generations"
	openAIImagesEditsURL       = "https://api.openai.com/v1/images/edits"

	apimartImagesGenerationsEndpoint = openAIImagesGenerationsEndpoint
	apimartImagesEditsEndpoint       = openAIImagesEditsEndpoint
	apimartMidjourneyEndpoint        = "/v1/midjourney/generations"
	apimartImagesUploadEndpoint      = "/v1/uploads/images"
	apimartImagesTaskEndpointPrefix  = "/v1/tasks/"
	apimartImagesPollInterval        = 3 * time.Second
	apimartImagesMaxPolls            = 80
	apimartImagesMaxResponseBytes    = 8 << 20
	apimartImagesMaxErrorBytes       = 2 << 20
	apimartImagesDefaultResolution   = "1k"

	openAIChatGPTStartURL                  = "https://chatgpt.com/"
	openAIChatGPTFilesURL                  = "https://chatgpt.com/backend-api/files"
	openAIImageBackendUserAgent            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	openAIImageMaxDownloadBytes            = 20 << 20 // 20MB per image download
	openAIImageMaxUploadPartSize           = 20 << 20 // 20MB per multipart upload part
	openAIImagesResponsesMainModel         = "gpt-5.4-mini"
	openAIImagesVerbatimPromptInstructions = "When invoking the image_generation tool, use the user's image prompt verbatim. Do not rewrite, expand, summarize, embellish, translate, normalize punctuation, or add or remove visual details or constraints. Preserve the original language, wording, capitalization, quotes, and punctuation exactly."

	openAIImageInputTransportExtraKey        = "image_input_transport"
	openAIImageInputTransportObjectURL       = "object_url"
	openAIImageInputUploadLimitBytesExtraKey = "image_upload_limit_bytes"
	openAIImageURLFieldsSupportedExtraKey    = "image_url_fields_supported"
	openAIImageInputObjectKeyPrefix          = "image-inputs"
	openAIImageInputObjectURLTTL             = 2 * time.Hour
)

type OpenAIImagesCapability string

const (
	OpenAIImagesCapabilityBasic  OpenAIImagesCapability = "images-basic"
	OpenAIImagesCapabilityNative OpenAIImagesCapability = "images-native"
)

type OpenAIImagesUpload struct {
	FieldName   string
	FileName    string
	ContentType string
	Data        []byte
	Width       int
	Height      int
}

type OpenAIImagesRequest struct {
	Endpoint           string
	ContentType        string
	Multipart          bool
	Model              string
	ExplicitModel      bool
	Prompt             string
	Stream             bool
	N                  int
	Size               string
	AspectRatio        string
	Resolution         string
	ExplicitSize       bool
	SizeTier           string
	ResponseFormat     string
	NSFWCheck          *bool
	Watermark          *bool
	SequentialMode     string
	SequentialMax      *int
	Quality            string
	Background         string
	OutputFormat       string
	Moderation         string
	InputFidelity      string
	Style              string
	OutputCompression  *int
	PartialImages      *int
	GoogleSearch       *bool
	GoogleImageSearch  *bool
	MidjourneyVersion  string
	MidjourneySpeed    string
	MidjourneyStylize  *int
	MidjourneyChaos    *int
	MidjourneyWeird    *int
	MidjourneyNiji     *bool
	MidjourneyRaw      *bool
	MidjourneyTile     *bool
	MidjourneyStop     *int
	HasMask            bool
	HasNativeOptions   bool
	RequiredCapability OpenAIImagesCapability
	InputImageURLs     []string
	MaskImageURL       string
	Uploads            []OpenAIImagesUpload
	MaskUpload         *OpenAIImagesUpload
	Body               []byte
	bodyHash           string
}

func (r *OpenAIImagesRequest) ModerationBody() []byte {
	if r == nil {
		return nil
	}
	payload := map[string]any{}
	if prompt := strings.TrimSpace(r.Prompt); prompt != "" {
		payload["prompt"] = prompt
	}
	images := r.moderationImages()
	if len(images) > 0 {
		payload["images"] = images
	}
	if len(payload) == 0 {
		return nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return body
}

func (r *OpenAIImagesRequest) moderationImages() []map[string]string {
	if r == nil {
		return nil
	}
	images := make([]map[string]string, 0, len(r.InputImageURLs)+len(r.Uploads)+1)
	for _, imageURL := range r.InputImageURLs {
		imageURL = strings.TrimSpace(imageURL)
		if imageURL != "" {
			images = append(images, map[string]string{"image_url": imageURL})
		}
	}
	for _, upload := range r.Uploads {
		if dataURL := upload.ModerationDataURL(); dataURL != "" {
			images = append(images, map[string]string{"image_url": dataURL})
		}
	}
	if maskURL := strings.TrimSpace(r.MaskImageURL); maskURL != "" {
		images = append(images, map[string]string{"image_url": maskURL})
	}
	if r.MaskUpload != nil {
		if dataURL := r.MaskUpload.ModerationDataURL(); dataURL != "" {
			images = append(images, map[string]string{"image_url": dataURL})
		}
	}
	return images
}

func (u OpenAIImagesUpload) ModerationDataURL() string {
	if len(u.Data) == 0 {
		return ""
	}
	contentType := strings.TrimSpace(u.ContentType)
	if contentType == "" {
		contentType = http.DetectContentType(u.Data)
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return ""
	}
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(u.Data))
}

func (r *OpenAIImagesRequest) IsEdits() bool {
	return r != nil && r.Endpoint == openAIImagesEditsEndpoint
}

func (r *OpenAIImagesRequest) StickySessionSeed() string {
	if r == nil {
		return ""
	}
	parts := []string{
		"openai-images",
		strings.TrimSpace(r.Endpoint),
		strings.TrimSpace(r.Model),
		strings.TrimSpace(r.Size),
		strings.TrimSpace(r.Prompt),
	}
	seed := strings.Join(parts, "|")
	if strings.TrimSpace(r.Prompt) == "" && r.bodyHash != "" {
		seed += "|body=" + r.bodyHash
	}
	return seed
}

func (s *OpenAIGatewayService) ParseOpenAIImagesRequest(c *gin.Context, body []byte) (*OpenAIImagesRequest, error) {
	if c == nil || c.Request == nil {
		return nil, fmt.Errorf("missing request context")
	}
	endpoint := normalizeOpenAIImagesEndpointPath(c.Request.URL.Path)
	if endpoint == "" {
		return nil, fmt.Errorf("unsupported images endpoint")
	}

	contentType := strings.TrimSpace(c.GetHeader("Content-Type"))
	req := &OpenAIImagesRequest{
		Endpoint:    endpoint,
		ContentType: contentType,
		N:           1,
		Body:        body,
	}
	if len(body) > 0 {
		sum := sha256.Sum256(body)
		req.bodyHash = hex.EncodeToString(sum[:8])
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err == nil && strings.EqualFold(mediaType, "multipart/form-data") {
		req.Multipart = true
		if parseErr := parseOpenAIImagesMultipartRequest(body, contentType, req); parseErr != nil {
			return nil, parseErr
		}
	} else {
		if len(body) == 0 {
			return nil, fmt.Errorf("request body is empty")
		}
		if !gjson.ValidBytes(body) {
			return nil, fmt.Errorf("failed to parse request body")
		}
		if parseErr := parseOpenAIImagesJSONRequest(body, req); parseErr != nil {
			return nil, parseErr
		}
	}

	applyOpenAIImagesDefaults(req)
	if err := validateOpenAIImagesModel(req.Model); err != nil {
		return nil, err
	}
	if err := validateOpenAIImagesReferenceLimit(req, req.Model); err != nil {
		return nil, err
	}
	if err := validateAPIMartGrokImagine20Request(req); err != nil {
		return nil, err
	}
	if err := validateAPIMartSeedream50Request(req); err != nil {
		return nil, err
	}
	req.SizeTier = normalizeOpenAIImageSizeTier(req.Size)
	req.RequiredCapability = classifyOpenAIImagesCapability(req)
	return req, nil
}

func parseOpenAIImagesJSONRequest(body []byte, req *OpenAIImagesRequest) error {
	if modelResult := gjson.GetBytes(body, "model"); modelResult.Exists() {
		req.Model = strings.TrimSpace(modelResult.String())
		req.ExplicitModel = req.Model != ""
	}
	req.Prompt = strings.TrimSpace(gjson.GetBytes(body, "prompt").String())

	if streamResult := gjson.GetBytes(body, "stream"); streamResult.Exists() {
		if streamResult.Type != gjson.True && streamResult.Type != gjson.False {
			return fmt.Errorf("invalid stream field type")
		}
		req.Stream = streamResult.Bool()
	}

	if nResult := gjson.GetBytes(body, "n"); nResult.Exists() {
		if nResult.Type != gjson.Number {
			return fmt.Errorf("invalid n field type")
		}
		req.N = int(nResult.Int())
		if req.N <= 0 {
			return fmt.Errorf("n must be greater than 0")
		}
	}

	if sizeResult := gjson.GetBytes(body, "size"); sizeResult.Exists() {
		req.Size = strings.TrimSpace(sizeResult.String())
		req.ExplicitSize = req.Size != ""
	}
	if aspectRatio := gjson.GetBytes(body, "aspect_ratio"); aspectRatio.Exists() {
		req.AspectRatio = strings.TrimSpace(aspectRatio.String())
	}
	if resolutionResult := gjson.GetBytes(body, "resolution"); resolutionResult.Exists() {
		req.Resolution = strings.TrimSpace(resolutionResult.String())
	}
	if req.Resolution == "" {
		if resolutionResult := gjson.GetBytes(body, "image_resolution"); resolutionResult.Exists() {
			req.Resolution = strings.TrimSpace(resolutionResult.String())
		}
	}
	req.ResponseFormat = strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "response_format").String()))
	if nsfw := gjson.GetBytes(body, "nsfw_check"); nsfw.Exists() {
		value, err := parseOpenAIImagesJSONBool(nsfw, "nsfw_check")
		if err != nil {
			return err
		}
		req.NSFWCheck = &value
	}
	if watermark := gjson.GetBytes(body, "watermark"); watermark.Exists() {
		value, err := parseOpenAIImagesJSONBool(watermark, "watermark")
		if err != nil {
			return err
		}
		req.Watermark = &value
	}
	req.SequentialMode = strings.TrimSpace(gjson.GetBytes(body, "sequential_image_generation").String())
	if maxImages := gjson.GetBytes(body, "sequential_image_generation_options.max_images"); maxImages.Exists() {
		if maxImages.Type != gjson.Number {
			return fmt.Errorf("invalid sequential_image_generation_options.max_images field type")
		}
		value := int(maxImages.Int())
		req.SequentialMax = &value
	}
	req.Quality = strings.TrimSpace(gjson.GetBytes(body, "quality").String())
	req.Background = strings.TrimSpace(gjson.GetBytes(body, "background").String())
	if req.Background == "" {
		background, ok, err := parseOpenAIImagesTransparentBackgroundJSON(gjson.GetBytes(body, "transparent_background"))
		if err != nil {
			return err
		}
		if ok {
			req.Background = background
		}
	}
	req.OutputFormat = strings.TrimSpace(gjson.GetBytes(body, "output_format").String())
	req.Moderation = strings.TrimSpace(gjson.GetBytes(body, "moderation").String())
	req.InputFidelity = strings.TrimSpace(gjson.GetBytes(body, "input_fidelity").String())
	req.Style = strings.TrimSpace(gjson.GetBytes(body, "style").String())
	req.HasMask = gjson.GetBytes(body, "mask").Exists()
	if outputCompression := gjson.GetBytes(body, "output_compression"); outputCompression.Exists() {
		if outputCompression.Type != gjson.Number {
			return fmt.Errorf("invalid output_compression field type")
		}
		v := int(outputCompression.Int())
		req.OutputCompression = &v
	}
	if partialImages := gjson.GetBytes(body, "partial_images"); partialImages.Exists() {
		if partialImages.Type != gjson.Number {
			return fmt.Errorf("invalid partial_images field type")
		}
		v := int(partialImages.Int())
		req.PartialImages = &v
	}
	if googleSearch := gjson.GetBytes(body, "google_search"); googleSearch.Exists() {
		value, err := parseOpenAIImagesJSONBool(googleSearch, "google_search")
		if err != nil {
			return err
		}
		req.GoogleSearch = &value
	}
	if googleImageSearch := gjson.GetBytes(body, "google_image_search"); googleImageSearch.Exists() {
		value, err := parseOpenAIImagesJSONBool(googleImageSearch, "google_image_search")
		if err != nil {
			return err
		}
		req.GoogleImageSearch = &value
	}
	req.MidjourneyVersion = strings.TrimSpace(gjson.GetBytes(body, "version").String())
	req.MidjourneySpeed = strings.TrimSpace(gjson.GetBytes(body, "speed").String())
	for _, field := range []struct {
		path string
		dest **int
	}{
		{path: "stylize", dest: &req.MidjourneyStylize},
		{path: "chaos", dest: &req.MidjourneyChaos},
		{path: "weird", dest: &req.MidjourneyWeird},
		{path: "stop", dest: &req.MidjourneyStop},
	} {
		if value := gjson.GetBytes(body, field.path); value.Exists() {
			if value.Type != gjson.Number {
				return fmt.Errorf("invalid %s field type", field.path)
			}
			v := int(value.Int())
			*field.dest = &v
		}
	}
	for _, field := range []struct {
		path string
		dest **bool
	}{
		{path: "niji", dest: &req.MidjourneyNiji},
		{path: "raw", dest: &req.MidjourneyRaw},
		{path: "tile", dest: &req.MidjourneyTile},
	} {
		if value := gjson.GetBytes(body, field.path); value.Exists() {
			parsed, err := parseOpenAIImagesJSONBool(value, field.path)
			if err != nil {
				return err
			}
			*field.dest = &parsed
		}
	}
	if err := appendOpenAIImagesJSONURLField(body, "image_urls", &req.InputImageURLs); err != nil {
		return err
	}
	if imageURL := strings.TrimSpace(gjson.GetBytes(body, "image_url").String()); imageURL != "" {
		req.InputImageURLs = append(req.InputImageURLs, imageURL)
	}
	if maskURL := strings.TrimSpace(gjson.GetBytes(body, "mask_url").String()); maskURL != "" {
		req.MaskImageURL = maskURL
		req.HasMask = true
	}
	if req.IsEdits() {
		images := gjson.GetBytes(body, "images")
		if images.Exists() {
			if !images.IsArray() {
				return fmt.Errorf("invalid images field type")
			}
			for _, item := range images.Array() {
				if imageURL := strings.TrimSpace(item.Get("image_url").String()); imageURL != "" {
					req.InputImageURLs = append(req.InputImageURLs, imageURL)
					continue
				}
				if item.Get("file_id").Exists() {
					return fmt.Errorf("images[].file_id is not supported (use images[].image_url instead)")
				}
			}
		}
		if maskImageURL := strings.TrimSpace(gjson.GetBytes(body, "mask.image_url").String()); maskImageURL != "" {
			req.MaskImageURL = maskImageURL
			req.HasMask = true
		}
		if gjson.GetBytes(body, "mask.file_id").Exists() {
			return fmt.Errorf("mask.file_id is not supported (use mask.image_url instead)")
		}
		if len(req.InputImageURLs) == 0 {
			return fmt.Errorf("images[].image_url is required")
		}
	}
	req.HasNativeOptions = hasOpenAINativeImageOptions(func(path string) bool {
		return gjson.GetBytes(body, path).Exists()
	})
	return nil
}

func parseOpenAIImagesTransparentBackgroundJSON(value gjson.Result) (string, bool, error) {
	if !value.Exists() {
		return "", false, nil
	}
	switch value.Type {
	case gjson.True, gjson.False:
		return openAIImagesBackgroundFromTransparent(value.Bool()), true, nil
	case gjson.String:
		parsed, err := strconv.ParseBool(strings.TrimSpace(value.String()))
		if err != nil {
			return "", false, fmt.Errorf("invalid transparent_background field type")
		}
		return openAIImagesBackgroundFromTransparent(parsed), true, nil
	default:
		return "", false, fmt.Errorf("invalid transparent_background field type")
	}
}

func openAIImagesBackgroundFromTransparent(transparent bool) string {
	if transparent {
		return "transparent"
	}
	return "opaque"
}

func parseOpenAIImagesJSONBool(value gjson.Result, field string) (bool, error) {
	switch value.Type {
	case gjson.True, gjson.False:
		return value.Bool(), nil
	default:
		return false, fmt.Errorf("invalid %s field type", field)
	}
}

func appendOpenAIImagesJSONURLField(body []byte, field string, out *[]string) error {
	value := gjson.GetBytes(body, field)
	if !value.Exists() {
		return nil
	}
	switch {
	case value.IsArray():
		for _, item := range value.Array() {
			if item.Type != gjson.String {
				return fmt.Errorf("invalid %s field type", field)
			}
			if url := strings.TrimSpace(item.String()); url != "" {
				*out = append(*out, url)
			}
		}
	case value.Type == gjson.String:
		if url := strings.TrimSpace(value.String()); url != "" {
			*out = append(*out, url)
		}
	default:
		return fmt.Errorf("invalid %s field type", field)
	}
	return nil
}

func parseOpenAIImagesMultipartRequest(body []byte, contentType string, req *OpenAIImagesRequest) error {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("invalid multipart content-type: %w", err)
	}
	boundary := strings.TrimSpace(params["boundary"])
	if boundary == "" {
		return fmt.Errorf("multipart boundary is required")
	}

	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read multipart body: %w", err)
		}
		name := strings.TrimSpace(part.FormName())
		if name == "" {
			_ = part.Close()
			continue
		}

		data, err := io.ReadAll(io.LimitReader(part, openAIImageMaxUploadPartSize))
		_ = part.Close()
		if err != nil {
			return fmt.Errorf("read multipart field %s: %w", name, err)
		}

		fileName := strings.TrimSpace(part.FileName())
		if fileName != "" {
			partContentType := strings.TrimSpace(part.Header.Get("Content-Type"))
			if name == "mask" && len(data) > 0 {
				req.HasMask = true
				width, height := parseOpenAIImageDimensions(part.Header)
				maskUpload := OpenAIImagesUpload{
					FieldName:   name,
					FileName:    fileName,
					ContentType: partContentType,
					Data:        data,
					Width:       width,
					Height:      height,
				}
				req.MaskUpload = &maskUpload
			}
			if name == "image" || strings.HasPrefix(name, "image[") {
				width, height := parseOpenAIImageDimensions(part.Header)
				req.Uploads = append(req.Uploads, OpenAIImagesUpload{
					FieldName:   name,
					FileName:    fileName,
					ContentType: partContentType,
					Data:        data,
					Width:       width,
					Height:      height,
				})
			}
			continue
		}

		value := strings.TrimSpace(string(data))
		switch name {
		case "model":
			req.Model = value
			req.ExplicitModel = value != ""
		case "prompt":
			req.Prompt = value
		case "size":
			req.Size = value
			req.ExplicitSize = value != ""
		case "resolution", "image_resolution":
			req.Resolution = value
		case "response_format":
			req.ResponseFormat = strings.ToLower(value)
		case "stream":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid stream field value")
			}
			req.Stream = parsed
		case "n":
			n, err := strconv.Atoi(value)
			if err != nil || n <= 0 {
				return fmt.Errorf("n must be a positive integer")
			}
			req.N = n
		case "quality":
			req.Quality = value
			req.HasNativeOptions = true
		case "background":
			req.Background = value
			req.HasNativeOptions = true
		case "transparent_background":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid transparent_background field value")
			}
			if req.Background == "" {
				req.Background = openAIImagesBackgroundFromTransparent(parsed)
			}
			req.HasNativeOptions = true
		case "output_format":
			req.OutputFormat = value
			req.HasNativeOptions = true
		case "moderation":
			req.Moderation = value
			req.HasNativeOptions = true
		case "input_fidelity":
			req.InputFidelity = value
			req.HasNativeOptions = true
		case "style":
			req.Style = value
			req.HasNativeOptions = true
		case "output_compression":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid output_compression field value")
			}
			req.OutputCompression = &n
			req.HasNativeOptions = true
		case "partial_images":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid partial_images field value")
			}
			req.PartialImages = &n
			req.HasNativeOptions = true
		case "google_search":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid google_search field value")
			}
			req.GoogleSearch = &parsed
			req.HasNativeOptions = true
		case "google_image_search":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid google_image_search field value")
			}
			req.GoogleImageSearch = &parsed
			req.HasNativeOptions = true
		case "image_url", "image_urls":
			req.InputImageURLs = append(req.InputImageURLs, parseOpenAIImagesURLFieldValue(value)...)
		case "mask_url":
			req.MaskImageURL = value
			req.HasMask = value != ""
		case "version":
			req.MidjourneyVersion = value
		case "speed":
			req.MidjourneySpeed = value
		case "stylize":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid stylize field value")
			}
			req.MidjourneyStylize = &n
		case "chaos":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid chaos field value")
			}
			req.MidjourneyChaos = &n
		case "weird":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid weird field value")
			}
			req.MidjourneyWeird = &n
		case "stop":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid stop field value")
			}
			req.MidjourneyStop = &n
		case "niji":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid niji field value")
			}
			req.MidjourneyNiji = &parsed
		case "raw":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid raw field value")
			}
			req.MidjourneyRaw = &parsed
		case "tile":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid tile field value")
			}
			req.MidjourneyTile = &parsed
		default:
			if isOpenAINativeImageOption(name) && value != "" {
				req.HasNativeOptions = true
			}
		}
	}

	if len(req.Uploads) == 0 && req.IsEdits() {
		return fmt.Errorf("image file is required")
	}
	return nil
}

func parseOpenAIImagesURLFieldValue(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if gjson.Valid(value) {
		parsed := gjson.Parse(value)
		if parsed.IsArray() {
			out := make([]string, 0, len(parsed.Array()))
			for _, item := range parsed.Array() {
				if url := strings.TrimSpace(item.String()); url != "" {
					out = append(out, url)
				}
			}
			return out
		}
	}
	return []string{value}
}

func parseOpenAIImageDimensions(_ textproto.MIMEHeader) (int, int) {
	return 0, 0
}

func applyOpenAIImagesDefaults(req *OpenAIImagesRequest) {
	if req == nil {
		return
	}
	if req.N <= 0 {
		req.N = 1
	}
	if strings.TrimSpace(req.Model) != "" {
		req.Model = strings.TrimSpace(req.Model)
		return
	}
	req.Model = "gpt-image-2"
}

func isOpenAIImageGenerationModel(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(normalized, "gpt-image-") ||
		isAPIMartGeminiImageModel(normalized) ||
		isAPIMartMidjourneyImageModel(normalized) ||
		isAPIMartGrokImagineImageModel(normalized) ||
		isAPIMartSeedreamImageModel(normalized)
}

func isAPIMartGeminiImageModel(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "gemini-3-pro-image-preview",
		"gemini-3-pro-image-preview-official",
		"gemini-3.1-flash-image-preview",
		"gemini-3.1-flash-image-preview-official":
		return true
	default:
		return false
	}
}

func isAPIMartGeminiFlashImageModel(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "gemini-3.1-flash-image-preview", "gemini-3.1-flash-image-preview-official":
		return true
	default:
		return false
	}
}

func isAPIMartMidjourneyImageModel(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "midjourney")
}

func isAPIMartGrokImagineImageModel(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "grok-imagine-1.5-apimart", "grok-imagine-1.5-edit-apimart",
		"grok-imagine-2.0-ext", "grok-imagine-image-2.0":
		return true
	default:
		return false
	}
}

func isAPIMartGrokImagine20ExtModel(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "grok-imagine-2.0-ext")
}

func isAPIMartGrokImagine20ImageModel(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "grok-imagine-image-2.0")
}

func isAPIMartSeedreamImageModel(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "doubao-seedance-4-0", "doubao-seedance-4-5",
		"seedream-5-0-pro", "seedream-5.0-pro",
		"seedream-5-0-lite", "seedream-5.0-lite":
		return true
	default:
		return false
	}
}

func isAPIMartSeedream50ProModel(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return normalized == "seedream-5-0-pro" || normalized == "seedream-5.0-pro"
}

func isAPIMartSeedream50LiteModel(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return normalized == "seedream-5-0-lite" || normalized == "seedream-5.0-lite"
}

func isAPIMartGrokImagineEditModel(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "grok-imagine-1.5-edit-apimart")
}

func validateOpenAIImagesModel(model string) error {
	model = strings.TrimSpace(model)
	if isOpenAIImageGenerationModel(model) {
		return nil
	}
	if model == "" {
		return fmt.Errorf("images endpoint requires an image model")
	}
	return fmt.Errorf("images endpoint requires an image model, got %q", model)
}

func validateOpenAIImagesReferenceLimit(req *OpenAIImagesRequest, model string) error {
	if req == nil {
		return nil
	}
	limit := 0
	switch {
	case isAPIMartGeminiImageModel(model):
		limit = 14
	case isAPIMartMidjourneyImageModel(model):
		limit = 4
	case isAPIMartGrokImagineImageModel(model):
		if isAPIMartGrokImagine20ExtModel(model) {
			limit = 0
		} else if isAPIMartGrokImagine20ImageModel(model) {
			limit = 3
		} else {
			limit = 1
		}
	case isAPIMartSeedreamImageModel(model):
		count := len(compactTrimmedStrings(req.InputImageURLs)) + len(req.Uploads)
		if count+maxInt(1, req.N) > 15 {
			return fmt.Errorf("%s supports at most 15 input and output images, got %d", strings.TrimSpace(model), count+maxInt(1, req.N))
		}
		return nil
	default:
		return nil
	}
	count := len(compactTrimmedStrings(req.InputImageURLs)) + len(req.Uploads)
	if count > limit {
		return fmt.Errorf("%s supports at most %d reference images, got %d", strings.TrimSpace(model), limit, count)
	}
	return nil
}

func validateAPIMartGrokImagine20Request(req *OpenAIImagesRequest) error {
	if req == nil {
		return nil
	}
	model := strings.TrimSpace(req.Model)
	if isAPIMartGrokImagine20ExtModel(model) {
		if req.N > 12 {
			return fmt.Errorf("%s supports at most 12 output images, got %d", model, req.N)
		}
		if len(compactTrimmedStrings(req.InputImageURLs))+len(req.Uploads) > 0 {
			return fmt.Errorf("%s does not support reference images", model)
		}
		if req.ResponseFormat != "" && !strings.EqualFold(req.ResponseFormat, "url") {
			return fmt.Errorf("%s only supports response_format=url", model)
		}
	}
	if isAPIMartGrokImagine20ImageModel(model) && req.N > 10 {
		return fmt.Errorf("%s supports at most 10 output images, got %d", model, req.N)
	}
	return nil
}

func validateAPIMartSeedream50Request(req *OpenAIImagesRequest) error {
	if req == nil {
		return nil
	}
	model := strings.TrimSpace(req.Model)
	inputCount := len(compactTrimmedStrings(req.InputImageURLs)) + len(req.Uploads)
	if isAPIMartSeedream50ProModel(model) {
		if req.N != 1 {
			return fmt.Errorf("%s supports exactly 1 output image, got %d", model, req.N)
		}
		if inputCount > 10 {
			return fmt.Errorf("%s supports at most 10 reference images, got %d", model, inputCount)
		}
		if req.Stream {
			return fmt.Errorf("%s does not support streaming", model)
		}
	}
	if isAPIMartSeedream50LiteModel(model) && inputCount+maxInt(1, req.N) > 15 {
		return fmt.Errorf("%s supports at most 15 input and output images, got %d", model, inputCount+maxInt(1, req.N))
	}
	return nil
}

type preparedOpenAIImageURLInputs struct {
	ImageURLs []string
	MaskURL   string
	keys      []string
}

func (p preparedOpenAIImageURLInputs) cleanup(service *OpenAIGatewayService) {
	if service == nil || len(p.keys) == 0 {
		return
	}
	service.cleanupOpenAIImageInputObjects(p.keys)
}

func shouldUseOpenAIImagesObjectURLTransport(account *Account, parsed *OpenAIImagesRequest) bool {
	if account == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(account.GetExtraString(openAIImageInputTransportExtraKey)), openAIImageInputTransportObjectURL) {
		return true
	}
	limit := openAIImagesAccountUploadLimitBytes(account)
	return limit > 0 && openAIImagesLocalInputExceedsLimit(parsed, limit)
}

func openAIImagesAccountUploadLimitBytes(account *Account) int64 {
	if account == nil || account.Extra == nil {
		return 0
	}
	value := ParseExtraInt(account.Extra[openAIImageInputUploadLimitBytesExtraKey])
	if value <= 0 {
		return 0
	}
	return int64(value)
}

func openAIImagesLocalInputExceedsLimit(parsed *OpenAIImagesRequest, limit int64) bool {
	if parsed == nil || limit <= 0 {
		return false
	}
	for _, upload := range parsed.Uploads {
		if int64(len(upload.Data)) > limit {
			return true
		}
	}
	if parsed.MaskUpload != nil && int64(len(parsed.MaskUpload.Data)) > limit {
		return true
	}
	for _, raw := range parsed.InputImageURLs {
		data, _, ok, err := parseOpenAIImagesDataURL(raw)
		if err == nil && ok && int64(len(data)) > limit {
			return true
		}
	}
	if data, _, ok, err := parseOpenAIImagesDataURL(parsed.MaskImageURL); err == nil && ok && int64(len(data)) > limit {
		return true
	}
	return false
}

func accountSupportsOpenAIImageURLFields(account *Account) bool {
	if account == nil || account.Extra == nil {
		return false
	}
	return parseOpenAIImagesExtraBool(account.Extra[openAIImageURLFieldsSupportedExtraKey])
}

func parseOpenAIImagesExtraBool(raw any) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "y", "on", "enabled":
			return true
		default:
			return false
		}
	case int:
		return value != 0
	case int64:
		return value != 0
	case float64:
		return value != 0
	case json.Number:
		i, err := value.Int64()
		return err == nil && i != 0
	default:
		return false
	}
}

func parseOpenAIImagesDataURL(raw string) ([]byte, string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(strings.ToLower(raw), "data:") {
		return nil, "", false, nil
	}
	comma := strings.Index(raw, ",")
	if comma < 0 {
		return nil, "", true, fmt.Errorf("invalid image data URL")
	}
	header := raw[len("data:"):comma]
	if !strings.Contains(strings.ToLower(header), ";base64") {
		return nil, "", true, fmt.Errorf("image data URL must be base64")
	}
	contentType := strings.TrimSpace(strings.Split(header, ";")[0])
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return nil, "", true, fmt.Errorf("image data URL must be image/*")
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw[comma+1:]))
	if err != nil {
		return nil, "", true, fmt.Errorf("decode image data URL: %w", err)
	}
	if len(data) == 0 {
		return nil, "", true, fmt.Errorf("image data URL is empty")
	}
	return data, contentType, true, nil
}

func (s *OpenAIGatewayService) prepareOpenAIImagesObjectURLInputs(
	ctx context.Context,
	account *Account,
	parsed *OpenAIImagesRequest,
) (preparedOpenAIImageURLInputs, error) {
	if parsed == nil {
		return preparedOpenAIImageURLInputs{}, fmt.Errorf("parsed images request is required")
	}
	prepared := preparedOpenAIImageURLInputs{
		ImageURLs: make([]string, 0, len(parsed.InputImageURLs)+len(parsed.Uploads)),
	}
	for _, imageURL := range parsed.InputImageURLs {
		converted, key, err := s.openAIImageInputURLAsObjectURL(ctx, account, imageURL)
		if err != nil {
			prepared.cleanup(s)
			return preparedOpenAIImageURLInputs{}, err
		}
		if strings.TrimSpace(converted) != "" {
			prepared.ImageURLs = append(prepared.ImageURLs, converted)
		}
		if key != "" {
			prepared.keys = append(prepared.keys, key)
		}
	}
	for _, upload := range parsed.Uploads {
		converted, key, err := s.uploadOpenAIImageInputObject(ctx, account, upload.Data, upload.ContentType, upload.FileName)
		if err != nil {
			prepared.cleanup(s)
			return preparedOpenAIImageURLInputs{}, err
		}
		prepared.ImageURLs = append(prepared.ImageURLs, converted)
		prepared.keys = append(prepared.keys, key)
	}

	maskURL := strings.TrimSpace(parsed.MaskImageURL)
	if maskURL != "" {
		converted, key, err := s.openAIImageInputURLAsObjectURL(ctx, account, maskURL)
		if err != nil {
			prepared.cleanup(s)
			return preparedOpenAIImageURLInputs{}, err
		}
		maskURL = converted
		if key != "" {
			prepared.keys = append(prepared.keys, key)
		}
	}
	if parsed.MaskUpload != nil {
		converted, key, err := s.uploadOpenAIImageInputObject(ctx, account, parsed.MaskUpload.Data, parsed.MaskUpload.ContentType, parsed.MaskUpload.FileName)
		if err != nil {
			prepared.cleanup(s)
			return preparedOpenAIImageURLInputs{}, err
		}
		maskURL = converted
		prepared.keys = append(prepared.keys, key)
	}
	prepared.MaskURL = maskURL
	return prepared, nil
}

func (s *OpenAIGatewayService) openAIImageInputURLAsObjectURL(
	ctx context.Context,
	account *Account,
	raw string,
) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", nil
	}
	data, contentType, ok, err := parseOpenAIImagesDataURL(raw)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return raw, "", nil
	}
	return s.uploadOpenAIImageInputObject(ctx, account, data, contentType, "image")
}

func (s *OpenAIGatewayService) uploadOpenAIImageInputObject(
	ctx context.Context,
	account *Account,
	data []byte,
	contentType string,
	filename string,
) (string, string, error) {
	if len(data) == 0 {
		return "", "", fmt.Errorf("image input is empty")
	}
	store, err := s.getOpenAIImageInputObjectStore(ctx)
	if err != nil {
		return "", "", err
	}
	contentType = normalizeImageMimeType(contentType, data)
	key := s.openAIImageInputObjectKey(account, data, contentType, filename)
	if _, err := store.Upload(ctx, key, bytes.NewReader(data), contentType); err != nil {
		return "", "", fmt.Errorf("upload image input object: %w", err)
	}
	objectURL, err := store.PresignURL(ctx, key, openAIImageInputObjectURLTTL)
	if err != nil {
		_ = store.Delete(context.Background(), key)
		return "", "", fmt.Errorf("presign image input object: %w", err)
	}
	if strings.TrimSpace(objectURL) == "" {
		_ = store.Delete(context.Background(), key)
		return "", "", fmt.Errorf("presign image input object returned empty url")
	}
	return strings.TrimSpace(objectURL), key, nil
}

func (s *OpenAIGatewayService) openAIImageInputObjectKey(account *Account, data []byte, contentType string, filename string) string {
	prefix := openAIImageInputObjectKeyPrefix
	if s != nil && s.cfg != nil {
		if configuredPrefix := strings.Trim(strings.TrimSpace(s.cfg.ImageCreator.ObjectStorage.Prefix), "/"); configuredPrefix != "" {
			prefix = configuredPrefix + "/" + openAIImageInputObjectKeyPrefix
		}
	}
	accountID := int64(0)
	if account != nil {
		accountID = account.ID
	}
	sum := sha256.Sum256(data)
	ext := imageExtension(contentType, filename, "png")
	return fmt.Sprintf("%s/%d/%s/%s-%s.%s", prefix, accountID, time.Now().UTC().Format("2006/01/02"), hex.EncodeToString(sum[:]), uuid.NewString(), ext)
}

func (s *OpenAIGatewayService) getOpenAIImageInputObjectStore(ctx context.Context) (BackupObjectStore, error) {
	if s == nil {
		return nil, fmt.Errorf("openai gateway service is unavailable")
	}
	if s.imageInputObjectStoreFactory == nil {
		return nil, fmt.Errorf("image input object storage factory is unavailable")
	}
	cfg := openAIImageInputObjectStorageConfig(s.cfg)
	if cfg == nil || !cfg.IsConfigured() {
		return nil, fmt.Errorf("image input object storage is not configured")
	}
	s.imageInputObjectStoreMu.Lock()
	defer s.imageInputObjectStoreMu.Unlock()
	if s.imageInputObjectStore != nil {
		return s.imageInputObjectStore, nil
	}
	store, err := s.imageInputObjectStoreFactory(ctx, cfg)
	if err != nil {
		return nil, err
	}
	s.imageInputObjectStore = store
	return store, nil
}

func openAIImageInputObjectStorageConfig(cfg *config.Config) *BackupS3Config {
	opts := imageCreatorOptionsFromConfig(cfg)
	if opts.ObjectStorage == nil || !opts.ObjectStorage.IsConfigured() {
		return nil
	}
	copied := *opts.ObjectStorage
	return &copied
}

func (s *OpenAIGatewayService) cleanupOpenAIImageInputObjects(keys []string) {
	keys = dedupeStrings(compactTrimmedStrings(keys))
	if len(keys) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store, err := s.getOpenAIImageInputObjectStore(ctx)
	if err != nil {
		logger.LegacyPrintf("service.openai_gateway", "[OpenAI] cleanup image input objects skipped: %s", sanitizeUpstreamErrorMessage(err.Error()))
		return
	}
	for _, key := range keys {
		if err := store.Delete(ctx, key); err != nil {
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI] cleanup image input object failed key=%s err=%s", key, sanitizeUpstreamErrorMessage(err.Error()))
		}
	}
}

func normalizeOpenAIImagesEndpointPath(path string) string {
	trimmed := strings.TrimSpace(path)
	switch {
	case strings.Contains(trimmed, "/images/generations"):
		return openAIImagesGenerationsEndpoint
	case strings.Contains(trimmed, "/midjourney/generations"):
		return openAIImagesGenerationsEndpoint
	case strings.Contains(trimmed, "/images/edits"):
		return openAIImagesEditsEndpoint
	default:
		return ""
	}
}

func classifyOpenAIImagesCapability(req *OpenAIImagesRequest) OpenAIImagesCapability {
	if req == nil {
		return OpenAIImagesCapabilityNative
	}
	if req.ExplicitModel || req.ExplicitSize {
		return OpenAIImagesCapabilityNative
	}
	model := strings.ToLower(strings.TrimSpace(req.Model))
	if !strings.HasPrefix(model, "gpt-image-") {
		return OpenAIImagesCapabilityNative
	}
	if req.Stream || req.N != 1 || req.HasMask || req.HasNativeOptions {
		return OpenAIImagesCapabilityNative
	}
	if req.IsEdits() && !req.Multipart {
		return OpenAIImagesCapabilityNative
	}
	if req.ResponseFormat != "" && req.ResponseFormat != "b64_json" {
		return OpenAIImagesCapabilityNative
	}
	return OpenAIImagesCapabilityBasic
}

func hasOpenAINativeImageOptions(exists func(path string) bool) bool {
	for _, path := range []string{
		"background",
		"transparent_background",
		"quality",
		"style",
		"output_format",
		"output_compression",
		"moderation",
		"input_fidelity",
		"partial_images",
		"google_search",
		"google_image_search",
	} {
		if exists(path) {
			return true
		}
	}
	return false
}

func isOpenAINativeImageOption(name string) bool {
	switch strings.TrimSpace(strings.ToLower(name)) {
	case "background", "transparent_background", "quality", "style", "output_format", "output_compression", "moderation", "input_fidelity", "partial_images", "google_search", "google_image_search":
		return true
	default:
		return false
	}
}

func normalizeOpenAIImageSizeTier(size string) string {
	return NormalizeImageBillingTierOrDefault(size)
}

func (s *OpenAIGatewayService) ForwardImages(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	parsed *OpenAIImagesRequest,
	channelMappedModel string,
) (*OpenAIForwardResult, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed images request is required")
	}
	switch account.Type {
	case AccountTypeAPIKey:
		return s.forwardOpenAIImagesAPIKey(ctx, c, account, body, parsed, channelMappedModel)
	case AccountTypeOAuth:
		return s.forwardOpenAIImagesOAuth(ctx, c, account, parsed, channelMappedModel)
	default:
		return nil, fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func (s *OpenAIGatewayService) forwardOpenAIImagesAPIKey(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	parsed *OpenAIImagesRequest,
	channelMappedModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()
	requestModel := strings.TrimSpace(parsed.Model)
	if mapped := strings.TrimSpace(channelMappedModel); mapped != "" {
		requestModel = mapped
	}
	if err := validateOpenAIImagesModel(requestModel); err != nil {
		return nil, err
	}
	upstreamModel := account.GetMappedModel(requestModel)
	if err := validateOpenAIImagesModel(upstreamModel); err != nil {
		return nil, err
	}
	if err := validateOpenAIImagesReferenceLimit(parsed, upstreamModel); err != nil {
		return nil, err
	}
	logger.LegacyPrintf(
		"service.openai_gateway",
		"[OpenAI] Images request routing request_model=%s upstream_model=%s endpoint=%s account_type=%s",
		strings.TrimSpace(parsed.Model),
		upstreamModel,
		parsed.Endpoint,
		account.Type,
	)
	upstreamCtx, releaseUpstreamCtx := detachUpstreamContext(ctx)
	defer releaseUpstreamCtx()

	token, _, err := s.GetAccessToken(upstreamCtx, account)
	if err != nil {
		return nil, err
	}
	if isAPIMartImagesHost(account) && isAPIMartImagesAsyncModel(upstreamModel) {
		return s.forwardAPIMartImages(upstreamCtx, c, account, parsed, token, requestModel, upstreamModel, startTime)
	}

	forwardBody, forwardContentType, err := rewriteOpenAIImagesModel(body, parsed.ContentType, upstreamModel)
	if err != nil {
		return nil, err
	}
	forwardParsed := parsed
	var preparedInputs preparedOpenAIImageURLInputs
	if accountSupportsOpenAIImageURLFields(account) && shouldUseOpenAIImagesObjectURLTransport(account, parsed) {
		preparedInputs, err = s.prepareOpenAIImagesObjectURLInputs(upstreamCtx, account, parsed)
		if err != nil {
			return nil, err
		}
		defer preparedInputs.cleanup(s)
		forwardBody, err = buildOpenAIImagesURLFieldsPayload(parsed, upstreamModel, preparedInputs.ImageURLs, preparedInputs.MaskURL)
		if err != nil {
			return nil, err
		}
		cloned := *parsed
		cloned.ContentType = "application/json"
		cloned.Multipart = false
		cloned.InputImageURLs = append([]string(nil), preparedInputs.ImageURLs...)
		cloned.Uploads = nil
		cloned.MaskImageURL = preparedInputs.MaskURL
		cloned.MaskUpload = nil
		forwardParsed = &cloned
		forwardContentType = "application/json"
	}
	if !forwardParsed.Multipart {
		setOpsUpstreamRequestBody(c, forwardBody)
	}

	if shouldSplitOpenAIImagesRequests(forwardParsed, upstreamModel) {
		return s.forwardSplitOpenAIImagesAPIKey(upstreamCtx, c, account, forwardBody, forwardParsed, token, requestModel, upstreamModel, startTime)
	}
	upstreamReq, err := s.buildOpenAIImagesRequest(upstreamCtx, c, account, forwardBody, forwardContentType, token, forwardParsed.Endpoint)
	if err != nil {
		return nil, err
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	upstreamStart := time.Now()
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			ProxyID:            opsUpstreamProxyID(account),
			ProxyName:          opsUpstreamProxyName(account),
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
			Kind:               "request_error",
			Message:            safeErr,
		})
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		if s.shouldFailoverOpenAIUpstreamResponse(resp.StatusCode, upstreamMsg, respBody) {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				ProxyID:            opsUpstreamProxyID(account),
				ProxyName:          opsUpstreamProxyName(account),
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
				Kind:               "failover",
				Message:            upstreamMsg,
			})
			s.handleFailoverSideEffects(upstreamCtx, resp, account, respBody)
			retryLimit, retryBackoffBase := openAISameAccountRetryPolicy(upstreamMsg, respBody)
			failoverErr := &UpstreamFailoverError{
				StatusCode:                  resp.StatusCode,
				ResponseBody:                respBody,
				RetryableOnSameAccount:      retryLimit > 0 || (account.IsPoolMode() && account.IsPoolModeRetryableStatus(resp.StatusCode)),
				SameAccountRetryLimit:       retryLimit,
				SameAccountRetryBackoffBase: retryBackoffBase,
			}
			return nil, s.applyOpenAIOAuth429Retry(account, resp.StatusCode, false, resp.Header, respBody, failoverErr)
		}
		return s.handleOpenAIImagesErrorResponse(upstreamCtx, resp, c, account, upstreamModel)
	}
	defer func() { _ = resp.Body.Close() }()

	var usage OpenAIUsage
	imageCount := parsed.N
	var firstTokenMs *int
	if parsed.Stream && isEventStreamResponse(resp.Header) {
		streamUsage, streamCount, streamSizes, ttft, err := s.handleOpenAIImagesStreamingResponse(resp, c, startTime)
		if err != nil {
			if streamCount > 0 {
				return &OpenAIForwardResult{
					RequestID:        resp.Header.Get("x-request-id"),
					Usage:            streamUsage,
					Model:            requestModel,
					UpstreamModel:    upstreamModel,
					Stream:           parsed.Stream,
					ResponseHeaders:  resp.Header.Clone(),
					Duration:         time.Since(startTime),
					FirstTokenMs:     ttft,
					ImageCount:       streamCount,
					ImageSize:        parsed.SizeTier,
					ImageQuality:     NormalizeImageQuality(parsed.Quality),
					ImageInputSize:   parsed.Size,
					ImageOutputSizes: streamSizes,
				}, err
			}
			return nil, err
		}
		usage = streamUsage
		imageCount = streamCount
		imageOutputSizes := streamSizes
		firstTokenMs = ttft
		return &OpenAIForwardResult{
			RequestID:        resp.Header.Get("x-request-id"),
			Usage:            usage,
			Model:            requestModel,
			UpstreamModel:    upstreamModel,
			Stream:           parsed.Stream,
			ResponseHeaders:  resp.Header.Clone(),
			Duration:         time.Since(startTime),
			FirstTokenMs:     firstTokenMs,
			ImageCount:       imageCount,
			ImageSize:        parsed.SizeTier,
			ImageQuality:     NormalizeImageQuality(parsed.Quality),
			ImageInputSize:   parsed.Size,
			ImageOutputSizes: imageOutputSizes,
		}, nil
	} else {
		nonStreamUsage, nonStreamCount, nonStreamSizes, err := s.handleOpenAIImagesNonStreamingResponse(resp, c)
		if err != nil {
			return nil, err
		}
		usage = nonStreamUsage
		if nonStreamCount > 0 {
			imageCount = nonStreamCount
		}
		return &OpenAIForwardResult{
			RequestID:        resp.Header.Get("x-request-id"),
			Usage:            usage,
			Model:            requestModel,
			UpstreamModel:    upstreamModel,
			Stream:           parsed.Stream,
			ResponseHeaders:  resp.Header.Clone(),
			Duration:         time.Since(startTime),
			FirstTokenMs:     firstTokenMs,
			ImageCount:       imageCount,
			ImageSize:        parsed.SizeTier,
			ImageQuality:     NormalizeImageQuality(parsed.Quality),
			ImageInputSize:   parsed.Size,
			ImageOutputSizes: nonStreamSizes,
		}, nil
	}
}

func (s *OpenAIGatewayService) forwardSplitOpenAIImagesAPIKey(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	parsed *OpenAIImagesRequest,
	token string,
	requestModel string,
	upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	if parsed.Stream {
		return s.forwardSplitOpenAIImagesAPIKeyStreaming(ctx, c, account, body, parsed, token, requestModel, upstreamModel, startTime)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	submitCount := splitOpenAIImagesRequestCount(parsed, upstreamModel)
	responseBodies := make([][]byte, 0, submitCount)
	var responseHeader http.Header
	statusCode := http.StatusOK
	requestID := ""
	var usage OpenAIUsage
	imageOutputSizes := make([]string, 0, submitCount)
	partialResult := func() *OpenAIForwardResult {
		imageCount := 0
		if len(responseBodies) > 0 {
			if _, count, err := buildSplitOpenAIImagesAPIResponse(responseBodies); err == nil {
				imageCount = count
			}
		}
		return &OpenAIForwardResult{
			RequestID:        requestID,
			Usage:            usage,
			Model:            requestModel,
			UpstreamModel:    upstreamModel,
			Stream:           false,
			ResponseHeaders:  responseHeader,
			Duration:         time.Since(startTime),
			ImageCount:       imageCount,
			ImageSize:        parsed.SizeTier,
			ImageQuality:     NormalizeImageQuality(parsed.Quality),
			ImageInputSize:   parsed.Size,
			ImageOutputSizes: imageOutputSizes,
		}
	}

	for i := 0; i < submitCount; i++ {
		forwardBody, forwardContentType, err := rewriteOpenAIImagesModelAndN(body, parsed.ContentType, upstreamModel, 1, true)
		if err != nil {
			return nil, err
		}
		setOpsUpstreamRequestBody(c, forwardBody)
		upstreamReq, err := s.buildOpenAIImagesRequest(ctx, c, account, forwardBody, forwardContentType, token, parsed.Endpoint)
		if err != nil {
			return nil, err
		}

		upstreamStart := time.Now()
		resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
		SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
		if err != nil {
			safeErr := sanitizeUpstreamErrorMessage(err.Error())
			setOpsUpstreamError(c, 0, safeErr, "")
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				ProxyID:            opsUpstreamProxyID(account),
				ProxyName:          opsUpstreamProxyName(account),
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: 0,
				UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
				Kind:               "request_error",
				Message:            safeErr,
			})
			if len(responseBodies) > 0 {
				return partialResult(), fmt.Errorf("upstream request failed: %s", safeErr)
			}
			return nil, fmt.Errorf("upstream request failed: %s", safeErr)
		}
		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
			_ = resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(respBody))
			upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
			upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
			if s.shouldFailoverOpenAIUpstreamResponse(resp.StatusCode, upstreamMsg, respBody) {
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					ProxyID:            opsUpstreamProxyID(account),
					ProxyName:          opsUpstreamProxyName(account),
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: resp.StatusCode,
					UpstreamRequestID:  resp.Header.Get("x-request-id"),
					UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
					Kind:               "failover",
					Message:            upstreamMsg,
				})
				s.handleFailoverSideEffects(ctx, resp, account, respBody)
				retryLimit, retryBackoffBase := openAISameAccountRetryPolicy(upstreamMsg, respBody)
				if len(responseBodies) > 0 {
					failoverErr := &UpstreamFailoverError{
						StatusCode:                  resp.StatusCode,
						ResponseBody:                respBody,
						RetryableOnSameAccount:      retryLimit > 0 || (account.IsPoolMode() && account.IsPoolModeRetryableStatus(resp.StatusCode)),
						SameAccountRetryLimit:       retryLimit,
						SameAccountRetryBackoffBase: retryBackoffBase,
					}
					return partialResult(), s.applyOpenAIOAuth429Retry(account, resp.StatusCode, false, resp.Header, respBody, failoverErr)
				}
				failoverErr := &UpstreamFailoverError{
					StatusCode:                  resp.StatusCode,
					ResponseBody:                respBody,
					RetryableOnSameAccount:      retryLimit > 0 || (account.IsPoolMode() && account.IsPoolModeRetryableStatus(resp.StatusCode)),
					SameAccountRetryLimit:       retryLimit,
					SameAccountRetryBackoffBase: retryBackoffBase,
				}
				return nil, s.applyOpenAIOAuth429Retry(account, resp.StatusCode, false, resp.Header, respBody, failoverErr)
			}
			result, err := s.handleOpenAIImagesErrorResponse(ctx, resp, c, account, upstreamModel)
			if len(responseBodies) > 0 {
				return partialResult(), err
			}
			return result, err
		}

		respBody, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
		_ = resp.Body.Close()
		if err != nil {
			if len(responseBodies) > 0 {
				return partialResult(), err
			}
			return nil, err
		}
		if responseHeader == nil {
			responseHeader = resp.Header.Clone()
			statusCode = resp.StatusCode
			requestID = resp.Header.Get("x-request-id")
		}
		responseBodies = append(responseBodies, respBody)
		if parsedUsage, ok := extractOpenAIUsageFromJSONBytes(respBody); ok {
			addOpenAIUsage(&usage, parsedUsage)
		}
		imageOutputSizes = append(imageOutputSizes, collectOpenAIResponseImageOutputSizesFromJSONBytes(respBody)...)
	}

	responseBody, imageCount, err := buildSplitOpenAIImagesAPIResponse(responseBodies)
	if err != nil {
		return nil, err
	}
	if responseHeader != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), responseHeader, s.responseHeaderFilter)
	}
	c.Data(statusCode, "application/json", responseBody)

	return &OpenAIForwardResult{
		RequestID:        requestID,
		Usage:            usage,
		Model:            requestModel,
		UpstreamModel:    upstreamModel,
		Stream:           false,
		ResponseHeaders:  responseHeader,
		Duration:         time.Since(startTime),
		ImageCount:       imageCount,
		ImageSize:        parsed.SizeTier,
		ImageQuality:     NormalizeImageQuality(parsed.Quality),
		ImageInputSize:   parsed.Size,
		ImageOutputSizes: imageOutputSizes,
	}, nil
}

func (s *OpenAIGatewayService) forwardSplitOpenAIImagesAPIKeyStreaming(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	parsed *OpenAIImagesRequest,
	token string,
	requestModel string,
	upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	// Keep streaming support conservative: each upstream call is still a single-image request,
	// and downstream receives the upstream stream for each image in order.
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	submitCount := splitOpenAIImagesRequestCount(parsed, upstreamModel)
	var usage OpenAIUsage
	imageCount := 0
	var imageOutputSizes []string
	var responseHeader http.Header
	requestID := ""
	var firstTokenMs *int
	streamStarted := false
	jsonBodies := make([][]byte, 0, submitCount)

	for i := 0; i < submitCount; i++ {
		forwardBody, forwardContentType, err := rewriteOpenAIImagesModelAndN(body, parsed.ContentType, upstreamModel, 1, true)
		if err != nil {
			return nil, err
		}
		setOpsUpstreamRequestBody(c, forwardBody)
		upstreamReq, err := s.buildOpenAIImagesRequest(ctx, c, account, forwardBody, forwardContentType, token, parsed.Endpoint)
		if err != nil {
			return nil, err
		}
		upstreamStart := time.Now()
		resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
		SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
		if err != nil {
			safeErr := sanitizeUpstreamErrorMessage(err.Error())
			setOpsUpstreamError(c, 0, safeErr, "")
			return nil, fmt.Errorf("upstream request failed: %s", safeErr)
		}
		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
			_ = resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(respBody))
			return s.handleOpenAIImagesErrorResponse(ctx, resp, c, account, upstreamModel)
		}
		if responseHeader == nil {
			responseHeader = resp.Header.Clone()
			requestID = resp.Header.Get("x-request-id")
		}
		if !isEventStreamResponse(resp.Header) {
			if streamStarted {
				_ = resp.Body.Close()
				return nil, fmt.Errorf("split image stream received non-stream response after streaming started")
			}
			respBody, readErr := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
			_ = resp.Body.Close()
			if readErr != nil {
				return nil, readErr
			}
			jsonBodies = append(jsonBodies, respBody)
			if parsedUsage, ok := extractOpenAIUsageFromJSONBytes(respBody); ok {
				addOpenAIUsage(&usage, parsedUsage)
			}
			imageOutputSizes = append(imageOutputSizes, collectOpenAIResponseImageOutputSizesFromJSONBytes(respBody)...)
			continue
		}
		streamStarted = true
		streamUsage, streamCount, streamSizes, ttft, err := s.handleOpenAIImagesSplitStreamingResponse(resp, c, startTime, i == submitCount-1)
		_ = resp.Body.Close()
		addOpenAIUsage(&usage, streamUsage)
		if streamCount > 0 {
			imageCount += streamCount
		}
		imageOutputSizes = append(imageOutputSizes, streamSizes...)
		if firstTokenMs == nil {
			firstTokenMs = ttft
		}
		if err != nil {
			return nil, err
		}
	}
	if !streamStarted {
		responseBody, count, err := buildSplitOpenAIImagesAPIResponse(jsonBodies)
		if err != nil {
			return nil, err
		}
		if responseHeader != nil {
			responseheaders.WriteFilteredHeaders(c.Writer.Header(), responseHeader, s.responseHeaderFilter)
		}
		c.Data(http.StatusOK, "application/json", responseBody)
		imageCount = count
	}
	if imageCount <= 0 {
		imageCount = submitCount
	}
	return &OpenAIForwardResult{
		RequestID:        requestID,
		Usage:            usage,
		Model:            requestModel,
		UpstreamModel:    upstreamModel,
		Stream:           true,
		ResponseHeaders:  responseHeader,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ImageCount:       imageCount,
		ImageSize:        parsed.SizeTier,
		ImageQuality:     NormalizeImageQuality(parsed.Quality),
		ImageInputSize:   parsed.Size,
		ImageOutputSizes: imageOutputSizes,
	}, nil
}

func (s *OpenAIGatewayService) handleOpenAIImagesSplitStreamingResponse(
	resp *http.Response,
	c *gin.Context,
	startTime time.Time,
	emitDone bool,
) (OpenAIUsage, int, []string, *int, error) {
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "text/event-stream"
	}
	c.Status(resp.StatusCode)
	c.Header("Content-Type", contentType)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return OpenAIUsage{}, 0, nil, nil, fmt.Errorf("streaming is not supported by response writer")
	}

	usage := OpenAIUsage{}
	imageCounter := newOpenAIImageOutputCounter()
	var firstTokenMs *int
	clientDisconnected := false
	var sseData openAISSEDataAccumulator
	suppressEvent := false

	processSSEData := func(dataBytes []byte) {
		mergeOpenAIUsage(&usage, dataBytes)
		imageCounter.AddSSEData(dataBytes)
	}
	flushSSEEvent := func() {
		sseData.Flush(processSSEData)
	}
	writeLine := func(line []byte) {
		if clientDisconnected || len(line) == 0 {
			return
		}
		if _, err := c.Writer.Write(line); err != nil {
			clientDisconnected = true
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Images stream client disconnected, continue draining upstream for billing")
			return
		}
		flusher.Flush()
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			if firstTokenMs == nil {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}
			trimmedLine := strings.TrimRight(string(line), "\r\n")
			if data, ok := extractOpenAISSEDataLine(trimmedLine); ok {
				sseData.AddLine(trimmedLine, processSSEData)
				if strings.TrimSpace(data) == "[DONE]" && !emitDone {
					suppressEvent = true
				}
			} else if strings.TrimSpace(trimmedLine) == "" {
				sseData.AddLine(trimmedLine, processSSEData)
				if suppressEvent {
					suppressEvent = false
					goto next
				}
			}
			if !suppressEvent {
				writeLine(line)
			}
		}
	next:
		if err == io.EOF {
			break
		}
		if err != nil {
			flushSSEEvent()
			return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, err
		}
	}
	flushSSEEvent()
	return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, nil
}

func (s *OpenAIGatewayService) buildOpenAIImagesRequest(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	contentType string,
	token string,
	endpoint string,
) (*http.Request, error) {
	targetURL := openAIImagesGenerationsURL
	if endpoint == openAIImagesEditsEndpoint {
		targetURL = openAIImagesEditsURL
	}
	baseURL := account.GetOpenAIBaseURL()
	if baseURL != "" {
		validatedURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return nil, err
		}
		targetURL = buildOpenAIImagesURL(validatedURL, endpoint)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	authHeaders, err := s.buildOpenAIAuthenticationHeaders(ctx, account, token)
	if err != nil {
		return nil, fmt.Errorf("build openai authentication headers: %w", err)
	}
	for key, values := range authHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	for key, values := range c.Request.Header {
		if !openaiPassthroughAllowedHeaders[strings.ToLower(key)] {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	customUA := account.GetOpenAIUserAgent()
	if customUA != "" {
		req.Header.Set("User-Agent", customUA)
	}
	if strings.TrimSpace(contentType) != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}

func buildOpenAIImagesURL(base string, endpoint string) string {
	return buildOpenAIEndpointURL(base, endpoint)
}

func isAPIMartImagesHost(account *Account) bool {
	if account == nil || !account.IsOpenAIApiKey() {
		return false
	}
	rawBase := strings.TrimSpace(account.GetOpenAIBaseURL())
	if rawBase == "" {
		return false
	}
	parsed, err := url.Parse(rawBase)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Hostname(), "api.apimart.ai")
}

func isAPIMartImagesAsyncModel(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return normalized == "gpt-image-2" ||
		normalized == "gpt-image-2-official" ||
		isAPIMartGeminiImageModel(normalized) ||
		isAPIMartMidjourneyImageModel(normalized) ||
		isAPIMartGrokImagineImageModel(normalized) ||
		isAPIMartSeedreamImageModel(normalized)
}

func (s *OpenAIGatewayService) forwardAPIMartImages(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *OpenAIImagesRequest,
	token string,
	requestModel string,
	upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	baseURL, err := s.validateUpstreamBaseURL(account.GetOpenAIBaseURL())
	if err != nil {
		return nil, err
	}
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	upstreamStart := time.Now()
	imageURLs, maskURL, cleanupInputs, err := s.prepareAPIMartImageInputs(ctx, account, token, proxyURL, baseURL, parsed)
	if err != nil {
		SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
		s.recordAPIMartImagesUpstreamError(c, account, err)
		return nil, err
	}
	defer cleanupInputs()

	requestID := ""
	images := make([]apimartImageResult, 0, maxInt(1, parsed.N))
	submitCount := apimartImagesSubmitCount(parsed, upstreamModel)
	for i := 0; i < submitCount; i++ {
		submitParsed := apimartImagesSubmitRequest(parsed, upstreamModel)
		submitBody, err := buildAPIMartImagesPayload(submitParsed, upstreamModel, imageURLs, maskURL)
		if err != nil {
			return nil, err
		}
		setOpsUpstreamRequestBody(c, submitBody)

		taskID, submitReqID, err := s.submitAPIMartImageTask(ctx, account, token, proxyURL, baseURL, upstreamModel, submitBody)
		if err != nil {
			SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
			s.recordAPIMartImagesUpstreamError(c, account, err)
			return nil, err
		}
		if requestID == "" {
			requestID = submitReqID
		}
		taskImages, err := s.pollAPIMartImageTask(ctx, account, token, proxyURL, baseURL, taskID, parsed)
		SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
		if err != nil {
			s.recordAPIMartImagesUpstreamError(c, account, err)
			return nil, err
		}
		if _, ok := apimartImageResultsCreditsCost(taskImages); !ok {
			err = fmt.Errorf("apimart image task completed without credits_cost")
			s.recordAPIMartImagesUpstreamError(c, account, err)
			return nil, err
		}
		images = append(images, taskImages...)
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("apimart image task completed without image urls")
	}

	imageOutputSizes := apimartImageResultSizes(images)
	costOverride := apimartImageResultCostOverride(images)
	responseCost := apimartImageResultResponseCost(images)
	totalCredits, _ := apimartImageResultsCreditsCost(images)
	responseCreditsCost := &totalCredits
	sizeResolution := resolveAPIMartImageBillingSize(parsed, imageOutputSizes)
	body, err := buildAPIMartOpenAIImagesResponse(apimartImageResultURLs(images), parsed, costOverride, responseCost, responseCreditsCost)
	if err != nil {
		return nil, err
	}
	c.Data(http.StatusOK, "application/json", body)

	return &OpenAIForwardResult{
		RequestID:          requestID,
		Usage:              OpenAIUsage{},
		Model:              requestModel,
		UpstreamModel:      upstreamModel,
		Stream:             false,
		Duration:           time.Since(startTime),
		ImageCount:         len(images),
		ImageSize:          sizeResolution.BillingSize,
		ImageOutputSizes:   imageOutputSizes,
		ImageOutputSize:    sizeResolution.OutputSize,
		ImageSizeSource:    sizeResolution.Source,
		ImageSizeBreakdown: sizeResolution.Breakdown,
		ImageQuality:       NormalizeImageQuality(parsed.Quality),
		ImageInputSize:     sizeResolution.InputSize,
		CostOverride:       costOverride,
	}, nil
}

func (s *OpenAIGatewayService) prepareAPIMartImageInputs(
	ctx context.Context,
	account *Account,
	token string,
	proxyURL string,
	baseURL string,
	parsed *OpenAIImagesRequest,
) ([]string, string, func(), error) {
	if parsed == nil {
		return nil, "", func() {}, fmt.Errorf("parsed images request is required")
	}
	useObjectURL := shouldUseOpenAIImagesObjectURLTransport(account, parsed)
	imageURLs := make([]string, 0, len(parsed.InputImageURLs)+len(parsed.Uploads))
	createdKeys := make([]string, 0, len(parsed.InputImageURLs)+len(parsed.Uploads)+1)
	appendKey := func(key string) {
		if strings.TrimSpace(key) != "" {
			createdKeys = append(createdKeys, key)
		}
	}
	cleanupCreated := func() {
		s.cleanupOpenAIImageInputObjects(createdKeys)
	}

	for _, rawURL := range parsed.InputImageURLs {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			continue
		}
		if !useObjectURL {
			imageURLs = append(imageURLs, rawURL)
			continue
		}
		converted, key, err := s.openAIImageInputURLAsObjectURL(ctx, account, rawURL)
		if err != nil {
			cleanupCreated()
			return nil, "", func() {}, err
		}
		imageURLs = append(imageURLs, converted)
		appendKey(key)
	}

	for _, upload := range parsed.Uploads {
		if !useObjectURL {
			uploadedURL, uploadErr := s.uploadAPIMartImage(ctx, account, token, proxyURL, baseURL, upload)
			if uploadErr != nil {
				return nil, "", func() {}, uploadErr
			}
			imageURLs = append(imageURLs, uploadedURL)
			continue
		}
		uploadedURL, key, uploadErr := s.uploadOpenAIImageInputObject(ctx, account, upload.Data, upload.ContentType, upload.FileName)
		if uploadErr != nil {
			cleanupCreated()
			return nil, "", func() {}, uploadErr
		}
		imageURLs = append(imageURLs, uploadedURL)
		appendKey(key)
	}

	maskURL := strings.TrimSpace(parsed.MaskImageURL)
	if parsed.MaskUpload != nil {
		if !useObjectURL {
			uploadedURL, uploadErr := s.uploadAPIMartImage(ctx, account, token, proxyURL, baseURL, *parsed.MaskUpload)
			if uploadErr != nil {
				return nil, "", func() {}, uploadErr
			}
			maskURL = uploadedURL
		} else {
			uploadedURL, key, uploadErr := s.uploadOpenAIImageInputObject(ctx, account, parsed.MaskUpload.Data, parsed.MaskUpload.ContentType, parsed.MaskUpload.FileName)
			if uploadErr != nil {
				cleanupCreated()
				return nil, "", func() {}, uploadErr
			}
			maskURL = uploadedURL
			appendKey(key)
		}
	} else if useObjectURL && maskURL != "" {
		converted, key, err := s.openAIImageInputURLAsObjectURL(ctx, account, maskURL)
		if err != nil {
			cleanupCreated()
			return nil, "", func() {}, err
		}
		maskURL = converted
		appendKey(key)
	}

	cleanup := func() {
		cleanupCreated()
	}
	return compactTrimmedStrings(imageURLs), maskURL, cleanup, nil
}

func splitOpenAIImagesRequestCount(parsed *OpenAIImagesRequest, upstreamModel string) int {
	if parsed == nil || parsed.N <= 1 {
		return 1
	}
	if oneImagePerRequestModel(upstreamModel) {
		return parsed.N
	}
	return 1
}

func shouldSplitOpenAIImagesRequests(parsed *OpenAIImagesRequest, upstreamModel string) bool {
	return splitOpenAIImagesRequestCount(parsed, upstreamModel) > 1
}

func splitOpenAIImagesSubmitRequest(parsed *OpenAIImagesRequest, upstreamModel string) *OpenAIImagesRequest {
	if parsed == nil || !oneImagePerRequestModel(upstreamModel) || parsed.N == 1 {
		return parsed
	}
	cloned := *parsed
	cloned.N = 1
	return &cloned
}

func oneImagePerRequestModel(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "gpt-image-2")
}

func apimartImagesSubmitCount(parsed *OpenAIImagesRequest, upstreamModel string) int {
	return splitOpenAIImagesRequestCount(parsed, upstreamModel)
}

func apimartImagesSubmitRequest(parsed *OpenAIImagesRequest, upstreamModel string) *OpenAIImagesRequest {
	return splitOpenAIImagesSubmitRequest(parsed, upstreamModel)
}

func buildAPIMartImagesPayload(parsed *OpenAIImagesRequest, upstreamModel string, imageURLs []string, maskURL string) ([]byte, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed images request is required")
	}
	upstreamModel = strings.TrimSpace(upstreamModel)
	if isAPIMartMidjourneyImageModel(upstreamModel) {
		return buildAPIMartMidjourneyImagesPayload(parsed, upstreamModel, imageURLs)
	}
	if isAPIMartGrokImagineImageModel(upstreamModel) {
		if isAPIMartGrokImagine20ExtModel(upstreamModel) || isAPIMartGrokImagine20ImageModel(upstreamModel) {
			return buildAPIMartGrokImagine20ImagesPayload(parsed, upstreamModel, imageURLs)
		}
		return buildAPIMartGrokImagineImagesPayload(parsed, upstreamModel, imageURLs)
	}
	if isAPIMartSeedream50ProModel(upstreamModel) || isAPIMartSeedream50LiteModel(upstreamModel) {
		return buildAPIMartSeedream50ImagesPayload(parsed, upstreamModel, imageURLs)
	}
	payload := map[string]any{
		"model":      upstreamModel,
		"prompt":     strings.TrimSpace(parsed.Prompt),
		"n":          parsed.N,
		"resolution": apimartImagesResolution(parsed),
	}
	if size := strings.TrimSpace(parsed.Size); size != "" {
		payload["size"] = size
	}
	if !isAPIMartGeminiImageModel(upstreamModel) {
		if !isAPIMartSeedreamImageModel(upstreamModel) {
			if quality := strings.TrimSpace(parsed.Quality); quality != "" {
				payload["quality"] = quality
			}
			if background := strings.TrimSpace(parsed.Background); background != "" {
				payload["background"] = background
			}
			if moderation := strings.TrimSpace(parsed.Moderation); moderation != "" {
				payload["moderation"] = moderation
			}
			if outputFormat := strings.TrimSpace(parsed.OutputFormat); outputFormat != "" {
				payload["output_format"] = outputFormat
			}
			if parsed.OutputCompression != nil {
				payload["output_compression"] = *parsed.OutputCompression
			}
		}
	}
	if strings.EqualFold(upstreamModel, "gpt-image-2") {
		payload["official_fallback"] = false
	}
	if isAPIMartGeminiImageModel(upstreamModel) && isAPIMartGeminiFlashImageModel(upstreamModel) {
		if parsed.GoogleSearch != nil {
			payload["google_search"] = *parsed.GoogleSearch
		}
		if parsed.GoogleImageSearch != nil {
			payload["google_image_search"] = *parsed.GoogleImageSearch
		}
	}
	imageURLs = compactTrimmedStrings(imageURLs)
	if len(imageURLs) > 0 {
		payload["image_urls"] = imageURLs
	}
	if maskURL := strings.TrimSpace(maskURL); maskURL != "" {
		payload["mask_url"] = maskURL
	}
	return json.Marshal(payload)
}

func buildAPIMartSeedream50ImagesPayload(parsed *OpenAIImagesRequest, upstreamModel string, imageURLs []string) ([]byte, error) {
	payload := map[string]any{
		"model":  upstreamModel,
		"prompt": strings.TrimSpace(parsed.Prompt),
		"n":      parsed.N,
	}
	if size := strings.TrimSpace(parsed.Size); size != "" {
		payload["size"] = size
	}
	if resolution := strings.TrimSpace(parsed.Resolution); resolution != "" {
		payload["resolution"] = resolution
	}
	imageURLs = compactTrimmedStrings(imageURLs)
	if len(imageURLs) > 0 {
		payload["image_urls"] = imageURLs
	}
	if parsed.NSFWCheck != nil {
		payload["nsfw_check"] = *parsed.NSFWCheck
	}
	if outputFormat := strings.TrimSpace(parsed.OutputFormat); outputFormat != "" {
		payload["output_format"] = outputFormat
	}
	if parsed.Watermark != nil {
		payload["watermark"] = *parsed.Watermark
	}
	if isAPIMartSeedream50LiteModel(upstreamModel) {
		mode := strings.TrimSpace(parsed.SequentialMode)
		if parsed.N > 1 && mode == "" {
			mode = "auto"
		}
		if mode != "" {
			payload["sequential_image_generation"] = mode
		}
		if parsed.SequentialMax != nil {
			payload["sequential_image_generation_options"] = map[string]int{"max_images": *parsed.SequentialMax}
		}
	}
	return json.Marshal(payload)
}

func buildAPIMartMidjourneyImagesPayload(parsed *OpenAIImagesRequest, upstreamModel string, imageURLs []string) ([]byte, error) {
	payload := map[string]any{
		"model":  upstreamModel,
		"prompt": strings.TrimSpace(parsed.Prompt),
		"n":      parsed.N,
	}
	if size := strings.TrimSpace(parsed.Size); size != "" {
		payload["size"] = size
	}
	if version := strings.TrimSpace(parsed.MidjourneyVersion); version != "" {
		payload["version"] = version
	}
	if speed := strings.TrimSpace(parsed.MidjourneySpeed); speed != "" {
		payload["speed"] = speed
	}
	if quality := strings.TrimSpace(parsed.Quality); quality != "" {
		payload["quality"] = quality
	}
	if parsed.MidjourneyStylize != nil {
		payload["stylize"] = *parsed.MidjourneyStylize
	}
	if parsed.MidjourneyChaos != nil {
		payload["chaos"] = *parsed.MidjourneyChaos
	}
	if parsed.MidjourneyWeird != nil {
		payload["weird"] = *parsed.MidjourneyWeird
	}
	if parsed.MidjourneyStop != nil && midjourneyVersionSupportsStop(parsed.MidjourneyVersion) {
		payload["stop"] = *parsed.MidjourneyStop
	}
	if parsed.MidjourneyNiji != nil {
		payload["niji"] = *parsed.MidjourneyNiji
	}
	if parsed.MidjourneyRaw != nil {
		payload["raw"] = *parsed.MidjourneyRaw
	}
	if parsed.MidjourneyTile != nil {
		payload["tile"] = *parsed.MidjourneyTile
	}
	imageURLs = compactTrimmedStrings(imageURLs)
	if len(imageURLs) > 0 {
		payload["image_urls"] = imageURLs
	}
	return json.Marshal(payload)
}

func midjourneyVersionSupportsStop(version string) bool {
	switch strings.TrimPrefix(strings.ToLower(strings.TrimSpace(version)), "v") {
	case "5", "5.1", "5.2", "6", "6.1":
		return true
	default:
		return false
	}
}

func buildAPIMartGrokImagineImagesPayload(parsed *OpenAIImagesRequest, upstreamModel string, imageURLs []string) ([]byte, error) {
	payload := map[string]any{
		"model":  upstreamModel,
		"prompt": strings.TrimSpace(parsed.Prompt),
		"n":      parsed.N,
	}
	if size := strings.TrimSpace(parsed.Size); size != "" {
		payload["size"] = size
	}
	imageURLs = compactTrimmedStrings(imageURLs)
	if isAPIMartGrokImagineEditModel(upstreamModel) {
		if len(imageURLs) == 0 {
			return nil, fmt.Errorf("%s requires image_urls", upstreamModel)
		}
		payload["image_urls"] = imageURLs[:1]
	}
	return json.Marshal(payload)
}

func buildAPIMartGrokImagine20ImagesPayload(parsed *OpenAIImagesRequest, upstreamModel string, imageURLs []string) ([]byte, error) {
	payload := map[string]any{
		"model":  upstreamModel,
		"prompt": strings.TrimSpace(parsed.Prompt),
		"n":      parsed.N,
	}
	if isAPIMartGrokImagine20ExtModel(upstreamModel) {
		if len(imageURLs) > 0 {
			return nil, fmt.Errorf("%s does not support reference images", upstreamModel)
		}
		if size := strings.TrimSpace(parsed.Size); size != "" {
			payload["size"] = size
		}
		if responseFormat := strings.TrimSpace(parsed.ResponseFormat); responseFormat != "" {
			payload["response_format"] = responseFormat
		}
	} else {
		if aspectRatio := strings.TrimSpace(parsed.AspectRatio); aspectRatio != "" {
			payload["aspect_ratio"] = aspectRatio
		}
		if len(imageURLs) > 0 {
			payload["image_urls"] = compactTrimmedStrings(imageURLs)
		}
		if quality := strings.TrimSpace(parsed.Quality); quality != "" && len(imageURLs) == 0 {
			payload["quality"] = quality
		}
	}
	if resolution := strings.TrimSpace(parsed.Resolution); resolution != "" {
		payload["resolution"] = resolution
	}
	if parsed.NSFWCheck != nil {
		payload["nsfw_check"] = *parsed.NSFWCheck
	}
	return json.Marshal(payload)
}

func buildOpenAIImagesURLFieldsPayload(parsed *OpenAIImagesRequest, upstreamModel string, imageURLs []string, maskURL string) ([]byte, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed images request is required")
	}
	payload := map[string]any{
		"model":  strings.TrimSpace(upstreamModel),
		"prompt": strings.TrimSpace(parsed.Prompt),
		"n":      parsed.N,
	}
	for _, field := range []struct {
		key   string
		value string
	}{
		{key: "size", value: parsed.Size},
		{key: "response_format", value: parsed.ResponseFormat},
		{key: "quality", value: parsed.Quality},
		{key: "background", value: parsed.Background},
		{key: "output_format", value: parsed.OutputFormat},
		{key: "moderation", value: parsed.Moderation},
		{key: "input_fidelity", value: parsed.InputFidelity},
		{key: "style", value: parsed.Style},
		{key: "version", value: parsed.MidjourneyVersion},
		{key: "speed", value: parsed.MidjourneySpeed},
	} {
		if trimmed := strings.TrimSpace(field.value); trimmed != "" {
			payload[field.key] = trimmed
		}
	}
	if parsed.OutputCompression != nil {
		payload["output_compression"] = *parsed.OutputCompression
	}
	if parsed.PartialImages != nil {
		payload["partial_images"] = *parsed.PartialImages
	}
	if parsed.GoogleSearch != nil {
		payload["google_search"] = *parsed.GoogleSearch
	}
	if parsed.GoogleImageSearch != nil {
		payload["google_image_search"] = *parsed.GoogleImageSearch
	}
	if parsed.MidjourneyStylize != nil {
		payload["stylize"] = *parsed.MidjourneyStylize
	}
	if parsed.MidjourneyChaos != nil {
		payload["chaos"] = *parsed.MidjourneyChaos
	}
	if parsed.MidjourneyWeird != nil {
		payload["weird"] = *parsed.MidjourneyWeird
	}
	if parsed.MidjourneyStop != nil {
		payload["stop"] = *parsed.MidjourneyStop
	}
	if parsed.MidjourneyNiji != nil {
		payload["niji"] = *parsed.MidjourneyNiji
	}
	if parsed.MidjourneyRaw != nil {
		payload["raw"] = *parsed.MidjourneyRaw
	}
	if parsed.MidjourneyTile != nil {
		payload["tile"] = *parsed.MidjourneyTile
	}
	imageURLs = compactTrimmedStrings(imageURLs)
	if len(imageURLs) > 0 {
		payload["image_urls"] = imageURLs
	}
	if maskURL := strings.TrimSpace(maskURL); maskURL != "" {
		payload["mask_url"] = maskURL
	}
	return json.Marshal(payload)
}

func apimartImagesResolution(parsed *OpenAIImagesRequest) string {
	if parsed == nil {
		return apimartImagesDefaultResolution
	}
	switch strings.ToLower(strings.TrimSpace(parsed.Resolution)) {
	case "1k":
		return "1k"
	case "2k":
		return "2k"
	case "4k":
		return "4k"
	}
	size := strings.TrimSpace(parsed.Size)
	if size == "" {
		return apimartImagesDefaultResolution
	}
	if strings.Contains(size, ":") {
		return apimartImagesDefaultResolution
	}
	if tier, ok := apimartKnownImageResolution(size); ok {
		return tier
	}
	switch strings.ToUpper(strings.TrimSpace(parsed.SizeTier)) {
	case ImageBillingSize1K:
		return "1k"
	case ImageBillingSize2K:
		return "2k"
	case ImageBillingSize4K:
		return "4k"
	default:
		return apimartImagesDefaultResolution
	}
}

func apimartImagesBillingSize(parsed *OpenAIImagesRequest) string {
	switch apimartImagesResolution(parsed) {
	case "1k":
		return ImageBillingSize1K
	case "2k":
		return ImageBillingSize2K
	case "4k":
		return ImageBillingSize4K
	default:
		return ImageBillingSize1K
	}
}

func apimartImagesBillingInputSize(parsed *OpenAIImagesRequest, outputSizes []string) string {
	if len(outputSizes) > 0 {
		if parsed == nil {
			return ""
		}
		size := strings.TrimSpace(parsed.Size)
		if _, ok := ClassifyImageBillingTier(size); ok {
			return size
		}
		return apimartImagesBillingSize(parsed)
	}
	if parsed != nil {
		size := strings.TrimSpace(parsed.Size)
		if _, ok := ClassifyImageBillingTier(size); ok {
			return size
		}
	}
	return apimartImagesBillingSize(parsed)
}

func resolveAPIMartImageBillingSize(parsed *OpenAIImagesRequest, outputSizes []string) ImageBillingSizeResolution {
	inputSize := apimartImagesBillingInputSize(parsed, outputSizes)
	outputSizes = compactTrimmedStrings(outputSizes)

	breakdown := map[string]int{}
	outputSize := firstDisplayImageOutputSize(outputSizes)
	outputTier := ""
	for _, output := range outputSizes {
		resolution, ok := apimartKnownImageResolution(output)
		if !ok {
			continue
		}
		tier := apimartImageResolutionBillingSize(resolution)
		if tier == "" {
			continue
		}
		breakdown[tier]++
		if imageTierRank(tier) > imageTierRank(outputTier) {
			outputTier = tier
		}
	}
	if outputTier != "" {
		return ImageBillingSizeResolution{
			BillingSize: outputTier,
			InputSize:   inputSize,
			OutputSize:  outputSize,
			Source:      ImageSizeSourceOutput,
			Breakdown:   normalizeImageSizeBreakdown(breakdown),
		}
	}

	if resolution := ResolveImageBillingSize(inputSize, outputSizes); resolution.Source == ImageSizeSourceOutput {
		return resolution
	}

	return ImageBillingSizeResolution{
		BillingSize: apimartImagesBillingSize(parsed),
		InputSize:   inputSize,
		OutputSize:  outputSize,
		Source:      ImageSizeSourceInput,
	}
}

func apimartImageResolutionBillingSize(resolution string) string {
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "1k":
		return ImageBillingSize1K
	case "2k":
		return ImageBillingSize2K
	case "4k":
		return ImageBillingSize4K
	default:
		return ""
	}
}

func apimartKnownImageResolution(size string) (string, bool) {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(size), "×", "x"))
	switch normalized {
	case "1024x1024", "1254x1254", "1536x1024", "1024x1536", "1024x768", "768x1024",
		"1280x1024", "1448x1086", "1024x1280", "1122x1402", "1536x864", "1672x941",
		"864x1536", "941x1672", "2048x1024", "1774x887", "1024x2048", "887x1774",
		"1881x836", "1536x512", "512x1536", "2016x864", "1915x821", "864x2016", "821x1915":
		return "1k", true
	case "2048x2048", "2048x1360", "1360x2048", "2048x1536", "1536x2048", "2560x2048",
		"2048x2560", "2048x1152", "1152x2048", "2688x1344", "1344x2688", "3072x1024",
		"1024x3072", "2688x1152", "1152x2688":
		return "2k", true
	case "2880x2880", "3520x2336", "2336x3520", "3312x2480", "2480x3312", "3216x2576",
		"2576x3216", "3840x2160", "2160x3840", "3840x1920", "1920x3840", "3840x1280",
		"1280x3840", "3840x1648", "1648x3840":
		return "4k", true
	default:
		return "", false
	}
}

func (s *OpenAIGatewayService) uploadAPIMartImage(
	ctx context.Context,
	account *Account,
	token string,
	proxyURL string,
	baseURL string,
	upload OpenAIImagesUpload,
) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeQuotes(firstNonEmptyString(upload.FileName, "image.png"))))
	if contentType := strings.TrimSpace(upload.ContentType); contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := writer.CreatePart(header)
	if err != nil {
		return "", fmt.Errorf("create apimart image upload part: %w", err)
	}
	if _, err := part.Write(upload.Data); err != nil {
		return "", fmt.Errorf("write apimart image upload part: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("finalize apimart image upload: %w", err)
	}

	targetURL := buildOpenAIEndpointURL(baseURL, apimartImagesUploadEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body.Bytes()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	respBody, _, err := s.doAPIMartImagesRequest(req, proxyURL, account)
	if err != nil {
		return "", err
	}
	uploadedURL := strings.TrimSpace(gjson.GetBytes(respBody, "url").String())
	if uploadedURL == "" {
		uploadedURL = strings.TrimSpace(gjson.GetBytes(respBody, "data.url").String())
	}
	if uploadedURL == "" {
		return "", fmt.Errorf("apimart image upload response missing url")
	}
	return uploadedURL, nil
}

func (s *OpenAIGatewayService) submitAPIMartImageTask(
	ctx context.Context,
	account *Account,
	token string,
	proxyURL string,
	baseURL string,
	upstreamModel string,
	body []byte,
) (string, string, error) {
	endpoint := apimartImagesGenerationsEndpoint
	if isAPIMartMidjourneyImageModel(upstreamModel) {
		endpoint = apimartMidjourneyEndpoint
	} else if isAPIMartGrokImagineEditModel(upstreamModel) {
		endpoint = apimartImagesEditsEndpoint
	}
	targetURL := buildOpenAIEndpointURL(baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	respBody, requestID, err := s.doAPIMartImagesRequest(req, proxyURL, account)
	if err != nil {
		return "", "", err
	}
	taskID := extractAPIMartImageTaskID(respBody)
	if taskID == "" {
		return "", requestID, fmt.Errorf("apimart image generation response missing task_id")
	}
	return taskID, requestID, nil
}

func (s *OpenAIGatewayService) pollAPIMartImageTask(
	ctx context.Context,
	account *Account,
	token string,
	proxyURL string,
	baseURL string,
	taskID string,
	parsed *OpenAIImagesRequest,
) ([]apimartImageResult, error) {
	targetURL := buildOpenAIEndpointURL(baseURL, apimartImagesTaskEndpointPrefix+url.PathEscape(taskID)) + "?language=zh"
	var lastStatus string
	var lastMessage string
	for attempt := 0; attempt < apimartImagesMaxPolls; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")

		respBody, _, err := s.doAPIMartImagesRequest(req, proxyURL, account)
		if err != nil {
			return nil, err
		}
		status := strings.ToLower(strings.TrimSpace(firstNonEmptyString(
			gjson.GetBytes(respBody, "data.status").String(),
			gjson.GetBytes(respBody, "status").String(),
		)))
		lastStatus = status
		lastMessage = extractAPIMartImageMessage(respBody)
		switch status {
		case "completed", "succeeded", "success":
			images := extractAPIMartImageResults(respBody, parsed)
			if len(images) == 0 {
				return nil, fmt.Errorf("apimart image task completed without image urls")
			}
			return images, nil
		case "failed", "cancelled", "canceled":
			if lastMessage == "" {
				lastMessage = "task failed"
			}
			return nil, fmt.Errorf("apimart image task %s: %s", status, sanitizeUpstreamErrorMessage(lastMessage))
		}

		timer := time.NewTimer(apimartImagesPollInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	if lastStatus == "" {
		lastStatus = "unknown"
	}
	if lastMessage == "" {
		lastMessage = "task polling timed out"
	}
	return nil, fmt.Errorf("apimart image task timeout: status=%s message=%s", lastStatus, sanitizeUpstreamErrorMessage(lastMessage))
}

func (s *OpenAIGatewayService) doAPIMartImagesRequest(req *http.Request, proxyURL string, account *Account) ([]byte, string, error) {
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		return nil, "", fmt.Errorf("apimart upstream request failed: %s", safeErr)
	}
	if resp == nil {
		return nil, "", fmt.Errorf("apimart upstream request failed: empty response")
	}
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, apimartImagesMaxResponseBytes))
	if readErr != nil {
		return nil, resp.Header.Get("x-request-id"), readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(body))
		if message == "" {
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		limitedBody := body
		if len(limitedBody) > apimartImagesMaxErrorBytes {
			limitedBody = limitedBody[:apimartImagesMaxErrorBytes]
		}
		return nil, resp.Header.Get("x-request-id"), &openAIImageStatusError{
			StatusCode:      resp.StatusCode,
			Message:         message,
			ResponseBody:    limitedBody,
			ResponseHeaders: resp.Header.Clone(),
			RequestID:       resp.Header.Get("x-request-id"),
			URL:             safeUpstreamURL(req.URL.String()),
		}
	}
	if code := gjson.GetBytes(body, "code"); code.Exists() && code.Int() != 0 && code.Int() != 200 {
		message := sanitizeUpstreamErrorMessage(extractAPIMartImageMessage(body))
		if message == "" {
			message = fmt.Sprintf("code %d", code.Int())
		}
		return nil, resp.Header.Get("x-request-id"), &openAIImageStatusError{
			StatusCode:      http.StatusBadGateway,
			Message:         message,
			ResponseBody:    body,
			ResponseHeaders: resp.Header.Clone(),
			RequestID:       resp.Header.Get("x-request-id"),
			URL:             safeUpstreamURL(req.URL.String()),
		}
	}
	return body, resp.Header.Get("x-request-id"), nil
}

func extractAPIMartImageTaskID(body []byte) string {
	for _, path := range []string{"data.0.task_id", "data.task_id", "task_id", "id"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ""
}

func extractAPIMartImageMessage(body []byte) string {
	for _, path := range []string{
		"message",
		"error.message",
		"data.message",
		"data.error",
		"data.error.message",
	} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ""
}

type apimartImageResult struct {
	URL         string
	Size        string
	Cost        *float64
	CreditsCost *float64
}

func extractAPIMartImageResults(body []byte, parsed *OpenAIImagesRequest) []apimartImageResult {
	var out []apimartImageResult
	fallbackSize := apimartExplicitPixelSize(parsed)
	taskCost := firstAPIMartTaskCost(body)
	taskCreditsCost := firstAPIMartTaskCreditsCost(body)
	for _, path := range []string{
		"data.result.images",
		"result.images",
		"data.images",
		"images",
	} {
		items := gjson.GetBytes(body, path)
		if !items.IsArray() {
			continue
		}
		for _, item := range items.Array() {
			cost := takeAPIMartTaskCost(&taskCost)
			creditsCost := takeAPIMartTaskCost(&taskCreditsCost)
			if item.Type == gjson.String {
				out = appendAPIMartImageResult(out, item.String(), fallbackSize, cost, creditsCost)
				continue
			}
			size := firstAPIMartImageResultSize(item, fallbackSize)
			if cost == nil {
				cost = firstAPIMartImageResultCost(item)
			}
			if creditsCost == nil {
				creditsCost = firstAPIMartImageResultCreditsCost(item)
			}
			urls := item.Get("url")
			if urls.IsArray() {
				for i, u := range urls.Array() {
					imageCost := cost
					imageCreditsCost := creditsCost
					if i > 0 && taskCost == nil {
						imageCost = nil
					}
					if i > 0 && taskCreditsCost == nil {
						imageCreditsCost = nil
					}
					out = appendAPIMartImageResult(out, u.String(), size, imageCost, imageCreditsCost)
				}
				cost = nil
				creditsCost = nil
			} else if urls.Type == gjson.String {
				out = appendAPIMartImageResult(out, urls.String(), size, cost, creditsCost)
				cost = nil
				creditsCost = nil
			}
			if value := strings.TrimSpace(item.Get("image_url").String()); value != "" {
				out = appendAPIMartImageResult(out, value, size, cost, creditsCost)
			}
		}
		if len(out) > 0 {
			break
		}
	}
	return compactAPIMartImageResults(out)
}

func extractAPIMartImageResultURLs(body []byte) []string {
	return apimartImageResultURLs(extractAPIMartImageResults(body, nil))
}

func appendAPIMartImageResult(out []apimartImageResult, rawURL string, size string, cost *float64, creditsCosts ...*float64) []apimartImageResult {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return out
	}
	var creditsCost *float64
	if len(creditsCosts) > 0 {
		creditsCost = creditsCosts[0]
	}
	return append(out, apimartImageResult{
		URL:         rawURL,
		Size:        normalizeAPIMartImageSize(size),
		Cost:        cost,
		CreditsCost: creditsCost,
	})
}

func compactAPIMartImageResults(results []apimartImageResult) []apimartImageResult {
	if len(results) == 0 {
		return nil
	}
	out := make([]apimartImageResult, 0, len(results))
	for _, result := range results {
		result.URL = strings.TrimSpace(result.URL)
		result.Size = normalizeAPIMartImageSize(result.Size)
		if result.URL != "" {
			out = append(out, result)
		}
	}
	return out
}

func apimartImageResultURLs(results []apimartImageResult) []string {
	if len(results) == 0 {
		return nil
	}
	out := make([]string, 0, len(results))
	for _, result := range results {
		if url := strings.TrimSpace(result.URL); url != "" {
			out = append(out, url)
		}
	}
	return out
}

func apimartImageResultSizes(results []apimartImageResult) []string {
	if len(results) == 0 {
		return nil
	}
	out := make([]string, 0, len(results))
	for _, result := range results {
		if size := normalizeAPIMartImageSize(result.Size); size != "" {
			out = append(out, size)
		}
	}
	return out
}

func apimartImageResultCostOverride(results []apimartImageResult) *CostBreakdown {
	totalCredits, ok := apimartImageResultsCreditsCost(results)
	if !ok {
		return nil
	}
	total := totalCredits / apimartCreditsPerCost
	return &CostBreakdown{
		TotalCost:   total,
		ActualCost:  0,
		BillingMode: string(BillingModeImage),
	}
}

func apimartImageResultsCreditsCost(results []apimartImageResult) (float64, bool) {
	total := 0.0
	found := false
	for _, result := range results {
		if result.CreditsCost != nil && *result.CreditsCost > 0 {
			total += *result.CreditsCost
			found = true
		}
	}
	return total, found
}

func apimartImageResultResponseCost(results []apimartImageResult) *float64 {
	totalCredits, ok := apimartImageResultsCreditsCost(results)
	if !ok {
		return nil
	}
	total := totalCredits / apimartCreditsPerCost
	return &total
}

func firstAPIMartImageResultSize(item gjson.Result, fallbackSize string) string {
	for _, path := range []string{
		"size",
		"output_size",
		"resolution_size",
		"dimensions",
		"metadata.size",
		"metadata.output_size",
	} {
		if size := normalizeAPIMartImageSize(item.Get(path).String()); size != "" {
			return size
		}
	}
	width, height := firstAPIMartImageDimensions(item)
	if width > 0 && height > 0 {
		return fmt.Sprintf("%dx%d", width, height)
	}
	return fallbackSize
}

func takeAPIMartTaskCost(cost **float64) *float64 {
	if cost == nil || *cost == nil {
		return nil
	}
	out := *cost
	*cost = nil
	return out
}

func firstAPIMartTaskCost(body []byte) *float64 {
	for _, path := range []string{
		"data.cost",
		"cost",
		"result.cost",
		"data.billing.cost",
		"billing.cost",
	} {
		value := gjson.GetBytes(body, path)
		if !value.Exists() {
			continue
		}
		cost := value.Float()
		if cost > 0 {
			return &cost
		}
	}
	return nil
}

func firstAPIMartTaskCreditsCost(body []byte) *float64 {
	for _, path := range []string{
		"data.credits_cost",
		"credits_cost",
		"result.credits_cost",
		"data.billing.credits_cost",
		"billing.credits_cost",
	} {
		value := gjson.GetBytes(body, path)
		if !value.Exists() {
			continue
		}
		creditsCost := value.Float()
		if creditsCost > 0 {
			return &creditsCost
		}
	}
	return nil
}

func firstAPIMartImageResultCost(item gjson.Result) *float64 {
	for _, path := range []string{
		"cost",
		"charge",
		"fee",
		"metadata.cost",
		"billing.cost",
	} {
		value := item.Get(path)
		if !value.Exists() {
			continue
		}
		cost := value.Float()
		if cost > 0 {
			return &cost
		}
	}
	return nil
}

func firstAPIMartImageResultCreditsCost(item gjson.Result) *float64 {
	for _, path := range []string{
		"credits_cost",
		"metadata.credits_cost",
		"billing.credits_cost",
	} {
		value := item.Get(path)
		if !value.Exists() {
			continue
		}
		creditsCost := value.Float()
		if creditsCost > 0 {
			return &creditsCost
		}
	}
	return nil
}

func firstAPIMartImageDimensions(item gjson.Result) (int64, int64) {
	for _, pair := range [][2]string{
		{"width", "height"},
		{"w", "h"},
		{"metadata.width", "metadata.height"},
		{"metadata.w", "metadata.h"},
	} {
		width := item.Get(pair[0]).Int()
		height := item.Get(pair[1]).Int()
		if width > 0 && height > 0 {
			return width, height
		}
	}
	return 0, 0
}

func apimartExplicitPixelSize(parsed *OpenAIImagesRequest) string {
	if parsed == nil || !parsed.ExplicitSize {
		return ""
	}
	size := normalizeAPIMartImageSize(parsed.Size)
	if _, _, ok := parseImageBillingDimensions(size); ok {
		return size
	}
	return ""
}

func buildAPIMartOpenAIImagesResponse(images []string, parsed *OpenAIImagesRequest, costOverride *CostBreakdown, responseCost *float64, responseCreditsCost *float64) ([]byte, error) {
	out := []byte(`{"created":0,"data":[]}`)
	out, _ = sjson.SetBytes(out, "created", time.Now().Unix())
	if responseCreditsCost != nil && *responseCreditsCost > 0 {
		out, _ = sjson.SetBytes(out, "credits_cost", *responseCreditsCost)
	}
	if responseCost != nil && *responseCost > 0 {
		out, _ = sjson.SetBytes(out, "cost", *responseCost)
	} else if costOverride != nil && costOverride.TotalCost > 0 {
		out, _ = sjson.SetBytes(out, "cost", costOverride.TotalCost)
	}
	prompt := ""
	if parsed != nil {
		prompt = strings.TrimSpace(parsed.Prompt)
	}
	for _, imageURL := range compactTrimmedStrings(images) {
		item := []byte(`{}`)
		item, _ = sjson.SetBytes(item, "url", imageURL)
		if prompt != "" {
			item, _ = sjson.SetBytes(item, "revised_prompt", prompt)
		}
		out, _ = sjson.SetRawBytes(out, "data.-1", item)
	}
	return out, nil
}

func buildSplitOpenAIImagesAPIResponse(responseBodies [][]byte) ([]byte, int, error) {
	out := []byte(`{"created":0,"data":[]}`)
	createdAt := int64(0)
	var usage OpenAIUsage
	imageCount := 0
	metadataCopied := false

	for _, body := range responseBodies {
		if len(body) == 0 || !gjson.ValidBytes(body) {
			return nil, 0, fmt.Errorf("invalid image response body")
		}
		root := gjson.ParseBytes(body)
		if createdAt <= 0 {
			createdAt = root.Get("created").Int()
		}
		data := root.Get("data")
		if data.IsArray() {
			for _, item := range data.Array() {
				out, _ = sjson.SetRawBytes(out, "data.-1", []byte(item.Raw))
				imageCount++
			}
		}
		if parsedUsage, ok := extractOpenAIUsageFromJSONBytes(body); ok {
			addOpenAIUsage(&usage, parsedUsage)
		}
		if !metadataCopied {
			for _, path := range []string{"background", "output_format", "quality", "size", "model"} {
				if value := root.Get(path); value.Exists() {
					out, _ = sjson.SetRawBytes(out, path, []byte(value.Raw))
				}
			}
			metadataCopied = true
		}
	}
	if createdAt <= 0 {
		createdAt = time.Now().Unix()
	}
	out, _ = sjson.SetBytes(out, "created", createdAt)
	if usage.hasValues() {
		out, _ = sjson.SetRawBytes(out, "usage", buildOpenAIUsageJSON(usage))
	}
	if imageCount == 0 {
		return nil, 0, fmt.Errorf("upstream did not return image output")
	}
	return out, imageCount, nil
}

func addOpenAIUsage(dst *OpenAIUsage, src OpenAIUsage) {
	if dst == nil {
		return
	}
	dst.InputTokens += src.InputTokens
	dst.ImageInputTokens += src.ImageInputTokens
	dst.OutputTokens += src.OutputTokens
	dst.CacheCreationInputTokens += src.CacheCreationInputTokens
	dst.CacheReadInputTokens += src.CacheReadInputTokens
	dst.ImageOutputTokens += src.ImageOutputTokens
}

func (u OpenAIUsage) hasValues() bool {
	return u.InputTokens > 0 ||
		u.ImageInputTokens > 0 ||
		u.OutputTokens > 0 ||
		u.CacheCreationInputTokens > 0 ||
		u.CacheReadInputTokens > 0 ||
		u.ImageOutputTokens > 0
}

func buildOpenAIUsageJSON(usage OpenAIUsage) []byte {
	out := []byte(`{}`)
	out, _ = sjson.SetBytes(out, "input_tokens", usage.InputTokens)
	out, _ = sjson.SetBytes(out, "output_tokens", usage.OutputTokens)
	if usage.ImageInputTokens > 0 {
		out, _ = sjson.SetBytes(out, "input_tokens_details.image_tokens", usage.ImageInputTokens)
	}
	if usage.CacheCreationInputTokens > 0 {
		out, _ = sjson.SetBytes(out, "cache_creation_input_tokens", usage.CacheCreationInputTokens)
	}
	if usage.CacheReadInputTokens > 0 {
		out, _ = sjson.SetBytes(out, "input_tokens_details.cached_tokens", usage.CacheReadInputTokens)
	}
	if usage.ImageOutputTokens > 0 {
		out, _ = sjson.SetBytes(out, "output_tokens_details.image_tokens", usage.ImageOutputTokens)
	}
	return out
}

func (s *OpenAIGatewayService) recordAPIMartImagesUpstreamError(c *gin.Context, account *Account, err error) {
	if err == nil {
		return
	}
	statusCode := 0
	requestID := ""
	upstreamURL := ""
	body := []byte(nil)
	if statusErr, ok := err.(*openAIImageStatusError); ok && statusErr != nil {
		statusCode = statusErr.StatusCode
		requestID = statusErr.RequestID
		upstreamURL = statusErr.URL
		body = statusErr.ResponseBody
	}
	message := sanitizeUpstreamErrorMessage(err.Error())
	setOpsUpstreamError(c, statusCode, message, "")
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		ProxyID:              opsUpstreamProxyID(account),
		ProxyName:            opsUpstreamProxyName(account),
		Platform:             account.Platform,
		AccountID:            account.ID,
		AccountName:          account.Name,
		UpstreamStatusCode:   statusCode,
		UpstreamRequestID:    requestID,
		UpstreamURL:          upstreamURL,
		Kind:                 "request_error",
		Message:              message,
		UpstreamResponseBody: truncateString(string(body), 2048),
	})
}

func rewriteOpenAIImagesModel(body []byte, contentType string, model string) ([]byte, string, error) {
	return rewriteOpenAIImagesModelAndN(body, contentType, model, 0, false)
}

func rewriteOpenAIImagesModelAndN(body []byte, contentType string, model string, n int, rewriteN bool) ([]byte, string, error) {
	model = strings.TrimSpace(model)
	if model == "" && !rewriteN {
		return body, contentType, nil
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err == nil && strings.EqualFold(mediaType, "multipart/form-data") {
		rewrittenBody, rewrittenType, rewriteErr := rewriteOpenAIImagesMultipartModelAndN(body, contentType, model, n, rewriteN)
		return rewrittenBody, rewrittenType, rewriteErr
	}
	rewritten, err := normalizeOpenAIImagesJSONCompatibilityFields(body)
	if err != nil {
		return nil, "", err
	}
	if model != "" {
		rewritten, err = sjson.SetBytes(rewritten, "model", model)
		if err != nil {
			return nil, "", fmt.Errorf("rewrite image request model: %w", err)
		}
	}
	if rewriteN {
		rewritten, err = sjson.SetBytes(rewritten, "n", n)
		if err != nil {
			return nil, "", fmt.Errorf("rewrite image request n: %w", err)
		}
	}
	return rewritten, contentType, nil
}

func normalizeOpenAIImagesJSONCompatibilityFields(body []byte) ([]byte, error) {
	rewritten, _, err := stripOpenAILocalGroupID(body)
	if err != nil {
		return nil, fmt.Errorf("remove image request local-only fields: %w", err)
	}
	transparentBackground := gjson.GetBytes(rewritten, "transparent_background")
	if !transparentBackground.Exists() {
		return rewritten, nil
	}
	if strings.TrimSpace(gjson.GetBytes(rewritten, "background").String()) == "" {
		background, ok, err := parseOpenAIImagesTransparentBackgroundJSON(transparentBackground)
		if err != nil {
			return nil, err
		}
		if ok {
			var setErr error
			rewritten, setErr = sjson.SetBytes(rewritten, "background", background)
			if setErr != nil {
				return nil, fmt.Errorf("rewrite image request background: %w", setErr)
			}
		}
	}
	rewritten, err = sjson.DeleteBytes(rewritten, "transparent_background")
	if err != nil {
		return nil, fmt.Errorf("remove image request transparent_background: %w", err)
	}
	return rewritten, nil
}

func rewriteOpenAIImagesMultipartModel(body []byte, contentType string, model string) ([]byte, string, error) {
	return rewriteOpenAIImagesMultipartModelAndN(body, contentType, model, 0, false)
}

func rewriteOpenAIImagesMultipartModelAndN(body []byte, contentType string, model string, n int, rewriteN bool) ([]byte, string, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, "", fmt.Errorf("parse multipart content-type: %w", err)
	}
	boundary := strings.TrimSpace(params["boundary"])
	if boundary == "" {
		return nil, "", fmt.Errorf("multipart boundary is required")
	}

	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	modelWritten := false
	nWritten := false
	backgroundWritten := false
	transparentBackgroundValue := ""
	transparentBackgroundSeen := false

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("read multipart body: %w", err)
		}

		formName := strings.TrimSpace(part.FormName())
		if part.FileName() == "" && formName == openAILocalGroupIDField {
			_ = part.Close()
			continue
		}
		if formName == "transparent_background" && part.FileName() == "" {
			data, readErr := io.ReadAll(io.LimitReader(part, openAIImageMaxUploadPartSize))
			_ = part.Close()
			if readErr != nil {
				return nil, "", fmt.Errorf("read multipart transparent_background: %w", readErr)
			}
			transparentBackgroundValue = strings.TrimSpace(string(data))
			transparentBackgroundSeen = true
			continue
		}

		if formName == "background" && part.FileName() == "" {
			data, readErr := io.ReadAll(io.LimitReader(part, openAIImageMaxUploadPartSize))
			_ = part.Close()
			if readErr != nil {
				return nil, "", fmt.Errorf("read multipart background: %w", readErr)
			}
			background := strings.TrimSpace(string(data))
			if background != "" {
				if err := writer.WriteField("background", background); err != nil {
					return nil, "", fmt.Errorf("rewrite multipart background: %w", err)
				}
				backgroundWritten = true
			}
			continue
		}

		partHeader := cloneMultipartHeader(part.Header)
		target, err := writer.CreatePart(partHeader)
		if err != nil {
			_ = part.Close()
			return nil, "", fmt.Errorf("create multipart part: %w", err)
		}
		if formName == "model" && part.FileName() == "" && model != "" {
			if _, err := target.Write([]byte(model)); err != nil {
				_ = part.Close()
				return nil, "", fmt.Errorf("rewrite multipart model: %w", err)
			}
			modelWritten = true
			_ = part.Close()
			continue
		}
		if formName == "n" && part.FileName() == "" && rewriteN {
			if _, err := target.Write([]byte(strconv.Itoa(n))); err != nil {
				_ = part.Close()
				return nil, "", fmt.Errorf("rewrite multipart n: %w", err)
			}
			nWritten = true
			_ = part.Close()
			continue
		}
		if _, err := io.Copy(target, part); err != nil {
			_ = part.Close()
			return nil, "", fmt.Errorf("copy multipart part: %w", err)
		}
		_ = part.Close()
	}

	if !modelWritten {
		if err := writer.WriteField("model", model); err != nil {
			return nil, "", fmt.Errorf("append multipart model field: %w", err)
		}
	}
	if rewriteN && !nWritten {
		if err := writer.WriteField("n", strconv.Itoa(n)); err != nil {
			return nil, "", fmt.Errorf("append multipart n field: %w", err)
		}
	}
	if transparentBackgroundSeen && !backgroundWritten {
		parsed, err := strconv.ParseBool(transparentBackgroundValue)
		if err != nil {
			return nil, "", fmt.Errorf("invalid transparent_background field value")
		}
		if err := writer.WriteField("background", openAIImagesBackgroundFromTransparent(parsed)); err != nil {
			return nil, "", fmt.Errorf("append multipart background field: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("finalize multipart body: %w", err)
	}
	return buffer.Bytes(), writer.FormDataContentType(), nil
}

func cloneMultipartHeader(src textproto.MIMEHeader) textproto.MIMEHeader {
	dst := make(textproto.MIMEHeader, len(src))
	for key, values := range src {
		copied := make([]string, len(values))
		copy(copied, values)
		dst[key] = copied
	}
	return dst
}

func (s *OpenAIGatewayService) handleOpenAIImagesNonStreamingResponse(resp *http.Response, c *gin.Context) (OpenAIUsage, int, []string, error) {
	body, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	if err != nil {
		return OpenAIUsage{}, 0, nil, err
	}
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := "application/json"
	if s.cfg != nil && !s.cfg.Security.ResponseHeaders.Enabled {
		if upstreamType := resp.Header.Get("Content-Type"); upstreamType != "" {
			contentType = upstreamType
		}
	}
	c.Data(resp.StatusCode, contentType, body)

	usage, _ := extractOpenAIUsageFromJSONBytes(body)
	return usage, extractOpenAIImageCountFromJSONBytes(body), collectOpenAIResponseImageOutputSizesFromJSONBytes(body), nil
}

func (s *OpenAIGatewayService) handleOpenAIImagesStreamingResponse(
	resp *http.Response,
	c *gin.Context,
	startTime time.Time,
) (OpenAIUsage, int, []string, *int, error) {
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "text/event-stream"
	}
	c.Status(resp.StatusCode)
	c.Header("Content-Type", contentType)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return OpenAIUsage{}, 0, nil, nil, fmt.Errorf("streaming is not supported by response writer")
	}

	usage := OpenAIUsage{}
	imageCounter := newOpenAIImageOutputCounter()
	var firstTokenMs *int
	clientDisconnected := false
	lastDownstreamWriteAt := time.Now()
	var fallbackBody bytes.Buffer
	fallbackBytes := int64(0)
	fallbackLimit := resolveUpstreamResponseReadLimit(s.cfg)
	seenSSEData := false
	fallbackTooLarge := false
	var sseData openAISSEDataAccumulator

	processSSEData := func(dataBytes []byte) {
		seenSSEData = true
		fallbackBody.Reset()
		fallbackBytes = 0
		mergeOpenAIUsage(&usage, dataBytes)
		imageCounter.AddSSEData(dataBytes)
	}

	flushSSEEvent := func() {
		sseData.Flush(processSSEData)
	}

	processLine := func(line []byte) {
		if len(line) == 0 {
			return
		}
		if firstTokenMs == nil {
			ms := int(time.Since(startTime).Milliseconds())
			firstTokenMs = &ms
		}
		if !clientDisconnected {
			if _, writeErr := c.Writer.Write(line); writeErr != nil {
				clientDisconnected = true
				logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Images stream client disconnected, continue draining upstream for billing")
			} else {
				flusher.Flush()
				lastDownstreamWriteAt = time.Now()
			}
		}

		trimmedLine := strings.TrimRight(string(line), "\r\n")
		if _, ok := extractOpenAISSEDataLine(trimmedLine); ok || strings.TrimSpace(trimmedLine) == "" {
			sseData.AddLine(trimmedLine, processSSEData)
			return
		}
		if !seenSSEData && !fallbackTooLarge {
			fallbackBytes += int64(len(line))
			if fallbackBytes <= fallbackLimit {
				_, _ = fallbackBody.Write(line)
			} else {
				fallbackTooLarge = true
				fallbackBody.Reset()
			}
		}
	}

	finalizeFallbackBody := func() {
		if seenSSEData || fallbackBody.Len() == 0 {
			return
		}
		body := bytes.TrimSpace(fallbackBody.Bytes())
		if len(body) == 0 {
			return
		}
		mergeOpenAIUsage(&usage, body)
		imageCounter.AddJSONResponse(body)
	}

	streamInterval := s.openAIImageStreamDataInterval()
	keepaliveInterval := s.openAIImageStreamKeepaliveInterval()
	if streamInterval <= 0 && keepaliveInterval <= 0 {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			processLine(line)
			if err == io.EOF {
				break
			}
			if err != nil {
				flushSSEEvent()
				return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, err
			}
		}
		flushSSEEvent()
		finalizeFallbackBody()
		return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, nil
	}

	type readEvent struct {
		line []byte
		err  error
	}
	events := make(chan readEvent, 16)
	done := make(chan struct{})
	sendEvent := func(ev readEvent) bool {
		select {
		case events <- ev:
			return true
		case <-done:
			return false
		}
	}
	var lastReadAt int64
	atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
	go func() {
		defer close(events)
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
			}
			if len(line) > 0 && !sendEvent(readEvent{line: line}) {
				return
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				_ = sendEvent(readEvent{err: err})
				return
			}
		}
	}()
	defer close(done)

	var intervalTicker *time.Ticker
	if streamInterval > 0 {
		intervalTicker = time.NewTicker(streamInterval)
		defer intervalTicker.Stop()
	}
	var intervalCh <-chan time.Time
	if intervalTicker != nil {
		intervalCh = intervalTicker.C
	}

	var keepaliveTicker *time.Ticker
	if keepaliveInterval > 0 {
		keepaliveTicker = time.NewTicker(keepaliveInterval)
		defer keepaliveTicker.Stop()
	}
	var keepaliveCh <-chan time.Time
	if keepaliveTicker != nil {
		keepaliveCh = keepaliveTicker.C
	}

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				flushSSEEvent()
				finalizeFallbackBody()
				return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, nil
			}
			if ev.err != nil {
				flushSSEEvent()
				return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, ev.err
			}
			processLine(ev.line)
		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			if clientDisconnected {
				return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, fmt.Errorf("image stream incomplete after timeout")
			}
			logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Images stream data interval timeout: interval=%s", streamInterval)
			_ = s.writeOpenAIImagesStreamEvent(c, flusher, "error", buildOpenAIImagesStreamErrorBody(fmt.Sprintf("upstream image stream idle for %s", streamInterval)))
			return usage, imageCounter.Count(), imageCounter.Sizes(), firstTokenMs, fmt.Errorf("image stream data interval timeout")
		case <-keepaliveCh:
			if clientDisconnected || time.Since(lastDownstreamWriteAt) < keepaliveInterval {
				continue
			}
			if _, writeErr := io.WriteString(c.Writer, ":\n\n"); writeErr != nil {
				clientDisconnected = true
				logger.LegacyPrintf("service.openai_gateway", "[OpenAI] Images stream client disconnected during keepalive, continue draining upstream for billing")
				continue
			}
			flusher.Flush()
			lastDownstreamWriteAt = time.Now()
		}
	}
}

func (s *OpenAIGatewayService) openAIImageStreamDataInterval() time.Duration {
	if s == nil || s.cfg == nil || s.cfg.Gateway.ImageStreamDataIntervalTimeout <= 0 {
		return 0
	}
	return time.Duration(s.cfg.Gateway.ImageStreamDataIntervalTimeout) * time.Second
}

func (s *OpenAIGatewayService) openAIImageStreamKeepaliveInterval() time.Duration {
	if s == nil || s.cfg == nil || s.cfg.Gateway.ImageStreamKeepaliveInterval <= 0 {
		return 0
	}
	return time.Duration(s.cfg.Gateway.ImageStreamKeepaliveInterval) * time.Second
}

func extractOpenAIImagesBillableCountFromJSONBytes(body []byte) int {
	if count := extractOpenAIImageCountFromJSONBytes(body); count > 0 {
		return count
	}
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return 0
	}
	if count := int(gjson.GetBytes(body, "usage.images").Int()); count > 0 {
		return count
	}
	if count := int(gjson.GetBytes(body, "tool_usage.image_gen.images").Int()); count > 0 {
		return count
	}
	eventType := strings.TrimSpace(gjson.GetBytes(body, "type").String())
	if eventType == "" || !strings.HasSuffix(eventType, ".completed") {
		return 0
	}
	if gjson.GetBytes(body, "b64_json").Exists() || gjson.GetBytes(body, "url").Exists() {
		return 1
	}
	return 0
}

func mergeOpenAIUsage(dst *OpenAIUsage, body []byte) {
	if dst == nil {
		return
	}
	if gjson.ValidBytes(body) && gjson.GetBytes(body, "type").String() == "response.completed" {
		if toolUsage, ok := openAIImagesToolUsageFromGJSON(gjson.GetBytes(body, "response.tool_usage.image_gen")); ok {
			*dst = toolUsage
			return
		}
	}
	if parsed, ok := extractOpenAIUsageFromJSONBytes(body); ok {
		if parsed.InputTokens > 0 {
			dst.InputTokens = parsed.InputTokens
		}
		if parsed.OutputTokens > 0 {
			dst.OutputTokens = parsed.OutputTokens
		}
		if parsed.CacheReadInputTokens > 0 {
			dst.CacheReadInputTokens = parsed.CacheReadInputTokens
		}
		if parsed.ImageInputTokens > 0 {
			dst.ImageInputTokens = parsed.ImageInputTokens
		}
		if parsed.ImageOutputTokens > 0 {
			dst.ImageOutputTokens = parsed.ImageOutputTokens
		}
	}
}

func extractOpenAIImageCountFromJSONBytes(body []byte) int {
	return countOpenAIResponseImageOutputsFromJSONBytes(body)
}

type openAIImagePointerInfo struct {
	Pointer     string
	DownloadURL string
	B64JSON     string
	MimeType    string
	Prompt      string
}

func collectOpenAIImagePointers(body []byte) []openAIImagePointerInfo {
	if len(body) == 0 {
		return nil
	}
	prompt := ""
	for _, path := range []string{
		"message.metadata.dalle.prompt",
		"metadata.dalle.prompt",
		"revised_prompt",
	} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			prompt = value
			break
		}
	}
	matches := openAIImagePointerMatches(body)
	out := make([]openAIImagePointerInfo, 0, len(matches))
	for _, pointer := range matches {
		out = append(out, openAIImagePointerInfo{Pointer: pointer, Prompt: prompt})
	}
	return mergeOpenAIImagePointerInfos(out, collectOpenAIImageInlineAssets(body, prompt))
}

func openAIImagePointerMatches(body []byte) []string {
	raw := string(body)
	matches := make([]string, 0, 4)
	for _, prefix := range []string{"file-service://", "sediment://"} {
		start := 0
		for {
			idx := strings.Index(raw[start:], prefix)
			if idx < 0 {
				break
			}
			idx += start
			end := idx + len(prefix)
			for end < len(raw) {
				ch := raw[end]
				if ch != '-' && ch != '_' &&
					(ch < '0' || ch > '9') &&
					(ch < 'a' || ch > 'z') &&
					(ch < 'A' || ch > 'Z') {
					break
				}
				end++
			}
			matches = append(matches, raw[idx:end])
			start = end
		}
	}
	return dedupeStrings(matches)
}

func mergeOpenAIImagePointerInfos(existing []openAIImagePointerInfo, next []openAIImagePointerInfo) []openAIImagePointerInfo {
	if len(next) == 0 {
		return existing
	}
	seen := make(map[string]openAIImagePointerInfo, len(existing)+len(next))
	out := make([]openAIImagePointerInfo, 0, len(existing)+len(next))
	for _, item := range existing {
		if key := item.identityKey(); key != "" {
			seen[key] = item
		}
		out = append(out, item)
	}
	for _, item := range next {
		key := item.identityKey()
		if key == "" {
			continue
		}
		if existingItem, ok := seen[key]; ok {
			merged := mergeOpenAIImagePointerInfo(existingItem, item)
			if merged != existingItem {
				for i := range out {
					if out[i].identityKey() == key {
						out[i] = merged
						break
					}
				}
				seen[key] = merged
			}
			continue
		}
		seen[key] = item
		out = append(out, item)
	}
	return out
}

func (i openAIImagePointerInfo) identityKey() string {
	switch {
	case strings.TrimSpace(i.Pointer) != "":
		return "pointer:" + strings.TrimSpace(i.Pointer)
	case strings.TrimSpace(i.DownloadURL) != "":
		return "download:" + strings.TrimSpace(i.DownloadURL)
	case strings.TrimSpace(i.B64JSON) != "":
		b64 := strings.TrimSpace(i.B64JSON)
		if len(b64) > 64 {
			b64 = b64[:64]
		}
		return "b64:" + b64
	default:
		return ""
	}
}

func mergeOpenAIImagePointerInfo(existing, next openAIImagePointerInfo) openAIImagePointerInfo {
	merged := existing
	if strings.TrimSpace(merged.Pointer) == "" {
		merged.Pointer = next.Pointer
	}
	if strings.TrimSpace(merged.DownloadURL) == "" {
		merged.DownloadURL = next.DownloadURL
	}
	if strings.TrimSpace(merged.B64JSON) == "" {
		merged.B64JSON = next.B64JSON
	}
	if strings.TrimSpace(merged.MimeType) == "" {
		merged.MimeType = next.MimeType
	}
	if strings.TrimSpace(merged.Prompt) == "" {
		merged.Prompt = next.Prompt
	}
	return merged
}

func resolveOpenAIImageBytes(
	ctx context.Context,
	client *req.Client,
	headers http.Header,
	conversationID string,
	pointer openAIImagePointerInfo,
) ([]byte, error) {
	if normalized := normalizeOpenAIImageBase64(pointer.B64JSON); normalized != "" {
		return base64.StdEncoding.DecodeString(normalized)
	}
	if downloadURL := strings.TrimSpace(pointer.DownloadURL); downloadURL != "" {
		return downloadOpenAIImageBytes(ctx, client, headers, downloadURL)
	}
	if strings.TrimSpace(pointer.Pointer) == "" {
		return nil, fmt.Errorf("image asset is missing pointer, url, and base64 data")
	}
	downloadURL, err := fetchOpenAIImageDownloadURL(ctx, client, headers, conversationID, pointer.Pointer)
	if err != nil {
		return nil, err
	}
	return downloadOpenAIImageBytes(ctx, client, headers, downloadURL)
}

func normalizeOpenAIImageBase64(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(raw), "data:") {
		if idx := strings.Index(raw, ","); idx >= 0 && idx+1 < len(raw) {
			raw = raw[idx+1:]
		}
	}
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "=") + strings.Repeat("=", (4-len(raw)%4)%4)
	if raw == "" {
		return ""
	}
	if _, err := base64.StdEncoding.DecodeString(raw); err != nil {
		return ""
	}
	return raw
}

func collectOpenAIImageInlineAssets(body []byte, fallbackPrompt string) []openAIImagePointerInfo {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil
	}
	var out []openAIImagePointerInfo
	walkOpenAIImageInlineAssets(decoded, strings.TrimSpace(fallbackPrompt), &out)
	return out
}

func walkOpenAIImageInlineAssets(node any, prompt string, out *[]openAIImagePointerInfo) {
	switch value := node.(type) {
	case map[string]any:
		localPrompt := prompt
		for _, key := range []string{"revised_prompt", "image_gen_title", "prompt"} {
			if v, ok := value[key].(string); ok && strings.TrimSpace(v) != "" {
				localPrompt = strings.TrimSpace(v)
				break
			}
		}
		item := openAIImagePointerInfo{
			Prompt:      localPrompt,
			Pointer:     firstNonEmptyString(value["asset_pointer"], value["pointer"]),
			DownloadURL: firstNonEmptyString(value["download_url"], value["url"], value["image_url"]),
			B64JSON:     firstNonEmptyString(value["b64_json"], value["base64"], value["image_base64"]),
			MimeType:    firstNonEmptyString(value["mime_type"], value["mimeType"], value["content_type"]),
		}
		switch {
		case strings.HasPrefix(strings.TrimSpace(item.Pointer), "file-service://"),
			strings.HasPrefix(strings.TrimSpace(item.Pointer), "sediment://"),
			isLikelyOpenAIImageDownloadURL(item.DownloadURL),
			normalizeOpenAIImageBase64(item.B64JSON) != "":
			*out = append(*out, item)
		}
		for _, child := range value {
			walkOpenAIImageInlineAssets(child, localPrompt, out)
		}
	case []any:
		for _, child := range value {
			walkOpenAIImageInlineAssets(child, prompt, out)
		}
	}
}

func firstNonEmptyString(values ...any) string {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func isLikelyOpenAIImageDownloadURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	if strings.HasPrefix(strings.ToLower(raw), "data:image/") {
		return true
	}
	if !strings.HasPrefix(strings.ToLower(raw), "http://") && !strings.HasPrefix(strings.ToLower(raw), "https://") {
		return false
	}
	lower := strings.ToLower(raw)
	return strings.Contains(lower, "/download") ||
		strings.Contains(lower, ".png") ||
		strings.Contains(lower, ".jpg") ||
		strings.Contains(lower, ".jpeg") ||
		strings.Contains(lower, ".webp")
}

func fetchOpenAIImageDownloadURL(
	ctx context.Context,
	client *req.Client,
	headers http.Header,
	conversationID string,
	pointer string,
) (string, error) {
	url := ""
	allowConversationRetry := false
	switch {
	case strings.HasPrefix(pointer, "file-service://"):
		fileID := strings.TrimPrefix(pointer, "file-service://")
		url = fmt.Sprintf("%s/%s/download", openAIChatGPTFilesURL, fileID)
	case strings.HasPrefix(pointer, "sediment://"):
		attachmentID := strings.TrimPrefix(pointer, "sediment://")
		url = fmt.Sprintf("https://chatgpt.com/backend-api/conversation/%s/attachment/%s/download", conversationID, attachmentID)
		allowConversationRetry = true
	default:
		return "", fmt.Errorf("unsupported image pointer: %s", pointer)
	}

	var lastErr error
	for attempt := 0; attempt < 8; attempt++ {
		var result struct {
			DownloadURL string `json:"download_url"`
		}
		resp, err := client.R().
			SetContext(ctx).
			SetHeaders(headerToMap(headers)).
			SetSuccessResult(&result).
			Get(url)
		if err != nil {
			lastErr = err
		} else if resp.IsSuccessState() && strings.TrimSpace(result.DownloadURL) != "" {
			return strings.TrimSpace(result.DownloadURL), nil
		} else {
			statusErr := newOpenAIImageStatusError(resp, "fetch image download url failed")
			if !allowConversationRetry || !isOpenAIImageTransientConversationNotFoundError(statusErr) {
				return "", statusErr
			}
			lastErr = statusErr
		}
		if attempt == 7 {
			break
		}
		timer := time.NewTimer(750 * time.Millisecond)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return "", ctx.Err()
		case <-timer.C:
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("fetch image download url failed")
	}
	return "", lastErr
}

func downloadOpenAIImageBytes(ctx context.Context, client *req.Client, headers http.Header, downloadURL string) ([]byte, error) {
	request := client.R().
		SetContext(ctx).
		DisableAutoReadResponse()

	if strings.HasPrefix(downloadURL, openAIChatGPTStartURL) {
		downloadHeaders := cloneHTTPHeader(headers)
		downloadHeaders.Set("Accept", "image/*,*/*;q=0.8")
		downloadHeaders.Del("Content-Type")
		request.SetHeaders(headerToMap(downloadHeaders))
	} else {
		userAgent := strings.TrimSpace(headers.Get("User-Agent"))
		if userAgent == "" {
			userAgent = openAIImageBackendUserAgent
		}
		request.SetHeader("User-Agent", userAgent)
	}

	resp, err := request.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, newOpenAIImageStatusError(resp, "download image bytes failed")
	}
	return io.ReadAll(io.LimitReader(resp.Body, openAIImageMaxDownloadBytes))
}

type openAIImageStatusError struct {
	StatusCode      int
	Message         string
	ResponseBody    []byte
	ResponseHeaders http.Header
	RequestID       string
	URL             string
}

func (e *openAIImageStatusError) Error() string {
	if e == nil {
		return "openai image backend request failed"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("openai image backend request failed: status %d", e.StatusCode)
	}
	return "openai image backend request failed"
}

func newOpenAIImageStatusError(resp *req.Response, fallback string) error {
	if resp == nil {
		if strings.TrimSpace(fallback) == "" {
			fallback = "openai image backend request failed"
		}
		return fmt.Errorf("%s", fallback)
	}

	statusCode := resp.StatusCode
	headers := http.Header(nil)
	requestID := ""
	requestURL := ""
	body := []byte(nil)

	if resp.Response != nil {
		headers = resp.Header.Clone()
		requestID = strings.TrimSpace(resp.Header.Get("x-request-id"))
		if resp.Request != nil && resp.Request.URL != nil {
			requestURL = resp.Request.URL.String()
		}
		if resp.Body != nil {
			body, _ = io.ReadAll(io.LimitReader(resp.Body, 2<<20))
			_ = resp.Body.Close()
		}
	}

	message := sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(body))
	if message == "" {
		prefix := strings.TrimSpace(fallback)
		if prefix == "" {
			prefix = "openai image backend request failed"
		}
		message = fmt.Sprintf("%s: status %d", prefix, statusCode)
	}

	return &openAIImageStatusError{
		StatusCode:      statusCode,
		Message:         message,
		ResponseBody:    body,
		ResponseHeaders: headers,
		RequestID:       requestID,
		URL:             requestURL,
	}
}

func isOpenAIImageTransientConversationNotFoundError(err error) bool {
	statusErr, ok := err.(*openAIImageStatusError)
	if !ok || statusErr == nil || statusErr.StatusCode != http.StatusNotFound {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(statusErr.Message))
	if strings.Contains(msg, "conversation_not_found") {
		return true
	}
	if strings.Contains(msg, "conversation") && strings.Contains(msg, "not found") {
		return true
	}
	bodyMsg := strings.ToLower(strings.TrimSpace(extractUpstreamErrorMessage(statusErr.ResponseBody)))
	if strings.Contains(bodyMsg, "conversation_not_found") {
		return true
	}
	return strings.Contains(bodyMsg, "conversation") && strings.Contains(bodyMsg, "not found")
}

func cloneHTTPHeader(src http.Header) http.Header {
	dst := make(http.Header, len(src))
	for key, values := range src {
		copied := make([]string, len(values))
		copy(copied, values)
		dst[key] = copied
	}
	return dst
}

func headerToMap(header http.Header) map[string]string {
	if len(header) == 0 {
		return nil
	}
	result := make(map[string]string, len(header))
	for key, values := range header {
		if len(values) == 0 {
			continue
		}
		result[key] = values[0]
	}
	return result
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

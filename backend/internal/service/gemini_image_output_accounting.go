package service

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const geminiImageOutputCounterKey = "gemini_image_output_counter"

// geminiImageOutputCounter keeps the largest image count seen in one payload.
// Gemini streams may repeat cumulative content, so summing chunks can overbill.
type geminiImageOutputCounter struct {
	count int
}

func beginGeminiImageOutputObservation(c *gin.Context) *geminiImageOutputCounter {
	if c == nil {
		return nil
	}
	counter := &geminiImageOutputCounter{}
	c.Set(geminiImageOutputCounterKey, counter)
	return counter
}

func geminiImageOutputCounterFromContext(c *gin.Context) *geminiImageOutputCounter {
	if c == nil {
		return nil
	}
	value, ok := c.Get(geminiImageOutputCounterKey)
	if !ok {
		return nil
	}
	counter, _ := value.(*geminiImageOutputCounter)
	return counter
}

func observeGeminiImageOutputs(c *gin.Context, payload []byte) {
	counter := geminiImageOutputCounterFromContext(c)
	if counter == nil {
		return
	}
	if count := countGeminiInlineImageOutputs(payload); count > counter.count {
		counter.count = count
	}
}

func observedGeminiImageOutputs(c *gin.Context) int {
	counter := geminiImageOutputCounterFromContext(c)
	if counter == nil {
		return 0
	}
	return counter.count
}

func resolveGeminiImageCount(c *gin.Context, originalModel, mappedModel string) int {
	if observed := observedGeminiImageOutputs(c); observed > 0 {
		return observed
	}
	if isImageGenerationModel(originalModel) || isImageGenerationModel(mappedModel) {
		return 1
	}
	return 0
}

func countGeminiInlineImageOutputs(payload []byte) int {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return 0
	}

	count := 0
	gjson.GetBytes(payload, "candidates").ForEach(func(_, candidate gjson.Result) bool {
		candidate.Get("content.parts").ForEach(func(_, part gjson.Result) bool {
			if geminiPartIsInlineImage(part) {
				count++
			}
			return true
		})
		return true
	})
	return count
}

func geminiPartIsInlineImage(part gjson.Result) bool {
	for _, key := range []string{"inlineData", "inline_data", "fileData", "file_data"} {
		image := part.Get(key)
		mimeType := image.Get("mimeType")
		if !mimeType.Exists() {
			mimeType = image.Get("mime_type")
		}
		if !isGeminiInlineImageMIMEType(strings.ToLower(strings.TrimSpace(mimeType.String()))) {
			continue
		}
		data := image.Get("data")
		if key == "fileData" || key == "file_data" {
			data = image.Get("fileUri")
			if !data.Exists() {
				data = image.Get("file_uri")
			}
		}
		// A part may expose both forms; count it only once and require actual data.
		if data.Type == gjson.String && strings.TrimSpace(data.String()) != "" {
			return true
		}
	}
	return false
}

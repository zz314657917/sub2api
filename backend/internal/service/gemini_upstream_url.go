package service

import (
	"errors"
	"fmt"
	"strings"
)

var geminiAIStudioActions = map[string]struct{}{
	"generateContent":       {},
	"streamGenerateContent": {},
	"countTokens":           {},
}

func buildGeminiAIStudioModelActionURL(baseURL, model, action string, stream bool) (string, error) {
	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBase == "" {
		return "", errors.New("gemini base url is required")
	}
	if model == "" {
		return "", errors.New("gemini model is required")
	}
	if err := validateUpstreamPathSegment("gemini model", model); err != nil {
		return "", err
	}
	trimmedAction := strings.TrimSpace(action)
	if _, ok := geminiAIStudioActions[trimmedAction]; !ok {
		return "", fmt.Errorf("unsupported gemini action: %s", trimmedAction)
	}

	fullURL := fmt.Sprintf("%s/v1beta/models/%s:%s", trimmedBase, model, trimmedAction)
	if stream {
		fullURL += "?alt=sse"
	}
	return fullURL, nil
}

func IsSafeGeminiModelPathSegment(model string) bool {
	return isSafeUpstreamPathSegment(model)
}

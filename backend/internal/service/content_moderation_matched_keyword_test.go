package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContentModerationKeywordBlockCarriesMatchedKeyword(t *testing.T) {
	keyword := "secret-token"
	cfg := defaultContentModerationConfig()
	input := ContentModerationCheckInput{UserID: 42, Endpoint: "/v1/chat/completions", Model: "test"}
	log := (&ContentModerationService{}).buildLog(input, cfg, ContentModerationActionKeywordBlock, true, contentModerationKeywordCategory, 1, map[string]float64{contentModerationKeywordCategory: 1}, "redacted", nil, nil, "")
	log.MatchedKeyword = keyword

	require.Equal(t, keyword, log.MatchedKeyword)
	require.Equal(t, ContentModerationActionKeywordBlock, log.Action)
}

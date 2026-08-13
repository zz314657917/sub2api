package service

import (
	"strings"
	"testing"
)

func TestChannelPricingCacheConflictClaudeExactDotHyphenAndWhitespace(t *testing.T) {
	err := validateChannelConfig([]ChannelModelPricing{
		{Platform: PlatformAnthropic, Models: []string{" Claude-Opus-4.6 "}},
		{Platform: PlatformAnthropic, Models: []string{"claude-opus-4-6"}},
	}, nil)

	if err == nil || !strings.Contains(err.Error(), "MODEL_PATTERN_CONFLICT") {
		t.Fatalf("expected normalized Claude exact-model conflict, got %v", err)
	}
}

func TestChannelPricingCacheConflictClaudeWildcardDotHyphen(t *testing.T) {
	err := validateChannelConfig([]ChannelModelPricing{
		{Platform: PlatformAnthropic, Models: []string{"claude-opus-4.*"}},
		{Platform: PlatformAnthropic, Models: []string{"claude-opus-4-6"}},
	}, nil)

	if err == nil || !strings.Contains(err.Error(), "MODEL_PATTERN_CONFLICT") {
		t.Fatalf("expected normalized Claude wildcard conflict, got %v", err)
	}
}

func TestChannelPricingCacheConflictNonClaudeDotHyphenCanCoexist(t *testing.T) {
	err := validateChannelConfig([]ChannelModelPricing{
		{Platform: PlatformOpenAI, Models: []string{"gpt-4.1"}},
		{Platform: PlatformOpenAI, Models: []string{"gpt-4-1"}},
	}, nil)

	if err != nil {
		t.Fatalf("expected non-Claude dot/hyphen models to coexist, got %v", err)
	}
}

func TestChannelPricingCacheConflictMappingKeepsLowerOnlySemantics(t *testing.T) {
	err := validateChannelConfig(nil, map[string]map[string]string{
		PlatformAnthropic: {
			"claude-opus-4.6": "first",
			"claude-opus-4-6": "second",
		},
	})

	if err != nil {
		t.Fatalf("expected mapping dot/hyphen sources to coexist, got %v", err)
	}
}

func TestChannelPricingCacheConflictAccountStatsKeepsLowerOnlySemantics(t *testing.T) {
	err := validatePricingEntries([]ChannelModelPricing{
		{Platform: PlatformAnthropic, Models: []string{"claude-opus-4.6"}},
		{Platform: PlatformAnthropic, Models: []string{"claude-opus-4-6"}},
	})

	if err != nil {
		t.Fatalf("expected account stats pricing dot/hyphen models to coexist, got %v", err)
	}
}

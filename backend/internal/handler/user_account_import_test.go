package handler

import "testing"

func TestMergeUserImportExtraPreservesExportMetadataAndInfersTier(t *testing.T) {
	extra := mergeUserImportExtra(
		map[string]any{
			"email": "owner@example.com",
		},
		map[string]any{
			"access_token": "redacted",
			"id_token":     "redacted",
		},
		"openai",
		"owner@example.com-plus",
	)

	if extra["email"] != "owner@example.com" {
		t.Fatalf("expected import email to be preserved, got %#v", extra)
	}
	if extra["share_display_tier"] != "plus" {
		t.Fatalf("expected plus display tier inferred from name, got %#v", extra)
	}
	if extra["share_display_percent_only"] != true {
		t.Fatalf("expected percent-only display default, got %#v", extra)
	}
}

func TestMergeUserImportExtraDoesNotOverrideExplicitDisplayTier(t *testing.T) {
	extra := mergeUserImportExtra(
		map[string]any{
			"share_display_tier":         "pro",
			"share_display_percent_only": false,
		},
		map[string]any{
			"email": "owner@example.com",
		},
		"openai",
		"owner@example.com-plus",
	)

	if extra["share_display_tier"] != "pro" {
		t.Fatalf("expected explicit tier to win, got %#v", extra)
	}
	if extra["share_display_percent_only"] != false {
		t.Fatalf("expected explicit percent-only flag to win, got %#v", extra)
	}
	if extra["email"] != "owner@example.com" {
		t.Fatalf("expected email to be copied from credentials, got %#v", extra)
	}
}

func TestMergeUserImportExtraOnlyInfersTierForOpenAI(t *testing.T) {
	extra := mergeUserImportExtra(
		nil,
		nil,
		"claude",
		"owner@example.com-plus",
	)

	if extra != nil {
		t.Fatalf("expected non-openai import to avoid share display tier inference, got %#v", extra)
	}
}

func TestMergeUserImportExtraDoesNotInferTierFromPartialWord(t *testing.T) {
	extra := mergeUserImportExtra(
		nil,
		nil,
		"openai",
		"owner-profile",
	)

	if extra != nil {
		t.Fatalf("expected partial word to avoid tier inference, got %#v", extra)
	}
}

func TestDefaultUserAccountNameUsesImportExtraEmail(t *testing.T) {
	name := defaultUserAccountName(
		"openai",
		"oauth",
		map[string]any{
			"access_token": "redacted",
		},
		map[string]any{
			"email": "owner@example.com",
		},
	)

	if name != "owner@example.com" {
		t.Fatalf("expected extra email to be used as import name, got %q", name)
	}
}

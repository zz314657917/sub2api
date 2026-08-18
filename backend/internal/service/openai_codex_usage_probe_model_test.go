package service

import (
	"testing"

	openaipkg "github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

func TestCodexUsageProbeModel(t *testing.T) {
	if openaipkg.CodexUsageProbeModel != "codex-auto-review" {
		t.Fatalf("CodexUsageProbeModel = %q, want %q", openaipkg.CodexUsageProbeModel, "codex-auto-review")
	}
	if openaipkg.DefaultTestModel == openaipkg.CodexUsageProbeModel {
		t.Fatalf("CodexUsageProbeModel must remain distinct from DefaultTestModel")
	}
}

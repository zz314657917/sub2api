package service

import "testing"

func TestBuildStatusSummaryIncludesExtraModelAvailability(t *testing.T) {
	latestByModel := map[string]*ChannelMonitorLatest{
		"gpt": {
			Model:  "gpt",
			Status: MonitorStatusOperational,
		},
		"gpt-image": {
			Model:  "gpt-image",
			Status: MonitorStatusDegraded,
		},
	}
	availByModel := map[string]*ChannelMonitorAvailability{
		"gpt": {
			Model:           "gpt",
			AvailabilityPct: 90.94,
		},
		"gpt-image": {
			Model:           "gpt-image",
			AvailabilityPct: 97.66,
		},
	}

	summary := buildStatusSummary(latestByModel, availByModel, "gpt", []string{"gpt-image"})

	if summary.Availability7d != 90.94 {
		t.Fatalf("primary availability = %v, want 90.94", summary.Availability7d)
	}
	if len(summary.ExtraModels) != 1 {
		t.Fatalf("extra models len = %d, want 1", len(summary.ExtraModels))
	}
	extra := summary.ExtraModels[0]
	if extra.Model != "gpt-image" {
		t.Fatalf("extra model = %q, want gpt-image", extra.Model)
	}
	if extra.Status != MonitorStatusDegraded {
		t.Fatalf("extra status = %q, want %q", extra.Status, MonitorStatusDegraded)
	}
	if extra.Availability7d != 97.66 {
		t.Fatalf("extra availability = %v, want 97.66", extra.Availability7d)
	}
}

package securityaudit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldStorePromptAuditEvent(t *testing.T) {
	tests := []struct {
		name            string
		storePassEvents bool
		decision        EventDecision
		want            bool
	}{
		{name: "pass disabled", storePassEvents: false, decision: EventPass, want: false},
		{name: "flag disabled", storePassEvents: false, decision: EventFlag, want: true},
		{name: "critical disabled", storePassEvents: false, decision: EventCritical, want: true},
		{name: "pass enabled", storePassEvents: true, decision: EventPass, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldStorePromptAuditEvent(tt.decision, tt.storePassEvents); got != tt.want {
				t.Fatalf("shouldStorePromptAuditEvent(%q, %t) = %t, want %t", tt.decision, tt.storePassEvents, got, tt.want)
			}
		})
	}
}

func TestFullPromptForStorageOnlyRetainsCriticalRisk(t *testing.T) {
	snapshot := PromptSnapshot{FullPrompt: "raw\x00critical prompt", ScanText: "transient scan text"}
	require.Equal(t, "rawcritical prompt", fullPromptForStorage(snapshot, &NormalizedResult{RiskLevel: RiskCritical}))
	require.Empty(t, fullPromptForStorage(snapshot, &NormalizedResult{RiskLevel: RiskHigh}))
	require.Empty(t, fullPromptForStorage(snapshot, nil))
	persisted := snapshotForEventStorage(snapshot, &NormalizedResult{RiskLevel: RiskCritical})
	require.Equal(t, "rawcritical prompt", persisted.FullPrompt)
	require.Empty(t, persisted.ScanText)
}

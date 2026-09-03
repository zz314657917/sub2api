package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeOpsUpstreamProxyAttribution(t *testing.T) {
	id := int64(7)
	ev := &OpsUpstreamErrorEvent{ProxyID: &id}
	normalizeOpsUpstreamProxyAttribution(ev)
	require.Equal(t, opsProxyNameUnnamed, ev.ProxyName)

	ev = &OpsUpstreamErrorEvent{ProxyName: "legacy-name"}
	normalizeOpsUpstreamProxyAttribution(ev)
	require.Nil(t, ev.ProxyID)
	require.Equal(t, opsProxyNameUnknown, ev.ProxyName)

	ev = &OpsUpstreamErrorEvent{ProxyName: opsProxyNameDirect}
	normalizeOpsUpstreamProxyAttribution(ev)
	require.Equal(t, opsProxyNameDirect, ev.ProxyName)
}

func TestNormalizeOpsUpstreamErrorsJSONLegacy(t *testing.T) {
	raw := `[ {"message":"old","proxy_name":"old-proxy"}, {"message":"direct","proxy_name":"direct/no_proxy"} ]`
	normalized, err := normalizeOpsUpstreamErrorsJSON(raw)
	require.NoError(t, err)
	var events []map[string]any
	require.NoError(t, json.Unmarshal([]byte(normalized), &events))
	require.Equal(t, "unknown", events[0]["proxy_name"])
	require.Nil(t, events[0]["proxy_id"])
	require.Equal(t, "direct/no_proxy", events[1]["proxy_name"])
}

func TestBoundOpsUpstreamErrorsKeepsNewestAndStampsDropped(t *testing.T) {
	events := make([]*OpsUpstreamErrorEvent, 0, opsUpstreamErrorsMaxEvents+5)
	for i := 0; i < opsUpstreamErrorsMaxEvents+5; i++ {
		events = append(events, &OpsUpstreamErrorEvent{Message: strings.Repeat("x", 10), UpstreamStatusCode: 500})
	}
	kept, dropped := boundOpsUpstreamErrors(events)
	require.Len(t, kept, opsUpstreamErrorsMaxEvents)
	require.Equal(t, 5, dropped)
	require.Equal(t, 5, kept[0].DroppedEarlierAttempts)
}

func TestSanitizeOpsUpstreamErrorsBodyWindow(t *testing.T) {
	events := make([]*OpsUpstreamErrorEvent, 0, opsUpstreamErrorsBodyWindow+1)
	for i := 0; i < opsUpstreamErrorsBodyWindow+1; i++ {
		events = append(events, &OpsUpstreamErrorEvent{Message: "m", Detail: "detail", UpstreamResponseBody: "body", UpstreamStatusCode: 500})
	}
	entry := &OpsInsertErrorLogInput{UpstreamErrors: events}
	require.NoError(t, sanitizeOpsUpstreamErrors(entry))
	require.NotNil(t, entry.UpstreamErrorsJSON)
	var out []*OpsUpstreamErrorEvent
	require.NoError(t, json.Unmarshal([]byte(*entry.UpstreamErrorsJSON), &out))
	require.Len(t, out, opsUpstreamErrorsBodyWindow+1)
	require.Empty(t, out[0].Detail)
	require.Empty(t, out[0].UpstreamResponseBody)
	require.Equal(t, "detail", out[len(out)-1].Detail)
}

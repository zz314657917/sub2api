package admin

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorAdminQuotaDTOContract(t *testing.T) {
	id := int64(81)
	snapshot := &domain.MonitorQuotaSnapshot{Source: "usage", Success: true, FetchedAt: time.Now()}
	monitor := &service.ChannelMonitor{ID: 3, APIKey: "local-secret", AccountID: &id,
		Provider: "openai", APIMode: "responses", CheckMode: "quota_probe", JitterSeconds: 5}
	encoded, err := json.Marshal(buildListItemResponse(monitor, service.MonitorStatusSummary{LatestQuota: snapshot}))
	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, json.Unmarshal(encoded, &result))
	require.Equal(t, "responses", result["api_mode"])
	require.Equal(t, "quota_probe", result["check_mode"])
	require.Equal(t, float64(81), result["account_id"])
	require.Equal(t, float64(5), result["jitter_seconds"])
	require.Contains(t, result, "latest_quota")
	require.NotContains(t, string(encoded), "local-secret")
	check, err := json.Marshal(checkResultToResponse(&service.CheckResult{Quota: snapshot}))
	require.NoError(t, err)
	require.Contains(t, string(check), `"quota":`)
	history, err := json.Marshal(historyEntryToResponse(&service.ChannelMonitorHistoryEntry{Quota: snapshot}))
	require.NoError(t, err)
	require.Contains(t, string(history), `"quota":`)
}

func TestChannelMonitorAdminJSONBindingAcceptsQuotaWithoutProbeSecrets(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/channel-monitors",
		strings.NewReader(`{"name":"quota","provider":"antigravity","check_mode":"quota","account_id":12,"interval_seconds":60,"jitter_seconds":4}`))
	c.Request.Header.Set("Content-Type", "application/json")
	var create channelMonitorCreateRequest
	require.NoError(t, c.ShouldBindJSON(&create))
	require.Empty(t, create.APIKey)
	require.Empty(t, create.Endpoint)
	require.Equal(t, "quota", create.CheckMode)
	require.Equal(t, int64(12), *create.AccountID)
	require.Equal(t, 4, create.JitterSeconds)
	c.Request = httptest.NewRequest("POST", "/templates",
		strings.NewReader(`{"name":"Kimi","provider":"kimi","api_mode":"chat_completions"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	var template channelMonitorTemplateCreateRequest
	require.NoError(t, c.ShouldBindJSON(&template))
	require.Equal(t, "kimi", template.Provider)
}

func TestChannelMonitorUpdateRequestAccountOmissionAndExplicitUnbind(t *testing.T) {
	var omitted, unbind channelMonitorUpdateRequest
	require.NoError(t, json.Unmarshal([]byte(`{"name":"renamed"}`), &omitted))
	require.Nil(t, omitted.AccountID)
	require.NoError(t, json.Unmarshal([]byte(`{"account_id":0,"check_mode":"probe"}`), &unbind))
	require.NotNil(t, unbind.AccountID)
	require.Zero(t, *unbind.AccountID)
}

package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorQuotaPublicDTOProjectionStaysPrivate(t *testing.T) {
	view := &service.UserMonitorView{
		ID: 1, Name: "quota channel", Provider: "kimi", PrimaryModel: "quota",
		PrimaryStatus: "operational", ExtraModels: []service.ExtraModelStatus{{Model: "a", Availability7d: 98}},
	}
	data, err := json.Marshal(userMonitorViewToItem(view))
	require.NoError(t, err)
	for _, secret := range []string{"account_id", "latest_quota", "balance", "credential", "api_key", "endpoint"} {
		require.NotContains(t, strings.ToLower(string(data)), secret)
	}
	require.Contains(t, string(data), `"availability_7d":98`)
	detail, err := json.Marshal(userMonitorDetailToResponse(&service.UserMonitorDetail{ID: 1}))
	require.NoError(t, err)
	require.NotContains(t, string(detail), "quota")
}

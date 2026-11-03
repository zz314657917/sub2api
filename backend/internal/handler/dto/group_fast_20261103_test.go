package dto

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGroupFast20261103DTO(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		group := &service.Group{ID: 7, Platform: service.PlatformOpenAI, ForceOpenAIFast: enabled}
		for _, dto := range []any{GroupFromService(group), GroupFromServiceShallow(group), GroupFromServiceAdmin(group)} {
			raw, err := json.Marshal(dto)
			require.NoError(t, err)
			require.True(t, gjson.GetBytes(raw, "force_openai_fast").Exists())
			require.Equal(t, enabled, gjson.GetBytes(raw, "force_openai_fast").Bool())
		}
	}
}

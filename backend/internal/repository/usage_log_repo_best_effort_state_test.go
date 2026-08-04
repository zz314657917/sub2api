package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildUsageLogBestEffortInsertStateQueryCastsSyntheticInputIndex(t *testing.T) {
	prepared := []usageLogInsertPrepared{
		prepareUsageLogInsert(&service.UsageLog{
			UserID:    17,
			APIKeyID:  3,
			RequestID: "s175-state-first",
			CreatedAt: time.Date(2026, 8, 4, 4, 0, 0, 0, time.UTC),
		}),
		prepareUsageLogInsert(&service.UsageLog{
			UserID:    18,
			APIKeyID:  4,
			RequestID: "s175-state-second",
			CreatedAt: time.Date(2026, 8, 4, 4, 1, 0, 0, time.UTC),
		}),
	}

	stateQuery, stateArgs := buildUsageLogBestEffortInsertStateQuery(prepared)
	require.Contains(t, stateQuery, "$1::integer")
	require.Contains(t, stateQuery, fmt.Sprintf("$%d::integer", len(prepared[0].args)+2))
	require.Equal(t, 0, stateArgs[0])
	require.Equal(t, 1, stateArgs[len(prepared[0].args)+1])

	nonStateQuery, nonStateArgs := buildUsageLogBestEffortInsertQuery(prepared)
	require.NotContains(t, nonStateQuery, "input_idx")
	require.NotContains(t, nonStateQuery, "$1::integer")
	require.Len(t, stateArgs, len(nonStateArgs)+len(prepared))
}

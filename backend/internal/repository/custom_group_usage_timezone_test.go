package repository

import (
	"testing"

	appTimezone "github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
)

func TestGroupUsageRepositoryTimezoneTestRestoresPreviousTimezone(t *testing.T) {
	previous := appTimezone.Location().String()
	require.NoError(t, appTimezone.Init("America/New_York"))
	t.Cleanup(func() { require.NoError(t, appTimezone.Init(previous)) })
	require.Equal(t, "America/New_York", appTimezone.Location().String())
}

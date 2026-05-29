package repository

import (
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildImageCreatorImageListWhereUsesCurrentUserAndFilters(t *testing.T) {
	where, args := buildImageCreatorImageListWhere(42, service.ImageCreatorImageListFilters{
		Search:      "snow",
		StartDate:   "2026-05-01",
		EndDate:     "2026-05-29",
		Format:      "webp",
		Orientation: "landscape",
		Resolution:  "2k",
		AspectRatio: "16:9",
		MinWidth:    1600,
		MinHeight:   900,
	})

	require.Contains(t, where, "images.user_id = $1")
	require.Contains(t, where, "LOWER(tasks.prompt)")
	require.Contains(t, where, "images.output_format =")
	require.Contains(t, where, "images.width > images.height")
	require.Contains(t, where, "GREATEST(images.width, images.height)")
	require.Contains(t, where, "images.width * 9 = images.height * 16")
	require.NotContains(t, strings.ToLower(where), "visibility")
	require.Len(t, args, 7)
	require.Equal(t, int64(42), args[0])
	require.Equal(t, "%snow%", args[1])
	require.Equal(t, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), args[2])
	require.Equal(t, time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC), args[3])
	require.Equal(t, "webp", args[4])
	require.Equal(t, 1600, args[5])
	require.Equal(t, 900, args[6])
}

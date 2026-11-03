//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupFast20261103AdminCreateUpdate(t *testing.T) {
	for _, platform := range []string{PlatformOpenAI, PlatformComposite, PlatformAnthropic, PlatformGemini, PlatformGrok} {
		for _, enabled := range []bool{false, true} {
			t.Run(platform+map[bool]string{true: "/enabled", false: "/disabled"}[enabled], func(t *testing.T) {
				repo := &groupRepoStubForAdmin{}
				invalidator := &authCacheInvalidatorStub{}
				svc := &adminServiceImpl{groupRepo: repo, authCacheInvalidator: invalidator}
				created, err := svc.CreateGroup(context.Background(), &CreateGroupInput{
					Name: "group fast", Platform: platform, RateMultiplier: 1,
					ModelMatchPatterns: []string{"*"}, ForceOpenAIFast: enabled,
				})
				require.NoError(t, err)
				want := enabled && (platform == PlatformOpenAI || platform == PlatformComposite)
				require.Equal(t, want, created.ForceOpenAIFast)
				require.Equal(t, want, repo.created.ForceOpenAIFast)
				created.ID = 7
				repo.getByID = created
				preserved, err := svc.UpdateGroup(context.Background(), 7, &UpdateGroupInput{})
				require.NoError(t, err)
				require.Equal(t, want, preserved.ForceOpenAIFast, "omitted updates must preserve enabled flag")
				on := true
				updated, err := svc.UpdateGroup(context.Background(), 7, &UpdateGroupInput{ForceOpenAIFast: &on})
				require.NoError(t, err)
				require.Equal(t, platform == PlatformOpenAI || platform == PlatformComposite, updated.ForceOpenAIFast)
				off := false
				updated, err = svc.UpdateGroup(context.Background(), 7, &UpdateGroupInput{ForceOpenAIFast: &off})
				require.NoError(t, err)
				require.False(t, updated.ForceOpenAIFast)
				require.Equal(t, []int64{7, 7, 7}, invalidator.groupIDs)
			})
		}
	}
}

func TestGroupFast20261103AdminPlatformChangeClearsFlag(t *testing.T) {
	group := &Group{ID: 7, Platform: PlatformOpenAI, Status: StatusActive, ForceOpenAIFast: true}
	repo := &groupRepoStubForAdmin{getByID: group}
	svc := &adminServiceImpl{groupRepo: repo}
	updated, err := svc.UpdateGroup(context.Background(), 7, &UpdateGroupInput{Platform: PlatformComposite})
	require.NoError(t, err)
	require.True(t, updated.ForceOpenAIFast)
	updated, err = svc.UpdateGroup(context.Background(), 7, &UpdateGroupInput{Platform: PlatformAnthropic})
	require.NoError(t, err)
	require.False(t, updated.ForceOpenAIFast)
	require.False(t, repo.updated.ForceOpenAIFast)
}

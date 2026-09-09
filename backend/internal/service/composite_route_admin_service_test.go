package service

import (
	"context"
	"errors"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type compositeRouteGroupReaderStub struct {
	group *Group
	err   error
}

func (s *compositeRouteGroupReaderStub) GetByIDLite(context.Context, int64) (*Group, error) {
	return s.group, s.err
}

type compositeRouteRepositoryStub struct {
	routesByGroup map[int64][]CompositeModelRoute
	listErr       error
	created       *CompositeModelRoute
	updated       *CompositeModelRoute
	deletedID     int64
}

func (s *compositeRouteRepositoryStub) ListByGroup(_ context.Context, groupID int64, _ bool) ([]CompositeModelRoute, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return append([]CompositeModelRoute(nil), s.routesByGroup[groupID]...), nil
}

func (s *compositeRouteRepositoryStub) Create(_ context.Context, route *CompositeModelRoute) error {
	s.created = route
	return nil
}

func (s *compositeRouteRepositoryStub) Update(_ context.Context, route *CompositeModelRoute) error {
	s.updated = route
	return nil
}

func (s *compositeRouteRepositoryStub) Delete(_ context.Context, id int64) error {
	s.deletedID = id
	return nil
}

func (s *compositeRouteRepositoryStub) DeleteByGroup(context.Context, int64) error {
	return nil
}

func newCompositeRouteAdminServiceForTest(group *Group, routes map[int64][]CompositeModelRoute) (*compositeRouteAdminService, *compositeRouteRepositoryStub) {
	repo := &compositeRouteRepositoryStub{routesByGroup: routes}
	return NewCompositeRouteAdminService(&compositeRouteGroupReaderStub{group: group}, repo).(*compositeRouteAdminService), repo
}

func TestCompositeRouteAdminRejectsNonCompositeGroups(t *testing.T) {
	svc, _ := newCompositeRouteAdminServiceForTest(&Group{ID: 7, Platform: PlatformOpenAI}, nil)
	ctx := context.Background()
	validInput := CompositeRouteInput{PublicModel: "gpt-6", TargetPlatform: PlatformOpenAI}

	operations := []struct {
		name string
		call func() error
	}{
		{"list", func() error { _, err := svc.List(ctx, 7); return err }},
		{"create", func() error { _, err := svc.Create(ctx, 7, validInput); return err }},
		{"update", func() error { _, err := svc.Update(ctx, 7, 9, validInput); return err }},
		{"delete", func() error { return svc.Delete(ctx, 7, 9) }},
		{"preview", func() error { _, err := svc.Preview(ctx, 7, CompositeRoutePreviewRequest{Model: "gpt-6"}); return err }},
	}

	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			err := operation.call()
			require.Error(t, err)
			require.Equal(t, "COMPOSITE_GROUP_REQUIRED", infraerrors.Reason(err))
		})
	}
}

func TestCompositeRouteAdminCreateNormalizesDefaults(t *testing.T) {
	svc, repo := newCompositeRouteAdminServiceForTest(&Group{ID: 7, Platform: PlatformComposite}, nil)

	route, err := svc.Create(context.Background(), 7, CompositeRouteInput{
		PublicModel:    "  gpt-6-astra  ",
		MatchType:      " EXACT ",
		TargetPlatform: " OPENAI ",
		Endpoint:       " RESPONSES ",
		Enabled:        true,
		Notes:          "  route note  ",
	})

	require.NoError(t, err)
	require.Same(t, repo.created, route)
	require.Equal(t, int64(7), route.GroupID)
	require.Equal(t, "gpt-6-astra", route.PublicModel)
	require.Equal(t, CompositeRouteMatchExact, route.MatchType)
	require.Equal(t, PlatformOpenAI, route.TargetPlatform)
	require.Equal(t, "gpt-6-astra", route.UpstreamModel)
	require.Equal(t, CompositeRouteEndpointResponses, route.Endpoint)
	require.Equal(t, 100, route.Priority)
	require.True(t, route.Enabled)
	require.Equal(t, "route note", route.Notes)
}

func TestCompositeRouteAdminPreviewFiltersStaticCandidates(t *testing.T) {
	svc, _ := newCompositeRouteAdminServiceForTest(&Group{ID: 7, Platform: PlatformComposite}, map[int64][]CompositeModelRoute{
		7: {
			{ID: 3, GroupID: 7, PublicModel: "gpt-6", MatchType: CompositeRouteMatchExact, Endpoint: CompositeRouteEndpointAny, Enabled: true},
			{ID: 4, GroupID: 7, PublicModel: "gpt-6", MatchType: CompositeRouteMatchExact, Endpoint: CompositeRouteEndpointResponses, Enabled: true},
			{ID: 5, GroupID: 7, PublicModel: "gpt-", MatchType: CompositeRouteMatchPrefix, Endpoint: CompositeRouteEndpointResponses, Enabled: true},
			{ID: 6, GroupID: 7, PublicModel: "gpt-6", MatchType: CompositeRouteMatchExact, Endpoint: CompositeRouteEndpointResponses, Enabled: false},
			{ID: 7, GroupID: 7, PublicModel: "claude-", MatchType: CompositeRouteMatchPrefix, Endpoint: CompositeRouteEndpointResponses, Enabled: true},
		},
	})

	preview, err := svc.Preview(context.Background(), 7, CompositeRoutePreviewRequest{Model: " gpt-6 ", Endpoint: " RESPONSES "})
	require.NoError(t, err)
	require.Equal(t, "gpt-6", preview.Model)
	require.Equal(t, CompositeRouteEndpointResponses, preview.Endpoint)
	require.Equal(t, []int64{3, 4, 5}, []int64{preview.Candidates[0].ID, preview.Candidates[1].ID, preview.Candidates[2].ID})

	preview, err = svc.Preview(context.Background(), 7, CompositeRoutePreviewRequest{Model: "unknown-model"})
	require.NoError(t, err)
	require.Empty(t, preview.Candidates)
}

func TestCompositeRouteAdminRejectsCrossGroupRouteAndPropagatesLookupErrors(t *testing.T) {
	svc, repo := newCompositeRouteAdminServiceForTest(&Group{ID: 7, Platform: PlatformComposite}, map[int64][]CompositeModelRoute{
		8: {{ID: 99, GroupID: 8}},
	})
	input := CompositeRouteInput{PublicModel: "gpt-6", TargetPlatform: PlatformOpenAI}

	_, err := svc.Update(context.Background(), 7, 99, input)
	require.ErrorIs(t, err, ErrCompositeRouteNotFound)
	require.Nil(t, repo.updated)
	require.Zero(t, repo.deletedID)

	err = svc.Delete(context.Background(), 7, 99)
	require.ErrorIs(t, err, ErrCompositeRouteNotFound)
	require.Zero(t, repo.deletedID)

	repo.listErr = errors.New("repository unavailable")
	_, err = svc.Update(context.Background(), 7, 99, input)
	require.ErrorIs(t, err, repo.listErr)
	err = svc.Delete(context.Background(), 7, 99)
	require.ErrorIs(t, err, repo.listErr)
}

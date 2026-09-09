package service

import (
	"context"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type compositeRouteGroupReader interface {
	GetByIDLite(context.Context, int64) (*Group, error)
}

type compositeRouteAdminService struct {
	groupRepo compositeRouteGroupReader
	routeRepo CompositeModelRouteRepository
}

func NewCompositeRouteAdminService(groupRepo compositeRouteGroupReader, routeRepo CompositeModelRouteRepository) CompositeRouteAdminService {
	return &compositeRouteAdminService{groupRepo: groupRepo, routeRepo: routeRepo}
}

// ProvideCompositeRouteAdminService keeps Wire on the established GroupRepository port
// while the feature itself depends only on the one read it needs.
func ProvideCompositeRouteAdminService(groupRepo GroupRepository, routeRepo CompositeModelRouteRepository) CompositeRouteAdminService {
	return NewCompositeRouteAdminService(groupRepo, routeRepo)
}

func (s *compositeRouteAdminService) List(ctx context.Context, groupID int64) ([]CompositeModelRoute, error) {
	if err := s.requireCompositeGroup(ctx, groupID); err != nil {
		return nil, err
	}
	return s.routeRepo.ListByGroup(ctx, groupID, true)
}

func (s *compositeRouteAdminService) Create(ctx context.Context, groupID int64, input CompositeRouteInput) (*CompositeModelRoute, error) {
	if err := s.requireCompositeGroup(ctx, groupID); err != nil {
		return nil, err
	}
	input, err := validateCompositeRouteInput(input)
	if err != nil {
		return nil, err
	}
	route := &CompositeModelRoute{
		GroupID:        groupID,
		PublicModel:    input.PublicModel,
		MatchType:      input.MatchType,
		TargetPlatform: input.TargetPlatform,
		UpstreamModel:  input.UpstreamModel,
		Endpoint:       input.Endpoint,
		Priority:       input.Priority,
		Enabled:        input.Enabled,
		Notes:          input.Notes,
	}
	if err := s.routeRepo.Create(ctx, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *compositeRouteAdminService) Update(ctx context.Context, groupID, routeID int64, input CompositeRouteInput) (*CompositeModelRoute, error) {
	if err := s.requireCompositeGroup(ctx, groupID); err != nil {
		return nil, err
	}
	exists, err := s.routeBelongsToGroup(ctx, groupID, routeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCompositeRouteNotFound
	}
	input, err = validateCompositeRouteInput(input)
	if err != nil {
		return nil, err
	}
	route := &CompositeModelRoute{
		ID:             routeID,
		GroupID:        groupID,
		PublicModel:    input.PublicModel,
		MatchType:      input.MatchType,
		TargetPlatform: input.TargetPlatform,
		UpstreamModel:  input.UpstreamModel,
		Endpoint:       input.Endpoint,
		Priority:       input.Priority,
		Enabled:        input.Enabled,
		Notes:          input.Notes,
	}
	if err := s.routeRepo.Update(ctx, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *compositeRouteAdminService) Delete(ctx context.Context, groupID, routeID int64) error {
	if err := s.requireCompositeGroup(ctx, groupID); err != nil {
		return err
	}
	exists, err := s.routeBelongsToGroup(ctx, groupID, routeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCompositeRouteNotFound
	}
	return s.routeRepo.Delete(ctx, routeID)
}

func (s *compositeRouteAdminService) Preview(ctx context.Context, groupID int64, input CompositeRoutePreviewRequest) (*CompositeRoutePreview, error) {
	if err := s.requireCompositeGroup(ctx, groupID); err != nil {
		return nil, err
	}
	input, err := normalizeCompositeRoutePreviewRequest(input)
	if err != nil {
		return nil, err
	}
	routes, err := s.routeRepo.ListByGroup(ctx, groupID, false)
	if err != nil {
		return nil, err
	}
	candidates := make([]CompositeModelRoute, 0, len(routes))
	for _, route := range routes {
		if compositeRouteMatches(route, input.Model, input.Endpoint) {
			candidates = append(candidates, route)
		}
	}
	return &CompositeRoutePreview{
		GroupID:    groupID,
		Model:      input.Model,
		Endpoint:   input.Endpoint,
		Candidates: candidates,
	}, nil
}

func (s *compositeRouteAdminService) requireCompositeGroup(ctx context.Context, groupID int64) error {
	if groupID <= 0 {
		return infraerrors.BadRequest("INVALID_GROUP_ID", "group_id must be positive")
	}
	if s == nil || s.groupRepo == nil || s.routeRepo == nil {
		return infraerrors.New(503, "COMPOSITE_ROUTE_ADMIN_UNAVAILABLE", "composite route administration is unavailable")
	}
	group, err := s.groupRepo.GetByIDLite(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil || group.Platform != PlatformComposite {
		return infraerrors.BadRequest("COMPOSITE_GROUP_REQUIRED", "composite routes require a composite group")
	}
	return nil
}

func (s *compositeRouteAdminService) routeBelongsToGroup(ctx context.Context, groupID, routeID int64) (bool, error) {
	if routeID <= 0 {
		return false, nil
	}
	routes, err := s.routeRepo.ListByGroup(ctx, groupID, true)
	if err != nil {
		return false, err
	}
	for _, route := range routes {
		if route.ID == routeID {
			return true, nil
		}
	}
	return false, nil
}

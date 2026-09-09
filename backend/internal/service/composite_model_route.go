package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	CompositeRouteMatchExact  = "exact"
	CompositeRouteMatchPrefix = "prefix"

	CompositeRouteEndpointAny             = "any"
	CompositeRouteEndpointMessages        = "messages"
	CompositeRouteEndpointCountTokens     = "count_tokens"
	CompositeRouteEndpointResponses       = "responses"
	CompositeRouteEndpointChatCompletions = "chat_completions"
	CompositeRouteEndpointEmbeddings      = "embeddings"
	CompositeRouteEndpointImages          = "images"
	CompositeRouteEndpointGemini          = "gemini"
)

var (
	ErrCompositeRouteNotFound = infraerrors.NotFound("COMPOSITE_ROUTE_NOT_FOUND", "composite route not found")
	ErrCompositeRouteExists   = infraerrors.Conflict("COMPOSITE_ROUTE_EXISTS", "composite route already exists")
)

// CompositeModelRoute maps one public model in a composite group to a concrete provider model.
// P1 stores and previews this configuration only; it does not route requests.
type CompositeModelRoute struct {
	ID             int64     `json:"id"`
	GroupID        int64     `json:"group_id"`
	PublicModel    string    `json:"public_model"`
	MatchType      string    `json:"match_type"`
	TargetPlatform string    `json:"target_platform"`
	UpstreamModel  string    `json:"upstream_model"`
	Endpoint       string    `json:"endpoint"`
	Priority       int       `json:"priority"`
	Enabled        bool      `json:"enabled"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CompositeRouteInput struct {
	PublicModel    string
	MatchType      string
	TargetPlatform string
	UpstreamModel  string
	Endpoint       string
	Priority       int
	Enabled        bool
	Notes          string
}

type CompositeRoutePreviewRequest struct {
	Model    string `json:"model"`
	Endpoint string `json:"endpoint"`
}

// CompositeRoutePreview intentionally contains candidates rather than a selected target.
// Gateway resolver precedence and account ownership are introduced in S295-P2.
type CompositeRoutePreview struct {
	GroupID    int64                 `json:"group_id"`
	Model      string                `json:"model"`
	Endpoint   string                `json:"endpoint"`
	Candidates []CompositeModelRoute `json:"candidates"`
}

type CompositeModelRouteRepository interface {
	ListByGroup(ctx context.Context, groupID int64, includeDisabled bool) ([]CompositeModelRoute, error)
	Create(ctx context.Context, route *CompositeModelRoute) error
	Update(ctx context.Context, route *CompositeModelRoute) error
	Delete(ctx context.Context, id int64) error
	DeleteByGroup(ctx context.Context, groupID int64) error
}

// CompositeRouteAdminService is deliberately separate from AdminService so existing
// administration test doubles do not grow with a feature that is not on the gateway path yet.
type CompositeRouteAdminService interface {
	List(ctx context.Context, groupID int64) ([]CompositeModelRoute, error)
	Create(ctx context.Context, groupID int64, input CompositeRouteInput) (*CompositeModelRoute, error)
	Update(ctx context.Context, groupID, routeID int64, input CompositeRouteInput) (*CompositeModelRoute, error)
	Delete(ctx context.Context, groupID, routeID int64) error
	Preview(ctx context.Context, groupID int64, input CompositeRoutePreviewRequest) (*CompositeRoutePreview, error)
}

func normalizeCompositeRouteInput(input CompositeRouteInput) CompositeRouteInput {
	input.PublicModel = strings.TrimSpace(input.PublicModel)
	input.MatchType = strings.ToLower(strings.TrimSpace(input.MatchType))
	if input.MatchType == "" {
		input.MatchType = CompositeRouteMatchExact
	}
	input.TargetPlatform = strings.ToLower(strings.TrimSpace(input.TargetPlatform))
	input.UpstreamModel = strings.TrimSpace(input.UpstreamModel)
	input.Endpoint = normalizeCompositeRouteEndpoint(input.Endpoint)
	if input.Endpoint == "" {
		input.Endpoint = CompositeRouteEndpointAny
	}
	if input.Priority == 0 {
		input.Priority = 100
	}
	input.Notes = strings.TrimSpace(input.Notes)
	if input.UpstreamModel == "" && input.MatchType == CompositeRouteMatchExact {
		input.UpstreamModel = input.PublicModel
	}
	return input
}

func validateCompositeRouteInput(input CompositeRouteInput) (CompositeRouteInput, error) {
	input = normalizeCompositeRouteInput(input)
	if input.PublicModel == "" {
		return CompositeRouteInput{}, compositeRouteBadRequest("public_model is required")
	}
	if input.MatchType != CompositeRouteMatchExact && input.MatchType != CompositeRouteMatchPrefix {
		return CompositeRouteInput{}, compositeRouteBadRequest("match_type must be exact or prefix")
	}
	if !isCompositeRouteTargetPlatform(input.TargetPlatform) {
		return CompositeRouteInput{}, compositeRouteBadRequest("target_platform must be a concrete provider")
	}
	if !isCompositeRouteEndpoint(input.Endpoint) {
		return CompositeRouteInput{}, compositeRouteBadRequest("endpoint is not supported")
	}
	return input, nil
}

func normalizeCompositeRoutePreviewRequest(input CompositeRoutePreviewRequest) (CompositeRoutePreviewRequest, error) {
	input.Model = strings.TrimSpace(input.Model)
	input.Endpoint = normalizeCompositeRouteEndpoint(input.Endpoint)
	if input.Endpoint == "" {
		input.Endpoint = CompositeRouteEndpointAny
	}
	if input.Model == "" {
		return CompositeRoutePreviewRequest{}, compositeRouteBadRequest("model is required")
	}
	if !isCompositeRouteEndpoint(input.Endpoint) {
		return CompositeRoutePreviewRequest{}, compositeRouteBadRequest("endpoint is not supported")
	}
	return input, nil
}

func normalizeCompositeRouteEndpoint(endpoint string) string {
	return strings.ToLower(strings.TrimSpace(endpoint))
}

func isCompositeRouteEndpoint(endpoint string) bool {
	switch endpoint {
	case CompositeRouteEndpointAny,
		CompositeRouteEndpointMessages,
		CompositeRouteEndpointCountTokens,
		CompositeRouteEndpointResponses,
		CompositeRouteEndpointChatCompletions,
		CompositeRouteEndpointEmbeddings,
		CompositeRouteEndpointImages,
		CompositeRouteEndpointGemini:
		return true
	default:
		return false
	}
}

func isCompositeRouteTargetPlatform(platform string) bool {
	switch platform {
	case PlatformAnthropic,
		PlatformOpenAI,
		PlatformGemini,
		PlatformAntigravity,
		PlatformGrok,
		PlatformKimi,
		PlatformZhipu,
		PlatformDeepseek:
		return true
	default:
		return false
	}
}

func compositeRouteMatches(route CompositeModelRoute, model, endpoint string) bool {
	if !route.Enabled {
		return false
	}
	if route.Endpoint != CompositeRouteEndpointAny && route.Endpoint != endpoint {
		return false
	}
	switch route.MatchType {
	case CompositeRouteMatchExact:
		return route.PublicModel == model
	case CompositeRouteMatchPrefix:
		return strings.HasPrefix(model, route.PublicModel)
	default:
		return false
	}
}

func compositeRouteBadRequest(message string) error {
	return infraerrors.BadRequest("INVALID_COMPOSITE_ROUTE", message)
}

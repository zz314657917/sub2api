package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	CanvasNodeTypeText         = "text"
	CanvasNodeTypeImage        = "image"
	CanvasNodeTypePrompt       = "prompt"
	CanvasNodeTypeLoop         = "loop"
	CanvasNodeTypeGroup        = "group"
	CanvasNodeTypeTextToImage  = "text_to_image"
	CanvasNodeTypeImageToImage = "image_to_image"
	CanvasNodeTypeResult       = "result"

	CanvasRunStatusPending   = "pending"
	CanvasRunStatusRunning   = "running"
	CanvasRunStatusSucceeded = "succeeded"
	CanvasRunStatusFailed    = "failed"
	CanvasRunStatusCanceled  = "canceled"

	defaultCanvasTitle          = "Untitled canvas"
	defaultCanvasRunTriggerType = "manual"
	defaultCanvasImageModel     = "gpt-image-2"
	defaultCanvasChatModel      = "gpt-5.4"
	maxCanvasTitleLength        = 200
	maxCanvasDescriptionLength  = 4000
	maxCanvasNodeCount          = 256
	maxCanvasEdgeCount          = 512
	maxCanvasJSONBytes          = 256 << 10
)

var canvasNodeTypes = []string{
	CanvasNodeTypeText,
	CanvasNodeTypeImage,
	CanvasNodeTypePrompt,
	CanvasNodeTypeLoop,
	CanvasNodeTypeGroup,
	CanvasNodeTypeTextToImage,
	CanvasNodeTypeImageToImage,
	CanvasNodeTypeResult,
}

var canvasNodeTypeSet = map[string]struct{}{
	CanvasNodeTypeText:         {},
	CanvasNodeTypeImage:        {},
	CanvasNodeTypePrompt:       {},
	CanvasNodeTypeLoop:         {},
	CanvasNodeTypeGroup:        {},
	CanvasNodeTypeTextToImage:  {},
	CanvasNodeTypeImageToImage: {},
	CanvasNodeTypeResult:       {},
}

type CanvasDocument struct {
	ID          int64          `json:"id"`
	UserID      int64          `json:"user_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Nodes       []CanvasNode   `json:"nodes"`
	Edges       []CanvasEdge   `json:"edges"`
	Viewport    map[string]any `json:"viewport"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type CanvasListItem struct {
	ID          int64          `json:"id"`
	UserID      int64          `json:"user_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Viewport    map[string]any `json:"viewport"`
	Metadata    map[string]any `json:"metadata"`
	NodeCount   int            `json:"node_count"`
	EdgeCount   int            `json:"edge_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type CanvasNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Position map[string]any `json:"position"`
	Data     map[string]any `json:"data"`
	Metadata map[string]any `json:"metadata"`
}

type CanvasEdge struct {
	ID           string         `json:"id"`
	Source       string         `json:"source"`
	Target       string         `json:"target"`
	SourceHandle string         `json:"source_handle,omitempty"`
	TargetHandle string         `json:"target_handle,omitempty"`
	Data         map[string]any `json:"data"`
	Metadata     map[string]any `json:"metadata"`
}

type CanvasSaveInput struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Nodes       []CanvasNode   `json:"nodes"`
	Edges       []CanvasEdge   `json:"edges"`
	Viewport    map[string]any `json:"viewport"`
	Metadata    map[string]any `json:"metadata"`
}

type CanvasListFilters struct {
	Limit  int
	Offset int
}

type CanvasRun struct {
	ID           int64          `json:"id"`
	UserID       int64          `json:"user_id"`
	CanvasID     int64          `json:"canvas_id,omitempty"`
	Status       string         `json:"status"`
	TriggerType  string         `json:"trigger_type"`
	APIKeyID     int64          `json:"api_key_id,omitempty"`
	Model        string         `json:"model"`
	Input        map[string]any `json:"input"`
	Output       map[string]any `json:"output"`
	ErrorMessage string         `json:"error_message,omitempty"`
	Metadata     map[string]any `json:"metadata"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	CanceledAt   *time.Time     `json:"canceled_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type CanvasRunCreateInput struct {
	CanvasID    int64          `json:"canvas_id"`
	APIKeyID    int64          `json:"api_key_id"`
	Model       string         `json:"model"`
	TriggerType string         `json:"trigger_type"`
	Input       map[string]any `json:"input"`
	Metadata    map[string]any `json:"metadata"`
}

type CanvasRunListFilters struct {
	CanvasID int64
	Limit    int
	Offset   int
}

type CanvasModelCatalog struct {
	Object            string             `json:"object"`
	Items             []ModelCatalogItem `json:"items"`
	ChatModels        []string           `json:"chat_models"`
	ImageModels       []string           `json:"image_models"`
	NodeTypes         []string           `json:"node_types"`
	DefaultChatModel  string             `json:"default_chat_model"`
	DefaultImageModel string             `json:"default_image_model"`
}

type CanvasRepository interface {
	ListCanvases(ctx context.Context, userID int64, filters CanvasListFilters) ([]CanvasListItem, int, error)
	GetCanvas(ctx context.Context, userID int64, canvasID int64) (*CanvasDocument, error)
	SaveCanvas(ctx context.Context, userID int64, input CanvasSaveInput) (*CanvasDocument, error)
	DeleteCanvas(ctx context.Context, userID int64, canvasID int64) error
	CreateCanvasRun(ctx context.Context, userID int64, input CanvasRunCreateInput) (*CanvasRun, error)
	ListCanvasRuns(ctx context.Context, userID int64, filters CanvasRunListFilters) ([]CanvasRun, int, error)
	GetCanvasRun(ctx context.Context, userID int64, runID int64) (*CanvasRun, error)
	CancelCanvasRun(ctx context.Context, userID int64, runID int64) (*CanvasRun, error)
}

type CanvasService struct {
	repo CanvasRepository
}

func NewCanvasService(repo CanvasRepository) *CanvasService {
	return &CanvasService{repo: repo}
}

func (s *CanvasService) ListCanvases(ctx context.Context, userID int64, filters CanvasListFilters) ([]CanvasListItem, int, error) {
	if err := validateCanvasUser(userID); err != nil {
		return nil, 0, err
	}
	filters = normalizeCanvasListFilters(filters)
	items, total, err := s.repo.ListCanvases(ctx, userID, filters)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		normalizeCanvasSummaryJSON(&items[i])
	}
	return items, total, nil
}

func (s *CanvasService) GetCanvas(ctx context.Context, userID int64, canvasID int64) (*CanvasDocument, error) {
	if err := validateCanvasUser(userID); err != nil {
		return nil, err
	}
	if canvasID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CANVAS_ID", "invalid canvas id")
	}
	doc, err := s.repo.GetCanvas(ctx, userID, canvasID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("CANVAS_NOT_FOUND", "canvas not found")
	}
	if err != nil {
		return nil, err
	}
	normalizeCanvasDocumentJSON(doc)
	return doc, nil
}

func (s *CanvasService) SaveCanvas(ctx context.Context, userID int64, input CanvasSaveInput) (*CanvasDocument, error) {
	if err := validateCanvasUser(userID); err != nil {
		return nil, err
	}
	normalized, err := normalizeAndValidateCanvasSaveInput(input)
	if err != nil {
		return nil, err
	}
	doc, err := s.repo.SaveCanvas(ctx, userID, normalized)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("CANVAS_NOT_FOUND", "canvas not found")
	}
	if err != nil {
		return nil, err
	}
	normalizeCanvasDocumentJSON(doc)
	return doc, nil
}

func (s *CanvasService) DeleteCanvas(ctx context.Context, userID int64, canvasID int64) error {
	if err := validateCanvasUser(userID); err != nil {
		return err
	}
	if canvasID <= 0 {
		return infraerrors.BadRequest("INVALID_CANVAS_ID", "invalid canvas id")
	}
	if err := s.repo.DeleteCanvas(ctx, userID, canvasID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return infraerrors.NotFound("CANVAS_NOT_FOUND", "canvas not found")
		}
		return err
	}
	return nil
}

func (s *CanvasService) CreateRun(ctx context.Context, userID int64, input CanvasRunCreateInput) (*CanvasRun, error) {
	if err := validateCanvasUser(userID); err != nil {
		return nil, err
	}
	normalized, err := normalizeAndValidateCanvasRunInput(input)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetCanvas(ctx, userID, normalized.CanvasID); err != nil {
		return nil, err
	}
	run, err := s.repo.CreateCanvasRun(ctx, userID, normalized)
	if err != nil {
		return nil, err
	}
	normalizeCanvasRunJSON(run)
	return run, nil
}

func (s *CanvasService) ListRuns(ctx context.Context, userID int64, filters CanvasRunListFilters) ([]CanvasRun, int, error) {
	if err := validateCanvasUser(userID); err != nil {
		return nil, 0, err
	}
	filters = normalizeCanvasRunListFilters(filters)
	runs, total, err := s.repo.ListCanvasRuns(ctx, userID, filters)
	if err != nil {
		return nil, 0, err
	}
	for i := range runs {
		normalizeCanvasRunJSON(&runs[i])
	}
	return runs, total, nil
}

func (s *CanvasService) GetRun(ctx context.Context, userID int64, runID int64) (*CanvasRun, error) {
	if err := validateCanvasUser(userID); err != nil {
		return nil, err
	}
	if runID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CANVAS_RUN_ID", "invalid canvas run id")
	}
	run, err := s.repo.GetCanvasRun(ctx, userID, runID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("CANVAS_RUN_NOT_FOUND", "canvas run not found")
	}
	if err != nil {
		return nil, err
	}
	normalizeCanvasRunJSON(run)
	return run, nil
}

func (s *CanvasService) CancelRun(ctx context.Context, userID int64, runID int64) (*CanvasRun, error) {
	if err := validateCanvasUser(userID); err != nil {
		return nil, err
	}
	if runID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CANVAS_RUN_ID", "invalid canvas run id")
	}
	run, err := s.repo.CancelCanvasRun(ctx, userID, runID)
	if err == nil {
		normalizeCanvasRunJSON(run)
		return run, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	existing, getErr := s.repo.GetCanvasRun(ctx, userID, runID)
	if errors.Is(getErr, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("CANVAS_RUN_NOT_FOUND", "canvas run not found")
	}
	if getErr != nil {
		return nil, getErr
	}
	if existing.Status == CanvasRunStatusCanceled {
		normalizeCanvasRunJSON(existing)
		return existing, nil
	}
	return nil, infraerrors.Conflict("CANVAS_RUN_NOT_CANCELABLE", "canvas run is not cancelable")
}

func (s *CanvasService) ListModels(_ context.Context, userID int64) (CanvasModelCatalog, error) {
	if err := validateCanvasUser(userID); err != nil {
		return CanvasModelCatalog{}, err
	}
	return DefaultCanvasModelCatalog(), nil
}

func DefaultCanvasModelCatalog() CanvasModelCatalog {
	models := BuildModelCatalog(PlatformOpenAI, []string{
		defaultCanvasChatModel,
		"gpt-5.5",
		defaultCanvasImageModel,
		"gpt-image-1.5",
	})
	return CanvasModelCatalog{
		Object:            "canvas_model_catalog",
		Items:             models.Items,
		ChatModels:        models.ChatModels,
		ImageModels:       models.ImageModels,
		NodeTypes:         append([]string(nil), canvasNodeTypes...),
		DefaultChatModel:  defaultCanvasChatModel,
		DefaultImageModel: defaultCanvasImageModel,
	}
}

func normalizeAndValidateCanvasSaveInput(input CanvasSaveInput) (CanvasSaveInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		input.Title = defaultCanvasTitle
	}
	if len(input.Title) > maxCanvasTitleLength {
		return input, infraerrors.BadRequest("CANVAS_TITLE_TOO_LONG", "canvas title is too long")
	}
	input.Description = strings.TrimSpace(input.Description)
	if len(input.Description) > maxCanvasDescriptionLength {
		return input, infraerrors.BadRequest("CANVAS_DESCRIPTION_TOO_LONG", "canvas description is too long")
	}
	input.Viewport = normalizeCanvasJSONMap(input.Viewport)
	input.Metadata = normalizeCanvasJSONMap(input.Metadata)
	if err := validateCanvasJSON("viewport", input.Viewport); err != nil {
		return input, err
	}
	if err := validateCanvasJSON("metadata", input.Metadata); err != nil {
		return input, err
	}
	if len(input.Nodes) > maxCanvasNodeCount {
		return input, infraerrors.BadRequest("CANVAS_TOO_MANY_NODES", fmt.Sprintf("canvas cannot contain more than %d nodes", maxCanvasNodeCount))
	}
	if len(input.Edges) > maxCanvasEdgeCount {
		return input, infraerrors.BadRequest("CANVAS_TOO_MANY_EDGES", fmt.Sprintf("canvas cannot contain more than %d edges", maxCanvasEdgeCount))
	}
	nodeIDs := make(map[string]struct{}, len(input.Nodes))
	for i := range input.Nodes {
		node, err := normalizeAndValidateCanvasNode(input.Nodes[i], nodeIDs)
		if err != nil {
			return input, err
		}
		input.Nodes[i] = node
		nodeIDs[node.ID] = struct{}{}
	}
	edgeIDs := make(map[string]struct{}, len(input.Edges))
	for i := range input.Edges {
		edge, err := normalizeAndValidateCanvasEdge(input.Edges[i], edgeIDs, nodeIDs)
		if err != nil {
			return input, err
		}
		input.Edges[i] = edge
		edgeIDs[edge.ID] = struct{}{}
	}
	return input, nil
}

func normalizeAndValidateCanvasNode(node CanvasNode, seen map[string]struct{}) (CanvasNode, error) {
	node.ID = strings.TrimSpace(node.ID)
	if node.ID == "" {
		return node, infraerrors.BadRequest("CANVAS_NODE_ID_REQUIRED", "canvas node id is required")
	}
	if _, ok := seen[node.ID]; ok {
		return node, infraerrors.BadRequest("CANVAS_NODE_ID_DUPLICATED", "canvas node id is duplicated")
	}
	node.Type = strings.ToLower(strings.TrimSpace(node.Type))
	if _, ok := canvasNodeTypeSet[node.Type]; !ok {
		return node, infraerrors.BadRequest("CANVAS_NODE_TYPE_INVALID", "canvas node type is invalid")
	}
	node.Position = normalizeCanvasJSONMap(node.Position)
	node.Data = normalizeCanvasJSONMap(node.Data)
	node.Metadata = normalizeCanvasJSONMap(node.Metadata)
	if err := validateCanvasJSON("node position", node.Position); err != nil {
		return node, err
	}
	if err := validateCanvasJSON("node data", node.Data); err != nil {
		return node, err
	}
	if err := validateCanvasJSON("node metadata", node.Metadata); err != nil {
		return node, err
	}
	return node, nil
}

func normalizeAndValidateCanvasEdge(edge CanvasEdge, seen map[string]struct{}, nodes map[string]struct{}) (CanvasEdge, error) {
	edge.ID = strings.TrimSpace(edge.ID)
	if edge.ID == "" {
		return edge, infraerrors.BadRequest("CANVAS_EDGE_ID_REQUIRED", "canvas edge id is required")
	}
	if _, ok := seen[edge.ID]; ok {
		return edge, infraerrors.BadRequest("CANVAS_EDGE_ID_DUPLICATED", "canvas edge id is duplicated")
	}
	edge.Source = strings.TrimSpace(edge.Source)
	edge.Target = strings.TrimSpace(edge.Target)
	edge.SourceHandle = strings.TrimSpace(edge.SourceHandle)
	edge.TargetHandle = strings.TrimSpace(edge.TargetHandle)
	if edge.Source == "" || edge.Target == "" {
		return edge, infraerrors.BadRequest("CANVAS_EDGE_ENDPOINT_REQUIRED", "canvas edge source and target are required")
	}
	if _, ok := nodes[edge.Source]; !ok {
		return edge, infraerrors.BadRequest("CANVAS_EDGE_SOURCE_INVALID", "canvas edge source node does not exist")
	}
	if _, ok := nodes[edge.Target]; !ok {
		return edge, infraerrors.BadRequest("CANVAS_EDGE_TARGET_INVALID", "canvas edge target node does not exist")
	}
	edge.Data = normalizeCanvasJSONMap(edge.Data)
	edge.Metadata = normalizeCanvasJSONMap(edge.Metadata)
	if err := validateCanvasJSON("edge data", edge.Data); err != nil {
		return edge, err
	}
	if err := validateCanvasJSON("edge metadata", edge.Metadata); err != nil {
		return edge, err
	}
	return edge, nil
}

func normalizeAndValidateCanvasRunInput(input CanvasRunCreateInput) (CanvasRunCreateInput, error) {
	if input.CanvasID <= 0 {
		return input, infraerrors.BadRequest("CANVAS_ID_REQUIRED", "canvas id is required")
	}
	if input.APIKeyID < 0 {
		return input, infraerrors.BadRequest("INVALID_API_KEY", "invalid api key id")
	}
	input.Model = strings.TrimSpace(input.Model)
	if input.Model == "" {
		input.Model = defaultCanvasImageModel
	}
	input.TriggerType = strings.TrimSpace(input.TriggerType)
	if input.TriggerType == "" {
		input.TriggerType = defaultCanvasRunTriggerType
	}
	input.Input = normalizeCanvasJSONMap(input.Input)
	input.Metadata = normalizeCanvasJSONMap(input.Metadata)
	if err := validateCanvasJSON("run input", input.Input); err != nil {
		return input, err
	}
	if err := validateCanvasJSON("run metadata", input.Metadata); err != nil {
		return input, err
	}
	return input, nil
}

func normalizeCanvasListFilters(filters CanvasListFilters) CanvasListFilters {
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func normalizeCanvasRunListFilters(filters CanvasRunListFilters) CanvasRunListFilters {
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func normalizeCanvasJSONMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(value))
	for key, item := range value {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = item
	}
	if len(out) == 0 {
		return map[string]any{}
	}
	return out
}

func validateCanvasJSON(name string, value map[string]any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return infraerrors.BadRequest("CANVAS_JSON_INVALID", name+" must be valid JSON")
	}
	if len(data) > maxCanvasJSONBytes {
		return infraerrors.BadRequest("CANVAS_JSON_TOO_LARGE", name+" is too large")
	}
	return nil
}

func validateCanvasUser(userID int64) error {
	if userID <= 0 {
		return infraerrors.Unauthorized("UNAUTHORIZED", "authentication required")
	}
	return nil
}

func normalizeCanvasDocumentJSON(doc *CanvasDocument) {
	if doc == nil {
		return
	}
	if doc.Nodes == nil {
		doc.Nodes = []CanvasNode{}
	}
	if doc.Edges == nil {
		doc.Edges = []CanvasEdge{}
	}
	doc.Viewport = normalizeCanvasJSONMap(doc.Viewport)
	doc.Metadata = normalizeCanvasJSONMap(doc.Metadata)
	for i := range doc.Nodes {
		doc.Nodes[i].Position = normalizeCanvasJSONMap(doc.Nodes[i].Position)
		doc.Nodes[i].Data = normalizeCanvasJSONMap(doc.Nodes[i].Data)
		doc.Nodes[i].Metadata = normalizeCanvasJSONMap(doc.Nodes[i].Metadata)
	}
	for i := range doc.Edges {
		doc.Edges[i].Data = normalizeCanvasJSONMap(doc.Edges[i].Data)
		doc.Edges[i].Metadata = normalizeCanvasJSONMap(doc.Edges[i].Metadata)
	}
}

func normalizeCanvasSummaryJSON(item *CanvasListItem) {
	if item == nil {
		return
	}
	item.Viewport = normalizeCanvasJSONMap(item.Viewport)
	item.Metadata = normalizeCanvasJSONMap(item.Metadata)
}

func normalizeCanvasRunJSON(run *CanvasRun) {
	if run == nil {
		return
	}
	run.Input = normalizeCanvasJSONMap(run.Input)
	run.Output = normalizeCanvasJSONMap(run.Output)
	run.Metadata = normalizeCanvasJSONMap(run.Metadata)
}

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type fakeCanvasRepo struct {
	documents map[int64]CanvasDocument
	runs      map[int64]CanvasRun

	lastListUserID   int64
	lastSaveUserID   int64
	lastRunUserID    int64
	lastRunInput     CanvasRunCreateInput
	lastCancelUserID int64
}

func newFakeCanvasRepo() *fakeCanvasRepo {
	now := time.Now()
	return &fakeCanvasRepo{
		documents: map[int64]CanvasDocument{
			12: {
				ID:        12,
				UserID:    42,
				Title:     "Campaign",
				Nodes:     []CanvasNode{{ID: "prompt", Type: CanvasNodeTypePrompt, Position: map[string]any{}, Data: map[string]any{}}},
				Edges:     []CanvasEdge{},
				Viewport:  map[string]any{},
				Metadata:  map[string]any{},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		runs: map[int64]CanvasRun{
			7: {
				ID:          7,
				UserID:      42,
				CanvasID:    12,
				Status:      CanvasRunStatusPending,
				TriggerType: defaultCanvasRunTriggerType,
				Input:       map[string]any{},
				Output:      map[string]any{},
				Metadata:    map[string]any{},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			8: {
				ID:          8,
				UserID:      42,
				CanvasID:    12,
				Status:      CanvasRunStatusSucceeded,
				TriggerType: defaultCanvasRunTriggerType,
				Input:       map[string]any{},
				Output:      map[string]any{},
				Metadata:    map[string]any{},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}
}

func (r *fakeCanvasRepo) ListCanvases(_ context.Context, userID int64, _ CanvasListFilters) ([]CanvasListItem, int, error) {
	r.lastListUserID = userID
	items := make([]CanvasListItem, 0)
	for _, doc := range r.documents {
		if doc.UserID != userID {
			continue
		}
		items = append(items, CanvasListItem{
			ID:        doc.ID,
			UserID:    doc.UserID,
			Title:     doc.Title,
			Viewport:  doc.Viewport,
			Metadata:  doc.Metadata,
			NodeCount: len(doc.Nodes),
			EdgeCount: len(doc.Edges),
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		})
	}
	return items, len(items), nil
}

func (r *fakeCanvasRepo) GetCanvas(_ context.Context, userID int64, canvasID int64) (*CanvasDocument, error) {
	doc, ok := r.documents[canvasID]
	if !ok || doc.UserID != userID {
		return nil, sql.ErrNoRows
	}
	copy := doc
	return &copy, nil
}

func (r *fakeCanvasRepo) SaveCanvas(_ context.Context, userID int64, input CanvasSaveInput) (*CanvasDocument, error) {
	r.lastSaveUserID = userID
	id := input.ID
	if id <= 0 {
		id = 20
	}
	if input.ID > 0 {
		doc, ok := r.documents[input.ID]
		if !ok || doc.UserID != userID {
			return nil, sql.ErrNoRows
		}
	}
	doc := CanvasDocument{
		ID:          id,
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Nodes:       input.Nodes,
		Edges:       input.Edges,
		Viewport:    input.Viewport,
		Metadata:    input.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.documents[id] = doc
	return &doc, nil
}

func (r *fakeCanvasRepo) DeleteCanvas(_ context.Context, userID int64, canvasID int64) error {
	doc, ok := r.documents[canvasID]
	if !ok || doc.UserID != userID {
		return sql.ErrNoRows
	}
	delete(r.documents, canvasID)
	return nil
}

func (r *fakeCanvasRepo) CreateCanvasRun(_ context.Context, userID int64, input CanvasRunCreateInput) (*CanvasRun, error) {
	r.lastRunUserID = userID
	r.lastRunInput = input
	run := CanvasRun{
		ID:          30,
		UserID:      userID,
		CanvasID:    input.CanvasID,
		Status:      CanvasRunStatusPending,
		TriggerType: input.TriggerType,
		APIKeyID:    input.APIKeyID,
		Model:       input.Model,
		Input:       input.Input,
		Output:      map[string]any{},
		Metadata:    input.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.runs[run.ID] = run
	return &run, nil
}

func (r *fakeCanvasRepo) ListCanvasRuns(_ context.Context, userID int64, filters CanvasRunListFilters) ([]CanvasRun, int, error) {
	items := make([]CanvasRun, 0)
	for _, run := range r.runs {
		if run.UserID != userID {
			continue
		}
		if filters.CanvasID > 0 && run.CanvasID != filters.CanvasID {
			continue
		}
		items = append(items, run)
	}
	return items, len(items), nil
}

func (r *fakeCanvasRepo) GetCanvasRun(_ context.Context, userID int64, runID int64) (*CanvasRun, error) {
	run, ok := r.runs[runID]
	if !ok || run.UserID != userID {
		return nil, sql.ErrNoRows
	}
	copy := run
	return &copy, nil
}

func (r *fakeCanvasRepo) CancelCanvasRun(_ context.Context, userID int64, runID int64) (*CanvasRun, error) {
	r.lastCancelUserID = userID
	run, ok := r.runs[runID]
	if !ok || run.UserID != userID || (run.Status != CanvasRunStatusPending && run.Status != CanvasRunStatusRunning) {
		return nil, sql.ErrNoRows
	}
	run.Status = CanvasRunStatusCanceled
	r.runs[runID] = run
	return &run, nil
}

func TestCanvasServiceSaveValidatesNodesEdgesAndUsesCurrentUser(t *testing.T) {
	repo := newFakeCanvasRepo()
	svc := NewCanvasService(repo)

	doc, err := svc.SaveCanvas(context.Background(), 42, CanvasSaveInput{
		Title: " Campaign ",
		Nodes: []CanvasNode{
			{ID: " prompt ", Type: "PROMPT", Position: map[string]any{"x": 10}, Data: map[string]any{"title": "Prompt"}},
			{ID: " result ", Type: CanvasNodeTypeResult, Position: map[string]any{}, Data: map[string]any{}},
		},
		Edges: []CanvasEdge{
			{ID: " edge ", Source: "prompt", Target: "result"},
		},
	})

	require.NoError(t, err)
	require.Equal(t, int64(42), repo.lastSaveUserID)
	require.Equal(t, "Campaign", doc.Title)
	require.Equal(t, "prompt", doc.Nodes[0].ID)
	require.Equal(t, CanvasNodeTypePrompt, doc.Nodes[0].Type)
	require.Equal(t, "edge", doc.Edges[0].ID)
}

func TestCanvasServiceSaveRejectsEdgesToMissingNodes(t *testing.T) {
	svc := NewCanvasService(newFakeCanvasRepo())

	_, err := svc.SaveCanvas(context.Background(), 42, CanvasSaveInput{
		Title: "Broken",
		Nodes: []CanvasNode{{ID: "prompt", Type: CanvasNodeTypePrompt}},
		Edges: []CanvasEdge{{ID: "edge", Source: "prompt", Target: "missing"}},
	})

	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))
	require.Equal(t, "CANVAS_EDGE_TARGET_INVALID", infraerrors.Reason(err))
}

func TestCanvasServiceCreateRunChecksCanvasOwnershipAndNormalizesDefaults(t *testing.T) {
	repo := newFakeCanvasRepo()
	svc := NewCanvasService(repo)

	run, err := svc.CreateRun(context.Background(), 42, CanvasRunCreateInput{CanvasID: 12})

	require.NoError(t, err)
	require.Equal(t, int64(42), repo.lastRunUserID)
	require.Equal(t, int64(12), repo.lastRunInput.CanvasID)
	require.Equal(t, defaultCanvasImageModel, repo.lastRunInput.Model)
	require.Equal(t, defaultCanvasRunTriggerType, repo.lastRunInput.TriggerType)
	require.Equal(t, CanvasRunStatusPending, run.Status)

	_, err = svc.CreateRun(context.Background(), 99, CanvasRunCreateInput{CanvasID: 12})
	require.Error(t, err)
	require.True(t, infraerrors.IsNotFound(err))
}

func TestCanvasServiceGetAndCancelRunUseCurrentUser(t *testing.T) {
	repo := newFakeCanvasRepo()
	svc := NewCanvasService(repo)

	run, err := svc.GetRun(context.Background(), 42, 7)
	require.NoError(t, err)
	require.Equal(t, int64(7), run.ID)

	canceled, err := svc.CancelRun(context.Background(), 42, 7)
	require.NoError(t, err)
	require.Equal(t, int64(42), repo.lastCancelUserID)
	require.Equal(t, CanvasRunStatusCanceled, canceled.Status)

	_, err = svc.CancelRun(context.Background(), 42, 8)
	require.Error(t, err)
	require.True(t, infraerrors.IsConflict(err))
}

package admin

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type compositeRouteAdminHandlerStub struct {
	service.CompositeRouteAdminService
	createGroupID int64
	createInput   service.CompositeRouteInput
	previewGroup  int64
	previewInput  service.CompositeRoutePreviewRequest
	calls         int
}

func (s *compositeRouteAdminHandlerStub) Create(_ context.Context, groupID int64, input service.CompositeRouteInput) (*service.CompositeModelRoute, error) {
	s.calls++
	s.createGroupID = groupID
	s.createInput = input
	return &service.CompositeModelRoute{ID: 11, GroupID: groupID, Enabled: input.Enabled}, nil
}

func (s *compositeRouteAdminHandlerStub) Preview(_ context.Context, groupID int64, input service.CompositeRoutePreviewRequest) (*service.CompositeRoutePreview, error) {
	s.calls++
	s.previewGroup = groupID
	s.previewInput = input
	return &service.CompositeRoutePreview{
		GroupID:  groupID,
		Model:    input.Model,
		Endpoint: input.Endpoint,
		Candidates: []service.CompositeModelRoute{{
			ID: 11, GroupID: groupID, PublicModel: input.Model, Enabled: true,
		}},
	}, nil
}

func (s *compositeRouteAdminHandlerStub) Delete(context.Context, int64, int64) error {
	s.calls++
	return nil
}

func newCompositeRouteHandlerRouter(stub service.CompositeRouteAdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	groupHandler := NewGroupHandler(nil, nil, nil)
	groupHandler.SetCompositeRouteAdminService(stub)
	router.POST("/groups/:id/composite-routes", groupHandler.CreateCompositeRoute)
	router.POST("/groups/:id/composite-routes/preview", groupHandler.PreviewCompositeRoute)
	router.DELETE("/groups/:id/composite-routes/:route_id", groupHandler.DeleteCompositeRoute)
	return router
}

func TestGroupHandlerCompositeRouteDefaultsEnabledAndMapsPreview(t *testing.T) {
	stub := &compositeRouteAdminHandlerStub{}
	router := newCompositeRouteHandlerRouter(stub)

	create := httptest.NewRequest(http.MethodPost, "/groups/7/composite-routes", bytes.NewBufferString(`{"public_model":"gpt-6","target_platform":"openai"}`))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	router.ServeHTTP(created, create)
	require.Equal(t, http.StatusCreated, created.Code)
	require.Equal(t, int64(7), stub.createGroupID)
	require.True(t, stub.createInput.Enabled)
	require.Equal(t, "gpt-6", stub.createInput.PublicModel)

	preview := httptest.NewRequest(http.MethodPost, "/groups/7/composite-routes/preview", bytes.NewBufferString(`{"model":"gpt-6","endpoint":"responses"}`))
	preview.Header.Set("Content-Type", "application/json")
	previewed := httptest.NewRecorder()
	router.ServeHTTP(previewed, preview)
	require.Equal(t, http.StatusOK, previewed.Code)
	require.Equal(t, int64(7), stub.previewGroup)
	require.Equal(t, service.CompositeRoutePreviewRequest{Model: "gpt-6", Endpoint: "responses"}, stub.previewInput)
	require.Contains(t, previewed.Body.String(), `"candidates":[{"id":11`)
}

func TestGroupHandlerCompositeRouteRejectsInvalidIDs(t *testing.T) {
	stub := &compositeRouteAdminHandlerStub{}
	router := newCompositeRouteHandlerRouter(stub)

	request := httptest.NewRequest(http.MethodPost, "/groups/not-a-number/composite-routes", bytes.NewBufferString(`{"public_model":"gpt-6","target_platform":"openai"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Zero(t, stub.calls)

	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/groups/7/composite-routes/0", nil))
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Zero(t, stub.calls)
}

func TestGroupHandlerCompositeRouteAcceptsCompositeGroupPlatform(t *testing.T) {
	for _, target := range []any{&CreateGroupRequest{}, &UpdateGroupRequest{}} {
		gin.SetMode(gin.TestMode)
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBufferString(`{"name":"composite","platform":"composite"}`))
		context.Request.Header.Set("Content-Type", "application/json")
		require.NoError(t, context.ShouldBindJSON(target))
	}
}

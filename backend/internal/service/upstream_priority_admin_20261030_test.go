package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type TestUpstreamPriorityAdminRepository struct {
	UserRepository
	user     *User
	getErr   error
	writeErr error
	updates  int
	deletes  int
}

func (r *TestUpstreamPriorityAdminRepository) GetByID(context.Context, int64) (*User, error) {
	return r.user, r.getErr
}

func (r *TestUpstreamPriorityAdminRepository) Update(_ context.Context, user *User, _ UserUpdateFields) error {
	r.updates++
	r.user = user
	return r.writeErr
}

func (r *TestUpstreamPriorityAdminRepository) Delete(context.Context, int64) error {
	r.deletes++
	return r.writeErr
}

func TestUpstreamPriorityAdminProtectionHTTP(t *testing.T) {
	for _, operation := range []string{"disable", "delete"} {
		t.Run(operation, func(t *testing.T) {
			repo := &TestUpstreamPriorityAdminRepository{user: &User{ID: 1, Role: RoleAdmin, Status: StatusActive}}
			svc := &adminServiceImpl{userRepo: repo}
			var err error
			reason := "ADMIN_DELETE_FORBIDDEN"
			if operation == "disable" {
				_, err = svc.UpdateUser(context.Background(), 1, &UpdateUserInput{Status: "disabled"})
				reason = "ADMIN_DISABLE_FORBIDDEN"
			} else {
				err = svc.DeleteUser(context.Background(), 1)
			}
			require.Error(t, err)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/users/1", nil)
			require.True(t, response.ErrorFrom(ctx, err))
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			require.JSONEq(t, `{"code":400,"reason":"`+reason+`","message":"cannot `+operation+` admin user; change the role to user first"}`, recorder.Body.String())
			require.Zero(t, repo.updates)
			require.Zero(t, repo.deletes)
			require.Equal(t, StatusActive, repo.user.Status)
		})
	}
}

func TestUpstreamPriorityAdminRegularAndRepositoryErrors(t *testing.T) {
	backendErr := errors.New("repository unavailable")
	for _, operation := range []string{"disable", "delete"} {
		for _, failure := range []string{"none", "notfound", "read", "write"} {
			t.Run(operation+"/"+failure, func(t *testing.T) {
				repo := &TestUpstreamPriorityAdminRepository{user: &User{ID: 7, Role: RoleUser, Status: StatusActive}}
				wantErr := error(nil)
				switch failure {
				case "notfound":
					repo.getErr, wantErr = ErrUserNotFound, ErrUserNotFound
				case "read":
					repo.getErr, wantErr = backendErr, backendErr
				case "write":
					repo.writeErr, wantErr = backendErr, backendErr
				}
				svc := &adminServiceImpl{userRepo: repo}
				var err error
				if operation == "disable" {
					_, err = svc.UpdateUser(context.Background(), 7, &UpdateUserInput{Status: "disabled"})
				} else {
					err = svc.DeleteUser(context.Background(), 7)
				}
				if wantErr == nil {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, err, wantErr)
					require.NotContains(t, infraerrors.Reason(err), "ADMIN_")
					recorder := httptest.NewRecorder()
					ctx, _ := gin.CreateTestContext(recorder)
					require.True(t, response.ErrorFrom(ctx, err))
					wantStatus := http.StatusInternalServerError
					if failure == "notfound" {
						wantStatus = http.StatusNotFound
					}
					require.Equal(t, wantStatus, recorder.Code)
				}
				wantWrites := 1
				if failure == "notfound" || failure == "read" {
					wantWrites = 0
				}
				if operation == "disable" {
					require.Equal(t, wantWrites, repo.updates)
					require.Zero(t, repo.deletes)
				} else {
					require.Equal(t, wantWrites, repo.deletes)
					require.Zero(t, repo.updates)
				}
			})
		}
	}
}

package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type pelicanAccessStub struct {
	groups []service.Group
	err    error
}

type pelicanAPIStub struct {
	pelicanAPI
	auth service.PelicanAuthorization
	id   int64
}

func (s *pelicanAPIStub) GetResult(_ context.Context, auth service.PelicanAuthorization, id int64) (*service.PelicanResult, error) {
	s.auth, s.id = auth, id
	return nil, service.ErrPelicanNotFound
}

func TestPelicanResultPassesAuthenticatedScope(t *testing.T) {
	stub := &pelicanAPIStub{}
	h := &PelicanTestHandler{svc: stub, groups: pelicanAccessStub{groups: []service.Group{{ID: 12}}}}
	c, w := pelicanContext("user", true)
	c.Params = gin.Params{{Key: "id", Value: "99"}}
	h.Result(c)
	if w.Code != 404 || stub.id != 99 || stub.auth.Admin || len(stub.auth.GroupIDs) != 1 || stub.auth.GroupIDs[0] != 12 {
		t.Fatalf("status=%d id=%d scope=%+v", w.Code, stub.id, stub.auth)
	}
}

func (s pelicanAccessStub) GetAvailableGroups(context.Context, int64) ([]service.Group, error) {
	return s.groups, s.err
}

func pelicanContext(role string, authenticated bool) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/pelican-tests", nil)
	if authenticated {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
	}
	c.Set(string(middleware.ContextKeyUserRole), role)
	return c, w
}

func TestPelicanAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name, role string
		logged     bool
		groups     []service.Group
		err        error
		status     int
		admin      bool
		count      int
	}{
		{"anonymous", "", false, nil, nil, 401, false, 0},
		{"empty access", "user", true, nil, nil, 200, false, 0},
		{"allowed groups", "user", true, []service.Group{{ID: 12}}, nil, 200, false, 1},
		{"lookup fails closed", "user", true, nil, errors.New("private database details"), 500, false, 0},
		{"admin", service.RoleAdmin, true, nil, nil, 200, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, w := pelicanContext(tc.role, tc.logged)
			h := &PelicanTestHandler{groups: pelicanAccessStub{tc.groups, tc.err}}
			auth, ok := h.authorization(c)
			if w.Code != tc.status || ok != (tc.status == 200) || auth.Admin != tc.admin || len(auth.GroupIDs) != tc.count {
				t.Fatalf("status=%d ok=%v auth=%+v", w.Code, ok, auth)
			}
			if strings.Contains(w.Body.String(), "private database") {
				t.Fatal("private error leaked")
			}
			if w.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("missing cache isolation")
			}
		})
	}
}

func TestPelicanErrorMappingRedactsDetails(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{service.ErrPelicanNotFound, 404}, {service.ErrPelicanConflict, 409}, {service.ErrPelicanInvalid, 400}, {errors.New("secret-provider-key"), 500},
	} {
		c, w := pelicanContext("user", true)
		pelicanError(c, fmt.Errorf("secret-provider-key: %w", tc.err))
		if w.Code != tc.status || strings.Contains(w.Body.String(), "secret-provider-key") {
			t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
		}
	}
}

func TestPelicanAdminHandlersRejectOrdinaryUser(t *testing.T) {
	h := &PelicanTestHandler{}
	for _, run := range []func(*gin.Context){h.ListPlans, h.SavePlan, h.DeletePlan, h.RunPlan} {
		c, w := pelicanContext("user", true)
		run(c)
		if w.Code != 403 {
			t.Fatalf("got %d", w.Code)
		}
	}
}

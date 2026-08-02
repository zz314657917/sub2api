package service

import (
	"context"
	"database/sql"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type userErrorOpsRepoStub struct {
	OpsRepository
	filter *OpsErrorLogFilter
	detail *OpsErrorLogDetail
	err    error
}

func (s *userErrorOpsRepoStub) ListErrorLogs(_ context.Context, filter *OpsErrorLogFilter) (*OpsErrorLogList, error) {
	s.filter = filter
	return &OpsErrorLogList{Errors: []*OpsErrorLog{{Phase: "request", Type: "rate_limit_error", RequestedModel: "gpt-test"}}, Total: 1, Page: 1, PageSize: 20}, nil
}

func (s *userErrorOpsRepoStub) GetErrorLogByID(context.Context, int64) (*OpsErrorLogDetail, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.detail, nil
}

func TestListUserErrorRequestsForcesOwnershipAndFilters(t *testing.T) {
	stub := &userErrorOpsRepoStub{}
	svc := &OpsService{opsRepo: stub}
	keyID := int64(7)
	in := &OpsErrorLogFilter{View: "errors", Phase: "upstream", Owner: "provider", APIKeyID: &keyID}
	out, err := svc.ListUserErrorRequests(context.Background(), 42, in)
	if err != nil {
		t.Fatal(err)
	}
	if stub.filter == nil || stub.filter.UserID == nil || *stub.filter.UserID != 42 {
		t.Fatalf("user ownership not forced: %+v", stub.filter)
	}
	if stub.filter.View != "all" || !stub.filter.ExcludeCountTokens || !stub.filter.ModelFuzzy || stub.filter.Phase != "" || stub.filter.Owner != "" || stub.filter.Source != "" {
		t.Fatalf("unsafe filter fields survived: %+v", stub.filter)
	}
	if stub.filter.APIKeyID == nil || *stub.filter.APIKeyID != keyID || len(out.Items) != 1 || out.Items[0].Category != "rate_limit" {
		t.Fatalf("unexpected result/filter: result=%+v filter=%+v", out, stub.filter)
	}
	if in.View != "errors" || in.Phase != "upstream" || in.UserID != nil {
		t.Fatal("caller filter was mutated")
	}
}

func TestGetUserErrorRequestDetailReturnsNotFoundForOtherUser(t *testing.T) {
	owner := int64(99)
	stub := &userErrorOpsRepoStub{detail: &OpsErrorLogDetail{OpsErrorLog: OpsErrorLog{ID: 5, UserID: &owner}}}
	svc := &OpsService{opsRepo: stub}
	got, err := svc.GetUserErrorRequestDetail(context.Background(), 42, 5)
	if got != nil || !infraerrors.IsNotFound(err) {
		t.Fatalf("want not found and no detail, got detail=%+v err=%v", got, err)
	}
	stub.err = sql.ErrNoRows
	got, err = svc.GetUserErrorRequestDetail(context.Background(), 42, 5)
	if got != nil || !infraerrors.IsNotFound(err) {
		t.Fatalf("missing row should be not found, got detail=%+v err=%v", got, err)
	}
}

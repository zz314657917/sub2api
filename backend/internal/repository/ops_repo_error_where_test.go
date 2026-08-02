package repository

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestBuildOpsErrorLogsWhere_QueryUsesQualifiedColumns(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		Query: "ACCESS_DENIED",
	}

	where, args := buildOpsErrorLogsWhere(filter)
	if where == "" {
		t.Fatalf("where should not be empty")
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if !strings.Contains(where, "e.request_id ILIKE $") {
		t.Fatalf("where should include qualified request_id condition: %s", where)
	}
	if !strings.Contains(where, "e.client_request_id ILIKE $") {
		t.Fatalf("where should include qualified client_request_id condition: %s", where)
	}
	if !strings.Contains(where, "e.error_message ILIKE $") {
		t.Fatalf("where should include qualified error_message condition: %s", where)
	}
}

func TestBuildOpsErrorLogsWhere_UserQueryUsesExistsSubquery(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		UserQuery: "admin@",
	}

	where, args := buildOpsErrorLogsWhere(filter)
	if where == "" {
		t.Fatalf("where should not be empty")
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if !strings.Contains(where, "EXISTS (SELECT 1 FROM users u WHERE u.id = e.user_id AND u.email ILIKE $") {
		t.Fatalf("where should include EXISTS user email condition: %s", where)
	}
}

func TestBuildOpsErrorLogsWhere_UserFiltersAndCountTokens(t *testing.T) {
	userID := int64(42)
	keyID := int64(7)
	where, args := buildOpsErrorLogsWhere(&service.OpsErrorLogFilter{
		UserID: &userID, APIKeyID: &keyID, Model: "gpt_", ModelFuzzy: true,
		ExcludeCountTokens: true, ErrorPhasesAny: []string{"upstream", "network"},
		View: "all",
	})
	for _, want := range []string{
		"e.user_id = $",
		"e.api_key_id = $",
		"ILIKE $",
		"e.is_count_tokens",
		"e.error_phase = ANY($",
	} {
		if !strings.Contains(where, want) {
			t.Fatalf("where missing %q: %s", want, where)
		}
	}
	if len(args) != 4 {
		t.Fatalf("args len=%d, want 4", len(args))
	}
}

func TestOpsErrorLogsOrderByWhitelist(t *testing.T) {
	if got := opsErrorLogsOrderBy(&service.OpsErrorLogFilter{SortBy: "model", SortOrder: "asc"}); got != "COALESCE(NULLIF(TRIM(e.requested_model), ''), e.model) ASC, e.id ASC" {
		t.Fatalf("unexpected model order: %s", got)
	}
	if got := opsErrorLogsOrderBy(&service.OpsErrorLogFilter{SortBy: "drop table", SortOrder: "asc"}); got != "e.created_at ASC, e.id ASC" {
		t.Fatalf("unexpected fallback order: %s", got)
	}
}

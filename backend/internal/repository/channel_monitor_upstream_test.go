package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/channelmonitor"
	"github.com/Wei-Shaw/sub2api/ent/channelmonitorrequesttemplate"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type monitorQuotaJSONArgument struct{}

func (monitorQuotaJSONArgument) Match(v driver.Value) bool {
	var raw []byte
	switch value := v.(type) {
	case string:
		raw = []byte(value)
	case []byte:
		raw = value
	default:
		return false
	}
	var snapshot domain.MonitorQuotaSnapshot
	return json.Unmarshal(raw, &snapshot) == nil && snapshot.Success && len(snapshot.Tiers) == 1 && snapshot.Tiers[0].UsedPercent == 42
}

func TestChannelMonitorHistoryQuotaPersistsThroughEntBuilder(t *testing.T) {
	var query string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_ string, actual string) error {
		query = actual
		return nil
	})))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	repo := NewChannelMonitorRepository(client, db)
	mock.ExpectQuery("insert quota").
		WithArgs(sqlmock.AnyArg(), "", "quota", int64(7), monitorQuotaJSONArgument{}, "operational").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(50)))
	err = repo.InsertHistoryBatch(context.Background(), []*service.ChannelMonitorHistoryRow{{
		MonitorID: 7, Model: "quota", Status: service.MonitorStatusOperational, CheckedAt: time.Now(),
		Quota: &domain.MonitorQuotaSnapshot{Source: "usage", Success: true, Tiers: []domain.MonitorQuotaTier{{Window: "5h", UsedPercent: 42}}},
	}})
	require.NoError(t, err)
	require.Contains(t, query, `"quota"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func monitorEntitySQLRow(accountID any, apiMode, checkMode string, jitter int) []driver.Value {
	now := time.Now().UTC()
	return []driver.Value{int64(88), now, now, "test", "openai", checkMode, accountID, apiMode, "",
		"NEW:", "quota", `[]`, "", true, 60, jitter, nil, int64(1), nil, `{}`, "off", nil}
}

func TestChannelMonitorRepositoryCRUDPersistsProtocolBindingJitter(t *testing.T) {
	var queries []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_ string, actual string) error {
		queries = append(queries, actual)
		return nil
	})))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	repo := NewChannelMonitorRepository(client, db)
	accountID := int64(12)
	monitor := &service.ChannelMonitor{Name: "test", Provider: "openai", APIMode: "responses", CheckMode: "quota",
		AccountID: &accountID, APIKey: "NEW:", PrimaryModel: "quota", Enabled: true,
		IntervalSeconds: 60, JitterSeconds: 4, CreatedBy: 1}
	mock.ExpectQuery("create").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "test", "openai", "quota", int64(12), "responses",
			"", "NEW:", "quota", sqlmock.AnyArg(), "", true, 60, 4, int64(1), sqlmock.AnyArg(), "off").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(88)))
	require.NoError(t, repo.Create(context.Background(), monitor))
	require.Equal(t, int64(88), monitor.ID)
	for _, field := range []string{`"api_mode"`, `"check_mode"`, `"account_id"`, `"jitter_seconds"`} {
		require.Contains(t, queries[0], field)
	}
	mock.ExpectQuery("get").WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows(channelmonitor.Columns).AddRow(monitorEntitySQLRow(int64(12), "responses", "quota", 4)...))
	loaded, err := repo.GetByID(context.Background(), 88)
	require.NoError(t, err)
	require.Equal(t, "responses", loaded.APIMode)
	require.Equal(t, "quota", loaded.CheckMode)
	require.Equal(t, int64(12), *loaded.AccountID)
	require.Equal(t, 4, loaded.JitterSeconds)
	loaded.AccountID = nil
	loaded.CheckMode = "probe"
	loaded.APIMode = "chat_completions"
	loaded.JitterSeconds = 0
	mock.ExpectBegin()
	mock.ExpectExec("update").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("updated").WillReturnRows(sqlmock.NewRows(channelmonitor.Columns).
		AddRow(monitorEntitySQLRow(nil, "chat_completions", "probe", 0)...))
	mock.ExpectCommit()
	require.NoError(t, repo.Update(context.Background(), loaded))
	require.Contains(t, strings.Join(queries, "\n"), `"account_id" = NULL`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelMonitorTemplateRepositoryProtocolRoundTripAndApplyGuard(t *testing.T) {
	var queries []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_ string, actual string) error {
		queries = append(queries, actual)
		return nil
	})))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	repo := NewChannelMonitorRequestTemplateRepository(client, db)
	tpl := &service.ChannelMonitorRequestTemplate{Name: "responses", Provider: "openai", APIMode: "responses", BodyOverrideMode: "off"}
	mock.ExpectQuery("create").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "responses", "openai", "responses", "", sqlmock.AnyArg(), "off").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))
	require.NoError(t, repo.Create(context.Background(), tpl))
	require.Contains(t, queries[0], `"api_mode"`)
	row := func() *sqlmock.Rows {
		now := time.Now()
		return sqlmock.NewRows(channelmonitorrequesttemplate.Columns).
			AddRow(int64(5), now, now, "responses", "openai", "responses", "", `{}`, "off", nil)
	}
	mock.ExpectQuery("get").WithArgs(int64(5)).WillReturnRows(row())
	loaded, err := repo.GetByID(context.Background(), 5)
	require.NoError(t, err)
	require.Equal(t, "responses", loaded.APIMode)
	mock.ExpectBegin()
	mock.ExpectExec("update template").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("updated template").WillReturnRows(row())
	mock.ExpectCommit()
	require.NoError(t, repo.Update(context.Background(), loaded))
	mock.ExpectQuery("apply template lookup").WithArgs(int64(5)).WillReturnRows(row())
	mock.ExpectExec("apply scoped protocol").WillReturnResult(sqlmock.NewResult(0, 1))
	count, err := repo.ApplyToMonitors(context.Background(), 5, []int64{88})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	query := queries[len(queries)-1]
	require.Contains(t, query, `"provider" =`)
	require.Contains(t, query, `"api_mode" =`)
	require.Contains(t, query, `"template_id" =`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelMonitorLatestAndHistoryDecodeQuota(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	repo := NewChannelMonitorRepository(client, db)
	now := time.Now().UTC()
	snapshot := `{"source":"usage","success":true,"tiers":[{"window":"5h","used_percent":42}]}`
	mock.ExpectQuery("SELECT DISTINCT ON \\(model\\)").WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"model", "status", "latency_ms", "ping_latency_ms", "checked_at", "quota"}).
			AddRow("quota", "operational", nil, nil, now, snapshot))
	latest, err := repo.ListLatestPerModel(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, latest, 1)
	require.True(t, latest[0].Quota.Success)
	require.Equal(t, float64(42), latest[0].Quota.Tiers[0].UsedPercent)
	mock.ExpectQuery("SELECT DISTINCT ON \\(monitor_id, model\\)").
		WillReturnRows(sqlmock.NewRows([]string{"monitor_id", "model", "status", "latency_ms", "ping_latency_ms", "checked_at", "quota"}).
			AddRow(7, "quota", "operational", nil, nil, now, snapshot))
	batch, err := repo.ListLatestForMonitorIDs(context.Background(), []int64{7})
	require.NoError(t, err)
	require.Equal(t, latest[0].Quota, batch[7][0].Quota)
	require.Nil(t, scanMonitorQuota(nil))
	require.Nil(t, scanMonitorQuota([]byte("invalid")))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelMonitorEntityMappingKeepsBindingsAndLegacyDefaults(t *testing.T) {
	id := int64(12)
	mapped := entToServiceMonitor(&dbent.ChannelMonitor{AccountID: &id, APIMode: "responses", CheckMode: "quota_probe", JitterSeconds: 4})
	require.Equal(t, id, *mapped.AccountID)
	require.Equal(t, "responses", mapped.APIMode)
	require.Equal(t, "quota_probe", mapped.CheckMode)
	require.Equal(t, 4, mapped.JitterSeconds)
	legacy := entToServiceMonitor(&dbent.ChannelMonitor{})
	require.Equal(t, "probe", legacy.CheckMode)
	require.Equal(t, "chat_completions", legacy.APIMode)
	require.Nil(t, legacy.AccountID)
	require.False(t, strings.Contains(legacy.APIKey, "plaintext"))
}

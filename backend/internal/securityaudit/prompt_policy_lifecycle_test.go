package securityaudit

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestPolicyLifecycleSaveDraftCommitAndConflicts(t *testing.T) {
	t.Run("commits a draft based on the locked active config", func(t *testing.T) {
		manager, mock := newPolicyLifecycleManager(t, nil)
		current := lifecycleStorageConfig(t, 7, 3)
		expectPolicyTransaction(mock, current, defaultPolicyHistory())
		mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditPolicyHistory, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		history, err := manager.SavePolicyDraft(context.Background(), PolicyDraftRequest{
			ExpectedConfigVersion: 7, ExpectedDraftVersion: 0, Rules: lifecyclePolicyRules("draft-policy"),
		}, 42)
		require.NoError(t, err)
		require.NotNil(t, history.Draft)
		require.Equal(t, 1, history.Draft.DraftVersion)
		require.Equal(t, int64(7), history.Draft.BaseConfigVersion)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rejects stale config before reading or writing draft history", func(t *testing.T) {
		manager, mock := newPolicyLifecycleManager(t, nil)
		expectPolicyTransactionStart(mock, lifecycleStorageConfig(t, 8, 3))
		mock.ExpectRollback()

		_, err := manager.SavePolicyDraft(context.Background(), PolicyDraftRequest{
			ExpectedConfigVersion: 7, ExpectedDraftVersion: 0, Rules: lifecyclePolicyRules("draft-policy"),
		}, 42)
		require.Error(t, err)
		require.Equal(t, ErrorCodeConfigConflict, infraerrors.Reason(err))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rejects stale draft without writing", func(t *testing.T) {
		manager, mock := newPolicyLifecycleManager(t, nil)
		current := lifecycleStorageConfig(t, 7, 3)
		history := defaultPolicyHistory()
		history.Draft = &PolicyDraft{DraftVersion: 3, BaseConfigVersion: 7, Rules: lifecyclePolicyRules("saved-draft")}
		expectPolicyTransaction(mock, current, history)
		mock.ExpectRollback()

		_, err := manager.SavePolicyDraft(context.Background(), PolicyDraftRequest{
			ExpectedConfigVersion: 7, ExpectedDraftVersion: 2, Rules: lifecyclePolicyRules("draft-policy"),
		}, 42)
		require.Error(t, err)
		require.Equal(t, ErrorCodeConfigConflict, infraerrors.Reason(err))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rejects a nonzero expected draft version when no draft exists", func(t *testing.T) {
		manager, mock := newPolicyLifecycleManager(t, nil)
		current := lifecycleStorageConfig(t, 7, 3)
		expectPolicyTransaction(mock, current, defaultPolicyHistory())
		mock.ExpectRollback()

		_, err := manager.SavePolicyDraft(context.Background(), PolicyDraftRequest{
			ExpectedConfigVersion: 7, ExpectedDraftVersion: 1, Rules: lifecyclePolicyRules("draft-policy"),
		}, 42)
		require.Error(t, err)
		require.Equal(t, ErrorCodeConfigConflict, infraerrors.Reason(err))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPolicyLifecyclePublishRejectsStaleAndMissingDrafts(t *testing.T) {
	tests := []struct {
		name    string
		current storageConfig
		history PolicyHistory
		req     PolicyPublishRequest
	}{
		{
			name: "stale config", current: lifecycleStorageConfig(t, 8, 3), history: policyHistoryWithDraft(7, 1),
			req: PolicyPublishRequest{ExpectedConfigVersion: 7, ExpectedDraftVersion: 1},
		},
		{
			name: "missing draft", current: lifecycleStorageConfig(t, 7, 3), history: defaultPolicyHistory(),
			req: PolicyPublishRequest{ExpectedConfigVersion: 7, ExpectedDraftVersion: 1},
		},
		{
			name: "stale draft version", current: lifecycleStorageConfig(t, 7, 3), history: policyHistoryWithDraft(7, 2),
			req: PolicyPublishRequest{ExpectedConfigVersion: 7, ExpectedDraftVersion: 1},
		},
		{
			name: "stale base config", current: lifecycleStorageConfig(t, 7, 3), history: policyHistoryWithDraft(6, 1),
			req: PolicyPublishRequest{ExpectedConfigVersion: 7, ExpectedDraftVersion: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, mock := newPolicyLifecycleManager(t, nil)
			if tt.name == "stale config" {
				expectPolicyTransactionStart(mock, tt.current)
			} else {
				expectPolicyTransaction(mock, tt.current, tt.history)
			}
			mock.ExpectRollback()

			_, err := manager.PublishPolicyDraft(context.Background(), tt.req, 42)
			require.Error(t, err)
			require.Equal(t, ErrorCodeConfigConflict, infraerrors.Reason(err))
			require.Nil(t, manager.snapshot.Load())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPolicyLifecyclePublishWriteFailureRollsBack(t *testing.T) {
	for _, secondWrite := range []bool{false, true} {
		t.Run(map[bool]string{false: "config write", true: "history write"}[secondWrite], func(t *testing.T) {
			manager, mock := newPolicyLifecycleManager(t, nil)
			current := lifecycleStorageConfig(t, 7, 3)
			expectPolicyTransaction(mock, current, policyHistoryWithDraft(7, 1))
			if secondWrite {
				mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditConfig, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditPolicyHistory, sqlmock.AnyArg()).WillReturnError(errors.New("history write failed"))
			} else {
				mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditConfig, sqlmock.AnyArg()).WillReturnError(errors.New("config write failed"))
			}
			mock.ExpectRollback()

			_, err := manager.PublishPolicyDraft(context.Background(), PolicyPublishRequest{ExpectedConfigVersion: 7, ExpectedDraftVersion: 1}, 42)
			require.Error(t, err)
			require.Nil(t, manager.snapshot.Load())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPolicyLifecyclePublishCommitControlsSnapshotAndNotification(t *testing.T) {
	t.Run("failed commit does not install active snapshot or notify redis", func(t *testing.T) {
		mini, redisClient, messages := lifecycleRedisSubscriber(t)
		defer mini.Close()
		manager, mock := newPolicyLifecycleManager(t, redisClient)
		current := lifecycleStorageConfig(t, 7, 3)
		expectPolicyPublishWrites(mock, current, policyHistoryWithDraft(7, 1))
		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		_, err := manager.PublishPolicyDraft(context.Background(), PolicyPublishRequest{ExpectedConfigVersion: 7, ExpectedDraftVersion: 1}, 42)
		require.EqualError(t, err, "commit failed")
		require.Nil(t, manager.snapshot.Load())
		assertNoLifecycleRedisMessage(t, messages)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successful commit installs snapshot then notifies redis", func(t *testing.T) {
		mini, redisClient, messages := lifecycleRedisSubscriber(t)
		defer mini.Close()
		manager, mock := newPolicyLifecycleManager(t, redisClient)
		current := lifecycleStorageConfig(t, 7, 3)
		expectPolicyPublishWritesWithPayloadAssertions(t, mock, current, policyHistoryWithDraft(7, 1))
		mock.ExpectCommit()

		published, err := manager.PublishPolicyDraft(context.Background(), PolicyPublishRequest{ExpectedConfigVersion: 7, ExpectedDraftVersion: 1}, 42)
		require.NoError(t, err)
		require.Equal(t, int64(8), published.ConfigVersion)
		active, ok := manager.Active()
		require.True(t, ok)
		require.Equal(t, int64(8), active.ConfigVersion)
		select {
		case message := <-messages:
			require.Equal(t, "8", message.Payload)
		case <-time.After(time.Second):
			t.Fatal("expected post-commit config invalidation")
		}
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPolicyLifecycleRollbackCommitsVersionAndHistory(t *testing.T) {
	mini, redisClient, messages := lifecycleRedisSubscriber(t)
	defer mini.Close()
	manager, mock := newPolicyLifecycleManager(t, redisClient)
	current := lifecycleStorageConfig(t, 7, 3)
	history := defaultPolicyHistory()
	history.ActiveVersion = 3
	history.Versions = []PolicyVersionRecord{{PolicyVersion: 2, PolicyID: "old-policy", Rules: lifecyclePolicyRules("old-policy"), ConfigVersion: 5}}
	expectPolicyTransaction(mock, current, history)
	mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditConfig, lifecycleJSONArgument(func(raw string) bool {
		next, err := ParseStorageConfig(raw)
		return err == nil && next.ConfigVersion == 8 && next.Rules.PolicyVersion == 2 && next.Rules.PolicyID == "old-policy"
	})).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditPolicyHistory, lifecycleJSONArgument(func(raw string) bool {
		persisted, err := parsePolicyHistory(raw)
		return err == nil && persisted.ActiveVersion == 2 && len(persisted.Versions) == 1 && persisted.Versions[0].RollbackCount == 1
	})).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rolledBack, err := manager.RollbackPolicy(context.Background(), 2, 7, 42)
	require.NoError(t, err)
	require.Equal(t, int64(8), rolledBack.ConfigVersion)
	require.Equal(t, 2, rolledBack.Rules.PolicyVersion)
	select {
	case message := <-messages:
		require.Equal(t, "8", message.Payload)
	case <-time.After(time.Second):
		t.Fatal("expected post-commit rollback invalidation")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPolicyLifecycleRollbackFailurePaths(t *testing.T) {
	tests := []struct {
		name    string
		current storageConfig
		history PolicyHistory
		expect  func(sqlmock.Sqlmock)
	}{
		{name: "stale config", current: lifecycleStorageConfig(t, 8, 3), history: defaultPolicyHistory(), expect: func(mock sqlmock.Sqlmock) { mock.ExpectRollback() }},
		{name: "missing target", current: lifecycleStorageConfig(t, 7, 3), history: defaultPolicyHistory(), expect: func(mock sqlmock.Sqlmock) { mock.ExpectRollback() }},
		{name: "history write failure", current: lifecycleStorageConfig(t, 7, 3), history: policyHistoryForRollback(), expect: func(mock sqlmock.Sqlmock) {
			mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditConfig, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditPolicyHistory, sqlmock.AnyArg()).WillReturnError(errors.New("history write failed"))
			mock.ExpectRollback()
		}},
		{name: "commit failure", current: lifecycleStorageConfig(t, 7, 3), history: policyHistoryForRollback(), expect: func(mock sqlmock.Sqlmock) {
			mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditConfig, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditPolicyHistory, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mini, redisClient, messages := lifecycleRedisSubscriber(t)
			defer mini.Close()
			manager, mock := newPolicyLifecycleManager(t, redisClient)
			if tt.name == "stale config" {
				expectPolicyTransactionStart(mock, tt.current)
			} else {
				expectPolicyTransaction(mock, tt.current, tt.history)
			}
			tt.expect(mock)
			_, err := manager.RollbackPolicy(context.Background(), 2, 7, 42)
			require.Error(t, err)
			require.Nil(t, manager.snapshot.Load())
			assertNoLifecycleRedisMessage(t, messages)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func newPolicyLifecycleManager(t *testing.T, redisClient *redis.Client) (*ConfigManager, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		require.NoError(t, db.Close())
	})
	return &ConfigManager{
		db: db, redis: redisClient, encryptor: prefixEncryptor{}, clock: fixedClock{now: time.Unix(1_700_000_000, 0)},
		settings: staticSettingRepository{values: map[string]string{SettingKeyRiskControl: "true"}},
	}, mock
}

func lifecycleStorageConfig(t *testing.T, configVersion int64, policyVersion int) storageConfig {
	t.Helper()
	config := DefaultStorageConfig()
	config.ConfigVersion = configVersion
	config.Rules = lifecyclePolicyRules("active-policy")
	config.Rules.PolicyVersion = policyVersion
	return config
}

func lifecyclePolicyRules(policyID string) RiskActionRules {
	return RiskActionRules{PolicyID: policyID, Defaults: map[string]RiskPolicyAction{
		"unsafe": {Action: ActionBlock, RiskLevel: RiskCritical},
	}}
}

func policyHistoryWithDraft(baseConfigVersion int64, draftVersion int) PolicyHistory {
	history := defaultPolicyHistory()
	history.Draft = &PolicyDraft{DraftVersion: draftVersion, BaseConfigVersion: baseConfigVersion, Rules: lifecyclePolicyRules("draft-policy")}
	return history
}

func policyHistoryForRollback() PolicyHistory {
	history := defaultPolicyHistory()
	history.ActiveVersion = 3
	history.Versions = []PolicyVersionRecord{{PolicyVersion: 2, PolicyID: "old-policy", Rules: lifecyclePolicyRules("old-policy"), ConfigVersion: 5}}
	return history
}

func expectPolicyTransactionStart(mock sqlmock.Sqlmock, current storageConfig) {
	raw, err := json.Marshal(current)
	if err != nil {
		panic(err)
	}
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WithArgs(promptAuditConfigLockKey).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT value FROM settings WHERE key=\\$1 FOR UPDATE").WithArgs(SettingKeyPromptAuditConfig).WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(string(raw)))
}

func expectPolicyTransaction(mock sqlmock.Sqlmock, current storageConfig, history PolicyHistory) {
	expectPolicyTransactionStart(mock, current)
	raw, err := marshalPolicyHistory(history)
	if err != nil {
		panic(err)
	}
	mock.ExpectQuery("SELECT value FROM settings WHERE key=\\$1 FOR UPDATE").WithArgs(SettingKeyPromptAuditPolicyHistory).WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(raw))
}

func expectPolicyPublishWrites(mock sqlmock.Sqlmock, current storageConfig, history PolicyHistory) {
	expectPolicyTransaction(mock, current, history)
	mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditConfig, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditPolicyHistory, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
}

func expectPolicyPublishWritesWithPayloadAssertions(t *testing.T, mock sqlmock.Sqlmock, current storageConfig, history PolicyHistory) {
	t.Helper()
	expectPolicyTransaction(mock, current, history)
	mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditConfig, lifecycleJSONArgument(func(raw string) bool {
		next, err := ParseStorageConfig(raw)
		return err == nil && next.ConfigVersion == 8 && next.Rules.PolicyVersion == 4 && next.Rules.PolicyID == "draft-policy"
	})).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO settings").WithArgs(SettingKeyPromptAuditPolicyHistory, lifecycleJSONArgument(func(raw string) bool {
		persisted, err := parsePolicyHistory(raw)
		return err == nil && persisted.Draft == nil && persisted.ActiveVersion == 4 && len(persisted.Versions) == 1 && persisted.Versions[0].ConfigVersion == 8 && persisted.Versions[0].Rules.PolicyID == "draft-policy"
	})).WillReturnResult(sqlmock.NewResult(1, 1))
}

type lifecycleJSONArgument func(string) bool

func (matches lifecycleJSONArgument) Match(value driver.Value) bool {
	raw, ok := value.(string)
	return ok && matches(raw)
}

func lifecycleRedisSubscriber(t *testing.T) (*miniredis.Miniredis, *redis.Client, <-chan *redis.Message) {
	t.Helper()
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	pubsub := client.Subscribe(context.Background(), ConfigInvalidationChannel)
	t.Cleanup(func() { require.NoError(t, pubsub.Close()) })
	_, err := pubsub.ReceiveTimeout(context.Background(), time.Second)
	require.NoError(t, err)
	return mini, client, pubsub.Channel()
}

func assertNoLifecycleRedisMessage(t *testing.T, messages <-chan *redis.Message) {
	t.Helper()
	select {
	case message := <-messages:
		t.Fatalf("unexpected pre-commit config invalidation: %q", message.Payload)
	case <-time.After(100 * time.Millisecond):
	}
}

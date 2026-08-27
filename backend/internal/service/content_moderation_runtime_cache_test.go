package service

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type countingContentModerationSettingRepo struct {
	*contentModerationTestSettingRepo
	getMultipleCalls atomic.Int64
}

func (r *countingContentModerationSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	r.getMultipleCalls.Add(1)
	return r.contentModerationTestSettingRepo.GetMultiple(ctx, keys)
}

func runtimeCacheTestConfig(t *testing.T, keywords ...string) string {
	t.Helper()
	cfg := defaultContentModerationConfig()
	cfg.Enabled = true
	cfg.Mode = ContentModerationModePreBlock
	cfg.KeywordBlockingMode = ContentModerationKeywordModeKeywordOnly
	cfg.BlockedKeywords = keywords
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	return string(raw)
}

func runtimeCacheTestInput(text string) ContentModerationCheckInput {
	return ContentModerationCheckInput{
		Protocol: ContentModerationProtocolOpenAIChat,
		Model:    "runtime-cache-test",
		Body:     []byte(`{"messages":[{"role":"user","content":"` + text + `"}]}`),
	}
}

func TestContentModerationRuntimeSnapshotCachesSettingsAndRefreshesAfterSave(t *testing.T) {
	baseRepo := &contentModerationTestSettingRepo{values: map[string]string{
		SettingKeyRiskControlEnabled:      "true",
		SettingKeyContentModerationConfig: runtimeCacheTestConfig(t, "old-keyword"),
	}}
	repo := &countingContentModerationSettingRepo{contentModerationTestSettingRepo: baseRepo}
	svc := runtimeCacheTestServiceForTest(repo, time.Hour)

	decision, err := svc.Check(context.Background(), runtimeCacheTestInput("clean prompt"))
	require.NoError(t, err)
	require.True(t, decision.Allowed)
	for range 10 {
		decision, err = svc.Check(context.Background(), runtimeCacheTestInput("clean prompt"))
		require.NoError(t, err)
		require.True(t, decision.Allowed)
	}
	require.Equal(t, int64(1), repo.getMultipleCalls.Load())

	keywords := []string{"new-keyword"}
	_, err = svc.UpdateConfig(context.Background(), UpdateContentModerationConfigInput{BlockedKeywords: &keywords})
	require.NoError(t, err)
	decision, err = svc.Check(context.Background(), runtimeCacheTestInput("new-keyword"))
	require.NoError(t, err)
	require.True(t, decision.Blocked)
	// The update replaces the in-memory snapshot, so no request-time settings read is needed.
	require.Equal(t, int64(1), repo.getMultipleCalls.Load())
}

func TestContentModerationRuntimeSnapshotConcurrentReadAndReplace(t *testing.T) {
	repo := &countingContentModerationSettingRepo{contentModerationTestSettingRepo: &contentModerationTestSettingRepo{values: map[string]string{
		SettingKeyRiskControlEnabled:      "true",
		SettingKeyContentModerationConfig: runtimeCacheTestConfig(t, "blocked"),
	}}}
	svc := runtimeCacheTestServiceForTest(repo, time.Hour)
	_, err := svc.Check(context.Background(), runtimeCacheTestInput("clean prompt"))
	require.NoError(t, err)

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				decision, checkErr := svc.Check(context.Background(), runtimeCacheTestInput("clean prompt"))
				if checkErr != nil {
					errs <- checkErr
					return
				}
				if decision == nil || !decision.Allowed {
					errs <- requireError("unexpected denied decision")
					return
				}
			}
		}()
	}
	for i := 0; i < 12; i++ {
		keywords := []string{"blocked-" + string(rune('a'+i))}
		if _, updateErr := svc.UpdateConfig(context.Background(), UpdateContentModerationConfigInput{BlockedKeywords: &keywords}); updateErr != nil {
			errs <- updateErr
		}
	}
	wg.Wait()
	close(errs)
	for checkErr := range errs {
		require.NoError(t, checkErr)
	}
}

func runtimeCacheTestServiceForTest(repo SettingRepository, ttl time.Duration) *ContentModerationService {
	return &ContentModerationService{
		settingRepo:     repo,
		repo:            &contentModerationTestRepo{},
		runtimeCacheTTL: ttl,
		keyHealth:       make(map[string]*contentModerationKeyHealth),
	}
}

type testError string

func (e testError) Error() string { return string(e) }

func requireError(message string) error { return testError(message) }

//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type updateServiceCacheStub struct {
	data string
}

func (s *updateServiceCacheStub) GetUpdateInfo(context.Context) (string, error) {
	if s.data == "" {
		return "", errors.New("cache miss")
	}
	return s.data, nil
}

func (s *updateServiceCacheStub) SetUpdateInfo(_ context.Context, data string, _ time.Duration) error {
	s.data = data
	return nil
}

type updateServiceGitHubClientStub struct {
	release *GitHubRelease
}

func (s *updateServiceGitHubClientStub) FetchLatestRelease(context.Context, string) (*GitHubRelease, error) {
	return s.release, nil
}

func (s *updateServiceGitHubClientStub) DownloadFile(context.Context, string, string, int64) error {
	panic("DownloadFile should not be called when no update is available")
}

func (s *updateServiceGitHubClientStub) FetchChecksumFile(context.Context, string) ([]byte, error) {
	panic("FetchChecksumFile should not be called when no update is available")
}

func TestUpdateServicePerformUpdateNoUpdateReturnsSentinel(t *testing.T) {
	svc := NewUpdateService(
		&updateServiceCacheStub{},
		&updateServiceGitHubClientStub{
			release: &GitHubRelease{
				TagName: "v0.1.132",
				Name:    "v0.1.132",
			},
		},
		"0.1.132",
		"release",
	)

	err := svc.PerformUpdate(context.Background())

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoUpdateAvailable))
	require.ErrorIs(t, err, ErrNoUpdateAvailable)
}

func TestCompareVersionsHyphenatedSuffix(t *testing.T) {
	if got := compareVersions("v0.1.183-custom", "v0.1.183"); got != 0 {
		t.Fatalf("equal versions got %d", got)
	}
	if compareVersions("v0.1.183-custom", "v0.1.184") >= 0 {
		t.Fatal("expected lower custom version")
	}
	if compareVersions("v0.1.184", "v0.1.183-custom") <= 0 {
		t.Fatal("expected higher version")
	}
	if got := compareVersions("v1.2.3", "v1.2.3"); got != 0 {
		t.Fatalf("ordinary equal versions got %d", got)
	}
	if compareVersions("v1.2.3", "v1.2.4") >= 0 {
		t.Fatal("expected ordinary lower version")
	}
	if compareVersions("v1.3.0", "v1.2.9") <= 0 {
		t.Fatal("expected ordinary higher version")
	}
}

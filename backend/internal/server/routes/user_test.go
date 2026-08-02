package routes

import (
	"os"
	"strings"
	"testing"
)

func TestUserUsageStaticRoutesRegisteredBeforeIDWildcard(t *testing.T) {
	source, err := os.ReadFile("user.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	dashboard := strings.Index(text, `usage.GET("/dashboard/leaderboard"`)
	dailyRewardClaim := strings.Index(text, `usage.POST("/dashboard/leaderboard/daily-reward/claim"`)
	errors := strings.Index(text, `usage.GET("/errors"`)
	errorDetail := strings.Index(text, `usage.GET("/errors/:id"`)
	snapshot := strings.Index(text, `usage.GET("/dashboard/snapshot-v2"`)
	wildcard := strings.Index(text, `usage.GET("/:id"`)
	if dashboard < 0 {
		t.Fatal("dashboard leaderboard route is not registered")
	}
	if dailyRewardClaim < 0 {
		t.Fatal("dashboard leaderboard daily reward claim route is not registered")
	}
	if errors < 0 || errorDetail < 0 {
		t.Fatal("user error routes are not registered")
	}
	if snapshot < 0 {
		t.Fatal("usage snapshot-v2 route is not registered")
	}
	if wildcard < 0 {
		t.Fatal("usage id wildcard route is not registered")
	}
	if dashboard > wildcard {
		t.Fatal("dashboard leaderboard route must be registered before /:id wildcard")
	}
	if dailyRewardClaim > wildcard {
		t.Fatal("dashboard leaderboard daily reward claim route must be registered before /:id wildcard")
	}
	if errors > wildcard || errorDetail > wildcard || snapshot > wildcard {
		t.Fatal("usage static routes must be registered before /:id wildcard")
	}
}

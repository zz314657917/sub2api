package service

import (
	"context"
	"errors"
	"testing"
)

func TestPelicanGroupProbeSingleAttemptAndRelease(t *testing.T) {
	for _, name := range []string{"success", "unavailable", "busy", "stale-member", "plan-changed", "provider-failure", "selection-error-with-lease"} {
		t.Run(name, func(t *testing.T) {
			a := reviewAccount(2, name != "stale-member")
			repo := &pelicanReviewRepo{released: make(chan struct{}, 1), canRun: map[int64]bool{2: name != "plan-changed"}}
			release := make(chan struct{})
			close(release)
			executor := &pelicanReviewExecutor{started: make(chan int64, 2), release: release}
			s := NewPelicanTestService(repo, pelicanReviewAccounts{fresh: map[int64]*Account{2: a}}, pelicanReviewGroups{}, executor)
			selected, released := 0, 0
			s.SetAccountSelector(pelicanSelectorFunc(func(_ context.Context, group int64, model string) (*AccountSelectionResult, error) {
				selected++
				if group != 7 || model != "gpt-test" {
					t.Fatalf("wrong route: %d %s", group, model)
				}
				if name == "unavailable" {
					return nil, errors.New("no capacity")
				}
				selection := &AccountSelectionResult{Account: a, Acquired: name != "busy", ReleaseFunc: func() { released++ }}
				if name == "selection-error-with-lease" {
					return selection, errors.New("selection failed")
				}
				return selection, nil
			}))
			if name == "provider-failure" {
				s.tester = pelicanRunnerTester{}
			}
			s.planGate <- struct{}{}
			s.run(context.Background(), &PelicanPlan{ID: 1, GroupID: 7, ModelID: "gpt-test", MinChars: 100, MaxResults: 20, RunGeneration: 1}, pelicanPromptSnapshot{prompt: PelicanPrompt, version: PelicanPromptVersion})
			wantCalls := 0
			if name == "success" || name == "provider-failure" {
				wantCalls = 1
			}
			if selected != 1 || repo.reserved != wantCalls || len(repo.saved) != 1 || len(repo.released) != 1 {
				t.Fatalf("select=%d reserve=%d results=%v finalized=%d", selected, repo.reserved, repo.saved, len(repo.released))
			}
			wantReleased := 1
			if name == "unavailable" {
				wantReleased = 0
			}
			if released != wantReleased {
				t.Fatalf("released=%d want=%d", released, wantReleased)
			}
			if name != "provider-failure" && len(executor.calls) != wantCalls {
				t.Fatalf("calls=%v", executor.calls)
			}
			if wantCalls == 0 && repo.saved[0].Status != "skipped" {
				t.Fatalf("result=%+v", repo.saved[0])
			}
		})
	}
}

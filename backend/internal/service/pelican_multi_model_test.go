package service

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type pelicanMultiModelExecutor struct {
	calls     []string
	cancel    context.CancelFunc
	deadlines []time.Time
}

func (e *pelicanMultiModelExecutor) RunPelicanTest(c context.Context, _ int64, model, _, _ string) (*PelicanResult, error) {
	e.calls = append(e.calls, model)
	d, _ := c.Deadline()
	e.deadlines = append(e.deadlines, d)
	if e.cancel != nil {
		e.cancel()
	}
	if len(e.calls) == 1 {
		return nil, errors.New("first-model failure")
	}
	return &PelicanResult{HTML: "<html>" + strings.Repeat("success", 30) + "</html>", CharCount: 200}, nil
}

type pelicanMultiModelQuotaRepo struct {
	*pelicanReviewRepo
	limit int
}

func (r *pelicanMultiModelQuotaRepo) ReserveAttempt(c context.Context, p, g int64) error {
	if r.reserved >= r.limit {
		return ErrPelicanQuota
	}
	return r.pelicanReviewRepo.ReserveAttempt(c, p, g)
}
func TestPelicanMultiModelInput(t *testing.T) {
	for _, tc := range []struct {
		body string
		ok   bool
		ids  []string
	}{
		{`{"model_id":"gpt-legacy"}`, true, []string{"gpt-legacy"}},
		{`{"model_id":"gpt-legacy","model_ids":[]}`, false, nil},
		{`{"model_id":"gpt-legacy","model_ids":null}`, false, nil},
		{`{"model_ids":[" gpt-a ","gpt-b"]}`, true, []string{"gpt-a", "gpt-b"}},
		{`{"model_ids":["gpt-a"," gpt-a "]}`, false, nil},
		{`{"model_ids":[" "]}`, false, nil},
		{`{"model_ids":["1","2","3","4","5","6","7","8","9","10","11"]}`, false, nil},
	} {
		t.Run(tc.body, func(t *testing.T) {
			var i PelicanPlanInput
			if err := json.Unmarshal([]byte(tc.body), &i); err != nil {
				t.Fatal(err)
			}
			err := normalizePelicanInput(&i)
			if (err == nil) != tc.ok {
				t.Fatalf("normalize=%v", err)
			}
			if tc.ok && (!reflect.DeepEqual(i.ModelIDs, tc.ids) || i.ModelID != tc.ids[0]) {
				t.Fatalf("%+v", i)
			}
		})
	}
}
func TestPelicanMultiModelRun(t *testing.T) {
	for _, tc := range []struct {
		name      string
		quota     int
		cancel    bool
		wantCalls int
	}{
		{"failure-then-success", 3, false, 3}, {"quota-stops-upstream", 1, false, 1}, {"parent-cancel-stops", 3, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &pelicanMultiModelQuotaRepo{pelicanReviewRepo: &pelicanReviewRepo{released: make(chan struct{}, 1)}, limit: tc.quota}
			a := reviewAccount(1, true)
			ex := &pelicanMultiModelExecutor{}
			c, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancel {
				ex.cancel = cancel
			}
			s := NewPelicanTestService(r, pelicanReviewAccounts{fresh: map[int64]*Account{1: a}}, pelicanReviewGroups{}, ex)
			selections, releases := 0, 0
			s.SetAccountSelector(pelicanSelectorFunc(func(ctx context.Context, _ int64, model string) (*AccountSelectionResult, error) {
				if selections != releases {
					t.Fatalf("prior model slot not released: %d/%d", selections, releases)
				}
				selections++
				return &AccountSelectionResult{Account: a, Acquired: true, ReleaseFunc: func() { releases++ }}, nil
			}))
			s.planGate <- struct{}{}
			s.run(c, &PelicanPlan{ID: 1, GroupID: 7, ModelID: "gpt-a", ModelIDs: []string{"gpt-a", "gpt-b", "gpt-c"}, MinChars: 100, MaxResults: 2, TimeoutSeconds: 30, RunGeneration: 1}, pelicanPromptSnapshot{})
			if len(ex.calls) != tc.wantCalls || r.reserved != tc.wantCalls || selections != releases {
				t.Fatalf("calls=%v reserved=%d selection/release=%d/%d", ex.calls, r.reserved, selections, releases)
			}
			if !tc.cancel && tc.quota == 3 {
				if r.saved[0].Status != "failed" || r.saved[1].Status != "success" {
					t.Fatalf("results=%+v", r.saved)
				}
				for j := 1; j < len(ex.deadlines); j++ {
					if ex.deadlines[j].Before(ex.deadlines[j-1]) {
						t.Fatal("later model inherited earlier deadline")
					}
				}
			}
		})
	}
}

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func pelicanReferenceFixture(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/pelican-radar-valid.json")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func pelicanReferenceMutate(t *testing.T, mutate func(map[string]any)) []byte {
	t.Helper()
	var source map[string]any
	if err := json.Unmarshal(pelicanReferenceFixture(t), &source); err != nil {
		t.Fatal(err)
	}
	mutate(source)
	body, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestPelicanReferenceSourceOracleAndDynamicWeights(t *testing.T) {
	snapshot, err := decodePelicanReferenceSnapshot(pelicanReferenceFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	score := snapshot.scores[pelicanReferenceKey{"gpt-6.1-sol", "low"}]
	// Independent numeric oracle from the source's final score (69, not raw 72).
	if score == nil || math.Abs(*score-106.16576086956522) > 1e-10 {
		t.Fatalf("score=%v", score)
	}
	zero := snapshot.scores[pelicanReferenceKey{"gpt-6.1-sol", "max"}]
	if zero == nil || *zero != 0 {
		t.Fatalf("zero/false must be recorded: %v", zero)
	}
	if snapshot.round != "codex-four-261-10c-20261029" || snapshot.started.Hour() != 8 || snapshot.completed.Hour() != 12 ||
		snapshot.sourceAt.Format(time.RFC3339Nano) != "2026-10-29T12:00:25.692213Z" {
		t.Fatalf("metadata=%+v", snapshot)
	}
	body := pelicanReferenceMutate(t, func(v map[string]any) {
		v["score_max"] = 200
		v["task_weights"] = map[string]any{"combined": 100, "pompeii": 50, "pelican": 25, "candy": 25}
	})
	dynamic, err := decodePelicanReferenceSnapshot(body)
	if err != nil || dynamic.scoreMax != 200 || math.Abs(*dynamic.scores[pelicanReferenceKey{"gpt-6.1-sol", "low"}]-140.96195652173913) > 1e-10 {
		t.Fatalf("dynamic score=%+v err=%v", dynamic, err)
	}
}

func TestPelicanReferenceSourceRejectsMalformedSnapshot(t *testing.T) {
	row := func(v map[string]any) map[string]any { return v["rows"].([]any)[0].(map[string]any) }
	cases := map[string]func(map[string]any){
		"schema":               func(v map[string]any) { v["schema"] = "future/2" },
		"protocol":             func(v map[string]any) { v["protocol"] = "unknown" },
		"round missing":        func(v map[string]any) { delete(v, "round_id") },
		"round control":        func(v map[string]any) { v["round_id"] = "x\nx" },
		"snapshot missing":     func(v map[string]any) { delete(v, "score_snapshot_at") },
		"date invalid":         func(v map[string]any) { v["score_snapshot_at"] = "2026-10-29" },
		"date hour not padded": func(v map[string]any) { v["score_snapshot_at"] = "2026-10-29T8:00:00Z" },
		"date decimal comma":   func(v map[string]any) { v["score_snapshot_at"] = "2026-10-29T08:00:00,1Z" },
		"date invalid offset":  func(v map[string]any) { v["score_snapshot_at"] = "2026-10-29T08:00:00+24:00" },
		"interval reversed": func(v map[string]any) {
			v["author_test_interval"].(map[string]any)["completed_at"] = "2026-10-28T12:00:00Z"
		},
		"maximum null":                  func(v map[string]any) { v["score_max"] = nil },
		"maximum nonnumeric":            func(v map[string]any) { v["score_max"] = "150" },
		"weight missing":                func(v map[string]any) { delete(v["task_weights"].(map[string]any), "candy") },
		"weight null":                   func(v map[string]any) { v["task_weights"].(map[string]any)["candy"] = nil },
		"weight zero":                   func(v map[string]any) { v["task_weights"].(map[string]any)["candy"] = 0 },
		"weight negative":               func(v map[string]any) { v["task_weights"].(map[string]any)["candy"] = -22.5 },
		"weight wrong sum":              func(v map[string]any) { v["task_weights"].(map[string]any)["candy"] = 20 },
		"weight extra":                  func(v map[string]any) { v["task_weights"].(map[string]any)["extra"] = 1 },
		"models empty":                  func(v map[string]any) { v["models"] = []string{} },
		"models duplicate":              func(v map[string]any) { v["models"] = []string{"gpt-6.1-sol", "gpt-6.1-sol"} },
		"models too long":               func(v map[string]any) { v["models"] = []string{strings.Repeat("x", 201)} },
		"name missing":                  func(v map[string]any) { delete(v["model_names"].(map[string]any), "gpt-6-luna") },
		"name duplicate":                func(v map[string]any) { v["model_names"].(map[string]any)["gpt-6-luna"] = "GPT-6.1 Sol" },
		"name canonical conflict":       func(v map[string]any) { v["model_names"].(map[string]any)["gpt-6-luna"] = "gpt-6.1-sol" },
		"efforts empty":                 func(v map[string]any) { v["efforts"] = []string{} },
		"efforts duplicate":             func(v map[string]any) { v["efforts"] = []string{"low", "low"} },
		"effort invalid":                func(v map[string]any) { v["efforts"] = []string{"LOW"} },
		"model efforts unknown":         func(v map[string]any) { v["model_efforts"].(map[string]any)["gpt-6-luna"] = []string{"medium"} },
		"model efforts duplicate":       func(v map[string]any) { v["model_efforts"].(map[string]any)["gpt-6-luna"] = []string{"low", "low"} },
		"row undeclared model":          func(v map[string]any) { row(v)["model"] = "unknown" },
		"row undeclared effort":         func(v map[string]any) { row(v)["effort"] = "medium" },
		"row duplicate":                 func(v map[string]any) { v["rows"] = append(v["rows"].([]any), row(v)) },
		"rows missing":                  func(v map[string]any) { delete(v, "rows") },
		"task wrong string":             func(v map[string]any) { row(v)["combined"] = "69" },
		"task wrong array":              func(v map[string]any) { row(v)["candy"] = []any{} },
		"task wrong bool":               func(v map[string]any) { row(v)["pompeii"] = false },
		"score string":                  func(v map[string]any) { row(v)["combined"].(map[string]any)["score"] = "69" },
		"score over max":                func(v map[string]any) { row(v)["pelican"].(map[string]any)["score"] = 100.1 },
		"score negative":                func(v map[string]any) { row(v)["pompeii"].(map[string]any)["score100"] = -1 },
		"correct number":                func(v map[string]any) { row(v)["candy"].(map[string]any)["correct"] = 1 },
		"correct string":                func(v map[string]any) { row(v)["candy"].(map[string]any)["correct"] = "true" },
		"missing one malformed another": func(v map[string]any) { row(v)["combined"] = nil; row(v)["candy"] = false },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := decodePelicanReferenceSnapshot(pelicanReferenceMutate(t, mutate)); err == nil {
				t.Fatal("malformed snapshot accepted")
			}
		})
	}
	for _, raw := range []string{
		`{"schema":"codex-four-task-results/1","schema":"codex-four-task-results/1"}`,
		`null`, `{`, `[]`, string(pelicanReferenceFixture(t)) + `{}`,
		strings.Replace(string(pelicanReferenceFixture(t)), `"score": 69`, `"score": 1e999`, 1),
	} {
		if _, err := decodePelicanReferenceSnapshot([]byte(raw)); err == nil {
			t.Fatalf("invalid JSON accepted: %.80s", raw)
		}
	}
}

func TestPelicanReferenceMissingPartsNeverZeroFilled(t *testing.T) {
	for _, task := range []string{"combined", "pompeii", "pelican", "candy"} {
		for _, missing := range []string{"object missing", "object null", "field missing", "field null"} {
			t.Run(task+"/"+missing, func(t *testing.T) {
				body := pelicanReferenceMutate(t, func(v map[string]any) {
					row := v["rows"].([]any)[0].(map[string]any)
					field := "score"
					if task == "pompeii" {
						field = "score100"
					}
					if task == "candy" {
						field = "correct"
					}
					switch missing {
					case "object missing":
						delete(row, task)
					case "object null":
						row[task] = nil
					case "field missing":
						delete(row[task].(map[string]any), field)
					case "field null":
						row[task].(map[string]any)[field] = nil
					}
				})
				snapshot, err := decodePelicanReferenceSnapshot(body)
				if err != nil || snapshot.scores[pelicanReferenceKey{"gpt-6.1-sol", "low"}] != nil {
					t.Fatalf("missing part got a score: snapshot=%+v err=%v", snapshot, err)
				}
			})
		}
	}
}

type pelicanReferenceTransport func(*http.Request) (*http.Response, error)

func (f pelicanReferenceTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func pelicanReferenceResponse(body []byte) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(body))}
}

func pelicanReferenceTestCache(t *testing.T) (*pelicanReferenceCache, *atomic.Int64) {
	t.Helper()
	cache := newPelicanReferenceCache()
	clock := &atomic.Int64{}
	clock.Store(time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC).UnixNano())
	cache.now = func() time.Time { return time.Unix(0, clock.Load()).UTC() }
	return cache, clock
}

func pelicanReferenceAwait(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for reference fetch")
	}
}

func pelicanReferencePending(c *pelicanReferenceCache) <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pending
}

func TestPelicanReferenceCacheFirstDedupAndCancellation(t *testing.T) {
	body := pelicanReferenceFixture(t)
	cache, _ := pelicanReferenceTestCache(t)
	var requests atomic.Int64
	started, release := make(chan struct{}), make(chan struct{})
	cache.client.Transport = pelicanReferenceTransport(func(r *http.Request) (*http.Response, error) {
		if requests.Add(1) == 1 {
			close(started)
		}
		if r.URL.String() != pelicanReferenceDataURL || r.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected upstream request: %s", r.URL)
		}
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > pelicanReferenceTimeout {
			t.Error("fetch lacks independent bounded deadline")
		}
		select {
		case <-release:
			return pelicanReferenceResponse(body), nil
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan struct{})
	go func() { cache.get(ctx); close(first) }()
	pelicanReferenceAwait(t, started)
	cancel()
	pelicanReferenceAwait(t, first)
	var wg sync.WaitGroup
	for range 24 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			snapshot, status := cache.get(context.Background())
			if snapshot == nil || status != "ready" {
				t.Errorf("snapshot=%v status=%s", snapshot, status)
			}
		}()
	}
	close(release)
	wg.Wait()
	if requests.Load() != 1 {
		t.Fatalf("cold requests=%d", requests.Load())
	}
}

func TestPelicanReferenceCacheTTLStaleBackoffAndRecovery(t *testing.T) {
	body := pelicanReferenceFixture(t)
	cache, clock := pelicanReferenceTestCache(t)
	var requests atomic.Int64
	var bad atomic.Bool
	started := make(chan struct{}, 4)
	release := make(chan struct{}, 4)
	cache.client.Transport = pelicanReferenceTransport(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		started <- struct{}{}
		<-release
		if r.Context().Err() != nil {
			t.Error("caller canceled shared refresh")
		}
		if bad.Load() {
			return pelicanReferenceResponse([]byte(`{"schema":"malformed"}`)), nil
		}
		return pelicanReferenceResponse(body), nil
	})
	release <- struct{}{}
	initial, state := cache.get(context.Background())
	if initial == nil || state != "ready" {
		t.Fatalf("initial=%v state=%s", initial, state)
	}
	pelicanReferenceAwait(t, started)
	// Independent user-duration oracle: never derive this boundary from the TTL constant.
	clock.Add(int64(2*time.Hour - time.Nanosecond))
	if got, state := cache.get(context.Background()); got != initial || state != "ready" || requests.Load() != 1 {
		t.Fatalf("cache expired early: state=%s requests=%d", state, requests.Load())
	}
	clock.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	for range 30 {
		got, state := cache.get(ctx)
		if got != initial || state != "stale" {
			t.Fatalf("expired cache blocked/replaced: %s", state)
		}
	}
	pelicanReferenceAwait(t, started)
	cancel()
	if requests.Load() != 2 {
		t.Fatalf("refresh requests=%d", requests.Load())
	}
	bad.Store(true)
	pending := pelicanReferencePending(cache)
	release <- struct{}{}
	pelicanReferenceAwait(t, pending)
	for range 8 {
		if got, state := cache.get(context.Background()); got != initial || state != "stale" {
			t.Fatal("malformed refresh replaced last good snapshot")
		}
	}
	clock.Add(int64(pelicanReferenceBackoff - time.Nanosecond))
	cache.get(context.Background())
	if requests.Load() != 2 {
		t.Fatal("failed refresh ignored backoff")
	}
	bad.Store(false)
	clock.Add(1)
	cache.get(context.Background())
	pelicanReferenceAwait(t, started)
	pending = pelicanReferencePending(cache)
	release <- struct{}{}
	pelicanReferenceAwait(t, pending)
	recovered, state := cache.get(context.Background())
	if recovered == initial || state != "ready" || !recovered.syncedAt.Equal(cache.now()) {
		t.Fatalf("recovery failed: state=%s", state)
	}
}

func TestPelicanReferenceCacheInitialFailureAndFetchBoundaries(t *testing.T) {
	for _, failure := range []string{"network", "status", "oversize", "malformed", "redirect"} {
		t.Run(failure, func(t *testing.T) {
			cache, clock := pelicanReferenceTestCache(t)
			var requests atomic.Int64
			cache.client.Transport = pelicanReferenceTransport(func(*http.Request) (*http.Response, error) {
				requests.Add(1)
				switch failure {
				case "network":
					return nil, errors.New("secret-provider-key")
				case "status":
					resp := pelicanReferenceResponse([]byte("secret-provider-key"))
					resp.StatusCode = 503
					return resp, nil
				case "oversize":
					return pelicanReferenceResponse(bytes.Repeat([]byte(" "), pelicanReferenceBodyLimit+1)), nil
				case "redirect":
					resp := pelicanReferenceResponse(nil)
					resp.StatusCode = 302
					resp.Header.Set("Location", "https://example.com/secret")
					return resp, nil
				default:
					return pelicanReferenceResponse([]byte("bad")), nil
				}
			})
			if snapshot, state := cache.get(context.Background()); snapshot != nil || state != "unavailable" {
				t.Fatalf("failed initial fetch: state=%s", state)
			}
			for range 10 {
				cache.get(context.Background())
			}
			if requests.Load() != 1 {
				t.Fatalf("failure storm requests=%d", requests.Load())
			}
			clock.Add(int64(pelicanReferenceBackoff))
			cache.get(context.Background())
			if requests.Load() != 2 {
				t.Fatalf("backoff retry requests=%d", requests.Load())
			}
		})
	}
}

func TestPelicanReferenceProjectionSettingsAndCatalogClone(t *testing.T) {
	body := pelicanReferenceFixture(t)
	cache, _ := pelicanReferenceTestCache(t)
	var requests atomic.Int64
	cache.client.Transport = pelicanReferenceTransport(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return pelicanReferenceResponse(body), nil
	})
	s := NewPelicanTestService(nil, nil, nil, nil)
	s.reference = cache
	repo := &pelicanSettingsRepoStub{value: `{"display_name":"x","prompt":"secret-prompt","enabled":true}`}
	s.SetSettingsRepository(repo)
	if disabled, err := s.ReferenceScores(context.Background(), PelicanAuthorization{}); err != nil || disabled.Status != "disabled" || requests.Load() != 0 {
		t.Fatalf("disabled fetch: %+v %v", disabled, err)
	}
	catalog, err := s.ReferenceCatalog(context.Background())
	if err != nil || catalog.Status != "ready" || len(catalog.Models) != 2 || requests.Load() != 1 {
		t.Fatalf("admin explicit catalog while disabled=%+v err=%v", catalog, err)
	}
	catalog.Models[0].Efforts[0] = "mutated"
	catalog.Models[0].DisplayName = "mutated"
	selections := []PelicanReferenceSelection{
		{"GPT-6.1 Sol", "low"}, {"gpt-6-luna", "ultra"}, {"gpt-6-luna", "low"},
		{"gpt-6-luna", "high"}, {"GPT-6.1 SOL", "low"}, {"gpt-6.1-sol-20261029", "low"}, {"gpt-6.1-sol", "max"},
	}
	updated, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{
		DisplayName: "x", Prompt: "secret-prompt", Radar: &PelicanReferenceSettings{Enabled: true, Selections: selections},
	})
	if err != nil || updated.Radar.Selections[0].Model != "gpt-6.1-sol" {
		t.Fatalf("update=%+v err=%v", updated, err)
	}
	repo.value = repo.set
	scores, err := s.ReferenceScores(context.Background(), PelicanAuthorization{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"available", "unsupported", "not_recorded", "not_recorded", "not_recorded", "not_recorded", "available"}
	for i, status := range want {
		if scores.Items[i].Status != status || (status == "available") != (scores.Items[i].Score != nil) {
			t.Fatalf("item %d=%+v want=%s", i, scores.Items[i], status)
		}
	}
	if scores.Items[0].DisplayName != "GPT-6.1 Sol" || scores.Items[6].Score == nil || *scores.Items[6].Score != 0 {
		t.Fatalf("bad projection=%+v", scores)
	}
	*scores.Items[0].Score = -99
	*scores.RoundID = "mutated"
	*scores.SyncedAt = time.Time{}
	*scores.ScoreMax = -1
	again, _ := s.ReferenceScores(context.Background(), PelicanAuthorization{})
	catalog, _ = s.ReferenceCatalog(context.Background())
	if *again.Items[0].Score < 0 || *again.RoundID == "mutated" || again.SyncedAt.IsZero() || *again.ScoreMax != 150 || catalog.Models[0].Efforts[0] != "low" ||
		requests.Load() != 1 {
		t.Fatal("exported mutable data contaminated shared cache")
	}
	// Older PUT preserves the pointer/object; an explicit empty array clears.
	legacy, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "renamed", Prompt: "secret-prompt"})
	if err != nil || !reflect.DeepEqual(legacy.Radar, updated.Radar) {
		t.Fatalf("legacy PUT lost radar: %+v %v", legacy, err)
	}
	repo.value = repo.set
	cleared, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "x", Prompt: "x", Radar: &PelicanReferenceSettings{Enabled: true}})
	if err != nil || len(cleared.Radar.Selections) != 0 || cleared.Radar.Selections == nil {
		t.Fatalf("clear=%+v %v", cleared, err)
	}
	repo.value = `{"display_name":"x","prompt":"x","enabled":false,"radar":{"enabled":true,"selections":[]}}`
	if _, err := s.ReferenceScores(context.Background(), PelicanAuthorization{}); !errors.Is(err, ErrPelicanDisabled) {
		t.Fatalf("disabled plaza err=%v", err)
	}
	if _, err := s.ReferenceScores(context.Background(), PelicanAuthorization{Admin: true}); err != nil {
		t.Fatalf("disabled plaza admin err=%v", err)
	}
}

func TestPelicanReferenceSettingsInvalidAndOrder(t *testing.T) {
	snapshot, err := decodePelicanReferenceSnapshot(pelicanReferenceFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for name, selections := range map[string][]PelicanReferenceSelection{
		"blank": {{"", "low"}}, "spaces": {{" x ", "low"}}, "control": {{"x\nx", "low"}}, "long": {{strings.Repeat("x", 201), "low"}},
		"effort": {{"x", "LOW"}}, "duplicate": {{"x", "low"}, {"x", "low"}},
		"alias duplicate": {{"GPT-6.1 Sol", "low"}, {"gpt-6.1-sol", "low"}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := normalizePelicanReferenceSettings(&PelicanReferenceSettings{Selections: selections}, snapshot.aliases); !errors.Is(err, ErrPelicanInvalid) {
				t.Fatalf("invalid settings err=%v", err)
			}
		})
	}
	order := []PelicanReferenceSelection{{"custom-model", "max"}, {"GPT-6.1 Sol", "low"}}
	got, err := normalizePelicanReferenceSettings(&PelicanReferenceSettings{Selections: order}, snapshot.aliases)
	if err != nil || got.Selections[0] != order[0] || got.Selections[1].Model != "gpt-6.1-sol" || order[1].Model != "GPT-6.1 Sol" {
		t.Fatalf("order/input mutation: %+v err=%v", got, err)
	}
}

func TestPelicanReferenceSettingsSelectionBoundCompatibility(t *testing.T) {
	for _, count := range []int{20, 64, 65} {
		t.Run(fmt.Sprintf("pairs_%d", count), func(t *testing.T) {
			selections := make([]PelicanReferenceSelection, count)
			for i := range selections {
				selections[i] = PelicanReferenceSelection{Model: fmt.Sprintf("custom-model-%d", i), Effort: "high"}
				if !pelicanReferenceIdentity(selections[i].Model) || !pelicanReferenceEffort(selections[i].Effort) {
					t.Fatalf("invalid boundary fixture: %+v", selections[i])
				}
			}
			input := &PelicanReferenceSettings{Enabled: true, Selections: selections}
			got, err := normalizePelicanReferenceSettings(input, nil)
			if count > 64 {
				if !errors.Is(err, ErrPelicanInvalid) || got != nil {
					t.Fatalf("65 distinct valid pairs accepted: got=%+v err=%v", got, err)
				}
			} else if err != nil || !reflect.DeepEqual(got, input) {
				t.Fatalf("%d valid pairs changed/rejected: got=%+v err=%v", count, got, err)
			}
		})
	}
}

func TestPelicanReferenceUnavailableDTO(t *testing.T) {
	s := NewPelicanTestService(nil, nil, nil, nil)
	repo := &pelicanSettingsRepoStub{value: `{"display_name":"x","prompt":"secret-prompt","radar":{"enabled":true,"selections":[{"model":"custom-model","effort":"low"}]}}`}
	s.SetSettingsRepository(repo)
	var requests atomic.Int64
	s.reference.client.Transport = pelicanReferenceTransport(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("secret-provider-key")
	})
	for range 2 {
		got, err := s.ReferenceScores(context.Background(), PelicanAuthorization{})
		if err != nil || got.Status != "unavailable" || !got.Enabled || len(got.Items) != 1 ||
			got.Items[0].Status != "unavailable" || got.Items[0].Score != nil || got.RoundID != nil || got.ScoreMax != nil ||
			got.TestStartedAt != nil || got.TestCompletedAt != nil || got.SourceSnapshotAt != nil || got.SyncedAt != nil {
			t.Fatalf("unavailable DTO=%+v err=%v", got, err)
		}
		body, _ := json.Marshal(got)
		if strings.Contains(string(body), "secret-") || got.SourceURL != "https://codexradar.com/#stationmaster-round" {
			t.Fatalf("unavailable leaked internal content: %s", body)
		}
	}
	if requests.Load() != 1 {
		t.Fatalf("unavailable request storm=%d", requests.Load())
	}
}

func TestPelicanReferenceSettingsColdAliasesAndSourceEvolution(t *testing.T) {
	s := NewPelicanTestService(nil, nil, nil, nil)
	repo := &pelicanSettingsRepoStub{value: `{"display_name":"x","prompt":"unchanged-prompt"}`}
	s.SetSettingsRepository(repo)
	s.reference.client.Transport = pelicanReferenceTransport(func(*http.Request) (*http.Response, error) {
		t.Error("settings validation must not fetch")
		return nil, errors.New("unexpected fetch")
	})
	_, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{
		DisplayName: "x", Prompt: "unchanged-prompt",
		Radar: &PelicanReferenceSettings{Enabled: true, Selections: []PelicanReferenceSelection{
			{"GPT-6.1 Sol", "max"}, {"gpt-6.1-sol", "max"},
		}},
	})
	if !errors.Is(err, ErrPelicanInvalid) || repo.sets != 0 {
		t.Fatalf("cold alias duplicate accepted: err=%v writes=%d", err, repo.sets)
	}
	// These were distinct custom values at save time. A newly published label
	// may map one to the other, but ordinary settings/prompt reads stay valid.
	repo.value = `{"display_name":"x","prompt":"unchanged-prompt","radar":{"enabled":true,"selections":[{"model":"New source label","effort":"max"},{"model":"gpt-6.1-sol","effort":"max"}]}}`
	snapshot, err := decodePelicanReferenceSnapshot(pelicanReferenceMutate(t, func(v map[string]any) {
		v["model_names"].(map[string]any)["gpt-6.1-sol"] = "New source label"
	}))
	if err != nil {
		t.Fatal(err)
	}
	s.reference.snapshot = snapshot
	settings, err := s.GetSettings(context.Background())
	if err != nil || settings.Prompt != "unchanged-prompt" || settings.Radar.Selections[0].Model != "New source label" {
		t.Fatalf("source evolution broke plaza settings: %+v %v", settings, err)
	}
}

// Explicit opt-in evidence against the captured upstream asset. Default tests
// stay offline and use only the small deterministic fixture.
func TestPelicanReferenceCapturedSource(t *testing.T) {
	path := os.Getenv("PELICAN_REFERENCE_SOURCE_FIXTURE")
	if path == "" {
		t.Skip("set PELICAN_REFERENCE_SOURCE_FIXTURE for captured source evidence")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := decodePelicanReferenceSnapshot(body)
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		ScoreMax float64            `json:"score_max"`
		Weights  map[string]float64 `json:"task_weights"`
		Rows     []struct {
			Model    string `json:"model"`
			Effort   string `json:"effort"`
			Combined struct {
				Score float64 `json:"score"`
			} `json:"combined"`
			Pompeii struct {
				Score float64 `json:"score100"`
			} `json:"pompeii"`
			Pelican struct {
				Score float64 `json:"score"`
			} `json:"pelican"`
			Candy struct {
				Correct bool `json:"correct"`
			} `json:"candy"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(body, &oracle); err != nil {
		t.Fatal(err)
	}
	for _, row := range oracle.Rows {
		want := row.Combined.Score * oracle.Weights["combined"] / 100
		want += row.Pompeii.Score * oracle.Weights["pompeii"] / 100
		want += row.Pelican.Score * oracle.Weights["pelican"] / 100
		if row.Candy.Correct {
			want += oracle.Weights["candy"]
		}
		got := snapshot.scores[pelicanReferenceKey{row.Model, row.Effort}]
		if got == nil || math.Abs(*got-want) > 1e-10 {
			t.Fatalf("%s/%s score=%v want=%.12f", row.Model, row.Effort, got, want)
		}
	}
	if len(snapshot.scores) != len(oracle.Rows) || snapshot.scoreMax != oracle.ScoreMax {
		t.Fatal("source catalog size/maximum mismatch")
	}
	t.Logf("captured source: models=%d rows=%d score_max=%g round=%s", len(snapshot.models), len(snapshot.scores), snapshot.scoreMax, snapshot.round)
}

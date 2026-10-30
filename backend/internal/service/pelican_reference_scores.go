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
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	pelicanReferenceDataURL   = "https://codexradar.com/data/codex-four-task-results.json"
	PelicanReferenceSourceURL = "https://codexradar.com/#stationmaster-round"
	pelicanReferenceTTL       = 30 * time.Minute
	pelicanReferenceBackoff   = time.Minute
	pelicanReferenceTimeout   = 10 * time.Second
	pelicanReferenceBodyLimit = 2 << 20
)

type PelicanReferenceSelection struct {
	Model  string `json:"model"`
	Effort string `json:"effort"`
}

type PelicanReferenceSettings struct {
	Enabled    bool                        `json:"enabled"`
	Selections []PelicanReferenceSelection `json:"selections"`
}

type PelicanReferenceItem struct {
	Model       string   `json:"model"`
	DisplayName string   `json:"display_name"`
	Effort      string   `json:"effort"`
	Status      string   `json:"status"`
	Score       *float64 `json:"score"`
}

type PelicanReferenceScores struct {
	Enabled          bool                   `json:"enabled"`
	Status           string                 `json:"status"`
	SourceURL        string                 `json:"source_url"`
	TestStartedAt    *time.Time             `json:"test_started_at"`
	TestCompletedAt  *time.Time             `json:"test_completed_at"`
	SourceSnapshotAt *time.Time             `json:"source_snapshot_at"`
	SyncedAt         *time.Time             `json:"synced_at"`
	RoundID          *string                `json:"round_id"`
	ScoreMax         *float64               `json:"score_max"`
	Items            []PelicanReferenceItem `json:"items"`
}

type PelicanReferenceModel struct {
	Model       string   `json:"model"`
	DisplayName string   `json:"display_name"`
	Efforts     []string `json:"efforts"`
}

type PelicanReferenceCatalog struct {
	Models   []PelicanReferenceModel `json:"models"`
	Status   string                  `json:"status"`
	SyncedAt *time.Time              `json:"synced_at"`
}

type pelicanReferenceKey struct{ model, effort string }

// Every map and slice in a published snapshot is immutable.
type pelicanReferenceSnapshot struct {
	models                                 []PelicanReferenceModel
	names                                  map[string]string
	aliases                                map[string]string
	efforts                                map[pelicanReferenceKey]bool
	scores                                 map[pelicanReferenceKey]*float64
	started, completed, sourceAt, syncedAt time.Time
	round                                  string
	scoreMax                               float64
}

type pelicanReferenceCache struct {
	mu          sync.Mutex
	snapshot    *pelicanReferenceSnapshot
	pending     chan struct{}
	nextAttempt time.Time
	failed      bool
	client      *http.Client
	now         func() time.Time
}

func newPelicanReferenceCache() *pelicanReferenceCache {
	return &pelicanReferenceCache{
		client: &http.Client{
			Timeout:       pelicanReferenceTimeout,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		now: time.Now,
	}
}

func (s *PelicanTestService) referenceCache() *pelicanReferenceCache {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reference == nil {
		s.reference = newPelicanReferenceCache()
	}
	return s.reference
}

func pelicanReferenceIdentity(value string) bool {
	return value != "" && len(value) <= 200 && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value && !strings.ContainsFunc(value, unicode.IsControl)
}

func pelicanReferenceEffort(value string) bool {
	switch value {
	case "low", "medium", "high", "xhigh", "max", "ultra":
		return true
	default:
		return false
	}
}

func normalizePelicanReferenceSettings(input *PelicanReferenceSettings, aliases map[string]string) (*PelicanReferenceSettings, error) {
	if input == nil {
		return nil, nil
	}
	if len(input.Selections) > 64 {
		return nil, ErrPelicanInvalid
	}
	result := &PelicanReferenceSettings{Enabled: input.Enabled, Selections: make([]PelicanReferenceSelection, 0, len(input.Selections))}
	seen := make(map[pelicanReferenceKey]bool, len(input.Selections))
	for _, item := range input.Selections {
		if !pelicanReferenceIdentity(item.Model) || !pelicanReferenceEffort(item.Effort) {
			return nil, ErrPelicanInvalid
		}
		if canonical, ok := aliases[item.Model]; ok {
			item.Model = canonical
		}
		key := pelicanReferenceKey{item.Model, item.Effort}
		if seen[key] {
			return nil, ErrPelicanInvalid
		}
		seen[key] = true
		result.Selections = append(result.Selections, item)
	}
	return result, nil
}

func (c *pelicanReferenceCache) aliases() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	// No fetching from settings reads or writes. Exact aliases come only from
	// the reviewed source model_names, then the latest validated snapshot.
	// This small cold-start list prevents alias/canonical duplicates in a PUT
	// before an administrator has explicitly fetched the catalog.
	aliases := map[string]string{
		"GPT-6.1 Sol":   "gpt-6.1-sol",
		"GPT-6 Astra":   "gpt-6-astra",
		"GPT-6.1 Terra": "gpt-6.1-terra",
		"GPT-6 Sol":     "gpt-6-sol",
		"GPT-6 Luna":    "gpt-6-luna",
		"GPT-5.6 Sol":   "gpt-5.6-sol",
		"GPT-5.6 Terra": "gpt-5.6-terra",
		"GPT-5.6 Luna":  "gpt-5.6-luna",
	}
	if c.snapshot != nil {
		for label, model := range c.snapshot.aliases {
			aliases[label] = model
		}
	}
	return aliases
}

// get waits only for a cold cache. A caller cancellation never cancels the
// bounded shared request, including the first request after process startup.
func (c *pelicanReferenceCache) get(ctx context.Context) (*pelicanReferenceSnapshot, string) {
	for {
		c.mu.Lock()
		now := c.now()
		snapshot := c.snapshot
		if snapshot != nil {
			stale := c.failed || !now.Before(snapshot.syncedAt.Add(pelicanReferenceTTL))
			if stale && c.pending == nil && !now.Before(c.nextAttempt) {
				c.startLocked()
			}
			c.mu.Unlock()
			if stale {
				return snapshot, "stale"
			}
			return snapshot, "ready"
		}
		if c.pending == nil && !now.Before(c.nextAttempt) {
			c.startLocked()
		}
		pending := c.pending
		c.mu.Unlock()
		if pending == nil {
			return nil, "unavailable"
		}
		select {
		case <-pending:
		case <-ctx.Done():
			return nil, "unavailable"
		}
	}
}

func (c *pelicanReferenceCache) startLocked() {
	pending := make(chan struct{})
	c.pending = pending
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), pelicanReferenceTimeout)
		defer cancel()
		snapshot, err := c.fetch(ctx)
		c.mu.Lock()
		defer c.mu.Unlock()
		now := c.now()
		if err == nil {
			snapshot.syncedAt = now
			c.snapshot = snapshot
			c.failed = false
			c.nextAttempt = time.Time{}
		} else {
			c.failed = true
			c.nextAttempt = now.Add(pelicanReferenceBackoff)
		}
		c.pending = nil
		close(pending)
	}()
}

func (c *pelicanReferenceCache) fetch(ctx context.Context) (*pelicanReferenceSnapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pelicanReferenceDataURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("reference source status rejected")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, pelicanReferenceBodyLimit+1))
	if err != nil {
		return nil, err
	}
	if len(body) > pelicanReferenceBodyLimit {
		return nil, errors.New("reference source body too large")
	}
	return decodePelicanReferenceSnapshot(body)
}

func (s *PelicanTestService) ReferenceScores(ctx context.Context, auth PelicanAuthorization) (*PelicanReferenceScores, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !auth.Admin && !settings.Enabled {
		return nil, ErrPelicanDisabled
	}
	result := &PelicanReferenceScores{SourceURL: PelicanReferenceSourceURL, Status: "disabled", Items: []PelicanReferenceItem{}}
	if settings.Radar == nil || !settings.Radar.Enabled {
		return result, nil
	}
	result.Enabled = true
	snapshot, status := s.referenceCache().get(ctx)
	result.Status = status
	if snapshot != nil {
		result.TestStartedAt = referenceCopy(snapshot.started)
		result.TestCompletedAt = referenceCopy(snapshot.completed)
		result.SourceSnapshotAt = referenceCopy(snapshot.sourceAt)
		result.SyncedAt = referenceCopy(snapshot.syncedAt)
		result.RoundID = referenceCopy(snapshot.round)
		result.ScoreMax = referenceCopy(snapshot.scoreMax)
	}
	for _, selected := range settings.Radar.Selections {
		item := PelicanReferenceItem{Model: selected.Model, DisplayName: selected.Model, Effort: selected.Effort, Status: "unavailable"}
		if snapshot != nil {
			if canonical, ok := snapshot.aliases[item.Model]; ok {
				item.Model = canonical
				item.DisplayName = canonical
			}
			item.Status = "not_recorded"
			if name, known := snapshot.names[item.Model]; known {
				item.DisplayName = name
				key := pelicanReferenceKey{item.Model, item.Effort}
				if !snapshot.efforts[key] {
					item.Status = "unsupported"
				} else if score := snapshot.scores[key]; score != nil {
					item.Status = "available"
					item.Score = referenceCopy(*score)
				}
			}
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (s *PelicanTestService) ReferenceCatalog(ctx context.Context) (*PelicanReferenceCatalog, error) {
	snapshot, status := s.referenceCache().get(ctx)
	result := &PelicanReferenceCatalog{Status: status, Models: []PelicanReferenceModel{}}
	if snapshot != nil {
		result.SyncedAt = referenceCopy(snapshot.syncedAt)
		for _, model := range snapshot.models {
			model.Efforts = append([]string{}, model.Efforts...)
			result.Models = append(result.Models, model)
		}
	}
	return result, nil
}

func referenceCopy[T any](value T) *T { return &value }

// Reject duplicate object keys before decoding: json.Unmarshal otherwise
// silently accepts a later conflicting metadata, alias, task or score value.
func referenceJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		keys := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || keys[name] {
				return errors.New("duplicate reference object key")
			}
			keys[name] = true
			if err := referenceJSONValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := referenceJSONValue(decoder); err != nil {
				return err
			}
		}
	default:
		return errors.New("unexpected reference delimiter")
	}
	_, err = decoder.Token()
	return err
}

func decodePelicanReferenceSnapshot(body []byte) (*pelicanReferenceSnapshot, error) {
	if !json.Valid(body) {
		return nil, errors.New("invalid reference JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := referenceJSONValue(decoder); err != nil {
		return nil, fmt.Errorf("reference JSON: %w", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, errors.New("trailing reference JSON")
	}
	var source struct {
		Schema   string `json:"schema"`
		Protocol string `json:"protocol"`
		Round    string `json:"round_id"`
		Snapshot string `json:"score_snapshot_at"`
		Interval struct {
			Started   string `json:"started_at"`
			Completed string `json:"completed_at"`
		} `json:"author_test_interval"`
		ScoreMax     *float64            `json:"score_max"`
		Weights      map[string]float64  `json:"task_weights"`
		Models       []string            `json:"models"`
		Names        map[string]string   `json:"model_names"`
		ModelEfforts map[string][]string `json:"model_efforts"`
		Efforts      []string            `json:"efforts"`
		Rows         []struct {
			Model    string          `json:"model"`
			Effort   string          `json:"effort"`
			Combined json.RawMessage `json:"combined"`
			Pompeii  json.RawMessage `json:"pompeii"`
			Pelican  json.RawMessage `json:"pelican"`
			Candy    json.RawMessage `json:"candy"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(body, &source); err != nil {
		return nil, err
	}
	if source.Schema != "codex-four-task-results/1" || source.Protocol != "codex-four-weighted-150-v2" || !pelicanReferenceIdentity(source.Round) {
		return nil, errors.New("unknown reference schema or metadata")
	}
	if source.ScoreMax == nil || !referencePositive(*source.ScoreMax) || len(source.Weights) != 4 {
		return nil, errors.New("invalid reference score maximum")
	}
	totalWeight := 0.0
	for _, key := range []string{"combined", "pompeii", "pelican", "candy"} {
		weight := source.Weights[key]
		if !referencePositive(weight) {
			return nil, errors.New("invalid reference weight")
		}
		totalWeight += weight
	}
	if math.Abs(totalWeight-*source.ScoreMax) > 1e-6 {
		return nil, errors.New("reference weights do not sum to maximum")
	}
	snapshot := &pelicanReferenceSnapshot{
		names: map[string]string{}, aliases: map[string]string{}, efforts: map[pelicanReferenceKey]bool{},
		scores: map[pelicanReferenceKey]*float64{}, round: source.Round, scoreMax: *source.ScoreMax,
	}
	var err error
	if snapshot.started, err = parsePelicanReferenceTime(source.Interval.Started); err != nil {
		return nil, errors.New("invalid reference start time")
	}
	if snapshot.completed, err = parsePelicanReferenceTime(source.Interval.Completed); err != nil || snapshot.completed.Before(snapshot.started) {
		return nil, errors.New("invalid reference completion time")
	}
	if snapshot.sourceAt, err = parsePelicanReferenceTime(source.Snapshot); err != nil {
		return nil, errors.New("invalid reference snapshot time")
	}
	if len(source.Models) == 0 || len(source.Models) > 128 || len(source.Names) != len(source.Models) || len(source.ModelEfforts) != len(source.Models) ||
		len(source.Efforts) == 0 || len(source.Efforts) > 6 || source.Rows == nil || len(source.Rows) > 768 {
		return nil, errors.New("invalid reference catalog bounds")
	}
	efforts := map[string]bool{}
	for _, effort := range source.Efforts {
		if !pelicanReferenceEffort(effort) || efforts[effort] {
			return nil, errors.New("invalid reference efforts")
		}
		efforts[effort] = true
	}
	for _, model := range source.Models {
		name := source.Names[model]
		declared := source.ModelEfforts[model]
		if !pelicanReferenceIdentity(model) || !pelicanReferenceIdentity(name) || snapshot.names[model] != "" || len(declared) == 0 || len(declared) > 6 {
			return nil, errors.New("invalid reference model")
		}
		if previous, exists := snapshot.aliases[name]; exists && previous != model {
			return nil, errors.New("ambiguous reference model label")
		}
		snapshot.names[model] = name
		snapshot.aliases[name] = model
		for _, effort := range declared {
			key := pelicanReferenceKey{model, effort}
			if !efforts[effort] || snapshot.efforts[key] {
				return nil, errors.New("invalid model effort declaration")
			}
			snapshot.efforts[key] = true
		}
		snapshot.models = append(snapshot.models, PelicanReferenceModel{Model: model, DisplayName: name, Efforts: append([]string{}, declared...)})
	}
	for alias, model := range snapshot.aliases {
		if _, canonical := snapshot.names[alias]; canonical && alias != model {
			return nil, errors.New("reference alias conflicts with canonical model")
		}
	}
	seen := map[pelicanReferenceKey]bool{}
	for _, row := range source.Rows {
		key := pelicanReferenceKey{row.Model, row.Effort}
		if !snapshot.efforts[key] || seen[key] {
			return nil, errors.New("duplicate or undeclared reference row")
		}
		seen[key] = true
		combined, err1 := referenceTaskNumber(row.Combined, "score")
		pompeii, err2 := referenceTaskNumber(row.Pompeii, "score100")
		pelican, err3 := referenceTaskNumber(row.Pelican, "score")
		candy, err4 := referenceTaskBool(row.Candy, "correct")
		if err := errors.Join(err1, err2, err3, err4); err != nil {
			return nil, err
		}
		if combined == nil || pompeii == nil || pelican == nil || candy == nil {
			snapshot.scores[key] = nil
			continue
		}
		score := *combined/100*source.Weights["combined"] + *pompeii/100*source.Weights["pompeii"] + *pelican/100*source.Weights["pelican"]
		if *candy {
			score += source.Weights["candy"]
		}
		snapshot.scores[key] = &score
	}
	return snapshot, nil
}

func referencePositive(value float64) bool {
	return value > 0 && !math.IsInf(value, 0) && !math.IsNaN(value)
}

// time.Parse also accepts a one-digit hour, decimal comma and some invalid
// timezone offsets, so constrain those forms before parsing RFC3339.
var pelicanReferenceTimePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](?:\.[0-9]+)?(?:Z|[+-](?:[01][0-9]|2[0-3]):[0-5][0-9])$`)

func parsePelicanReferenceTime(value string) (time.Time, error) {
	if len(value) > 100 || !pelicanReferenceTimePattern.MatchString(value) {
		return time.Time{}, errors.New("invalid RFC3339 reference timestamp")
	}
	return time.Parse(time.RFC3339Nano, value)
}

func referenceTaskValue(raw json.RawMessage, field string) (json.RawMessage, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var task map[string]json.RawMessage
	if err := json.Unmarshal(raw, &task); err != nil || task == nil {
		return nil, errors.New("invalid reference task object")
	}
	value := task[field]
	if len(value) == 0 || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil, nil
	}
	return value, nil
}

func referenceTaskNumber(raw json.RawMessage, field string) (*float64, error) {
	value, err := referenceTaskValue(raw, field)
	if err != nil || value == nil {
		return nil, err
	}
	var score float64
	if err := json.Unmarshal(value, &score); err != nil || math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 100 {
		return nil, errors.New("invalid reference task score")
	}
	return &score, nil
}

func referenceTaskBool(raw json.RawMessage, field string) (*bool, error) {
	value, err := referenceTaskValue(raw, field)
	if err != nil || value == nil {
		return nil, err
	}
	var correct bool
	if err := json.Unmarshal(value, &correct); err != nil {
		return nil, errors.New("invalid reference task correctness")
	}
	return &correct, nil
}

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type pelicanReferenceAPIStub struct {
	pelicanAPI
	auth service.PelicanAuthorization
	err  error
}

func (s *pelicanReferenceAPIStub) ReferenceScores(_ context.Context, auth service.PelicanAuthorization) (*service.PelicanReferenceScores, error) {
	s.auth = auth
	return &service.PelicanReferenceScores{
		Enabled: true, Status: "unavailable", SourceURL: service.PelicanReferenceSourceURL,
		Items: []service.PelicanReferenceItem{{Model: "custom-model", DisplayName: "custom-model", Effort: "low", Status: "unavailable"}},
	}, s.err
}

func (s *pelicanReferenceAPIStub) ReferenceCatalog(context.Context) (*service.PelicanReferenceCatalog, error) {
	return &service.PelicanReferenceCatalog{Status: "unavailable", Models: []service.PelicanReferenceModel{}}, s.err
}

func TestPelicanReferenceHandlerAuthAndRedaction(t *testing.T) {
	for _, tc := range []struct {
		name, role string
		logged     bool
		err        error
		status     int
	}{
		{"anonymous", "", false, nil, 401},
		{"ordinary", "user", true, nil, 200},
		{"admin", service.RoleAdmin, true, nil, 200},
		{"disabled", "user", true, service.ErrPelicanDisabled, 403},
		{"internal", "user", true, errors.New("secret-provider-token prompt database"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &pelicanReferenceAPIStub{err: tc.err}
			h := &PelicanTestHandler{svc: stub, groups: pelicanAccessStub{groups: []service.Group{{ID: 12}}}}
			c, w := pelicanContext(tc.role, tc.logged)
			h.ReferenceScores(c)
			if w.Code != tc.status || w.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatalf("response=%d %s", w.Code, w.Body.String())
			}
			for _, private := range []string{"secret-provider", "prompt", "group_ids", "account_id", "settings"} {
				if strings.Contains(w.Body.String(), private) {
					t.Fatalf("private data %s leaked: %s", private, w.Body.String())
				}
			}
			if tc.status == 200 {
				if stub.auth.Admin != (tc.role == service.RoleAdmin) {
					t.Fatalf("auth=%+v", stub.auth)
				}
				var result struct {
					Data map[string]any `json:"data"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				for _, field := range []string{"test_started_at", "test_completed_at", "source_snapshot_at", "synced_at", "round_id", "score_max"} {
					if value, exists := result.Data[field]; !exists || value != nil {
						t.Fatalf("missing/null contract %s=%v", field, value)
					}
				}
			}
		})
	}
	h := &PelicanTestHandler{svc: &pelicanReferenceAPIStub{}}
	for _, role := range []string{"user", ""} {
		c, w := pelicanContext(role, true)
		h.ReferenceCatalog(c)
		if w.Code != 403 {
			t.Fatalf("catalog non-admin status=%d", w.Code)
		}
	}
	c, w := pelicanContext(service.RoleAdmin, true)
	h.ReferenceCatalog(c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"models":[]`) {
		t.Fatalf("catalog=%d %s", w.Code, w.Body.String())
	}
}

type pelicanReferenceSettingsRepo struct {
	service.SettingRepository
	mu     sync.Mutex
	value  string
	writes int
}

func (r *pelicanReferenceSettingsRepo) GetValue(context.Context, string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.value, nil
}

func (r *pelicanReferenceSettingsRepo) Set(_ context.Context, _ string, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value = value
	r.writes++
	return nil
}

type pelicanReferenceHTTPTransport func(*http.Request) (*http.Response, error)

func (f pelicanReferenceHTTPTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestPelicanReferenceRuntimeHTTPWithFixtureAndSettings(t *testing.T) {
	body, err := os.ReadFile("../service/testdata/pelican-radar-valid.json")
	if err != nil {
		t.Fatal(err)
	}
	var sourceRequests atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceRequests.Add(1)
		if r.URL.Path != "/data/codex-four-task-results.json" {
			t.Errorf("unexpected source path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer source.Close()
	target, err := url.Parse(source.URL)
	if err != nil {
		t.Fatal(err)
	}
	// Production still requests the fixed HTTPS URL. Only the test transport
	// sends that request to the local fixture server, without adding a URL API.
	originalTransport := http.DefaultTransport
	http.DefaultTransport = pelicanReferenceHTTPTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://codexradar.com/data/codex-four-task-results.json" {
			t.Errorf("unexpected source URL: %s", r.URL)
		}
		local := r.Clone(r.Context())
		copyURL := *r.URL
		copyURL.Scheme, copyURL.Host = target.Scheme, target.Host
		local.URL = &copyURL
		return originalTransport.RoundTrip(local)
	})
	defer func() { http.DefaultTransport = originalTransport }()
	repo := &pelicanReferenceSettingsRepo{value: `{"display_name":"fixture","prompt":"secret-prompt","enabled":true}`}
	svc := service.NewPelicanTestService(nil, nil, nil, nil)
	svc.SetSettingsRepository(repo)
	h := &PelicanTestHandler{svc: svc, groups: pelicanAccessStub{}}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(func(c *gin.Context) {
		role := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if role != "user" && role != service.RoleAdmin {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
		c.Set(string(middleware.ContextKeyUserRole), role)
		c.Next()
	})
	router.GET("/api/v1/pelican-tests/reference-scores", h.ReferenceScores)
	router.GET("/api/v1/admin/pelican-reference-catalog", h.ReferenceCatalog)
	router.GET("/api/v1/admin/pelican-test-settings", h.GetSettings)
	router.PUT("/api/v1/admin/pelican-test-settings", h.UpdateSettings)
	app := httptest.NewServer(router)
	defer app.Close()
	// Capture the original transport so local test calls do not use the source
	// interceptor. The production source fetch uses http.DefaultTransport.
	client := &http.Client{Transport: originalTransport}
	call := func(method, path, role, body string) (int, []byte) {
		t.Helper()
		req, err := http.NewRequest(method, app.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if role != "" {
			req.Header.Set("Authorization", "Bearer "+role)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode, data
	}
	readPath := "/api/v1/pelican-tests/reference-scores"
	settingsPath := "/api/v1/admin/pelican-test-settings"
	catalogPath := "/api/v1/admin/pelican-reference-catalog"
	if status, _ := call("GET", readPath, "", ""); status != 401 {
		t.Fatal("anonymous read allowed")
	}
	if status, data := call("GET", readPath, "user", ""); status != 200 || !strings.Contains(string(data), `"status":"disabled"`) || sourceRequests.Load() != 0 {
		t.Fatalf("disabled read=%d %s requests=%d", status, data, sourceRequests.Load())
	}
	if status, _ := call("GET", catalogPath, "user", ""); status != 403 || sourceRequests.Load() != 0 {
		t.Fatal("ordinary catalog allowed/fetched")
	}
	if status, data := call("GET", catalogPath, service.RoleAdmin, ""); status != 200 || sourceRequests.Load() != 1 || !strings.Contains(string(data), `"display_name":"GPT-6.1 Sol"`) {
		t.Fatalf("admin catalog=%d %s", status, data)
	}
	settings := `{"display_name":"fixture","prompt":"secret-prompt","radar":{"enabled":true,"selections":[{"model":"GPT-6.1 Sol","effort":"low"},{"model":"gpt-6-luna","effort":"ultra"},{"model":"custom-model","effort":"max"}]}}`
	if status, _ := call("PUT", settingsPath, "user", settings); status != 403 || repo.writes != 0 {
		t.Fatal("ordinary settings write allowed")
	}
	if status, data := call("PUT", settingsPath, service.RoleAdmin, settings); status != 200 {
		t.Fatalf("admin settings=%d %s", status, data)
	}
	status, data := call("GET", readPath, "user", "")
	var envelope struct {
		Data service.PelicanReferenceScores `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	got := envelope.Data
	if status != 200 || got.Status != "ready" || len(got.Items) != 3 || got.Items[0].Model != "gpt-6.1-sol" ||
		got.Items[0].Score == nil || math.Abs(*got.Items[0].Score-106.16576086956522) > 1e-10 ||
		got.Items[1].Status != "unsupported" || got.Items[2].Status != "not_recorded" || sourceRequests.Load() != 1 {
		t.Fatalf("runtime DTO=%d %s requests=%d", status, data, sourceRequests.Load())
	}
	for _, secret := range []string{"secret-prompt", "raw_score_total", "task_weights", "group_id", "account_id"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("public response leaked %s", secret)
		}
	}
	t.Logf("fixture source requests=%d authenticated DTO=%s", sourceRequests.Load(), data)
	// Older PUT keeps the saved radar object and selection order.
	if status, data := call("PUT", settingsPath, service.RoleAdmin, `{"display_name":"legacy","prompt":"secret-prompt"}`); status != 200 || !strings.Contains(string(data), `"model":"gpt-6.1-sol"`) {
		t.Fatalf("legacy update=%d %s", status, data)
	}
	if status, _ := call("PUT", readPath, "user", `{}`); status != http.StatusMethodNotAllowed {
		t.Fatalf("user read endpoint writable: %d", status)
	}
	if status, data := call("PUT", settingsPath, service.RoleAdmin, `{"display_name":"fixture","prompt":"secret-prompt","enabled":false}`); status != 200 {
		t.Fatalf("disable plaza=%d %s", status, data)
	}
	if status, _ := call("GET", readPath, "user", ""); status != 403 {
		t.Fatalf("plaza-disabled read=%d", status)
	}
	if status, _ := call("GET", readPath, service.RoleAdmin, ""); status != 200 {
		t.Fatalf("plaza-disabled admin read=%d", status)
	}
	if status, data := call("PUT", settingsPath, service.RoleAdmin, `{"display_name":"fixture","prompt":"secret-prompt","radar":{"enabled":false,"selections":[]}}`); status != 200 || !strings.Contains(string(data), `"selections":[]`) {
		t.Fatalf("explicit clear=%d %s", status, data)
	}
}

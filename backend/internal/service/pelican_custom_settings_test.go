package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type pelicanDefaultSettingsRepo struct{ SettingRepository }

func (pelicanDefaultSettingsRepo) GetValue(context.Context, string) (string, error) {
	return "", ErrSettingNotFound
}

type pelicanSettingsRepoStub struct {
	SettingRepository
	value    string
	err      error
	writeErr error
	set      string
	sets     int
}

type pelicanAccessRepo struct {
	PelicanTestRepository
	galleryCalls int
	historyCalls int
	resultCalls  int
}

func (r *pelicanAccessRepo) ListGallery(_ context.Context, _ PelicanAuthorization, _ int64, _ int64, page, pageSize int) (*PelicanListResponse, error) {
	r.galleryCalls++
	return &PelicanListResponse{Page: page, PageSize: pageSize}, nil
}

func (r *pelicanAccessRepo) ListHistory(_ context.Context, _ PelicanAuthorization, _ int64, _ int64) ([]PelicanResult, error) {
	r.historyCalls++
	return []PelicanResult{}, nil
}

func (r *pelicanAccessRepo) GetResult(_ context.Context, _ PelicanAuthorization, id int64) (*PelicanResult, error) {
	r.resultCalls++
	return &PelicanResult{ID: id}, nil
}

func pelicanSettingsUpdate(settings PelicanTestSettings) PelicanTestSettingsUpdate {
	return PelicanTestSettingsUpdate{DisplayName: settings.DisplayName, Prompt: settings.Prompt, Enabled: &settings.Enabled}
}

func (r *pelicanSettingsRepoStub) GetValue(context.Context, string) (string, error) {
	return r.value, r.err
}
func (r *pelicanSettingsRepoStub) Set(_ context.Context, _ string, value string) error {
	r.set = value
	r.sets++
	if r.writeErr != nil {
		return r.writeErr
	}
	return r.err
}

func TestPelicanSettingsDefaultAndWriteFailures(t *testing.T) {
	repo := &pelicanSettingsRepoStub{err: ErrSettingNotFound}
	s := NewPelicanTestService(nil, nil, nil, nil)
	s.SetSettingsRepository(repo)
	settings, err := s.GetSettings(context.Background())
	if err != nil || settings != defaultPelicanTestSettings() {
		t.Fatalf("settings=%+v err=%v", settings, err)
	}
	repo.err = nil
	repo.value = `{"display_name":"鹈鹕测试","prompt":"x","enabled":true}`
	repo.writeErr = errors.New("write failed")
	if _, err := s.UpdateSettings(context.Background(), pelicanSettingsUpdate(defaultPelicanTestSettings())); err == nil || repo.sets != 1 {
		t.Fatalf("write failure err=%v sets=%d", err, repo.sets)
	}
	before := repo.sets
	if _, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "x", Prompt: " \n "}); !errors.Is(err, ErrPelicanInvalid) || repo.sets != before {
		t.Fatalf("invalid update err=%v writes=%d", err, repo.sets)
	}
	repo.writeErr = nil
	repo.value = "{"
	if _, err := s.GetSettings(context.Background()); err == nil {
		t.Fatal("malformed JSON accepted")
	}
}

func TestPelicanSettingsRejectInvalidLimitsWithoutWrite(t *testing.T) {
	repo := &pelicanSettingsRepoStub{value: `{"display_name":"鹈鹕测试","prompt":"x","enabled":true}`}
	s := NewPelicanTestService(nil, nil, nil, nil)
	s.SetSettingsRepository(repo)
	for _, input := range []PelicanTestSettings{
		{DisplayName: "", Prompt: "x"},
		{DisplayName: strings.Repeat("名", 41), Prompt: "x"},
		{DisplayName: "x", Prompt: " \n\t "},
		{DisplayName: "x", Prompt: strings.Repeat("提", 20001)},
	} {
		before := repo.sets
		if _, err := s.UpdateSettings(context.Background(), pelicanSettingsUpdate(input)); !errors.Is(err, ErrPelicanInvalid) || repo.sets != before {
			t.Fatalf("input=%+v err=%v writes=%d", input, err, repo.sets)
		}
	}
}

func TestPelicanSettingsPreservePromptAndValidatePersistedValue(t *testing.T) {
	repo := &pelicanSettingsRepoStub{value: `{"display_name":"  自定义页面  ","prompt":"  第一行\n第二行  "}`}
	s := NewPelicanTestService(nil, nil, nil, nil)
	s.SetSettingsRepository(repo)
	settings, err := s.GetSettings(context.Background())
	if err != nil || settings.Prompt != "  第一行\n第二行  " || settings.DisplayName != "  自定义页面  " {
		t.Fatalf("settings=%+v err=%v", settings, err)
	}
	snapshot, err := s.promptSnapshot(context.Background())
	if err != nil || snapshot.prompt != settings.Prompt || snapshot.version == PelicanPromptVersion {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	if _, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "  新名称  ", Prompt: settings.Prompt}); err != nil || repo.set == "" {
		t.Fatalf("update err=%v value=%q", err, repo.set)
	}
	if _, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "", Prompt: settings.Prompt}); !errors.Is(err, ErrPelicanInvalid) {
		t.Fatalf("blank name err=%v", err)
	}
	repo.value = `{"display_name":"","prompt":"x"}`
	if _, err := s.GetSettings(context.Background()); err == nil || errors.Is(err, ErrPelicanInvalid) {
		t.Fatalf("invalid persisted settings must be internal error, got %v", err)
	}
}

func TestPelicanSettingsReadFailurePreventsClaim(t *testing.T) {
	repo := &pelicanRunnerRepo{claim: make(chan struct{}, 1), release: make(chan struct{}, 1)}
	s := NewPelicanTestService(repo, nil, nil, nil)
	s.SetSettingsRepository(&pelicanSettingsRepoStub{err: errors.New("settings unavailable")})
	s.Start(context.Background())
	defer s.Stop(context.Background())
	if err := s.RunNow(context.Background(), 1); err == nil {
		t.Fatal("expected settings read failure")
	}
	select {
	case <-repo.claim:
		t.Fatal("claim ran after settings read failure")
	default:
	}
}

func TestPelicanSettingsEnabledLegacyAndOptionalUpdate(t *testing.T) {
	repo := &pelicanSettingsRepoStub{value: `{"display_name":"旧页面","prompt":"旧提示词","enabled":false}`}
	s := NewPelicanTestService(nil, nil, nil, nil)
	s.SetSettingsRepository(repo)
	settings, err := s.GetSettings(context.Background())
	if err != nil || settings.Enabled {
		t.Fatalf("explicit false lost: %+v err=%v", settings, err)
	}
	updated, err := s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "旧页面", Prompt: "旧提示词"})
	if err != nil || updated.Enabled {
		t.Fatalf("omitted enabled must preserve false: %+v err=%v", updated, err)
	}
	repo.value = `{"display_name":"旧页面","prompt":"旧提示词"}`
	settings, err = s.GetSettings(context.Background())
	if err != nil || !settings.Enabled {
		t.Fatalf("legacy settings must default enabled: %+v err=%v", settings, err)
	}
	falseValue := false
	updated, err = s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "旧页面", Prompt: "旧提示词", Enabled: &falseValue})
	if err != nil || updated.Enabled {
		t.Fatalf("explicit false update failed: %+v err=%v", updated, err)
	}
	trueValue := true
	updated, err = s.UpdateSettings(context.Background(), PelicanTestSettingsUpdate{DisplayName: "旧页面", Prompt: "旧提示词", Enabled: &trueValue})
	if err != nil || !updated.Enabled {
		t.Fatalf("explicit true update failed: %+v err=%v", updated, err)
	}
}

func TestPelicanDisabledUserAccessFailsClosedBeforeRepository(t *testing.T) {
	repo := &pelicanAccessRepo{}
	s := NewPelicanTestService(repo, nil, nil, nil)
	s.SetSettingsRepository(&pelicanSettingsRepoStub{value: `{"display_name":"鹈鹕测试","prompt":"x","enabled":false}`})
	user := PelicanAuthorization{}
	if _, err := s.ListGallery(context.Background(), user, 0, 0, 1, 24); !errors.Is(err, ErrPelicanDisabled) {
		t.Fatalf("gallery err=%v", err)
	}
	if _, err := s.ListHistory(context.Background(), user, 1, 1); !errors.Is(err, ErrPelicanDisabled) {
		t.Fatalf("history err=%v", err)
	}
	if _, err := s.GetResult(context.Background(), user, 9); !errors.Is(err, ErrPelicanDisabled) {
		t.Fatalf("result err=%v", err)
	}
	if repo.galleryCalls != 0 || repo.historyCalls != 0 || repo.resultCalls != 0 {
		t.Fatalf("disabled user reached repository: %+v", repo)
	}
	admin := PelicanAuthorization{Admin: true}
	if _, err := s.ListGallery(context.Background(), admin, 0, 0, 1, 24); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListHistory(context.Background(), admin, 1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetResult(context.Background(), admin, 9); err != nil {
		t.Fatal(err)
	}
	if repo.galleryCalls != 1 || repo.historyCalls != 1 || repo.resultCalls != 1 {
		t.Fatalf("admin was not allowed: %+v", repo)
	}
}

func TestPelicanSettingsReadFailureFailsClosedBeforeUserRepositoryAccess(t *testing.T) {
	repo := &pelicanAccessRepo{}
	s := NewPelicanTestService(repo, nil, nil, nil)
	s.SetSettingsRepository(&pelicanSettingsRepoStub{err: errors.New("settings unavailable")})
	if _, err := s.ListGallery(context.Background(), PelicanAuthorization{}, 0, 0, 1, 24); err == nil || errors.Is(err, ErrPelicanDisabled) {
		t.Fatalf("settings error must remain fail-closed internal error: %v", err)
	}
	if repo.galleryCalls != 0 {
		t.Fatalf("repository called after settings error: %d", repo.galleryCalls)
	}
}

func TestPelicanDisabledPreventsClaim(t *testing.T) {
	repo := &pelicanRunnerRepo{claim: make(chan struct{}, 1), release: make(chan struct{}, 1)}
	s := NewPelicanTestService(repo, nil, nil, nil)
	s.SetSettingsRepository(&pelicanSettingsRepoStub{value: `{"display_name":"鹈鹕测试","prompt":"x","enabled":false}`})
	s.Start(context.Background())
	defer s.Stop(context.Background())
	if err := s.RunNow(context.Background(), 1); !errors.Is(err, ErrPelicanDisabled) {
		t.Fatalf("err=%v", err)
	}
	select {
	case <-repo.claim:
		t.Fatal("disabled plan was claimed")
	default:
	}
	if err := s.claimAndRun(context.Background(), 1, false); !errors.Is(err, ErrPelicanDisabled) {
		t.Fatalf("scheduled err=%v", err)
	}
	select {
	case <-repo.claim:
		t.Fatal("disabled scheduled plan was claimed")
	default:
	}
}

func TestPelicanRunAndCancellationKeepOriginalSnapshot(t *testing.T) {
	release := make(chan struct{})
	executor := &pelicanReviewExecutor{started: make(chan int64, 2), release: release, cancelled: make(chan struct{}, 2)}
	accounts := []Account{*reviewAccount(1, true), *reviewAccount(2, true)}
	fresh := map[int64]*Account{1: reviewAccount(1, true), 2: reviewAccount(2, true)}
	s, runRepo := newPelicanReviewService(accounts, fresh, executor)
	defer s.Stop(context.Background())
	settings := &pelicanSettingsRepoStub{value: `{"display_name":"旧名","prompt":"  旧提示词\n原文  "}`}
	s.SetSettingsRepository(settings)
	snapshot, err := s.promptSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RunNow(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	for range 1 {
		select {
		case <-executor.started:
		case <-time.After(time.Second):
			t.Fatal("provider did not start")
		}
	}
	settings.value = `{"display_name":"新名","prompt":"下一轮提示词","enabled":false}`
	close(release)
	select {
	case <-runRepo.released:
	case <-time.After(time.Second):
		t.Fatal("run did not finish")
	}
	executor.mu.Lock()
	prompts := append([]string(nil), executor.prompts...)
	executor.mu.Unlock()
	if len(prompts) != 1 {
		t.Fatalf("prompts=%q", prompts)
	}
	for _, prompt := range prompts {
		if prompt != snapshot.prompt {
			t.Fatalf("prompt=%q want=%q", prompt, snapshot.prompt)
		}
	}
	runRepo.mu.Lock()
	results := append([]PelicanResult(nil), runRepo.saved...)
	runRepo.mu.Unlock()
	if len(results) != 1 {
		t.Fatalf("results=%+v", results)
	}
	for _, result := range results {
		if result.PromptVersion != snapshot.version {
			t.Fatalf("result version=%q want=%q", result.PromptVersion, snapshot.version)
		}
	}
}

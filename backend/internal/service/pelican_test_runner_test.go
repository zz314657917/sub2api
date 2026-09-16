package service

import (
	"context"
	"testing"
	"time"
)

type pelicanRunnerRepo struct {
	PelicanTestRepository
	claim   chan struct{}
	release chan struct{}
}

func (r *pelicanRunnerRepo) Claim(context.Context, int64, bool, time.Time) (*PelicanPlan, error) {
	select {
	case r.claim <- struct{}{}:
	default:
	}
	return &PelicanPlan{ID: 1, GroupID: 1, ModelID: "m", MinChars: 100, MaxResults: 1, RunGeneration: 1}, nil
}
func (r *pelicanRunnerRepo) Release(context.Context, int64, int64) error {
	r.release <- struct{}{}
	return nil
}
func (r *pelicanRunnerRepo) SaveResult(context.Context, *PelicanResult, int, int64) error { return nil }

type pelicanRunnerAccounts struct {
	AccountRepository
	block <-chan struct{}
}

func (a pelicanRunnerAccounts) ListByGroup(ctx context.Context, _ int64) ([]Account, error) {
	select {
	case <-a.block:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return []Account{}, nil
}

type pelicanRunnerTester struct{}

func (pelicanRunnerTester) RunPelicanTest(context.Context, int64, string) (*PelicanResult, error) {
	return nil, nil
}

func TestPelicanRunnerRejectsOverlappingManualRun(t *testing.T) {
	r := &pelicanRunnerRepo{claim: make(chan struct{}, 2), release: make(chan struct{}, 2)}
	block := make(chan struct{})
	s := NewPelicanTestService(r, pelicanRunnerAccounts{block: block}, nil, pelicanRunnerTester{})
	s.Start(context.Background())
	defer s.Stop(context.Background())
	if err := s.RunNow(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	select {
	case <-r.claim:
	case <-time.After(time.Second):
		t.Fatal("claim did not start")
	}
	if err := s.RunNow(context.Background(), 2); err != ErrPelicanConflict {
		t.Fatalf("overlap err=%v", err)
	}
	close(block)
	select {
	case <-r.release:
	case <-time.After(time.Second):
		t.Fatal("run did not release")
	}
}

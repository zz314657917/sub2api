package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPelicanGroupOrderSettings(t *testing.T) {
	ctx := context.Background()
	repo := &pelicanSettingsRepoStub{value: `{"display_name":"鹈鹕测试","prompt":"x","group_order":[9,3]}`}
	s := NewPelicanTestService(nil, nil, nil, nil)
	s.SetSettingsRepository(repo)
	input := PelicanTestSettingsUpdate{DisplayName: "鹈鹕测试", Prompt: "x"}
	got, err := s.UpdateSettings(ctx, input)
	if err != nil || !reflect.DeepEqual(got.GroupOrder, []int64{9, 3}) {
		t.Fatalf("legacy update discarded order: %+v %v", got, err)
	}
	repo.value = repo.set
	empty := []int64{}
	input.GroupOrder = &empty
	got, err = s.UpdateSettings(ctx, input)
	if err != nil || len(got.GroupOrder) != 0 {
		t.Fatalf("explicit clear failed: %+v %v", got, err)
	}
	repo.value = repo.set
	got, err = s.GetSettings(ctx)
	if err != nil || len(got.GroupOrder) != 0 {
		t.Fatalf("clear did not persist: %+v %v", got, err)
	}
	for _, order := range [][]int64{{0}, {-1}, {3, 3}, make([]int64, 1001)} {
		before := repo.sets
		input.GroupOrder = &order
		if _, err := s.UpdateSettings(ctx, input); !errors.Is(err, ErrPelicanInvalid) || repo.sets != before {
			t.Fatalf("invalid order wrote settings: %v %v", order, err)
		}
	}
	repo.value = `{"display_name":"鹈鹕测试","prompt":"x","group_order":[3,3]}`
	if _, err := s.GetSettings(ctx); err == nil {
		t.Fatal("invalid persisted order accepted")
	}
}

type pelicanOrderedRepo struct {
	pelicanAccessRepo
	order []int64
	auth  PelicanAuthorization
}

func (r *pelicanOrderedRepo) ListGalleryWithGroupOrder(_ context.Context, auth PelicanAuthorization, _, _ int64, page, size int, order []int64) (*PelicanListResponse, error) {
	r.order = append([]int64(nil), order...)
	r.auth = auth
	return &PelicanListResponse{Page: page, PageSize: size}, nil
}

func TestPelicanGroupOrderReachesRepositoryWithoutChangingAuthorization(t *testing.T) {
	ctx := context.Background()
	repo := &pelicanOrderedRepo{}
	settings := &pelicanSettingsRepoStub{value: `{"display_name":"鹈鹕测试","prompt":"x","enabled":true,"group_order":[9,3]}`}
	s := NewPelicanTestService(repo, nil, nil, nil)
	s.SetSettingsRepository(settings)
	auth := PelicanAuthorization{GroupIDs: []int64{3}}
	got, err := s.ListGallery(ctx, auth, 0, 0, 2, 6)
	if err != nil || got.Page != 2 || got.PageSize != 6 || !reflect.DeepEqual(repo.order, []int64{9, 3}) || !reflect.DeepEqual(repo.auth, auth) || repo.galleryCalls != 0 {
		t.Fatalf("ordering dispatch: %+v %+v %v", got, repo, err)
	}
	settings.err = errors.New("unavailable")
	if _, err := s.ListGallery(ctx, PelicanAuthorization{Admin: true}, 0, 0, 1, 6); err == nil {
		t.Fatal("admin gallery ignored unavailable settings")
	}
}

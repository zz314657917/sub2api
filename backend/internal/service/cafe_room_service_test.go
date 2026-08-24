package service

import (
	"context"
	"errors"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type cafeAccountStub struct {
	status   string
	platform string
	groupIDs []int64
}

type cafeRoomRepositoryStub struct {
	plan           *CafeRoomPlan
	accounts       map[int64]cafeAccountStub
	rooms          map[int64]*CafeRoom
	assigned       map[int64]bool
	live           map[int64]bool
	openRoundError map[int64]error
	deleted        []int64
	nextRoomID     int64
	optionParams   CafeRoomAccountOptionParams
}

func newCafeRoomRepositoryStub() *cafeRoomRepositoryStub {
	return &cafeRoomRepositoryStub{
		plan: &CafeRoomPlan{
			ID:                10,
			Title:             "Pixel Cafe",
			Status:            GroupBuyPlanStatusActive,
			TargetGroupID:     20,
			FulfillmentMode:   CafeRoomFulfillmentMode,
			AutoCreateRoomKey: true,
			TotalShares:       4,
			SeatCount:         4,
			TimeoutMinutes:    60,
			ValidityDays:      30,
			GroupPlatform:     "openai",
			GroupAccessMode:   CafeRoomGroupAccessMode,
			TargetGroupStatus: StatusActive,
		},
		accounts: map[int64]cafeAccountStub{
			1: {status: StatusActive, platform: "openai", groupIDs: []int64{20}},
		},
		rooms:          map[int64]*CafeRoom{},
		assigned:       map[int64]bool{},
		live:           map[int64]bool{},
		openRoundError: map[int64]error{},
	}
}

func (r *cafeRoomRepositoryStub) List(context.Context, pagination.PaginationParams, string, string, string) ([]CafeRoom, *pagination.PaginationResult, error) {
	rooms := make([]CafeRoom, 0, len(r.rooms))
	for _, room := range r.rooms {
		rooms = append(rooms, *room)
	}
	return rooms, &pagination.PaginationResult{Page: 1, PageSize: 20, Total: int64(len(rooms)), Pages: 1}, nil
}

func (r *cafeRoomRepositoryStub) GetByID(_ context.Context, id int64) (*CafeRoom, error) {
	room := r.rooms[id]
	if room == nil {
		return nil, ErrCafeRoomNotFound
	}
	copy := *room
	return &copy, nil
}

func (r *cafeRoomRepositoryStub) GetPlan(context.Context, int64) (*CafeRoomPlan, error) {
	if r.plan == nil {
		return nil, ErrCafePlanNotFound
	}
	copy := *r.plan
	return &copy, nil
}

func (r *cafeRoomRepositoryStub) GetAccount(_ context.Context, id int64) (string, string, []int64, error) {
	account, ok := r.accounts[id]
	if !ok {
		return "", "", nil, nil
	}
	return account.status, account.platform, account.groupIDs, nil
}

func (r *cafeRoomRepositoryStub) ListAccountOptions(_ context.Context, params CafeRoomAccountOptionParams) ([]CafeRoomAccountOption, *pagination.PaginationResult, error) {
	r.optionParams = params
	return []CafeRoomAccountOption{{ID: 1, Name: "Cafe account", Platform: "openai", Status: StatusActive}}, &pagination.PaginationResult{Page: 1, PageSize: 20, Total: 1, Pages: 1}, nil
}

func (r *cafeRoomRepositoryStub) HasOperationalAccount(_ context.Context, accountID, excludeRoomID int64) (bool, error) {
	if !r.assigned[accountID] {
		return false, nil
	}
	for id, room := range r.rooms {
		if id != excludeRoomID && room.AccountID != nil && *room.AccountID == accountID && (room.Status == CafeRoomStatusEnabled || room.Status == CafeRoomStatusMaintenance) {
			return true, nil
		}
	}
	return false, nil
}

func (r *cafeRoomRepositoryStub) HasLiveRound(_ context.Context, roomID int64) (bool, error) {
	return r.live[roomID], nil
}

func (r *cafeRoomRepositoryStub) Create(_ context.Context, room *CafeRoom) (*CafeRoom, error) {
	r.nextRoomID++
	copy := *room
	copy.ID = r.nextRoomID
	r.rooms[copy.ID] = &copy
	if copy.AccountID != nil && (copy.Status == CafeRoomStatusEnabled || copy.Status == CafeRoomStatusMaintenance) {
		r.assigned[*copy.AccountID] = true
	}
	return &copy, nil
}

func (r *cafeRoomRepositoryStub) Update(_ context.Context, room *CafeRoom) (*CafeRoom, error) {
	if current := r.rooms[room.ID]; current != nil && r.live[room.ID] && (current.PlanID != room.PlanID || current.AccountID == nil || room.AccountID == nil || *current.AccountID != *room.AccountID || room.Status != CafeRoomStatusEnabled) {
		return nil, ErrCafeRoomLive
	}
	copy := *room
	r.rooms[copy.ID] = &copy
	return &copy, nil
}

func (r *cafeRoomRepositoryStub) Delete(_ context.Context, id int64) error {
	if room := r.rooms[id]; room != nil && room.AccountID != nil {
		r.assigned[*room.AccountID] = false
	}
	delete(r.rooms, id)
	r.deleted = append(r.deleted, id)
	return nil
}

func (r *cafeRoomRepositoryStub) CreateOpenRound(_ context.Context, roomID int64, now time.Time) (*CafeRound, error) {
	if err := r.openRoundError[roomID]; err != nil {
		return nil, err
	}
	room := r.rooms[roomID]
	return &CafeRound{
		ID:                roomID + 100,
		PlanID:            room.PlanID,
		CafeRoomID:        &roomID,
		AssignedAccountID: room.AccountID,
		Status:            CafeRoundStatusOpen,
		TotalShares:       r.plan.TotalShares,
		TotalSeats:        r.plan.SeatCount,
		DeadlineAt:        now.Add(time.Hour),
	}, nil
}

func TestCafeRoomServiceCreateValidatesStatusAndCompatibility(t *testing.T) {
	accountID := int64(1)
	input := CafeRoomInput{Code: "ROOM-001", Name: "Room 1", PlanID: 10, AccountID: &accountID, Status: CafeRoomStatusEnabled}

	t.Run("invalid status", func(t *testing.T) {
		repo := newCafeRoomRepositoryStub()
		input.Status = "online"
		_, err := NewCafeRoomService(repo).Create(context.Background(), input)
		require.Equal(t, "CAFE_ROOM_INVALID_STATUS", infraerrors.Reason(err))
		input.Status = CafeRoomStatusEnabled
	})

	t.Run("wrong fulfillment mode", func(t *testing.T) {
		repo := newCafeRoomRepositoryStub()
		repo.plan.FulfillmentMode = "aggregate_tier"
		_, err := NewCafeRoomService(repo).Create(context.Background(), input)
		require.ErrorIs(t, err, ErrCafePlanInvalid)
	})

	t.Run("wrong group access mode", func(t *testing.T) {
		repo := newCafeRoomRepositoryStub()
		repo.plan.GroupAccessMode = "normal"
		_, err := NewCafeRoomService(repo).Create(context.Background(), input)
		require.ErrorIs(t, err, ErrCafeGroupInvalid)
	})

	t.Run("account must be active and assigned to target group", func(t *testing.T) {
		repo := newCafeRoomRepositoryStub()
		repo.accounts[1] = cafeAccountStub{status: "disabled", platform: "openai", groupIDs: []int64{20}}
		_, err := NewCafeRoomService(repo).Create(context.Background(), input)
		require.ErrorIs(t, err, ErrCafeAccountIncompatible)
	})

	t.Run("operational account is unique", func(t *testing.T) {
		repo := newCafeRoomRepositoryStub()
		repo.assigned[1] = true
		repo.rooms[99] = &CafeRoom{ID: 99, AccountID: &accountID, Status: CafeRoomStatusEnabled}
		_, err := NewCafeRoomService(repo).Create(context.Background(), input)
		require.ErrorIs(t, err, ErrCafeAccountAssigned)
	})
}

func TestCafeRoomServiceAccountOptionsValidatePlanButAllowBoundedHydration(t *testing.T) {
	repo := newCafeRoomRepositoryStub()
	svc := NewCafeRoomService(repo)

	_, _, err := svc.ListAccountOptions(context.Background(), CafeRoomAccountOptionParams{PlanID: 0})
	require.ErrorIs(t, err, ErrCafeRoomInvalid)

	repo.plan.FulfillmentMode = "aggregate_tier"
	_, _, err = svc.ListAccountOptions(context.Background(), CafeRoomAccountOptionParams{PlanID: 10})
	require.ErrorIs(t, err, ErrCafePlanInvalid)

	items, result, err := svc.ListAccountOptions(context.Background(), CafeRoomAccountOptionParams{IDs: []int64{1}})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), result.Total)
	require.Equal(t, []int64{1}, repo.optionParams.IDs)

	_, _, err = svc.ListAccountOptions(context.Background(), CafeRoomAccountOptionParams{IDs: make([]int64, 51)})
	require.ErrorIs(t, err, ErrCafeRoomInvalid)
}

func TestCafeRoomServiceUpdateAndDeleteRespectLiveState(t *testing.T) {
	repo := newCafeRoomRepositoryStub()
	accountID := int64(1)
	repo.rooms[1] = &CafeRoom{ID: 1, Code: "ROOM-001", Name: "Room 1", PlanID: 10, AccountID: &accountID, Status: CafeRoomStatusEnabled}
	svc := NewCafeRoomService(repo)

	planID := int64(11)
	repo.live[1] = true
	_, err := svc.Update(context.Background(), 1, CafeRoomUpdateInput{PlanID: &planID})
	require.ErrorIs(t, err, ErrCafeRoomLive)

	repo.live[1] = false
	err = svc.Delete(context.Background(), 1)
	require.ErrorIs(t, err, ErrCafeRoomEnabled)

	repo.rooms[1].Status = CafeRoomStatusMaintenance
	repo.live[1] = true
	err = svc.Delete(context.Background(), 1)
	require.ErrorIs(t, err, ErrCafeRoomLive)
}

func TestCafeRoomServiceBulkCreateReturnsPerAccountResults(t *testing.T) {
	repo := newCafeRoomRepositoryStub()
	repo.accounts[2] = cafeAccountStub{status: "disabled", platform: "openai", groupIDs: []int64{20}}
	repo.accounts[3] = cafeAccountStub{status: StatusActive, platform: "openai", groupIDs: []int64{20}}
	repo.openRoundError[2] = errors.New("round create failed")

	result := NewCafeRoomService(repo).BulkCreate(context.Background(), CafeRoomBulkInput{
		PlanID:          10,
		AccountIDs:      []int64{1, 2, 3},
		CodePrefix:      "CAFE-",
		StartNumber:     8,
		CreateOpenRound: true,
	})

	require.Len(t, result.Created, 1)
	require.Equal(t, int64(1), result.Created[0].AccountID)
	require.Equal(t, "CAFE-008", result.Created[0].Room.Code)
	require.NotNil(t, result.Created[0].Round)
	require.Len(t, result.Failed, 2)
	require.Equal(t, int64(2), result.Failed[0].AccountID)
	require.Equal(t, "CAFE_ACCOUNT_INCOMPATIBLE", result.Failed[0].Code)
	require.Equal(t, int64(3), result.Failed[1].AccountID)
	require.Equal(t, []int64{2}, repo.deleted)
}

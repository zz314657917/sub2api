package repository

import (
	"context"
	"time"

	"entgo.io/ent/dialect"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type cafeRoomRepository struct {
	client *dbent.Client
}

func NewCafeRoomRepository(client *dbent.Client) service.CafeRoomRepository {
	return &cafeRoomRepository{client: client}
}

func (r *cafeRoomRepository) List(ctx context.Context, params pagination.PaginationParams, status, zone, search string) ([]service.CafeRoom, *pagination.PaginationResult, error) {
	q := r.client.CafeRoom.Query().WithPlan(func(planQ *dbent.GroupBuyPlanQuery) {
		planQ.Where(groupbuyplan.DeletedAtIsNil())
		planQ.WithTargetGroup()
	})
	if status != "" {
		q = q.Where(caferoom.StatusEQ(status))
	}
	if zone != "" {
		q = q.Where(caferoom.ZoneKeyEQ(zone))
	}
	if search != "" {
		q = q.Where(caferoom.Or(
			caferoom.CodeContainsFold(search),
			caferoom.NameContainsFold(search),
		))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	page := params.Page
	if page < 1 {
		page = 1
	}
	params.Page = page
	rooms, err := q.
		Order(dbent.Asc(caferoom.FieldFeatured), dbent.Asc(caferoom.FieldSortOrder), dbent.Asc(caferoom.FieldID)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	result := make([]service.CafeRoom, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, cafeRoomToService(room))
	}
	return result, paginationResultFromTotal(int64(total), params), nil
}

func (r *cafeRoomRepository) GetByID(ctx context.Context, id int64) (*service.CafeRoom, error) {
	room, err := r.client.CafeRoom.Query().
		Where(caferoom.IDEQ(id)).
		WithPlan(func(planQ *dbent.GroupBuyPlanQuery) {
			planQ.Where(groupbuyplan.DeletedAtIsNil())
			planQ.WithTargetGroup()
		}).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafeRoomNotFound
		}
		return nil, err
	}
	converted := cafeRoomToService(room)
	return &converted, nil
}

func (r *cafeRoomRepository) GetPlan(ctx context.Context, id int64) (*service.CafeRoomPlan, error) {
	plan, err := r.client.GroupBuyPlan.Query().
		Where(groupbuyplan.IDEQ(id), groupbuyplan.DeletedAtIsNil()).
		WithTargetGroup().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafePlanNotFound
		}
		return nil, err
	}
	converted := cafePlanToService(plan)
	return &converted, nil
}

func (r *cafeRoomRepository) GetAccount(ctx context.Context, id int64) (status, platform string, groupIDs []int64, err error) {
	item, err := r.client.Account.Query().
		Where(account.IDEQ(id)).
		WithGroups().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return "", "", nil, nil
		}
		return "", "", nil, err
	}
	groupIDs = make([]int64, 0, len(item.Edges.Groups))
	for _, group := range item.Edges.Groups {
		groupIDs = append(groupIDs, group.ID)
	}
	return item.Status, item.Platform, groupIDs, nil
}

func (r *cafeRoomRepository) HasOperationalAccount(ctx context.Context, accountID, excludeRoomID int64) (bool, error) {
	q := r.client.CafeRoom.Query().Where(
		caferoom.AccountIDEQ(accountID),
		caferoom.StatusIn(service.CafeRoomStatusEnabled, service.CafeRoomStatusMaintenance),
	)
	if excludeRoomID > 0 {
		q = q.Where(caferoom.IDNEQ(excludeRoomID))
	}
	return q.Exist(ctx)
}

func (r *cafeRoomRepository) HasLiveRound(ctx context.Context, roomID int64) (bool, error) {
	return r.client.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDEQ(roomID),
		groupbuyround.StatusIn(service.CafeRoundStatusOpen, "activating", "active"),
	).Exist(ctx)
}

func (r *cafeRoomRepository) Create(ctx context.Context, room *service.CafeRoom) (*service.CafeRoom, error) {
	if room == nil || room.AccountID == nil || *room.AccountID <= 0 {
		return nil, service.ErrCafeRoomInvalid
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	txClient := tx.Client()
	defer func() { _ = tx.Rollback() }()

	if err := lockAccountAndCheckAssignment(ctx, txClient, *room.AccountID, 0); err != nil {
		return nil, err
	}
	metadata := room.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	created, err := txClient.CafeRoom.Create().
		SetCode(room.Code).
		SetName(room.Name).
		SetPlanID(room.PlanID).
		SetNillableAccountID(room.AccountID).
		SetZoneKey(room.ZoneKey).
		SetThemeKey(room.ThemeKey).
		SetSceneSlotKey(room.SceneSlotKey).
		SetStatus(room.Status).
		SetFeatured(room.Featured).
		SetSortOrder(room.SortOrder).
		SetMetadata(metadata).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	converted := cafeRoomToService(created)
	return &converted, nil
}

func (r *cafeRoomRepository) Update(ctx context.Context, room *service.CafeRoom) (*service.CafeRoom, error) {
	if room == nil || room.ID <= 0 || room.AccountID == nil || *room.AccountID <= 0 {
		return nil, service.ErrCafeRoomInvalid
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	txClient := tx.Client()
	defer func() { _ = tx.Rollback() }()
	roomQ := txClient.CafeRoom.Query().Where(caferoom.IDEQ(room.ID))
	if txClient.Driver().Dialect() != dialect.SQLite {
		roomQ = roomQ.ForUpdate()
	}
	if _, err := roomQ.Only(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafeRoomNotFound
		}
		return nil, err
	}
	if err := lockAccountAndCheckAssignment(ctx, txClient, *room.AccountID, room.ID); err != nil {
		return nil, err
	}
	metadata := room.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	updated, err := txClient.CafeRoom.UpdateOneID(room.ID).
		SetCode(room.Code).
		SetName(room.Name).
		SetPlanID(room.PlanID).
		SetAccountID(*room.AccountID).
		SetZoneKey(room.ZoneKey).
		SetThemeKey(room.ThemeKey).
		SetSceneSlotKey(room.SceneSlotKey).
		SetStatus(room.Status).
		SetFeatured(room.Featured).
		SetSortOrder(room.SortOrder).
		SetMetadata(metadata).
		Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafeRoomNotFound
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	converted := cafeRoomToService(updated)
	return &converted, nil
}

func (r *cafeRoomRepository) Delete(ctx context.Context, id int64) error {
	if err := r.client.CafeRoom.DeleteOneID(id).Exec(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrCafeRoomNotFound
		}
		return err
	}
	return nil
}

func (r *cafeRoomRepository) CreateOpenRound(ctx context.Context, roomID int64, now time.Time) (*service.CafeRound, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	txClient := tx.Client()
	defer func() { _ = tx.Rollback() }()
	roomQ := txClient.CafeRoom.Query().Where(caferoom.IDEQ(roomID)).WithPlan(func(planQ *dbent.GroupBuyPlanQuery) {
		planQ.Where(groupbuyplan.DeletedAtIsNil())
		planQ.WithTargetGroup()
	})
	if txClient.Driver().Dialect() != dialect.SQLite {
		roomQ = roomQ.ForUpdate()
	}
	room, err := roomQ.Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafeRoomNotFound
		}
		return nil, err
	}
	if room.Status != service.CafeRoomStatusEnabled {
		return nil, service.ErrCafeRoomDisabled
	}
	if room.AccountID == nil || *room.AccountID <= 0 || room.Edges.Plan == nil {
		return nil, service.ErrCafeRoomInvalid
	}
	exists, err := txClient.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDEQ(roomID),
		groupbuyround.StatusIn(service.CafeRoundStatusOpen, "activating", "active"),
	).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, service.ErrCafeRoundExists
	}
	plan := room.Edges.Plan
	totalShares := plan.TotalShares
	if totalShares <= 0 {
		totalShares = plan.SeatCount
	}
	if totalShares <= 0 {
		return nil, service.ErrCafePlanInvalid
	}
	timeoutMinutes := plan.TimeoutMinutes
	if timeoutMinutes <= 0 {
		timeoutMinutes = 1440
	}
	created, err := txClient.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetCafeRoomID(room.ID).
		SetAssignedAccountID(*room.AccountID).
		SetRoomCodeSnapshot(room.Code).
		SetRoomNameSnapshot(room.Name).
		SetStatus(service.CafeRoundStatusOpen).
		SetTotalShares(totalShares).
		SetTotalSeats(totalShares).
		SetStartedAt(now).
		SetDeadlineAt(now.Add(time.Duration(timeoutMinutes) * time.Minute)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	converted := cafeRoundToService(created)
	return &converted, nil
}

func lockAccountAndCheckAssignment(ctx context.Context, client *dbent.Client, accountID, excludeRoomID int64) error {
	accountQ := client.Account.Query().Where(account.IDEQ(accountID))
	if client.Driver().Dialect() != dialect.SQLite {
		accountQ = accountQ.ForUpdate()
	}
	if _, err := accountQ.Only(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrCafeAccountNotFound
		}
		return err
	}
	roomQ := client.CafeRoom.Query().Where(
		caferoom.AccountIDEQ(accountID),
		caferoom.StatusIn(service.CafeRoomStatusEnabled, service.CafeRoomStatusMaintenance),
	)
	if excludeRoomID > 0 {
		roomQ = roomQ.Where(caferoom.IDNEQ(excludeRoomID))
	}
	assigned, err := roomQ.Exist(ctx)
	if err != nil {
		return err
	}
	if assigned {
		return service.ErrCafeAccountAssigned
	}
	return nil
}

func cafeRoomToService(room *dbent.CafeRoom) service.CafeRoom {
	converted := service.CafeRoom{
		ID:           room.ID,
		Code:         room.Code,
		Name:         room.Name,
		PlanID:       room.PlanID,
		AccountID:    room.AccountID,
		ZoneKey:      room.ZoneKey,
		ThemeKey:     room.ThemeKey,
		SceneSlotKey: room.SceneSlotKey,
		Status:       room.Status,
		Featured:     room.Featured,
		SortOrder:    room.SortOrder,
		Metadata:     room.Metadata,
		CreatedAt:    room.CreatedAt,
		UpdatedAt:    room.UpdatedAt,
	}
	if room.Edges.Plan != nil {
		plan := cafePlanToService(room.Edges.Plan)
		converted.Plan = &plan
	}
	return converted
}

func cafePlanToService(plan *dbent.GroupBuyPlan) service.CafeRoomPlan {
	converted := service.CafeRoomPlan{
		ID:              plan.ID,
		Title:           plan.Title,
		TargetGroupID:   plan.TargetGroupID,
		FulfillmentMode: plan.FulfillmentMode,
		TotalShares:     plan.TotalShares,
		SeatCount:       plan.SeatCount,
		TimeoutMinutes:  plan.TimeoutMinutes,
		ValidityDays:    plan.ValidityDays,
	}
	if plan.Edges.TargetGroup != nil {
		converted.GroupPlatform = plan.Edges.TargetGroup.Platform
		converted.GroupAccessMode = plan.Edges.TargetGroup.AccessMode
	}
	return converted
}

func cafeRoundToService(round *dbent.GroupBuyRound) service.CafeRound {
	return service.CafeRound{
		ID:                round.ID,
		PlanID:            round.PlanID,
		CafeRoomID:        round.CafeRoomID,
		AssignedAccountID: round.AssignedAccountID,
		RoomCodeSnapshot:  round.RoomCodeSnapshot,
		RoomNameSnapshot:  round.RoomNameSnapshot,
		Status:            round.Status,
		TotalShares:       round.TotalShares,
		TotalSeats:        round.TotalSeats,
		DeadlineAt:        round.DeadlineAt,
		CreatedAt:         round.CreatedAt,
		UpdatedAt:         round.UpdatedAt,
	}
}

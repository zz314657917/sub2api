package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	dbgroup "github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type cafeRoomRepository struct {
	client *dbent.Client
}

var cafeCommercialLockStatuses = []string{
	service.CafeRoundStatusOpen,
	service.CafeRoundStatusAwaitingAccount,
	service.CafeRoundStatusActivating,
	service.CafeRoundStatusActive,
	service.CafeRoundStatusRefunding,
}

func NewCafeRoomRepository(client *dbent.Client) service.CafeRoomRepository {
	return &cafeRoomRepository{client: client}
}

func (r *cafeRoomRepository) List(ctx context.Context, params pagination.PaginationParams, status, zone, search string) ([]service.CafeRoom, *pagination.PaginationResult, error) {
	q := r.client.CafeRoom.Query().WithPlan(func(planQ *dbent.GroupBuyPlanQuery) {
		planQ.Where(groupbuyplan.DeletedAtIsNil())
		planQ.WithTargetGroup()
		planQ.WithRounds(func(roundQ *dbent.GroupBuyRoundQuery) {
			roundQ.Where(groupbuyround.StatusIn(cafeCommercialLockStatuses...)).Order(dbent.Desc(groupbuyround.FieldID)).Limit(1)
		})
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
		Order(cafeRoomListOrder(params)...).
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
			planQ.WithRounds(func(roundQ *dbent.GroupBuyRoundQuery) {
				roundQ.Where(groupbuyround.StatusIn(cafeCommercialLockStatuses...)).Order(dbent.Desc(groupbuyround.FieldID)).Limit(1)
			})
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

// ResolveDefaultRoomManagedGroupID returns the system-owned default group.
// Startup normally creates it; the lazy ensure keeps direct repository users
// and tests safe as well.
func (r *cafeRoomRepository) ResolveDefaultRoomManagedGroupID(ctx context.Context) (int64, error) {
	group, err := ensurePixelCafeDefaultManagedGroup(ctx, r.client)
	if err != nil {
		return 0, err
	}
	return group.ID, nil
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

func (r *cafeRoomRepository) ListAccountOptions(ctx context.Context, params service.CafeRoomAccountOptionParams) ([]service.CafeRoomAccountOption, *pagination.PaginationResult, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	q := r.client.Account.Query().Where(account.DeletedAtIsNil())
	if len(params.IDs) > 0 {
		q = q.Where(account.IDIn(params.IDs...))
	} else {
		plan, err := r.client.GroupBuyPlan.Query().
			Where(groupbuyplan.IDEQ(params.PlanID), groupbuyplan.DeletedAtIsNil()).
			WithTargetGroup().
			Only(ctx)
		if err != nil {
			if dbent.IsNotFound(err) {
				return nil, nil, service.ErrCafePlanNotFound
			}
			return nil, nil, err
		}
		if err := validateCafePlanEntity(plan); err != nil {
			return nil, nil, err
		}
		q = q.Where(
			account.StatusEQ(service.StatusActive),
			account.PlatformEQ(plan.Edges.TargetGroup.Platform),
			account.HasGroupsWith(dbgroup.IDEQ(plan.TargetGroupID), dbgroup.DeletedAtIsNil()),
			account.Not(account.HasCafeRoomsWith(caferoom.StatusIn(service.CafeRoomStatusEnabled, service.CafeRoomStatusMaintenance), cafeRoomAccountOptionExclusion(params.ExcludeRoomID))),
		)
		if search := strings.TrimSpace(params.Search); search != "" {
			q = q.Where(account.Or(
				account.NameContainsFold(search),
				account.PlatformContainsFold(search),
				func(selector *sql.Selector) {
					selector.Where(sqljson.StringContains(account.FieldCredentials, search, sqljson.Path("email")))
				},
			))
		}
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	items, err := q.Order(account.ByID(sql.OrderAsc())).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, nil, err
	}
	options := make([]service.CafeRoomAccountOption, 0, len(items))
	for _, item := range items {
		options = append(options, service.CafeRoomAccountOption{ID: item.ID, Name: item.Name, Platform: item.Platform, Status: item.Status, EmailMasked: cafeRoomAccountEmailMasked(item)})
	}
	return options, &pagination.PaginationResult{Page: page, PageSize: pageSize, Total: int64(total), Pages: int((total + pageSize - 1) / pageSize)}, nil
}

func cafeRoomAccountOptionExclusion(excludeRoomID int64) func(*sql.Selector) {
	return func(selector *sql.Selector) {
		if excludeRoomID > 0 {
			selector.Where(sql.NEQ(selector.C(caferoom.FieldID), excludeRoomID))
		}
	}
}

func cafeRoomAccountEmailMasked(item *dbent.Account) string {
	if item == nil {
		return ""
	}
	value, ok := item.Credentials["email"].(string)
	return service.MaskCafeEmail(value, ok)
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
		groupbuyround.StatusIn(cafeCommercialLockStatuses...),
	).Exist(ctx)
}

func (r *cafeRoomRepository) Create(ctx context.Context, room *service.CafeRoom) (*service.CafeRoom, error) {
	if room == nil {
		return nil, service.ErrCafeRoomInvalid
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	txClient := tx.Client()
	defer func() { _ = tx.Rollback() }()

	var plan *dbent.GroupBuyPlan
	if room.Plan != nil && room.PlanID <= 0 {
		plan, err = createCafeRoomPlanInTx(ctx, txClient, room)
		if err != nil {
			return nil, err
		}
		room.PlanID = plan.ID
	} else {
		plan, err = lockCafeRoomPlan(ctx, txClient, room.PlanID)
		if err != nil {
			return nil, err
		}
		used, err := txClient.CafeRoom.Query().Where(caferoom.PlanIDEQ(room.PlanID)).Exist(ctx)
		if err != nil {
			return nil, err
		}
		if used {
			return nil, service.ErrCafePlanAssigned
		}
		if room.Plan != nil {
			plan, err = updateCafeRoomPlanInTx(ctx, txClient, plan, room)
			if err != nil {
				return nil, err
			}
		}
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
	created.Edges.Plan = plan
	converted := cafeRoomToService(created)
	return &converted, nil
}

func (r *cafeRoomRepository) Update(ctx context.Context, room *service.CafeRoom) (*service.CafeRoom, error) {
	if room == nil || room.ID <= 0 {
		return nil, service.ErrCafeRoomInvalid
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	txClient := tx.Client()
	defer func() { _ = tx.Rollback() }()
	roomQ := txClient.CafeRoom.Query().Where(caferoom.IDEQ(room.ID)).WithPlan(func(planQ *dbent.GroupBuyPlanQuery) {
		planQ.WithTargetGroup()
	})
	if txClient.Driver().Dialect() != dialect.SQLite {
		roomQ = roomQ.ForUpdate()
	}
	currentRoom, err := roomQ.Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafeRoomNotFound
		}
		return nil, err
	}
	live, err := txClient.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDEQ(room.ID),
		groupbuyround.StatusIn(cafeCommercialLockStatuses...),
	).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if live && (currentRoom.PlanID != room.PlanID || room.Status != service.CafeRoomStatusEnabled) {
		return nil, service.ErrCafeRoomLive
	}
	plan, err := lockCafeRoomPlan(ctx, txClient, room.PlanID)
	if err != nil {
		return nil, err
	}
	if currentRoom.PlanID != room.PlanID {
		used, err := txClient.CafeRoom.Query().Where(caferoom.PlanIDEQ(room.PlanID), caferoom.IDNEQ(room.ID)).Exist(ctx)
		if err != nil {
			return nil, err
		}
		if used {
			return nil, service.ErrCafePlanAssigned
		}
	}
	if room.Plan != nil {
		if live && cafeCommercialPlanChanged(plan, room.Plan) {
			return nil, service.ErrCafeRoomLive
		}
		plan, err = updateCafeRoomPlanInTx(ctx, txClient, plan, room)
		if err != nil {
			return nil, err
		}
	}
	metadata := room.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	updated, err := txClient.CafeRoom.UpdateOneID(room.ID).
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
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafeRoomNotFound
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	updated.Edges.Plan = plan
	converted := cafeRoomToService(updated)
	return &converted, nil
}

func (r *cafeRoomRepository) Delete(ctx context.Context, id int64) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	txClient := tx.Client()
	defer func() { _ = tx.Rollback() }()
	roomQ := txClient.CafeRoom.Query().Where(caferoom.IDEQ(id))
	if txClient.Driver().Dialect() != dialect.SQLite {
		roomQ = roomQ.ForUpdate()
	}
	room, err := roomQ.Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrCafeRoomNotFound
		}
		return err
	}
	live, err := txClient.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDEQ(id),
		groupbuyround.StatusIn(cafeCommercialLockStatuses...),
	).Exist(ctx)
	if err != nil {
		return err
	}
	if live {
		return service.ErrCafeRoomLive
	}
	if err := txClient.CafeRoom.DeleteOneID(id).Exec(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrCafeRoomNotFound
		}
		return err
	}
	if _, err := txClient.GroupBuyPlan.Update().
		Where(groupbuyplan.IDEQ(room.PlanID), groupbuyplan.DeletedAtIsNil()).
		SetStatus(service.GroupBuyPlanStatusDisabled).
		SetDeletedAt(time.Now()).
		Save(ctx); err != nil {
		return err
	}
	return tx.Commit()
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
	if room.Edges.Plan == nil {
		return nil, service.ErrCafeRoomInvalid
	}
	plan, err := lockCafeRoomPlan(ctx, txClient, room.PlanID)
	if err != nil {
		return nil, err
	}
	exists, err := txClient.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDEQ(roomID),
		groupbuyround.StatusIn(cafeCommercialLockStatuses...),
	).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, service.ErrCafeRoundExists
	}
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
	_, tier, maxBuyers, maxSharesPerUser, fulfillmentTimeout := normalizedCafePlanEntityPolicy(plan)
	if plan.Edges.TargetGroup == nil {
		return nil, service.ErrCafeGroupInvalid
	}
	created, err := txClient.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetCafeRoomID(room.ID).
		SetRoomCodeSnapshot(room.Code).
		SetRoomNameSnapshot(room.Name).
		SetStatus(service.CafeRoundStatusOpen).
		SetTotalShares(totalShares).
		SetTotalSeats(totalShares).
		SetStartedAt(now).
		SetDeadlineAt(now.Add(time.Duration(timeoutMinutes) * time.Minute)).
		SetCafeFulfillmentVersion("membership_share").
		SetSubscriptionTier(tier).
		SetMaxBuyers(maxBuyers).
		SetMaxSharesPerUser(maxSharesPerUser).
		SetFulfillmentTimeoutMinutes(fulfillmentTimeout).
		SetValidityDaysSnapshot(plan.ValidityDays).
		SetTargetGroupIDSnapshot(plan.TargetGroupID).
		SetPlatformSnapshot(plan.Edges.TargetGroup.Platform).
		SetQuotaPerShareSnapshot(plan.RoomKeyQuotaUsd).
		SetRateLimit5hPerShareSnapshot(plan.RoomKeyRateLimit5h).
		SetRateLimit1dPerShareSnapshot(plan.RoomKeyRateLimit1d).
		SetRateLimit7dPerShareSnapshot(plan.RoomKeyRateLimit7d).
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

func createCafeRoomPlanInTx(ctx context.Context, client *dbent.Client, room *service.CafeRoom) (*dbent.GroupBuyPlan, error) {
	if room == nil || room.Plan == nil {
		return nil, service.ErrCafePlanInvalid
	}
	targetGroup, err := loadCafeRoomTargetGroup(ctx, client, room.Plan.TargetGroupID)
	if err != nil {
		return nil, err
	}
	rules, mapping := cafeRoomTierPolicy(room.Plan.TotalShares, room.Plan.TargetGroupID)
	plan, err := client.GroupBuyPlan.Create().
		SetTitle(room.Name).
		SetDescription(room.Description).
		SetProductKey(service.GroupBuyProductTokenPinPinPin).
		SetTotalShares(room.Plan.TotalShares).
		SetSeatCount(room.Plan.TotalShares).
		SetSubscriptionTier(room.Plan.SubscriptionTier).
		SetMaxBuyers(room.Plan.MaxBuyers).
		SetMaxSharesPerUser(room.Plan.MaxSharesPerUser).
		SetPricePerShare(room.Plan.PricePerShare).
		SetPricePerSeat(room.Plan.PricePerShare).
		SetPriceLabel(room.Plan.PriceLabel).
		SetQuotaPerShareLabel(room.Plan.QuotaPerShareLabel).
		SetQuotaLabel(room.Plan.QuotaPerShareLabel).
		SetFulfillmentTimeoutMinutes(room.Plan.FulfillmentTimeoutMinutes).
		SetTargetGroupID(room.Plan.TargetGroupID).
		SetTierGroupIds(mapping).
		SetTierRules(rules).
		SetValidityDays(room.Plan.ValidityDays).
		SetTimeoutMinutes(room.Plan.TimeoutMinutes).
		SetLaunchMode(service.GroupBuyLaunchModeManual).
		SetFulfillmentMode(service.CafeRoomFulfillmentMode).
		SetRoomKeyQuotaUsd(room.Plan.RoomKeyQuotaUsd).
		SetRoomKeyRateLimit5h(room.Plan.RoomKeyRateLimit5h).
		SetRoomKeyRateLimit1d(room.Plan.RoomKeyRateLimit1d).
		SetRoomKeyRateLimit7d(room.Plan.RoomKeyRateLimit7d).
		SetAutoCreateRoomKey(true).
		SetRefundMode(room.Plan.RefundMode).
		SetAgreementText(room.Plan.AgreementText).
		SetStatus(service.GroupBuyPlanStatusActive).
		SetSortOrder(room.SortOrder).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	plan.Edges.TargetGroup = targetGroup
	if err := validateCafePlanEntity(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func updateCafeRoomPlanInTx(ctx context.Context, client *dbent.Client, plan *dbent.GroupBuyPlan, room *service.CafeRoom) (*dbent.GroupBuyPlan, error) {
	if plan == nil || room == nil || room.Plan == nil {
		return nil, service.ErrCafePlanInvalid
	}
	targetGroup, err := loadCafeRoomTargetGroup(ctx, client, room.Plan.TargetGroupID)
	if err != nil {
		return nil, err
	}
	rules, mapping := cafeRoomTierPolicy(room.Plan.TotalShares, room.Plan.TargetGroupID)
	updated, err := client.GroupBuyPlan.UpdateOneID(plan.ID).
		SetTitle(room.Name).
		SetDescription(room.Description).
		SetProductKey(service.GroupBuyProductTokenPinPinPin).
		SetTotalShares(room.Plan.TotalShares).
		SetSeatCount(room.Plan.TotalShares).
		SetSubscriptionTier(room.Plan.SubscriptionTier).
		SetMaxBuyers(room.Plan.MaxBuyers).
		SetMaxSharesPerUser(room.Plan.MaxSharesPerUser).
		SetPricePerShare(room.Plan.PricePerShare).
		SetPricePerSeat(room.Plan.PricePerShare).
		SetPriceLabel(room.Plan.PriceLabel).
		SetQuotaPerShareLabel(room.Plan.QuotaPerShareLabel).
		SetQuotaLabel(room.Plan.QuotaPerShareLabel).
		SetFulfillmentTimeoutMinutes(room.Plan.FulfillmentTimeoutMinutes).
		SetTargetGroupID(room.Plan.TargetGroupID).
		SetTierGroupIds(mapping).
		SetTierRules(rules).
		SetValidityDays(room.Plan.ValidityDays).
		SetTimeoutMinutes(room.Plan.TimeoutMinutes).
		SetLaunchMode(service.GroupBuyLaunchModeManual).
		SetFulfillmentMode(service.CafeRoomFulfillmentMode).
		SetRoomKeyQuotaUsd(room.Plan.RoomKeyQuotaUsd).
		SetRoomKeyRateLimit5h(room.Plan.RoomKeyRateLimit5h).
		SetRoomKeyRateLimit1d(room.Plan.RoomKeyRateLimit1d).
		SetRoomKeyRateLimit7d(room.Plan.RoomKeyRateLimit7d).
		SetAutoCreateRoomKey(true).
		SetRefundMode(room.Plan.RefundMode).
		SetAgreementText(room.Plan.AgreementText).
		SetStatus(service.GroupBuyPlanStatusActive).
		SetSortOrder(room.SortOrder).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	updated.Edges.TargetGroup = targetGroup
	if err := validateCafePlanEntity(updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func loadCafeRoomTargetGroup(ctx context.Context, client *dbent.Client, groupID int64) (*dbent.Group, error) {
	group, err := client.Group.Query().Where(dbgroup.IDEQ(groupID), dbgroup.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, service.ErrCafeGroupInvalid
	}
	if group.Status != service.StatusActive || group.SubscriptionType != service.SubscriptionTypeSubscription || group.AccessMode != service.CafeRoomGroupAccessMode {
		return nil, service.ErrCafeGroupInvalid
	}
	return group, nil
}

func cafeRoomTierPolicy(totalShares int, groupID int64) ([]domain.GroupBuyTierRule, map[string]int64) {
	mapping := make(map[string]int64, totalShares)
	for share := 1; share <= totalShares; share++ {
		mapping[fmt.Sprintf("%d", share)] = groupID
	}
	return []domain.GroupBuyTierRule{{MinShares: 1, MaxShares: totalShares, TargetGroupID: groupID, Label: fmt.Sprintf("1-%d 份", totalShares)}}, mapping
}

func cafeCommercialPlanChanged(current *dbent.GroupBuyPlan, next *service.CafeRoomPlan) bool {
	if current == nil || next == nil {
		return true
	}
	return nullableCafeString(current.Description) != next.Description ||
		current.SubscriptionTier != next.SubscriptionTier ||
		current.TotalShares != next.TotalShares ||
		current.MaxBuyers != next.MaxBuyers ||
		current.MaxSharesPerUser != next.MaxSharesPerUser ||
		current.PricePerShare != next.PricePerShare ||
		current.PriceLabel != next.PriceLabel ||
		current.QuotaPerShareLabel != next.QuotaPerShareLabel ||
		current.TimeoutMinutes != next.TimeoutMinutes ||
		current.FulfillmentTimeoutMinutes != next.FulfillmentTimeoutMinutes ||
		current.ValidityDays != next.ValidityDays ||
		current.TargetGroupID != next.TargetGroupID ||
		current.RoomKeyQuotaUsd != next.RoomKeyQuotaUsd ||
		current.RoomKeyRateLimit5h != next.RoomKeyRateLimit5h ||
		current.RoomKeyRateLimit1d != next.RoomKeyRateLimit1d ||
		current.RoomKeyRateLimit7d != next.RoomKeyRateLimit7d ||
		current.RefundMode != next.RefundMode ||
		nullableCafeString(current.AgreementText) != next.AgreementText
}

func nullableCafeString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func lockCafeRoomAccount(ctx context.Context, client *dbent.Client, accountID, excludeRoomID int64, plan *dbent.GroupBuyPlan) error {
	accountQ := client.Account.Query().Where(account.IDEQ(accountID)).WithGroups()
	if client.Driver().Dialect() != dialect.SQLite {
		accountQ = accountQ.ForUpdate()
	}
	assignedAccount, err := accountQ.Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrCafeAccountNotFound
		}
		return err
	}
	if assignedAccount.Status != service.StatusActive {
		return service.ErrCafeAccountIncompatible
	}
	if err := validateCafePlanFields(plan); err != nil {
		return err
	}
	if plan.Edges.TargetGroup == nil {
		group, err := client.Group.Query().Where(dbgroup.IDEQ(plan.TargetGroupID), dbgroup.DeletedAtIsNil()).Only(ctx)
		if err != nil {
			if dbent.IsNotFound(err) {
				return service.ErrCafeGroupInvalid
			}
			return err
		}
		plan.Edges.TargetGroup = group
	}
	if plan.Edges.TargetGroup.Status != service.StatusActive || plan.Edges.TargetGroup.AccessMode != service.CafeRoomGroupAccessMode {
		return service.ErrCafeGroupInvalid
	}
	if assignedAccount.Platform != plan.Edges.TargetGroup.Platform || !accountBelongsToGroup(assignedAccount, plan.TargetGroupID) {
		return service.ErrCafeAccountIncompatible
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

func validateCafePlanEntity(plan *dbent.GroupBuyPlan) error {
	if err := validateCafePlanFields(plan); err != nil {
		return err
	}
	if plan.Edges.TargetGroup == nil || plan.Edges.TargetGroup.Status != service.StatusActive || plan.Edges.TargetGroup.SubscriptionType != service.SubscriptionTypeSubscription || plan.Edges.TargetGroup.AccessMode != service.CafeRoomGroupAccessMode {
		return service.ErrCafeGroupInvalid
	}
	return nil
}

func validateCafePlanFields(plan *dbent.GroupBuyPlan) error {
	if plan == nil {
		return service.ErrCafePlanInvalid
	}
	totalShares, tier, maxBuyers, maxSharesPerUser, fulfillmentTimeout := normalizedCafePlanEntityPolicy(plan)
	if plan.Status != service.GroupBuyPlanStatusActive || plan.FulfillmentMode != service.CafeRoomFulfillmentMode || !plan.AutoCreateRoomKey || plan.ValidityDays <= 0 ||
		(tier != "plus" && tier != "pro") || totalShares < 1 || totalShares > 10 ||
		maxBuyers < 1 || maxBuyers > totalShares || maxSharesPerUser < 1 || maxSharesPerUser > totalShares || fulfillmentTimeout <= 0 {
		return service.ErrCafePlanInvalid
	}
	return nil
}

func normalizedCafePlanEntityPolicy(plan *dbent.GroupBuyPlan) (int, string, int, int, int) {
	if plan == nil {
		return 0, "", 0, 0, 0
	}
	totalShares := plan.TotalShares
	if totalShares <= 0 {
		totalShares = plan.SeatCount
	}
	tier := strings.ToLower(strings.TrimSpace(plan.SubscriptionTier))
	if tier == "" {
		tier = "plus"
	}
	maxBuyers := plan.MaxBuyers
	if maxBuyers <= 0 && totalShares > 0 {
		maxBuyers = totalShares
		if maxBuyers > 4 {
			maxBuyers = 4
		}
	}
	maxSharesPerUser := plan.MaxSharesPerUser
	if maxSharesPerUser <= 0 {
		maxSharesPerUser = totalShares
	}
	fulfillmentTimeout := plan.FulfillmentTimeoutMinutes
	if fulfillmentTimeout <= 0 {
		fulfillmentTimeout = 1440
	}
	return totalShares, tier, maxBuyers, maxSharesPerUser, fulfillmentTimeout
}

func accountBelongsToGroup(item *dbent.Account, groupID int64) bool {
	if item == nil {
		return false
	}
	for _, group := range item.Edges.Groups {
		if group.ID == groupID {
			return true
		}
	}
	return false
}

func lockCafeRoomPlan(ctx context.Context, client *dbent.Client, planID int64) (*dbent.GroupBuyPlan, error) {
	planQ := client.GroupBuyPlan.Query().Where(
		groupbuyplan.IDEQ(planID),
		groupbuyplan.DeletedAtIsNil(),
	).WithTargetGroup()
	if client.Driver().Dialect() != dialect.SQLite {
		planQ = planQ.ForUpdate()
	}
	plan, err := planQ.Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrCafePlanNotFound
		}
		return nil, err
	}
	if err := validateCafePlanFields(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func cafeRoomListOrder(params pagination.PaginationParams) []caferoom.OrderOption {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderAsc)
	if sortBy == "" {
		return []caferoom.OrderOption{caferoom.ByFeatured(sql.OrderDesc()), caferoom.BySortOrder(sql.OrderAsc()), caferoom.ByID(sql.OrderAsc())}
	}
	var order func(...sql.OrderTermOption) caferoom.OrderOption
	switch sortBy {
	case "id":
		order = caferoom.ByID
	case "code":
		order = caferoom.ByCode
	case "name":
		order = caferoom.ByName
	case "status":
		order = caferoom.ByStatus
	case "featured":
		order = caferoom.ByFeatured
	case "sort_order":
		order = caferoom.BySortOrder
	default:
		order = caferoom.BySortOrder
	}
	if sortOrder == pagination.SortOrderDesc {
		return []caferoom.OrderOption{order(sql.OrderDesc()), caferoom.ByID(sql.OrderDesc())}
	}
	return []caferoom.OrderOption{order(sql.OrderAsc()), caferoom.ByID(sql.OrderAsc())}
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
		converted.Description = plan.Description
	}
	return converted
}

func cafePlanToService(plan *dbent.GroupBuyPlan) service.CafeRoomPlan {
	totalShares, tier, maxBuyers, maxSharesPerUser, fulfillmentTimeout := normalizedCafePlanEntityPolicy(plan)
	converted := service.CafeRoomPlan{
		ID:                        plan.ID,
		Title:                     plan.Title,
		Description:               nullableCafeString(plan.Description),
		ProductKey:                plan.ProductKey,
		Status:                    plan.Status,
		TargetGroupID:             plan.TargetGroupID,
		FulfillmentMode:           plan.FulfillmentMode,
		AutoCreateRoomKey:         plan.AutoCreateRoomKey,
		TotalShares:               totalShares,
		SubscriptionTier:          tier,
		MaxBuyers:                 maxBuyers,
		MaxSharesPerUser:          maxSharesPerUser,
		FulfillmentTimeoutMinutes: fulfillmentTimeout,
		SeatCount:                 plan.SeatCount,
		PricePerShare:             plan.PricePerShare,
		PriceLabel:                plan.PriceLabel,
		QuotaPerShareLabel:        plan.QuotaPerShareLabel,
		TimeoutMinutes:            plan.TimeoutMinutes,
		ValidityDays:              plan.ValidityDays,
		RoomKeyQuotaUsd:           plan.RoomKeyQuotaUsd,
		RoomKeyRateLimit5h:        plan.RoomKeyRateLimit5h,
		RoomKeyRateLimit1d:        plan.RoomKeyRateLimit1d,
		RoomKeyRateLimit7d:        plan.RoomKeyRateLimit7d,
		LaunchMode:                plan.LaunchMode,
		RefundMode:                plan.RefundMode,
		AgreementText:             nullableCafeString(plan.AgreementText),
		SortOrder:                 plan.SortOrder,
	}
	if plan.Edges.TargetGroup != nil {
		converted.GroupPlatform = plan.Edges.TargetGroup.Platform
		converted.GroupAccessMode = plan.Edges.TargetGroup.AccessMode
		converted.TargetGroupStatus = plan.Edges.TargetGroup.Status
	}
	if len(plan.Edges.Rounds) > 0 {
		converted.CurrentRoundStatus = plan.Edges.Rounds[0].Status
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

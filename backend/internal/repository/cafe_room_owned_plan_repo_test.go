package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	dbgroup "github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func newCafeOwnedPlanTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	db, err := sql.Open("sqlite", fmt.Sprintf("file:cafe_owned_plan_%s?mode=memory&cache=shared&_fk=1", t.Name()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, db))))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func createCafeOwnedPlanGroup(t *testing.T, ctx context.Context, client *dbent.Client) int64 {
	t.Helper()
	group, err := client.Group.Create().
		SetName("Cafe managed group").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		SetAccessMode(service.CafeRoomGroupAccessMode).
		Save(ctx)
	require.NoError(t, err)
	return group.ID
}

func TestCafeRoomRepositoryEnsuresAndProtectsSystemDefaultManagedGroup(t *testing.T) {
	ctx := context.Background()
	client := newCafeOwnedPlanTestClient(t)
	repo := NewCafeRoomRepository(client)

	arbitrary, err := client.Group.Create().
		SetName("Existing non-default managed group").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		SetAccessMode(service.CafeRoomGroupAccessMode).
		SetSortOrder(-2000).
		Save(ctx)
	require.NoError(t, err)

	resolved, err := repo.ResolveDefaultRoomManagedGroupID(ctx)
	require.NoError(t, err)
	require.NotEqual(t, arbitrary.ID, resolved)
	managed, err := client.Group.Query().Where(
		dbgroup.IDEQ(resolved),
		dbgroup.DuplicateOperationIDEQ(service.CafeDefaultManagedGroupMarker),
	).Only(ctx)
	require.NoError(t, err)
	require.True(t, isValidPixelCafeDefaultManagedGroup(managed))

	repeated, err := repo.ResolveDefaultRoomManagedGroupID(ctx)
	require.NoError(t, err)
	require.Equal(t, resolved, repeated)

	groupRepo := NewGroupRepository(client, nil)
	groupView, err := groupRepo.GetByIDLite(ctx, resolved)
	require.NoError(t, err)
	groupView.Status = service.StatusDisabled
	require.ErrorIs(t, groupRepo.Update(ctx, groupView), service.ErrCafeDefaultGroupProtected)
	require.ErrorIs(t, groupRepo.Delete(ctx, resolved), service.ErrCafeDefaultGroupProtected)

	// Directly emulate an old local database where the feature data was deleted.
	require.NoError(t, client.Group.DeleteOneID(resolved).Exec(ctx))
	recreated, err := repo.ResolveDefaultRoomManagedGroupID(ctx)
	require.NoError(t, err)
	require.NotEqual(t, resolved, recreated)
}

func cafeOwnedPlanRoom(code string, groupID int64) *service.CafeRoom {
	return &service.CafeRoom{
		Code: code, Name: code, Description: "owned plan", Status: service.CafeRoomStatusEnabled,
		ZoneKey: "featured", ThemeKey: "warm_wood",
		Plan: &service.CafeRoomPlan{
			Title: code, Description: "owned plan", SubscriptionTier: "pro", TotalShares: 10,
			MaxBuyers: 4, MaxSharesPerUser: 10, PricePerShare: 12,
			TimeoutMinutes: 60, FulfillmentTimeoutMinutes: 1440, ValidityDays: 30,
			TargetGroupID: groupID, RefundMode: service.GroupBuyRefundModeBalanceCredit,
		},
	}
}

func TestCafeRoomRepositoryOwnedPlanLifecycle(t *testing.T) {
	ctx := context.Background()
	client := newCafeOwnedPlanTestClient(t)
	repo := NewCafeRoomRepository(client)
	groupID := createCafeOwnedPlanGroup(t, ctx, client)

	room, err := repo.Create(ctx, cafeOwnedPlanRoom("ROOM-OWNED-1", groupID))
	require.NoError(t, err)
	require.Positive(t, room.PlanID)
	require.NotNil(t, room.Plan)
	require.Equal(t, "pro", room.Plan.SubscriptionTier)
	require.Equal(t, 1, client.GroupBuyPlan.Query().Where(groupbuyplan.DeletedAtIsNil()).CountX(ctx))

	_, err = repo.Create(ctx, cafeOwnedPlanRoom("ROOM-OWNED-1", groupID))
	require.Error(t, err)
	require.Equal(t, 1, client.GroupBuyPlan.Query().Where(groupbuyplan.DeletedAtIsNil()).CountX(ctx), "room failure must roll back its freshly-created plan")

	_, err = repo.Create(ctx, &service.CafeRoom{Code: "ROOM-LEGACY-DUP", Name: "duplicate", PlanID: room.PlanID, Status: service.CafeRoomStatusEnabled})
	require.ErrorIs(t, err, service.ErrCafePlanAssigned)

	round, err := repo.CreateOpenRound(ctx, room.ID, room.CreatedAt)
	require.NoError(t, err)
	changed := *room
	changedPlan := *room.Plan
	changed.Plan = &changedPlan
	changed.Plan.PricePerShare++
	_, err = repo.Update(ctx, &changed)
	require.ErrorIs(t, err, service.ErrCafeRoomLive)

	current, err := repo.GetByID(ctx, room.ID)
	require.NoError(t, err)
	blockedDescription := *current
	blockedDescription.Description = "description cannot change while active"
	blockedPlan := *current.Plan
	blockedPlan.Description = blockedDescription.Description
	blockedDescription.Plan = &blockedPlan
	_, err = repo.Update(ctx, &blockedDescription)
	require.ErrorIs(t, err, service.ErrCafeRoomLive)

	visual := *current
	visual.Name = "Renamed while active"
	_, err = repo.Update(ctx, &visual)
	require.NoError(t, err)
	require.ErrorIs(t, repo.Delete(ctx, room.ID), service.ErrCafeRoomLive)
	client.GroupBuyRound.UpdateOneID(round.ID).SetStatus(service.GroupBuyRoundStatusCompleted).SaveX(ctx)

	terminal, err := repo.GetByID(ctx, room.ID)
	require.NoError(t, err)
	terminalPlan := *terminal.Plan
	terminalPlan.PricePerShare++
	terminal.Plan = &terminalPlan
	updated, err := repo.Update(ctx, terminal)
	require.NoError(t, err)
	require.Equal(t, terminalPlan.PricePerShare, updated.Plan.PricePerShare)
}

func TestCafeRoomRepositoryPauseRequiresAnEmptyOpenRoundAndAllowsReopen(t *testing.T) {
	ctx := context.Background()
	client := newCafeOwnedPlanTestClient(t)
	repo := NewCafeRoomRepository(client)
	groupID := createCafeOwnedPlanGroup(t, ctx, client)
	room, err := repo.Create(ctx, cafeOwnedPlanRoom("ROOM-PAUSE-1", groupID))
	require.NoError(t, err)

	round, err := repo.CreateOpenRound(ctx, room.ID, time.Now().UTC())
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("pause-test@example.com").SetPasswordHash("test").Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(room.PlanID).
		SetUserID(user.ID).
		SetStatus(service.GroupBuySeatStatusLocked).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.PauseOpenRound(ctx, room.ID, time.Now().UTC())
	require.ErrorIs(t, err, service.ErrCafeRoundNotEmpty)
	require.Equal(t, service.CafeRoundStatusOpen, client.GroupBuyRound.GetX(ctx, round.ID).Status)

	client.GroupBuySeat.UpdateOneID(seat.ID).SetStatus(service.GroupBuySeatStatusReleased).SaveX(ctx)
	membership, err := client.CafeRoundMembership.Create().
		SetRoundID(round.ID).
		SetUserID(user.ID).
		SetStatus(service.GroupBuySeatStatusPaid).
		SetPaidShares(1).
		Save(ctx)
	require.NoError(t, err)
	_, err = repo.PauseOpenRound(ctx, room.ID, time.Now().UTC())
	require.ErrorIs(t, err, service.ErrCafeRoundNotEmpty)

	client.CafeRoundMembership.UpdateOneID(membership.ID).SetPaidShares(0).SaveX(ctx)
	pausedAt := time.Now().UTC().Truncate(time.Second)
	paused, err := repo.PauseOpenRound(ctx, room.ID, pausedAt)
	require.NoError(t, err)
	require.Equal(t, service.GroupBuyRoundStatusCancelled, paused.Status)
	stored := client.GroupBuyRound.GetX(ctx, round.ID)
	require.Equal(t, service.GroupBuyRoundStatusCancelled, stored.Status)
	require.NotNil(t, stored.ClosedAt)
	require.Equal(t, "paused by administrator", *stored.CloseReason)

	reopened, err := repo.CreateOpenRound(ctx, room.ID, pausedAt.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, service.CafeRoundStatusOpen, reopened.Status)
	require.NotEqual(t, round.ID, reopened.ID)
}

func TestCafeRoomRepositoryListDefaultsToPriorityInsteadOfFeatured(t *testing.T) {
	ctx := context.Background()
	client := newCafeOwnedPlanTestClient(t)
	repo := NewCafeRoomRepository(client)
	groupID := createCafeOwnedPlanGroup(t, ctx, client)

	lowerPriority := cafeOwnedPlanRoom("ROOM-PRIORITY-20", groupID)
	lowerPriority.Featured = true
	lowerPriority.SortOrder = 20
	createdLower, err := repo.Create(ctx, lowerPriority)
	require.NoError(t, err)
	higherPriority := cafeOwnedPlanRoom("ROOM-PRIORITY-10", groupID)
	higherPriority.Featured = false
	higherPriority.SortOrder = 10
	createdHigher, err := repo.Create(ctx, higherPriority)
	require.NoError(t, err)

	rooms, _, err := repo.List(ctx, pagination.PaginationParams{Page: 1, PageSize: 20}, "", "", "")
	require.NoError(t, err)
	require.Len(t, rooms, 2)
	require.Equal(t, createdHigher.ID, rooms[0].ID)
	require.Equal(t, createdLower.ID, rooms[1].ID)
}

func TestCafeRoomRepositoryDeleteSoftDeletesOwnedPlan(t *testing.T) {
	ctx := context.Background()
	client := newCafeOwnedPlanTestClient(t)
	repo := NewCafeRoomRepository(client)
	groupID := createCafeOwnedPlanGroup(t, ctx, client)
	room, err := repo.Create(ctx, cafeOwnedPlanRoom("ROOM-DELETE-1", groupID))
	require.NoError(t, err)
	room.Status = service.CafeRoomStatusMaintenance
	room, err = repo.Update(ctx, room)
	require.NoError(t, err)
	require.NoError(t, repo.Delete(ctx, room.ID))

	plan, err := client.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(room.PlanID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, service.GroupBuyPlanStatusDisabled, plan.Status)
	require.NotNil(t, plan.DeletedAt)
	require.False(t, client.GroupBuyRound.Query().Where(groupbuyround.PlanIDEQ(room.PlanID)).ExistX(ctx))
}

func TestCafeRoomRepositoryLegacyUnassignedPlanSyncsRoomCopy(t *testing.T) {
	ctx := context.Background()
	client := newCafeOwnedPlanTestClient(t)
	repo := NewCafeRoomRepository(client)
	groupID := createCafeOwnedPlanGroup(t, ctx, client)
	legacyPlan, err := client.GroupBuyPlan.Create().
		SetTitle("Detached legacy title").
		SetDescription("Detached legacy description").
		SetPricePerShare(12).
		SetPricePerSeat(12).
		SetTargetGroupID(groupID).
		SetFulfillmentMode(service.CafeRoomFulfillmentMode).
		Save(ctx)
	require.NoError(t, err)

	created, err := service.NewCafeRoomService(repo).Create(ctx, service.CafeRoomInput{
		Code: "ROOM-LEGACY-OWNED", Name: "Legacy owned room", Description: "Room-owned description",
		PlanID: legacyPlan.ID, Status: service.CafeRoomStatusEnabled,
	})
	require.NoError(t, err)
	require.Equal(t, legacyPlan.ID, created.PlanID)
	reloaded, err := client.GroupBuyPlan.Get(ctx, legacyPlan.ID)
	require.NoError(t, err)
	require.Equal(t, "Legacy owned room", reloaded.Title)
	require.NotNil(t, reloaded.Description)
	require.Equal(t, "Room-owned description", *reloaded.Description)
}

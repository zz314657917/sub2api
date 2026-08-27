package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type cafePublicSettingsStub struct {
	enabled bool
	err     error
}

func (s cafePublicSettingsStub) GetPublicSettings(context.Context) (*PublicSettings, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &PublicSettings{PixelCafeEnabled: s.enabled}, nil
}

func newCafePublicTestClient(t *testing.T, name string) *dbent.Client {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, db))))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestCafePublicServiceListsOnlyEnabledRoomPlansAndRedactsOperations(t *testing.T) {
	ctx := context.Background()
	client := newCafePublicTestClient(t, "cafe_public_list")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	roomPlan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 3)
	_, err := client.GroupBuyPlan.UpdateOneID(roomPlan.ID).
		SetFulfillmentMode(CafeRoomFulfillmentMode).
		SetQuotaPerShareLabel("每份独立 Key 额度").
		SetRoomKeyQuotaUsd(500).
		SetRoomKeyRateLimit5h(0).
		SetRoomKeyRateLimit1d(0).
		SetRoomKeyRateLimit7d(100).
		Save(ctx)
	require.NoError(t, err)

	room, err := client.CafeRoom.Create().
		SetCode("CAFE-001").
		SetName("OpenAI 包间").
		SetPlanID(roomPlan.ID).
		SetZoneKey("openai").
		SetThemeKey("green").
		SetStatus(CafeRoomStatusEnabled).
		SetFeatured(true).
		SetMetadata(map[string]any{"account_email": "private@example.com"}).
		Save(ctx)
	require.NoError(t, err)
	round, err := client.GroupBuyRound.Create().
		SetPlanID(roomPlan.ID).
		SetCafeRoomID(room.ID).
		SetStatus(CafeRoundStatusOpen).
		SetTotalShares(3).
		SetTotalSeats(3).
		SetPaidShares(1).
		SetPaidSeats(1).
		SetDeadlineAt(time.Now().Add(time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	user := createGroupBuyTestUser(t, ctx, client, "cafe-public@example.com")
	_, err = client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(roomPlan.ID).
		SetUserID(user.ID).
		SetSeatNo(1).
		SetStatus(GroupBuySeatStatusPaid).
		Save(ctx)
	require.NoError(t, err)

	aggregatePlan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 2)
	_, err = client.CafeRoom.Create().
		SetCode("CAFE-AGGREGATE").
		SetName("Legacy Room").
		SetPlanID(aggregatePlan.ID).
		SetStatus(CafeRoomStatusEnabled).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.CafeRoom.Create().
		SetCode("CAFE-DISABLED").
		SetName("Hidden Room").
		SetPlanID(roomPlan.ID).
		SetStatus(CafeRoomStatusDisabled).
		Save(ctx)
	require.NoError(t, err)

	svc := NewCafePublicService(client, cafePublicSettingsStub{enabled: true})
	rooms, page, err := svc.List(ctx, user.ID, CafePublicListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, rooms, 1)
	require.Equal(t, room.ID, rooms[0].ID)
	require.Equal(t, "每份独立 Key 额度", rooms[0].Plan.QuotaPerShareLabel)
	require.Equal(t, float64(500), rooms[0].Plan.RoomKeyQuotaUsd)
	require.Zero(t, rooms[0].Plan.RoomKeyRateLimit5h)
	require.Zero(t, rooms[0].Plan.RoomKeyRateLimit1d)
	require.Equal(t, float64(100), rooms[0].Plan.RoomKeyRateLimit7d)
	require.Equal(t, "available", rooms[0].PurchaseState)
	require.Len(t, rooms[0].SeatVisuals, 3)
	require.Equal(t, "paid", rooms[0].SeatVisuals[0].State)
	require.True(t, rooms[0].SeatVisuals[0].IsMine)
	require.NotEmpty(t, rooms[0].SeatVisuals[0].AvatarSeed)

	payload, err := json.Marshal(rooms[0])
	require.NoError(t, err)
	encoded := string(payload)
	for _, prohibited := range []string{
		"account_id", "assigned_account_id", "target_group_id", "user_id", "account_email", "metadata", "private@example.com",
	} {
		require.NotContains(t, encoded, prohibited)
	}

	detail, err := svc.Get(ctx, user.ID, room.ID)
	require.NoError(t, err)
	require.Equal(t, "cafe.v1", detail.APIVersion)
	require.Equal(t, "full_then_assign_account", detail.Rules.Activation)
	require.Equal(t, "automatic_after_fulfillment_timeout", detail.Rules.Refund)
	require.True(t, detail.Rules.OneKeyPerMember)

	overview, err := svc.Overview(ctx, user.ID, 1)
	require.NoError(t, err)
	require.Len(t, overview.Rooms, 1)
	require.Len(t, overview.Zones, 1)
	require.Equal(t, "openai", overview.Zones[0].Key)
	require.False(t, overview.Lobby.Available)
	require.Equal(t, "今日使用用户", overview.Lobby.Label)
}

func TestCafePublicServiceOverviewOrdersByPriorityInsteadOfFeatured(t *testing.T) {
	ctx := context.Background()
	client := newCafePublicTestClient(t, "cafe_public_overview_all_rooms")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 2)
	_, err := client.GroupBuyPlan.UpdateOneID(plan.ID).SetFulfillmentMode(CafeRoomFulfillmentMode).Save(ctx)
	require.NoError(t, err)

	regularRoom, err := client.CafeRoom.Create().
		SetCode("CAFE-REGULAR").
		SetName("普通包间").
		SetPlanID(plan.ID).
		SetZoneKey("openai").
		SetStatus(CafeRoomStatusEnabled).
		SetFeatured(false).
		SetSortOrder(10).
		Save(ctx)
	require.NoError(t, err)
	featuredRoom, err := client.CafeRoom.Create().
		SetCode("CAFE-FEATURED").
		SetName("精选包间").
		SetPlanID(plan.ID).
		SetZoneKey("openai").
		SetStatus(CafeRoomStatusEnabled).
		SetFeatured(true).
		SetSortOrder(20).
		Save(ctx)
	require.NoError(t, err)

	svc := NewCafePublicService(client, cafePublicSettingsStub{enabled: true})
	overview, err := svc.Overview(ctx, 1, 10)
	require.NoError(t, err)
	require.Len(t, overview.Rooms, 2)
	require.Equal(t, regularRoom.ID, overview.Rooms[0].ID)
	require.Equal(t, featuredRoom.ID, overview.Rooms[1].ID)
	require.False(t, overview.Rooms[0].Featured)
	require.True(t, overview.Rooms[1].Featured)
	require.Len(t, overview.Zones, 1)
	require.Equal(t, 2, overview.Zones[0].RoomCount)
}

func TestCafePublicServiceFailsClosedWhenFeatureIsDisabled(t *testing.T) {
	client := newCafePublicTestClient(t, "cafe_public_disabled")
	svc := NewCafePublicService(client, cafePublicSettingsStub{enabled: false})
	_, _, err := svc.List(context.Background(), 1, CafePublicListParams{})
	require.ErrorIs(t, err, ErrCafeDisabled)
	require.Equal(t, "CAFE_DISABLED", infraerrors.Reason(err))
}

func TestCafePublicServiceListsMyRoomsWithSafeStatusProjection(t *testing.T) {
	ctx := context.Background()
	client := newCafePublicTestClient(t, "cafe_my_rooms")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 5)
	_, err := client.GroupBuyPlan.UpdateOneID(plan.ID).SetFulfillmentMode(CafeRoomFulfillmentMode).Save(ctx)
	require.NoError(t, err)
	now := time.Date(2026, 8, 3, 16, 15, 0, 0, time.UTC)
	account, err := client.Account.Create().SetName("ChatGPT Pro 主账号").SetPlatform(PlatformOpenAI).
		SetType("api_key").SetStatus(StatusActive).SetCredentials(map[string]interface{}{"email": "owner@example.com", "api_key": "must-not-leak"}).
		SetExtra(map[string]interface{}{"codex_7d_used_percent": 37.5, "codex_7d_reset_at": now.Add(4 * 24 * time.Hour).Format(time.RFC3339), "private_note": "must-not-leak-extra"}).Save(ctx)
	require.NoError(t, err)
	room, err := client.CafeRoom.Create().SetCode("CAFE-MY-001").SetName("我的 ChatGPT 包间").SetPlanID(plan.ID).SetAccountID(account.ID).
		SetZoneKey("openai").SetThemeKey("warm_wood").SetStatus(CafeRoomStatusEnabled).Save(ctx)
	require.NoError(t, err)
	round, err := client.GroupBuyRound.Create().SetPlanID(plan.ID).SetCafeRoomID(room.ID).SetStatus("active").
		SetTotalShares(5).SetTotalSeats(5).SetPaidShares(5).SetPaidSeats(5).SetDeadlineAt(now.Add(-time.Hour)).Save(ctx)
	require.NoError(t, err)
	user := createGroupBuyTestUser(t, ctx, client, "cafe-my-rooms@example.com")
	otherUser := createGroupBuyTestUser(t, ctx, client, "cafe-my-rooms-other@example.com")
	activeSeat, err := client.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(user.ID).SetSeatNo(1).
		SetStatus(GroupBuySeatStatusActive).SetActivatedAt(now.Add(-time.Hour)).SetExpiresAt(now.Add(time.Hour)).Save(ctx)
	require.NoError(t, err)
	managedKey, err := client.APIKey.Create().SetUserID(user.ID).SetKey("sk-cafe-my-rooms-private").SetName("ChatGPT 包间 CAFE-MY-001 / 座位 1").
		SetStatus("disabled").SetQuota(100).SetQuotaUsed(12.3).SetRateLimit5h(10).SetRateLimit1d(20).SetRateLimit7d(80).SetUsage5h(2.5).SetUsage7d(18.75).
		SetWindow5hStart(now.Add(-2 * time.Hour)).SetWindow7dStart(now.Add(-72 * time.Hour)).
		SetManagedSourceType(APIKeyManagedSourceCafeRoomSeat).SetManagedSourceID(activeSeat.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.UpdateOneID(activeSeat.ID).SetBoundAPIKeyID(managedKey.ID).Save(ctx)
	require.NoError(t, err)
	paidSeat, err := client.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(user.ID).SetSeatNo(2).SetStatus(GroupBuySeatStatusPaid).Save(ctx)
	require.NoError(t, err)
	unsafeKey, err := client.APIKey.Create().SetUserID(otherUser.ID).SetKey("sk-not-for-client").SetName("ordinary key").SetStatus("active").
		SetManagedSourceType(APIKeyManagedSourceCafeRoomSeat).SetManagedSourceID(paidSeat.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.UpdateOneID(paidSeat.ID).SetBoundAPIKeyID(unsafeKey.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(user.ID).SetSeatNo(3).
		SetStatus(GroupBuySeatStatusLocked).SetLockedUntil(now.Add(time.Minute)).Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(user.ID).SetSeatNo(4).
		SetStatus(GroupBuySeatStatusLocked).SetLockedUntil(now.Add(-time.Minute)).Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(user.ID).SetSeatNo(5).
		SetStatus(GroupBuySeatStatusActive).SetExpiresAt(now.Add(-time.Minute)).Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(otherUser.ID).SetSeatNo(6).SetStatus(GroupBuySeatStatusPaid).Save(ctx)
	require.NoError(t, err)
	nonCafeRound, err := client.GroupBuyRound.Create().SetPlanID(plan.ID).SetStatus(GroupBuyRoundStatusOpen).
		SetTotalShares(1).SetTotalSeats(1).SetDeadlineAt(now.Add(time.Hour)).Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.Create().SetRoundID(nonCafeRound.ID).SetPlanID(plan.ID).SetUserID(user.ID).SetStatus(GroupBuySeatStatusPaid).Save(ctx)
	require.NoError(t, err)

	svc := NewCafePublicService(client, cafePublicSettingsStub{enabled: true})
	svc.now = func() time.Time { return now }
	items, page, err := svc.MyRooms(ctx, user.ID, CafeMyRoomsListParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(5), page.Total)
	require.Len(t, items, 5)
	var activeItem *CafeMyRoom
	var paidItem *CafeMyRoom
	for index := range items {
		item := &items[index]
		if item.Seat.ID == activeSeat.ID {
			activeItem = item
		}
		if item.Seat.ID == paidSeat.ID {
			paidItem = item
		}
	}
	require.NotNil(t, activeItem)
	require.Equal(t, room.ID, activeItem.Room.ID)
	require.Equal(t, plan.ID, activeItem.Plan.ID)
	require.NotNil(t, activeItem.Account)
	require.Equal(t, "ChatGPT Pro 主账号", activeItem.Account.Name)
	require.Equal(t, PlatformOpenAI, activeItem.Account.Platform)
	require.Equal(t, "o***r@example.com", activeItem.Account.EmailMasked)
	require.Len(t, activeItem.MemberAvatars, 4)
	require.NotEmpty(t, activeItem.MemberAvatars[0].AvatarSeed)
	require.NotNil(t, activeItem.Account.Remaining7dPercent)
	require.Equal(t, 62.5, *activeItem.Account.Remaining7dPercent)
	require.NotNil(t, activeItem.ManagedAPIKey)
	require.Equal(t, activeSeat.ActivatedAt, activeItem.ActivatedAt)
	require.Equal(t, activeSeat.ExpiresAt, activeItem.ExpiresAt)
	require.Equal(t, managedKey.ID, activeItem.ManagedAPIKey.ID)
	require.Equal(t, 2.5, activeItem.ManagedAPIKey.Usage5h)
	require.Equal(t, 18.75, activeItem.ManagedAPIKey.Usage7d)
	require.NotNil(t, activeItem.ManagedAPIKey.ResetAt5h)
	require.Equal(t, now.Add(3*time.Hour), *activeItem.ManagedAPIKey.ResetAt5h)
	require.NotNil(t, activeItem.ManagedAPIKey.ResetAt7d)
	require.Equal(t, now.Add(4*24*time.Hour), *activeItem.ManagedAPIKey.ResetAt7d)
	require.True(t, activeItem.ManagedAPIKey.Protected)
	require.NotNil(t, paidItem)
	require.Nil(t, paidItem.ManagedAPIKey)

	encoded, err := json.Marshal(items)
	require.NoError(t, err)
	for _, prohibited := range []string{"\"key\":", "masked_key", "user_id", "group_id", "managed_source_id", "account_id", "credentials", "\"extra\":", "codex_7d_used_percent", "window_5h_start", "window_7d_start", "must-not-leak", "owner@example.com", "sk-cafe-my-rooms-private", "sk-not-for-client", "cafe-my-rooms@example.com", "cafe-my-rooms-other@example.com"} {
		require.NotContains(t, string(encoded), prohibited)
	}

	activeItems, activePage, err := svc.MyRooms(ctx, user.ID, CafeMyRoomsListParams{Page: 1, PageSize: 20, Statuses: []string{CafeMyRoomStatusActive}})
	require.NoError(t, err)
	require.Equal(t, int64(1), activePage.Total)
	require.Len(t, activeItems, 1)
	require.Equal(t, activeSeat.ID, activeItems[0].Seat.ID)
	waitingItems, waitingPage, err := svc.MyRooms(ctx, user.ID, CafeMyRoomsListParams{Page: 1, PageSize: 20, Statuses: []string{CafeMyRoomStatusWaiting}})
	require.NoError(t, err)
	require.Equal(t, int64(2), waitingPage.Total)
	require.Len(t, waitingItems, 2)
	historyItems, historyPage, err := svc.MyRooms(ctx, user.ID, CafeMyRoomsListParams{Page: 1, PageSize: 1, Statuses: []string{CafeMyRoomStatusHistory}})
	require.NoError(t, err)
	require.Equal(t, int64(2), historyPage.Total)
	require.Len(t, historyItems, 1)
	require.Equal(t, 2, historyPage.Pages)
}

func TestCafeMyRoomManagedKeyProjectsOnlyCurrentFiniteWindows(t *testing.T) {
	now := time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC)
	active5hStart := now.Add(-90 * time.Minute)
	elapsed7dStart := now.Add(-8 * 24 * time.Hour)

	projected := cafeMyRoomManagedKey(&dbent.APIKey{
		ID: 42, RateLimit5h: 20, RateLimit7d: 80, Usage5h: 5.4, Usage7d: 31.2,
		Window5hStart: &active5hStart, Window7dStart: &elapsed7dStart,
	}, now)
	require.NotNil(t, projected)
	require.Equal(t, 5.4, projected.Usage5h)
	require.NotNil(t, projected.ResetAt5h)
	require.Equal(t, now.Add(210*time.Minute), *projected.ResetAt5h)
	require.Zero(t, projected.Usage7d)
	require.Nil(t, projected.ResetAt7d)

	notStarted := cafeMyRoomManagedKey(&dbent.APIKey{RateLimit5h: 20, Usage5h: 9}, now)
	require.Zero(t, notStarted.Usage5h)
	require.Nil(t, notStarted.ResetAt5h)

	unlimited := cafeMyRoomManagedKey(&dbent.APIKey{RateLimit5h: 0, Usage5h: 9, Window5hStart: &active5hStart}, now)
	require.Equal(t, 9.0, unlimited.Usage5h)
	require.Nil(t, unlimited.ResetAt5h)
}

func TestCafeAccountRemaining7dPercentUsesOnlyOpenAISnapshot(t *testing.T) {
	now := time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC)
	openAIAccount := &dbent.Account{Platform: PlatformOpenAI, Extra: map[string]interface{}{
		"codex_7d_used_percent": 37.5,
		"codex_7d_reset_at":     now.Add(24 * time.Hour).Format(time.RFC3339),
	}}
	remaining := cafeAccountRemaining7dPercent(openAIAccount, now)
	require.NotNil(t, remaining)
	require.Equal(t, 62.5, *remaining)

	require.Nil(t, cafeAccountRemaining7dPercent(&dbent.Account{Platform: PlatformOpenAI}, now))
	require.Nil(t, cafeAccountRemaining7dPercent(&dbent.Account{Platform: PlatformAnthropic, Extra: openAIAccount.Extra}, now))
}

func TestCafeMyRoomStatusParserFailsClosed(t *testing.T) {
	for _, raw := range []string{"active,active", "active,unknown", "active,,waiting", ","} {
		_, err := ParseCafeMyRoomStatuses(raw)
		require.ErrorIs(t, err, ErrCafeMyRoomsInvalidStatus, raw)
		require.Equal(t, "CAFE_MY_ROOMS_INVALID_STATUS", infraerrors.Reason(err), raw)
	}
}

func TestCafeAccountEmailProjectionFailsClosed(t *testing.T) {
	for _, account := range []*dbent.Account{
		nil,
		{Credentials: nil},
		{Credentials: map[string]interface{}{"email": ""}},
		{Credentials: map[string]interface{}{"email": 123}},
	} {
		require.Empty(t, cafeAccountEmailMasked(account))
	}
	require.Equal(t, "a***e@example.com", cafeAccountEmailMasked(&dbent.Account{
		Credentials: map[string]interface{}{"email": " alice@example.com "},
	}))
}

func TestCafeMyRoomAccountNameMasksEmailShapedLabels(t *testing.T) {
	masked := cafeAccountDisplayName("owner@example.com")
	require.Equal(t, "o***r@example.com", masked)
	encoded, err := json.Marshal(CafeMyRoom{Account: &CafeMyRoomAccount{Name: masked, EmailMasked: cafeAccountEmailMasked(&dbent.Account{Credentials: map[string]interface{}{"email": "owner@example.com", "api_key": "must-not-leak"}})}})
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "owner@example.com")
	require.NotContains(t, string(encoded), "must-not-leak")
	require.Equal(t, "Cafe account", cafeAccountDisplayName("Cafe account"))
}

func TestPublicCafeSeatVisualDoesNotUseRawUserIdentity(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	seatNo := 2
	round := &dbent.GroupBuyRound{ID: 100, TotalSeats: 3, Edges: dbent.GroupBuyRoundEdges{
		Seats: []*dbent.GroupBuySeat{{ID: 200, UserID: 9001, SeatNo: &seatNo, Status: GroupBuySeatStatusLocked, LockedUntil: pointerTime(now.Add(time.Minute))}},
	}}
	visuals := publicCafeSeatVisuals(round, 9001, now)
	require.Len(t, visuals, 3)
	require.Equal(t, "locked", visuals[1].State)
	require.True(t, visuals[1].IsMine)
	require.NotContains(t, visuals[1].AvatarSeed, "9001")
	require.Equal(t, visuals[1].AvatarSeed, publicCafeAvatarSeed(round.ID, seatNo))
}

func pointerTime(value time.Time) *time.Time {
	return &value
}

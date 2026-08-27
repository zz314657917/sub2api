package repository

import (
	"context"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const pixelCafeDefaultManagedGroupName = "像素网吧默认托管分组"

// ensurePixelCafeDefaultManagedGroup maintains the one system-owned group used
// by all newly-created Pixel Cafe rooms. The immutable operation marker is the
// identity; the visible name and sort order may still be edited by an admin.
func ensurePixelCafeDefaultManagedGroup(ctx context.Context, client *dbent.Client) (*dbent.Group, error) {
	if client == nil {
		return nil, fmt.Errorf("nil ent client")
	}
	existing, err := client.Group.Query().Where(
		group.DuplicateOperationIDEQ(service.CafeDefaultManagedGroupMarker),
		group.DeletedAtIsNil(),
	).Only(ctx)
	if err == nil {
		if !isValidPixelCafeDefaultManagedGroup(existing) {
			return nil, service.ErrCafeDefaultGroupProtected
		}
		return existing, nil
	}
	if !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("query Pixel Cafe default managed group: %w", err)
	}

	created, err := client.Group.Create().
		SetName(pixelCafeDefaultManagedGroupName).
		SetDescription("系统自动维护，供像素网吧房间托管订阅和账号路由使用").
		SetDuplicateOperationID(service.CafeDefaultManagedGroupMarker).
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		SetAccessMode(service.CafeRoomGroupAccessMode).
		SetModelMatchPatterns([]string{"*"}).
		SetRateMultiplier(1).
		SetDefaultValidityDays(30).
		SetSortOrder(-1000).
		Save(ctx)
	if err == nil {
		return created, nil
	}
	if dbent.IsConstraintError(err) {
		// Concurrent starts may race. Read back the marker winner.
		winner, queryErr := client.Group.Query().Where(
			group.DuplicateOperationIDEQ(service.CafeDefaultManagedGroupMarker),
			group.DeletedAtIsNil(),
		).Only(ctx)
		if queryErr == nil && isValidPixelCafeDefaultManagedGroup(winner) {
			return winner, nil
		}
	}
	return nil, fmt.Errorf("create Pixel Cafe default managed group: %w", err)
}

func isValidPixelCafeDefaultManagedGroup(item *dbent.Group) bool {
	return item != nil && item.Status == service.StatusActive &&
		item.Platform == service.PlatformOpenAI &&
		item.SubscriptionType == service.SubscriptionTypeSubscription &&
		item.AccessMode == service.CafeRoomGroupAccessMode
}

func isPixelCafeDefaultManagedGroup(item *dbent.Group) bool {
	return item != nil && item.DuplicateOperationID != nil &&
		*item.DuplicateOperationID == service.CafeDefaultManagedGroupMarker &&
		item.DeletedAt == nil
}

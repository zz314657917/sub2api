package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

func IsStudioBridgeGatewayContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value, _ := ctx.Value(ctxkey.StudioBridgeGateway).(bool)
	return value
}

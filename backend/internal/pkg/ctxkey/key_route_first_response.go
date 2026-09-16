package ctxkey

import "context"

// KeyRouteFirstResponseAttempt marks an upstream request whose context must
// remain cancelable while the route dispatcher waits for its first response.
const KeyRouteFirstResponseAttempt Key = "ctx_key_route_first_response_attempt"

func WithKeyRouteFirstResponseAttempt(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, KeyRouteFirstResponseAttempt, true)
}

func IsKeyRouteFirstResponseAttempt(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	marked, _ := ctx.Value(KeyRouteFirstResponseAttempt).(bool)
	return marked
}

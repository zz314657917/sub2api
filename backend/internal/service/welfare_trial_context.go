package service

import "context"

type newUserTrialContextKey struct{}

// NewUserTrialSession marks a gateway request that should be billed against the
// non-transferable new-user trial pool instead of the user's wallet balance.
type NewUserTrialSession struct {
	TrialID   int64
	UserID    int64
	RequestID string
	QuotaLeft float64
}

func WithNewUserTrialSession(ctx context.Context, session *NewUserTrialSession) context.Context {
	if ctx == nil || session == nil {
		return ctx
	}
	return context.WithValue(ctx, newUserTrialContextKey{}, session)
}

func NewUserTrialSessionFromContext(ctx context.Context) (*NewUserTrialSession, bool) {
	if ctx == nil {
		return nil, false
	}
	session, ok := ctx.Value(newUserTrialContextKey{}).(*NewUserTrialSession)
	if !ok || session == nil || session.TrialID <= 0 || session.UserID <= 0 || session.RequestID == "" {
		return nil, false
	}
	return session, true
}

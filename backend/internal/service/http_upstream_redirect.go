package service

import "context"

type httpUpstreamDisableRedirectsContextKey struct{}

// WithHTTPUpstreamRedirectsDisabled prevents credential-bearing requests from
// following redirects through the shared upstream client.
func WithHTTPUpstreamRedirectsDisabled(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, httpUpstreamDisableRedirectsContextKey{}, true)
}

// HTTPUpstreamRedirectsDisabled reports whether redirects must be returned to
// the caller instead of followed.
func HTTPUpstreamRedirectsDisabled(ctx context.Context) bool {
	return ctx != nil && ctx.Value(httpUpstreamDisableRedirectsContextKey{}) == true
}

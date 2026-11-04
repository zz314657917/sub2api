package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type upstreamFiveAuthRepository struct {
	service.APIKeyRepository
	lookup func(context.Context, string) (*service.APIKey, error)
}

func (r *upstreamFiveAuthRepository) GetByKeyForAuth(ctx context.Context, key string) (*service.APIKey, error) {
	return r.lookup(ctx, key)
}

func TestUpstreamFiveAuthCancellationAndLiveRequestFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, google := range []bool{false, true} {
		for _, tc := range []struct {
			name            string
			cancel          bool
			err             error
			want            int
			requestDeadline bool
		}{
			{"cancelled_lookup", true, context.Canceled, 499, false},
			{"cancelled_driver_error", true, errors.New("driver connection closed"), 499, false},
			{"live_db_error", false, errors.New("db unavailable"), 500, false},
			{"live_db_deadline", false, fmt.Errorf("query: %w", context.DeadlineExceeded), 500, false},
			{"live_db_cancelled", false, fmt.Errorf("query: %w", context.Canceled), 500, false},
			{"missing_key", false, service.ErrAPIKeyNotFound, 401, false},
			{"request_deadline", false, context.DeadlineExceeded, 500, true},
		} {
			t.Run(fmt.Sprintf("google_%t/%s", google, tc.name), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if tc.requestDeadline {
					deadlineCtx, deadlineCancel := context.WithDeadline(ctx, time.Now().Add(-time.Second))
					defer deadlineCancel()
					ctx = deadlineCtx
					require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
				}
				calls := 0
				repo := &upstreamFiveAuthRepository{lookup: func(context.Context, string) (*service.APIKey, error) {
					calls++
					if tc.cancel {
						cancel()
					}
					return nil, tc.err
				}}
				cfg := &config.Config{RunMode: config.RunModeSimple}
				svc := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
				mw := gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg))
				if google {
					mw = APIKeyAuthWithSubscriptionGoogle(svc, nil, cfg)
				}
				router := gin.New()
				router.Use(mw)
				router.GET("/t", func(c *gin.Context) { t.Error("failed auth reached handler") })
				req := httptest.NewRequest(http.MethodGet, "/t", nil).WithContext(ctx)
				req.Header.Set("Authorization", "Bearer sk-private-test-value")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				require.Equal(t, tc.want, w.Code)
				require.Equal(t, 1, calls)
				require.NotContains(t, w.Body.String(), "sk-private-test-value")
				require.NotContains(t, w.Body.String(), tc.err.Error())
				if tc.want == 499 {
					require.Empty(t, w.Body.String())
				} else if google {
					require.Contains(t, w.Body.String(), `"code":`)
					require.Contains(t, w.Body.String(), `"status":`)
				} else {
					require.Contains(t, w.Body.String(), `"code":`)
					require.Contains(t, w.Body.String(), `"message":`)
				}
			})
		}
	}
}

func TestUpstreamFiveAuthInvalidUTF8BothEnvelopes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, google := range []bool{false, true} {
		t.Run(fmt.Sprintf("google_%t", google), func(t *testing.T) {
			repo := &upstreamFiveAuthRepository{lookup: func(context.Context, string) (*service.APIKey, error) {
				t.Fatal("invalid UTF-8 reached repository")
				return nil, nil
			}}
			cfg := &config.Config{RunMode: config.RunModeSimple}
			svc := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
			mw := gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg))
			if google {
				mw = APIKeyAuthWithSubscriptionGoogle(svc, nil, cfg)
			}
			router := gin.New()
			router.Use(mw)
			router.GET("/t", func(c *gin.Context) { t.Error("invalid auth reached handler") })
			req := httptest.NewRequest(http.MethodGet, "/t", nil)
			req.Header.Set("Authorization", "Bearer sk-\xff\xfe")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusUnauthorized, w.Code)
			require.Contains(t, w.Body.String(), "Invalid API key")
		})
	}
}
